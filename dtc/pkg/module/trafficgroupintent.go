// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

type TrafficGroupIntent struct {
	Metadata Metadata `json:"metadata"`
}

type TrafficGroupIntentManager interface {
	CreateTrafficGroupIntent(ctx context.Context, tci TrafficGroupIntent, project, compositeapp, compositeappversion, deploymentIntentGroupName string, exists bool) (TrafficGroupIntent, error)

	GetTrafficGroupIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig string) (TrafficGroupIntent, error)
	GetTrafficGroupIntents(ctx context.Context, project, compositeapp, compositeappversion, dig string) ([]TrafficGroupIntent, error)
	DeleteTrafficGroupIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig string) error
}

type TrafficGroupIntentDbClient struct {
	db ClientDbInfo
}

// TrafficGroupIntentKey is the key structure that is used in the database
type TrafficGroupIntentKey struct {
	TrafficGroupIntentName    string `json:"trafficGroupIntent"`
	Project                   string `json:"project"`
	CompositeApp              string `json:"compositeApp"`
	CompositeAppVersion       string `json:"compositeAppVersion"`
	DeploymentIntentGroupName string `json:"deploymentIntentGroup"`
}

func NewTrafficGroupIntentClient() *TrafficGroupIntentDbClient {
	return &TrafficGroupIntentDbClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}

func (v TrafficGroupIntentDbClient) CreateTrafficGroupIntent(ctx context.Context, tci TrafficGroupIntent, project, compositeapp, compositeappversion, deploymentintentgroupname string, exists bool) (TrafficGroupIntent, error) {

	//Construct key and tag to select the entry
	key := TrafficGroupIntentKey{
		TrafficGroupIntentName:    tci.Metadata.Name,
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: deploymentintentgroupname,
	}
	//Check if this TrafficGroupIntent already exists
	_, err := v.GetTrafficGroupIntent(ctx, tci.Metadata.Name, project, compositeapp, compositeappversion, deploymentintentgroupname)
	if err == nil && !exists {
		return TrafficGroupIntent{}, pkgerrors.New("TrafficGroupIntent already exists")
	}

	err = db.DBconn.Insert(ctx, v.db.storeName, key, nil, v.db.tagMeta, tci)
	if err != nil {
		return TrafficGroupIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return tci, nil
}

// GetTrafficGroupIntent returns the TrafficGroupIntent for corresponding name
func (v *TrafficGroupIntentDbClient) GetTrafficGroupIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig string) (TrafficGroupIntent, error) {

	//Construct key and tag to select the entry
	key := TrafficGroupIntentKey{
		TrafficGroupIntentName:    name,
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: dig,
	}

	value, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return TrafficGroupIntent{}, err
	} else if len(value) == 0 {
		return TrafficGroupIntent{}, pkgerrors.New("Traffic group intent not found")

	}

	//value is a byte array
	if value != nil {
		tgi := TrafficGroupIntent{}
		err = db.DBconn.Unmarshal(value[0], &tgi)
		if err != nil {
			return TrafficGroupIntent{}, err
		}
		return tgi, nil
	}

	return TrafficGroupIntent{}, pkgerrors.New("Unknown Error")

}

// GetTrafficGroupIntents returns all of the TrafficGroupIntents
func (v *TrafficGroupIntentDbClient) GetTrafficGroupIntents(ctx context.Context, project, compositeapp, compositeappversion, dig string) ([]TrafficGroupIntent, error) {

	//Construct key and tag to select the entry
	key := TrafficGroupIntentKey{
		TrafficGroupIntentName:    "",
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: dig,
	}

	var resp []TrafficGroupIntent
	values, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []TrafficGroupIntent{}, err
	}

	for _, value := range values {
		tgi := TrafficGroupIntent{}
		err = db.DBconn.Unmarshal(value, &tgi)
		if err != nil {
			return []TrafficGroupIntent{}, err
		}
		resp = append(resp, tgi)
	}

	return resp, nil
}

// Delete the  TrafficGroupIntent from database
func (v *TrafficGroupIntentDbClient) DeleteTrafficGroupIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig string) error {

	//Construct key and tag to select the entry
	key := TrafficGroupIntentKey{
		TrafficGroupIntentName:    name,
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: dig,
	}

	err := db.DBconn.Remove(ctx, v.db.storeName, key)
	return err
}
