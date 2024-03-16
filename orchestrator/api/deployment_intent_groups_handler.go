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

var dpiJSONFile string = "json-schemas/deployment-group-intent.json"

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type deploymentIntentGroupHandler struct {
	client moduleLib.DeploymentIntentGroupManager
}

// createDeploymentIntentGroupHandler handles the create operation of DeploymentIntentGroup
func (h deploymentIntentGroupHandler) createDeploymentIntentGroupHandler(w http.ResponseWriter, r *http.Request) {

	var d moduleLib.DeploymentIntentGroup

	err := json.NewDecoder(r.Body).Decode(&d)
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
	err, httpError := validation.ValidateJsonSchemaData(dpiJSONFile, d)
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

	dIntent, _, createErr := h.client.CreateDeploymentIntentGroup(ctx, d, projectName, compositeAppName, version, true)
	if createErr != nil {
		apiErr := apierror.HandleErrors(vars, createErr, d, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(dIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h deploymentIntentGroupHandler) getDeploymentIntentGroupHandler(w http.ResponseWriter, r *http.Request) {

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

	di := vars["deploymentIntentGroup"]
	if v == "" {
		log.Error("Missing name of DeploymentIntentGroup in GET request", log.Fields{})
		http.Error(w, "Missing name of DeploymentIntentGroup in GET request", http.StatusBadRequest)
		return
	}

	dIntentGrp, err := h.client.GetDeploymentIntentGroup(ctx, di, p, ca, v)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(dIntentGrp)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h deploymentIntentGroupHandler) getAllDeploymentIntentGroupsHandler(w http.ResponseWriter, r *http.Request) {
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

	diList, err := h.client.GetAllDeploymentIntentGroups(ctx, p, ca, v)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(diList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h deploymentIntentGroupHandler) deleteDeploymentIntentGroupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	di := vars["deploymentIntentGroup"]

	err := h.client.DeleteDeploymentIntentGroup(ctx, di, p, ca, v)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// putDeploymentIntentGroupHandler handles the update operation of DeploymentIntentGroup
func (h deploymentIntentGroupHandler) putDeploymentIntentGroupHandler(w http.ResponseWriter, r *http.Request) {
	var dig moduleLib.DeploymentIntentGroup
	ctx := r.Context()
	vars := mux.Vars(r)
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]

	err := json.NewDecoder(r.Body).Decode(&dig)
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
	err, code := validation.ValidateJsonSchemaData(dpiJSONFile, dig)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), code)
		return
	}

	deploymentIntentGroup, digExists, err := h.client.CreateDeploymentIntentGroup(ctx, dig, p, ca, v, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, dig, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	statusCode := http.StatusCreated
	if digExists {
		// resource does have a current representation and that representation is successfully modified
		statusCode = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(deploymentIntentGroup)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
