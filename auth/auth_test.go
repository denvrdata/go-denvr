package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denvrdata/go-denvr/auth"
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

	t.Run(
		"NewAuth",
		func(t *testing.T) {
			assert.Equal(
				t,
				auth.Auth{server.URL, "access1", "refresh", time.Now().Unix() + 60, time.Now().Unix() + 3600},
				auth.NewAuth(
					server.URL,
					"alice@denvrtest.com",
					"alice.is.the.best",
				),
			)
		},
	)
}
