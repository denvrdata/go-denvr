package virtual_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/denvrdata/go-denvr/api/v1/servers/virtual"
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
		"/api/v1/servers/virtual/GetConfigurations",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			resp.Write(
				[]byte(`{
					"items": [
						{
							"id": 5,
					      	"user_friendly_name": "A100_40GB_PCIe_1x",
						    "name": "A100_40GB_PCIe_1x",
						    "description": null,
						    "os_version": "20.04",
						    "os_type": "Ubuntu",
						    "storage": 1700,
						    "gpu_type": "nvidia.com/A100PCIE40GB",
						    "gpu_family": "NVIDIA A100",
						    "gpu_brand": "Nvidia",
						    "gpu_name": "A100 40GB PCIe",
						    "type": "nvidia.com/A100PCIE40GB",
						    "brand_family": "NVIDIA A100",
						    "brand": "Nvidia",
						    "text_name": "A100 40GB PCIe",
						    "gpus": 1,
						    "vcpus": 14,
						    "memory": 112,
						    "price": 2.05,
						    "compute_network": null,
						    "is_gpu_platform": true,
							"clusters": [
					        	"Hou1",
						        "Msc1"
							]
					    },
					    {
						    "id": 6,
						    "user_friendly_name": "A100_40GB_PCIe_2x",
						    "name": "A100_40GB_PCIe_2x",
						    "description": null,
						    "os_version": "20.04",
						    "os_type": "Ubuntu",
						    "storage": 3400,
						    "gpu_type": "nvidia.com/A100PCIE40GB",
						    "gpu_family": "NVIDIA A100",
						    "gpu_brand": "Nvidia",
						    "gpu_name": "A100 40GB PCIe",
						    "type": "nvidia.com/A100PCIE40GB",
						    "brand_family": "NVIDIA A100",
						    "brand": "Nvidia",
						    "text_name": "A100 40GB PCIe",
						    "gpus": 2,
						    "vcpus": 28,
						    "memory": 224,
						    "price": 4.1,
						    "compute_network": null,
						    "is_gpu_platform": true,
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

			c := virtual.NewClient()
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
