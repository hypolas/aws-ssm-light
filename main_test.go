package main

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSecretsManagerClient is a mock implementation of SecretsManagerClient
type MockSecretsManagerClient struct {
	mock.Mock
}

func (m *MockSecretsManagerClient) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*secretsmanager.GetSecretValueOutput), args.Error(1)
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		envVars  map[string]string
		expected Config
		wantErr  bool
	}{
		{
			name:    "missing secret ID",
			args:    []string{"aws-ssm"},
			wantErr: true,
		},
		{
			name:    "secret ID with region from env",
			args:    []string{"aws-ssm", "my-secret"},
			envVars: map[string]string{"AWS_REGION": "us-east-1"},
			expected: Config{
				SecretID: "my-secret",
				Region:   "us-east-1",
			},
			wantErr: false,
		},
		{
			name: "secret ID with region from args",
			args: []string{"aws-ssm", "my-secret", "eu-west-1"},
			expected: Config{
				SecretID: "my-secret",
				Region:   "eu-west-1",
			},
			wantErr: false,
		},
		{
			name: "region from args overrides env",
			args: []string{"aws-ssm", "my-secret", "eu-west-1"},
			envVars: map[string]string{"AWS_REGION": "us-east-1"},
			expected: Config{
				SecretID: "my-secret",
				Region:   "eu-west-1",
			},
			wantErr: false,
		},
		{
			name:    "no region specified",
			args:    []string{"aws-ssm", "my-secret"},
			wantErr: true,
		},
		{
			name: "ARN as secret ID",
			args: []string{"aws-ssm", "arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret-AbCdEf", "us-east-1"},
			expected: Config{
				SecretID: "arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret-AbCdEf",
				Region:   "us-east-1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Clear AWS_REGION if not in envVars to ensure clean test
			if _, exists := tt.envVars["AWS_REGION"]; !exists {
				os.Unsetenv("AWS_REGION")
			}

			got, err := ParseArgs(tt.args)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestFormatOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text secret",
			input:    "plain-text-password",
			expected: "plain-text-password",
		},
		{
			name:     "JSON secret",
			input:    `{"username": "admin", "password": "secret123"}`,
			expected: `{"username": "admin", "password": "secret123"}`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "malformed JSON",
			input:    `{"username": "admin", "password":`,
			expected: `{"username": "admin", "password":`,
		},
		{
			name:     "number as string",
			input:    "12345",
			expected: "12345",
		},
		{
			name:     "JSON array",
			input:    `["value1", "value2", "value3"]`,
			expected: `["value1", "value2", "value3"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatOutput(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestApp_GetSecret(t *testing.T) {
	tests := []struct {
		name          string
		secretID      string
		mockResponse  *secretsmanager.GetSecretValueOutput
		mockError     error
		expectedValue string
		wantErr       bool
	}{
		{
			name:     "successful secret retrieval",
			secretID: "my-secret",
			mockResponse: &secretsmanager.GetSecretValueOutput{
				SecretString: stringPtr("my-secret-value"),
			},
			expectedValue: "my-secret-value",
			wantErr:       false,
		},
		{
			name:     "JSON secret retrieval",
			secretID: "json-secret",
			mockResponse: &secretsmanager.GetSecretValueOutput{
				SecretString: stringPtr(`{"key": "value"}`),
			},
			expectedValue: `{"key": "value"}`,
			wantErr:       false,
		},
		{
			name:      "AWS error",
			secretID:  "non-existent-secret",
			mockError: assert.AnError,
			wantErr:   true,
		},
		{
			name:     "nil secret string",
			secretID: "binary-secret",
			mockResponse: &secretsmanager.GetSecretValueOutput{
				SecretString: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockSecretsManagerClient)
			
			if tt.mockError != nil {
				mockClient.On("GetSecretValue", mock.Anything, mock.MatchedBy(func(input *secretsmanager.GetSecretValueInput) bool {
					return *input.SecretId == tt.secretID
				})).Return((*secretsmanager.GetSecretValueOutput)(nil), tt.mockError)
			} else {
				mockClient.On("GetSecretValue", mock.Anything, mock.MatchedBy(func(input *secretsmanager.GetSecretValueInput) bool {
					return *input.SecretId == tt.secretID
				})).Return(tt.mockResponse, nil)
			}

			app := &App{
				Client: mockClient,
				Config: Config{
					SecretID: tt.secretID,
					Region:   "us-east-1",
				},
			}

			got, err := app.GetSecret(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, got)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestApp_GetSecret_Integration(t *testing.T) {
	// Skip integration tests in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires AWS credentials and should only run in CI or with proper setup
	// You can set AWS_SKIP_INTEGRATION_TESTS=true to skip these tests
	if os.Getenv("AWS_SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Integration tests disabled via AWS_SKIP_INTEGRATION_TESTS")
	}

	// This would require actual AWS credentials and a test secret
	// For now, we'll skip it unless explicitly enabled
	t.Skip("Integration test requires AWS credentials and test secret")
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}