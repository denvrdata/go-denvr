package response

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
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

	// Try and fix malformed time formats prior to calling JSON unmarshalling
	bodyBytes = fixMalformedTimeFormats(bodyBytes)

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

// fixMalformedTimeFormats fixes common time format issues in JSON before unmarshaling
// https://github.com/denvrdata/DenvrDashboard/issues/3048
// TODO: This seems kinda ugly, maybe there's a better way by reusing the existing definition?
func fixMalformedTimeFormats(data []byte) []byte {
	// Pattern to match timestamps like "0001-01-01T00:00:00" (zero time without timezone)
	// and replace them with null since this typically indicates missing data
	zeroTimeRegex := regexp.MustCompile(`"0001-01-01T00:00:00(?:\.\d+)?"`)
	data = zeroTimeRegex.ReplaceAll(data, []byte("null"))

	// Pattern to match timestamps that need timezone added
	// Only matches timestamps that end with seconds/nanoseconds followed immediately by quote
	// (no existing timezone suffix like Z, +05:00, -07:00, etc.)
	noTimezoneRegex := regexp.MustCompile(`"(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?)"`)

	// Replace each match, but only if it doesn't already have timezone info
	result := noTimezoneRegex.ReplaceAllFunc(data, func(match []byte) []byte {
		str := string(match)
		// Check if this timestamp already has timezone info by looking at what comes before the closing quote
		if str[len(str)-2] == 'Z' || (len(str) > 7 && (str[len(str)-7] == '+' || str[len(str)-7] == '-')) {
			return match // Already has timezone, don't modify
		}
		// Add Z timezone
		return []byte(str[:len(str)-1] + `Z"`)
	})

	return result
}
