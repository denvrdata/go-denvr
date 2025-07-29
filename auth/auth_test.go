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
			httpClient := &http.Client{}
			result := auth.NewAuth(
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
