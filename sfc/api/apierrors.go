package api

import (
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
)

var apiErrors = []apierror.APIError{
	{ID: "SFC Intent not found", Message: "SFC Intent not found", Status: http.StatusNotFound},
	{ID: "SFC Client Selector Intent not found", Message: "SFC Client Selector Intent not found", Status: http.StatusNotFound},
	{ID: "SFC Provider Network Intent not found", Message: "SFC Provider Network Intent not found", Status: http.StatusNotFound},
	{ID: "SFC Provider Network Intent already exists", Message: "SFC Provider Network Intent already exists", Status: http.StatusConflict},
	{ID: "SFC Client Selector Intent already exists", Message: "SFC Client Selector Intent already exists", Status: http.StatusConflict},
	{ID: "SFC Intent already exists", Message: "SFC Intent already exists", Status: http.StatusConflict},
}
