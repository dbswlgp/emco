// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	"gitlab.com/project-emco/core/emco-base/src/tac/pkg/model"
	"gitlab.com/project-emco/core/emco-base/src/tac/pkg/module"
)

// manager for the interface
type workerHandler struct {
	client module.WorkerIntentManager
}

// File with json schema to validate user input
var WorkerIntentJSONFile string = "json-schemas/deploy_worker.json"

// struct used to manage the user submitted variables in the URL
type dwVars struct {
	workers,
	tacIntent,
	project,
	cApp,
	cAppVer,
	dig string
}

// _dwVars packages all the user submitted variables from the URL, and packages it into a struct
func _dwVars(vars map[string]string) dwVars {
	return dwVars{
		workers:   vars["workers"],
		tacIntent: vars["tac-intent"],
		project:   vars["project"],
		cApp:      vars["compositeApp"],
		cAppVer:   vars["compositeAppVersion"],
		dig:       vars["deploymentIntentGroup"],
	}
}

// handleWorkerCreate - Will register a worker with TAC.
func (h workerHandler) handleWorkerCreate(w http.ResponseWriter, r *http.Request) {
	h.handleWorkerCreateOrUpdate(w, r)
}

func (h workerHandler) handleWorkerUpdate(w http.ResponseWriter, r *http.Request) {
	h.handleWorkerCreateOrUpdate(w, r)
}

func (h workerHandler) handleWorkerCreateOrUpdate(w http.ResponseWriter, r *http.Request) {
	// get vars from URL
	vars := _dwVars(mux.Vars(r))

	// log to the user that we are in the createOrUpdate function
	logutils.Info("workerCreateOrUpdate", logutils.Fields{
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer,
		"dig": vars.dig, "tac-intent": vars.tacIntent,
	})

	// get the data from the request body
	var wi model.WorkerIntent
	fmt.Println(r.Body)
	err := json.NewDecoder(r.Body).Decode(&wi)

	if err != nil {
		// see if there was an error decoding the tac intent body.
		switch {
		case err == io.EOF: // this usually means there are missing fields, or just no content entirely.
			apiErr := apierror.HandleErrors(mux.Vars(r), errors.New("empty Post Body"), nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		case err != nil:
			apiErr := apierror.HandleErrors(mux.Vars(r), errors.New("error decoding json body"), nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(WorkerIntentJSONFile, wi)
	if err != nil {
		logutils.Error(err.Error(), logutils.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Send verified struct to backend to be stored in the mongodb
	var resp model.WorkerIntent
	if len(vars.workers) == 0 {
		// if there wasn't a workers in the URL then this is a create
		resp, err = h.client.CreateOrUpdateWorkerIntent(wi, vars.tacIntent, vars.project, vars.cApp, vars.cAppVer, vars.dig, false)
	} else {
		// else this is an update
		resp, err = h.client.CreateOrUpdateWorkerIntent(wi, vars.tacIntent, vars.project, vars.cApp, vars.cAppVer, vars.dig, true)
	}

	// error putting item into db, print error
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	// If we have reached this point, we have successfully created or updated a tac intent. Return success to user.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logutils.Error(":: Error encoding create workflow intent response ::", logutils.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// log success for us.
	logutils.Info("createHandler API success", logutils.Fields{
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig,
	})
}

func (h workerHandler) handleWorkerGet(w http.ResponseWriter, r *http.Request) {
	// response variabels
	var resp interface{}
	var err error

	// get vars from URL
	vars := _dwVars(mux.Vars(r))

	// log to the user that we are in the createOrUpdate function
	logutils.Info("workerGet", logutils.Fields{
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig,
	})

	// make the correct request
	// if there is no workers value in the vars then it is get many
	// if there is a value in the workers value then it is get one
	if len(vars.workers) == 0 {
		logutils.Info("Get All Workers", logutils.Fields{})
		resp, err = h.client.GetWorkerIntents(vars.project, vars.cApp, vars.cAppVer, vars.dig, vars.tacIntent)
	} else {
		logutils.Info("Get Just One Worker", logutils.Fields{})
		resp, err = h.client.GetWorkerIntent(vars.workers, vars.project, vars.cApp, vars.cAppVer, vars.dig, vars.tacIntent)
	}

	// handle error if it exists
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	// Send the response to the client.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logutils.Error(":: Error encoding tac intent(s) ::",
			logutils.Fields{"Error": err, "tacIntent": vars.tacIntent,
				"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log the success
	logutils.Info("getWorker Success", logutils.Fields{"worker": vars.workers, "tacIntent": vars.tacIntent,
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig})
}

func (h workerHandler) handleWorkerDelete(w http.ResponseWriter, r *http.Request) {
	// get vars from URL
	vars := _dwVars(mux.Vars(r))

	// log to the user that we are in the createOrUpdate function
	logutils.Info("workerDelete", logutils.Fields{
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer,
		"dig": vars.dig, "tac-intent": vars.tacIntent,
	})

	// attempt to delete the requested intent, and handle any error that may come.
	err := h.client.DeleteWorkerIntents(vars.project, vars.cApp, vars.cAppVer, vars.dig, vars.tacIntent, vars.workers)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
	}

	// write success back to the user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

	// write the success to logs
	logutils.Info("Successfully Deleted Worker Intent", logutils.Fields{
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer,
		"dig": vars.dig, "tac-intent": vars.tacIntent,
	})
}
