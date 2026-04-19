.PHONY: build build-gui run test clean

build:
	@echo "Building AutoClickAccepted (Console mode)..."
	@go build -o autoclick.exe ./cmd/main.go

build-gui:
	@echo "Building AutoClickAccepted (GUI mode - no console window)..."
	@go build -ldflags "-H windowsgui" -o autoclick_gui.exe ./cmd/main.go

run:
	@go run ./cmd/main.go

test:
	@go test -v ./...

clean:
	@rm -f autoclick.exe autoclick_gui.exe
