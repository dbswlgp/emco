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

var inServerIntJSONFile string = "json-schemas/inbound-server.json"

type inboundserverintentHandler struct {
	client module.InboundServerIntentManager
}

// Check for valid format of input parameters
func validateInboundServerIntentInputs(isi module.InboundServerIntent) error {
	// validate metadata
	err := module.IsValidMetadata(isi.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid inbound server intent metadata")
	}
	return nil
}

func (h inboundserverintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var isi module.InboundServerIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, deploymentIntentGroupName, project, compositeApp, compositeAppVersion)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&isi)
	switch {
	case err == io.EOF:
		log.Error(":: Empty inbound server POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding inbound server POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	err, httpError := validation.ValidateJsonSchemaData(inServerIntJSONFile, isi)
	if err != nil {
		log.Error(":: Error validating inbound server POST data ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if isi.Metadata.Name == "" {
		log.Error(":: Missing name in inbound server POST request ::", log.Fields{})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateInboundServerIntentInputs(isi)
	if err != nil {
		log.Error(":: Invalid create inbound server body inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateServerInboundIntent(ctx, isi, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, isi, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create inbound server response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
func (h inboundserverintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var isi module.InboundServerIntent
	vars := mux.Vars(r)
	name := vars["inboundServerIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, deploymentIntentGroupName, project, compositeApp, compositeAppVersion)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&isi)

	switch {
	case err == io.EOF:
		log.Error(":: Empty inbound server PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding inbound server PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if isi.Metadata.Name == "" {
		log.Error(":: Missing name in inbound server PUT request ::", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if isi.Metadata.Name != name {
		log.Error(":: Mismatched name in inbound server PUT request ::", log.Fields{})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateInboundServerIntentInputs(isi)
	if err != nil {
		log.Error(":: Invalid inbound server PUT inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateServerInboundIntent(ctx, isi, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, isi, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding inbound server update response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h inboundserverintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["inboundServerIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]

	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetServerInboundIntents(ctx, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName)
	} else {
		ret, err = h.client.GetServerInboundIntent(ctx, name, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName)
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
		log.Error(":: Error encoding get inbound server response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h inboundserverintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["inboundServerIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]

	err := h.client.DeleteServerInboundIntent(ctx, name, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
