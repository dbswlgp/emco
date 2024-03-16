// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

type InboundServerIntent struct {
	Metadata Metadata               `json:"metadata"`
	Spec     InbondServerIntentSpec `json:"spec"`
}

type InbondServerIntentSpec struct {
	AppName         string                `json:"app"`
	AppLabel        string                `json:"appLabel"`
	ServiceName     string                `json:"serviceName"`
	ExternalName    string                `json:"externalName", default:""`
	Port            int                   `json:"port"`
	Protocol        string                `json:"protocol"`
	ExternalSupport bool                  `json:"externalSupport", default:false`
	ServiceMesh     string                `json:"serviceMesh", default:"none"`
	Management      ServiceMeshManagement `json:"management"`
	External        ExternalInfo          `json:"external"`
	EdgeCNF         string                `json:"edgeCNF", default: "none"`
}

type ServiceMeshManagement struct {
	SidecarProxy string `json:"sidecarProxy"`
	TlsType      string `json:"tlsType"`
}

type ExternalInfo struct {
	ExternalCerts ExternalCertInfo `json:"externalCerts"`
}
type ExternalCertInfo struct {
	ServiceCertificate string `json:"serviceCertificate"`
	ServicePrivateKey  string `json:"servicePrivateKey"`
	CaCertificate      string `json:"caCertificate"`
}

type InboundServerIntentManager interface {
	CreateServerInboundIntent(ctx context.Context, isi InboundServerIntent, project, compositeapp, compositeappversion, deploymentIntentGroupName, trafficintentgroupname string, exists bool) (InboundServerIntent, error)
	GetServerInboundIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig, trafficintentgroupname string) (InboundServerIntent, error)

	GetServerInboundIntents(ctx context.Context, project, compositeapp, compositeappversion, dig, intentName string) ([]InboundServerIntent, error)
	DeleteServerInboundIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig, trafficintentgroupname string) error
}

type InboundServerIntentDbClient struct {
	db ClientDbInfo
}

// ServerInboundIntentKey is the key structure that is used in the database
type InboundServerIntentKey struct {
	Project                   string `json:"project"`
	CompositeApp              string `json:"compositeApp"`
	CompositeAppVersion       string `json:"compositeAppVersion"`
	DeploymentIntentGroupName string `json:"deploymentIntentGroup"`
	TrafficGroupIntentName    string `json:"trafficGroupIntent"`
	ServerInboundIntentName   string `json:"inboundServerIntent"`
}

func NewServerInboundIntentClient() *InboundServerIntentDbClient {
	return &InboundServerIntentDbClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}

func (v InboundServerIntentDbClient) CreateServerInboundIntent(ctx context.Context, isi InboundServerIntent, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname string, exists bool) (InboundServerIntent, error) {

	//Construct key and tag to select the entry
	key := InboundServerIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: deploymentintentgroupname,
		TrafficGroupIntentName:    trafficintentgroupname,
		ServerInboundIntentName:   isi.Metadata.Name,
	}

	//Check if this ServerInboundIntent already exists
	_, err := v.GetServerInboundIntent(ctx, isi.Metadata.Name, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname)
	if err == nil && !exists {
		return InboundServerIntent{}, pkgerrors.New("ServerInboundIntent already exists")
	}

	err = db.DBconn.Insert(ctx, v.db.storeName, key, nil, v.db.tagMeta, isi)
	if err != nil {
		return InboundServerIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return isi, nil
}

// GetServerInboundIntent returns the ServerInboundIntent for corresponding name
func (v *InboundServerIntentDbClient) GetServerInboundIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig, trafficintentgroupname string) (InboundServerIntent, error) {

	//Construct key and tag to select the entry
	key := InboundServerIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: dig,
		TrafficGroupIntentName:    trafficintentgroupname,
		ServerInboundIntentName:   name,
	}

	value, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return InboundServerIntent{}, err
	} else if len(value) == 0 {
		return InboundServerIntent{}, pkgerrors.New("Inbound server intent not found")
	}

	//value is a byte array
	if value != nil {
		wi := InboundServerIntent{}
		err = db.DBconn.Unmarshal(value[0], &wi)
		if err != nil {
			return InboundServerIntent{}, err
		}
		return wi, nil
	}

	return InboundServerIntent{}, pkgerrors.New("Unknown Error")
}

// GetServerInboundIntents returns all of the ServerInboundIntents
func (v *InboundServerIntentDbClient) GetServerInboundIntents(ctx context.Context, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname string) ([]InboundServerIntent, error) {

	//Construct key and tag to select the entry
	key := InboundServerIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: deploymentintentgroupname,
		TrafficGroupIntentName:    trafficintentgroupname,
		ServerInboundIntentName:   "",
	}

	var resp []InboundServerIntent
	values, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []InboundServerIntent{}, err
	}

	for _, value := range values {
		is := InboundServerIntent{}
		err = db.DBconn.Unmarshal(value, &is)
		if err != nil {
			return []InboundServerIntent{}, err
		}
		resp = append(resp, is)
	}

	return resp, nil
}

// Delete the  ServerInboundIntents from database
func (v *InboundServerIntentDbClient) DeleteServerInboundIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig, trafficintentgroupname string) error {

	//Construct key and tag to select the entry
	key := InboundServerIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: dig,
		TrafficGroupIntentName:    trafficintentgroupname,
		ServerInboundIntentName:   name,
	}

	err := db.DBconn.Remove(ctx, v.db.storeName, key)
	return err
}
