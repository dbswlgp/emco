// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2022 Intel Corporation

package module

import (
	"context"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"

	pkgerrors "github.com/pkg/errors"
)

type LogicalCloudClient struct {
	storeName string
	tagMeta   string
	tagState  string
}

func NewLogicalCloudClient() *LogicalCloudClient {
	return &LogicalCloudClient{
		storeName: "resources",
		tagMeta:   "data",
		tagState:  "stateInfo",
	}
}

// NOTE: this method is a duplicate of the identically-named one in dcm/pkg/module/logicalcloud.go
// due to current cross-reference (cyclic dependency) issue between DCM and Orchestrator
func (v *LogicalCloudClient) Get(ctx context.Context, project, logicalCloudName string) (common.LogicalCloud, error) {

	//Construct the composite key to select the entry
	key := common.LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloudName,
	}
	value, err := db.DBconn.Find(ctx, v.storeName, key, v.tagMeta)
	if err != nil {
		return common.LogicalCloud{}, err
	}

	if len(value) == 0 {
		return common.LogicalCloud{}, pkgerrors.New("Logical Cloud not found")
	}

	//value is a byte array
	if value != nil {
		lc := common.LogicalCloud{}
		err = db.DBconn.Unmarshal(value[0], &lc)
		if err != nil {
			return common.LogicalCloud{}, err
		}
		return lc, nil
	}

	return common.LogicalCloud{}, pkgerrors.New("Unknown Error")
}

// NOTE: this method is a duplicate of the identically-named one in dcm/pkg/module/logicalcloud.go
// due to current cross-reference (cyclic dependency) issue between DCM and Orchestrator
func (v *LogicalCloudClient) GetState(ctx context.Context, p string, lc string) (state.StateInfo, error) {

	key := common.LogicalCloudKey{
		Project:          p,
		LogicalCloudName: lc,
	}

	result, err := db.DBconn.Find(ctx, v.storeName, key, v.tagState)
	if err != nil {
		return state.StateInfo{}, err
	}

	if len(result) == 0 {
		return state.StateInfo{}, pkgerrors.New("LogicalCloud StateInfo not found")
	}

	if result != nil {
		s := state.StateInfo{}
		err = db.DBconn.Unmarshal(result[0], &s)
		if err != nil {
			return state.StateInfo{}, err
		}
		return s, nil
	}

	return state.StateInfo{}, pkgerrors.New("Unknown Error")
}
