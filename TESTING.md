# Test configuration for AWS SSM

## Running Tests

### Unit Tests
```bash
# Run all unit tests
go test -v ./...

# Run short tests only (skips integration tests)
go test -short -v ./...

# Run tests with race detection
go test -race -short ./...

# Using Makefile
make test
make test-short
make test-race
```

### Integration Tests
Integration tests require AWS credentials and can run against real AWS services.

```bash
# Set up test environment
export AWS_INTEGRATION_TEST_REGION="us-east-1"
export AWS_INTEGRATION_TEST_SECRET="your-test-secret-name"

# Run integration tests
make test-integration

# Or directly with go
AWS_SKIP_INTEGRATION_TESTS=false go test -v ./...
```

### Coverage
```bash
# Generate coverage report
make test-coverage

# This creates:
# - coverage.out (coverage data)
# - coverage.html (HTML report)
```

### Benchmarks
```bash
# Run benchmarks
make benchmark

# Or directly
go test -bench=. -benchmem ./...
```

## Test Categories

### Unit Tests (`main_test.go`)
- **ParseArgs**: Tests command line argument parsing
- **FormatOutput**: Tests output formatting (JSON detection)
- **App.GetSecret**: Tests secret retrieval with mocked AWS client

### Integration Tests (`integration_test.go`)
- **E2E with Mock AWS**: End-to-end testing with mocked responses
- **Real AWS Tests**: Tests against actual AWS (requires credentials)
- **Concurrent Access**: Tests thread safety
- **Error Handling**: Tests various error scenarios
- **Benchmarks**: Performance testing

## Test Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `AWS_SKIP_INTEGRATION_TESTS` | Skip integration tests (`true`/`false`) | No |
| `AWS_INTEGRATION_TEST_REGION` | AWS region for integration tests | For integration tests |
| `AWS_INTEGRATION_TEST_SECRET` | Test secret name/ARN | For integration tests |

## CI/CD Testing

### GitHub Actions
The project includes comprehensive CI/CD testing:

- **Multi-OS Testing**: Tests on Ubuntu, Windows, macOS
- **Multi-Go Version**: Tests with Go 1.21 and 1.22
- **Coverage Reports**: Generates and uploads coverage reports
- **Security Scanning**: Runs gosec for security analysis
- **Linting**: Uses golangci-lint for code quality

### Local Development
```bash
# Run all development checks
make dev

# Run CI equivalent locally
make ci-test

# Install development tools
make install-tools
```

## Test Structure

```
.
├── main.go              # Main application code
├── main_test.go         # Unit tests
├── integration_test.go  # Integration and E2E tests
├── Makefile            # Development commands
└── .github/workflows/
    ├── test.yml        # Comprehensive test workflow
    └── build.yml       # Build workflow (includes basic tests)
```

## Mock Testing

The project uses `testify/mock` for creating mock AWS clients:

```go
// Example mock setup
mockClient := new(MockSecretsManagerClient)
mockClient.On("GetSecretValue", ctx, input).Return(output, nil)

app := &App{Client: mockClient, Config: cfg}
result, err := app.GetSecret(ctx)
```

## Coverage Goals

- **Target**: >90% code coverage
- **Current focus**: Business logic functions
- **Exclusions**: Main function, error paths that require real AWS

## Best Practices

1. **Test Isolation**: Each test is independent
2. **Mock External Dependencies**: AWS clients are mocked
3. **Test Data**: Use predictable test data
4. **Error Testing**: Test both success and failure paths
5. **Performance**: Include benchmarks for critical paths