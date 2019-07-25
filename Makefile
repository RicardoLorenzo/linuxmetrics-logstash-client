EXE = linuxmetrics-logstash
PACKAGE = github.com/RicardoLorenzo/linuxmetrics-logstash
BASE 	= $(GOPATH)/src/$(PACKAGE)
VERSION = 0.1.00

.PHONY : all clean fmt test test-junit build

all : fmt test build

build : test
	@go get github.com/c9s/goprocinfo/linux
	@GOOS=darwin GOARCH=amd64 go build  -o $(GOPATH)/bin/$(EXE)-$(VERSION)-osx $(PACKAGE)
	@GOOS=linux GOARCH=amd64 go build  -o $(GOPATH)/bin/$(EXE)-$(VERSION)-linux $(PACKAGE)

test : fmt
	@go test -v -cover ./...

fmt :
	@gofmt -w $(BASE)/*.go
