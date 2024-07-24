SHELL := /bin/bash

init:
	go mod init

tidy:
	go mod tidy

run:
	go run main.go

run-user-service:
	go run user_service/main.go

run-user-service-c:
	go run user_service/main.go -m=consumer

run-gateway-service:
	go run gateway/main.go

build:
	go build

test:
	go test ./...

coverage-report:
	go test -coverprofile cover.out ./handler/...
	go tool cover -html cover.out

proto-gen:
	protoc --go_out=user_service/proto --proto_path=user_service/proto user_service/proto/*.proto --go-grpc_out=user_service/proto
	protoc --go_out=gateway/proto --proto_path=gateway/proto gateway/proto/*.proto --go-grpc_out=gateway/proto