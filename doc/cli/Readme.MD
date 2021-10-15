# Cli usage tutorial

## building pparallel-bench

For Linux:
```
$ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/pparallel-bench ./cmd/pparallel-bench/main.go
```
For Windows:
```
$ CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/pparallel-bench ./cmd/pparallel-bench/main.go
```
## installing the binary 

```
go install github.com/jhoelzel/pparallel-bench
```