package response_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/denvrdata/go-denvr/response"
)

// TestStruct represents a simple test structure with a data field
type TestStruct struct {
	Data *string `json:"data"`
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name        string
		response    *http.Response
		wantData    *string
		wantErr     bool
		errContains string
	}{
		{
			name: "empty response body",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte{})),
			},
			wantErr:     true,
			errContains: "unexpected end of JSON input",
		},
		{
			name: "invalid JSON response",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte("invalid json"))),
			},
			wantErr:     true,
			errContains: "invalid character",
		},
		{
			name: "valid JSON with success",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewReader([]byte(`{
					"result": {"data": "foo"},
					"success": true,
					"error": {"code": 0, "message": ""}
				}`))),
			},
			wantData: strPtr("foo"),
			wantErr:  false,
		},
		{
			name: "valid JSON with failure",
			response: &http.Response{
				StatusCode: http.StatusNotFound,
				Status:     "404 Not Found",
				Body: io.NopCloser(bytes.NewReader([]byte(`{
					"result": null,
					"success": false,
					"error": {
						"code": 404,
						"message": "The application 'my-missing-app' could not be found."
					}
				}`))),
			},
			wantErr:     true,
			errContains: "my-missing-app",
		},
		{
			name: "valid incorrect JSON with failure",
			response: &http.Response{
				StatusCode: http.StatusNotFound,
				Status:     "404 Not Found",
				Body: io.NopCloser(bytes.NewReader([]byte(`{
					"result": null,
					"success": false,
					"bad_error": {
						"code": 404,
						"message": "The application 'my-missing-app' could not be found."
					}
				}`))),
			},
			wantErr:     true,
			errContains: "404 Not Found",
		},
		{
			name: "direct JSON parsing fallback",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"data": "bar"}`))),
			},
			wantData: strPtr("bar"),
			wantErr:  false,
		},
		{
			name: "invalid direct JSON parsing fallback",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"foo": "bar"}`))),
			},
			wantData: nil,
			wantErr:  false,
		},
		{
			name: "error status with invalid JSON",
			response: &http.Response{
				StatusCode: http.StatusBadRequest,
				Status:     "400 Bad Request",
				Body:       io.NopCloser(bytes.NewReader([]byte(`invalid json`))),
			},
			wantErr:     true,
			errContains: "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := response.ParseResponse[TestStruct](tt.response)

			// Check error conditions
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseResponse() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ParseResponse() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			// Check success conditions
			if err != nil {
				t.Errorf("ParseResponse() unexpected error = %v", err)
				return
			}

			if result == nil {
				t.Errorf("ParseResponse() result is nil")
				return
			}

			// Compare the actual data with expected data
			if tt.wantData == nil && result.Data != nil {
				t.Errorf("ParseResponse() result.Data = %v, want nil", *result.Data)
			} else if tt.wantData != nil {
				if result.Data == nil {
					t.Errorf("ParseResponse() result.Data is nil, want %v", *tt.wantData)
				} else if *result.Data != *tt.wantData {
					t.Errorf("ParseResponse() result.Data = %v, want %v", *result.Data, *tt.wantData)
				}
			}
		})
	}
}

// Helper function to create a string pointer
func strPtr(s string) *string {
	return &s
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return s != "" && substr != "" && (s == substr || bytes.Contains([]byte(s), []byte(substr)))
}
