# Test Makefile for Expense Tracker API

.PHONY: test test-unit test-integration test-coverage test-verbose clean

# Run all tests
test:
	go test ./tests/unit/... ./tests/integration/... -v

# Run only unit tests
test-unit:
	go test ./tests/unit/... -v

# Run only integration tests  
test-integration:
	go test ./tests/integration/... -v

# Run tests with coverage
test-coverage:
	go test ./tests/unit/... ./tests/integration/... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with verbose output and race detection
test-verbose:
	go test ./tests/unit/... ./tests/integration/... -v -race -timeout=30s

# Run specific test
test-specific:
	@read -p "Enter test name: " test; \
	go test ./tests/unit/... ./tests/integration/... -v -run $$test

# Clean test artifacts
clean:
	rm -f coverage.out coverage.html
	go clean -testcache

# Install test dependencies
deps:
	go mod tidy
	go get github.com/stretchr/testify/assert
	go get github.com/stretchr/testify/suite
	go get github.com/DATA-DOG/go-sqlmock

# Run tests in watch mode (requires entr)
test-watch:
	find . -name "*.go" | entr -r make test-unit

# Benchmark tests
benchmark:
	go test ./tests/unit/... -bench=. -benchmem

# Generate test report
report: test-coverage
	@echo "=== Test Summary ==="
	@go test ./tests/unit/... ./tests/integration/... -v | grep -E "(PASS|FAIL|RUN)"
	@echo "=== Coverage Report ==="
	@go tool cover -func=coverage.out