// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package networkintents

import (
	clusterPkg "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	ncmtypes "gitlab.com/project-emco/core/emco-base/src/ncm/pkg/module/types"
	nettypes "gitlab.com/project-emco/core/emco-base/src/ncm/pkg/networkintents/types"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"context"

	pkgerrors "github.com/pkg/errors"
)

// ProviderNet contains the parameters needed for dynamic networks
type ProviderNet struct {
	Metadata mtypes.Metadata `json:"metadata"`
	Spec     ProviderNetSpec `json:"spec"`
}

type ProviderNetSpec struct {
	CniType         string                `json:"cniType" yaml:"cniType"`
	Ipv4Subnets     []nettypes.Ipv4Subnet `json:"ipv4Subnets,omitempty" yaml:"ipv4Subnets,omitempty"`
	Ipv6Subnets     []nettypes.Ipv6Subnet `json:"ipv6Subnets,omitempty" yaml:"ipv6Subnets,omitempty"`
	ProviderNetType string                `json:"providerNetType" yaml:"providerNetType"`
	Vlan            nettypes.Vlan         `json:"vlan" yaml:"vlan"`
}

// structure for the Network Custom Resource
type CrProviderNet struct {
	ApiVersion      string            `yaml:"apiVersion"`
	Kind            string            `yaml:"kind"`
	MetaData        metav1.ObjectMeta `yaml:"metadata"`
	ProviderNetSpec ProviderNetSpec   `yaml:"spec"`
}

const PROVIDER_NETWORK_APIVERSION = "k8s.plugin.opnfv.org/v1alpha1"
const PROVIDER_NETWORK_KIND = "ProviderNetwork"

// ProviderNetKey is the key structure that is used in the database
type ProviderNetKey struct {
	ClusterProviderName string `json:"clusterProvider"`
	ClusterName         string `json:"cluster"`
	ProviderNetName     string `json:"providerNetwork"`
}

// Manager is an interface exposing the ProviderNet functionality
type ProviderNetManager interface {
	CreateProviderNet(ctx context.Context, pr ProviderNet, clusterProvider, cluster string, exists bool) (ProviderNet, error)
	GetProviderNet(ctx context.Context, name, clusterProvider, cluster string) (ProviderNet, error)
	GetProviderNets(ctx context.Context, clusterProvider, cluster string) ([]ProviderNet, error)
	DeleteProviderNet(ctx context.Context, name, clusterProvider, cluster string) error
}

// ProviderNetClient implements the Manager
// It will also be used to maintain some localized state
type ProviderNetClient struct {
	db ncmtypes.ClientDbInfo
}

// NewProviderNetClient returns an instance of the ProviderNetClient
// which implements the Manager
func NewProviderNetClient() *ProviderNetClient {
	return &ProviderNetClient{
		db: ncmtypes.ClientDbInfo{
			StoreName: "resources",
			TagMeta:   "data",
		},
	}
}

// CreateProviderNet - create a new ProviderNet
func (v *ProviderNetClient) CreateProviderNet(ctx context.Context, p ProviderNet, clusterProvider, cluster string, exists bool) (ProviderNet, error) {

	// verify cluster exists and in state to add provider networks
	s, err := clusterPkg.NewClusterClient().GetClusterState(ctx, clusterProvider, cluster)
	if err != nil {
		return ProviderNet{}, err
	}
	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return ProviderNet{}, pkgerrors.Wrap(err, "Error getting current state from Cluster stateInfo: "+cluster)
	}
	switch stateVal {
	case state.StateEnum.Approved:
		return ProviderNet{}, pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Approved)
	case state.StateEnum.Terminated:
		break
	case state.StateEnum.Created:
		break
	case state.StateEnum.Applied:
		return ProviderNet{}, pkgerrors.Errorf("Existing cluster provider network intents must be terminated before creating: " + cluster)
	case state.StateEnum.Instantiated:
		return ProviderNet{}, pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Instantiated)
	default:
		return ProviderNet{}, pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + stateVal)
	}

	//Construct key and tag to select the entry
	key := ProviderNetKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		ProviderNetName:     p.Metadata.Name,
	}

	//Check if this ProviderNet already exists
	_, err = v.GetProviderNet(ctx, p.Metadata.Name, clusterProvider, cluster)
	if err == nil && !exists {
		return ProviderNet{}, pkgerrors.New("ProviderNet already exists")
	}

	err = db.DBconn.Insert(ctx, v.db.StoreName, key, nil, v.db.TagMeta, p)
	if err != nil {
		return ProviderNet{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetProviderNet returns the ProviderNet for corresponding name
func (v *ProviderNetClient) GetProviderNet(ctx context.Context, name, clusterProvider, cluster string) (ProviderNet, error) {

	//Construct key and tag to select the entry
	key := ProviderNetKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		ProviderNetName:     name,
	}

	value, err := db.DBconn.Find(ctx, v.db.StoreName, key, v.db.TagMeta)
	if err != nil {
		return ProviderNet{}, err
	}

	if len(value) == 0 {
		return ProviderNet{}, pkgerrors.New("ProviderNet not found")
	}

	//value is a byte array
	if value != nil {
		cp := ProviderNet{}
		err = db.DBconn.Unmarshal(value[0], &cp)
		if err != nil {
			return ProviderNet{}, err
		}
		return cp, nil
	}

	return ProviderNet{}, pkgerrors.New("Unknown Error")
}

// GetProviderNetList returns all of the ProviderNet for corresponding name
func (v *ProviderNetClient) GetProviderNets(ctx context.Context, clusterProvider, cluster string) ([]ProviderNet, error) {

	//Construct key and tag to select the entry
	key := ProviderNetKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		ProviderNetName:     "",
	}

	var resp []ProviderNet
	values, err := db.DBconn.Find(ctx, v.db.StoreName, key, v.db.TagMeta)
	if err != nil {
		return []ProviderNet{}, err
	}

	for _, value := range values {
		cp := ProviderNet{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ProviderNet{}, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// Delete the  ProviderNet from database
func (v *ProviderNetClient) DeleteProviderNet(ctx context.Context, name, clusterProvider, cluster string) error {
	// verify cluster is in a state where provider network intent can be deleted
	s, err := clusterPkg.NewClusterClient().GetClusterState(ctx, clusterProvider, cluster)
	if err != nil {
		return err
	}
	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting current state from Cluster stateInfo: "+cluster)
	}
	switch stateVal {
	case state.StateEnum.Approved:
		return pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Approved)
	case state.StateEnum.Terminated:
		break
	case state.StateEnum.TerminateStopped:
		break
	case state.StateEnum.Created:
		break
	case state.StateEnum.Applied:
		return pkgerrors.Errorf("Cluster provider network intents must be terminated before deleting: " + cluster)
	case state.StateEnum.InstantiateStopped:
		return pkgerrors.Errorf("Cluster provider network intents must be terminated before deleting: " + cluster)
	case state.StateEnum.Instantiated:
		return pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Instantiated)
	default:
		return pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + stateVal)
	}

	//Construct key and tag to select the entry
	key := ProviderNetKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		ProviderNetName:     name,
	}

	err = db.DBconn.Remove(ctx, v.db.StoreName, key)
	return err
}
