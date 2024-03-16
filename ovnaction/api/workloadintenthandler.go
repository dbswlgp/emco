// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/ovnaction/pkg/module"

	"github.com/gorilla/mux"
)

var workloadIntJSONFile string = "json-schemas/network-workload.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type workloadintentHandler struct {
	// Interface that implements workload intent operations
	// We will set this variable with a mock interface for testing
	client moduleLib.WorkloadIntentManager
}

// Create handles creation of the Network entry in the database
func (h workloadintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var wi moduleLib.WorkloadIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	netControlIntent := vars["netControllerIntent"]

	err := json.NewDecoder(r.Body).Decode(&wi)

	switch {
	case err == io.EOF:
		log.Error(":: Empty workload intent POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding workload intent POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(workloadIntJSONFile, wi)
	if err != nil {
		log.Error(":: Invalid workload intent POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateWorkloadIntent(ctx, wi, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, wi, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create workload intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles creation/update of the Network entry in the database
func (h workloadintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var wi moduleLib.WorkloadIntent
	vars := mux.Vars(r)
	name := vars["workloadIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	netControlIntent := vars["netControllerIntent"]

	err := json.NewDecoder(r.Body).Decode(&wi)

	switch {
	case err == io.EOF:
		log.Error(":: Empty workload intent PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding workload intent PUT body ::", log.Fields{"Error": err, "Body": wi})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(workloadIntJSONFile, wi)
	if err != nil {
		log.Error(":: Invalid workload intent PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name in URL should match name in body
	if wi.Metadata.Name != name {
		log.Error(":: Mismatched network workload intent name in PUT request ::", log.Fields{"URL name": name, "Metadata name": wi.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateWorkloadIntent(ctx, wi, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, wi, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding update workload intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular Network Name
// Returns a Network
func (h workloadintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["workloadIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	netControlIntent := vars["netControllerIntent"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetWorkloadIntents(ctx, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
	} else {
		ret, err = h.client.GetWorkloadIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding get workload intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular Network  Name
func (h workloadintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["workloadIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	netControlIntent := vars["netControllerIntent"]

	err := h.client.DeleteWorkloadIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
