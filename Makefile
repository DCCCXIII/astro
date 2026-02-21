default: build

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

test: vet
	go test -v ./...

test-short: vet
	go test ./...

build: test-short
	go build .
