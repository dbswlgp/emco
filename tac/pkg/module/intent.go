// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"context"
	"fmt"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"go.temporal.io/sdk/client"

	"github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/tac/pkg/model"
	wfMod "gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/module"
	enums "go.temporal.io/api/enums/v1"
	history "go.temporal.io/api/history/v1"
)

// WfhClientSpec is the network endpoint at which the

// WorkflowIntentClient implements the Manager
// It will also be used to maintain some localized state
type WorkflowIntentClient struct {
	db db.DbInfo
}

func NewWorkflowIntentClient() *WorkflowIntentClient {
	return &WorkflowIntentClient{
		db: db.DbInfo{
			StoreName: "resources", // should remain the same
			TagMeta:   "data",      // should remain the same
		},
	}
}

// Manager to expose the workflow intent functions
type WorkflowIntentManager interface {
	// LCM hook Routes
	CreateWorkflowHookIntent(ctx context.Context, wfh model.WorkflowHookIntent, project, cApp, cAppVer, dig string, exists bool) (model.WorkflowHookIntent, error)
	GetWorkflowHookIntent(ctx context.Context, name, project, cApp, cAppVer, dig string) (model.WorkflowHookIntent, error)
	GetWorkflowHookIntents(ctx context.Context, project, cApp, cAppVer, dig string) ([]model.WorkflowHookIntent, error)
	DeleteWorkflowHookIntent(ctx context.Context, name, project, cApp, cAppVer, dig string) error
	// monitor the status of workflows, and cancel workflows
	GetStatusWorkflowIntent(ctx context.Context, name, project, cApp, cAppVer, dig string, query *wfMod.WfTemporalStatusQuery) (*wfMod.WfTemporalStatusResponse, error)
	CancelWorkflowIntent(ctx context.Context, name, project, cApp, cAppVer, dig string, req *model.WfhTemporalCancelRequest) error
	// action controller helper functions
	GetSpecificHooks(ctx context.Context, project, cApp, cAppVer, dig, hook string) ([]model.WorkflowHookIntent, error)
}

// CreateWorkflowHookIntent - create a new hook for a workflow intent
func (v *WorkflowIntentClient) CreateWorkflowHookIntent(ctx context.Context, wfh model.WorkflowHookIntent, project, cApp, cAppVer, dig string, exists bool) (model.WorkflowHookIntent, error) {
	log.Warn("CreateWFH", log.Fields{"WfHook": wfh, "project": project,
		"cApp": cApp})

	// create key and tag for the hook
	key := model.WorkflowHookKey{
		WorkflowHook:        wfh.Metadata.Name,
		Project:             project,
		CompositeApp:        cApp,
		CompositeAppVersion: cAppVer,
		DigName:             dig,
	}

	//Check if this WorkflowHook already exists
	_, err := v.GetWorkflowHookIntent(ctx, wfh.Metadata.Name, project, cApp, cAppVer, dig)
	if err == nil && !exists {
		return model.WorkflowHookIntent{}, errors.New("workflow Hook Intent Already exists")
	}

	// if it doesn't exist already insert it into the db
	err = db.DBconn.Insert(ctx, v.db.StoreName, key, nil, v.db.TagMeta, wfh)
	if err != nil {
		return model.WorkflowHookIntent{}, err
	}

	// return no error and the hook
	return wfh, nil
}

// GetWorkflowHookIntent - get specific hook
func (v *WorkflowIntentClient) GetWorkflowHookIntent(ctx context.Context, name, project, cApp, cAppVer, dig string) (model.WorkflowHookIntent, error) {
	// create key and tag for the hook
	key := model.WorkflowHookKey{
		WorkflowHook:        name,
		Project:             project,
		CompositeApp:        cApp,
		CompositeAppVersion: cAppVer,
		DigName:             dig,
	}

	// look for the workflow hook in the db
	value, err := db.DBconn.Find(ctx, v.db.StoreName, key, v.db.TagMeta)
	if err != nil {
		// if there was an error return it to caller
		return model.WorkflowHookIntent{}, err
	} else if len(value) == 0 {
		// if it dne then return a nil
		return model.WorkflowHookIntent{}, errors.New("Workflow Hook not found")
	}

	//value is a byte array
	if value == nil {
		return model.WorkflowHookIntent{}, errors.New("Unknown Error")
	}

	wfh := model.WorkflowHookIntent{}
	if err = db.DBconn.Unmarshal(value[0], &wfh); err != nil {
		return model.WorkflowHookIntent{}, err
	}

	log.Warn("GetWFH", log.Fields{"WfHook": wfh, "project": project,
		"cApp": cApp})

	return wfh, nil
}

