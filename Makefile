.DEFAULT_GOAL := help
.PHONY : check lint install-linters dep test build

VERSION := $(shell git describe --always)

OPTS?=GO111MODULE=on

TEST_OPTS:=-tags no_ci -cover -timeout=5m

RACE_FLAG:=-race
GOARCH:=$(shell go env GOARCH)

ifneq (,$(findstring 64,$(GOARCH)))
    TEST_OPTS:=$(TEST_OPTS) $(RACE_FLAG)
endif

test: ## Run tests
	-go clean -testcache &>/dev/null
	${OPTS} go test ${TEST_OPTS} ./...

build-vpn-client:
	./build.sh
