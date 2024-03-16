// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"

	"github.com/gorilla/mux"
)

var appDepJSONFile string = "json-schemas/app-dependency.json"

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type appDependencyHandler struct {
	client moduleLib.AppDependencyManager
}

// createAppDependencyHandler handles the create operation of App Dependency
func (h appDependencyHandler) createAppDependencyHandler(w http.ResponseWriter, r *http.Request) {

	var d moduleLib.AppDependency

	ctx := r.Context()

	str, code, err := validateBody(r, &d)
	if err != nil {
		http.Error(w, str, code)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(appDepJSONFile, d)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	p, ca, v, app, _, errString, code := validateParams(vars, false)
	if code != 0 {
		http.Error(w, errString, code)
		return
	}

	dIntent, createErr := h.client.CreateAppDependency(ctx, d, p, ca, v, app, false)
	if createErr != nil {
		apiErr := apierror.HandleErrors(vars, createErr, d, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(dIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h appDependencyHandler) getAppDependencyHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	vars := mux.Vars(r)

	p, ca, v, app, dp, errString, code := validateParams(vars, true)
	if code != 0 {
		http.Error(w, errString, code)
		return
	}
	dAppDep, err := h.client.GetAppDependency(ctx, dp, p, ca, v, app)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(dAppDep)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h appDependencyHandler) getAllAppDependencyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	pList := []string{"project", "compositeApp", "compositeAppVersion"}
	err := validation.IsValidParameterPresent(vars, pList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p, ca, v, app, _, errString, code := validateParams(vars, false)
	if code != 0 {
		http.Error(w, errString, code)
		return
	}

	diList, err := h.client.GetAllAppDependency(ctx, p, ca, v, app)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(diList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h appDependencyHandler) deleteappDependencyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	p, ca, v, app, dp, errString, code := validateParams(vars, true)
	if code != 0 {
		http.Error(w, errString, code)
		return
	}
	// If doesn't exist return
	_, err := h.client.GetAppDependency(ctx, dp, p, ca, v, app)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	err = h.client.DeleteAppDependency(ctx, dp, p, ca, v, app)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h appDependencyHandler) updateAppDependencyHandler(w http.ResponseWriter, r *http.Request) {

	var d moduleLib.AppDependency

	str, code, err := validateBody(r, &d)
	if err != nil {
		http.Error(w, str, code)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(appDepJSONFile, d)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	p, ca, v, app, dep, errString, code := validateParams(vars, true)
	if code != 0 {
		http.Error(w, errString, code)
		return
	}

	if dep != d.MetaData.Name {
		log.Error(":: Mismatched name in PUT request ::", log.Fields{"URL name": dep, "Metadata name": d.MetaData.Name})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)

	}
	dIntent, createErr := h.client.CreateAppDependency(ctx, d, p, ca, v, app, true)
	if createErr != nil {
		apiErr := apierror.HandleErrors(vars, createErr, d, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(dIntent)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func validateParams(vars map[string]string, depExists bool) (p, ca, v, app, dep, err string, code int) {
	p = vars["project"]
	if p == "" {
		code = http.StatusBadRequest
		err = "Missing projectName in request"
		log.Error(err, log.Fields{})
		return
	}
	ca = vars["compositeApp"]
	if ca == "" {
		code = http.StatusBadRequest
		err = "Missing compositeAppName in request"
		log.Error(err, log.Fields{})
		return
	}
	v = vars["compositeAppVersion"]
	if v == "" {
		code = http.StatusBadRequest
		err = "Missing compositeAppName version in request"
		log.Error(err, log.Fields{})
		return
	}
	app = vars["app"]
	if app == "" {
		code = http.StatusBadRequest
		err = "Missing AppName in request"
		log.Error(err, log.Fields{})
		return
	}
	if depExists {
		dep = vars["dependency"]
		if dep == "" {
			code = http.StatusBadRequest
			err = "Missing dependency in request"
			log.Error(err, log.Fields{})
			return
		}
	}
	return
}

func validateBody(r *http.Request, d *moduleLib.AppDependency) (string, int, error) {

	err := json.NewDecoder(r.Body).Decode(&d)
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		return "Empty body", http.StatusBadRequest, err
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		return "Invalid Body", http.StatusUnprocessableEntity, err

	}
	return "", 0, err
}
