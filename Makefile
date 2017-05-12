NAME := packet-guardian
DESC := A captive portal for today's networks
VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
CGO_ENABLED ?= 0

PWD := $(shell pwd)
GOBIN := $(PWD)/bin
CODECLIMATE_CODE := $(PWD)

ifeq ($(shell uname -o), Cygwin)
CODECLIMATE_CODE := //c/cygwin64$(PWD)
PWD := $(shell cygpath -w -a `pwd`)
GOBIN := $(PWD)\bin
endif

BUILDTAGS ?= dball
LDFLAGS := -X 'main.version=$(VERSION)' \
			-X 'main.buildTime=$(BUILDTIME)' \
			-X 'main.builder=$(BUILDER)' \
			-X 'main.goversion=$(GOVERSION)'

.PHONY: all doc fmt alltests test coverage benchmark lint vet dhcp management dist clean docker server

all: test server

server:
	GOBIN="$(GOBIN)" go install -v -ldflags "$(LDFLAGS)" -tags '$(BUILDTAGS)' ./cmd/dhcp

# development tasks
doc:
	@godoc -http=:6060 -index

fmt:
	@go fmt $$(go list ./... | grep -v 'vendor/')

alltests: test lint vet

test:
ifdef verbose
	@go test -race -v $$(go list ./... | grep -v 'vendor/')
else
	@go test -race $$(go list ./... | grep -v 'vendor/')
endif

coverage:
	@go test -cover $$(go list ./... | grep -v 'vendor/')

benchmark:
	@echo "Running tests..."
	@go test -bench=. $$(go list ./... | grep -v 'vendor/')

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
lint:
	@golint $$(go list ./... | grep -v 'vendor/')

vet:
	@go vet $$(go list ./... | grep -v 'vendor/')
