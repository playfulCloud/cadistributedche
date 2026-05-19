APP_NAME=cadistributedche
MAIN_PATH=./cmd/cadistributedche
BIN_DIR=bin

run:
	go run $(MAIN_PATH)

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(MAIN_PATH)

test:
	gotestsum ./...

race:
	go test -race ./...

test-v:
	go test -v ./...

coverage:
	go test -cover ./...

lint:
	golangci-lint run

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR)

check: fmt vet test

.PHONY: run build test test-v coverage fmt vet tidy clean check
