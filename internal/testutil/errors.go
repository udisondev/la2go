package testutil

import "errors"

// ErrSimulated is a sentinel error for testing error handling paths
var ErrSimulated = errors.New("simulated error for testing")
