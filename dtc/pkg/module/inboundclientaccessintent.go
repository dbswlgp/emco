// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"context"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

type InboundClientsAccessIntent struct {
	Metadata Metadata                       `json:"metadata"`
	Spec     InboundClientsAccessIntentSpec `json:"spec"`
}

type InboundClientsAccessIntentSpec struct {
	Action string   `json:"action"`
	Url    []string `json:"url"`
	Access []string `json:"access"`
}

type InboundClientsAccessIntentManager interface {
	CreateClientsAccessInboundIntent(ctx context.Context, tci InboundClientsAccessIntent, project, compositeapp, compositeappversion, deploymentIntentGroupName, trafficintentgroupName, inboundintentname, inboundclientsintentname string, exists bool) (InboundClientsAccessIntent, error)
	GetClientsAccessInboundIntents(ctx context.Context, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundintentname, inboundclientsintentname string) ([]InboundClientsAccessIntent, error)
	GetClientsAccessInboundIntent(ctx context.Context, name, project, compositeapp, compositeappversion, deploymentintentGroupName, trafficintentgroupname, inboundintentname, inboundclientsintentname string) (InboundClientsAccessIntent, error)
	DeleteClientsAccessInboundIntent(ctx context.Context, name, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname, inboundclientsintentname string) error
}

type InboundClientsAccessIntentDbClient struct {
	db ClientDbInfo
}

// ClientsAccessInboundIntentKey is the key structure that is used in the database
type InboundClientsAccessIntentKey struct {
	Project                        string `json:"project"`
	CompositeApp                   string `json:"compositeApp"`
	CompositeAppVersion            string `json:"compositeAppVersion"`
	DeploymentIntentGroupName      string `json:"deploymentIntentGroup"`
	TrafficGroupIntentName         string `json:"trafficGroupIntent"`
	InboundServerIntentName        string `json:"inboundServerIntent"`
	InboundClientsIntentName       string `json:"inboundClientsIntent"`
	InboundClientsAccessIntentName string `json:"inboundClientsAccessIntent"`
}

func NewClientsAccessInboundIntentClient() *InboundClientsAccessIntentDbClient {
	return &InboundClientsAccessIntentDbClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}

func (v InboundClientsAccessIntentDbClient) CreateClientsAccessInboundIntent(ctx context.Context, icai InboundClientsAccessIntent, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname, inboundclientsintentname string, exists bool) (InboundClientsAccessIntent, error) {

	//Construct key and tag to select the entry
	key := InboundClientsAccessIntentKey{
		Project:                        project,
		CompositeApp:                   compositeapp,
		CompositeAppVersion:            compositeappversion,
		DeploymentIntentGroupName:      deploymentintentgroupname,
		TrafficGroupIntentName:         trafficintentgroupname,
		InboundServerIntentName:        inboundserverintentname,
		InboundClientsIntentName:       inboundclientsintentname,
		InboundClientsAccessIntentName: icai.Metadata.Name,
	}

	//Check if this InboundClientsAccessIntent already exists
	_, err := v.GetClientsAccessInboundIntent(ctx, icai.Metadata.Name, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname, inboundclientsintentname)
	if err == nil && !exists {
		return InboundClientsAccessIntent{}, pkgerrors.New("InboundClientsAccessIntent already exists")
	}

	err = db.DBconn.Insert(ctx, v.db.storeName, key, nil, v.db.tagMeta, icai)
	if err != nil {
		return InboundClientsAccessIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return icai, nil

}

// GetClientsAccessInboundIntent returns the InboundClientsAccessIntent
func (v *InboundClientsAccessIntentDbClient) GetClientsAccessInboundIntent(ctx context.Context, name, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname, inboundclientsintentname string) (InboundClientsAccessIntent, error) {

	//Construct key and tag to select the entry
	key := InboundClientsAccessIntentKey{
		Project:                        project,
		CompositeApp:                   compositeapp,
		CompositeAppVersion:            compositeappversion,
		DeploymentIntentGroupName:      deploymentintentgroupname,
		TrafficGroupIntentName:         trafficintentgroupname,
		InboundServerIntentName:        inboundserverintentname,
		InboundClientsIntentName:       inboundclientsintentname,
		InboundClientsAccessIntentName: name,
	}

	value, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return InboundClientsAccessIntent{}, err
	} else if len(value) == 0 {
		return InboundClientsAccessIntent{}, pkgerrors.New("Inbound clients access intent not found")
	}

	//value is a byte array
	if value != nil {
		ici := InboundClientsAccessIntent{}
		err = db.DBconn.Unmarshal(value[0], &ici)
		if err != nil {
			return InboundClientsAccessIntent{}, err
		}
		return ici, nil
	}

	return InboundClientsAccessIntent{}, pkgerrors.New("Unknown Error")
}

// GetClientsAccessInboundIntents returns all of the InboundClientsAccessIntent for corresponding name
func (v *InboundClientsAccessIntentDbClient) GetClientsAccessInboundIntents(ctx context.Context, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname, inboundclientsintentname string) ([]InboundClientsAccessIntent, error) {

	//Construct key and tag to select the entry
	key := InboundClientsAccessIntentKey{
		Project:                        project,
		CompositeApp:                   compositeapp,
		CompositeAppVersion:            compositeappversion,
		DeploymentIntentGroupName:      deploymentintentgroupname,
		TrafficGroupIntentName:         trafficintentgroupname,
		InboundServerIntentName:        inboundserverintentname,
		InboundClientsIntentName:       inboundclientsintentname,
		InboundClientsAccessIntentName: "",
	}

	var resp []InboundClientsAccessIntent
	values, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []InboundClientsAccessIntent{}, err
	}

	for _, value := range values {
		icai := InboundClientsAccessIntent{}
		err = db.DBconn.Unmarshal(value, &icai)
		if err != nil {
			return []InboundClientsAccessIntent{}, err
		}
		resp = append(resp, icai)
	}

	return resp, nil

}

// Delete the ClientsInboundAccessIntent from database
func (v *InboundClientsAccessIntentDbClient) DeleteClientsAccessInboundIntent(ctx context.Context, name, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname, inboundclientsintentname string) error {

	//Construct key and tag to select the entry
	key := InboundClientsAccessIntentKey{
		Project:                        project,
		CompositeApp:                   compositeapp,
		CompositeAppVersion:            compositeappversion,
		DeploymentIntentGroupName:      deploymentintentgroupname,
		TrafficGroupIntentName:         trafficintentgroupname,
		InboundServerIntentName:        inboundserverintentname,
		InboundClientsIntentName:       inboundclientsintentname,
		InboundClientsAccessIntentName: name,
	}

	err := db.DBconn.Remove(ctx, v.db.storeName, key)
	return err
}
