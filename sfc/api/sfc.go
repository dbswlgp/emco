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

var sfcJSONFile string = "json-schemas/sfc.json"

// Create handles creation of the SFC entry in the database
func (h sfcIntentHandler) createSfcHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sfc model.SfcIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]

	err := json.NewDecoder(r.Body).Decode(&sfc)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcJSONFile, sfc)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateSfcIntent(ctx, sfc, project, compositeApp, compositeAppVersion, deployIntentGroup, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, sfc, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create SFC Intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles update of the SFC entry in the database
func (h sfcIntentHandler) putSfcHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sfc model.SfcIntent
	vars := mux.Vars(r)
	name := vars["sfcIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]

	err := json.NewDecoder(r.Body).Decode(&sfc)

	switch {
	case err == io.EOF:
		log.Error(":: Empty SFC PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding SFC PUT body ::", log.Fields{"Error": err, "Body": sfc})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(sfcJSONFile, sfc)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name in URL should match name in body
	if sfc.Metadata.Name != name {
		log.Error(":: Mismatched SFC name in PUT request ::", log.Fields{"URL name": name, "Metadata name": sfc.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateSfcIntent(ctx, sfc, project, compositeApp, compositeAppVersion, deployIntentGroup, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, sfc, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding update SFC response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular SFC Name
// Returns an SfcIntent
func (h sfcIntentHandler) getSfcHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["sfcIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetAllSfcIntents(ctx, project, compositeApp, compositeAppVersion, deployIntentGroup)
	} else {
		ret, err = h.client.GetSfcIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup)
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
		log.Error(":: Error encoding get SFC Intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular SFC
func (h sfcIntentHandler) deleteSfcHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["sfcIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]

	err := h.client.DeleteSfcIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
