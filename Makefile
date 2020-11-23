FILE_HASH?=$(shell git ls-files | xargs sha256sum | cut -d" " -f1 | sha256sum | cut -d" " -f1)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

build:
	@echo "-- building binary"
	go build -ldflags "-X main.buildHash=${FILE_HASH} -X main.buildTime=${BUILD_TIME}" -o ./bin/bot ./cmd

lint:
	@echo "-- linter running"
	golangci-lint run -c .golangci.yaml ./pkg...
	golangci-lint run -c .golangci.yaml ./cmd...