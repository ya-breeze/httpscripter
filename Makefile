all: lint test

# Lazy way to install ginkgo
/go/bin/ginkgo:
	go mod tidy
	go install github.com/onsi/ginkgo/v2/ginkgo
	@echo "✅ ginkgo installed"

.PHONY: test
test: /go/bin/ginkgo
	/go/bin/ginkgo -r -v
	@echo "✅ Tests passed"

/go/bin/golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint
	@echo "✅ golangci-lint installed"

lint: /go/bin/golangci-lint
	/go/bin/golangci-lint run
	@echo "✅ Lint passed"
