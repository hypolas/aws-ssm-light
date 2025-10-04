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

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <secret-id> [region]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Environment variables:\n")
		fmt.Fprintf(os.Stderr, "  AWS_REGION: AWS region (can be overridden by second argument)\n")
		fmt.Fprintf(os.Stderr, "  AWS_ACCESS_KEY_ID: AWS access key\n")
		fmt.Fprintf(os.Stderr, "  AWS_SECRET_ACCESS_KEY: AWS secret key\n")
		os.Exit(1)
	}

	secretID := os.Args[1]
	region := os.Getenv("AWS_REGION")

	// Override region if provided as argument
	if len(os.Args) > 2 {
		region = os.Args[2]
	}

	if region == "" {
		log.Fatal("AWS region must be specified either via AWS_REGION environment variable or as second argument")
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Create Secrets Manager client
	client := secretsmanager.NewFromConfig(cfg)

	// Get secret value
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretID,
	}

	result, err := client.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Fatalf("Failed to get secret value: %v", err)
	}

	if result.SecretString == nil {
		log.Fatal("Secret does not contain a string value")
	}

	// Try to parse as JSON first, if it fails, output as is
	var jsonData interface{}
	if err := json.Unmarshal([]byte(*result.SecretString), &jsonData); err != nil {
		// Not JSON, output the string directly
		fmt.Print(*result.SecretString)
	} else {
		// Is JSON, output the raw JSON string (like AWS CLI does)
		fmt.Print(*result.SecretString)
	}
}
