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

var addIntentJSONFile string = "json-schemas/deployment-intent.json"

type intentHandler struct {
	client moduleLib.IntentManager
}

// Add Intent in Deployment Group
func (h intentHandler) addIntentHandler(w http.ResponseWriter, r *http.Request) {
	var i moduleLib.Intent
	ctx := r.Context()
	err := json.NewDecoder(r.Body).Decode(&i)
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
	err, httpError := validation.ValidateJsonSchemaData(addIntentJSONFile, i)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	d := vars["deploymentIntentGroup"]

	intent, _, addError := h.client.AddIntent(ctx, i, p, ca, v, d, true)
	if addError != nil {
		apiErr := apierror.HandleErrors(vars, addError, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(intent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

/*
getIntentByNameHandler handles the URL
URL: /v2/projects/{p}/composite-apps/{compositeApp}/{version}/
deployment-intent-groups/{deploymentIntentGroup}/intents?intent=<intent>
*/
func (h intentHandler) getIntentByNameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	pList := []string{"project", "compositeApp", "compositeAppVersion", "deploymentIntentGroup"}
	err := validation.IsValidParameterPresent(vars, pList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	di := vars["deploymentIntentGroup"]

	iN := r.URL.Query().Get("intent")
	if iN == "" {
		log.Error("Missing appName in GET request", log.Fields{})
		http.Error(w, "Missing appName in GET request", http.StatusBadRequest)
		return
	}

	mapOfIntents, err := h.client.GetIntentByName(ctx, iN, p, ca, v, di)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(mapOfIntents)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

/*
getAllIntentsHandler handles the URL
URL: /v2/projects/{project}/composite-apps/{compositeApp}/{version}/
deployment-intent-groups/{deploymentIntentGroup}/intents
*/
func (h intentHandler) getAllIntentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	pList := []string{"project", "compositeApp", "compositeAppVersion", "deploymentIntentGroup"}
	err := validation.IsValidParameterPresent(vars, pList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	di := vars["deploymentIntentGroup"]

	mapOfIntents, err := h.client.GetAllIntents(ctx, p, ca, v, di)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(mapOfIntents)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h intentHandler) getIntentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	i := vars["groupIntent"]
	if i == "" {
		log.Error("Missing intentName in GET request", log.Fields{})
		http.Error(w, "Missing intentName in GET request", http.StatusBadRequest)
		return
	}

	p := vars["project"]
	if p == "" {
		log.Error("Missing projectName in GET request", log.Fields{})
		http.Error(w, "Missing projectName in GET request", http.StatusBadRequest)
		return
	}
	ca := vars["compositeApp"]
	if ca == "" {
		log.Error("Missing compositeAppName in GET request", log.Fields{})
		http.Error(w, "Missing compositeAppName in GET request", http.StatusBadRequest)
		return
	}

	v := vars["compositeAppVersion"]
	if v == "" {
		log.Error("Missing version of compositeApp in GET request", log.Fields{})
		http.Error(w, "Missing version of compositeApp in GET request", http.StatusBadRequest)
		return
	}

	di := vars["deploymentIntentGroup"]
	if di == "" {
		log.Error("Missing name of DeploymentIntentGroup in GET request", log.Fields{})
		http.Error(w, "Missing name of DeploymentIntentGroup in GET request", http.StatusBadRequest)
		return
	}

	intent, err := h.client.GetIntent(ctx, i, p, ca, v, di)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(intent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h intentHandler) deleteIntentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	i := vars["groupIntent"]
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	di := vars["deploymentIntentGroup"]

	err := h.client.DeleteIntent(ctx, i, p, ca, v, di)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// putIntentHandler handles the update operations on intent
func (h intentHandler) putIntentHandler(w http.ResponseWriter, r *http.Request) {
	var i moduleLib.Intent
	ctx := r.Context()
	vars := mux.Vars(r)
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	dig := vars["deploymentIntentGroup"]

	err := json.NewDecoder(r.Body).Decode(&i)
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
	err, httpError := validation.ValidateJsonSchemaData(addIntentJSONFile, i)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	intent, iExists, err := h.client.AddIntent(ctx, i, p, ca, v, dig, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	statusCode := http.StatusCreated
	if iExists {
		// resource does have a current representation and that representation is successfully modified
		statusCode = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(intent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
