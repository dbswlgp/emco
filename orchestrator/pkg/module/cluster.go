// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2022 Intel Corporation

package module

import (
	"context"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

type ClusterClient struct {
	storeName string
	tagMeta   string
}

func NewClusterClient() *ClusterClient {
	return &ClusterClient{
		storeName: "resources",
		tagMeta:   "data",
	}
}

// NOTE: this method is a duplicate of the identically-named one in dcm/pkg/module/cluster.go
// due to current cross-reference (cyclic dependency) issue between DCM and Orchestrator
func (v *ClusterClient) GetAllClusters(ctx context.Context, project, logicalCloud string) ([]common.Cluster, error) {
	//Construct the composite key to select clusters
	key := common.ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: "",
	}
	var resp []common.Cluster
	values, err := db.DBconn.Find(ctx, v.storeName, key, v.tagMeta)
	if err != nil {
		return []common.Cluster{}, err
	}
	if len(values) == 0 {
		return []common.Cluster{}, pkgerrors.New("No Cluster References associated")
	}

	for _, value := range values {
		cl := common.Cluster{}
		err = db.DBconn.Unmarshal(value, &cl)
		if err != nil {
			return []common.Cluster{}, err
		}
		resp = append(resp, cl)
	}

	return resp, nil
}