// GetWorkflowHookIntents - get all current registered hooks
func (v *WorkflowIntentClient) GetWorkflowHookIntents(ctx context.Context, project, cApp, cAppVer, dig string) ([]model.WorkflowHookIntent, error) {
	// create key and tag for the hook
	key := model.WorkflowHookKey{
		WorkflowHook:        "",
		Project:             project,
		CompositeApp:        cApp,
		CompositeAppVersion: cAppVer,
		DigName:             dig,
	}

	var resp []model.WorkflowHookIntent
	values, err := db.DBconn.Find(ctx, v.db.StoreName, key, v.db.TagMeta)
	if err != nil {
		return []model.WorkflowHookIntent{}, err
	}

	for _, value := range values {
		wfh := model.WorkflowHookIntent{}
		err = db.DBconn.Unmarshal(value, &wfh)
		if err != nil {
			return []model.WorkflowHookIntent{}, err
		}
		resp = append(resp, wfh)
	}

	return resp, nil
}

// DeleteWorkflowHookIntent - delete specific hook
func (v *WorkflowIntentClient) DeleteWorkflowHookIntent(ctx context.Context, name, project, cApp, cAppVer, dig string) error {

	// create key and tag for the hook
	key := model.WorkflowHookKey{
		WorkflowHook:        name,
		Project:             project,
		CompositeApp:        cApp,
		CompositeAppVersion: cAppVer,
		DigName:             dig,
	}

	err := db.DBconn.Remove(ctx, v.db.StoreName, key)
	return err
}

