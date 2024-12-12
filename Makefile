.DEFAULT_GOAL := all
.PHONY: all
all: ## Build pipeline
all: mod inst gen fmt build lint test

.PHONY: precommit
precommit: ## Validate the branch before commit
precommit: all vuln

.PHONY: ci
ci: ## CI Build pipeline
ci: precommit diff

.PHONY: clean
clean: ## Cleanup artifacts of the build pipeline
	$(call print-target)
	rm -f coverage.*
	rm -f '"$(shell go env GOCACHE)/../golangci-lint"'
	go clean -i -cache -testcache -modcache -fuzzcache -x

.PHONY: mod
mod: ## Add missing or remove unused modules from go.mod
	$(call print-target)
	go mod tidy

PHONY: inst
inst: ## Install tools
	$(call print-target)
	cd tools && GOBIN=$(shell pwd)/tools/bin go install $(shell cd tools && go list -e -f '{{ join .Imports " " }}' -tags=tools)

.PHONY: fmt
fmt: ## Format code
	go fmt ./...

.PHONY: gen
gen: ## Code generation
	$(call print-target)
	go generate ./...

.PHONY: build
build: ## Build
build: fmt gen
	go build -o bin/

.PHONY: lint
lint: ## Lint and (attempt) fix (golangci-lint)
	$(call print-target)
	tools/bin/golangci-lint run --fix

.PHONY: vuln
vuln: ## Look for vulnerabilities (https://vuln.go.dev/)
	$(call print-target)
	tools/bin/govulncheck ./...

.PHONY: test
test: ## Test
	$(call print-target)
	go test -race -covermode=atomic -coverprofile=coverage.out -coverpkg=./... ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: diff
diff: ## Fail if branch isn't clean (i.e. git diff isn't empty)
	$(call print-target)
	git diff --exit-code
	RES=$$(git status --porcelain) ; if [ -n "$$RES" ]; then echo $$RES && exit 1 ; fi

.PHONY: help
help: ## Shows this list of Makefile targets
	@echo
	@printf "\033[32m[ Makefile Targets ]\033[0m\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}'

define print-target
	@echo
    @printf "\033[32m*\033[0m Executing target: \033[36m$@\033[0m\n"
endef
