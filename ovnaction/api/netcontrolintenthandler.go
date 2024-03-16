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

var netCntIntJSONFile string = "json-schemas/metadata.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type netcontrolintentHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client moduleLib.NetControlIntentManager
}

// Create handles creation of the NetControlIntent entry in the database
func (h netcontrolintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var nci moduleLib.NetControlIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]

	err := json.NewDecoder(r.Body).Decode(&nci)

	switch {
	case err == io.EOF:
		log.Error(":: Empty network control intent POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding network control intent POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(netCntIntJSONFile, nci)
	if err != nil {
		log.Error(":: Invalid network control intent body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateNetControlIntent(ctx, nci, project, compositeApp, compositeAppVersion, deployIntentGroup, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nci, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create network control intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles creation/update of the NetControlIntent entry in the database
func (h netcontrolintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var nci moduleLib.NetControlIntent
	vars := mux.Vars(r)
	name := vars["netControllerIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]

	err := json.NewDecoder(r.Body).Decode(&nci)

	switch {
	case err == io.EOF:
		log.Error(":: Empty network control intent PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding network control intent PUT body ::", log.Fields{"Error": err, "Body": nci})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(netCntIntJSONFile, nci)
	if err != nil {
		log.Error(":: Invalid network control intent body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name in URL should match name in body
	if nci.Metadata.Name != name {
		log.Error(":: Mismatched network control intent name in PUT request ::", log.Fields{"URL name": name, "Metadata name": nci.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateNetControlIntent(ctx, nci, project, compositeApp, compositeAppVersion, deployIntentGroup, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nci, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding update network control intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular NetControlIntent Name
// Returns a NetControlIntent
func (h netcontrolintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["netControllerIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetNetControlIntents(ctx, project, compositeApp, compositeAppVersion, deployIntentGroup)
	} else {
		ret, err = h.client.GetNetControlIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup)
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
		log.Error(":: Error encoding get network control intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular NetControlIntent  Name
func (h netcontrolintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["netControllerIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]

	err := h.client.DeleteNetControlIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
