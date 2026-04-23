package holded

import "errors"

var (
	ErrMissingClient = errors.New("api client not available in context")
	ErrWriteDisabled = errors.New("write operations are disabled (HOLDED_ALLOW_WRITE=false)")
)
