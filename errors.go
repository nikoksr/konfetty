package konfetty

import "errors"

var (
	// ErrCircularReference is returned when a circular reference is detected in the config structure.
	ErrCircularReference = errors.New("circular reference detected")

	// ErrNilConfig is returned when the config passed to applyDefaults is nil.
	ErrNilConfig = errors.New("config cannot be nil")

	// ErrNotPointer is returned when the config passed to applyDefaults is not a pointer.
	ErrNotPointer = errors.New("config must be a pointer to a struct")
)
