# 单元测试
.PHONY: ut
ut:
	@go test -race ./...

.PHONY: setup
setup:
	@sh ./scripts/setup.sh

.PHONY: lint
lint:
	make -C app/article-service lint

.PHONY: fmt
fmt:
	@sh ./scripts/fmt.sh

.PHONY: tidy
tidy:
	@go mod tidy -v

.PHONY: check
check:
	@$(MAKE) --no-print-directory fmt
	@$(MAKE) --no-print-directory tidy

# e2e 测试
.PHONY: e2e
e2e:
	make -C app/article-service e2e_test

