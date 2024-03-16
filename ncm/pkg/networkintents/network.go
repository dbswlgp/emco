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

// Network contains the parameters needed for dynamic networks
type Network struct {
	Metadata mtypes.Metadata `json:"metadata" yaml:"metadata"`
	Spec     NetworkSpec     `json:"spec" yaml:"spec"`
}

type NetworkSpec struct {
	CniType     string                `json:"cniType" yaml:"cniType"`
	Ipv4Subnets []nettypes.Ipv4Subnet `json:"ipv4Subnets,omitempty" yaml:"ipv4Subnets,omitempty"`
	Ipv6Subnets []nettypes.Ipv6Subnet `json:"ipv6Subnets,omitempty" yaml:"ipv6Subnets,omitempty"`
}

// NetworkKey is the key structure that is used in the database
type NetworkKey struct {
	ClusterProviderName string `json:"clusterProvider"`
	ClusterName         string `json:"cluster"`
	NetworkName         string `json:"network"`
}

// structure for the Network Custom Resource
type CrNetwork struct {
	ApiVersion  string            `yaml:"apiVersion"`
	Kind        string            `yaml:"kind"`
	MetaData    metav1.ObjectMeta `yaml:"metadata"`
	NetworkSpec NetworkSpec       `yaml:"spec"`
}

const NETWORK_APIVERSION = "k8s.plugin.opnfv.org/v1alpha1"
const NETWORK_KIND = "Network"

// Manager is an interface exposing the Network functionality
type NetworkManager interface {
	CreateNetwork(ctx context.Context, pr Network, clusterProvider, cluster string, exists bool) (Network, error)
	GetNetwork(ctx context.Context, name, clusterProvider, cluster string) (Network, error)
	GetNetworks(ctx context.Context, clusterProvider, cluster string) ([]Network, error)
	DeleteNetwork(ctx context.Context, name, clusterProvider, cluster string) error
}

// NetworkClient implements the Manager
// It will also be used to maintain some localized state
type NetworkClient struct {
	db ncmtypes.ClientDbInfo
}

// NewNetworkClient returns an instance of the NetworkClient
// which implements the Manager
func NewNetworkClient() *NetworkClient {
	return &NetworkClient{
		db: ncmtypes.ClientDbInfo{
			StoreName: "resources",
			TagMeta:   "data",
		},
	}
}

// CreateNetwork - create a new Network
func (v *NetworkClient) CreateNetwork(ctx context.Context, p Network, clusterProvider, cluster string, exists bool) (Network, error) {

	//Construct key and tag to select the entry
	key := NetworkKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		NetworkName:         p.Metadata.Name,
	}

	//Check if cluster exists and in a state for adding network intents
	s, err := clusterPkg.NewClusterClient().GetClusterState(ctx, clusterProvider, cluster)
	if err != nil {
		return Network{}, err
	}
	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return Network{}, pkgerrors.Wrap(err, "Error getting current state from Cluster stateInfo: "+cluster)
	}
	switch stateVal {
	case state.StateEnum.Approved:
		return Network{}, pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Approved)
	case state.StateEnum.Terminated:
		break
	case state.StateEnum.Created:
		break
	case state.StateEnum.Applied:
		return Network{}, pkgerrors.Errorf("Existing cluster network intents must be terminated before creating: " + cluster)
	case state.StateEnum.Instantiated:
		return Network{}, pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Instantiated)
	default:
		return Network{}, pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + stateVal)
	}

	//Check if this Network already exists
	_, err = v.GetNetwork(ctx, p.Metadata.Name, clusterProvider, cluster)
	if err == nil && !exists {
		return Network{}, pkgerrors.New("Network already exists")
	}

	err = db.DBconn.Insert(ctx, v.db.StoreName, key, nil, v.db.TagMeta, p)
	if err != nil {
		return Network{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetNetwork returns the Network for corresponding name
func (v *NetworkClient) GetNetwork(ctx context.Context, name, clusterProvider, cluster string) (Network, error) {

	//Construct key and tag to select the entry
	key := NetworkKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		NetworkName:         name,
	}

	value, err := db.DBconn.Find(ctx, v.db.StoreName, key, v.db.TagMeta)
	if err != nil {
		return Network{}, err
	}

	if len(value) == 0 {
		return Network{}, pkgerrors.New("Network not found")
	}

	//value is a byte array
	if value != nil {
		cp := Network{}
		err = db.DBconn.Unmarshal(value[0], &cp)
		if err != nil {
			return Network{}, err
		}
		return cp, nil
	}

	return Network{}, pkgerrors.New("Unknown Error")
}

// GetNetworkList returns all of the Network for corresponding name
func (v *NetworkClient) GetNetworks(ctx context.Context, clusterProvider, cluster string) ([]Network, error) {

	//Construct key and tag to select the entry
	key := NetworkKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		NetworkName:         "",
	}

	var resp []Network
	values, err := db.DBconn.Find(ctx, v.db.StoreName, key, v.db.TagMeta)
	if err != nil {
		return []Network{}, err
	}

	for _, value := range values {
		cp := Network{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []Network{}, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// Delete the  Network from database
func (v *NetworkClient) DeleteNetwork(ctx context.Context, name, clusterProvider, cluster string) error {
	// verify cluster is in a state where network intent can be deleted
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
		return pkgerrors.Errorf("Cluster network intents must be terminated before deleting: " + cluster)
	case state.StateEnum.InstantiateStopped:
		return pkgerrors.Errorf("Cluster network intents must be terminated before deleting: " + cluster)
	case state.StateEnum.Instantiated:
		return pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Instantiated)
	default:
		return pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + stateVal)
	}

	//Construct key and tag to select the entry
	key := NetworkKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		NetworkName:         name,
	}

	err = db.DBconn.Remove(ctx, v.db.StoreName, key)
	return err
}
