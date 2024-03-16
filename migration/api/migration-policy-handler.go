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
addMigrationPolicyHandler handles the URL
URL: /v2/projects/{project-name}/migration-policies
*/
// Add Migration Policy in project
func (h MigrationPlacementIntentHandler) addMigrationPolicyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var migration migrationModel.MigrationPolicy
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: addMigrationPolicyHandler .. start ::", log.Fields{"req": string(reqDump)})

	err := json.NewDecoder(r.Body).Decode(&migration)
	switch {
	case err == io.EOF:
		log.Error(":: addMigrationPolicyHandler .. Empty POST body ::", log.Fields{"Error": err})
		http.Error(w, "addMigrationPolicyHandler .. Empty POST body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: addMigrationPolicyHandler .. Error decoding POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Validate json schema
	err, httpError := validation.ValidateJsonSchemaData(migrationPolicyJSONFile, migration)
	if err != nil {
		log.Error(":: addMigrationPolicyHandler .. JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	p := vars["project-name"]
	//ca := vars["composite-app-name"]
	//v := vars["composite-app-version"]
	//d := vars["deployment-intent-group-name"]

	log.Info(":: addMigrationPolicyHandler .. Req ::", log.Fields{"project": p, "policy-name": migration.MetaData.Name})
	policy, err := h.client.AddPolicy(ctx, migration, p, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, migration, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(policy)
	if err != nil {
		log.Error(":: addMigrationPolicyHandler ..  Encoder error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: addMigrationPolicyHandler .. end ::", log.Fields{"policy": policy})
}

/*
getIntentByNameHandler handles the URL
URL: /v2/projects/{project-name}/migration-policies
*/
// Query Migration Policy in project
func (h MigrationPlacementIntentHandler) getMigrationPolicyByNameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: getMigrationPolicyByNameHandler .. start ::", log.Fields{"req": string(reqDump)})

	p, _, err := parseMigrationPolicyReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pN := r.URL.Query().Get("policy")
	if pN == "" {
		log.Error(":: getMigrationPolicyByNameHandler .. Missing policy-name in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(w, "getMigrationPolicyByNameHandler .. Missing policy-name in request", http.StatusBadRequest)
		return
	}

	policy, err := h.client.GetPolicyByName(ctx, pN, p)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(policy)
	if err != nil {
		log.Error(":: getMigrationPolicyByNameHandler .. Encoder error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: getMigrationPolicyByNameHandler .. end ::", log.Fields{"policy": policy})
}

/*
getMigrationPolicyHandler/getMigrationPolicyHandlers handles the URL
URL: /v2/projects/{project-name}/migration-policies/{policy-name}
*/
// Get Migration Policy in project
func (h MigrationPlacementIntentHandler) getMigrationPolicyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: getMigrationPolicyHandler .. start ::", log.Fields{"req": string(reqDump)})
	p, pn, err := parseMigrationPolicyReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: getMigrationPolicyHandler .. Req ::", log.Fields{"project": p, "policy-name": pn})

	var policies interface{}
	if len(pn) == 0 {
		policies, err = h.client.GetAllPolicies(ctx, p)
	} else {
		policies, _, err = h.client.GetPolicy(ctx, pn, p)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(policies)
	if err != nil {
		log.Error(":: getMigrationPolicyHandler .. Encoder failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: getMigrationPolicyHandler .. end ::", log.Fields{"policies": policies})
}

/*
putMigrationPlacementIntentHandler handles the URL
URL: /v2/projects/{project-name}/migration-policies/{policy-name}
*/
// Update Migration Policy in project
func (h MigrationPlacementIntentHandler) putMigrationPolicyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var migration migrationModel.MigrationPolicy
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: putMigrationPolicyHandler .. start ::", log.Fields{"req": string(reqDump)})

	err := json.NewDecoder(r.Body).Decode(&migration)
	switch {
	case err == io.EOF:
		log.Error(":: putMigrationPolicyHandler .. Empty PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: putMigrationPolicyHandler .. decoding resource PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Validate json schema
	err, httpError := validation.ValidateJsonSchemaData(migrationPolicyJSONFile, migration)
	if err != nil {
		log.Error(":: putMigrationPolicyHandler .. JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Parse request parameters
	p, pn, err := parseMigrationPolicyReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if migration.MetaData.Name != pn {
		log.Error(":: putMigrationPolicyHandler .. Mismatched name in PUT request ::", log.Fields{"bodyname": migration.MetaData.Name, "policy": pn})
		http.Error(w, "putMigrationPolicyHandler .. Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	log.Info(":: putMigrationPolicyHandler .. Req ::", log.Fields{"project": p, "policy-name": pn})
	policy, err := h.client.AddPolicy(ctx, migration, p, true)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, migration, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(policy)
	if err != nil {
		log.Error(":: putMigrationPolicyHandler .. encoding failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: putMigrationPolicyHandler .. end ::", log.Fields{"policy": policy})
}

/*
deleteMigrationPolicyHandler handles the URL
URL: /v2/projects/{project-name}/migration-policies/{policy-name}
*/
// Delete Migration Policy in project
func (h MigrationPlacementIntentHandler) deleteMigrationPolicyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: deleteMigrationPolicyHandler .. start ::", log.Fields{"req": string(reqDump)})

	// Parse request parameters
	p, pn, err := parseMigrationPolicyReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: deleteMigrationPolicyHandler .. Req ::", log.Fields{"project": p, "policy-name": pn})

	_, _, err = h.client.GetPolicy(ctx, pn, p)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	err = h.client.DeletePolicy(ctx, pn, p)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Info(":: deleteMigrationPolicyHandler .. end ::", log.Fields{"req": string(reqDump)})
}

/*
deleteAllMigrationPolicysHandler handles the URL
URL: /v2/projects/{project-name}/migration-policies
*/
// Delete all Migration Policies in project
func (h MigrationPlacementIntentHandler) deleteAllMigrationPoliciesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: deleteAllMigrationPoliciesHandler .. start ::", log.Fields{"req": string(reqDump)})

	// Parse request parameters
	p, pn, err := parseMigrationPolicyReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: deleteAllMigrationPoliciesHandler .. Req ::", log.Fields{"project": p, "policy-name": pn})

	migrationpolicies, err := h.client.GetAllPolicies(ctx, p)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	for _, migrationPolicy := range migrationpolicies {
		err = h.client.DeletePolicy(ctx, migrationPolicy.MetaData.Name, p)
		if err != nil {
			apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
	log.Info(":: deleteAllMigrationPoliciesHandler .. end ::", log.Fields{"req": string(reqDump)})
}

/* Parse Http request Parameters */
func parseMigrationPolicyReqParameters(w *http.ResponseWriter, r *http.Request) (string, string, error) {
	vars := mux.Vars(r)

	pn := vars["policy-name"]

	p := vars["project-name"]
	if p == "" {
		log.Error(":: parseMigrationPolicyReqParameters ..  Missing projectName in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseMigrationPolicyReqParameters .. Missing projectName in request", http.StatusBadRequest)
		return "", "", pkgerrors.New("Missing project-name")
	}

	return p, pn, nil
}
