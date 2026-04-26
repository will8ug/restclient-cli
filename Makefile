.PHONY: build test vet lint fmt clean run install

BINARY := restclient-cli
BUILD_DIR := ./bin
CMD_DIR := ./cmd

GOPATH := $(shell go env GOPATH)
GOLANGCI_LINT := $(GOPATH)/bin/golangci-lint

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) $(CMD_DIR)

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

lint:
	@if [ -x "$(GOLANGCI_LINT)" ]; then \
		$(GOLANGCI_LINT) run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

fmt:
	go fmt ./...

clean:
	rm -rf $(BUILD_DIR)

run: build
	$(BUILD_DIR)/$(BINARY) examples/jsonplaceholder.http

install:
	go install $(CMD_DIR)
