BINARY := jrnl-md
MODULE := github.com/glw907/jrnl-md
VERSION := 2.0.0

.PHONY: build test vet lint install check clean

build:
	go build -o $(BINARY) ./cmd/jrnl-md

test:
	go test ./...

vet:
	go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

install:
	go install ./cmd/jrnl-md

check: vet test

clean:
	rm -f $(BINARY)
