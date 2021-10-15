// timescaledb-parallel-query creates some query performance data for a batch
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jhoelzel/pparallel-bench/internal/db"
	"github.com/jhoelzel/pparallel-bench/internal/resultaggregator"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	binName    = "pparallel-bench"
	version    = "0.1.0"
	tabCharStr = "\\t"
)

// Flag vars
var (
	postgresConnect string
	overrides       []db.Overrideable
	schemaName      string
	tableName       string
	queryCmd        string

	splitCharacter string
	fromFile       string
	toJson         string
	skipHeader     bool
	headerLinesCnt int

	tokenSize   int
	workers     int
	batchSize   int
	verbose     bool
	showVersion bool
)

// Parse args
func init() {
	var dbName string
	flag.StringVar(&postgresConnect, "connection", "host=localhost user=postgres sslmode=disable", "PostgreSQL connection url")
	flag.StringVar(&dbName, "db-name", "", "Database where the destination table exists")
	flag.StringVar(&tableName, "table", "cpu_usage", "Destination table for insertions")
	flag.StringVar(&schemaName, "schema", "public", "Destination table's schema")
	flag.StringVar(&queryCmd, "query", "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  %s WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;", "The query executed against the db. %s will be replaced by shema.tale")

	flag.StringVar(&splitCharacter, "split", ",", "Character to split by")
	flag.StringVar(&fromFile, "file", "", "File to read from rather than stdin")
	flag.StringVar(&toJson, "toJson", "", "return output to a JSON file")
	flag.BoolVar(&skipHeader, "skip-header", false, "Skip the first line of the input")
	flag.IntVar(&headerLinesCnt, "header-line-count", 1, "Number of header lines")

	flag.IntVar(&tokenSize, "token-size", bufio.MaxScanTokenSize, "Maximum size to use for tokens. By default, this is 64KB, so any value less than that will be ignored")
	flag.IntVar(&batchSize, "batch-size", 50, "Number of concurrent query executions of a worker")
	flag.IntVar(&workers, "workers", 1, "Number of parallel requests to make")
	flag.BoolVar(&verbose, "verbose", false, "Print more information about copying statistics")

	flag.BoolVar(&showVersion, "version", false, "Show the version of this tool")

	flag.Parse()

	if dbName != "" {
		overrides = append(overrides, db.OverrideDBName(dbName))
	}
}

//This program servers for performance benchmarking of queries. By giving you the possibility to define workers and batches parallel execution and more is possible
func main() {
	showTheVersion()

	scanner, file := prepareScanner()

	if headerLinesCnt <= 0 {
		fmt.Printf("WARNING: provided --header-line-count (%d) must be greater than 0\n", headerLinesCnt)
		os.Exit(1)
	}

	if tokenSize != 0 && tokenSize < bufio.MaxScanTokenSize {
		fmt.Printf("WARNING: provided --token-size (%d) is smaller than default (%d), ignoring\n", tokenSize, bufio.MaxScanTokenSize)
	} else if tokenSize > bufio.MaxScanTokenSize {
		buf := make([]byte, tokenSize)
		scanner.Buffer(buf, tokenSize)
	}
	//pepare buffered channels
	var wg sync.WaitGroup
	batchChan := make(chan []string, workers)
	resultChan := make(chan []resultaggregator.IdentifiedResults, workers)

	//initialize the resultaggregator
	fullTableName := fmt.Sprintf(`"%s"."%s"`, schemaName, tableName)
	queryResult := resultaggregator.NewQueryResult(fullTableName, queryCmd)

	// Generate  workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go processBatches(queryResult.QueryCmd, &wg, batchChan, resultChan)
	}

	//scan the input
	start := time.Now()
	rowsRead := scan(batchSize, scanner, batchChan)
	timeSpentScanning := time.Since(start)
	//when reading from file we need to close the file once its no longer used
	if file != nil {
		file.Close()
	}
	// after scanning all the workers have been dispatched so we can close the channel
	close(batchChan)
	wg.Wait()
	//now that all processing is done we will evaluate the results and also close resultchan
	close(resultChan)
	readStart := time.Now()
	readResults(resultChan, &queryResult)
	//and return the performance metrics
	queryResult.TimeSpentReading = time.Since(readStart)
	queryResult.TimeSpentScanning = timeSpentScanning
	queryResult.TimeSpentTotal = time.Since(start)
	queryResult.RowsRead = rowsRead
	queryResult.ProcessingComplete(verbose, workers)
	//write json output
	if toJson != "" {
		queryResult.ToJsonfile(toJson)
	}
}

