package aperture

// Code represents status of feature execution.
type Code uint8

// User passes a code to indicate status of feature execution.
//
//go:generate enumer -type=Code -output=code-string.go
const (
	// Ok indicates successful feature execution.
	Ok Code = iota
	// Error indicate error on feature execution.
	Error
)
