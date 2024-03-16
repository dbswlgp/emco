// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	"context"
	pkgerrors "github.com/pkg/errors"
)

// WorkloadIfIntent contains the parameters needed for dynamic networks
type WorkloadIfIntent struct {
	Metadata Metadata             `json:"metadata"`
	Spec     WorkloadIfIntentSpec `json:"spec"`
}

type WorkloadIfIntentSpec struct {
	IfName         string `json:"interface"`
	NetworkName    string `json:"name"`
	DefaultGateway string `json:"defaultGateway"`       // optional, default value is "false"
	IpAddr         string `json:"ipAddress,omitempty"`  // optional, if not provided then will be dynamically allocated
	MacAddr        string `json:"macAddress,omitempty"` // optional, if not provided then will be dynamically allocated
}

// WorkloadIfIntentKey is the key structure that is used in the database
type WorkloadIfIntentKey struct {
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	DigName             string `json:"deploymentIntentGroup"`
	NetControlIntent    string `json:"netControllerIntent"`
	WorkloadIntent      string `json:"workloadIntent"`
	WorkloadIfIntent    string `json:"interfaceIntent"`
}

// Manager is an interface exposing the WorkloadIfIntent functionality
type WorkloadIfIntentManager interface {
	CreateWorkloadIfIntent(ctx context.Context, wi WorkloadIfIntent, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string, exists bool) (WorkloadIfIntent, error)
	GetWorkloadIfIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string) (WorkloadIfIntent, error)
	GetWorkloadIfIntents(ctx context.Context, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string) ([]WorkloadIfIntent, error)
	DeleteWorkloadIfIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string) error
}

// WorkloadIfIntentClient implements the Manager
// It will also be used to maintain some localized state
type WorkloadIfIntentClient struct {
	db ClientDbInfo
}

// NewWorkloadIfIntentClient returns an instance of the WorkloadIfIntentClient
// which implements the Manager
func NewWorkloadIfIntentClient() *WorkloadIfIntentClient {
	return &WorkloadIfIntentClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}

// CreateWorkloadIfIntent - create a new WorkloadIfIntent
func (v *WorkloadIfIntentClient) CreateWorkloadIfIntent(ctx context.Context, wif WorkloadIfIntent, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string, exists bool) (WorkloadIfIntent, error) {

	//Construct key and tag to select the entry
	key := WorkloadIfIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      workloadintent,
		WorkloadIfIntent:    wif.Metadata.Name,
	}

	//Check if this WorkloadIfIntent already exists
	_, err := v.GetWorkloadIfIntent(ctx, wif.Metadata.Name, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent)
	if err == nil && !exists {
		return WorkloadIfIntent{}, pkgerrors.New("WorkloadIfIntent already exists")
	}

	err = db.DBconn.Insert(ctx, v.db.storeName, key, nil, v.db.tagMeta, wif)
	if err != nil {
		return WorkloadIfIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return wif, nil
}

// GetWorkloadIfIntent returns the WorkloadIfIntent for corresponding name
func (v *WorkloadIfIntentClient) GetWorkloadIfIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string) (WorkloadIfIntent, error) {

	//Construct key and tag to select the entry
	key := WorkloadIfIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      workloadintent,
		WorkloadIfIntent:    name,
	}

	value, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return WorkloadIfIntent{}, err
	}

	if len(value) == 0 {
		return WorkloadIfIntent{}, pkgerrors.New("WorkloadIfIntent not found")
	}

	//value is a byte array
	if value != nil {
		wif := WorkloadIfIntent{}
		err = db.DBconn.Unmarshal(value[0], &wif)
		if err != nil {
			return WorkloadIfIntent{}, err
		}
		return wif, nil
	}

	return WorkloadIfIntent{}, pkgerrors.New("Unknown Error")
}

// GetWorkloadIfIntentList returns all of the WorkloadIfIntent for corresponding name
func (v *WorkloadIfIntentClient) GetWorkloadIfIntents(ctx context.Context, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string) ([]WorkloadIfIntent, error) {

	//Construct key and tag to select the entry
	key := WorkloadIfIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      workloadintent,
		WorkloadIfIntent:    "",
	}

	var resp []WorkloadIfIntent
	values, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []WorkloadIfIntent{}, err
	}

	for _, value := range values {
		wif := WorkloadIfIntent{}
		err = db.DBconn.Unmarshal(value, &wif)
		if err != nil {
			return []WorkloadIfIntent{}, err
		}
		resp = append(resp, wif)
	}

	return resp, nil
}

// Delete the  WorkloadIfIntent from database
func (v *WorkloadIfIntentClient) DeleteWorkloadIfIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string) error {

	//Construct key and tag to select the entry
	key := WorkloadIfIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      workloadintent,
		WorkloadIfIntent:    name,
	}

	err := db.DBconn.Remove(ctx, v.db.storeName, key)
	return err
}
