package session_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/denvrdata/go-denvr/result"
	"github.com/denvrdata/go-denvr/session"
	"github.com/stretchr/testify/assert"
)

func TestSession(t *testing.T) {
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

	f := result.Wrap(os.CreateTemp("", "test-newconfig-tmpfile-")).Unwrap()
	defer f.Close()
	defer os.Remove(f.Name())
	result.Wrap(f.Write([]byte(content))).Unwrap()

	t.Run(
		"NewSession",
		func(t *testing.T) {
			os.Setenv("DENVR_CONFIG", f.Name())
			s := session.NewSession()
			assert.Equal(t, server.URL, s.Config.Server)
			assert.Equal(t, int64(5), s.Config.Retries)

			t.Cleanup(func() { os.Unsetenv("DENVR_CONFIG") })
		},
	)

}
