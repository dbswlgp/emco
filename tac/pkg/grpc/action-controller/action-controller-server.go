// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package actioncontroller

import (
	"context"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdate"
	"gitlab.com/project-emco/core/emco-base/src/tac/internal/action"
)

type actionControllerServer struct {
	contextupdate.UnimplementedContextupdateServer
}

// Post Event action controller
func (ac *actionControllerServer) PostEvent(ctx context.Context, req *contextupdate.PostEventRequest) (*contextupdate.PostEventResponse, error) {
	err := action.PostEvent(ctx, req.AppContext, req.EventType)
	if err != nil {
		return &contextupdate.PostEventResponse{Success: false, PostEventMessage: err.Error()}, err
	}

	return &contextupdate.PostEventResponse{Success: true, PostEventMessage: "Post Context Updated successfully."}, nil
}

// Terminate action controller
func (ac *actionControllerServer) TerminateAppContext(ctx context.Context, req *contextupdate.TerminateRequest) (*contextupdate.TerminateResponse, error) {
	err := action.TerminateAppContext(ctx, req.AppContext)
	if err != nil {
		return &contextupdate.TerminateResponse{AppContextTerminated: false, AppContextTerminatedMessage: err.Error()}, err
	}

	return &contextupdate.TerminateResponse{AppContextTerminated: true, AppContextTerminatedMessage: "Context Terminated successfully."}, nil
}

// Update or Create action controller
func (ac *actionControllerServer) UpdateAppContext(ctx context.Context, req *contextupdate.ContextUpdateRequest) (*contextupdate.ContextUpdateResponse, error) {

	err := action.UpdateAppContext(ctx, req.IntentName, req.AppContext, req.UpdateFromAppContext)
	if err != nil {
		return &contextupdate.ContextUpdateResponse{AppContextUpdated: false, AppContextUpdateMessage: err.Error()}, err
	}

	return &contextupdate.ContextUpdateResponse{AppContextUpdated: true, AppContextUpdateMessage: "Context updated successfully."}, nil
}

// NewActionControllerServer exported
func NewActionControllerServer() *actionControllerServer {
	s := &actionControllerServer{}
	return s
}
