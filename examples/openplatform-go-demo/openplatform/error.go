package openplatform

import "fmt"

const (
	CodeSuccess       = 200
	CodeTokenInvalid  = 40001
	CodeQuotaExceeded = 40002
)

// OpenPlatformError is a gateway business error.
type OpenPlatformError struct {
	Code    int
	Message string
}

func (e *OpenPlatformError) Error() string {
	return fmt.Sprintf("openplatform error code=%d message=%s", e.Code, e.Message)
}

func (e *OpenPlatformError) IsTokenInvalid() bool   { return e.Code == CodeTokenInvalid }
func (e *OpenPlatformError) IsQuotaExceeded() bool { return e.Code == CodeQuotaExceeded }
