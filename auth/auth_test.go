package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denvrdata/go-denvr/auth"
	"github.com/stretchr/testify/assert"
)

func TestNewAuth(t *testing.T) {
	t.Run(
		"NoCredentialsKey",
		func(t *testing.T) {
			content := map[string]any{}
			httpClient := &http.Client{}

			defer func() {
				if r := recover(); r != nil {
					assert.Contains(t, fmt.Sprintf("%v", r), "Authentication failed.")
				} else {
					t.Fatal("Expected panic but none occurred")
				}
			}()

			auth.NewAuth("/path/to/config.toml", content, "https://api.example.com", httpClient)
		},
	)

	t.Run(
		"EmptyCredentials",
		func(t *testing.T) {
			content := map[string]any{
				"credentials": map[string]any{},
			}
			httpClient := &http.Client{}

			defer func() {
				if r := recover(); r != nil {
					assert.Contains(t, fmt.Sprintf("%v", r), "Authentication failed.")
				} else {
					t.Fatal("Expected panic but none occurred")
				}
			}()

			auth.NewAuth("/path/to/config.toml", content, "https://api.example.com", httpClient)
		},
	)

	t.Run(
		"CredentialsWithApiKey",
		func(t *testing.T) {
			content := map[string]any{
				"credentials": map[string]any{
					"apikey": "test-api-key-123",
				},
			}
			httpClient := &http.Client{}
			result := auth.NewAuth("/path/to/config.toml", content, "https://api.example.com", httpClient)
			assert.NotNil(t, result)

			// Test the intercept function
			req, _ := http.NewRequest("GET", "https://api.example.com/test", nil)
			result.Intercept(req.Context(), req)
			assert.Equal(t, "ApiKey test-api-key-123", req.Header.Get("Authorization"))
		},
	)
	server := httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
				writer.Write(
					[]byte(`{
						"result": {
							"accessToken": "access1",
							"refreshToken": "refresh",
							"expireInSeconds": 60,
							"refreshTokenExpireInSeconds": 3600
						}
					}`),
				)
			},
		),
	)
	t.Run(
		"CredentialsWithUsernamePassword",
		func(t *testing.T) {
			content := map[string]any{
				"credentials": map[string]any{
					"username": "testuser",
					"password": "testpass",
				},
			}
			httpClient := &http.Client{}
			result := auth.NewAuth("/path/to/config.toml", content, server.URL, httpClient)
			assert.NotNil(t, result)

			// Test the intercept function
			req, _ := http.NewRequest("GET", "https://api.example.com/test", nil)
			result.Intercept(req.Context(), req)
			assert.Equal(t, "Bearer access1", req.Header.Get("Authorization"))
		},
	)
}

func TestBearer(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
				writer.Write(
					[]byte(`{
						"result": {
							"accessToken": "access1",
							"refreshToken": "refresh",
							"expireInSeconds": 60,
							"refreshTokenExpireInSeconds": 3600
						}
					}`),
				)
			},
		),
	)

	t.Run(
		"NewAuth",
		func(t *testing.T) {
			httpClient := &http.Client{}
			result := auth.NewBearer(
				server.URL,
				"alice@denvrtest.com",
				"alice.is.the.best",
				httpClient,
			)

			assert.Equal(t, server.URL, result.Server)
			assert.Equal(t, "access1", result.AccessToken)
			assert.Equal(t, "refresh", result.RefreshToken)
			assert.NotNil(t, result.Client)
			// Allow some tolerance for timing differences
			assert.InDelta(t, time.Now().Unix()+60, result.AccessExpires, 5)
			assert.InDelta(t, time.Now().Unix()+3600, result.RefreshExpires, 5)
		},
	)
}
