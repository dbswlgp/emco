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

var inClientsIntJSONFile string = "json-schemas/inbound-clients.json"

type inboundclientsintentHandler struct {
	client module.InboundClientsIntentManager
}

// Check for valid format of input parameters
func validateInboundClientsIntentInputs(ici module.InboundClientsIntent) error {
	// validate metadata
	err := module.IsValidMetadata(ici.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid inbound clients intent metadata")
	}
	return nil
}

func (h inboundclientsintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var ici module.InboundClientsIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	inboundIntentName := vars["inboundServerIntent"]
	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, deploymentIntentGroupName, project, compositeApp, compositeAppVersion)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&ici)

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

	err, httpError := validation.ValidateJsonSchemaData(inClientsIntJSONFile, ici)
	if err != nil {
		log.Error(":: Error validating inbound clients POST data ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if ici.Metadata.Name == "" {
		log.Error(":: Missing name in inbound clients POST request ::", log.Fields{})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateInboundClientsIntentInputs(ici)
	if err != nil {
		log.Error(":: Invalid create inbound clients body inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClientsInboundIntent(ctx, ici, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, ici, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create inbound clients response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
func (h inboundclientsintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var ici module.InboundClientsIntent
	vars := mux.Vars(r)
	name := vars["inboundClientsIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	inboundIntentName := vars["inboundServerIntent"]

	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, deploymentIntentGroupName, project, compositeApp, compositeAppVersion)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&ici)

	switch {
	case err == io.EOF:
		log.Error(":: Empty inbound clients PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding inbound clients PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if ici.Metadata.Name == "" {
		log.Error(":: Missing name in inbound clients PUT request ::", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if ici.Metadata.Name != name {
		log.Error(":: Mismatched name in inbound clients PUT request ::", log.Fields{})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateInboundClientsIntentInputs(ici)
	if err != nil {
		log.Error(":: Invalid inbound clients PUT inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClientsInboundIntent(ctx, ici, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, ici, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding inbound clients update response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h inboundclientsintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["inboundClientsIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	inboundIntentName := vars["inboundServerIntent"]

	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetClientsInboundIntents(ctx, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName)
	} else {
		ret, err = h.client.GetClientsInboundIntent(ctx, name, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName)
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
		log.Error(":: Error encoding get inbound clients response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h inboundclientsintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["inboundClientsIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	inboundIntentName := vars["inboundServerIntent"]

	err := h.client.DeleteClientsInboundIntent(ctx, name, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
