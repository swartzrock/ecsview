BINARY_NAME=bin/ecsview

linter := golangci-lint

all: build

build:
	@go build -o $(BINARY_NAME)

clean:
	@go clean
	@rm -f $(BINARY_NAME)

run: lint build
	@./$(BINARY_NAME)

lint:
	@goimports -local=github.com/swartzrock/ecsview -w .
	@gofmt -s -w .
	@$(linter) run

install: lint build
	@go install
	@echo "Installed ecsview into ${shell which ecsview}"
