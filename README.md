
# pparallel-bench

`pparallel-bench` is a command line program for parallelizing PostgreSQL's built-in `query` functionality for bulk measuring of performance data with [TimescaleDB.](//github.com/timescale/timescaledb/)
Its main use is to measure performance of queries that run against a database. The main idea is to have one main query and to provide a CSV of data for its execution and to collect metrics on the results.
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
For Windows (untested):
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

The unverbose metrics this tool measures are:

- #of queries processed
- total processing time across all queries
- the minimum query time (for a single query)
- the median query time
- the average query time
- and the maximum query time

The verbose metrics this tool additionally measures are:

Processingstats:

- time spent per batch executed
- Time spent scanning the input
- Time spent reading the responses

Queries:

- #of results received
- avg results per query


### Regular output with 200 Entries, 2 workers and a batchsize of 25 without verbose output

Example:

```
$ ./bin/timescale-parallelquery  -db-name homework -table cpu_usage -skip-header -file data/faulty_query_params.csv -workers 2 -connection  $(DBCONNECTIONSTRING)  -batch-size 25 --query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  %s WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
```

```
########################################
Processing Complete
2021-10-15 12:51:34.5185548 +0000 UTC m=+1.202699401
################
Querystats:
MinimumQueryTime: 8.6977ms
MaximumQueryTime: 18.4664ms
TotalQueryTime: 2.2428712s
MeanQueryTime: 10.725 ms
MedianQueryTime: 11 ms
Queries: 200
########################################
Querying  200 entries, took 1.2019485s with 2 worker(s) (mean rate 89.171416/sec)   
```

### Regular output with 200 Entries, 2 workers and a batchsize of 25 with verbose output

Example: 

```
$ ./bin/timescale-parallelquery  -db-name homework -table cpu_usage -skip-header -file data/faulty_query_params.csv -workers 2 -connection  $(DBCONNECTIONSTRING)  -batch-size 25 -verbose --query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  %s WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"

```

```
Skipping the first 1 lines of the input.
[BATCH] took 308.2699ms, batch size 25, row rate 81.097765/sec
[BATCH] took 309.3971ms, batch size 25, row rate 80.802309/sec
[BATCH] took 281.4304ms, batch size 25, row rate 88.831910/sec
[BATCH] took 281.1626ms, batch size 25, row rate 88.916520/sec
[BATCH] took 257.2092ms, batch size 25, row rate 97.197145/sec
[BATCH] took 268.5397ms, batch size 25, row rate 93.096105/sec
[BATCH] took 256.9481ms, batch size 25, row rate 97.295913/sec
[BATCH] took 260.5247ms, batch size 25, row rate 95.960191/sec
########################################
Processing Complete
2021-10-15 12:52:42.3156183 +0000 UTC m=+1.136168001
################
Processingstats:
Time spent scanning the input: 605.5154ms
Time spent reading the responses: 219.7µs
################
Querystats:
MinimumQueryTime: 8.5291ms
MaximumQueryTime: 18.0183ms
TotalQueryTime: 2.1009662s
MeanQueryTime: 9.99 ms
MedianQueryTime: 10 ms
Queries: 200
Result Rows received: 12163
Avg results per query: 60.815
########################################
Querying  200 entries, took 1.1349651s with 2 worker(s) (mean rate 95.194297/sec)
```

### Problematic output with 200 Entries, 2 workers and a batchsize of 25 with verbose output

Behaviour: When content with errors is provided, the tool will show those in the logs and skip ahead. Error output includes the query and its parameter for easy testing:

```
Skipping the first 1 lines of the input.
[BATCH] took 272.8114ms, batch size 25, row rate 91.638399/sec
[BATCH] took 283.7918ms, batch size 25, row rate 88.092750/sec
Unable to execute query sql: expected 3 arguments, got 2
for Query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  "public"."cpu_usage" WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
with parameter [host_000006 ]
Unable to execute query pq: invalid input syntax for type timestamp with time zone: "2017-01-01 23:as:52"
for Query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  "public"."cpu_usage" WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
with parameter [host_000003 2017-01-01 22:05:52 2017-01-01 23:as:52]
[BATCH] took 239.2407ms, batch size 25, row rate 104.497270/sec
[BATCH] took 252.8844ms, batch size 25, row rate 98.859400/sec
Unable to execute query pq: invalid input syntax for type timestamp with time zone: "2017-01-01 17:05:x"
for Query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  "public"."cpu_usage" WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
with parameter [host_000001 2017-01-01 17:05:x 2017-01-01 18:05:12]
Unable to execute query pq: invalid input syntax for type timestamp with time zone: "2017-01-a 14:29:53"
for Query "SELECT time_bucket('1 minutes', ts) AS t, min(usage) AS min_cpu, max(usage) AS max_cpu,avg(usage) AS avg_cpu FROM  "public"."cpu_usage" WHERE host = $1 AND ts > $2 AND ts < $3 GROUP BY t ORDER BY t DESC;"
with parameter [host_000002 2017-01-02 13:29:53 2017-01-a 14:29:53]
[BATCH] took 255.2225ms, batch size 25, row rate 97.953746/sec
[BATCH] took 254.363ms, batch size 25, row rate 98.284735/sec
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
[BATCH] took 288.4073ms, batch size 25, row rate 86.682965/sec
[BATCH] took 450.528ms, batch size 25, row rate 55.490447/sec
########################################
Processing Complete
2021-10-15 12:53:54.2585477 +0000 UTC m=+1.251625401
################
Processingstats:
Time spent scanning the input: 545.5381ms
Time spent reading the responses: 169.3µs
################
Querystats:
MinimumQueryTime: 8.3168ms
MaximumQueryTime: 191.1425ms
TotalQueryTime: 2.1468537s
MeanQueryTime: 10.651041666666666 ms
MedianQueryTime: 10 ms
Queries: 192
Result Rows received: 13354
Avg results per query: 69.55208333333333
########################################
Querying  200 entries, took 1.2507416s with 2 worker(s) (mean rate 93.159585/sec)
########################################
            WARNING!
########################################
Faulty Queries: 8 
There has been more input found than queries executed,
please consult the Log for Errors!!
########################################
```

### Retrieving Results

This tool also provides the possibility to save all results in a json file for reference. In ord to use this please utilize the --toJson flag with a file path of your chosing.

NOTE: As mentioned earlier, all results are grouped by the first parameter in your inputcsv. this ideally should be the first parameter in your where clause. This enables for instance that you could query for multiple hostnames in your datbase and later summarize them for a more detailed analysis.

```
$ .... -toJson ./resultdata.json
```

