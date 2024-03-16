// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/module"
)

type workflowIntentHandler struct {
	client moduleLib.WorkflowIntentManager
}

// location of the files
var wfiJSONFile string = "json-schemas/workflow_intent.json"
var crJSONFile string = "json-schemas/cancel_request.json"

func (h workflowIntentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var wfi moduleLib.WorkflowIntent
	vars := mux.Vars(r)
	project := vars["project"]
	cApp := vars["compositeApp"]
	cAppVer := vars["compositeAppVersion"]
	dig := vars["deploymentIntentGroup"]

	err := json.NewDecoder(r.Body).Decode(&wfi)

	switch {
	case err == io.EOF:
		errmsg := ":: Empty workflow intent POST body ::"
		log.Error(errmsg, log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		errmsg := ":: Error decoding workflow intent POST body ::"
		log.Error(errmsg, log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	log.Info("createHandler API start", log.Fields{
		"project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig,
	})

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(wfiJSONFile, wfi)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateWorkflowIntent(ctx, wfi, project, cApp, cAppVer, dig, false)
	if err != nil {
		//apiErr := apierror.HandleErrors(vars, err, wfi, apiErrors)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create workflow intent response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info("createHandler API success", log.Fields{
		"project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig,
	})
}

func (h workflowIntentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var ret interface{}
	var err error

	vars := mux.Vars(r)
	name := vars["workflow-intent-name"]
	project := vars["project"]
	cApp := vars["compositeApp"]
	cAppVer := vars["compositeAppVersion"]
	dig := vars["deploymentIntentGroup"]

	log.Info("getHandler API start", log.Fields{"name": name,
		"project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig,
	})

	if len(name) == 0 {
		ret, err = h.client.GetWorkflowIntents(ctx, project, cApp, cAppVer, dig)
	} else {
		ret, err = h.client.GetWorkflowIntent(ctx, name, project, cApp, cAppVer, dig)
	}

	if err != nil {
		log.Error(":: Error getting workflow intent(s) ::",
			log.Fields{"Error": err, "name": name, "project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding workflow intent(s) ::",
			log.Fields{"Error": err, "name": name, "project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info("getHandler API success", log.Fields{"name": name,
		"project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig,
	})
}

func (h workflowIntentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["workflow-intent-name"]
	project := vars["project"]
	cApp := vars["compositeApp"]
	cAppVer := vars["compositeAppVersion"]
	dig := vars["deploymentIntentGroup"]

	log.Info("deleteHandler API start", log.Fields{"name": name,
		"project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig,
	})

	err := h.client.DeleteWorkflowIntent(ctx, name, project, cApp, cAppVer, dig)
	if err != nil {
		log.Error(":: Error deleting workflow intent::",
			log.Fields{"Error": err, "name": name, "project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	log.Info("deleteHandler API success", log.Fields{"name": name,
		"project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig,
	})
}

func (h workflowIntentHandler) startHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["workflow-intent-name"]
	project := vars["project"]
	cApp := vars["compositeApp"]
	cAppVer := vars["compositeAppVersion"]
	dig := vars["deploymentIntentGroup"]

	log.Info("startHandler API start", log.Fields{"name": name,
		"project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig})
	err := h.client.StartWorkflowIntent(ctx, name, project, cApp, cAppVer, dig)
	if err != nil {
		log.Error(":: Error starting workflow intent ::",
			log.Fields{"Error": err, "name": name, "project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Info("startHandler API return", log.Fields{"name": name,
		"project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig})
}

func (h workflowIntentHandler) statusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["workflow-intent-name"]
	project := vars["project"]
	cApp := vars["compositeApp"]
	cAppVer := vars["compositeAppVersion"]
	dig := vars["deploymentIntentGroup"]

	query, err := buildStatusQuery(r)
	if err != nil {
		errmsg := err.Error()
		log.Error(":: Error: "+errmsg, log.Fields{"name": name,
			"project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig,
			"workflow intent name": name})
		http.Error(w, errmsg, http.StatusBadRequest)
		return
	}

	log.Info("statusHandler API", log.Fields{"name": name,
		"project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig,
		"intent name": name, "workflowID": query.WfID, "runID": query.RunID,
		"waitForResult":     query.WaitForResult,
		"runDescribeWfExec": query.RunDescribeWfExec,
		"getWfHistory":      query.GetWfHistory,
		"queryType":         query.QueryType, "queryParams": query.QueryParams,
	})

	ret, err := h.client.GetStatusWorkflowIntent(ctx, name,
		project, cApp, cAppVer, dig, query)
	if err != nil {
		errmsg := "failed to get workflow status"
		log.Error(":: Error: "+errmsg, log.Fields{"name": name,
			"project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig,
			"intent name": name, "workflowID": query.WfID, "runID": query.RunID,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding workflow intent status ::",
			log.Fields{"Error": err, "name": name, "project": project, "cApp": cApp,
				"cAppVer": cAppVer, "dig": dig, "intent name": name})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info("statusHandler API success", log.Fields{"name": name, "project": project,
		"cApp": cApp, "cAppVer": cAppVer, "dig": dig, "intent name": name,
	})
}

func buildStatusQuery(r *http.Request) (*moduleLib.WfTemporalStatusQuery, error) {
	query := &moduleLib.WfTemporalStatusQuery{}

	err := r.ParseForm()
	if err != nil {
		return query, err
	}

	errmsg := ""
	for key := range r.Form {
		value := r.FormValue(key)
		switch key {
		case "temporalServer":
			// Cut leading/trailing quotes, if any
			if len(value) > 0 && (value[0] == '"' || value[0] == '\'') {
				value = value[1:]
			}
			if len(value) > 0 && (value[len(value)-1] == '"' || value[len(value)-1] == '\'') {
				value = value[:len(value)-1]
			}
			// TODO Use github.com/asaskevich/govalidator to validate this
			query.TemporalServer = value
		case "workflowID":
			query.WfID = value
		case "runID":
			query.RunID = value
		case "queryType":
			query.QueryType = value
		case "waitForResult":
			switch value {
			case "true":
				query.WaitForResult = true
			case "false", "":
				query.WaitForResult = false
			default:
				fmt.Sprintf(errmsg, "%s must be a boolean but is %s", key, value)
			}
		case "runDescribeWfExec":
			switch value {
			case "true":
				query.RunDescribeWfExec = true
			case "false", "":
				query.RunDescribeWfExec = false
			default:
				fmt.Sprintf(errmsg, "%s must be a boolean but is %s", key, value)
			}
		case "getWfHistory":
			switch value {
			case "true":
				query.GetWfHistory = true
			case "false", "":
				query.GetWfHistory = false
			default:
				fmt.Sprintf(errmsg, "%s must be a boolean but is %s", key, value)
			}
		// TODO Add queryParams in the future
		default:
			errmsg = "Unknown query parameter: " + key
		}
		if errmsg != "" {
			return query, fmt.Errorf(errmsg)
		}
	}

	return query, nil
}

func (h workflowIntentHandler) cancelHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var cancelReq moduleLib.WfTemporalCancelRequest

	vars := mux.Vars(r)
	name := vars["workflow-intent-name"]
	project := vars["project"]
	cApp := vars["compositeApp"]
	cAppVer := vars["compositeAppVersion"]
	dig := vars["deploymentIntentGroup"]

	if requestDump, err := httputil.DumpRequest(r, true); err != nil {
		log.Error("Failed to dump request", log.Fields{"error": err})
	} else {
		log.Info("cancelHandler", log.Fields{"reqDump": string(requestDump),
			"cancelReq": cancelReq}) // XXX
	}

	err := json.NewDecoder(r.Body).Decode(&cancelReq)
	switch {
	case err == io.EOF:
		errmsg := ":: Empty workflow cancel request POST body ::"
		log.Error(errmsg, log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		errmsg := ":: Error decoding workflow cancel request POST body ::"
		log.Error(errmsg, log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(crJSONFile, cancelReq)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	log.Info("cancelHandler", log.Fields{"cancelReq": cancelReq}) // XXX

	if cancelReq.Spec.TemporalServer == "" {
		errmsg := ":: Temporal Server endpoint is required."
		log.Error(errmsg, log.Fields{"name": name,
			"project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig,
			"cancelReq": cancelReq})
		http.Error(w, errmsg, http.StatusBadRequest)
		return
	}

	log.Info("cancelHandler API start", log.Fields{"name": name,
		"project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig,
		"cancelReq": cancelReq,
	})

	err = h.client.CancelWorkflowIntent(ctx, name, project, cApp, cAppVer, dig, &cancelReq)
	if err != nil {
		errmsg := ":: Error cancelling workflow::"
		if cancelReq.Spec.Terminate {
			errmsg = ":: Error terminating workflow::"
		}
		log.Error(errmsg, log.Fields{"Error": err, "name": name, "project": project,
			"cApp": cApp, "cAppVer": cAppVer, "dig": dig})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	log.Info("cancelHandler API success", log.Fields{"name": name,
		"project": project, "cApp": cApp, "cAppVer": cAppVer, "dig": dig,
	})
}
