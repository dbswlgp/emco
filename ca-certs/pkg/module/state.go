// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"context"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
)

// StateManager exposes all the caCert state functionalities
type StateManager interface {
	Create(ctx context.Context, contextID string) error
	Get(ctx context.Context) (state.StateInfo, error)
	Update(ctx context.Context, newState state.StateValue, contextID string, createIfNotExists bool) error
}

// StateClient holds the client properties
type StateClient struct {
	dbInfo db.DbInfo
	dbKey  interface{}
}

// NewStateClient returns an instance of the StateClient which implements the Manager
func NewStateClient(dbKey interface{}) *StateClient {
	return &StateClient{
		dbInfo: db.DbInfo{
			StoreName: "resources",
			TagState:  "stateInfo"},
		dbKey: dbKey}
}

// Create the stateInfo resource in mongo
func (c *StateClient) Create(ctx context.Context, contextID string) error {
	// create the stateInfo
	a := state.ActionEntry{
		State:     state.StateEnum.Created,
		ContextId: contextID,
		TimeStamp: time.Now(),
	}

	s := state.StateInfo{}
	s.Actions = append(s.Actions, a)

	return db.DBconn.Insert(ctx, c.dbInfo.StoreName, c.dbKey, nil, c.dbInfo.TagState, s)
}

// Get the stateInfo from mongo
func (c *StateClient) Get(ctx context.Context) (state.StateInfo, error) {
	values, err := db.DBconn.Find(ctx, c.dbInfo.StoreName, c.dbKey, c.dbInfo.TagState)
	if err != nil {
		return state.StateInfo{}, err
	}

	if len(values) == 0 ||
		(len(values) > 0 &&
			values[0] == nil) {
		return state.StateInfo{}, emcoerror.NewEmcoError(
			emcoerror.StateInfoNotFound,
			emcoerror.NotFound,
		)
	}

	if len(values) > 0 &&
		values[0] != nil {
		s := state.StateInfo{}
		if err = db.DBconn.Unmarshal(values[0], &s); err != nil {
			return state.StateInfo{}, err
		}
		return s, nil
	}

	return state.StateInfo{}, emcoerror.NewEmcoError(
		emcoerror.UnknownErrorMessage,
		emcoerror.Unknown,
	)
}

// Update the stateInfo
func (c *StateClient) Update(ctx context.Context, newState state.StateValue,
	contextID string, createIfNotExists bool) error {
	s, err := c.Get(ctx)
	if err == nil { // state exists
		revision, err := state.GetLatestRevisionFromStateInfo(s)
		if err != nil {
			return err
		}

		a := state.ActionEntry{
			State:     newState,
			ContextId: contextID,
			TimeStamp: time.Now(),
			Revision:  revision + 1,
		}

		s.StatusContextId = contextID
		s.Actions = append(s.Actions, a)

		if err = db.DBconn.Insert(ctx, c.dbInfo.StoreName, c.dbKey, nil, c.dbInfo.TagState, s); err != nil {
			return err
		}

		return nil
	}

	switch e := err.(type) {
	case *emcoerror.Error:
		if e.Reason == emcoerror.NotFound &&
			createIfNotExists {
			return c.Create(ctx, contextID)
		}
	}

	return err
}

// Delete the stateInfo
func (c *StateClient) Delete(ctx context.Context) error {
	return db.DBconn.Remove(ctx, c.dbInfo.StoreName, c.dbKey)
}

// VerifyState verifies the enrollment\distribution state
func (sc *StateClient) VerifyState(ctx context.Context, event common.EmcoEvent) (string, error) {
	var contextID string
	// check for previous instantiation state
	s, err := sc.Get(ctx)
	if err != nil {
		return contextID, err
	}

	contextID = state.GetLastContextIdFromStateInfo(s)
	if contextID != "" {
		status, err := state.GetAppContextStatus(ctx, contextID)
		if err != nil {
			return contextID, err
		}

		switch status.Status {
		case appcontext.AppContextStatusEnum.Terminating:
			err := emcoerror.NewEmcoError(
				(&emcoerror.StateError{
					Resource: "CaCert",
					Event:    event,
					Status:   status.Status,
				}).Error(),
				emcoerror.Conflict,
			)
			logutils.Error("",
				logutils.Fields{
					"Error":     err.Error(),
					"ContextID": contextID})
			return contextID, err
		case appcontext.AppContextStatusEnum.Instantiating:
			err := emcoerror.NewEmcoError(
				(&emcoerror.StateError{
					Resource: "CaCert",
					Event:    event,
					Status:   status.Status}).Error(),
				emcoerror.Conflict,
			)
			logutils.Error("",
				logutils.Fields{
					"Error":     err.Error(),
					"ContextID": contextID})
			return contextID, err
		case appcontext.AppContextStatusEnum.TerminateFailed:
			err := emcoerror.NewEmcoError(
				(&emcoerror.StateError{
					Resource: "CaCert",
					Event:    event,
					Status:   status.Status}).Error(),
				emcoerror.Conflict,
			)
			logutils.Error("",
				logutils.Fields{
					"Error":     err.Error(),
					"ContextID": contextID})
			return contextID, err
		case appcontext.AppContextStatusEnum.Terminated:
			// handle events specific use cases
			switch event {
			case common.Instantiate:
				return contextID, nil
			case common.Terminate:
				err := emcoerror.NewEmcoError(
					(&emcoerror.StateError{
						Resource: "CaCert",
						Event:    event,
						Status:   status.Status}).Error(),
					emcoerror.Conflict,
				)
				logutils.Error("",
					logutils.Fields{
						"Error":     err.Error(),
						"ContextID": contextID})
				return contextID, err
			}
		case appcontext.AppContextStatusEnum.Instantiated:
			switch event {
			case common.Instantiate:
				err := emcoerror.NewEmcoError(
					(&emcoerror.StateError{
						Resource: "CaCert",
						Event:    event,
						Status:   status.Status}).Error(),
					emcoerror.Conflict,
				)
				logutils.Error("",
					logutils.Fields{
						"Error":     err.Error(),
						"ContextID": contextID})
				return contextID, err
			case common.Terminate:
				return contextID, nil
			}
		case appcontext.AppContextStatusEnum.InstantiateFailed:
			switch event {
			case common.Instantiate:
				err := emcoerror.NewEmcoError(
					(&emcoerror.StateError{
						Resource: "CaCert",
						Event:    event,
						Status:   status.Status}).Error(),
					emcoerror.Conflict,
				)
				logutils.Error("",
					logutils.Fields{
						"Error":     err.Error(),
						"ContextID": contextID})
				return contextID, err
			case common.Terminate:
				// Terminate anyway
				return contextID, nil
			}
		default:
			err := emcoerror.NewEmcoError(
				(&emcoerror.StateError{
					Resource: "CaCert",
					Event:    event,
					Status:   status.Status}).Error(),
				emcoerror.Conflict,
			)
			logutils.Error("",
				logutils.Fields{
					"Error":     err.Error(),
					"ContextID": contextID,
					"Status":    status.Status})
			return contextID, err
		}
	}

	return contextID, nil
}
