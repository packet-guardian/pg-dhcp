.PHONY: all doc fmt alltests test coverage benchmark lint vet dhcp management dist clean docker

all: test

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
