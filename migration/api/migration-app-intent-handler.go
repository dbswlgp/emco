// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"fmt"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"

	migrationModel "gitlab.com/project-emco/core/emco-base/src/migration/pkg/model"
)

/*
addMigrationAppIntentHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/migration-intents/{intent-name}/app-intents
*/
// Add Migration App Intent
func (h MigrationPlacementIntentHandler) addMigrationAppIntentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var migration migrationModel.MigrationAppIntent
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: addMigrationAppIntentHandler .. start ::", log.Fields{"req": string(reqDump)})

	err := json.NewDecoder(r.Body).Decode(&migration)
	switch {
	case err == io.EOF:
		log.Error(":: addMigrationAppIntentHandler ..Empty POST body ::", log.Fields{"Error": err})
		http.Error(w, "addMigrationAppIntentHandler .. Empty POST body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: addMigrationAppIntentHandler .. Error decoding POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(migrationAppIntentJSONFile, migration)
	if err != nil {
		log.Error(":: addMigrationAppIntentHandler .. JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]
	i := vars["intent-name"]

	log.Info(":: AddAppIntent .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "app-intent-name": migration.MetaData.Name})
	consumer, err := h.client.AddAppIntent(ctx, migration, p, ca, v, di, i, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, migration, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(consumer)
	if err != nil {
		log.Error(":: addMigrationAppIntentHandler .. Encoder error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: addMigrationAppIntentHandler .. end ::", log.Fields{"consumer": consumer})
}

/*
getMigrationAppIntentByNameHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/migration-intents/{intent-name}/app-intents?app-intent-name=<app-intent-name>
*/
// Query Migration App Intent
func (h MigrationPlacementIntentHandler) getMigrationAppIntentByNameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: getMigrationAppIntentHandlerByName .. start ::", log.Fields{"req": string(reqDump)})

	p, ca, v, di, i, _, err := parseMigrationAppIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cN := r.URL.Query().Get("consumer")
	if cN == "" {
		log.Error(":: getMigrationAppIntentHandlerByName .. Missing intent-name in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(w, "getMigrationAppIntentHandlerByName .. Missing intent-name in request", http.StatusBadRequest)
		return
	}

	consumer, err := h.client.GetAppIntentByName(ctx, cN, p, ca, v, di, i)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(consumer)
	if err != nil {
		log.Error(":: getMigrationAppIntentHandlerByName .. Encoder error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: getMigrationAppIntentHandlerByName .. end ::", log.Fields{"consumer": consumer})
}

/*
getMigrationAppIntentHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/migration-intents/{intent-name}/app-intents/{app-intent-name}
*/
// Get Migration App Intent
func (h MigrationPlacementIntentHandler) getMigrationAppIntentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: getMigrationAppIntentHandler .. start ::", log.Fields{"req": string(reqDump)})
	p, ca, v, di, i, name, err := parseMigrationAppIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: getMigrationAppIntentHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": name})

	var consumers interface{}
	if len(name) == 0 {
		consumers, err = h.client.GetAllAppIntents(ctx, p, ca, v, di, i)
	} else {
		consumers, _, err = h.client.GetAppIntent(ctx, name, p, ca, v, di, i)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(consumers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: getMigrationAppIntentHandler .. end ::", log.Fields{"consumers": consumers})
}

/*
putMigrationAppIntentHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/migration-intents/{intent-name}/app-intents/{app-intent-name}
*/
// Update Migration App Intent
func (h MigrationPlacementIntentHandler) putMigrationAppIntentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var migration migrationModel.MigrationAppIntent
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: putMigrationAppIntentHandler .. start ::", log.Fields{"req": string(reqDump)})

	err := json.NewDecoder(r.Body).Decode(&migration)
	switch {
	case err == io.EOF:
		log.Error(":: putMigrationAppIntentHandler .. Empty  body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: putMigrationAppIntentHandler .. Error decoding resource body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(migrationAppIntentJSONFile, migration)
	if err != nil {
		log.Error(":: putMigrationAppIntentHandler .. JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Parse request parameters
	p, ca, v, di, i, name, err := parseMigrationAppIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if migration.MetaData.Name != name {
		log.Error(":: putMigrationAppIntentHandler ..Mismatched name in PUT request ::", log.Fields{"bodyname": migration.MetaData.Name, "name": name})
		http.Error(w, "putMigrationAppIntentHandler ..Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	log.Info(":: putMigrationAppIntentHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": name})
	consumer, err := h.client.AddAppIntent(ctx, migration, p, ca, v, di, i, true)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, migration, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(consumer)
	if err != nil {
		log.Error(":: putMigrationAppIntentHandler .. encoding failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: putMigrationAppIntentHandler .. end ::", log.Fields{"consumer": consumer})
}

/*
deleteMigrationAppIntentHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/migration-intents/{intent-name}/app-intents/{app-intent-name}
*/
// Delete Migration App Intent
func (h MigrationPlacementIntentHandler) deleteMigrationAppIntentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: deleteMigrationAppIntentHandler .. start ::", log.Fields{"req": string(reqDump)})

	// Parse request parameters
	p, ca, v, di, i, name, err := parseMigrationAppIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: deleteMigrationAppIntentHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "app-intent-name": name})

	_, _, err = h.client.GetAppIntent(ctx, name, p, ca, v, di, i)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	err = h.client.DeleteAppIntent(ctx, name, p, ca, v, di, i)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Info(":: deleteMigrationAppIntentHandler .. end ::", log.Fields{"req": string(reqDump)})
}

/*
deleteAllMigrationAppIntentsHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/migration-intents/{intent-name}/app-intents
*/
// Delete all Migration App Intents
func (h MigrationPlacementIntentHandler) deleteAllMigrationAppIntentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: deleteAllMigrationAppIntentsHandler .. start ::", log.Fields{"req": string(reqDump)})

	// Parse request parameters
	p, ca, v, di, i, name, err := parseMigrationAppIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: deleteAllMigrationAppIntentsHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "app-intent-name": name})

	migrationAppIntents, err := h.client.GetAllAppIntents(ctx, p, ca, v, di, i)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	for _, migrationAppIntent := range migrationAppIntents {
		err = h.client.DeleteAppIntent(ctx, migrationAppIntent.MetaData.Name, p, ca, v, di, i)
		if err != nil {
			apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
	log.Info(":: deleteAllMigrationAppIntentsHandler .. end ::", log.Fields{"consumers": migrationAppIntents})
}

/* Parse Http request Parameters */
func parseMigrationAppIntentReqParameters(w *http.ResponseWriter, r *http.Request) (string, string, string, string, string, string, error) {
	vars := mux.Vars(r)
	fmt.Println(vars)

	cn := vars["app-intent-name"]

	i := vars["intent-name"]
	if i == "" {
		log.Error(":: parseMigrationIntentReqParameters ..  Missing intentName in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseMigrationIntentReqParameters .. Missing name of intentName in request", http.StatusBadRequest)
		return "", "", "", "", "", "", pkgerrors.New("Missing intent-name")
	}

	p, ca, v, di, i, err := parseMigrationIntentReqParameters(w, r)
	if err != nil {
		log.Error(":: parseMigrationAppIntentReqParameters .. Failed intent validation ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseMigrationAppIntentReqParameters .. Failed intent validation", http.StatusBadRequest)
		return "", "", "", "", "", "", err
	}

	return p, ca, v, di, i, cn, nil
}
