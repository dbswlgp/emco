// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/model"
)

var sfcProviderNetworkJSONFile string = "json-schemas/sfc-provider-network.json"

// Create handles creation of the SFC Provider Network entry in the database
func (h sfcProviderNetworkIntentHandler) createProviderNetworkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sfcProviderNetwork model.SfcProviderNetworkIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	sfcIntent := vars["sfcIntent"]

	err := json.NewDecoder(r.Body).Decode(&sfcProviderNetwork)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC Provider Network POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC Provider Network POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcProviderNetworkJSONFile, sfcProviderNetwork)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateSfcProviderNetworkIntent(ctx, sfcProviderNetwork, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, sfcProviderNetwork, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create SFC Provider Network response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles update of the SFC Provider Network entry in the database
func (h sfcProviderNetworkIntentHandler) putProviderNetworkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sfcProviderNetwork model.SfcProviderNetworkIntent
	vars := mux.Vars(r)
	name := vars["sfcProviderNetwork"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	sfcIntent := vars["sfcIntent"]

	err := json.NewDecoder(r.Body).Decode(&sfcProviderNetwork)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC Provider Network PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC Provider Network PUT body ::", log.Fields{"Error": err, "Body": sfcProviderNetwork})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcProviderNetworkJSONFile, sfcProviderNetwork)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name in URL should match name in body
	if sfcProviderNetwork.Metadata.Name != name {
		log.Error(":: Mismatched SFC Provider Network name in PUT request ::", log.Fields{"URL name": name, "Metadata name": sfcProviderNetwork.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateSfcProviderNetworkIntent(ctx, sfcProviderNetwork, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, sfcProviderNetwork, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding update SFC Provider Network response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular SFC Provider Network Name
// Returns a SFC Provider Network
func (h sfcProviderNetworkIntentHandler) getProviderNetworkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["sfcProviderNetwork"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	sfcIntent := vars["sfcIntent"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetAllSfcProviderNetworkIntents(ctx, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent)
	} else {
		ret, err = h.client.GetSfcProviderNetworkIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent)
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
		log.Error(":: Error encoding get SFC Provider Network response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular SfcProviderNetwork
func (h sfcProviderNetworkIntentHandler) deleteProviderNetworkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["sfcProviderNetwork"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	sfcIntent := vars["sfcIntent"]

	err := h.client.DeleteSfcProviderNetworkIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
