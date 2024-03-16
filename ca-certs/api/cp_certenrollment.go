// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
)

// cpCertEnrollmentHandler implements the clusterProvider caCert enrollment handler functions
type cpCertEnrollmentHandler struct {
	manager clusterprovider.CaCertEnrollmentManager
}

// handleInstantiate handles the route for instantiating the caCert enrollment
func (h *cpCertEnrollmentHandler) handleInstantiate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// get the route variables
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.Instantiate(ctx, vars.cert, vars.clusterProvider); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleStatus handles the route for getting the status of the caCert enrollment
func (h *cpCertEnrollmentHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// get the route variables
	vars := _cpVars(mux.Vars(r))

	qParams, err := _statusQueryParams(r)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
	}

	stat, err := h.manager.Status(ctx, vars.cert, vars.clusterProvider,
		qParams.qInstance,
		qParams.qType,
		qParams.qOutput,
		qParams.fApps,
		qParams.fClusters,
		qParams.fResources)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	sendResponse(w, stat, http.StatusOK)
}

// handleTerminate handles the route for terminating the caCert enrollment
func (h *cpCertEnrollmentHandler) handleTerminate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// get the route variables
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.Terminate(ctx, vars.cert, vars.clusterProvider); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleUpdate handles the route for updating the caCert enrollment
func (h *cpCertEnrollmentHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// get the route variables
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.Update(ctx, vars.cert, vars.clusterProvider); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
