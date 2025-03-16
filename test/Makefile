# Makefile for running Agent SDK Go tests

# Test variables
TEST_PACKAGE ?= ./...
COVERAGE_FILE ?= coverage.out

# Default rule - run all tests
.PHONY: test
test:
	go test -v $(TEST_PACKAGE)

# Run tests with coverage
.PHONY: test-cover
test-cover:
	go test -v -coverprofile=$(COVERAGE_FILE) $(TEST_PACKAGE)
	go tool cover -func=$(COVERAGE_FILE)

# Generate an HTML coverage report
.PHONY: cover-html
cover-html: test-cover
	go tool cover -html=$(COVERAGE_FILE)

# Test a specific package
.PHONY: test-agent
test-agent:
	go test -v ./agent/...

.PHONY: test-model
test-model:
	go test -v ./model/...

.PHONY: test-runner
test-runner:
	go test -v ./runner/...

.PHONY: test-tool
test-tool:
	go test -v ./tool/...

.PHONY: test-tracing
test-tracing:
	go test -v ./tracing/...

.PHONY: test-integration
test-integration:
	go test -v ./integration/...

# Run all tests with verbose output and race detection
.PHONY: test-race
test-race:
	go test -v -race $(TEST_PACKAGE)

# Clean up test artifacts
.PHONY: clean
clean:
	rm -f $(COVERAGE_FILE)

# CI test target with buildvcs=false flag
ci-test:
	go test -buildvcs=false -v $(TEST_PACKAGE)

# CI test with race detection
ci-test-race:
	go test -buildvcs=false -race -v $(TEST_PACKAGE)

# CI test with coverage
ci-test-coverage:
	go test -buildvcs=false -v -coverprofile=coverage.out $(TEST_PACKAGE)
	go tool cover -func=coverage.out 