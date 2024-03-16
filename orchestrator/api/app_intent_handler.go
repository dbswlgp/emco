// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

var appIntentJSONFile string = "json-schemas/generic-placement-intent-app.json"

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type appIntentHandler struct {
	client moduleLib.AppIntentManager
}

// createAppIntentHandler handles the create operation of intent
func (h appIntentHandler) createAppIntentHandler(w http.ResponseWriter, r *http.Request) {

	var a moduleLib.AppIntent

	err := json.NewDecoder(r.Body).Decode(&a)
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
	err, httpError := validation.ValidateJsonSchemaData(appIntentJSONFile, a)
	if err != nil {
		handleJsonSchemaValidationError(w, err, httpError)
		return
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	projectName := vars["project"]
	compositeAppName := vars["compositeApp"]
	version := vars["compositeAppVersion"]
	intent := vars["genericPlacementIntent"]
	digName := vars["deploymentIntentGroup"]

	appIntent, _, createErr := h.client.CreateAppIntent(ctx, a, projectName, compositeAppName, version, intent, digName, true)
	if createErr != nil {
		apiErr := apierror.HandleErrors(vars, createErr, a, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(appIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h appIntentHandler) getAppIntentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

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

	i := vars["genericPlacementIntent"]
	if i == "" {
		log.Error("Missing genericPlacementIntentName in GET request", log.Fields{})
		http.Error(w, "Missing genericPlacementIntentName in GET request", http.StatusBadRequest)
		return
	}

	dig := vars["deploymentIntentGroup"]
	if dig == "" {
		log.Error("Missing deploymentIntentGroupName in GET request", log.Fields{})
		http.Error(w, "Missing deploymentIntentGroupName in GET request", http.StatusBadRequest)
		return
	}

	ai := vars["genericAppPlacementIntent"]
	if ai == "" {
		log.Error("Missing appIntentName in GET request", log.Fields{})
		http.Error(w, "Missing appIntentName in GET request", http.StatusBadRequest)
		return
	}

	appIntent, err := h.client.GetAppIntent(ctx, ai, p, ca, v, i, dig)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(appIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

/*
getAllIntentsByAppHandler handles the URL:
/v2/project/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intent/{genericPlacementIntent}/app-intents?app-name=<app-name>
*/
func (h appIntentHandler) getAllIntentsByAppHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	vars := mux.Vars(r)
	pList := []string{"project", "compositeApp", "compositeAppVersion", "genericPlacementIntent"}
	err := validation.IsValidParameterPresent(vars, pList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	i := vars["genericPlacementIntent"]
	digName := vars["deploymentIntentGroup"]

	aN := r.URL.Query().Get("app-name")
	if aN == "" {
		log.Error("Missing appName in GET request", log.Fields{})
		http.Error(w, "Missing appName in GET request", http.StatusBadRequest)
		return
	}

	specData, err := h.client.GetAllIntentsByApp(ctx, aN, p, ca, v, i, digName)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(specData)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

/*
getAllAppIntentsHandler handles the URL:
/v2/project/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intent/{genericPlacementIntent}/app-intents
*/
func (h appIntentHandler) getAllAppIntentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	pList := []string{"project", "compositeApp", "compositeAppVersion", "genericPlacementIntent"}
	err := validation.IsValidParameterPresent(vars, pList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	i := vars["genericPlacementIntent"]
	digName := vars["deploymentIntentGroup"]

	applicationsAndClusterInfo, err := h.client.GetAllAppIntents(ctx, p, ca, v, i, digName)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(applicationsAndClusterInfo)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func (h appIntentHandler) deleteAppIntentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	i := vars["genericPlacementIntent"]
	ai := vars["genericAppPlacementIntent"]
	digName := vars["deploymentIntentGroup"]

	err := h.client.DeleteAppIntent(ctx, ai, p, ca, v, i, digName)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// putAppIntentHandler handles the put operation of intent
func (h appIntentHandler) putAppIntentHandler(w http.ResponseWriter, r *http.Request) {
	var ai moduleLib.AppIntent
	ctx := r.Context()
	vars := mux.Vars(r)
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	gpi := vars["genericPlacementIntent"]
	dig := vars["deploymentIntentGroup"]

	// Verify JSON Body
	err := json.NewDecoder(r.Body).Decode(&ai)
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

	err, code := validation.ValidateJsonSchemaData(appIntentJSONFile, ai)
	if err != nil {
		handleJsonSchemaValidationError(w, err, code)
		return
	}

	// Update App Intent
	appIntent, aiExists, err := h.client.CreateAppIntent(ctx, ai, p, ca, v, gpi, dig, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, ai, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	statusCode := http.StatusCreated
	if aiExists {
		// resource does have a current representation and that representation is successfully modified
		statusCode = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(appIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleJsonSchemaValidationError(w http.ResponseWriter, err error, status int) {
	log.Error(err.Error(), log.Fields{})
	if strings.Contains(err.Error(), "clusterProvider is required") {
		http.Error(w, "Missing clusterProvider in an intent", status)
		return
	}
	if strings.Contains(err.Error(), "cluster is required") {
		http.Error(w, "Missing cluster or clusterLabel", status)
		return
	}
	if strings.Contains(err.Error(), "Must not validate the schema (not)") {
		http.Error(w, "Only one of cluster name or cluster label allowed", status)
		return
	}
	if strings.Contains(err.Error(), "app is required") {
		http.Error(w, "Missing app for the intent", status)
		return
	}
	if strings.Contains(err.Error(), "name is required") {
		http.Error(w, "Missing name for the intent", status)
		return
	}
	// default
	http.Error(w, err.Error(), status)
}
