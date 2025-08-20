package config

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/denvrdata/go-denvr/auth"
	"github.com/denvrdata/go-denvr/result"
	"github.com/hashicorp/go-retryablehttp"
)

type Config struct {
	Auth    auth.Auth
	Server  string
	API     string
	Cluster string
	Tenant  string
	VPCId   string
	RPool   string
	Client  *http.Client
}

func NewConfig(paths ...string) Config {
	// Default config file location as our fallback
	path := filepath.Join(
		result.Wrap(os.UserHomeDir()).Unwrap(), ".config", "denvr.toml",
	)

	if len(paths) > 1 {
		// Error if we're given more than 1 path
		panic("NewConfig only accepts 0 or 1 argument, representing the config path.")
	} else if len(paths) > 0 {
		// Extract the explicity config path as the highest priority option if given
		path = paths[0]
	} else if os.Getenv("DENVR_CONFIG") != "" {
		// Seting the env is the next highest priority
		path = os.Getenv("DENVR_CONFIG")
	}

	var content map[string]any
	result.Wrap(toml.DecodeFile(path, &content)).Unwrap()

	defaults := struct {
		Server  string
		API     string
		Cluster string
		Tenant  string
		VPCId   string
		RPool   string
		Retries int64
	}{
		Server:  "https://api.cloud.denvrdata.com/",
		API:     "v1",
		Cluster: "Msc1",
		Tenant:  "",
		VPCId:   "",
		RPool:   "on-demand",
		Retries: 3,
	}
	if def, ok := content["defaults"].(map[string]any); ok {
		if server, ok := def["server"].(string); ok {
			// Trim any trailing slashes to make sure the paths work
			defaults.Server = strings.Trim(server, "/")
		}
		if api, ok := def["api"].(string); ok {
			defaults.API = api
		}
		if cluster, ok := def["cluster"].(string); ok {
			defaults.Cluster = cluster
		}
		if tenant, ok := def["tenant"].(string); ok {
			defaults.Tenant = tenant
		} else {
			panic(fmt.Sprintf("A tenant value must be specified in the config %s", path))
		}
		if vpcid, ok := def["vpcid"].(string); ok {
			defaults.VPCId = vpcid
		} else {
			defaults.VPCId = defaults.Tenant
		}
		if rpool, ok := def["rpool"].(string); ok {
			defaults.RPool = rpool
		}
		if retries, ok := def["retries"].(int64); ok {
			defaults.Retries = retries
		}
	}

	// Create a retryable HTTP client for use both in our auth code and the API client code.
	client := retryablehttp.NewClient()
	client.RetryMax = int(defaults.Retries)
	client.RetryWaitMin = 2 * time.Second
	client.RetryWaitMax = 60 * time.Second
	client.Backoff = retryablehttp.DefaultBackoff
	client.HTTPClient.Timeout = 60 * time.Second

	return Config{
		auth.NewAuth(path, content, defaults.Server, client.StandardClient()),
		defaults.Server,
		defaults.API,
		defaults.Cluster,
		defaults.Tenant,
		defaults.VPCId,
		defaults.RPool,
		client.StandardClient(),
	}
}
