.DEFAULT_GOAL := help
.PHONY : check lint install-linters dep test build

VERSION := $(shell git describe --always)

BIN := ${PWD}/bin
OPTS?=GO111MODULE=on
BIN_DIR?=./bin

TEST_OPTS:=-tags no_ci -cover -timeout=5m

RACE_FLAG:=-race
GOARCH:=$(shell go env GOARCH)

ifneq (,$(findstring 64,$(GOARCH)))
    TEST_OPTS:=$(TEST_OPTS) $(RACE_FLAG)
endif

BUILD_OPTS?="-ldflags=$(BUILDINFO)"
BUILD_OPTS_DEPLOY?="-ldflags=$(BUILDINFO) -w -s"

build: ## Build binaries into ./bin
	mkdir -p ${BIN}; go build ${BUILD_OPTS} -o ${BIN} ./cmd/*

daemon: build ## Build and register dmsg-daemon.service
	sudo cp ${BIN}/dmsg-daemon /usr/local/bin/
	awk '{sub("CSVPATH","${PWD}/dmsg-clients.csv")}1' ./scripts/dmsg-daemon.service > temp.txt && mv temp.txt ./dmsg-daemon.service
	sudo mv ${PWD}/dmsg-daemon.service /etc/systemd/system/
	sudo systemctl daemon-reload

deamon-start-on-reboot:
	sudo systemctl enable dmsg-daemon

daemon-start:
	sudo systemctl start dmsg-daemon

daemon-stop:
	sudo systemctl stop dmsg-daemon