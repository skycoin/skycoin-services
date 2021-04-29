# Basic go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Binary names
BINARY_NAME=manifest
MAIN=manifest.go
CUSTOMEtYPES=manifestTypes.go
MANIFESTUTIL=manifestUtil.go

# CLI command and flags
INIT=init
COMMIT=commit
PRINTJSON=-print-json
META=-meta

run: init  commit

build: 
		$(GOBUILD) -o $(BINARY_NAME) -v
test: 
		$(GOTEST) -v ./...
clean: 
		$(GOCLEAN)
		rm -f $(BINARY_NAME)
init:
		$(GORUN) $(MAIN) $(CUSTOMEtYPES) $(MANIFESTUTIL) $(INIT) 

commit:
		$(GORUN) $(MAIN) $(CUSTOMEtYPES) $(MANIFESTUTIL) $(COMMIT) 

commit-json:
		$(GORUN) $(MAIN) $(CUSTOMEtYPES) $(MANIFESTUTIL) $(COMMIT) $(PRINTJSON) 

commit-meta:
		$(GORUN) $(MAIN) $(CUSTOMEtYPES) $(MANIFESTUTIL) $(COMMIT) $(PRINTJSON) $(META) 
 