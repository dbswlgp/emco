// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// WorkloadIntent contains the parameters needed for dynamic networks
type WorkloadIntent struct {
	Metadata Metadata           `json:"metadata"`
	Spec     WorkloadIntentSpec `json:"spec"`
}

type WorkloadIntentSpec struct {
	AppName          string `json:"app"`
	WorkloadResource string `json:"workloadResource"`
	Type             string `json:"type"`
}

// WorkloadIntentKey is the key structure that is used in the database
type WorkloadIntentKey struct {
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	DigName             string `json:"deploymentIntentGroup"`
	NetControlIntent    string `json:"netControllerIntent"`
	WorkloadIntent      string `json:"workloadIntent"`
}

// Manager is an interface exposing the WorkloadIntent functionality
type WorkloadIntentManager interface {
	CreateWorkloadIntent(ctx context.Context, wi WorkloadIntent, project, compositeapp, compositeappversion, dig, netcontrolintent string, exists bool) (WorkloadIntent, error)
	GetWorkloadIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig, netcontrolintent string) (WorkloadIntent, error)
	GetWorkloadIntents(ctx context.Context, project, compositeapp, compositeappversion, dig, netcontrolintent string) ([]WorkloadIntent, error)
	DeleteWorkloadIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig, netcontrolintent string) error
}

// WorkloadIntentClient implements the Manager
// It will also be used to maintain some localized state
type WorkloadIntentClient struct {
	db ClientDbInfo
}

// NewWorkloadIntentClient returns an instance of the WorkloadIntentClient
// which implements the Manager
func NewWorkloadIntentClient() *WorkloadIntentClient {
	return &WorkloadIntentClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}

// CreateWorkloadIntent - create a new WorkloadIntent
func (v *WorkloadIntentClient) CreateWorkloadIntent(ctx context.Context, wi WorkloadIntent, project, compositeapp, compositeappversion, dig, netcontrolintent string, exists bool) (WorkloadIntent, error) {

	//Construct key and tag to select the entry
	key := WorkloadIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      wi.Metadata.Name,
	}

	//Check if this WorkloadIntent already exists
	_, err := v.GetWorkloadIntent(ctx, wi.Metadata.Name, project, compositeapp, compositeappversion, dig, netcontrolintent)
	if err == nil && !exists {
		return WorkloadIntent{}, pkgerrors.New("WorkloadIntent already exists")
	}

	err = db.DBconn.Insert(ctx, v.db.storeName, key, nil, v.db.tagMeta, wi)
	if err != nil {
		return WorkloadIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return wi, nil
}

// GetWorkloadIntent returns the WorkloadIntent for corresponding name
func (v *WorkloadIntentClient) GetWorkloadIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig, netcontrolintent string) (WorkloadIntent, error) {

	//Construct key and tag to select the entry
	key := WorkloadIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      name,
	}

	value, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return WorkloadIntent{}, err
	}

	if len(value) == 0 {
		return WorkloadIntent{}, pkgerrors.New("WorkloadIntent not found")
	}

	//value is a byte array
	if value != nil {
		wi := WorkloadIntent{}
		err = db.DBconn.Unmarshal(value[0], &wi)
		if err != nil {
			return WorkloadIntent{}, err
		}
		return wi, nil
	}

	return WorkloadIntent{}, pkgerrors.New("Unknown Error")
}

// GetWorkloadIntentList returns all of the WorkloadIntent for corresponding name
func (v *WorkloadIntentClient) GetWorkloadIntents(ctx context.Context, project, compositeapp, compositeappversion, dig, netcontrolintent string) ([]WorkloadIntent, error) {

	//Construct key and tag to select the entry
	key := WorkloadIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      "",
	}

	var resp []WorkloadIntent
	values, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []WorkloadIntent{}, err
	}

	for _, value := range values {
		wi := WorkloadIntent{}
		err = db.DBconn.Unmarshal(value, &wi)
		if err != nil {
			return []WorkloadIntent{}, err
		}
		resp = append(resp, wi)
	}

	return resp, nil
}

// Delete the  WorkloadIntent from database
func (v *WorkloadIntentClient) DeleteWorkloadIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig, netcontrolintent string) error {

	//Construct key and tag to select the entry
	key := WorkloadIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      name,
	}

	err := db.DBconn.Remove(ctx, v.db.storeName, key)
	return err
}
