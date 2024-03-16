package api

import (
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
)

var apiErrors = []apierror.APIError{
	{ID: "ClusterProvider already exists", Message: "ClusterProvider already exists", Status: http.StatusConflict},
	{ID: "Cluster provider not found", Message: "Cluster provider not found", Status: http.StatusNotFound},
	{ID: "Cluster already exists", Message: "Cluster already exists", Status: http.StatusConflict},
	{ID: "Cluster not found", Message: "Cluster not found", Status: http.StatusNotFound},
	{ID: "Cluster StateInfo not found", Message: "Cluster StateInfo not found", Status: http.StatusNotFound},
	{ID: "Cluster Label already exists", Message: "Cluster Label already exists", Status: http.StatusConflict},
	{ID: "Cluster label not found", Message: "Cluster label not found", Status: http.StatusNotFound},
	{ID: "Cluster KV Pair already exists", Message: "Cluster KV Pair already exists", Status: http.StatusConflict},
	{ID: "Cluster Sync Objects already exists", Message: "Cluster Sync Objects already exists", Status: http.StatusConflict},
	{ID: "Cluster key value pair not found", Message: "Cluster key value pair not found", Status: http.StatusNotFound},
	{ID: "Cluster sync object not found", Message: "Cluster sync object not found", Status: http.StatusNotFound},
	{ID: "ClmController already exists", Message: "ClmController already exists", Status: http.StatusConflict},
	{ID: "ClmController not found", Message: "ClmController not found", Status: http.StatusNotFound},
	{ID: "Cluster KV pair key value not found", Message: "Cluster KV pair key value not found", Status: http.StatusNotFound},
	{ID: "Cluster Sync Object key value not found", Message: "Cluster Sync Object key value not found", Status: http.StatusNotFound},
	{ID: "Error creating cloud config", Message: "Error creating cloud config", Status: http.StatusInternalServerError},
	{ID: "CLM failed publishing event", Message: "CLM failed publishing event to clm-controller", Status: http.StatusInternalServerError},
	{ID: "Error getting current state", Message: "Error getting current state from Cluster stateInfo", Status: http.StatusInternalServerError},
	{ID: "Cluster network intents must be terminated before it can be deleted", Message: "Cluster network intents must be terminated before it can be deleted", Status: http.StatusInternalServerError},
	{ID: "Network intents for cluster have not completed terminating", Message: "Network intents for cluster have not completed terminating", Status: http.StatusConflict},
	{ID: "Error deleting appcontext for Cluster", Message: "Error deleting appcontext for Cluster", Status: http.StatusInternalServerError},
}
