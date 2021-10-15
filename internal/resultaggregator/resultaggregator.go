//resultaggregator aggregates all results by Identificator
package resultaggregator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"
)

//Hostresult represtend multiple query results for a single host
type IdentifiedResults struct {
	Identificator string
	Results       []map[string]string
	ExecutionTime time.Duration
	RetrievalTime time.Duration
}

//QueryResults are the master object holders. they offer result representation and specific type based logic
type QueryResults struct {
	TableName         string
	QueryCmd          string
	Results           map[string]IdentifiedResults
	MinimumQueryTime  time.Duration
	QueryTimes        []float64
	MaximumQueryTime  time.Duration
	TotalQueryTime    time.Duration
	QueryCounter      int64
	ResultCounter     int64
	RowsRead          int64
	TimeSpentScanning time.Duration
	TimeSpentReading  time.Duration
	TimeSpentTotal    time.Duration
}

//NewQueryResult returns a new QueryResult which has been initialized
func NewQueryResult(fullTableName string, queryCMD string) QueryResults {
	queryResult := QueryResults{}
	queryResult.TableName = fullTableName
	//if no %s is given in the query we can assume that we do not need to replace the tablename in the query
	if strings.Contains(queryCMD, "%s") {
		queryResult.QueryCmd = fmt.Sprintf(queryCMD, fullTableName)
	} else {
		queryResult.QueryCmd = queryCMD
	}

	queryResult.Results = make(map[string]IdentifiedResults)
	queryResult.QueryCounter = 0
	queryResult.ResultCounter = 0
	return queryResult
}

//Addresults aggregates the results from all Identificators into grouped results
func (h *QueryResults) AddResults(hostres IdentifiedResults) {
	if h.QueryCounter == 0 {
		h.MinimumQueryTime = hostres.ExecutionTime
		h.MaximumQueryTime = hostres.ExecutionTime
	}

	if host, ok := h.Results[hostres.Identificator]; ok {
		host.Results = append(host.Results, hostres.Results...)
	} else {
		h.Results[hostres.Identificator] = hostres
	}
	if hostres.ExecutionTime < h.MinimumQueryTime {
		h.MinimumQueryTime = hostres.ExecutionTime
	}
	if hostres.ExecutionTime > h.MaximumQueryTime {
		h.MaximumQueryTime = hostres.ExecutionTime
	}
	h.ResultCounter += int64(len(hostres.Results))
	h.QueryTimes = append(h.QueryTimes, float64(hostres.ExecutionTime.Milliseconds()))
	h.TotalQueryTime += hostres.ExecutionTime
	h.QueryCounter += 1

}

//CalcMean calculates the mean response time in milliseconds
func (h *QueryResults) CalcMean() float64 {
	total := 0.0
	for _, v := range h.QueryTimes {
		total += v
	}
	return total / float64(len(h.QueryTimes))
}

//CalcMedian calculates the mean response time in milliseconds
func (h *QueryResults) CalcMedian() float64 {
	sort.Float64s(h.QueryTimes)        // sort the numbers
	meanIndex := len(h.QueryTimes) / 2 // get index
	if meanIndex%2 == 0 {              //even
		return (h.QueryTimes[meanIndex-1] + h.QueryTimes[meanIndex]) / 2
	}
	return h.QueryTimes[meanIndex]
}

//ToJsonResult() returns a json representation of the results
func (h *QueryResults) ToJsonResult() ([]byte, error) {
	return json.MarshalIndent(h, "", " ")
}

//processingComplete sends the output to the cli. The verbose setting provides more output details
func (h *QueryResults) ProcessingComplete(verbose bool, workercount int) {
	fmt.Println("########################################")
	fmt.Println("Processing Complete")
	fmt.Println(time.Now())
	fmt.Println("################")
	if verbose {
		fmt.Println("Processingstats:")
		fmt.Println("Time spent scanning the input:", h.TimeSpentScanning)
		fmt.Println("Time spent reading the responses:", h.TimeSpentReading)
		fmt.Println("################")
	}
	fmt.Println("Querystats:")
	fmt.Println("MinimumQueryTime:", h.MinimumQueryTime)
	fmt.Println("MaximumQueryTime:", h.MaximumQueryTime)
	fmt.Println("TotalQueryTime:", h.TotalQueryTime)
	fmt.Println("MeanQueryTime:", h.CalcMean(), "ms")
	fmt.Println("MedianQueryTime:", h.CalcMedian(), "ms")
	fmt.Println("Queries:", h.QueryCounter)
	if verbose {
		queryResultRate := float64(h.ResultCounter) / float64(h.QueryCounter)
		fmt.Println("Result Rows received:", h.ResultCounter)
		fmt.Println("Avg results per query:", queryResultRate)
	}
	fmt.Println("########################################")
	rowRate := float64(h.RowsRead) / float64(h.TotalQueryTime.Seconds())
	fmt.Printf("Querying  %d entries, took %v with %d worker(s) (mean rate %f/sec)\n", h.RowsRead, h.TimeSpentTotal, workercount, rowRate)
	if int64(h.RowsRead) != int64(h.QueryCounter) {
		faulty := h.RowsRead - h.QueryCounter
		strFaulty := fmt.Sprintf("Faulty Queries: %d \n", faulty)
		fmt.Fprintf(os.Stderr, "########################################\n")
		fmt.Fprintf(os.Stderr, "            WARNING!\n")
		fmt.Fprintf(os.Stderr, "########################################\n")
		fmt.Fprintf(os.Stderr, strFaulty)
		fmt.Fprintf(os.Stderr, "There has been more input found than queries executed,\n")
		fmt.Fprintf(os.Stderr, "please consult the Log for Errors!!\n")
		fmt.Fprintf(os.Stderr, "########################################\n")
	}

}

//ToJsonfile writes the entire QueryResult to Json
func (h *QueryResults) ToJsonfile(filename string) error {
	json, err := h.ToJsonResult()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, json, 0644)
	if err != nil {
		return err
	}
	return nil
}
