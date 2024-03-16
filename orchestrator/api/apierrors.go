package api

import (
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
)

var apiErrors = []apierror.APIError{
	{ID: "Project not found", Message: "Project not found", Status: http.StatusNotFound},
	{ID: "CompositeProfile not found", Message: "CompositeProfile not found", Status: http.StatusNotFound},
	{ID: "DeploymentIntentGroup not found", Message: "DeploymentIntentGroup not found", Status: http.StatusNotFound},
	{ID: "Intent not found", Message: "Intent not found", Status: http.StatusNotFound},
	{ID: "AppIntent not found", Message: "AppIntent not found", Status: http.StatusNotFound},
	{ID: "Cluster not found", Message: "Cluster not found", Status: http.StatusNotFound},
	{ID: "GenericPlacementIntent not found", Message: "GenericPlacementIntent not found", Status: http.StatusNotFound},
	{ID: "LogicalCloud not found", Message: "LogicalCloud not found", Status: http.StatusNotFound},
	{ID: "Controller not found", Message: "Controller not found", Status: http.StatusNotFound},
	{ID: "Intent already exists", Message: "Intent already exists", Status: http.StatusConflict},
	{ID: "AppIntent already exists", Message: "AppIntent already exists", Status: http.StatusConflict},
	{ID: "AppProfile already exists", Message: "AppProfile already exists", Status: http.StatusConflict},
	{ID: "AppProfile not found", Message: "AppProfile not found", Status: http.StatusNotFound},
	{ID: "App already has an AppProfile", Message: "App already has an AppProfile", Status: http.StatusConflict},
	{ID: "App already exists", Message: "App already exists", Status: http.StatusConflict},
	{ID: "App not found", Message: "App not found", Status: http.StatusNotFound},
	{ID: "CompositeApp already exists", Message: "CompositeApp already exists", Status: http.StatusConflict},
	{ID: "CompositeApp not found", Message: "CompositeApp not found", Status: http.StatusNotFound},
	{ID: "CompositeProfile already exists", Message: "CompositeProfile already exists", Status: http.StatusConflict},
	{ID: "DeploymentIntent already exists", Message: "DeploymentIntent already exists", Status: http.StatusConflict},
	{ID: "Project already exists", Message: "Project already exists", Status: http.StatusConflict},
	{ID: "Controller already exists", Message: "Controller already exists", Status: http.StatusConflict},
	{ID: "The DeploymentIntentGroup is not updated", Message: "The specified DeploymentIntentGroup is not in Created status", Status: http.StatusConflict},
	{ID: "AppDependency not found", Message: "AppDependency not found", Status: http.StatusNotFound},
}

var lcErrors = []apierror.APIError{
	{ID: "The specified Logical Cloud doesn't provide the mandatory clusters", Message: "The specified Logical Cloud doesn't provide the mandatory clusters", Status: http.StatusBadRequest},
	{ID: "Failed to obtain Logical Cloud specified", Message: "Failed to obtain Logical Cloud specified", Status: http.StatusBadRequest},
	{ID: "Logical Cloud is not currently applied", Message: "Logical Cloud is not currently applied", Status: http.StatusConflict},
	{ID: "Logical Cloud has never been applied", Message: "Logical Cloud has never been applied", Status: http.StatusConflict},
	{ID: "Error reading Logical Cloud context", Message: "Error reading Logical Cloud context", Status: http.StatusInternalServerError},
	{ID: "No Qualified Clusters to deploy App", Message: "No Qualified Clusters to deploy App", Status: http.StatusInternalServerError},
}
