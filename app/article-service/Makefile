
.PHONY: e2e_test
e2e_test: ## 单元测试
	@sh ./scripts/integrate_test.sh

.PHONY: lint
lint:
	go mod tidy
	golangci-lint run -v --timeout=5m
