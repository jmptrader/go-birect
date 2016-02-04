package birect

import "github.com/marcuswestin/go-errs"

var (
	// NewError creates an error with debugging information, such as stack traces, etc.
	NewError = errs.New

	// WrapError wraps an error with debugging information, such as stack traces, etc.
	WrapError = errs.Wrap

	// DefaultPublicErrorMessage will be set as the public error message for any error without one.
	DefaultPublicErrorMessage = "Oops! Something went wrong - please try again."
)
