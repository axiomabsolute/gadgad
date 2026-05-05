.PHONY: setup build test lint

setup:
	pre-commit install

build:
	go build ./...

test:
	go test ./...

lint:
	golangci-lint run ./...
