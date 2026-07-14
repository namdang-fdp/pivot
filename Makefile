.PHONY: fmt vet test test-race lint build check clean

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

vet:
	go vet ./...

test:
	go test ./...

test-race:
	go test -race ./...

lint:
	golangci-lint run

build:
	go build ./cmd/pivot

check:
	test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))"
	go vet ./...
	go test ./...
	go test -race ./...
	golangci-lint run
	go build ./cmd/pivot

clean:
	rm -rf pivot dist coverage.out coverage.html
