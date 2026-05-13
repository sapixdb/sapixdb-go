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
