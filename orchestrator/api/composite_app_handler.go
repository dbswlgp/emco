// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

var caJSONFile string = "json-schemas/composite-app.json"

// compositeAppHandler to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type compositeAppHandler struct {
	// Interface that implements CompositeApp operations
	// We will set this variable with a mock interface for testing
	client moduleLib.CompositeAppManager
}

// createHandler handles creation of the CompositeApp entry in the database
func (h compositeAppHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var c moduleLib.CompositeApp

	err := json.NewDecoder(r.Body).Decode(&c)
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
	err, httpError := validation.ValidateJsonSchemaData(caJSONFile, c)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	projectName := vars["project"]

	ret, err := h.client.CreateCompositeApp(ctx, c, projectName, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, c, apiErrors)
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

// getHandler handles GET operations on a particular CompositeApp Name
// Returns a compositeApp
func (h compositeAppHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["compositeApp"]
	version := vars["compositeAppVersion"]
	projectName := vars["project"]

	ret, err := h.client.GetCompositeApp(ctx, name, version, projectName)
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

// getAllCompositeAppsHandler handles the GetAllComppositeApps, returns a list of compositeApps under a project
func (h compositeAppHandler) getAllCompositeAppsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	pName := vars["project"]

	var caList []moduleLib.CompositeApp

	cApps, err := h.client.GetAllCompositeApps(ctx, pName)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	for _, cApp := range cApps {
		caList = append(caList, moduleLib.CompositeApp{Metadata: cApp.Metadata, Spec: cApp.Spec})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(caList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	return
}

// deleteHandler handles DELETE operations on a particular CompositeApp Name
func (h compositeAppHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["compositeApp"]
	version := vars["compositeAppVersion"]
	projectName := vars["project"]

	err := h.client.DeleteCompositeApp(ctx, name, version, projectName)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h compositeAppHandler) updateHandler(w http.ResponseWriter, r *http.Request) {
	var c moduleLib.CompositeApp

	err := json.NewDecoder(r.Body).Decode(&c)
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
	err, httpError := validation.ValidateJsonSchemaData(caJSONFile, c)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	projectName := vars["project"]

	ret, err := h.client.CreateCompositeApp(ctx, c, projectName, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, c, apiErrors)
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
