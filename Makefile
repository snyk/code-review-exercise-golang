APP?=npmjs-deps-fetcher
GITHUB_SHA?=dev

.DEFAULT_GOAL := all

.PHONY: all
all: mod gen fmt build lint test

.PHONY: build
build: ## Build the app Go binary
	$(call print-target)
	go build -o .bin/ ./cmd/${APP}/...

.PHONY: clean
clean: ## Cleanup artifacts of the build pipeline
	$(call print-target)
	rm -f test/results/*
	go tool golangci-lint cache clean
	go clean -i -cache -testcache -modcache -fuzzcache -x

.PHONY: docker-build
docker-build: ## Build the docker image for the service
	$(call print-target)
	docker build --build-arg APP=${APP} -t ${APP}:${GITHUB_SHA} .

.PHONY: docker-run
docker-run: docker-build ## Run the docker image for the service
	$(call print-target)
	docker run -t -p 8080:8080 ${APP}:${GITHUB_SHA}

.PHONY: download
download: ## Download dependencies to local cache
	$(call print-target)
	go mod download

.PHONY: fmt
fmt: ## Format source code based on golangci
	$(call print-target)
	go tool golangci-lint run --fix -v ./...

.PHONY: gen
gen: ## Code generation
	$(call print-target)
	go generate ./...

.PHONY: help
help: ## List Makefile targets
	@echo
	@printf "\033[32m[ Makefile Targets ]\033[0m\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}'
	@echo

.PHONY: lint
lint: ## Lint using golangci-lint
	$(call print-target)
	go tool golangci-lint run -v ./...

.PHONY: mod
mod: ## Add missing or remove unused modules from go.mod
	$(call print-target)
	go mod tidy

.PHONY: run
run: ## Run service
	$(call print-target)
	go run ./cmd/${APP}/...

.PHONY: test
test: ## Run unit tests
	$(call print-target)
	mkdir -p test/results
	go tool gotestsum --junitfile test/results/unit-tests.xml -- -race -covermode=atomic -coverprofile=test/results/cover.out -v ./...
	go tool cover -html=test/results/cover.out -o test/results/coverage.html

define print-target
	@echo
    @printf "\033[32m*\033[0m Executing target: \033[36m$@\033[0m\n"
endef
