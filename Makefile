.PHONY: all test lint

all: test lint

test:
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out -o=coverage.html

lint:
	golint