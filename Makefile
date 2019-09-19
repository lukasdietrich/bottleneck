.PHONY: all install-tools test lint

all: clean test lint

install-tools:
	go get -u golang.org/x/lint/golint
	go get -u github.com/jstemmer/go-junit-report

clean:
	rm -f *.html *.xml *.txt *.log

test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic | tee test.log
	go-junit-report < test.log > junit.xml

lint:
	golint