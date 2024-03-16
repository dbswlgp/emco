package api

import (
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
)

var apiErrors = []apierror.APIError{
	{ID: "Consumer already exists", Message: "Consumer already exists", Status: http.StatusConflict},
	{ID: "Consumer not found", Message: "Consumer not found", Status: http.StatusNotFound},
	{ID: "Resource already exists", Message: "Resource already exists", Status: http.StatusConflict},
	{ID: "Resource not found", Message: "Resource not found", Status: http.StatusNotFound},
	{ID: "Intent already exists", Message: "Intent already exists", Status: http.StatusConflict},
	{ID: "Intent not found", Message: "Intent not found", Status: http.StatusNotFound},
}
