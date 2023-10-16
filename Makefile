NAME := pg-dhcp
DESC := DHCP server
VERSION ?= $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
BUILDTIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDER ?= $(shell echo "`git config user.name` <`git config user.email`>")
CGO_ENABLED := 0

PWD := $(shell pwd)
GOBIN := $(PWD)/bin
UID := $(shell id -u)
GID := $(shell id -g)

ifeq ($(shell uname -o), Cygwin)
PWD := $(shell cygpath -w -a `pwd`)
GOBIN := $(PWD)\bin
endif

LDFLAGS := -X 'main.version=$(VERSION)' \
			-X 'main.buildTime=$(BUILDTIME)' \
			-X 'main.builder=$(BUILDER)' \
			-X 'main.goversion=$(GOVERSION)'

.PHONY: all doc fmt alltests test coverage benchmark lint vet dhcp management dist clean docker build build-cmd build-tools build-cli build-db-edit build-db-migrate

all: test build

build:
	docker run \
		--rm \
		-v "$(PWD)":/usr/src/myapp \
		-w /usr/src/myapp \
		-e XDG_CACHE_HOME=/tmp/.cache \
		-e "BUILDER=$(BUILDER)" \
		-e "VERSION=$(VERSION)" \
		-e "BUILDTIME=$(BUILDTIME)" \
		docker.io/golang:1.20-bullseye \
		make build-cmd

build-tools:
	docker run \
		--rm \
		-v "$(PWD)":/usr/src/myapp \
		-w /usr/src/myapp \
		-e XDG_CACHE_HOME=/tmp/.cache \
		-e "BUILDER=$(BUILDER)" \
		-e "VERSION=$(VERSION)" \
		-e "BUILDTIME=$(BUILDTIME)" \
		docker.io/golang:1.20-bullseye \
		make build-cli build-db-edit build-db-migrate

build-cmd:
	go build -o bin/dhcp -v -ldflags "$(LDFLAGS)" -tags '$(BUILDTAGS)' ./cmd/dhcp/...

build-cli:
	go build -o bin/dhcp-cli -v -ldflags "$(LDFLAGS)" -tags '$(BUILDTAGS)' ./cmd/cli/...

build-db-edit:
	go build -o bin/dhcp-edit -v -ldflags "$(LDFLAGS)" -tags '$(BUILDTAGS)' ./cmd/db-edit/...

build-db-migrate:
	go build -o bin/dhcp-migrate -v -ldflags "$(LDFLAGS)" -tags '$(BUILDTAGS)' ./cmd/db-migrate/...

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

integration-test:
	@go test -race -tags mysql $$(go list ./... | grep -v 'vendor/')

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
