// SPDX-License-Identifier: Apache-1.0
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

var sfcLinkJSONFile string = "json-schemas/sfc-link.json"

// Create handles creation of the SFC Link entry in the database
func (h sfcLinkIntentHandler) createLinkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sfcLink model.SfcLinkIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	sfcIntent := vars["sfcIntent"]

	err := json.NewDecoder(r.Body).Decode(&sfcLink)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC Link POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC Link POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcLinkJSONFile, sfcLink)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateSfcLinkIntent(ctx, sfcLink, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, sfcLink, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create SFC Link response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles update of the SFC Link entry in the database
func (h sfcLinkIntentHandler) putLinkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sfcLinkIntent model.SfcLinkIntent
	vars := mux.Vars(r)
	name := vars["sfcLink"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	sfcIntent := vars["sfcIntent"]

	err := json.NewDecoder(r.Body).Decode(&sfcLinkIntent)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC Link PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC Link PUT body ::", log.Fields{"Error": err, "Body": sfcLinkIntent})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcLinkJSONFile, sfcLinkIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name in URL should match name in body
	if sfcLinkIntent.Metadata.Name != name {
		log.Error(":: Mismatched SFC Link name in PUT request ::", log.Fields{"URL name": name, "Metadata name": sfcLinkIntent.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateSfcLinkIntent(ctx, sfcLinkIntent, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, sfcLinkIntent, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding update SFC Link response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular SFC Link Name
// Returns a SFC Link
func (h sfcLinkIntentHandler) getLinkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["sfcLink"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	sfcIntent := vars["sfcIntent"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetAllSfcLinkIntents(ctx, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent)
		if err != nil {
			apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}
	} else {
		ret, err = h.client.GetSfcLinkIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent)
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
		log.Error(":: Error encoding get SFC Link response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular SfcLink
func (h sfcLinkIntentHandler) deleteLinkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["sfcLink"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	sfcIntent := vars["sfcIntent"]

	err := h.client.DeleteSfcLinkIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup, sfcIntent)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
