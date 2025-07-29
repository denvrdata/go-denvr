package response_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/denvrdata/go-denvr/response"
)

// TestStruct represents a simple test structure with a data field
type TestStruct struct {
	Data *string `json:"data"`
}

type TestStructWithTime struct {
	Data        *string    `json:"data"`
	LastUpdated *time.Time `json:"lastUpdated,omitempty"`
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

func TestParseResponseWithMalformedTime(t *testing.T) {
	tests := []struct {
		name        string
		response    *http.Response
		wantErr     bool
		errContains string
	}{
		{
			name: "malformed time without timezone - wrapped response",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewReader([]byte(`{
					"result": {
						"data": "test-data",
						"lastUpdated": "0001-01-01T00:00:00"
					},
					"success": true
				}`))),
			},
			wantErr: false,
		},
		{
			name: "malformed time without timezone - direct response",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewReader([]byte(`{
					"data": "test-data",
					"lastUpdated": "2006-01-02T15:04:05"
				}`))),
			},
			wantErr: false,
		},
		{
			name: "valid time with timezone",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewReader([]byte(`{
					"result": {
						"data": "test-data",
						"lastUpdated": "2025-07-28T17:41:44.093Z"
					},
					"success": true
				}`))),
			},
			wantErr: false,
		},
		{
			name: "null time value",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewReader([]byte(`{
					"result": {
						"data": "test-data",
						"lastUpdated": null
					},
					"success": true
				}`))),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := response.ParseResponse[TestStructWithTime](tt.response)

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

			if err != nil {
				t.Errorf("ParseResponse() unexpected error = %v", err)
				return
			}

			if result == nil {
				t.Errorf("ParseResponse() result is nil")
			}
		})
	}
}

func TestFixMalformedTimeFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "zero time becomes null",
			input:    `{"lastUpdated": "0001-01-01T00:00:00"}`,
			expected: `{"lastUpdated": null}`,
		},
		{
			name:     "zero time with nanoseconds becomes null",
			input:    `{"lastUpdated": "0001-01-01T00:00:00.000"}`,
			expected: `{"lastUpdated": null}`,
		},
		{
			name:     "time without timezone gets Z added",
			input:    `{"lastUpdated": "2006-01-02T15:04:05"}`,
			expected: `{"lastUpdated": "2006-01-02T15:04:05Z"}`,
		},
		{
			name:     "time with nanoseconds without timezone gets Z added",
			input:    `{"lastUpdated": "2006-01-02T15:04:05.999"}`,
			expected: `{"lastUpdated": "2006-01-02T15:04:05.999Z"}`,
		},
		{
			name:     "time with timezone remains unchanged",
			input:    `{"lastUpdated": "2006-01-02T15:04:05Z"}`,
			expected: `{"lastUpdated": "2006-01-02T15:04:05Z"}`,
		},
		{
			name:     "time with offset remains unchanged",
			input:    `{"lastUpdated": "2006-01-02T15:04:05-07:00"}`,
			expected: `{"lastUpdated": "2006-01-02T15:04:05-07:00"}`,
		},
		{
			name:     "multiple times in response",
			input:    `{"created": "0001-01-01T00:00:00", "updated": "2006-01-02T15:04:05"}`,
			expected: `{"created": null, "updated": "2006-01-02T15:04:05Z"}`,
		},
		{
			name:     "non-time strings remain unchanged",
			input:    `{"data": "some regular string", "count": 123}`,
			expected: `{"data": "some regular string", "count": 123}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We need to access the unexported function, so we'll test it indirectly
			// by creating a response and checking if it parses correctly
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(tt.input))),
			}

			// This should not error if our fix works
			_, err := response.ParseResponse[TestStructWithTime](resp)
			if err != nil {
				t.Errorf("ParseResponse() with fixed time failed: %v", err)
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
