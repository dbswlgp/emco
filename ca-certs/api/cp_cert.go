// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

// cpCertHandler implements the clusterProvider caCert handler functions
type cpCertHandler struct {
	manager clusterprovider.CaCertManager
}

// handleCertificateCreate handles the route for creating a new caCert
func (h *cpCertHandler) handleCertificateCreate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCertificate(w, r)
}

// handleCertificateDelete handles the route for deleting a caCert
func (h *cpCertHandler) handleCertificateDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// get the route variables
	vars := _cpVars(mux.Vars(r))

	if err := h.manager.DeleteCert(ctx, vars.cert, vars.clusterProvider); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleCertificateGet handles the route for retrieving a caCert
func (h *cpCertHandler) handleCertificateGet(w http.ResponseWriter, r *http.Request) {
	var (
		certs interface{}
		err   error
	)

	ctx := r.Context()
	// get the route variables
	vars := _cpVars(mux.Vars(r))
	if len(vars.cert) == 0 {
		certs, err = h.manager.GetAllCert(ctx, vars.clusterProvider)
	} else {
		certs, err = h.manager.GetCert(ctx, vars.cert, vars.clusterProvider)
	}

	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	sendResponse(w, certs, http.StatusOK)
}

// handleCertificateUpdate handles the route for updating a caCert
func (h *cpCertHandler) handleCertificateUpdate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCertificate(w, r)
}

// createOrUpdateCertificate create/update the caCert  based on the request method
func (h *cpCertHandler) createOrUpdateCertificate(w http.ResponseWriter, r *http.Request) {
	var cert module.CaCert

	// validate the request body before storing it in the database
	if code, err := validateRequestBody(r.Body, &cert, CertificateSchemaJson); err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	ctx := r.Context()
	// get the route variables
	vars := _cpVars(mux.Vars(r))

	methodPost := false
	if r.Method == http.MethodPost {
		methodPost = true
	}

	if !methodPost {
		// name in the URL should match the name in the body
		if cert.MetaData.Name != vars.cert {
			err := "caCert name is not matching with the name in the request"
			logutils.Error(err,
				logutils.Fields{
					"Cert":     cert,
					"CertName": vars.cert})
			http.Error(w, err, http.StatusBadRequest)
			return
		}
	}

	crt, certExists, err := h.manager.CreateCert(ctx, cert, vars.clusterProvider, methodPost)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	code := http.StatusCreated
	if certExists {
		// caCert does have a current representation and that representation is successfully modified
		code = http.StatusOK
	}

	sendResponse(w, crt, code)
}
