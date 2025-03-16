# Agent SDK Go Tests

This directory contains the tests for the Agent SDK Go. The tests are organized by package and functionality.

## Test Organization

- **agent**: Tests for the agent package functionality
- **model**: Tests for the model package functionality
- **runner**: Tests for the runner package functionality
- **tool**: Tests for the tool package functionality
- **tracing**: Tests for the tracing package functionality
- **integration**: End-to-end integration tests that test multiple components together

## Running Tests

To run all tests, use:

```bash
go test -v ./test/...
```

To run tests for a specific package, use:

```bash
go test -v ./test/agent/...
go test -v ./test/model/...
# etc.
```

## Integration Tests

The integration tests are particularly useful for testing the full functionality of the SDK. They test multiple components together to ensure that they work correctly in combination.

The integration tests include:

- **simple_agent_test.go**: Tests basic agent functionality with different tools and output types
- **multi_agent_test.go**: Tests the handoff functionality between multiple agents

## Test Coverage

To run the tests with coverage, use:

```bash
go test -v -cover ./test/...
```

To generate a coverage report, use:

```bash
go test -v -coverprofile=coverage.out ./test/...
go tool cover -html=coverage.out
```

## Writing New Tests

When writing new tests, please follow these guidelines:

1. Place tests in the appropriate directory based on what they're testing
2. Use descriptive test names that indicate what functionality is being tested
3. Use the standard Go testing pattern of `func TestXxx(t *testing.T)`
4. Include comments that describe what the test is checking
5. Include setup, execution, and validation sections in each test
6. Use helper functions for common setup/teardown code
7. For integration tests, ensure they test realistic scenarios 