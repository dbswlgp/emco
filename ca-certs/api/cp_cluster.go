// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

// cpClusterHandler implements the clusterProvider caCert clusterGroup handler functions
type cpClusterHandler struct {
	manager clusterprovider.ClusterGroupManager
}

// handleClusterCreate handles the route for creating a new caCert clusterGroup
func (h *cpClusterHandler) handleClusterCreate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCluster(w, r)
}

// handleClusterDelete handles the route for deleting a caCert clusterGroup
func (h *cpClusterHandler) handleClusterDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// get the route variables
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.DeleteClusterGroup(ctx, vars.cert, vars.cluster, vars.clusterProvider); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleClusterGet handles the route for retrieving a caCert clusterGroup
func (h *cpClusterHandler) handleClusterGet(w http.ResponseWriter, r *http.Request) {
	var (
		clusters interface{}
		err      error
	)

	ctx := r.Context()
	// get the route variables
	vars := _cpVars(mux.Vars(r))
	if len(vars.cluster) == 0 {
		clusters, err = h.manager.GetAllClusterGroups(ctx, vars.cert, vars.clusterProvider)
	} else {
		clusters, err = h.manager.GetClusterGroup(ctx, vars.cert, vars.cluster, vars.clusterProvider)
	}

	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	sendResponse(w, clusters, http.StatusOK)
}

// handleClusterUpdate handles the route for updating a caCert clusterGroup
func (h *cpClusterHandler) handleClusterUpdate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCluster(w, r)
}

// createOrUpdateCluster create/update the caCert clusterGroup based on the request method
func (h *cpClusterHandler) createOrUpdateCluster(w http.ResponseWriter, r *http.Request) {
	var cluster module.ClusterGroup

	// validate the request body before storing it in the database
	if code, err := validateRequestBody(r.Body, &cluster, ClusterSchemaJson); err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	if len(cluster.Spec.Label) == 0 && len(cluster.Spec.Cluster) == 0 {
		err := "cluster label or cluster name must be provided"
		logutils.Error(err,
			logutils.Fields{
				"Cluster": cluster})
		http.Error(w, err, http.StatusBadRequest)
		return

	}

	ctx := r.Context()
	// get the route variables
	vars := _cpVars(mux.Vars(r))

	methodPost := false
	if r.Method == http.MethodPost {
		methodPost = true
	}

	if !methodPost {
		// name in the URL should match the name in the body
		if cluster.MetaData.Name != vars.cluster {
			err := "clusterGroup name is not matching with the name in the request"
			logutils.Error(err,
				logutils.Fields{
					"ClusterGroup": cluster})
			http.Error(w, err, http.StatusBadRequest)
			return
		}
	}

	clr, clusterExists, err := h.manager.CreateClusterGroup(ctx, cluster, vars.cert, vars.clusterProvider, methodPost)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	code := http.StatusCreated
	if clusterExists {
		// clusterGroup does have a current representation and that representation is successfully modified
		code = http.StatusOK
	}

	sendResponse(w, clr, code)
}
