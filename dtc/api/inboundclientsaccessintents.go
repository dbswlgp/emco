// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

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

var inClientsAccessIntJSONFile string = "json-schemas/inbound-clients-access.json"

type inboundclientsaccessintentHandler struct {
	client module.InboundClientsAccessIntentManager
}

// Check for valid format of input parameters
func validateInboundClientsAccessIntentInputs(icai module.InboundClientsAccessIntent) error {
	// validate metadata
	err := module.IsValidMetadata(icai.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid inbound clients access intent metadata")
	}
	return nil
}

func (h inboundclientsaccessintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var icai module.InboundClientsAccessIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	inboundIntentName := vars["inboundServerIntent"]
	inboundClientIntentName := vars["inboundClientsIntent"]
	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, deploymentIntentGroupName, project, compositeApp, compositeAppVersion)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&icai)

	switch {
	case err == io.EOF:
		log.Error(":: Empty inbound clients POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding inbound clients POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(inClientsAccessIntJSONFile, icai)
	if err != nil {
		log.Error(":: Error validating inbound clients POST data ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if icai.Metadata.Name == "" {
		log.Error(":: Missing name in inbound clients access POST request ::", log.Fields{})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateInboundClientsAccessIntentInputs(icai)
	if err != nil {
		log.Error(":: Invalid create inbound clients access body inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClientsAccessInboundIntent(ctx, icai, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName, inboundClientIntentName, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, icai, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create inbound clients access response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
func (h inboundclientsaccessintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var icai module.InboundClientsAccessIntent
	vars := mux.Vars(r)
	name := vars["inboundClientsAccessIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	inboundIntentName := vars["inboundServerIntent"]
	inboundClientIntentName := vars["inboundClientsIntent"]

	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, deploymentIntentGroupName, project, compositeApp, compositeAppVersion)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&icai)

	switch {
	case err == io.EOF:
		log.Error(":: Empty inbound clients access PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding inbound clients access PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if icai.Metadata.Name == "" {
		log.Error(":: Missing name in inbound clients access PUT request ::", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if icai.Metadata.Name != name {
		log.Error(":: Mismatched name in inbound clients access PUT request ::", log.Fields{})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateInboundClientsAccessIntentInputs(icai)
	if err != nil {
		log.Error(":: Invalid inbound clients access PUT inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClientsAccessInboundIntent(ctx, icai, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName, inboundClientIntentName, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, icai, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding inbound clients access update response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h inboundclientsaccessintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["inboundClientsAccessIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	inboundIntentName := vars["inboundServerIntent"]
	inboundClientIntentName := vars["inboundClientsIntent"]

	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetClientsAccessInboundIntents(ctx, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName, inboundClientIntentName)
	} else {
		ret, err = h.client.GetClientsAccessInboundIntent(ctx, name, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName, inboundClientIntentName)
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
		log.Error(":: Error encoding get inbound clients access response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h inboundclientsaccessintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["inboundClientsAccessIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	inboundIntentName := vars["inboundServerIntent"]
	inboundClientIntentName := vars["inboundClientsIntent"]

	err := h.client.DeleteClientsAccessInboundIntent(ctx, name, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName, inboundClientIntentName)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
