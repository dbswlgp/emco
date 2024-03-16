// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package migrationcontrollerclient

import (
	"context"
	"time"

	pkgerrors "github.com/pkg/errors"
	misctrlclientpb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/migrationcontroller"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/rpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
)

// InvokeMigrationApps ..  will make the grpc call to the specified controller
func InvokeMigrationApps(ctx context.Context, misCtrl controller.Controller, appContextId string, app string) error {
	controllerName := misCtrl.Metadata.Name
	log.Info("MigrationApps .. start", log.Fields{"controllerName": controllerName, "Host": misCtrl.Spec.Host, "Port": misCtrl.Spec.Port, "appContextId": appContextId})

	var err error
	var rpcClient misctrlclientpb.MigrationControllerClient
	var ctrlRes *misctrlclientpb.MigrationResponse

	timeout := time.Duration(config.GetConfiguration().GrpcCallTimeout)
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Millisecond)
	defer cancel()


	// Fetch Grpc Connection handle
	conn := rpc.GetRpcConn(ctx, misCtrl.Metadata.Name)
	if conn != nil {
		rpcClient = misctrlclientpb.NewMigrationControllerClient(conn)
		ctrlReq := new(misctrlclientpb.MigrationRequest)
		ctrlReq.AppContext = appContextId
		ctrlReq.App = app
		ctrlRes, err = rpcClient.MigrationApps(ctx, ctrlReq)

		if err == nil {
			log.Info("Response from MigrationApps GRPC call", log.Fields{"status": ctrlRes.Status, "message": ctrlRes.Message})
		}
	} else {
		log.Error("MigrationApps Failed - Could not get client connection", log.Fields{"controllerName": controllerName, "appContextId": appContextId})
		return pkgerrors.Errorf("MigrationApps Failed - Could not get client connection. controllerName[%v] appContextId[%v]", controllerName, appContextId)
	}

	if err == nil {
		if ctrlRes.Status {
			log.Info("MigrationApps Successful", log.Fields{
				"Controller": controllerName,
				"AppContext": appContextId,
				"Message":    ctrlRes.Message})
			return nil
		}
		log.Error("MigrationApps UnSuccessful - Received message", log.Fields{"message": ctrlRes.Message, "controllerName": controllerName, "appContextId": appContextId})
		return pkgerrors.Errorf("MigrationApps UnSuccessful - Received message[%v] for controllerName[%v] appContextId[%v]", ctrlRes.Message, controllerName, appContextId)
	}
	log.Error("MigrationApps Failed - Received error message", log.Fields{"controllerName": controllerName, "appContextId": appContextId})


	return err
}
