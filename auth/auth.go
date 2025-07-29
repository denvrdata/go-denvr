package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/denvrdata/go-denvr/result"
)

type Auth struct {
	Server         string
	AccessToken    string
	RefreshToken   string
	AccessExpires  int64
	RefreshExpires int64
	Client         *http.Client
}

func NewAuth(server string, username string, password string, client *http.Client) Auth {
	data := result.Wrap(
		json.Marshal(
			map[string]string{
				"userNameOrEmailAddress": username,
				"password":               password,
			},
		),
	).Unwrap()

	req := result.Wrap(
		http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("%s/api/TokenAuth/Authenticate", server),
			bytes.NewBuffer(data),
		),
	).Unwrap()
	req.Header.Set("Content-Type", "application/json")

	resp := result.Wrap(client.Do(req)).Unwrap()

	defer resp.Body.Close()

	// A bit ugly, but we'll define our specific response content to decode
	var content struct {
		Result struct {
			AccessToken                 string `json:"accessToken"`
			RefreshToken                string `json:"refreshToken"`
			ExpireInSeconds             int64  `json:"expireInSeconds"`
			RefreshTokenExpireInSeconds int64  `json:"refreshTokenExpireInSeconds"`
		} `json:"result"`
	}
	result.Wrap(content, json.NewDecoder(resp.Body).Decode(&content)).Unwrap()

	return Auth{
		server,
		content.Result.AccessToken,
		content.Result.RefreshToken,
		time.Now().Unix() + content.Result.ExpireInSeconds,
		time.Now().Unix() + content.Result.RefreshTokenExpireInSeconds,
		client,
	}
}

func (auth Auth) Token() string {
	t := time.Now().Unix()

	if t > auth.RefreshExpires {
		panic("Auth refresh token has expired. Unable to refresh access token.")
	}

	if t > auth.AccessExpires {
		req := result.Wrap(
			http.NewRequest(
				http.MethodGet,
				fmt.Sprintf("%s/api/TokenAuth/RefreshToken", auth.Server),
				nil,
			),
		).Unwrap()

		query := req.URL.Query()
		query.Add("refreshToken", auth.RefreshToken)
		req.URL.RawQuery = query.Encode()

		resp := result.Wrap(auth.Client.Do(req)).Unwrap()
		defer resp.Body.Close()

		// A bit ugly, but we'll define our specific response content to decode
		var content struct {
			Result struct {
				AccessToken          string `json:"accessToken"`
				EncryptedAccessToken string `json:"encryptedAccessToken"`
				ExpireInSeconds      int64  `json:"expireInSeconds"`
			} `json:"result"`
		}
		result.Wrap(content, json.NewDecoder(resp.Body).Decode(&content)).Unwrap()

		auth.AccessToken = content.Result.AccessToken
		auth.AccessExpires = t + content.Result.ExpireInSeconds
	}

	return auth.AccessToken
}

func (auth Auth) Intercept(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.Token()))
	return nil
}
