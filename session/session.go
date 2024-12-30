package session

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/denvrdata/go-denvr/config"
	"github.com/denvrdata/go-denvr/result"
)

type Session struct {
	Config     config.Config
	HTTPClient *http.Client
}

func NewSession() Session {
	config := config.NewConfig()
	client := retryablehttp.NewClient()
	client.RetryMax = int(config.Retries)

	return Session{config, client.StandardClient()}
}

func (s Session) request(req *http.Request, content any) any {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.Config.Auth.Token()))
	resp := result.Wrap(s.HTTPClient.Do(req)).Unwrap()

	defer resp.Body.Close()

	// Simply passing the content to decode allows us to pass either:
	// - map[string]inferface{} or
	// - struct { ... }
	result.Wrap(content, json.NewDecoder(resp.Body).Decode(&content)).Unwrap()
	return content
}
