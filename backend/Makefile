.PHONY: run build test vet tidy clean

run: ## Run the API locally
	go run ./cmd

build: ## Build a production binary into ./bin
	mkdir -p bin
	go build -o bin/goinvoicer ./cmd

test: ## Run tests
	go test ./...

vet: ## Static checks
	go vet ./...

tidy: ## Clean up deps
	go mod tidy

clean: ## Remove build output
	rm -rf bin
