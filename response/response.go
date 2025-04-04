package response

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ErrorResponse represents an error response from our Denvr API Server.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Response represents a generic response from our Denvr API Server which unwraps either a result or an error.
type Response[T any] struct {
	Result  *T             `json:"result"`
	Error   *ErrorResponse `json:"error"`
	Success bool           `json:"success"`
}

func ParseResponse[T any](rsp *http.Response) (*T, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	// First try to parse the response as a Response[T]
	var resp Response[T]
	err = json.Unmarshal(bodyBytes, &resp)
	if err != nil {
		return nil, err
	}

	// At this point we've either extracted the additona error message details or not
	// and should not proceed any further.
	if 400 <= rsp.StatusCode {
		if resp.Error == nil {
			return nil, fmt.Errorf(rsp.Status)
		} else {
			return nil, fmt.Errorf("%s - %s", rsp.Status, resp.Error.Message)
		}
	}

	// If we fail to parse the Response structure then we should fallback to parsing the passed in type.
	if resp.Result == nil {
		var fallback T
		if err = json.Unmarshal(bodyBytes, &fallback); err != nil {
			return nil, fmt.Errorf("Failed to parse response into %T: %v", fallback, err)
		}
		return &fallback, nil
	}

	return resp.Result, nil
}
