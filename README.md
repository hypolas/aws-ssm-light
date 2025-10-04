# AWS SSM

[![Build and Release](https://github.com/hypolas/aws-ssm-light/actions/workflows/build.yml/badge.svg)](https://github.com/hypolas/aws-ssm-light/actions/workflows/build.yml)
[![Renovate](https://github.com/hypolas/aws-ssm-light/actions/workflows/renovate.yml/badge.svg)](https://github.com/hypolas/aws-ssm-light/actions/workflows/renovate.yml)
[![Latest Release](https://img.shields.io/github/v/release/hypolas/aws-ssm-light)](https://github.com/hypolas/aws-ssm-light/releases/latest)

**Tags:** `aws-secrets-manager` `docker` `golang` `cli-tool` `kubernetes` `microservices` `aws-cli-alternative`

A lightweight Go program that replaces `aws secretsmanager get-secret-value` to retrieve secrets from AWS Secrets Manager.

## Why AWS SSM?

This binary was specifically created to **avoid the complete and heavy installation of AWS CLI in Docker images** while maintaining the ability to retrieve secrets from AWS Secrets Manager efficiently.

### The Problem
- **AWS CLI is heavy**: >100MB download + dependencies
- **Slow container startup**: Large image size impacts deployment speed
- **Security surface**: More components = more potential vulnerabilities
- **Resource usage**: High memory footprint for simple secret retrieval

### The Solution
- **Lightweight binary**: ~10MB vs >100MB for AWS CLI
- **Single purpose**: Only secret retrieval, no bloat
- **Fast startup**: Minimal memory footprint (~5MB vs ~50MB)
- **Docker optimized**: Perfect for container environments
- **Drop-in replacement**: Compatible output format

### Use Cases
- 🐳 **Docker containers** requiring AWS secrets
- 🚀 **CI/CD pipelines** needing lightweight secret access
- 📦 **Kubernetes pods** with minimal resource requirements  
- 🔧 **Microservices** requiring fast startup times
- 🏗️ **Build environments** where AWS CLI is overkill

## Features

- ✅ Retrieves secrets from AWS Secrets Manager
- ✅ Compatible with standard AWS credentials (environment variables, IAM roles, etc.)
- ✅ Lighter than AWS CLI (binary ~10MB vs ~100MB+)
- ✅ Output compatible with `--query SecretString --output text`
- ✅ Multi-platform support (Windows, Linux, macOS on AMD64 and ARM64)
- ✅ Automatic releases with checksums for security verification
- ✅ Cross-region secret access support

## Installation

### Quick Install (Recommended)

#### Linux (AMD64)
```bash
curl -L https://github.com/hypolas/aws-ssm-light/releases/latest/download/aws-ssm-linux-amd64 -o aws-ssm && \
chmod +x aws-ssm && \
sudo mv aws-ssm /usr/local/bin/
```

#### macOS (Apple Silicon)
```bash
curl -L https://github.com/hypolas/aws-ssm-light/releases/latest/download/aws-ssm-darwin-arm64 -o aws-ssm && \
chmod +x aws-ssm && \
sudo mv aws-ssm /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -L https://github.com/hypolas/aws-ssm-light/releases/latest/download/aws-ssm-darwin-amd64 -o aws-ssm && \
chmod +x aws-ssm && \
sudo mv aws-ssm /usr/local/bin/
```

#### Windows (PowerShell)
```powershell
Invoke-WebRequest -Uri "https://github.com/hypolas/aws-ssm-light/releases/latest/download/aws-ssm-windows-amd64.exe" -OutFile "aws-ssm.exe"
# Move to a directory in your PATH
```

### Manual Download

Download the appropriate binary for your platform from the [releases page](https://github.com/hypolas/aws-ssm-light/releases/latest):

- **Windows (Intel/AMD)**: `aws-ssm-windows-amd64.exe`
- **Windows (ARM)**: `aws-ssm-windows-arm64.exe`
- **Linux (Intel/AMD)**: `aws-ssm-linux-amd64`
- **Linux (ARM)**: `aws-ssm-linux-arm64`
- **macOS (Intel)**: `aws-ssm-darwin-amd64`
- **macOS (Apple Silicon)**: `aws-ssm-darwin-arm64`

## Usage

```bash
# Basic usage
aws-ssm <secret-id>

# With specific region
aws-ssm <secret-id> <region>

# Show version information
aws-ssm --version

# Show help
aws-ssm --help
```

## Command Line Examples

```bash
# Retrieve a secret by name (uses default region from AWS_REGION)
aws-ssm "azure-devops-token"

# Retrieve a secret with explicit region
aws-ssm "azure-devops-token" "eu-west-1"

# Retrieve a secret using full ARN
aws-ssm "arn:aws:secretsmanager:eu-west-1:123456789012:secret:azure-devops-token-AbCdEf"

# Use in a script to store secret in a variable
SECRET=$(aws-ssm "my-database-password")
echo "Retrieved secret: $SECRET"

# Use with different regions
aws-ssm "prod-api-key" "us-east-1"
aws-ssm "dev-database-url" "eu-central-1"

# Example replacing AWS CLI command:
# Instead of: aws secretsmanager get-secret-value --secret-id "my-secret" --query SecretString --output text
# Use: aws-ssm "my-secret"
```

## Environment Variables

- `AWS_REGION`: Default AWS region
- `AWS_ACCESS_KEY_ID`: AWS access key
- `AWS_SECRET_ACCESS_KEY`: AWS secret key
- `AWS_SESSION_TOKEN`: Session token (for temporary roles)

## Development

### Prerequisites

- Go 1.21 or later
- AWS credentials configured
- Git

### Building from Source

```bash
# Clone the repository
git clone https://github.com/hypolas/aws-ssm-light.git
cd aws-ssm

# Build for your platform
go build -o aws-ssm main.go

# Build for specific platforms
GOOS=linux GOARCH=amd64 go build -o aws-ssm-linux-amd64 main.go
GOOS=windows GOARCH=amd64 go build -o aws-ssm-windows-amd64.exe main.go
GOOS=darwin GOARCH=arm64 go build -o aws-ssm-darwin-arm64 main.go
```

## Development

### Prerequisites

- Go 1.21 or later
- AWS credentials configured
- Git

### Building from Source

```bash
# Clone the repository
git clone https://github.com/hypolas/aws-ssm-light.git
cd aws-ssm

# Build for your platform
go build -o aws-ssm main.go

# Build for specific platforms
GOOS=linux GOARCH=amd64 go build -o aws-ssm-linux-amd64 main.go
GOOS=windows GOARCH=amd64 go build -o aws-ssm-windows-amd64.exe main.go
GOOS=darwin GOARCH=arm64 go build -o aws-ssm-darwin-arm64 main.go
```

#### Windows Users

For Windows development, use the provided PowerShell build script:

```powershell
# Build with version information
.\build.ps1

# Or build manually
go build -o aws-ssm.exe main.go

# Test the version
.\aws-ssm.exe --version
```

### Testing

The project includes comprehensive unit and integration tests:

```bash
# Run all tests
go test -v ./...

# Run only unit tests (skip integration tests)
go test -short -v ./...

# Run tests with coverage
go test -coverprofile=coverage.out .
go tool cover -html=coverage.out -o coverage.html

# Run with race detection
go test -race -short ./...

# Using Makefile (recommended)
make test           # Run all tests
make test-short     # Run unit tests only
make test-coverage  # Generate coverage report
make benchmark      # Run performance benchmarks
make security       # Run security analysis
```

#### Security Testing

The project includes automated security scanning using gosec:

```bash
# Install security tools
make install-tools

# Run security analysis
make security

# Or run gosec directly
gosec ./...
```

Security checks include:
- ✅ **SQL injection** detection
- ✅ **Hardcoded credentials** detection
- ✅ **File traversal** vulnerabilities
- ✅ **Command injection** risks
- ✅ **TLS/crypto** best practices
- ✅ **Unsafe operations** detection

#### Test Coverage
- **Current Coverage**: ~49% (focused on business logic)
- **Target**: >90% for critical functions
- **Test Types**: Unit tests, integration tests, benchmarks, security scans

#### Running Integration Tests
Integration tests require AWS credentials:

```bash
# Set up test environment
export AWS_INTEGRATION_TEST_REGION="us-east-1"
export AWS_INTEGRATION_TEST_SECRET="your-test-secret-name"

# Run integration tests
make test-integration
```

See [TESTING.md](TESTING.md) for detailed testing information.

### Running Tests

```bash
go test -v ./...
```

## CI/CD

This project uses GitHub Actions for automated building, testing, and releasing:

- **🔄 Automated Builds**: Multi-platform binaries built on every push
- **🧪 Comprehensive Testing**: Unit, integration, and benchmark tests
- **� Security Scanning**: Automated vulnerability detection with gosec
- **📊 Code Quality**: Linting with golangci-lint and coverage reporting
- **�📦 Automatic Releases**: Version bumping based on conventional commits
- **� Binary Security**: SHA256 checksums for all releases
- **🔧 Dependency Management**: Renovate bot for automatic dependency updates

### Security Features

- **Gosec Integration**: Scans for common Go security issues
- **Multi-platform Testing**: Ensures security across Windows, Linux, macOS
- **Dependency Scanning**: Renovate monitors for security updates
- **Code Quality Gates**: All tests must pass before merge
- **Binary Verification**: SHA256 checksums prevent tampering

### Release Process

#### Automatic Releases
Releases are automatically created based on commit messages:

```bash
# Creates a minor version bump (v1.0.0 → v1.1.0)
git commit -m "feat: add new secret caching feature"

# Creates a patch version bump (v1.0.0 → v1.0.1)
git commit -m "fix: handle timeout errors properly"
```

#### Manual Releases
1. Go to **Actions** → **Auto Release** → **Run workflow**
2. Enter the desired version (e.g., `v1.2.0`)
3. The workflow will create the tag and release automatically

### Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make your changes and commit using conventional commits
4. Push to your fork and submit a pull request

## Build

```bash
# For Linux (in container)
GOOS=linux GOARCH=amd64 go build -o aws-ssm main.go

# For Windows
GOOS=windows GOARCH=amd64 go build -o aws-ssm.exe main.go

# For macOS
GOOS=darwin GOARCH=amd64 go build -o aws-ssm main.go
```

## Docker Integration

The binary was specifically designed to **replace AWS CLI in Docker images**, providing a lightweight alternative for secret retrieval.

### Docker Size Comparison

| Solution | Image Size Impact | Memory Usage | Startup Time |
|----------|------------------|--------------|--------------|
| AWS CLI | +100MB+ | ~50MB+ | ~1-2s |
| aws-ssm | +10MB | ~5MB | ~50ms |

### Example Dockerfile

```dockerfile
FROM ubuntu:latest

# Instead of installing AWS CLI (100MB+)
# RUN apt-get update && apt-get install -y awscli

# Use lightweight aws-ssm (10MB)
RUN curl -L https://github.com/hypolas/aws-ssm-light/releases/latest/download/aws-ssm-linux-amd64 -o /usr/local/bin/aws-ssm && \
    chmod +x /usr/local/bin/aws-ssm

# Perfect for CI/CD and production containers
# Use in your scripts instead of AWS CLI
# aws-ssm "my-secret" instead of aws secretsmanager get-secret-value --secret-id "my-secret" --query SecretString --output text
```

### Multi-stage Docker Build

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -ldflags="-w -s" -o aws-ssm main.go

# Production stage - minimal image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/aws-ssm /usr/local/bin/
ENTRYPOINT ["aws-ssm"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-with-secrets
spec:
  template:
    spec:
      containers:
      - name: app
        image: your-app:latest
        env:
        - name: DATABASE_PASSWORD
          value: "$(aws-ssm 'prod/database/password')"
        # Minimal resource requirements thanks to lightweight binary
        resources:
          requests:
            memory: "32Mi"
            cpu: "10m"
```

## Verification

### Verify Binary Integrity

Each release includes SHA256 checksums. To verify your download:

```bash
# Download the binary and checksum
curl -L https://github.com/hypolas/aws-ssm-light/releases/latest/download/aws-ssm-linux-amd64 -o aws-ssm
curl -L https://github.com/hypolas/aws-ssm-light/releases/latest/download/aws-ssm-linux-amd64.sha256 -o aws-ssm.sha256

# Verify the checksum
sha256sum -c aws-ssm.sha256
```

### Verify Installation

```bash
# Check version
aws-ssm --version

# Test with a dummy secret (will fail if secret doesn't exist, but confirms AWS connectivity)
aws-ssm "test-secret-name"
```

## Performance Comparison

| Tool | Binary Size | Memory Usage | Startup Time |
|------|-------------|--------------|--------------|
| AWS CLI | ~100MB+ | ~50MB+ | ~1-2s |
| aws-ssm | ~10MB | ~5MB | ~50ms |

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📖 [Documentation](https://github.com/hypolas/aws-ssm-light/wiki)
- 🐛 [Report Issues](https://github.com/hypolas/aws-ssm-light/issues)
- 💬 [Discussions](https://github.com/hypolas/aws-ssm-light/discussions)
- 📋 [Changelog](CHANGELOG.md)