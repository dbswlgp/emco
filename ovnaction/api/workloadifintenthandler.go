// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/ovnaction/pkg/module"

	"github.com/gorilla/mux"
)

var netIfJSONFile string = "json-schemas/network-load-interface.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type workloadifintentHandler struct {
	// Interface that implements workload intent operations
	// We will set this variable with a mock interface for testing
	client moduleLib.WorkloadIfIntentManager
}

// Check for valid format of input parameters
func validateWorkloadIfIntentInputs(wif moduleLib.WorkloadIfIntent) error {

	errs := validation.IsValidName(wif.Spec.IfName)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid interface name = [%v], errors: %v", wif.Spec.IfName, errs)
	}

	errs = validation.IsValidName(wif.Spec.NetworkName)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid network name = [%v], errors: %v", wif.Spec.NetworkName, errs)
	}

	// optional - only validate if supplied
	if len(wif.Spec.DefaultGateway) > 0 {
		errs = validation.IsValidName(wif.Spec.DefaultGateway)
		if len(errs) > 0 {
			return pkgerrors.Errorf("Invalid default interface = [%v], errors: %v", wif.Spec.DefaultGateway, errs)
		}
	}

	// optional - only validate if supplied
	if len(wif.Spec.IpAddr) > 0 {
		err := validation.IsIp(wif.Spec.IpAddr)
		if err != nil {
			return pkgerrors.Errorf("Invalid IP address = [%v], errors: %v", wif.Spec.IpAddr, err)
		}
	}

	// optional - only validate if supplied
	if len(wif.Spec.MacAddr) > 0 {
		err := validation.IsMac(wif.Spec.MacAddr)
		if err != nil {
			return pkgerrors.Errorf("Invalid MAC address = [%v], errors: %v", wif.Spec.MacAddr, err)
		}
	}
	return nil
}

// Create handles creation of the Network entry in the database
func (h workloadifintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var wif moduleLib.WorkloadIfIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	netControlIntent := vars["netControllerIntent"]
	workloadIntent := vars["workloadIntent"]

	err := json.NewDecoder(r.Body).Decode(&wif)

	switch {
	case err == io.EOF:
		log.Error(":: Empty workload interface POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding workload interface POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(netIfJSONFile, wif)
	if err != nil {
		log.Error(":: Invalid workload interface POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// set default value
	if len(wif.Spec.DefaultGateway) == 0 {
		wif.Spec.DefaultGateway = "false" // set default value
	}

	err = validateWorkloadIfIntentInputs(wif)
	if err != nil {
		log.Error(":: Invalid workload interface body inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateWorkloadIfIntent(ctx, wif, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, workloadIntent, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, wif, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create workload interface response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles creation/update of the Network entry in the database
func (h workloadifintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var wif moduleLib.WorkloadIfIntent
	vars := mux.Vars(r)
	name := vars["interfaceIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	netControlIntent := vars["netControllerIntent"]
	workloadIntent := vars["workloadIntent"]

	err := json.NewDecoder(r.Body).Decode(&wif)

	switch {
	case err == io.EOF:
		log.Error(":: Empty workload interface PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding workload interface PUT body ::", log.Fields{"Error": err, "Body": wif})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if wif.Metadata.Name == "" {
		log.Error(":: Missing workload interface name in PUT request ::", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if wif.Metadata.Name != name {
		log.Error(":: Mismatched workload interface name in PUT request ::", log.Fields{"URL name": name, "Metadata name": wif.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	// set default value
	if len(wif.Spec.DefaultGateway) == 0 {
		wif.Spec.DefaultGateway = "false" // set default value
	}

	err = validateWorkloadIfIntentInputs(wif)
	if err != nil {
		log.Error(":: Invalid workload interface inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateWorkloadIfIntent(ctx, wif, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, workloadIntent, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, wif, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding update workload interface response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular Network Name
// Returns a Network
func (h workloadifintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["interfaceIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	netControlIntent := vars["netControllerIntent"]
	workloadIntent := vars["workloadIntent"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetWorkloadIfIntents(ctx, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, workloadIntent)
	} else {
		ret, err = h.client.GetWorkloadIfIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, workloadIntent)
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
		log.Error(":: Error encoding get workload interface response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular Network  Name
func (h workloadifintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["interfaceIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deployIntentGroup := vars["deploymentIntentGroup"]
	netControlIntent := vars["netControllerIntent"]
	workloadIntent := vars["workloadIntent"]

	err := h.client.DeleteWorkloadIfIntent(ctx, name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, workloadIntent)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
