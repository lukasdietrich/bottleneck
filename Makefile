.PHONY: all install-tools test lint

all: test lint

install-tools:
	go get -u github.com/jstemmer/go-junit-report
	go get -u github.com/axw/gocov/gocov
	go get -u github.com/AlekSi/gocov-xml
	go get -u golang.org/x/lint/golint
	go get -u github.com/matm/gocov-html

test:
	mkdir -p reports
	go test -v -coverprofile=reports/coverage.out -covermode=count > reports/test.log
	go-junit-report < reports/test.log > reports/junit.xml
	gocov convert reports/coverage.out > reports/coverage.json
	gocov-xml < reports/coverage.json > reports/coverage.xml
	gocov-html < reports/coverage.json > reports/coverage.html

lint:
	golint