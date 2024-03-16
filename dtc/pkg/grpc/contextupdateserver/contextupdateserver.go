// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package contextupdateserver

import (
	"context"
	"fmt"
	"sort"

	contextpb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdate"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	client "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdateclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
)

type contextupdateServer struct {
	contextpb.UnimplementedContextupdateServer
}

func (cs *contextupdateServer) UpdateAppContext(ctx context.Context, req *contextpb.ContextUpdateRequest) (*contextpb.ContextUpdateResponse, error) {
	log.Info("Received Update App Context request", log.Fields{
		"AppContextId": req.AppContext,
		"IntentName":   req.IntentName,
	})

	cc := controller.NewControllerClient("resources", "data", "dtc")
	clist, err := cc.GetControllers(ctx)
	if err != nil {
		log.Error("Error getting controllers", log.Fields{
			"error": err,
		})
		return &contextpb.ContextUpdateResponse{AppContextUpdated: false, AppContextUpdateMessage: fmt.Sprintf("Error getting controllers for intent %v and Id: %v", req.IntentName, req.AppContext)}, err
	}

	// Sort the list based on priority
	sort.SliceStable(clist, func(i, j int) bool {
		return clist[i].Spec.Priority > clist[j].Spec.Priority

	})

	for _, c := range clist {
		err := client.InvokeContextUpdate(ctx, c.Metadata.Name, req.IntentName, req.AppContext, "")
		if err != nil {
			log.Error("invoke context update failed for sub controller", log.Fields{
				"error": err,
				"Name":  c.Metadata.Name,
			})
		}
	}

	return &contextpb.ContextUpdateResponse{AppContextUpdated: true, AppContextUpdateMessage: fmt.Sprintf("Successful application of intent %v to %v", req.IntentName, req.AppContext)}, nil
}

// NewContextUpdateServer exported
func NewContextupdateServer() *contextupdateServer {
	s := &contextupdateServer{}
	return s
}
