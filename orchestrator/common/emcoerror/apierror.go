// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package emcoerror

import (
	"net/http"
)

// API error is intended for the API handlers
// It defines the HTTP status code and the explanation of the encountered error
type APIError struct {
	Message string
	Status  int
}

// Each ErrorReason must map to a single HTTP status code
var StatusCode = map[ErrorReason]int{
	// 4xx
	BadRequest:          http.StatusBadRequest,
	Conflict:            http.StatusConflict,
	NotFound:            http.StatusNotFound,
	PreconditionFailed:  http.StatusPreconditionFailed,
	RequestTimeout:      http.StatusRequestTimeout,
	UnprocessableEntity: http.StatusUnprocessableEntity,
	// 5xx
	Unknown: http.StatusInternalServerError,
}

// HandleAPIError returns the HTTP status code and the message
func HandleAPIError(err error) APIError {
	switch e := err.(type) {
	case *Error:
		if status, ok := StatusCode[e.Reason]; ok {
			return APIError{Message: e.Error(), Status: status}
		}
	}

	return APIError{Message: err.Error(), Status: http.StatusInternalServerError}
}
