package api

import (
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
)

var apiErrors = []apierror.APIError{
	{ID: "InboundClientsIntent already exists", Message: "InboundClientsIntent already exists", Status: http.StatusConflict},
	{ID: "Inbound clients intent not found", Message: "Inbound clients intent not found", Status: http.StatusNotFound},
	{ID: "ServerInboundIntent already exists", Message: "ServerInboundIntent already exists", Status: http.StatusConflict},
	{ID: "Inbound server intent not found", Message: "Inbound server intent not found", Status: http.StatusNotFound},
	{ID: "Traffic group intent not found", Message: "Traffic group intent not found", Status: http.StatusNotFound},
	{ID: "TrafficGroupIntent already exists", Message: "TrafficGroupIntent already exists", Status: http.StatusConflict},
	{ID: "Controller already exists", Message: "Controller already exists", Status: http.StatusConflict},
	{ID: "Controller not found", Message: "Controller not found", Status: http.StatusNotFound},
}
