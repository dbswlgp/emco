// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"

	"github.com/gorilla/mux"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

var gpiJSONFile string = "json-schemas/generic-placement-intent.json"

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type genericPlacementIntentHandler struct {
	client moduleLib.GenericPlacementIntentManager
}

// createGenericPlacementIntentHandler handles the create operation of intent
func (h genericPlacementIntentHandler) createGenericPlacementIntentHandler(w http.ResponseWriter, r *http.Request) {

	var g moduleLib.GenericPlacementIntent

	err := json.NewDecoder(r.Body).Decode(&g)
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

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(gpiJSONFile, g)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	projectName := vars["project"]
	compositeAppName := vars["compositeApp"]
	version := vars["compositeAppVersion"]
	digName := vars["deploymentIntentGroup"]

	gPIntent, _, createErr := h.client.CreateGenericPlacementIntent(ctx, g, projectName, compositeAppName, version, digName, true)
	if createErr != nil {
		apiErr := apierror.HandleErrors(vars, createErr, g, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(gPIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getGenericPlacementHandler handles the GET operations on intent
func (h genericPlacementIntentHandler) getGenericPlacementHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	intentName := vars["genericPlacementIntent"]
	if intentName == "" {
		log.Error("Missing genericPlacementIntentName in GET request", log.Fields{})
		http.Error(w, "Missing genericPlacementIntentName in GET request", http.StatusBadRequest)
		return
	}
	projectName := vars["project"]
	if projectName == "" {
		log.Error("Missing projectName in GET request", log.Fields{})
		http.Error(w, "Missing projectName in GET request", http.StatusBadRequest)
		return
	}
	compositeAppName := vars["compositeApp"]
	if compositeAppName == "" {
		log.Error("Missing compositeAppName in GET request", log.Fields{})
		http.Error(w, "Missing compositeAppName in GET request", http.StatusBadRequest)
		return
	}

	version := vars["compositeAppVersion"]
	if version == "" {
		log.Error("Missing version in GET request", log.Fields{})
		http.Error(w, "Missing version in GET request", http.StatusBadRequest)
		return
	}

	dig := vars["deploymentIntentGroup"]
	if dig == "" {
		log.Error("Missing deploymentIntentGroupName in GET request", log.Fields{})
		http.Error(w, "Missing deploymentIntentGroupName in GET request", http.StatusBadRequest)
		return
	}

	gPIntent, err := h.client.GetGenericPlacementIntent(ctx, intentName, projectName, compositeAppName, version, dig)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(gPIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h genericPlacementIntentHandler) getAllGenericPlacementIntentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	pList := []string{"project", "compositeApp", "compositeAppVersion"}
	err := validation.IsValidParameterPresent(vars, pList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	digName := vars["deploymentIntentGroup"]

	gpList, err := h.client.GetAllGenericPlacementIntents(ctx, p, ca, v, digName)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(gpList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// deleteGenericPlacementHandler handles the delete operations on intent
func (h genericPlacementIntentHandler) deleteGenericPlacementHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	i := vars["genericPlacementIntent"]
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	digName := vars["deploymentIntentGroup"]

	err := h.client.DeleteGenericPlacementIntent(ctx, i, p, ca, v, digName)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// putGenericPlacementHandler handles the update operations on intent
func (h genericPlacementIntentHandler) putGenericPlacementHandler(w http.ResponseWriter, r *http.Request) {
	var gpi moduleLib.GenericPlacementIntent
	ctx := r.Context()
	vars := mux.Vars(r)
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	dig := vars["deploymentIntentGroup"]

	// Verify JSON Body
	err := json.NewDecoder(r.Body).Decode(&gpi)
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

	err, httpError := validation.ValidateJsonSchemaData(gpiJSONFile, gpi)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Update generic placement intent
	genericPlacementIntent, gpiExists, err := h.client.CreateGenericPlacementIntent(ctx, gpi, p, ca, v, dig, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, gpi, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	statusCode := http.StatusCreated
	if gpiExists {
		// resource does have a current representation and that representation is successfully modified
		statusCode = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(genericPlacementIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