//showTheVersion returns the verison number GOOS and GOARCH
func showTheVersion() {
	if showVersion {
		fmt.Printf("%s %s (%s %s)\n", binName, version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
}

//readResults reads the results from the return channel and aggregates them with the resultaggregator
func readResults(resultChan chan []resultaggregator.IdentifiedResults, queryResult *resultaggregator.QueryResults) {
	for cRess := range resultChan {
		for _, cRes := range cRess {
			queryResult.AddResults(cRes)
		}
	}
}

//prepareScanner takes a csv file either from stdin or file and returns it
//the file is only returned so it can be close after scanning instead of much to early
//the alternative is to read everything into ram directly and return that, but that is actually more memory than we need
func prepareScanner() (*bufio.Scanner, *os.File) {
	var scanner *bufio.Scanner
	if len(fromFile) > 0 {
		file, err := os.Open(fromFile)
		if err != nil {
			log.Fatal(err)
		}
		scanner = bufio.NewScanner(file)
		return scanner, file
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			fmt.Println("data is being piped to stdin")
		} else {
			log.Fatal("there seems to be no stdin data or input from --file aborting.")
		}
		scanner = bufio.NewScanner(os.Stdin)
	}
	return scanner, nil
}

//scan reads lines from a bufio.Scanner, each which should be in CSV format
//if a itemsPerBatch is provided the batch will be cut into smaller pieces
func scan(itemsPerBatch int, scanner *bufio.Scanner, batchChan chan []string) int64 {
	rows := make([]string, 0, itemsPerBatch)
	var linesRead int64

	if skipHeader {
		skipCsvHeader(scanner)
	}

	for scanner.Scan() {
		rows = append(rows, scanner.Text())
		if len(rows) >= itemsPerBatch { // dispatch to COPY worker & reset
			batchChan <- rows
			rows = make([]string, 0, itemsPerBatch)
		}
		linesRead++
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input: %s", err.Error())
	}

	// Finished reading input, make sure last batch goes out.
	if len(rows) > 0 {
		batchChan <- rows
	}
	return linesRead
}

//skipCsvHeader is skipping the first line of csv input because it contains headers that are not needed
func skipCsvHeader(scanner *bufio.Scanner) {
	if verbose {
		fmt.Printf("Skipping the first %d lines of the input.\n", headerLinesCnt)
	}
	for i := 0; i < headerLinesCnt; i++ {
		scanner.Scan()
	}
}

// processBatches reads batches from channel c and returns worked on batches to main thread
func processBatches(queryCmd string, wg *sync.WaitGroup, c chan []string, r chan []resultaggregator.IdentifiedResults) {
	dbx, err := db.Connect(postgresConnect, overrides...)
	if err != nil {
		//not being able to connecto the database is worth a panic
		fmt.Print("Unable to connect to the db ")
		panic(err)
	}
	defer dbx.Close()

	delimStr := "'" + splitCharacter + "'"
	useSplitChar := splitCharacter
	if splitCharacter == tabCharStr {
		delimStr = "E" + delimStr
		// Need to covert the string-ified version of the character to actual
		// character for correct split
		useSplitChar = "\t"
	}
	var container []resultaggregator.IdentifiedResults

	for batch := range c {
		start := time.Now()
		batchresult, err := processBatch2(dbx, batch, queryCmd, useSplitChar)
		if err != nil {
			panic(err)
		}
		container = append(container, batchresult...)

		if verbose {
			took := time.Since(start)
			fmt.Printf("[BATCH] took %v, batch size %d, row rate %f/sec\n", took, batchSize, float64(batchSize)/float64(took.Seconds()))
		}
	}
	r <- container
	defer wg.Done()
}

//prepareIdentifiedResult will create the holder object and split the csv input line into an interface array for execution in the db
func prepareIdentifiedResult(line string, splitChar string) (resultaggregator.IdentifiedResults, []interface{}) {
	sp := strings.Split(line, splitChar)
	args := make([]interface{}, len(sp))
	for i, v := range sp {
		args[i] = v
	}
	holder := resultaggregator.IdentifiedResults{
		Identificator: fmt.Sprint(args[0]), //we allways assume the first val in the where clause to be the identificator
		ExecutionTime: 0,
		RetrievalTime: 0,
	}
	return holder, args
}

//processBatch executes the queries for the batch as a prepared statement and returns the objects back to the caller
func processBatch2(db *sqlx.DB, batch []string, queryCmd, splitChar string) ([]resultaggregator.IdentifiedResults, error) {
	stmt, err := db.Prepare(queryCmd)
	if err != nil {
		return nil, err
	}
	batchresult := []resultaggregator.IdentifiedResults{}
	for _, line := range batch {
		holder, args := prepareIdentifiedResult(line, splitChar)
		//execute the query
		queryStart := time.Now()
		rows, err := stmt.Query(args...)
		holder.ExecutionTime = time.Since(queryStart)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to execute query %v\n", err)
			fmt.Fprintf(os.Stderr, "for Query \"%s\"\n", queryCmd)
			fmt.Fprintf(os.Stderr, "with parameter %v\n", args)
			//TODO: this error could be caught better for now we will ignore it
			continue
		}
		retirievalStart := time.Now()
		cols, _ := rows.Columns()
		for rows.Next() {
			//so we can scan the rows with automatic parameter adjustment
			columns, columnPointers := prePareColPointers(cols)
			err = rows.Scan(columnPointers...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to scan %v\n", err)
				fmt.Fprintf(os.Stderr, "with parameter %v\n", args)
				//TODO: this error could be caught better for now we will ignore it
				continue
			}
			colval := getColVal(cols, columns)
			holder.Results = append(holder.Results, colval)
		}
		holder.RetrievalTime = time.Since(retirievalStart)
		if err = rows.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to get query results of query %v\n", err)
		}
		if err != nil {
			return nil, err
		}
		batchresult = append(batchresult, holder)
		rows.Close()
	}
	err = stmt.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to close the db connection %v\n", err)
		return nil, err
	}
	return batchresult, nil
}

//prePareColPointers  prepares the arguments for row.scan in order to fetch data without structure
func prePareColPointers(cols []string) ([]string, []interface{}) {
	//we create an array of columlength
	columns := make([]string, len(cols))
	//and and []interface to catch it
	columnPointers := make([]interface{}, len(cols))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}
	return columns, columnPointers
}

//getColVal extracts the values of the current row scanned into a string map
func getColVal(cols []string, columns []string) map[string]string {
	colval := make(map[string]string)
	for i, colName := range cols {
		colval[colName] = columns[i]
	}
	return colval
}
