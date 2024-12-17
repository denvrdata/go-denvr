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

func TestAuth(t *testing.T) {
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

	expected := config.Config{
		auth.NewAuth(server.URL, "test@foobar.com", "test.foo.bar.baz"),
		server.URL,
		"v2",
		"Hou1",
		"denvr",
		"denvr",
		"reserved-denvr",
		5,
	}

	f := result.Wrap(os.CreateTemp("", "test-newconfig-tmpfile-")).Unwrap()
	defer f.Close()
	defer os.Remove(f.Name())
	result.Wrap(f.Write([]byte(content))).Unwrap()

	t.Run(
		"NewConfig",
		func(t *testing.T) {
			assert.Equal(t, expected, config.NewConfig(f.Name()))
		},
	)
}
