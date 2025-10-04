package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Version information - set during build with -ldflags
var (
	Version    = "dev"
	GitCommit  = "unknown"
	BuildTime  = "unknown"
	Maintainer = "Nicolas HYPOLITE"
)

// SecretsManagerClient interface for testing
type SecretsManagerClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// Config holds the application configuration
type Config struct {
	SecretID string
	Region   string
}

// App holds the application dependencies
type App struct {
	Client SecretsManagerClient
	Config Config
}

// ShowVersion displays version information
func ShowVersion() {
	fmt.Printf("aws-ssm version %s\n", Version)
	if GitCommit != "unknown" {
		fmt.Printf("Git commit: %s\n", GitCommit)
	}
	if BuildTime != "unknown" {
		fmt.Printf("Built: %s\n", BuildTime)
	}
	if Maintainer != "" {
		// Replace underscore with space for display
		maintainerName := strings.ReplaceAll(Maintainer, "_", " ")
		fmt.Printf("Maintainer: %s\n", maintainerName)
	}
}

// ShowUsage displays usage information
func ShowUsage(progName string) {
	fmt.Fprintf(os.Stderr, "Usage: %s <secret-id> [region]\n", progName)
	fmt.Fprintf(os.Stderr, "       %s --version\n", progName)
	fmt.Fprintf(os.Stderr, "       %s --help\n", progName)
	fmt.Fprintf(os.Stderr, "\nArguments:\n")
	fmt.Fprintf(os.Stderr, "  secret-id    AWS Secrets Manager secret ID or ARN\n")
	fmt.Fprintf(os.Stderr, "  region       AWS region (optional, overrides AWS_REGION)\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  --version    Show version information\n")
	fmt.Fprintf(os.Stderr, "  --help       Show this help message\n")
	fmt.Fprintf(os.Stderr, "\nEnvironment variables:\n")
	fmt.Fprintf(os.Stderr, "  AWS_REGION: AWS region (can be overridden by second argument)\n")
	fmt.Fprintf(os.Stderr, "  AWS_ACCESS_KEY_ID: AWS access key\n")
	fmt.Fprintf(os.Stderr, "  AWS_SECRET_ACCESS_KEY: AWS secret key\n")
	fmt.Fprintf(os.Stderr, "  AWS_SESSION_TOKEN: Session token (for temporary roles)\n")
}

// ParseArgs parses command line arguments and environment variables
func ParseArgs(args []string) (Config, error) {
	if len(args) < 2 {
		return Config{}, fmt.Errorf("insufficient arguments")
	}

	// Handle version flag
	if args[1] == "--version" || args[1] == "-v" {
		ShowVersion()
		os.Exit(0)
	}

	// Handle help flag
	if args[1] == "--help" || args[1] == "-h" {
		ShowUsage(args[0])
		os.Exit(0)
	}

	secretID := args[1]
	region := os.Getenv("AWS_REGION")

	// Override region if provided as argument
	if len(args) > 2 {
		region = args[2]
	}

	if region == "" {
		return Config{}, fmt.Errorf("AWS region must be specified either via AWS_REGION environment variable or as second argument")
	}

	return Config{
		SecretID: secretID,
		Region:   region,
	}, nil
}

// GetSecret retrieves a secret from AWS Secrets Manager
func (app *App) GetSecret(ctx context.Context) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &app.Config.SecretID,
	}

	result, err := app.Client.GetSecretValue(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get secret value: %w", err)
	}

	if result.SecretString == nil {
		return "", fmt.Errorf("secret does not contain a string value")
	}

	return *result.SecretString, nil
}

// FormatOutput formats the secret output (handles JSON detection)
func FormatOutput(secretValue string) string {
	// Try to parse as JSON first, if it fails, output as is
	var jsonData interface{}
	if err := json.Unmarshal([]byte(secretValue), &jsonData); err != nil {
		// Not JSON, output the string directly
		return secretValue
	}
	// Is JSON, output the raw JSON string (like AWS CLI does)
	return secretValue
}

// NewApp creates a new application instance
func NewApp(cfg Config) (*App, error) {
	// Load AWS configuration
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Secrets Manager client
	client := secretsmanager.NewFromConfig(awsCfg)

	return &App{
		Client: client,
		Config: cfg,
	}, nil
}

func main() {
	cfg, err := ParseArgs(os.Args)
	if err != nil {
		if err.Error() == "insufficient arguments" {
			ShowUsage(os.Args[0])
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	app, err := NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	secretValue, err := app.GetSecret(context.TODO())
	if err != nil {
		log.Fatalf("Failed to get secret: %v", err)
	}

	output := FormatOutput(secretValue)
	fmt.Print(output)
}
