package api

import (
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
)

var apiErrors = []apierror.APIError{
	{ID: "SFC Client Intent already exists", Message: "SFC Client Intent already exists", Status: http.StatusConflict},
	{ID: "SFC Client Intent not found", Message: "SFC Client Intent not found", Status: http.StatusNotFound},
}
