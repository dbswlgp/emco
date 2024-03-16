// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package clmcontrollereventchannelclient

import (
	"context"
	"time"

	pkgerrors "github.com/pkg/errors"
	clmcontrollerpb "gitlab.com/project-emco/core/emco-base/src/clm/pkg/grpc/controller-eventchannel"
	clmModel "gitlab.com/project-emco/core/emco-base/src/clm/pkg/model"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/rpc"
)

// SendControllerEvent ..  will make the grpc call to the specified controller
func SendControllerEvent(ctx context.Context, providerName string, clusterName string, event clmcontrollerpb.ClmControllerEventType, clmCtrl clmModel.Controller) error {
	controllerName := clmCtrl.Metadata.Name
	log.Info("SendControllerEvent .. start", log.Fields{"provider-name": providerName, "cluster-name": clusterName, "event": event, "controller": clmCtrl})
	var err error
	var rpcClient clmcontrollerpb.ClmControllerEventChannelClient
	var ctrlRes *clmcontrollerpb.ClmControllerEventResponse
	ctx, cancel := context.WithTimeout(ctx, 600*time.Second)
	defer cancel()

	// Fetch Grpc Connection handle
	conn := rpc.GetRpcConn(ctx, clmCtrl.Metadata.Name)
	if conn != nil {
		rpcClient = clmcontrollerpb.NewClmControllerEventChannelClient(conn)
		ctrlReq := new(clmcontrollerpb.ClmControllerEventRequest)
		ctrlReq.ProviderName = providerName
		ctrlReq.ClusterName = clusterName
		ctrlReq.Event = event
		log.Info("SendControllerEvent .. Sending event", log.Fields{"controller": clmCtrl, "ctrlReq": ctrlReq})
		ctrlRes, err = rpcClient.Publish(ctx, ctrlReq)
		if err == nil {
			log.Info("Response from SendControllerEvent GRPC call", log.Fields{"status": ctrlRes.Status, "message": ctrlRes.Message})
		}
	} else {
		log.Error("SendControllerEvent Failed - Could not get client connection to grpc-server.", log.Fields{"controllerName": controllerName})
		return pkgerrors.Errorf("SendControllerEvent Failed - Could not get client connection to grpc-server[%v]", controllerName)
	}

	if err == nil {
		if ctrlRes.Status {
			log.Info("SendControllerEvent Successful", log.Fields{
				"provider-name": providerName,
				"cluster-name":  clusterName,
				"event":         event,
				"controller":    clmCtrl})
			return nil
		}
		log.Error("SendControllerEvent UnSuccessful - Received message", log.Fields{"message": ctrlRes.Message, "controllerName": controllerName})
		return pkgerrors.Errorf("SendControllerEvent UnSuccessful - Received message: %v", ctrlRes.Message)
	}
	log.Error("SendControllerEvent Failed - Received error message", log.Fields{"message": ctrlRes.Message, "controllerName": controllerName})
	return err
}
