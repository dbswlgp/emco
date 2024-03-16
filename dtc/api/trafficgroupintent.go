// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/dtc/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	orcmod "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

var TrGroupIntJSONFile string = "json-schemas/metadata.json"

type trafficgroupintentHandler struct {
	client module.TrafficGroupIntentManager
}

// Check for valid format of input parameters
func validateTrafficGroupIntentInputs(tgi module.TrafficGroupIntent) error {
	// validate metadata
	err := module.IsValidMetadata(tgi.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid traffic group intent metadata")
	}
	return nil
}

func (h trafficgroupintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var tgi module.TrafficGroupIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]

	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, deploymentIntentGroupName, project, compositeApp, compositeAppVersion)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&tgi)
	switch {
	case err == io.EOF:
		log.Error(":: Empty traffic group POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding traffic group POST body ::", log.Fields{"Error": err})

		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	err, httpError := validation.ValidateJsonSchemaData(TrGroupIntJSONFile, tgi)
	if err != nil {
		log.Error(":: Error validating traffic group POST data ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if tgi.Metadata.Name == "" {
		log.Error(":: Missing name in traffic group POST request ::", log.Fields{})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateTrafficGroupIntentInputs(tgi)
	if err != nil {
		log.Error(":: Invalid create traffic group body inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateTrafficGroupIntent(ctx, tgi, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, tgi, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create traffic group response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
func (h trafficgroupintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var tgi module.TrafficGroupIntent
	vars := mux.Vars(r)
	name := vars["trafficGroupIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]

	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, deployIntentGroup, project, compositeApp, compositeAppVersion)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&tgi)

	switch {
	case err == io.EOF:
		log.Error(":: Empty traffic group PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding traffic group PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if tgi.Metadata.Name == "" {
		log.Error(":: Missing name in traffic group PUT request ::", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if tgi.Metadata.Name != name {
		log.Error(":: Mismatched name in traffic group PUT request ::", log.Fields{})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateTrafficGroupIntentInputs(tgi)
	if err != nil {
		log.Error(":: Invalid traffic group PUT inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateTrafficGroupIntent(ctx, tgi, project, compositeApp, compositeAppVersion, deployIntentGroup, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, tgi, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding traffic group update response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h trafficgroupintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["trafficGroupIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetTrafficGroupIntents(ctx, project, compositeApp, compositeAppVersion, deployIntentGroup)
	} else {
		ret, err = h.client.GetTrafficGroupIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding get traffic group response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h trafficgroupintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["trafficGroupIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]

	err := h.client.DeleteTrafficGroupIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
