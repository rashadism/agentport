.PHONY: build run run-local test clean vet fmt

# Build the application
build:
	go build -o fantasydemo ./cmd/fantasydemo

# Run the application
run: build
	./fantasydemo

# Run locally with .env file
run-local:
	@export $$(cat .env | xargs) && go run ./cmd/fantasydemo

# Run tests
test:
	go test -v ./...

# Run go vet
vet:
	go vet ./...

# Format code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	rm -f fantasydemo
	go clean

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/fantasydemo-linux-amd64 ./cmd/fantasydemo
	GOOS=linux GOARCH=arm64 go build -o bin/fantasydemo-linux-arm64 ./cmd/fantasydemo
	GOOS=darwin GOARCH=amd64 go build -o bin/fantasydemo-darwin-amd64 ./cmd/fantasydemo
	GOOS=darwin GOARCH=arm64 go build -o bin/fantasydemo-darwin-arm64 ./cmd/fantasydemo
