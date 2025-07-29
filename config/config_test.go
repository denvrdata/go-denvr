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

func TestConfig(t *testing.T) {
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
		auth.NewAuth(server.URL, "test@foobar.com", "test.foo.bar.baz", httpClient),
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
			assert.Equal(t, expected.Auth.Server, result.Auth.Server)
			assert.Equal(t, expected.Auth.AccessToken, result.Auth.AccessToken)
			assert.Equal(t, expected.Auth.RefreshToken, result.Auth.RefreshToken)
			// Compare Auth Client field (just check that both are non-nil)
			assert.NotNil(t, expected.Auth.Client)
			assert.NotNil(t, result.Auth.Client)
		},
	)
}
