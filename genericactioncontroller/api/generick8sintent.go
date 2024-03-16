// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	"github.com/gorilla/mux"
)

var GenericK8sIntentSchemaJson string = "json-schemas/genericK8sIntent.json"

// genericK8sIntentHandler implements the handler functions
type genericK8sIntentHandler struct {
	client module.GenericK8sIntentManager
}

type gkiVars struct {
	compositeApp,
	deploymentIntentGroup,
	intent,
	project,
	version string
}

// handleGenericK8sIntentCreate handles the route for creating a new GenericK8sIntent
func (h genericK8sIntentHandler) handleGenericK8sIntentCreate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateIntent(w, r)
}

// handleGenericK8sIntentDelete handles the route for deleting GenericK8sIntent from the database
func (h genericK8sIntentHandler) handleGenericK8sIntentDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := _gkiVars(mux.Vars(r))
	if err := h.client.DeleteGenericK8sIntent(ctx, vars.intent, vars.project, vars.compositeApp,
		vars.version, vars.deploymentIntentGroup); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGenericK8sIntentGet handles the route for retrieving a GenericK8sIntent from the database
func (h genericK8sIntentHandler) handleGenericK8sIntentGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var (
		genericK8sIntent interface{}
		err              error
	)

	vars := _gkiVars(mux.Vars(r))
	if len(vars.intent) == 0 {
		genericK8sIntent, err = h.client.GetAllGenericK8sIntents(ctx, vars.project, vars.compositeApp,
			vars.version, vars.deploymentIntentGroup)
	} else {
		genericK8sIntent, err = h.client.GetGenericK8sIntent(ctx, vars.intent, vars.project,
			vars.compositeApp, vars.version, vars.deploymentIntentGroup)
	}

	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	sendResponse(w, genericK8sIntent, http.StatusOK)
}

// handleGenericK8sIntentUpdate handles the route for updating the existing GenericK8sIntent
func (h genericK8sIntentHandler) handleGenericK8sIntentUpdate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateIntent(w, r)
}

// createOrUpdateIntent create/update the GenericK8sIntent based on the request method
func (h genericK8sIntentHandler) createOrUpdateIntent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var genericK8sIntent module.GenericK8sIntent
	if err := validateRequestBody(r.Body, &genericK8sIntent, GenericK8sIntentSchemaJson); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	vars := _gkiVars(mux.Vars(r))

	methodPost := false
	if r.Method == http.MethodPost {
		methodPost = true
	}

	if !methodPost {
		// name in the URL should match the name in the body
		if genericK8sIntent.Metadata.Name != vars.intent {
			logutils.Error("The intent name is not matching with the name in the request",
				logutils.Fields{"GenericK8sIntent": genericK8sIntent,
					"IntentName": vars.intent})
			http.Error(w, "the intent name is not matching with the name in the request",
				http.StatusBadRequest)
			return
		}
	}

	gki, gkiExists, err := h.client.CreateGenericK8sIntent(ctx, genericK8sIntent,
		vars.project, vars.compositeApp, vars.version, vars.deploymentIntentGroup,
		methodPost)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	code := http.StatusCreated
	if gkiExists {
		// genericK8sIntent does have a current representation and that representation is successfully modified
		code = http.StatusOK
	}

	sendResponse(w, gki, code)
}

// validateGenericK8sIntentData validate the GenericK8sIntent payload for the required values
func validateGenericK8sIntentData(gki module.GenericK8sIntent) error {
	if len(gki.Metadata.Name) == 0 {
		logutils.Error("GenericK8sIntent name may not be empty",
			logutils.Fields{})
		return emcoerror.NewEmcoError(
			"genericK8sIntent name may not be empty",
			emcoerror.BadRequest,
		)
	}

	return nil
}

// _gkiVars returns the route variables for the current request
func _gkiVars(vars map[string]string) gkiVars {
	return gkiVars{
		compositeApp:          vars["compositeApp"],
		deploymentIntentGroup: vars["deploymentIntentGroup"],
		intent:                vars["genericK8sIntent"],
		project:               vars["project"],
		version:               vars["compositeAppVersion"],
	}
}
