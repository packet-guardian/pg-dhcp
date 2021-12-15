.PHONE: doc lint test vet full generate

default: lint vet test

doc:
	godoc -http=:6060 -index

lint:
	golint ./...

test:
	go test -v ./...

vet:
	go vet ./...

full: lint vet test
