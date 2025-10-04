package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
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

// ParseArgs parses command line arguments and environment variables
func ParseArgs(args []string) (Config, error) {
	if len(args) < 2 {
		return Config{}, fmt.Errorf("usage: %s <secret-id> [region]", args[0])
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Environment variables:\n")
		fmt.Fprintf(os.Stderr, "  AWS_REGION: AWS region (can be overridden by second argument)\n")
		fmt.Fprintf(os.Stderr, "  AWS_ACCESS_KEY_ID: AWS access key\n")
		fmt.Fprintf(os.Stderr, "  AWS_SECRET_ACCESS_KEY: AWS secret key\n")
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
