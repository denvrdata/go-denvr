package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/denvrdata/go-denvr/result"
)

type Auth interface {
	Intercept(ctx context.Context, req *http.Request) error
}

func NewAuth(path string, content map[string]any, server string, client *http.Client) Auth {
	// Use environment variables as our default
	credentials := struct {
		Apikey   string
		Username string
		Password string
	}{
		Apikey:   os.Getenv("DENVR_APIKEY"),
		Username: os.Getenv("DENVR_USERNAME"),
		Password: os.Getenv("DENVR_PASSWORD"),
	}

	// Check if a credentials section even exists
	// TODO: This code seems ugly and should be placed in an accessor utility function
	if cred, ok := content["credentials"].(map[string]any); ok {
		// Check if we need to load the Apikey
		if credentials.Apikey == "" {
			if apikey, ok := cred["apikey"].(string); ok {
				return ApiKey{apikey}
			}
		}
		// Check if we need to load a username
		if credentials.Username == "" {
			if username, ok := cred["username"].(string); ok {
				credentials.Username = username
			}
		}
		// Check if we need to load a password
		if credentials.Password == "" {
			if password, ok := cred["password"].(string); ok {
				credentials.Password = password
			}
		}
	}

	if credentials.Apikey != "" {
		return NewApiKey(credentials.Apikey)
	} else if credentials.Username != "" && credentials.Password != "" {
		return NewBearer(server, credentials.Username, credentials.Password, client)
	} else {
		panic(
			fmt.Sprintf(
				"Authentication failed. "+
					"Please provide credentials via environment variables or the [credentials] section in %s. "+
					"See https://github.com/denvrdata/go-denvr#configuration for more details.",
				path,
			),
		)
	}
}

type ApiKey struct {
	Key string
}

func NewApiKey(key string) ApiKey {
	return ApiKey{key}
}

func (auth ApiKey) Intercept(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", fmt.Sprintf("ApiKey %s", auth.Key))
	return nil
}

type Bearer struct {
	Server         string
	AccessToken    string
	RefreshToken   string
	AccessExpires  int64
	RefreshExpires int64
	Client         *http.Client
}

func NewBearer(server string, username string, password string, client *http.Client) Bearer {
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

	return Bearer{
		server,
		content.Result.AccessToken,
		content.Result.RefreshToken,
		time.Now().Unix() + content.Result.ExpireInSeconds,
		time.Now().Unix() + content.Result.RefreshTokenExpireInSeconds,
		client,
	}
}

func (auth Bearer) Token() string {
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

func (auth Bearer) Intercept(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.Token()))
	return nil
}
