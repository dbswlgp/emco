// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	"gitlab.com/project-emco/core/emco-base/src/sfcclient/pkg/model"
	"gitlab.com/project-emco/core/emco-base/src/sfcclient/pkg/module"

	"github.com/gorilla/mux"
)

var sfcClientJSONFile string = "json-schemas/sfc-client.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type sfcHandler struct {
	// Interface that implements workload intent operations
	// We will set this variable with a mock interface for testing
	client module.SfcManager
}

// Create handles creation of the SFC Client Intent entry in the database
func (h sfcHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sfcClient model.SfcClientIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]

	err := json.NewDecoder(r.Body).Decode(&sfcClient)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC Client Intent POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC Client Intent POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcClientJSONFile, sfcClient)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateSfcClientIntent(ctx, sfcClient, project, compositeApp, compositeAppVersion, deployIntentGroup, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, sfcClient, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create SFC Client Intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles update of the SFC Client Intent entry in the database
func (h sfcHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sfcClient model.SfcClientIntent
	vars := mux.Vars(r)
	name := vars["sfcClientIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]

	err := json.NewDecoder(r.Body).Decode(&sfcClient)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC Client Intent PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC Client Intent PUT body ::", log.Fields{"Error": err, "Body": sfcClient})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcClientJSONFile, sfcClient)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name in URL should match name in body
	if sfcClient.Metadata.Name != name {
		log.Error(":: Mismatched SFC Client Intent name in PUT request ::", log.Fields{"URL name": name, "Metadata name": sfcClient.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateSfcClientIntent(ctx, sfcClient, project, compositeApp, compositeAppVersion, deployIntentGroup, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, sfcClient, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding update SFC Client Intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular SFC Client Intent Name
// Returns an SfcIntent
func (h sfcHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["sfcClientIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetAllSfcClientIntents(ctx, project, compositeApp, compositeAppVersion, deployIntentGroup)
	} else {
		ret, err = h.client.GetSfcClientIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup)
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
		log.Error(":: Error encoding get SFC Client Intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular SFC Client Intent
func (h sfcHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["sfcClientIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]

	err := h.client.DeleteSfcClientIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
