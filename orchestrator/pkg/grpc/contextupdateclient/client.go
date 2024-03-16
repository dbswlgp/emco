// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package contextupdateclient

import (
	"context"
	"time"

	pkgerrors "github.com/pkg/errors"
	contextpb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdate"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/rpc"
)

// InvokeContextUpdate will make the grpc call to the specified controller
// The controller will take the specified intentName and update the AppContext
// appropriatly based on its operation as a placement or action controller.
func InvokeContextUpdate(ctx context.Context, controllerName, intentName, appContextId, updateFromAppContextId string) error {
	var err error
	var rpcClient contextpb.ContextupdateClient
	var updateRes *contextpb.ContextUpdateResponse

	timeout := time.Duration(config.GetConfiguration().GrpcCallTimeout)
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Millisecond)
	defer cancel()

	conn := rpc.GetRpcConn(ctx, controllerName)
	if conn != nil {
		rpcClient = contextpb.NewContextupdateClient(conn)
		updateReq := new(contextpb.ContextUpdateRequest)
		updateReq.AppContext = appContextId
		updateReq.UpdateFromAppContext = updateFromAppContextId
		updateReq.IntentName = intentName
		updateRes, err = rpcClient.UpdateAppContext(ctx, updateReq)
	} else {
		return pkgerrors.Errorf("ContextUpdate Failed - Could not get ContextupdateClient: %v", controllerName)
	}

	if err == nil {
		if updateRes.AppContextUpdated {
			log.Info("ContextUpdate Passed", log.Fields{
				"Controller": controllerName,
				"Intent":     intentName,
				"AppContext": appContextId,
				"Message":    updateRes.AppContextUpdateMessage,
			})
			return nil
		} else {
			return pkgerrors.Errorf("ContextUpdate Failed: %v", updateRes.AppContextUpdateMessage)
		}
	}
	return err
}

// InvokeContextTerminate will make the grpc call to the specified controller
// The controller will take the specified AppContext and take steps to
// based on action controller to terminate the AppContext
func InvokeContextTerminate(ctx context.Context, controllerName, appContextId string) error {
	var err error
	var rpcClient contextpb.ContextupdateClient
	var terminateRes *contextpb.TerminateResponse

	timeout := time.Duration(config.GetConfiguration().GrpcCallTimeout)

	ctx, cancel := context.WithTimeout(ctx, timeout*time.Millisecond)
	defer cancel()

	conn := rpc.GetRpcConn(ctx, controllerName)
	if conn != nil {
		rpcClient = contextpb.NewContextupdateClient(conn)
		req := new(contextpb.TerminateRequest)
		req.AppContext = appContextId
		terminateRes, err = rpcClient.TerminateAppContext(ctx, req)
	} else {
		return pkgerrors.Errorf("ContextUpdate Failed - Could not get ContextupdateClient: %v", controllerName)
	}

	if err == nil {
		if terminateRes.AppContextTerminated {
			log.Info("Terminate Context Passed", log.Fields{
				"Controller": controllerName,
				"AppContext": appContextId,
				"Message":    terminateRes.AppContextTerminatedMessage,
			})
			return nil
		} else {
			return pkgerrors.Errorf("ContextUpdate Failed: %v", terminateRes.AppContextTerminatedMessage)
		}
	}
	return err
}

// InvokePostEvent will make the grpc call to the specified controller
// The controller will take the specified AppContext and take steps to
// based on action controller to do Post event processing
func InvokePostEvent(ctx context.Context, controllerName, appContextId string, eventType string) error {
	var err error
	var rpcClient contextpb.ContextupdateClient
	var postEventRes *contextpb.PostEventResponse

	timeout := time.Duration(config.GetConfiguration().GrpcCallTimeout)
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Millisecond)
	defer cancel()

	conn := rpc.GetRpcConn(ctx, controllerName)
	if conn != nil {
		rpcClient = contextpb.NewContextupdateClient(conn)
		req := new(contextpb.PostEventRequest)
		req.AppContext = appContextId
		req.EventType = contextpb.EventType(contextpb.EventType_value[eventType])
		postEventRes, err = rpcClient.PostEvent(ctx, req)
	} else {
		return pkgerrors.Errorf("ContextUpdate Failed - Could not get ContextupdateClient: %v", controllerName)
	}

	if err == nil {
		if postEventRes.Success {
			log.Info("Post Event Success", log.Fields{
				"Controller": controllerName,
				"AppContext": appContextId,
				"Message":    postEventRes.PostEventMessage,
				"Event":      eventType,
			})
			return nil
		} else {
			return pkgerrors.Errorf("PostEvent Failed: %v", postEventRes.PostEventMessage)
		}
	}
	return err
}
