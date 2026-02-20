default: build

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build .
