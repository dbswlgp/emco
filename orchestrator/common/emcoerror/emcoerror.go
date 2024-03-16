// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

/*
	Package emcoerror standardizes the error handling
*/
package emcoerror

// ErrorReason is an enumeration of potential failure reasons
// Each emco error type must be associated with an ErrorReason

type ErrorReason int

const (
	BadRequest ErrorReason = iota
	Conflict
	NotFound
	PreconditionFailed
	RequestTimeout
	Unknown
	UnprocessableEntity
	// Add additional reason(s)
)

const (
	InternalServerErrorMessage string = "The server encountered an internal error and was unable to complete your request"
	StateInfoNotFound          string = "Stateinfo not found"
	UnknownErrorMessage        string = "Unknown error"
)

// Type Error implements the emcoerror
type Error struct {
	error
	Message string
	Reason  ErrorReason
	Cause   *Error
	// Add additional properties(s)

}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Message + e.Cause.Error()
	}

	return e.Message
}

// NewEmcoError returns an instance of emco error
// constructed using the provided message, and reason
func NewEmcoError(message string, reason ErrorReason) *Error {
	return &Error{
		Message: message,
		Reason:  reason,
	}

}

// NewEmcoErrorWithCause returns an instance of emco error
// constructed using the provided message, reason, and cause
func NewEmcoErrorWithCause(message string, reason ErrorReason, cause *Error) *Error {
	return &Error{
		Message: message,
		Reason:  reason,
		Cause:   cause,
	}

}
