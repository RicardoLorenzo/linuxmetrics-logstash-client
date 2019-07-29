EXE = linuxmetrics-logstash
PACKAGE = github.com/RicardoLorenzo/linuxmetrics-logstash-client
BASE 	= $(GOPATH)/src/$(PACKAGE)
VERSION = 0.1.00

.PHONY : all clean fmt test test-junit build

all : fmt test build

build : test
	@go get -d ./...
	@GOOS=darwin GOARCH=amd64 go build  -o $(GOPATH)/bin/$(EXE)-$(VERSION)-osx $(PACKAGE)
	@GOOS=linux GOARCH=amd64 go build  -o $(GOPATH)/bin/$(EXE)-$(VERSION)-linux $(PACKAGE)

test : fmt
	@go get -d ./...
	# TODO: A test is failing on dependency
	#@go test -v -cover ./...

fmt :
	@gofmt -w $(BASE)/*.go
