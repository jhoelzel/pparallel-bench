
# pparallel-bench

`pparallel-bench` is a command line program for parallelizing PostgreSQL's built-in `query` functionality for bulk measuring of performance data with [TimescaleDB.](//github.com/timescale/timescaledb/)
Its main use is to measure performance of queries that run against a database. The mean idea is to have one main query and to provide a CSV of data for its execution.
The Query as well as the data can be freely defined, but the first parameter of the query should be the column on which you want your results to be grouped at.

NOTE:
This Project is inspired by <https://github.com/timescale/timescaledb-parallel-copy> and this project also uses its intnal DB module

## Getting started

Before using this program to bulk query data, your database should be installed with the TimescaleDB extension and the target table should already be made a hypertable.
Tutorials for this purpose can be found in the official timescaldb documentation <https://docs.timescale.com/timescaledb/latest/how-to-guides/install-timescaledb/>.
If you already have a database set up you are ready to go.

## Defining the CSV file

The tool requires a csv file to be either provided via stdin or the --file parameter. Beware that all columns will be provided as parameter for the query.
Example data can look like this:

```
host_000008,2017-01-01 08:59:22,2017-01-01 09:59:22
```

but by using the -split flag you are able to provide your own seperators.

NOTE:
If you want to provide headers for your file you need to skip them with the -skip-header flag

```
hostname,start_time,end_time
host_000008,2017-01-01 08:59:22,2017-01-01 09:59:22
```

## Defining your Query

pparallel-bench can be used with any query you like, but since we are provided an input file for your query, the number of columns provided needs to match the number of parameters of the query.
if you want to provide your table name programatically through the shema and table name, your query needs to contain "%s" where the tablename is going to be placed

Example:

```
--query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  %s WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
```

but the tool acceptes any other form of query too:
Example:

```
--query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  public.cpu_usage WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
```

## Defining workers

In oder to use the power of go to query your server in parralel, this tool uses the --worker flag to define the number of concurrent workers that should query the server.
It defaults to one but can be used at will.

## Defining batches

Batches are used to limit the amout of Queries sent to your server at any given time. This brings rate-limiting and also makes this tool useable without flooding the server. Scenarios include for instance a production server that should be tested every hour for query performance but not take the entire server down while at it.
-batch-size <int> defines the Number of concurrent query executions of a worker.

## Defining the connection

this tool assumes a DB Server running on your local host on the default settings. If that is not the case a connection string can be provided using -connection  $(DBCONNECTIONSTRING)

Example:

```
-connection  "host=10.37.37.1 port=5432 user=postgres password=supersecureandmegahidden" sslmode==disable
```

## Building pparallel-bench

For Linux:
```
$ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/pparallel-bench ./cmd/pparallel-bench/main.go
```
For Windows:
```
$ CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/pparallel-bench ./cmd/pparallel-bench/main.go
```

## Installing into $GOPATH/bin 

You can install the tool directly to your bin folder with:
```
$ go install <Thisrepourl>
```


### Using pparallel-bench

Using pparallel-bench after installation with:

```bash
# single-threaded
$ pparallel-bench --db-name homework --table cpu_usage --file foo.csv --query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  %s WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"

# 2 workers
$ pparallel-bench --db-name homework --table cpu_usage --file foo.csv \
    --workers 2 --query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  %s WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"

# 2 workers verbose output
$ pparallel-bench --db-name homework --table cpu_usage --file foo.csv \
    --workers 2 --verbose --query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  %s WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"

# if your input file containe header you can skip them with --skip-header
$ pparallel-bench --db-name homework --table cpu_usage --file foo.csv \
    --skip-header --query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  %s WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
```

or if you rather prefer to execute the binary:

```bash
# single-threaded
$ ./bin/pparallel-bench --db-name homework --table cpu_usage --file foo.csv --query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  %s WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"

# 2 workers
$ ./bin/pparallel-bench --db-name homework --table cpu_usage --file foo.csv \
    --workers 2 --query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  %s WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"

# 2 workers verbose output
$ ./bin/pparallel-bench --db-name homework --table cpu_usage --file foo.csv \
    --workers 2 --verbose --query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  %s WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"

# if your input file containe header you can skip them with --skip-header
$ ./bin/pparallel-bench --db-name homework --table cpu_usage --file foo.csv \
    --skip-header --query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  %s WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
```


Other options and flags are also available:

```
$ pparallel-bench --help

Usage of timescaledb-parallel-copy:
  -batch-size int
        Number of concurrent query executions of a worker (default 50)
  -connection string
        PostgreSQL connection url (default "host=localhost user=postgres sslmode=disable")
  -db-name string
        Database where the destination table exists
  -file string
        File to read from rather than stdin
  -header-line-count int
        Number of header lines (default 1)
  -query string
        The query executed against the db. %s will be replaced by shema.tale (default "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  %s WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;")
  -schema string
        Destination table's schema (default "public")
  -skip-header
        Skip the first line of the input
  -split string
        Character to split by (default ",")
  -table string
        Destination table for insertions (default "cpu_usage")
  -toJson string
        return output to a JSON file
  -token-size int
        Maximum size to use for tokens. By default, this is 64KB, so any value less than that will be ignored (default 65536)
  -verbose
        Print more information about copying statistics
  -version
        Show the version of this tool
  -workers int
        Number of parallel requests to make (default 1)
```

## Result metrics 

The unverbose metrics are:
- #of queries processed
- total processing time across all queries
- the minimum query time (for a single query)
- the median query time
- the average query time
- and the maximum query time

### Regular output with 200 Entries, 2 workers and a batchsize of 25 without verbose output

