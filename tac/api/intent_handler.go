// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httputil"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	"gitlab.com/project-emco/core/emco-base/src/tac/pkg/model"
	"gitlab.com/project-emco/core/emco-base/src/tac/pkg/module"
	wfMod "gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/module"

	"github.com/gorilla/mux"
)

type intentHandler struct {
	client module.WorkflowIntentManager
}

// struct used to manage the user submitted variables in the URL
type wfhVars struct {
	tacIntent,
	project,
	cApp,
	cAppVer,
	dig string
}

// _wfhVars returns the route variables for the current request
func _wfhVars(vars map[string]string) wfhVars {
	return wfhVars{
		tacIntent: vars["tac-intent"],
		project:   vars["project"],
		cApp:      vars["compositeApp"],
		cAppVer:   vars["compositeAppVersion"],
		dig:       vars["deploymentIntentGroup"],
	}
}

// Files with json schema to validate user input
var TacIntentJSONFile string = "json-schemas/tac_intent.json"
var CrJSONFile string = "json-schemas/cancel_request.json"
var SqJSONFile string = "json-schemas/temporal_status_query.json"

// handle TAC intent
func (h intentHandler) handleTacIntentCreate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdate(w, r)
}

// createOrUpdate consolidates the functionality of create workflow intent and update workflow intent into one cleaner function
func (h intentHandler) createOrUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// get all variables from url
	vars := _wfhVars(mux.Vars(r))

	// log to the user that we are in the createOrUpdate function
	logutils.Info("createOrUpdate API start", logutils.Fields{
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig,
	})

	// decode info from json body into workflow intent struct
	var wfh model.WorkflowHookIntent
	err := json.NewDecoder(r.Body).Decode(&wfh)

	// see if there was an error decoding the workflow body.
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

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(TacIntentJSONFile, wfh)
	if err != nil {
		logutils.Error(err.Error(), logutils.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// The data has been validated, and put into a developer defined struct. Send to module to submit into data base
	// if len of vars.tacIntent == 0 that means we are on the create route, if not then we are in update
	var ret model.WorkflowHookIntent
	if len(vars.tacIntent) == 0 {
		ret, err = h.client.CreateWorkflowHookIntent(ctx, wfh, vars.project, vars.cApp, vars.cAppVer, vars.dig, false)
	} else {
		ret, err = h.client.CreateWorkflowHookIntent(ctx, wfh, vars.project, vars.cApp, vars.cAppVer, vars.dig, true)
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
	err = json.NewEncoder(w).Encode(ret)
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

// handleTacIntentGet gets one or many tac intent from the DB
func (h intentHandler) handleTacIntentGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// response variables
	var resp interface{}
	var err error

	vars := _wfhVars(mux.Vars(r))

	// log request
	logutils.Info("get tac intent", logutils.Fields{"tacIntent": vars.tacIntent,
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig,
	})

	// make the request
	if len(vars.tacIntent) == 0 {
		resp, err = h.client.GetWorkflowHookIntents(ctx, vars.project, vars.cApp, vars.cAppVer, vars.dig)
	} else {
		resp, err = h.client.GetWorkflowHookIntent(ctx, vars.tacIntent, vars.project, vars.cApp, vars.cAppVer, vars.dig)
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
	logutils.Info("getHandler API success", logutils.Fields{"Error": err, "tacIntent": vars.tacIntent,
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig})
}

func (h intentHandler) handleTacIntentPut(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdate(w, r)
}

// handleTacIntentDelete - delete a TAC Intent
func (h intentHandler) handleTacIntentDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// get the variables from the URL
	vars := _wfhVars(mux.Vars(r))

	logutils.Info("Delete TAC Intent", logutils.Fields{"TacIntent": vars.tacIntent,
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig,
	})

	// delete the requested item from the backend
	err := h.client.DeleteWorkflowHookIntent(ctx, vars.tacIntent, vars.project, vars.cApp, vars.cAppVer, vars.dig)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	// write back to the user.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

	logutils.Info("Delete Tac Intent", logutils.Fields{"TacIntent": vars.tacIntent,
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig})
}

// get the status of the current TAC Intent
func (h intentHandler) handleTemporalWorkflowHookStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// grab variables from URL
	vars := _wfhVars(mux.Vars(r))

	// unmarshall data from the request, and verify it exists.
	query := wfMod.WfTemporalStatusQuery{}
	err := json.NewDecoder(r.Body).Decode(&query)
	switch {
	case err == io.EOF:
		errmsg := ":: Empty Workflow Status POST body ::"
		logutils.Error(errmsg, logutils.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		errmsg := ":: Error decoding workflow status POST body ::"
		logutils.Error(errmsg, logutils.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(SqJSONFile, query)
	if err != nil {
		logutils.Error(err.Error(), logutils.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// log that the workflow is starting
	logutils.Info("statusHandler API", logutils.Fields{"name": vars.tacIntent,
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig,
		"workflowID": query.WfID, "runID": query.RunID,
		"waitForResult":     query.WaitForResult,
		"runDescribeWfExec": query.RunDescribeWfExec,
		"getWfHistory":      query.GetWfHistory,
		"queryType":         query.QueryType, "queryParams": query.QueryParams,
	})

	// make a request to the backend with the required data.
	ret, err := h.client.GetStatusWorkflowIntent(ctx, vars.tacIntent, vars.project, vars.cApp, vars.cAppVer, vars.dig, &query)
	if err != nil {
		errmsg := "failed to get workflow status"
		logutils.Error(":: Error: "+errmsg, logutils.Fields{"name": vars.tacIntent,
			"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig,
			"workflowID": query.WfID, "runID": query.RunID,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// write response back to user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		logutils.Error(":: Error encoding workflow intent status ::",
			logutils.Fields{"name": vars.tacIntent,
				"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// log response for us.
	logutils.Info("statusHandler API success", logutils.Fields{"name": vars.tacIntent,
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig,
	})
}

// Cancel the selected workflow hook
func (h intentHandler) handleTemporalWorkflowHookCancel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var cancelReq model.WfhTemporalCancelRequest

	vars := _wfhVars(mux.Vars(r))

	if requestDump, err := httputil.DumpRequest(r, true); err != nil {
		logutils.Error("Failed to dump request", logutils.Fields{"error": err})
	} else {
		logutils.Info("cancelHandler", logutils.Fields{"reqDump": string(requestDump),
			"cancelReq": cancelReq}) // XXX
	}

	err := json.NewDecoder(r.Body).Decode(&cancelReq)
	switch {
	case err == io.EOF:
		errmsg := ":: Empty workflow cancel request POST body ::"
		logutils.Error(errmsg, logutils.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		errmsg := ":: Error decoding workflow cancel request POST body ::"
		logutils.Error(errmsg, logutils.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(CrJSONFile, cancelReq)
	if err != nil {
		logutils.Error(err.Error(), logutils.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	logutils.Info("cancelHandler", logutils.Fields{"cancelReq": cancelReq}) // XXX

	if cancelReq.Spec.TemporalServer == "" {
		errmsg := ":: Temporal Server endpoint is required."
		logutils.Error(errmsg, logutils.Fields{"name": vars.tacIntent,
			"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig,
			"cancelReq": cancelReq})
		http.Error(w, errmsg, http.StatusBadRequest)
		return
	}

	logutils.Info("cancelHandler API start", logutils.Fields{"name": vars.tacIntent,
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig,
		"cancelReq": cancelReq,
	})

	err = h.client.CancelWorkflowIntent(ctx, vars.tacIntent, vars.project, vars.cApp, vars.cAppVer, vars.dig, &cancelReq)
	if err != nil {
		errmsg := ":: Error cancelling workflow::"
		if cancelReq.Spec.Terminate {
			errmsg = ":: Error terminating workflow::"
		}
		logutils.Error(errmsg, logutils.Fields{"Error": err, "name": vars.tacIntent,
			"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	logutils.Info("cancelHandler API success", logutils.Fields{"name": vars.tacIntent,
		"project": vars.project, "cApp": vars.cApp, "cAppVer": vars.cAppVer, "dig": vars.dig,
	})
}
