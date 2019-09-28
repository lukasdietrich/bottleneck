.PHONY: all install-tools test lint

all: clean test lint

install-tools:
	GO111MODULE=on go get \
		github.com/golangci/golangci-lint/cmd/golangci-lint@v1.19.1 \
		github.com/jstemmer/go-junit-report
	go mod tidy

clean:
	rm -f *.html *.xml *.txt *.log

test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./... | tee test.log
	go-junit-report < test.log > junit.xml

lint:
	golangci-lint run