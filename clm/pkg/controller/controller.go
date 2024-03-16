// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package controller

import (
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	rpc "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/rpc"

	"context"

	clmModel "gitlab.com/project-emco/core/emco-base/src/clm/pkg/model"
)

// ControllerManager is an interface exposes the Controller functionality
type ControllerManager interface {
	CreateController(ctx context.Context, ms clmModel.Controller, mayExist bool) (clmModel.Controller, error)
	GetController(ctx context.Context, name string) (clmModel.Controller, error)
	GetControllers(ctx context.Context) ([]clmModel.Controller, error)
	InitControllers(ctx context.Context)
	DeleteController(ctx context.Context, name string) error
}

// ControllerClient implements the Manager
// It will also be used to maintain some localized state
type ControllerClient struct {
	collectionName string
	tagMeta        string
	tagGroup       string
}

// NewControllerClient returns an instance of the ControllerClient
// which implements the Manager
func NewControllerClient() *ControllerClient {
	return &ControllerClient{
		collectionName: "resources",
		tagMeta:        "data",
		tagGroup:       "cluster",
	}
}

// CreateController a new collection based on the Controller
func (mc *ControllerClient) CreateController(ctx context.Context, m clmModel.Controller, mayExist bool) (clmModel.Controller, error) {

	log.Info("CLM CreateController .. start", log.Fields{"Controller": m, "exists": mayExist})

	//Construct the composite key to select the entry
	key := clmModel.ControllerKey{
		ControllerName:  m.Metadata.Name,
		ControllerGroup: mc.tagGroup,
	}

	//Check if this Controller already exists
	_, err := mc.GetController(ctx, m.Metadata.Name)
	if err == nil && !mayExist {
		return clmModel.Controller{}, pkgerrors.New("ClmController already exists")
	}

	err = db.DBconn.Insert(ctx, mc.collectionName, key, nil, mc.tagMeta, m)
	if err != nil {
		return clmModel.Controller{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	// send message to create/update the  rpc connection
	rpc.UpdateRpcConn(m.Metadata.Name, m.Spec.Host, m.Spec.Port)

	log.Info("CLM CreateController .. end", log.Fields{"Controller": m, "exists": mayExist})
	return m, nil
}

// GetController returns the Controller for corresponding name
func (mc *ControllerClient) GetController(ctx context.Context, name string) (clmModel.Controller, error) {

	//Construct the composite key to select the entry
	key := clmModel.ControllerKey{
		ControllerName:  name,
		ControllerGroup: mc.tagGroup,
	}
	value, err := db.DBconn.Find(ctx, mc.collectionName, key, mc.tagMeta)
	if err != nil {
		return clmModel.Controller{}, err
	}

	if len(value) == 0 {
		return clmModel.Controller{}, pkgerrors.New("ClmController not found")
	}

	if value != nil {
		microserv := clmModel.Controller{}
		err = db.DBconn.Unmarshal(value[0], &microserv)
		if err != nil {
			return clmModel.Controller{}, err
		}
		return microserv, nil
	}

	return clmModel.Controller{}, pkgerrors.New("Unknown Error")
}

// GetControllers returns all the  Controllers that are registered
func (mc *ControllerClient) GetControllers(ctx context.Context) ([]clmModel.Controller, error) {

	//Construct the composite key to select the entry
	key := clmModel.ControllerKey{
		ControllerName:  "",
		ControllerGroup: mc.tagGroup,
	}

	var resp []clmModel.Controller
	values, err := db.DBconn.Find(ctx, mc.collectionName, key, mc.tagMeta)
	if err != nil {
		return []clmModel.Controller{}, err
	}

	for _, value := range values {
		microserv := clmModel.Controller{}
		err = db.DBconn.Unmarshal(value, &microserv)
		if err != nil {
			return []clmModel.Controller{}, err
		}

		resp = append(resp, microserv)
	}

	return resp, nil
}

// DeleteController the  Controller from database
func (mc *ControllerClient) DeleteController(ctx context.Context, name string) error {

	//Construct the composite key to select the entry
	key := clmModel.ControllerKey{
		ControllerName:  name,
		ControllerGroup: mc.tagGroup,
	}
	err := db.DBconn.Remove(ctx, mc.collectionName, key)
	if err != nil {
		return err
	}

	// send message to close rpc connection
	rpc.RemoveRpcConn(name)

	return nil
}

// InitControllers initializes connctions for Controllers in the DB
func (mc *ControllerClient) InitControllers(ctx context.Context) {
	vals, _ := mc.GetControllers(ctx)
	for _, v := range vals {
		log.Info("Initializing RPC connection for Controller", log.Fields{
			"Controller": v.Metadata.Name,
		})
		rpc.UpdateRpcConn(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
	}
}
