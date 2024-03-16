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
	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"

	"github.com/gorilla/mux"
)

var caprofileJSONFile string = "json-schemas/metadata.json"

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type compositeProfileHandler struct {
	client moduleLib.CompositeProfileManager
}

// createCompositeProfileHandler handles the create operation of intent
func (h compositeProfileHandler) createHandler(w http.ResponseWriter, r *http.Request) {

	var cpf moduleLib.CompositeProfile

	err := json.NewDecoder(r.Body).Decode(&cpf)
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
	err, httpError := validation.ValidateJsonSchemaData(caprofileJSONFile, cpf)
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

	cProf, createErr := h.client.CreateCompositeProfile(ctx, cpf, projectName, compositeAppName, version, false)
	if createErr != nil {
		apiErr := apierror.HandleErrors(vars, createErr, cpf, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(cProf)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getHandler handles the GET operations on CompositeProfile
func (h compositeProfileHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	cProfName := vars["compositeProfile"]

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

	// handle the get all composite profile case
	if len(cProfName) == 0 {
		var retList []moduleLib.CompositeProfile

		ret, err := h.client.GetCompositeProfiles(ctx, projectName, compositeAppName, version)
		if err != nil {
			apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}

		for _, cl := range ret {
			retList = append(retList, moduleLib.CompositeProfile{Metadata: cl.Metadata})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(retList)
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	cProf, err := h.client.GetCompositeProfile(ctx, cProfName, projectName, compositeAppName, version)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(cProf)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// deleteHandler handles the delete operations on CompostiteProfile
func (h compositeProfileHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	c := vars["compositeProfile"]
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]

	err := h.client.DeleteCompositeProfile(ctx, c, p, ca, v)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h compositeProfileHandler) updateHandler(w http.ResponseWriter, r *http.Request) {

	var cpf moduleLib.CompositeProfile

	err := json.NewDecoder(r.Body).Decode(&cpf)
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
	err, httpError := validation.ValidateJsonSchemaData(caprofileJSONFile, cpf)
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

	cProf, createErr := h.client.CreateCompositeProfile(ctx, cpf, projectName, compositeAppName, version, true)
	if createErr != nil {
		apiErr := apierror.HandleErrors(vars, createErr, cpf, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(cProf)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
