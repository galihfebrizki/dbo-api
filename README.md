# DPO SERVICE API
DPO Service API

## How to run all system
```
docker compose up
```

## How to run
- make sure you have setup .env file in project folder
```
go run ./cmd/app/

```

## How to build
```
go build -o bca-server ./cmd/app/
```

## How to use google wire
- install google wire
- add new handler, service and repository in internal folder
- add new set to wire.go file
- run :
```
wire ./...
```
