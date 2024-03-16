// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	pkgerrors "github.com/pkg/errors"
	netintents "gitlab.com/project-emco/core/emco-base/src/ncm/pkg/networkintents"
	nettypes "gitlab.com/project-emco/core/emco-base/src/ncm/pkg/networkintents/types"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"

	"github.com/gorilla/mux"
)

var pnetJSONFile string = "json-schemas/provider-network.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type providernetHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client netintents.ProviderNetManager
}

// Check for valid format of input parameters
func validateProviderNetInputs(p netintents.ProviderNet) error {
	// validate name
	errs := validation.IsValidName(p.Metadata.Name)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid provider network name=[%v], errors: %v", p.Metadata.Name, errs)
	}

	// validate cni type
	found := false
	for _, val := range nettypes.CNI_TYPES {
		if p.Spec.CniType == val {
			found = true
			break
		}
	}
	if !found {
		return pkgerrors.Errorf("Invalid cni type: %v", p.Spec.CniType)
	}

	// validate the provider network type
	found = false
	for _, val := range nettypes.PROVIDER_NET_TYPES {
		if strings.ToUpper(p.Spec.ProviderNetType) == val {
			found = true
			break
		}
	}
	if !found {
		return pkgerrors.Errorf("Invalid provider network type: %v", p.Spec.ProviderNetType)
	}

	// validate the subnets
	subnets := p.Spec.Ipv4Subnets
	for _, subnet := range subnets {
		err := nettypes.ValidateSubnet(subnet)
		if err != nil {
			return pkgerrors.Wrap(err, "invalid subnet")
		}
	}

	// validate the VLAN ID
	errs = validation.IsValidNumberStr(p.Spec.Vlan.VlanId, 0, 4095)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid VlAN ID %v - error: %v", p.Spec.Vlan.VlanId, errs)
	}

	// validate the VLAN Node Selector value
	expectLabels := false
	found = false
	for _, val := range nettypes.VLAN_NODE_SELECTORS {
		if strings.ToLower(p.Spec.Vlan.VlanNodeSelector) == val {
			found = true
			if val == nettypes.VLAN_NODE_SPECIFIC {
				expectLabels = true
			}
			break
		}
	}
	if !found {
		return pkgerrors.Errorf("Invalid VlAN Node Selector %v", p.Spec.Vlan.VlanNodeSelector)
	}

	// validate the node label list
	gotLabels := false
	for _, label := range p.Spec.Vlan.NodeLabelList {
		errs = validation.IsValidLabel(label)
		if len(errs) > 0 {
			return pkgerrors.Errorf("Invalid Label=%v - errors: %v", label, errs)
		}
		gotLabels = true
	}

	// Need at least one label if node selector value was "specific"
	// (if selector is "any" - don't care if labels were supplied or not
	if expectLabels && !gotLabels {
		return pkgerrors.Errorf("Node Labels required for VlAN node selector \"%v\"", nettypes.VLAN_NODE_SPECIFIC)
	}

	return nil
}

// Create handles creation of the ProviderNet entry in the database
func (h providernetHandler) createProviderNetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var p netintents.ProviderNet
	vars := mux.Vars(r)
	clusterProvider := vars["clusterProvider"]
	cluster := vars["cluster"]

	err := json.NewDecoder(r.Body).Decode(&p)

	switch {
	case err == io.EOF:
		log.Error(":: Empty provider network POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding provider network POST body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(pnetJSONFile, p)
	if err != nil {
		log.Error(":: Invalid provider network POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing provider network name in POST body ::", log.Fields{})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateProviderNetInputs(p)
	if err != nil {
		log.Error(":: Invalid provider network body inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateProviderNet(ctx, p, clusterProvider, cluster, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, p, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create provider network response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles creation/update of the ProviderNet entry in the database
func (h providernetHandler) putProviderNetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var p netintents.ProviderNet
	vars := mux.Vars(r)
	clusterProvider := vars["clusterProvider"]
	cluster := vars["cluster"]
	name := vars["providerNetwork"]

	err := json.NewDecoder(r.Body).Decode(&p)

	switch {
	case err == io.EOF:
		log.Error(":: Empty provider network PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing provider network name in PUT request ::", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if p.Metadata.Name != name {
		log.Error(":: Mismatched provider network name in PUT request ::", log.Fields{"URL name": name, "Metadata name": p.Metadata.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateProviderNetInputs(p)
	if err != nil {
		log.Error(":: Invalid provider network PUT inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateProviderNet(ctx, p, clusterProvider, cluster, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, p, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding provider network update response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular ProviderNet Name
// Returns a ProviderNet
func (h providernetHandler) getProviderNetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	clusterProvider := vars["clusterProvider"]
	cluster := vars["cluster"]
	name := vars["providerNetwork"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetProviderNets(ctx, clusterProvider, cluster)
	} else {
		ret, err = h.client.GetProviderNet(ctx, name, clusterProvider, cluster)
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
		log.Error(":: Error encoding get provider network response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular ProviderNet  Name
func (h providernetHandler) deleteProviderNetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	clusterProvider := vars["clusterProvider"]
	cluster := vars["cluster"]
	name := vars["providerNetwork"]

	err := h.client.DeleteProviderNet(ctx, name, clusterProvider, cluster)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
