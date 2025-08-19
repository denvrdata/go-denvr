package config_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/denvrdata/go-denvr/auth"
	"github.com/denvrdata/go-denvr/config"
	"github.com/denvrdata/go-denvr/result"
	"github.com/stretchr/testify/assert"
)

func TestConfigNoCredentials(t *testing.T) {
	content := `[defaults]
		server = "http://localhost:8080"
		api = "v2"
		cluster = "Hou1"
		tenant = "denvr"
		vpcid = "denvr"
		rpool = "reserved-denvr"
		retries = 5`

	f := result.Wrap(os.CreateTemp("", "test-newconfig-tmpfile-")).Unwrap()
	defer f.Close()
	defer os.Remove(f.Name())
	result.Wrap(f.Write([]byte(content))).Unwrap()

	t.Run(
		"NoCredentialsNoEnvVars",
		func(t *testing.T) {
			// Ensure DENVR_APIKEY is not set
			os.Unsetenv("DENVR_APIKEY")

			defer func() {
				if r := recover(); r != nil {
					assert.Contains(t, fmt.Sprintf("%v", r), "Authentication failed.")
				} else {
					t.Fatal("Expected panic but none occurred")
				}
			}()

			config.NewConfig(f.Name())
		},
	)
}

func TestConfigWithEnvAPIKey(t *testing.T) {
	content := `[defaults]
        server = "http://localhost:8080"
        api = "v2"
        cluster = "Hou1"
        tenant = "denvr"
        vpcid = "denvr"
        rpool = "reserved-denvr"
        retries = 5`

	f := result.Wrap(os.CreateTemp("", "test-newconfig-tmpfile-")).Unwrap()
	defer f.Close()
	defer os.Remove(f.Name())
	result.Wrap(f.Write([]byte(content))).Unwrap()

	t.Run(
		"NoCredentialsSectionWithEnvAPIKey",
		func(t *testing.T) {
			// Set DENVR_APIKEY environment variable
			os.Setenv("DENVR_APIKEY", "foo.bar.baz")
			defer os.Unsetenv("DENVR_APIKEY")

			result := config.NewConfig(f.Name())

			assert.Equal(t, "http://localhost:8080", result.Server)
			assert.Equal(t, "v2", result.API)
			assert.Equal(t, "Hou1", result.Cluster)
			assert.Equal(t, "denvr", result.Tenant)
			assert.Equal(t, "denvr", result.VPCId)
			assert.Equal(t, "reserved-denvr", result.RPool)
			assert.NotNil(t, result.Client)

			// Check that Auth is set with the API key
			assert.NotNil(t, result.Auth)
			apikey_auth := result.Auth.(auth.ApiKey)
			assert.Equal(t, "foo.bar.baz", apikey_auth.Key)
		},
	)
}

func TestConfigWithAPIKeyInCredentials(t *testing.T) {
	content := `[defaults]
        server = "http://localhost:8080"
        api = "v2"
        cluster = "Hou1"
        tenant = "denvr"
        vpcid = "denvr"
        rpool = "reserved-denvr"
        retries = 5

        [credentials]
        apikey = "foo.bar.baz"`

	f := result.Wrap(os.CreateTemp("", "test-newconfig-tmpfile-")).Unwrap()
	defer f.Close()
	defer os.Remove(f.Name())
	result.Wrap(f.Write([]byte(content))).Unwrap()

	t.Run(
		"CredentialsWithAPIKey",
		func(t *testing.T) {
			result := config.NewConfig(f.Name())

			assert.Equal(t, "http://localhost:8080", result.Server)
			assert.Equal(t, "v2", result.API)
			assert.Equal(t, "Hou1", result.Cluster)
			assert.Equal(t, "denvr", result.Tenant)
			assert.Equal(t, "denvr", result.VPCId)
			assert.Equal(t, "reserved-denvr", result.RPool)
			assert.NotNil(t, result.Client)

			// Check that Auth is set with the API key
			assert.NotNil(t, result.Auth)
			apikey_auth := result.Auth.(auth.ApiKey)
			assert.Equal(t, "foo.bar.baz", apikey_auth.Key)
		},
	)
}

func TestBearerConfig(t *testing.T) {
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

	content := fmt.Sprintf(
		`[defaults]
        server = "%s"
        api = "v2"
        cluster = "Hou1"
        tenant = "denvr"
        vpcid = "denvr"
        rpool = "reserved-denvr"
        retries = 5

        [credentials]
        username = "test@foobar.com"
        password = "test.foo.bar.baz"`,
		server.URL,
	)

	// Create an HTTP client for the test
	httpClient := &http.Client{}

	expected := config.Config{
		auth.NewBearer(server.URL, "test@foobar.com", "test.foo.bar.baz", httpClient),
		server.URL,
		"v2",
		"Hou1",
		"denvr",
		"denvr",
		"reserved-denvr",
		httpClient,
	}

	f := result.Wrap(os.CreateTemp("", "test-newconfig-tmpfile-")).Unwrap()
	defer f.Close()
	defer os.Remove(f.Name())
	result.Wrap(f.Write([]byte(content))).Unwrap()

	t.Run(
		"NewConfig",
		func(t *testing.T) {
			result := config.NewConfig(f.Name())

			// Compare all fields except Auth since it contains a non-comparable Client field
			assert.Equal(t, expected.Server, result.Server)
			assert.Equal(t, expected.API, result.API)
			assert.Equal(t, expected.Cluster, result.Cluster)
			assert.Equal(t, expected.Tenant, result.Tenant)
			assert.Equal(t, expected.VPCId, result.VPCId)
			assert.Equal(t, expected.RPool, result.RPool)
			// Compare Client field (just check that both are non-nil)
			assert.NotNil(t, expected.Client)
			assert.NotNil(t, result.Client)

			// Compare Auth fields individually
			expected_auth := expected.Auth.(auth.Bearer)
			result_auth := result.Auth.(auth.Bearer)
			assert.Equal(t, expected_auth.Server, result_auth.Server)
			assert.Equal(t, expected_auth.AccessToken, result_auth.AccessToken)
			assert.Equal(t, expected_auth.RefreshToken, result_auth.RefreshToken)
			// Compare Auth Client field (just check that both are non-nil)
			assert.NotNil(t, expected_auth.Client)
			assert.NotNil(t, result_auth.Client)
		},
	)
}
