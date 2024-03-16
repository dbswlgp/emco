package api

import (
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
)

var apiErrors = []apierror.APIError{
	{ID: "NetControlIntent already exists", Message: "NetControlIntent already exists", Status: http.StatusConflict},
	{ID: "Net Control Intent not found", Message: "Net Control Intent not found", Status: http.StatusNotFound},
	{ID: "WorkloadIfIntent already exists", Message: "WorkloadIfIntent already exists", Status: http.StatusConflict},
	{ID: "WorkloadIfIntent not found", Message: "WorkloadIfIntent not found", Status: http.StatusNotFound},
	{ID: "WorkloadIntent already exists", Message: "WorkloadIntent already exists", Status: http.StatusConflict},
	{ID: "WorkloadIntent not found", Message: "WorkloadIntent not found", Status: http.StatusNotFound},
}
