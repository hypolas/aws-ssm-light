package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSecretManager_RealAWS contains integration tests that run against real AWS
// These tests require:
// 1. Valid AWS credentials
// 2. AWS_INTEGRATION_TEST_REGION environment variable
// 3. AWS_INTEGRATION_TEST_SECRET environment variable (pointing to a test secret)
func TestSecretManager_RealAWS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	region := os.Getenv("AWS_INTEGRATION_TEST_REGION")
	testSecret := os.Getenv("AWS_INTEGRATION_TEST_SECRET")

	if region == "" || testSecret == "" {
		t.Skip("Skipping AWS integration tests - set AWS_INTEGRATION_TEST_REGION and AWS_INTEGRATION_TEST_SECRET")
	}

	t.Run("real AWS secret retrieval", func(t *testing.T) {
		cfg := Config{
			SecretID: testSecret,
			Region:   region,
		}

		app, err := NewApp(cfg)
		require.NoError(t, err, "Failed to create app")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		secretValue, err := app.GetSecret(ctx)
		require.NoError(t, err, "Failed to get secret")
		assert.NotEmpty(t, secretValue, "Secret value should not be empty")
	})
}

// TestE2E_WithMockAWS provides end-to-end testing with a mock AWS environment
func TestE2E_WithMockAWS(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		envVars        map[string]string
		mockResponse   *secretsmanager.GetSecretValueOutput
		mockError      error
		expectedOutput string
		expectError    bool
	}{
		{
			name: "successful e2e flow with plain text secret",
			args: []string{"aws-ssm", "test-secret"},
			envVars: map[string]string{
				"AWS_REGION": "us-east-1",
			},
			mockResponse: &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String("my-plain-text-secret"),
			},
			expectedOutput: "my-plain-text-secret",
			expectError:    false,
		},
		{
			name: "successful e2e flow with JSON secret",
			args: []string{"aws-ssm", "json-secret"},
			envVars: map[string]string{
				"AWS_REGION": "eu-west-1",
			},
			mockResponse: &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String(`{"username":"admin","password":"secret123"}`),
			},
			expectedOutput: `{"username":"admin","password":"secret123"}`,
			expectError:    false,
		},
		{
			name: "e2e flow with region override",
			args: []string{"aws-ssm", "test-secret", "ap-southeast-1"},
			envVars: map[string]string{
				"AWS_REGION": "us-east-1", // Should be overridden
			},
			mockResponse: &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String("region-override-test"),
			},
			expectedOutput: "region-override-test",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			originalEnv := make(map[string]string)
			for key, value := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("Failed to set env var %s: %v", key, err)
				}
			}
			defer func() {
				for key, originalValue := range originalEnv {
					if originalValue == "" {
						if err := os.Unsetenv(key); err != nil {
							t.Fatalf("Failed to unset env var %s: %v", key, err)
						}
					} else {
						if err := os.Setenv(key, originalValue); err != nil {
							t.Fatalf("Failed to restore env var %s: %v", key, err)
						}
					}
				}
			}()

			// Parse arguments
			cfg, err := ParseArgs(tt.args)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Create mock client
			mockClient := new(MockSecretsManagerClient)
			if tt.mockError != nil {
				mockClient.On("GetSecretValue", context.Background(), &secretsmanager.GetSecretValueInput{
					SecretId: aws.String(cfg.SecretID),
				}).Return((*secretsmanager.GetSecretValueOutput)(nil), tt.mockError)
			} else {
				mockClient.On("GetSecretValue", context.Background(), &secretsmanager.GetSecretValueInput{
					SecretId: aws.String(cfg.SecretID),
				}).Return(tt.mockResponse, nil)
			}

			// Create app with mock client
			app := &App{
				Client: mockClient,
				Config: cfg,
			}

			// Get secret
			secretValue, err := app.GetSecret(context.Background())
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Format output
			output := FormatOutput(secretValue)
			assert.Equal(t, tt.expectedOutput, output)

			mockClient.AssertExpectations(t)
		})
	}
}

// BenchmarkGetSecret benchmarks the secret retrieval performance
func BenchmarkGetSecret(b *testing.B) {
	mockClient := new(MockSecretsManagerClient)
	mockResponse := &secretsmanager.GetSecretValueOutput{
		SecretString: aws.String("benchmark-secret-value"),
	}

	mockClient.On("GetSecretValue", context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String("benchmark-secret"),
	}).Return(mockResponse, nil)

	app := &App{
		Client: mockClient,
		Config: Config{
			SecretID: "benchmark-secret",
			Region:   "us-east-1",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := app.GetSecret(context.Background())
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// BenchmarkFormatOutput benchmarks the output formatting
func BenchmarkFormatOutput(b *testing.B) {
	tests := []struct {
		name  string
		input string
	}{
		{"plain_text", "simple-password"},
		{"small_json", `{"key": "value"}`},
		{"large_json", `{"username": "admin", "password": "very-long-password-that-might-be-typical-in-real-world-usage", "database": "production", "host": "db.example.com", "port": 5432}`},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				FormatOutput(tt.input)
			}
		})
	}
}

// TestConcurrentAccess tests concurrent access to secrets
func TestConcurrentAccess(t *testing.T) {
	const numGoroutines = 10

	mockClient := new(MockSecretsManagerClient)
	mockResponse := &secretsmanager.GetSecretValueOutput{
		SecretString: aws.String("concurrent-test-secret"),
	}

	// Set up expectations for concurrent calls
	mockClient.On("GetSecretValue", context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String("concurrent-secret"),
	}).Return(mockResponse, nil).Times(numGoroutines)

	app := &App{
		Client: mockClient,
		Config: Config{
			SecretID: "concurrent-secret",
			Region:   "us-east-1",
		},
	}

	// Channel to collect results
	results := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		go func() {
			secretValue, err := app.GetSecret(context.Background())
			if err != nil {
				errors <- err
				return
			}
			results <- secretValue
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		select {
		case result := <-results:
			assert.Equal(t, "concurrent-test-secret", result)
		case err := <-errors:
			t.Errorf("Unexpected error in goroutine: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for goroutine to complete")
		}
	}

	mockClient.AssertExpectations(t)
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		mockError   error
		expectError string
	}{
		{
			name:        "network timeout",
			mockError:   context.DeadlineExceeded,
			expectError: "failed to get secret value: context deadline exceeded",
		},
		{
			name:        "generic AWS error",
			mockError:   assert.AnError,
			expectError: "failed to get secret value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockSecretsManagerClient)
			mockClient.On("GetSecretValue", context.Background(), &secretsmanager.GetSecretValueInput{
				SecretId: aws.String("error-test-secret"),
			}).Return((*secretsmanager.GetSecretValueOutput)(nil), tt.mockError)

			app := &App{
				Client: mockClient,
				Config: Config{
					SecretID: "error-test-secret",
					Region:   "us-east-1",
				},
			}

			_, err := app.GetSecret(context.Background())
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)

			mockClient.AssertExpectations(t)
		})
	}
}