// GetStatusWorkflowIntent performs different types of Temporal workflow
// status queries depending on the flags specified in the status API call.
func (v *WorkflowIntentClient) GetStatusWorkflowIntent(ctx context.Context, name, project, cApp, cAppVer, dig string, query *wfMod.WfTemporalStatusQuery) (*wfMod.WfTemporalStatusResponse, error) {

	log.Info("Entered GetStatusWorkflowIntent", log.Fields{"project": project,
		"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
		"intent name": name, "query": query, "queryType": query.QueryType})

	resp := wfMod.WfTemporalStatusResponse{
		WfID:  query.WfID,
		RunID: query.RunID,
	}

	clientOptions := client.Options{HostPort: query.TemporalServer}
	c, err := client.NewClient(clientOptions)
	if err != nil {
		wrapErr := fmt.Errorf("failed to connect to Temporal server (%s). Error: %s",
			query.TemporalServer, err.Error())
		log.Error(wrapErr.Error(), log.Fields{"project": project,
			"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
			"intent name": name, "query": query})
		return &resp, wrapErr
	}

	wfCtx := context.Background() // TODO include query options later

	if query.RunDescribeWfExec {
		log.Info("Running DescribeWorkflowExecution", log.Fields{"project": project,
			"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
			"intent name": name, "query": query})
		result, err := c.DescribeWorkflowExecution(wfCtx, query.WfID, query.RunID)
		if err != nil {
			log.Error("DescribeWorkflowExecution error", log.Fields{"project": project,
				"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
				"intent name": name, "query": query, "error": err.Error()})
			return &resp, err
		}
		log.Info("DescribeWorkflowExecution success", log.Fields{"project": project,
			"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
			"intent name": name, "query": query})
		resp.WfExecDesc = *result
	}

	if query.GetWfHistory {
		msg := "Getting workflow history"
		if query.WaitForResult {
			msg += ": will now block till workflow completes"
		}
		log.Info(msg, log.Fields{"name": name, "project": project, "cApp": cApp,
			"cAppVer": cAppVer, "dig": dig, "intent name": name,
			"workflow ID": query.WfID, "workflow Run ID": query.RunID,
			"Temporal Server": query.TemporalServer,
		})

		resp.WfHistory = []history.HistoryEvent{}
		iter := c.GetWorkflowHistory(wfCtx, query.WfID, query.RunID,
			query.WaitForResult, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		for iter.HasNext() {
			event, err := iter.Next()
			if err != nil {
				log.Error("c.GetWorkflowHistory error", log.Fields{"project": project,
					"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
					"intent name": name, "query": query, "error": err.Error()})
				return &resp, err
			}
			resp.WfHistory = append(resp.WfHistory, *event)
		}
		log.Info("Got workflow history", log.Fields{"name": name, "project": project,
			"cApp": cApp, "cAppVer": cAppVer, "dig": dig, "intent name": name,
			"workflow ID": query.WfID, "workflow Run ID": query.RunID,
			"Temporal Server": query.TemporalServer,
		})
	}

	if query.QueryType != "" {
		log.Info("Querying workflow", log.Fields{"name": name, "project": project,
			"cApp": cApp, "cAppVer": cAppVer, "dig": dig, "intent name": name,
			"workflow ID": query.WfID, "workflow Run ID": query.RunID,
			"queryType": query.QueryType, "queryARgs": query.QueryParams,
			"Temporal Server": query.TemporalServer,
		})
		queryWithOptions := &client.QueryWorkflowWithOptionsRequest{
			WorkflowID:           query.WfID,
			RunID:                query.RunID,
			QueryType:            query.QueryType,
			Args:                 query.QueryParams,
			QueryRejectCondition: enums.QUERY_REJECT_CONDITION_NONE,
		}
		response, err := c.QueryWorkflowWithOptions(wfCtx, queryWithOptions)
		if err != nil {
			wrapErr := fmt.Errorf("query failed (%s). Error: %s",
				query.QueryType, err.Error())
			log.Error(wrapErr.Error(), log.Fields{"project": project,
				"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
				"intent name": name, "query": query})
			return &resp, wrapErr
		}
		// type of response: *QueryWorkflowWithOptionsResponse
		// TODO handle response.QueryRejected (reason why query was rejected, if any
		if !response.QueryResult.HasValue() {
			wrapErr := fmt.Errorf("got no result for query (%s)", query.QueryType)
			log.Error(wrapErr.Error(), log.Fields{"project": project,
				"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
				"intent name": name, "query": query})
			return &resp, wrapErr
		}

		err = response.QueryResult.Get(&resp.WfQueryResult)
		if err != nil {
			wrapErr := fmt.Errorf("failed to get result for query (%s). Error: %s",
				query.QueryType, err.Error())
			log.Error(wrapErr.Error(), log.Fields{"project": project,
				"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
				"intent name": name, "query": query})
			return &resp, wrapErr
		}
		log.Info("Query got result", log.Fields{"project": project,
			"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
			"intent name": name, "query": query, "result": resp.WfQueryResult})
	}

	if query.WaitForResult {
		// TODO Status query can take options for timeouts and retries.
		log.Info("GetStatusWorkflowIntent will now block till workflow completes",
			log.Fields{"name": name, "project": project, "cApp": cApp,
				"cAppVer": cAppVer, "dig": dig, "intent name": name,
				"workflow ID": query.WfID, "workflow Run ID": query.RunID,
				"Temporal Server": query.TemporalServer,
			})
		workflowRun := c.GetWorkflow(wfCtx, query.WfID, query.RunID)

		var result interface{}
		_ = workflowRun.Get(wfCtx, &result)
		log.Info("Workflow got result", log.Fields{"project": project,
			"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
			"intent name": name, "query": query, "result": result})
	}

	return &resp, nil
}

// Cancel/terminate the  Workflow
func (v *WorkflowIntentClient) CancelWorkflowIntent(ctx context.Context, name, project, cApp, cAppVer, dig string, req *model.WfhTemporalCancelRequest) error {
	var err error

	//Construct key and tag to select the entry
	key := model.WorkflowHookKey{
		WorkflowHook:        name,
		Project:             project,
		CompositeApp:        cApp,
		CompositeAppVersion: cAppVer,
		DigName:             dig,
	}

	value, err := db.DBconn.Find(ctx, v.db.StoreName, key, v.db.TagMeta)
	if err != nil {
		log.Error("CancelWorkflowHookIntent: Error getting intent",
			log.Fields{"project": project, "composite app": cApp,
				"composite app version": cAppVer, "DIG": dig,
				"intent name": name, "error": err,
			})
		return err
	} else if len(value) == 0 {
		log.Error("CancelWorkflowHookIntent: Intent not found",
			log.Fields{"project": project, "composite app": cApp,
				"composite app version": cAppVer, "DIG": dig,
				"intent name": name, "error": err,
			})
		return errors.New("Workflow Intent Hook not found")
	}

	//value is a byte array
	if value == nil {
		log.Error("CancelWorkflowHookIntent: Intent value invalid",
			log.Fields{"project": project, "composite app": cApp,
				"composite app version": cAppVer, "DIG": dig,
				"intent name": name, "error": err,
			})
		return errors.New("Unknown Error")
	}

	wfh := model.WorkflowHookIntent{}
	if err = db.DBconn.Unmarshal(value[0], &wfh); err != nil {
		log.Error("CancelWorkflowIntent: Can't decode intent",
			log.Fields{"project": project, "composite app": cApp,
				"composite app version": cAppVer, "DIG": dig,
				"intent name": name, "error": err,
			})
		return err
	}

	spec := req.Spec

	wfID := wfh.Spec.WfTemporalSpec.WfStartOpts.ID
	if spec.WfID != "" { // wfID in the request overrides the one in the intent
		wfID = spec.WfID
	}

	clientOptions := client.Options{HostPort: spec.TemporalServer}
	c, err := client.NewClient(clientOptions)
	if err != nil {
		wrapErr := fmt.Errorf("CancelWorkflowIntent: Failed to "+
			"connect to Temporal server (%s). Error: %s",
			spec.TemporalServer, err.Error())
		log.Error(wrapErr.Error(), log.Fields{"project": project,
			"composite app": cApp, "composite app version": cAppVer, "DIG": dig,
			"intent name": name, "cancel request": req})
		return wrapErr
	}

	wfCtx := context.Background() // TODO include options later

	if spec.Terminate {
		log.Info("CancelWorkflowIntent: Calling TerminateWorkflow", log.Fields{
			"wfID": wfID, "spec.runID": spec.RunID, "spec.reason": spec.Reason,
			"spec.Details": spec.Details})
		err = c.TerminateWorkflow(wfCtx, wfID, spec.RunID, spec.Reason, spec.Details)
	} else {
		log.Info("CancelWorkflowIntent: Calling CancelWorkflow", log.Fields{
			"wfID": wfID, "spec.runID": spec.RunID})
		err = c.CancelWorkflow(wfCtx, wfID, spec.RunID)
	}

	// Caller logs the error
	return err
}

// get all of the hooks of a specific kind (ie pre-install, post-install, pre-delete, etc.)
func (v *WorkflowIntentClient) GetSpecificHooks(ctx context.Context, project, cApp, cAppVer, dig, hook string) ([]model.WorkflowHookIntent, error) {

	// Get all workflow hook intents, and grab only the pre-install hooks
	hooks, err := v.GetWorkflowHookIntents(ctx, project, cApp, cAppVer, dig)
	if err != nil {
		return []model.WorkflowHookIntent{}, err
	}

	var pre []model.WorkflowHookIntent

	// iterate through hooks, and all pre-install hooks to the list
	for _, h := range hooks {
		if h.Spec.HookType == hook {
			pre = append(pre, h)
		}
	}

	// return all pre install hooks for this temporal action intent
	return pre, nil
}
