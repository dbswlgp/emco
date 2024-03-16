// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/dcm/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
)

var clusterReferenceJSONValidation string = "json-schemas/cluster-reference.json"

// clusterHandler is used to store backend implementations objects
type clusterHandler struct {
	client module.ClusterManager
}

// createHandler handles creation of the cluster reference entry in the database
func (h clusterHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	project := vars["project"]
	logicalCloud := vars["logicalCloud"]
	var v common.Cluster
	var err error

	// Check logical cloud exists and is valid
	_, err = module.NewLogicalCloudClient().Get(ctx, project, logicalCloud)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&v)
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(clusterReferenceJSONValidation, v)
	if err != nil {
		log.Error(":: Invalid Cluster Reference JSON ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Cluster Reference Name is required.
	if v.MetaData.Name == "" {
		msg := "Missing name in POST request"
		log.Error(msg, log.Fields{})
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateCluster(ctx, project, logicalCloud, v)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, v, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getAllHandler handles GET operations over cluster references
// Returns a list of Cluster References
func (h clusterHandler) getAllHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	project := vars["project"]
	logicalCloud := vars["logicalCloud"]
	var ret interface{}
	var err error

	// Check logical cloud exists and is valid
	_, err = module.NewLogicalCloudClient().Get(ctx, project, logicalCloud)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	ret, err = h.client.GetAllClusters(ctx, project, logicalCloud)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getHandler handles GET operations on a particular name
// Returns a Cluster Reference
func (h clusterHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	project := vars["project"]
	logicalCloud := vars["logicalCloud"]
	name := vars["clusterReference"]
	var ret interface{}
	var err error

	// Check logical cloud exists and is valid
	_, err = module.NewLogicalCloudClient().Get(ctx, project, logicalCloud)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	ret, err = h.client.GetCluster(ctx, project, logicalCloud, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UpdateHandler handles Update operations on a particular cluster reference
func (h clusterHandler) updateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var v common.Cluster
	vars := mux.Vars(r)
	project := vars["project"]
	logicalCloud := vars["logicalCloud"]
	name := vars["clusterReference"]
	var err error

	// Check logical cloud exists and is valid
	_, err = module.NewLogicalCloudClient().Get(ctx, project, logicalCloud)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&v)
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(clusterReferenceJSONValidation, v)
	if err != nil {
		log.Error(":: Invalid Cluster Reference JSON ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if v.MetaData.Name == "" {
		log.Error("API: Missing name in PUT request", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.UpdateCluster(ctx, project, logicalCloud, name, v)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, v, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}
}

//deleteHandler handles DELETE operations on a particular record
func (h clusterHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	project := vars["project"]
	logicalCloud := vars["logicalCloud"]
	name := vars["clusterReference"]
	var err error

	// Check logical cloud exists and is valid
	_, err = module.NewLogicalCloudClient().Get(ctx, project, logicalCloud)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	err = h.client.DeleteCluster(ctx, project, logicalCloud, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getConfigHandler handles GET operations on kubeconfigs
// Returns a kubeconfig file
func (h clusterHandler) getConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	project := vars["project"]
	logicalCloud := vars["logicalCloud"]
	name := vars["clusterReference"]
	var err error

	// Check logical cloud exists and is valid
	_, err = module.NewLogicalCloudClient().Get(ctx, project, logicalCloud)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	_, err = h.client.GetCluster(ctx, project, logicalCloud, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	cfg, err := h.client.GetClusterConfig(ctx, project, logicalCloud, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, err = io.WriteString(w, cfg)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
