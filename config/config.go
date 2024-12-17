package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/denvrdata/go-denvr/auth"
	"github.com/denvrdata/go-denvr/result"
)

type Config struct {
	Auth    auth.Auth
	Server  string
	API     string
	Cluster string
	Tenant  string
	VPCId   string
	RPool   string
	Retries int64
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

	var content map[string]interface{}
	result.Wrap(toml.DecodeFile(path, &content)).Unwrap()

	// Use environment variables as our default
	credentials := struct {
		Username string
		Password string
	}{
		Username: os.Getenv("DENVR_USERNAME"),
		Password: os.Getenv("DENVR_PASSWORD"),
	}

	// Check if a credentials section even exists
	// TODO: This code seems ugly and should be placed in an accessor utility function
	if cred, ok := content["credentials"].(map[string]interface{}); ok {
		// Check if we need to load a username
		if credentials.Username == "" {
			if username, ok := cred["username"].(string); ok {
				credentials.Username = username
			} else {
				panic(fmt.Sprintf("Could not find username in \"DENVR_USERNAME\" or %s", path))
			}
		}
		// Check if we need to load a password
		if credentials.Password == "" {
			if password, ok := cred["password"].(string); ok {
				credentials.Password = password
			} else {
				panic(fmt.Sprintf("Could not find password in \"DENVR_PASSWORD\" or %s", path))
			}
		}
	}

	defaults := struct {
		Server  string
		API     string
		Cluster string
		Tenant  string
		VPCId   string
		RPool   string
		Retries int64
	}{
		Server:  "https://api.cloud.denvrdata.com",
		API:     "v1",
		Cluster: "Msc1",
		Tenant:  "",
		VPCId:   "",
		RPool:   "on-demand",
		Retries: 3,
	}
	if def, ok := content["defaults"].(map[string]interface{}); ok {
		if server, ok := def["server"].(string); ok {
			defaults.Server = server
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

	return Config{
		auth.NewAuth(defaults.Server, credentials.Username, credentials.Password),
		defaults.Server,
		defaults.API,
		defaults.Cluster,
		defaults.Tenant,
		defaults.VPCId,
		defaults.RPool,
		defaults.Retries,
	}
}
