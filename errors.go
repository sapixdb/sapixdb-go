package sapixdb

import "fmt"

// SapixError is the base error for non-2xx responses.
type SapixError struct {
	Message string
	Status  int
	Code    string
}

func (e *SapixError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("sapixdb: %s (status=%d, code=%s)", e.Message, e.Status, e.Code)
	}
	return fmt.Sprintf("sapixdb: %s (status=%d)", e.Message, e.Status)
}

// SapixNetworkError is returned when the agent is unreachable.
type SapixNetworkError struct {
	Cause error
}

func (e *SapixNetworkError) Error() string {
	return fmt.Sprintf("sapixdb: network error: %v", e.Cause)
}

func (e *SapixNetworkError) Unwrap() error { return e.Cause }

// SapixNotFoundError is returned when a record does not exist.
type SapixNotFoundError struct {
	RecordID string
}

func (e *SapixNotFoundError) Error() string {
	return fmt.Sprintf("sapixdb: record not found: %s", e.RecordID)
}

// SapixRateLimitError is returned when an API key exceeds its configured
// requests-per-second limit (HTTP 429).
//
// Check for it by type-asserting the error returned by any SDK method:
//
//	_, err := orders.Write(ctx, data)
//	var rle *sapixdb.SapixRateLimitError
//	if errors.As(err, &rle) {
//	    delay := time.Second
//	    if rle.RPSLimit > 0 {
//	        delay = time.Second / time.Duration(rle.RPSLimit)
//	    }
//	    time.Sleep(delay)
//	    _, err = orders.Write(ctx, data) // retry once
//	}
type SapixRateLimitError struct {
	Message string
	// RPSLimit is the key's configured limit parsed from the server message, or 0 if not available.
	RPSLimit int
}

func (e *SapixRateLimitError) Error() string {
	if e.RPSLimit > 0 {
		return fmt.Sprintf("sapixdb: rate limit exceeded (%d req/s): %s", e.RPSLimit, e.Message)
	}
	return fmt.Sprintf("sapixdb: rate limit exceeded: %s", e.Message)
}
