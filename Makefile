.PHONY: all install-tools test lint

all: test lint

install-tools:
	go get -u golang.org/x/lint/golint
	go get -u github.com/jstemmer/go-junit-report
	go get -u github.com/t-yuki/gocover-cobertura

test:
	mkdir -p reports
	go test -v -coverprofile=reports/coverage.out -covermode=count > reports/test.log
	go-junit-report < reports/test.log > reports/junit.xml
	gocover-cobertura < reports/coverage.out > reports/coverage.xml

lint:
	golint