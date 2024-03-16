package api

import (
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
)

var apiErrors = []apierror.APIError{
	{ID: "Error getting current state", Message: "Error getting current state from Cluster stateInfo", Status: http.StatusInternalServerError},
	{ID: "Cluster is in an invalid state", Message: "Cluster is in an invalid state", Status: http.StatusConflict},
	{ID: "Existing cluster network intents must be terminated before creating", Message: "Existing cluster network intents must be terminated before creating", Status: http.StatusBadRequest},
	{ID: "Network already exists", Message: "Network already exists", Status: http.StatusConflict},
	{ID: "Network not found", Message: "Network not found", Status: http.StatusNotFound},
	{ID: "Cluster network intents must be terminated before deleting", Message: "Cluster network intents must be terminated before deleting", Status: http.StatusConflict},
	{ID: "Existing cluster provider network intents must be terminated before creating", Message: "Existing cluster provider network intents must be terminated before creating", Status: http.StatusBadRequest},
	{ID: "ProviderNet already exists", Message: "ProviderNet already exists", Status: http.StatusConflict},
	{ID: "ProviderNet not found", Message: "ProviderNet not found", Status: http.StatusNotFound},
	{ID: "Cluster provider network intents must be terminated before deleting", Message: "Cluster provider network intents must be terminated before deleting", Status: http.StatusConflict},
	{ID: "Could not get find rsync by name", Message: "Could not get find rsync by name", Status: http.StatusNotFound},
	{ID: "Error creating AppContext", Message: "Error creating AppContext", Status: http.StatusInternalServerError},
	{ID: "Error creating AppContext CompositeApp", Message: "Error creating AppContext CompositeApp", Status: http.StatusInternalServerError},
	{ID: "Error adding App to AppContext", Message: "Error adding App to AppContext", Status: http.StatusInternalServerError},
	{ID: "Error adding network intent app order instruction", Message: "Error adding network intent app order instruction", Status: http.StatusInternalServerError},
	{ID: "Error adding network intent app dependency instruction", Message: "Error adding network intent app dependency instruction", Status: http.StatusInternalServerError},
	{ID: "Error adding Cluster to AppContext", Message: "Error adding Cluster to AppContext", Status: http.StatusInternalServerError},
	{ID: "Error finding StateInfo for cluster", Message: "Error finding StateInfo for cluster", Status: http.StatusInternalServerError},
	{ID: "Cluster state not found", Message: "Cluster state not found", Status: http.StatusNotFound},
	{ID: "Error updating the stateInfo of cluster", Message: "Error updating the stateInfo of cluster", Status: http.StatusInternalServerError},
	{ID: "Cluster network intents are not instantiating or terminating", Message: "Cluster network intents are not instantiating or terminating", Status: http.StatusInternalServerError},
	{ID: "Cluster network intents have not been applied", Message: "Cluster network intents have not been applied", Status: http.StatusConflict},
	{ID: "Cluster network intents termination already stopped", Message: "Cluster network intents termination already stopped", Status: http.StatusConflict},
	{ID: "Cluster network intents instantiation already stopped", Message: "Cluster network intents instantiation already stopped", Status: http.StatusConflict},
}
