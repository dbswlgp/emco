// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (

	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"

	migrationModel "gitlab.com/project-emco/core/emco-base/src/migration/pkg/model"
)

/*
addMigrationIntentHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/migration-intents
*/
// Add Migration Intent in Deployment Group
func (h MigrationPlacementIntentHandler) addMigrationIntentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var migration migrationModel.DeploymentMigrationIntent
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: addMigrationIntentHandler .. start ::", log.Fields{"req": string(reqDump)})

	err := json.NewDecoder(r.Body).Decode(&migration)
	switch {
	case err == io.EOF:
		log.Error(":: addMigrationIntentHandler .. Empty POST body ::", log.Fields{"Error": err})
		http.Error(w, "addMigrationIntentHandler .. Empty POST body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: addMigrationIntentHandler .. Error decoding POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Validate json schema
	err, httpError := validation.ValidateJsonSchemaData(migrationIntentJSONFile, migration)
	if err != nil {
		log.Error(":: addMigrationIntentHandler .. JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	d := vars["deployment-intent-group-name"]

	log.Info(":: addMigrationIntentHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": d, "intent-name": migration.MetaData.Name})
	intent, err := h.client.AddIntent(ctx, migration, p, ca, v, d, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, migration, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(intent)
	if err != nil {
		log.Error(":: addMigrationIntentHandler ..  Encoder error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: addMigrationIntentHandler .. end ::", log.Fields{"intent": intent})
}

/*
getMigrationIntentByNameHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}?intent=<intent>
*/
// Query Migration Intent in Deployment Group
func (h MigrationPlacementIntentHandler) getMigrationIntentByNameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: getMigrationIntentByNameHandler .. start ::", log.Fields{"req": string(reqDump)})

	p, ca, v, di, _, err := parseMigrationIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	iN := r.URL.Query().Get("intent")
	if iN == "" {
		log.Error(":: getMigrationIntentByNameHandler .. Missing intent-name in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(w, "getMigrationIntentByNameHandler .. Missing intent-name in request", http.StatusBadRequest)
		return
	}

	intent, err := h.client.GetIntentByName(ctx, iN, p, ca, v, di)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(intent)
	if err != nil {
		log.Error(":: getMigrationIntentByNameHandler .. Encoder error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: getMigrationIntentByNameHandler .. end ::", log.Fields{"intent": intent})
}

/*
getMigrationIntentHandler/getMigrationIntentHandlers handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/migration-intents/{intent-name}
*/
// Get Migration Intent in Deployment Group
func (h MigrationPlacementIntentHandler) getMigrationIntentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: getMigrationIntentHandler .. start ::", log.Fields{"req": string(reqDump)})
	p, ca, v, di, name, err := parseMigrationIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: getMigrationIntentHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": name})

	var intents interface{}
	if len(name) == 0 {
		intents, err = h.client.GetAllIntents(ctx, p, ca, v, di)
	} else {
		intents, _, err = h.client.GetIntent(ctx, name, p, ca, v, di)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(intents)
	if err != nil {
		log.Error(":: getMigrationIntentHandler .. Encoder failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: getMigrationIntentHandler .. end ::", log.Fields{"intents": intents})
}

/*
putMigrationIntentHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/migration-intents/{intent-name}
*/
// Update Migration Intent in Deployment Group
func (h MigrationPlacementIntentHandler) putMigrationIntentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var migration migrationModel.DeploymentMigrationIntent
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: putMigrationIntentHandler .. start ::", log.Fields{"req": string(reqDump)})

	err := json.NewDecoder(r.Body).Decode(&migration)
	switch {
	case err == io.EOF:
		log.Error(":: putMigrationIntentHandler .. Empty PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: putMigrationIntentHandler .. decoding resource PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Validate json schema
	err, httpError := validation.ValidateJsonSchemaData(migrationIntentJSONFile, migration)
	if err != nil {
		log.Error(":: putMigrationIntentHandler .. JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Parse request parameters
	p, ca, v, di, name, err := parseMigrationIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if migration.MetaData.Name != name {
		log.Error(":: putMigrationIntentHandler .. Mismatched name in PUT request ::", log.Fields{"bodyname": migration.MetaData.Name, "name": name})
		http.Error(w, "putMigrationIntentHandler .. Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	log.Info(":: putMigrationIntentHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": name})
	intent, err := h.client.AddIntent(ctx, migration, p, ca, v, di, true)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, migration, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(intent)
	if err != nil {
		log.Error(":: putMigrationIntentHandler .. encoding failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: putMigrationIntentHandler .. end ::", log.Fields{"intent": intent})
}

/*
deleteMigrationIntentHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/migration-intents/{intent-name}
*/
// Delete Migration Intent in Deployment Group
func (h MigrationPlacementIntentHandler) deleteMigrationIntentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: deleteMigrationIntentHandler .. start ::", log.Fields{"req": string(reqDump)})

	// Parse request parameters
	p, ca, v, di, name, err := parseMigrationIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: deleteMigrationIntentHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": name})

	_, _, err = h.client.GetIntent(ctx, name, p, ca, v, di)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	err = h.client.DeleteIntent(ctx, name, p, ca, v, di)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Info(":: deleteMigrationIntentHandler .. end ::", log.Fields{"req": string(reqDump)})
}

/*
deleteAllMigrationIntentsHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/migration-intents
*/
// Delete all Migration Intents in Deployment Group
func (h MigrationPlacementIntentHandler) deleteAllMigrationIntentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: deleteAllMigrationIntentsHandler .. start ::", log.Fields{"req": string(reqDump)})

	// Parse request parameters
	p, ca, v, di, name, err := parseMigrationIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: deleteAllMigrationIntentsHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": name})

	migrationintents, err := h.client.GetAllIntents(ctx, p, ca, v, di)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	for _, migrationIntent := range migrationintents {
		err = h.client.DeleteIntent(ctx, migrationIntent.MetaData.Name, p, ca, v, di)
		if err != nil {
			apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
	log.Info(":: deleteAllMigrationIntentsHandler .. end ::", log.Fields{"req": string(reqDump)})
}

/* Parse Http request Parameters */
func parseMigrationIntentReqParameters(w *http.ResponseWriter, r *http.Request) (string, string, string, string, string, error) {
	vars := mux.Vars(r)

	i := vars["intent-name"]

	p := vars["project-name"]
	if p == "" {
		log.Error(":: parseMigrationIntentReqParameters ..  Missing projectName in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseMigrationIntentReqParameters .. Missing projectName in request", http.StatusBadRequest)
		return "", "", "", "", "", pkgerrors.New("Missing project-name")
	}
	ca := vars["composite-app-name"]
	if ca == "" {
		log.Error(":: parseMigrationIntentReqParameters ..  Missing compositeAppName in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseMigrationIntentReqParameters .. Missing compositeAppName in request", http.StatusBadRequest)
		return "", "", "", "", "", pkgerrors.New("Missing composite-app-name")
	}

	v := vars["composite-app-version"]
	if v == "" {
		log.Error(":: parseMigrationIntentReqParameters ..  version intentName in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseMigrationIntentReqParameters .. Missing version of compositeApp in request", http.StatusBadRequest)
		return "", "", "", "", "", pkgerrors.New("Missing composite-app-version")
	}

	di := vars["deployment-intent-group-name"]
	if di == "" {
		log.Error(":: parseMigrationIntentReqParameters ..  Missing DeploymentIntentGroup in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseMigrationIntentReqParameters .. Missing name of DeploymentIntentGroup in request", http.StatusBadRequest)
		return "", "", "", "", "", pkgerrors.New("Missing deployment-intent-group-name")
	}

	return p, ca, v, di, i, nil
}
