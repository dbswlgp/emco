// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package intent

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"net/http"
)

func (c Client) CreatePolicyIntentHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	intentData := new(Intent)
	if err := json.NewDecoder(r.Body).Decode(intentData); err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Copying key information to Policy Spec
	// This is a redundant information in database
	// orchestrator/pkg/infra/db library's find() method limits the fields that it returns
	// It takes 'tag' as parameter and set projection to include only 'tag'
	// But we need whole record when building reverse map during the boot. Hence, this workaround.
	intentData.Spec.Project = v["project"]
	intentData.Spec.CompositeApp = v["compositeApp"]
	intentData.Spec.CompositeAppVersion = v["compositeAppVersion"]
	intentData.Spec.DeploymentIntentGroup = v["deploymentIntentGroup"]
	intentData.Spec.PolicyIntentID = v["policyIntentId"]
	request := &Request{
		Project:               v["project"],
		CompositeApp:          v["compositeApp"],
		CompositeAppVersion:   v["compositeAppVersion"],
		DeploymentIntentGroup: v["deploymentIntentGroup"],
		PolicyIntentId:        v["policyIntentId"],
		IntentData:            intentData,
	}
	intent, err := c.CreateIntent(ctx, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(intent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("Create Policy intent processed successfully", log.Fields{"IntentID": request.PolicyIntentId})
}

func (c Client) GetPolicyIntentHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	request := &Request{
		Project:               v["project"],
		CompositeApp:          v["compositeApp"],
		CompositeAppVersion:   v["compositeAppVersion"],
		DeploymentIntentGroup: v["deploymentIntentGroup"],
		PolicyIntentId:        v["policyIntentId"],
	}
	response, err := c.GetIntent(ctx, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if response == nil {
		http.Error(w, "404 Policy Intent not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("Get Policy intent processed successfully", log.Fields{"IntentID": request.PolicyIntentId})
}

func (c Client) DeletePolicyIntentHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	request := &Request{
		Project:               v["project"],
		CompositeApp:          v["compositeApp"],
		CompositeAppVersion:   v["compositeAppVersion"],
		DeploymentIntentGroup: v["deploymentIntentGroup"],
		PolicyIntentId:        v["policyIntentId"],
	}
	if err := c.DeleteIntent(ctx, request); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	log.Info("Delete Policy intent processed successfully", log.Fields{"IntentID": request.PolicyIntentId})
}
