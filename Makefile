.PHONY: build run test clean

build:
	@echo "Building AutoClickAccepted..."
	@go build -o bin/autoclick.exe ./cmd/main.go

run:
	@go run ./cmd/main.go

test:
	@go test -v ./...

clean:
	@rm -rf bin/