```
########################################
Processing Complete
################
Querystats:
MinimumQueryTime: 4.1421ms
MaximumQueryTime: 11.3781ms
TotalQueryTime: 1.2689326s
MeanQueryTime: 5.83 ms
MedianQueryTime: 6 ms
Queries: 200
########################################
Querying  200 entries, took 706.2456ms with 2 worker(s) (mean rate 157.612784/sec)
```

### Regular output with 200 Entries, 2 workers and a batchsize of 25 with verbose output

command: 

make docker-cron-run-verbose DBHOST=<YourIP> DBPORT=<YourPort> DBUSER=<YourUser> DBPW=<DBPORT>

OR
./bin/timescale-parallelquery  -db-name homework -table cpu_usage -skip-header -file data/faulty_query_params.csv -workers $(WORKERS) -connection  $(DBCONNECTIONSTRING)  -batch-size $(BATCHSIZE) -verbose 

```
Skipping the first 1 lines of the input.
[BATCH] took 221.9052ms, batch size 25, row rate 112.660722/sec
[BATCH] took 235.6608ms, batch size 25, row rate 106.084678/sec
[BATCH] took 217.6501ms, batch size 25, row rate 114.863260/sec
[BATCH] took 203.8431ms, batch size 25, row rate 122.643347/sec
[BATCH] took 155.2121ms, batch size 25, row rate 161.069917/sec
[BATCH] took 167.8463ms, batch size 25, row rate 148.945791/sec
[BATCH] took 159.0662ms, batch size 25, row rate 157.167267/sec
[BATCH] took 157.3773ms, batch size 25, row rate 158.853913/sec
########################################
Processing Complete
################
Processingstats:
Time spent scanning the input: 460.3905ms
Time spent reading the responses: 669.8µs
################
Querystats:
MinimumQueryTime: 4.0404ms
MaximumQueryTime: 19.7758ms
TotalQueryTime: 1.38782s
MeanQueryTime: 6.44 ms
MedianQueryTime: 6 ms
Queries: 200
########################################
Querying  200 entries, took 786.2167ms with 2 worker(s) (mean rate 144.110908/sec)   
```

### Problematic output with 200 Entries, 2 workers and a batchsize of 25 with verbose output

Behaviour: When content with errors is provided, the tool will show those in the logs and skip ahead. Error output includes the query and its parameter for easy testing:

```
Skipping the first 1 lines of the input.
[BATCH] took 156.1017ms, batch size 25, row rate 160.152003/sec
[BATCH] took 159.4769ms, batch size 25, row rate 156.762515/sec
Unable to execute query sql: expected 3 arguments, got 2
for Query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  "public"."cpu_usage" WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
with parameter [host_000006 ]
Unable to execute query pq: invalid input syntax for type timestamp with time zone: "2017-01-01 23:as:52"
for Query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  "public"."cpu_usage" WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
with parameter [host_000003 2017-01-01 22:05:52 2017-01-01 23:as:52]
[BATCH] took 253.0613ms, batch size 25, row rate 98.790293/sec
[BATCH] took 265.2057ms, batch size 25, row rate 94.266451/sec
Unable to execute query pq: invalid input syntax for type timestamp with time zone: "2017-01-01 17:05:x"
for Query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  "public"."cpu_usage" WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
with parameter [host_000001 2017-01-01 17:05:x 2017-01-01 18:05:12]
Unable to execute query pq: invalid input syntax for type timestamp with time zone: "2017-01-a 14:29:53"
for Query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  "public"."cpu_usage" WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
with parameter [host_000002 2017-01-02 13:29:53 2017-01-a 14:29:53]
[BATCH] took 225.5779ms, batch size 25, row rate 110.826460/sec
[BATCH] took 225.909ms, batch size 25, row rate 110.664028/sec
Unable to execute query pq: invalid input syntax for type timestamp with time zone: "2017-a-02 09:13:47"
for Query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  "public"."cpu_usage" WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
with parameter [host_000001 2017-01-02 08:13:47 2017-a-02 09:13:47]
Unable to execute query pq: invalid input syntax for type timestamp with time zone: "2017-a-02 02:28:31"
for Query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  "public"."cpu_usage" WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
with parameter [host_000005 2017-a-02 02:28:31 2017-01-02 03:28:31]
Unable to execute query pq: invalid input syntax for type timestamp with time zone: "2017-01-01 a:40:32"
for Query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  "public"."cpu_usage" WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
with parameter [host_000005 2017-01-01 a:40:32 2017-01-01 12:40:32]
Unable to execute query pq: invalid input syntax for type timestamp with time zone: "a-01-02 19:21:03"
for Query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  "public"."cpu_usage" WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
with parameter [host_000002 2017-01-02 18:21:03 a-01-02 19:21:03]
[BATCH] took 186.4929ms, batch size 25, row rate 134.053361/sec
[BATCH] took 271.043ms, batch size 25, row rate 92.236287/sec
########################################
Processing Complete
################
Processingstats:
Time spent scanning the input: 436.7902ms
Time spent reading the responses: 528.2µs
################
Querystats:
MinimumQueryTime: 3.7385ms
MaximumQueryTime: 19.3715ms
TotalQueryTime: 1.5025836s
MeanQueryTime: 7.354166666666667 ms
MedianQueryTime: 7 ms
Queries: 192
########################################
Querying  200 entries, took 934.3796ms with 2 worker(s) (mean rate 133.104075/sec)
########################################
            WARNING!
########################################
There has been more input found than queries executed,
please consult the Log for Errors!!
########################################
```

### Contributing
We welcome contributions to this utility, which like TimescaleDB is released under the Apache2 Open Source License.  The same [Contributors Agreement](//github.com/timescale/timescaledb/blob/master/CONTRIBUTING.md) applies; please sign the [Contributor License Agreement](https://cla-assistant.io/timescale/timescaledb-parallel-copy) (CLA) if you're a new contributor.
