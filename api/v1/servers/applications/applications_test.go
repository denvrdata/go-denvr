package applications_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/denvrdata/go-denvr/api/v1/servers/applications"
	"github.com/denvrdata/go-denvr/result"
)

func TestClient(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"/api/TokenAuth/Authenticate",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write(
				[]byte(`{
					"result": {
						"accessToken": "access1",
						"refreshToken": "refresh",
						"expireInSeconds": 600,
						"refreshTokenExpireInSeconds": 3600
					}
				}`),
			)
		},
	)

	mux.HandleFunc(
		"/api/v1/servers/applications/GetConfigurations",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write(
				[]byte(`{
					"items": [
				 		{
					      "name": "g-nvidia-8xa100-80gb-sxm-198vcpu-970gb",
					      "description": "8x NVIDIA A100 SXM GPUs, 198 vCPUs, 970GB RAM",
					      "gpuCount": 8,
					      "gpuType": "nvidia.com/A100SXM480GB",
					      "gpuBrand": "NVIDIA",
					      "gpuName": "NVIDIA A100",
					      "vcpusCount": 198,
					      "memoryGb": 970,
					      "directAttachedStorageGb": 17000,
					      "pricePerHour": 0.0,
					      "clusters": [
					        "Hou1",
					        "Msc1"
					      ]
					    },
					    {
					      "name": "g-nvidia-8xh100-80gb-sxm-198vcpu-970gb",
					      "description": "8x NVIDIA H100 SXM GPUs, 198 vCPUs, 970GB RAM",
					      "gpuCount": 8,
					      "gpuType": "nvidia.com/H100SXM480GB",
					      "gpuBrand": "NVIDIA",
					      "gpuName": "NVIDIA H100",
					      "vcpusCount": 198,
					      "memoryGb": 970,
					      "directAttachedStorageGb": 20000,
					      "pricePerHour": 19.52,
					      "clusters": [
					        "Hou1",
					        "Msc1"
					      ]
					    }
					]
				}`),
			)
		},
	)
	server := httptest.NewServer(mux)
	defer server.Close()

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
		"TestClient",
		func(t *testing.T) {
			// Use the DENVR_CONFIG environment variable for our tests
			os.Setenv("DENVR_CONFIG", f.Name())

			c := applications.NewClient()
			// with default behaviour
			{
				resp, err := c.GetConfigurations(context.TODO())
				if err != nil {
					log.Fatal(err)
				}

				fmt.Printf("Response: %v\n", resp)
			}

			// with a raw http.Response
			{

				resp, err := c.GetConfigurationsRaw(context.TODO())
				if err != nil {
					log.Fatal(err)
				}
				if resp.StatusCode != http.StatusOK {
					log.Fatalf("Expected HTTP 200 but received %d", resp.StatusCode)

				}
			}

			t.Cleanup(func() { os.Unsetenv("DENVR_CONFIG") })
		},
	)
}
