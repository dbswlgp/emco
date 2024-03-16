// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

type InboundClientsIntent struct {
	Metadata Metadata                 `json:"metadata"`
	Spec     InboundClientsIntentSpec `json:"spec"`
}

type InboundClientsIntentSpec struct {
	AppName     string   `json:"app"`
	AppLabel    string   `json:"appLabel"`
	ServiceName string   `json:"serviceName"`
	Namespaces  []string `json:"namespaces"`
	IpRange     []string `json:"cidrs"`
}

type InboundClientsIntentManager interface {
	CreateClientsInboundIntent(ctx context.Context, tci InboundClientsIntent, project, compositeapp, compositeappversion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName string, exists bool) (InboundClientsIntent, error)
	GetClientsInboundIntents(ctx context.Context, project, compositeapp, compositeappversion, deploymentIntentGroupName, trafficintentgroupname, inboundIntentName string) ([]InboundClientsIntent, error)
	GetClientsInboundIntent(ctx context.Context, name, project, compositeapp, compositeappversion, deploymentIntentGroupName, trafficintentgroupname, inboundIntentName string) (InboundClientsIntent, error)
	DeleteClientsInboundIntent(ctx context.Context, name, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname string) error
}

type InboundClientsIntentDbClient struct {
	db ClientDbInfo
}

// ClientsInboundIntentKey is the key structure that is used in the database
type InboundClientsIntentKey struct {
	Project                   string `json:"project"`
	CompositeApp              string `json:"compositeApp"`
	CompositeAppVersion       string `json:"compositeAppVersion"`
	DeploymentIntentGroupName string `json:"deploymentIntentGroup"`
	TrafficGroupIntentName    string `json:"trafficGroupIntent"`
	InboundServerIntentName   string `json:"inboundServerIntent"`
	InboundClientsIntentName  string `json:"inboundClientsIntent"`
}

func NewClientsInboundIntentClient() *InboundClientsIntentDbClient {
	return &InboundClientsIntentDbClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}

func (v InboundClientsIntentDbClient) CreateClientsInboundIntent(ctx context.Context, ici InboundClientsIntent, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname string, exists bool) (InboundClientsIntent, error) {

	//Construct key and tag to select the entry
	key := InboundClientsIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: deploymentintentgroupname,
		TrafficGroupIntentName:    trafficintentgroupname,
		InboundServerIntentName:   inboundserverintentname,
		InboundClientsIntentName:  ici.Metadata.Name,
	}

	//Check if this InboundClientsIntent already exists
	_, err := v.GetClientsInboundIntent(ctx, ici.Metadata.Name, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname)
	if err == nil && !exists {
		return InboundClientsIntent{}, pkgerrors.New("InboundClientsIntent already exists")
	}

	err = db.DBconn.Insert(ctx, v.db.storeName, key, nil, v.db.tagMeta, ici)
	if err != nil {
		return InboundClientsIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return ici, nil

}

// GetClientsInboundIntent returns the InboundClientsIntent
func (v *InboundClientsIntentDbClient) GetClientsInboundIntent(ctx context.Context, name, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname string) (InboundClientsIntent, error) {

	//Construct key and tag to select the entry
	key := InboundClientsIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: deploymentintentgroupname,
		TrafficGroupIntentName:    trafficintentgroupname,
		InboundServerIntentName:   inboundserverintentname,
		InboundClientsIntentName:  name,
	}

	value, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return InboundClientsIntent{}, err
	} else if len(value) == 0 {
		return InboundClientsIntent{}, pkgerrors.New("Inbound clients intent not found")
	}

	//value is a byte array
	if value != nil {
		ici := InboundClientsIntent{}
		err = db.DBconn.Unmarshal(value[0], &ici)
		if err != nil {
			return InboundClientsIntent{}, err
		}
		return ici, nil
	}

	return InboundClientsIntent{}, pkgerrors.New("Unknown Error")
}

// GetClientsInboundIntents returns all of the InboundClientsIntent for corresponding name
func (v *InboundClientsIntentDbClient) GetClientsInboundIntents(ctx context.Context, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname string) ([]InboundClientsIntent, error) {

	//Construct key and tag to select the entry
	key := InboundClientsIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: deploymentintentgroupname,
		TrafficGroupIntentName:    trafficintentgroupname,
		InboundServerIntentName:   inboundserverintentname,
		InboundClientsIntentName:  "",
	}

	var resp []InboundClientsIntent
	values, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []InboundClientsIntent{}, err
	}

	for _, value := range values {
		ici := InboundClientsIntent{}
		err = db.DBconn.Unmarshal(value, &ici)
		if err != nil {
			return []InboundClientsIntent{}, err
		}
		resp = append(resp, ici)
	}

	return resp, nil

}

// Delete the  ClientsInboundIntent from database
func (v *InboundClientsIntentDbClient) DeleteClientsInboundIntent(ctx context.Context, name, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname string) error {

	//Construct key and tag to select the entry
	key := InboundClientsIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: deploymentintentgroupname,
		TrafficGroupIntentName:    trafficintentgroupname,
		InboundServerIntentName:   inboundserverintentname,
		InboundClientsIntentName:  name,
	}

	err := db.DBconn.Remove(ctx, v.db.storeName, key)
	return err
}
