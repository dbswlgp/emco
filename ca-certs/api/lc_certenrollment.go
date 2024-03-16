// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
)

// lcCertEnrollmentHandler implements the logicalCloud caCert enrollment handler functions
type lcCertEnrollmentHandler struct {
	manager logicalcloud.CaCertEnrollmentManager
}

// handleInstantiate handles the route for instantiating the caCert enrollment
func (h *lcCertEnrollmentHandler) handleInstantiate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// get the route variables
	vars := _lcVars(mux.Vars(r))
	if err := h.manager.Instantiate(ctx, vars.cert, vars.project); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleStatus handles the route for getting the status of the caCert enrollment
func (h *lcCertEnrollmentHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// get the route variables
	vars := _lcVars(mux.Vars(r))

	qParams, err := _statusQueryParams(r)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
	}

	stat, err := h.manager.Status(ctx, vars.cert, vars.project,
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
func (h *lcCertEnrollmentHandler) handleTerminate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// get the route variables
	vars := _lcVars(mux.Vars(r))
	if err := h.manager.Terminate(ctx, vars.cert, vars.project); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleUpdate handles the route for updating the caCert enrollment
func (h *lcCertEnrollmentHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// get the route variables
	vars := _lcVars(mux.Vars(r))
	if err := h.manager.Update(ctx, vars.cert, vars.project); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
