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
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/model"

	"github.com/gorilla/mux"
)

var sfcClientSelectorJSONFile string = "json-schemas/sfc-client-selector.json"

// Create handles creation of the SFC Client Selector entry in the database
func (h sfcClientSelectorIntentHandler) createClientSelectorHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sfcClientSelector model.SfcClientSelectorIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	sfcIntent := vars["sfcIntent"]

	err := json.NewDecoder(r.Body).Decode(&sfcClientSelector)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC Client Selector POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC Client Selector POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcClientSelectorJSONFile, sfcClientSelector)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateSfcClientSelectorIntent(ctx, sfcClientSelector, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, sfcClientSelector, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create SFC Client Selector response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles update of the SFC Client Selector entry in the database
func (h sfcClientSelectorIntentHandler) putClientSelectorHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sfcClientSelectorIntent model.SfcClientSelectorIntent
	vars := mux.Vars(r)
	name := vars["sfcClientSelector"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	sfcIntent := vars["sfcIntent"]

	err := json.NewDecoder(r.Body).Decode(&sfcClientSelectorIntent)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC Client Selector PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC Client Selector PUT body ::", log.Fields{"Error": err, "Body": sfcClientSelectorIntent})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcClientSelectorJSONFile, sfcClientSelectorIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name in URL should match name in body
	if sfcClientSelectorIntent.Metadata.Name != name {
		log.Error(":: Mismatched SFC Client Selector name in PUT request ::", log.Fields{"URL name": name, "Metadata name": sfcClientSelectorIntent.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateSfcClientSelectorIntent(ctx, sfcClientSelectorIntent, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, sfcClientSelectorIntent, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding update SFC Client Selector response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular SFC Client Selector Name
// Returns a SFC Client Selector
func (h sfcClientSelectorIntentHandler) getClientSelectorHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["sfcClientSelector"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	sfcIntent := vars["sfcIntent"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetAllSfcClientSelectorIntents(ctx, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent)
		if err != nil {
			apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}
	} else {
		ret, err = h.client.GetSfcClientSelectorIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent)
		if err != nil {
			apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding get SFC Client Selector response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular SfcClientSelector
func (h sfcClientSelectorIntentHandler) deleteClientSelectorHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["sfcClientSelector"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	sfcIntent := vars["sfcIntent"]

	err := h.client.DeleteSfcClientSelectorIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
