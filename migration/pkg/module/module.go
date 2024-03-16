// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"encoding/json"

	migrationModel "gitlab.com/project-emco/core/emco-base/src/migration/pkg/model"
)

// MigrationPlacementManager .. Manager is an interface exposing the HpaPlacementIntent functionality
type MigrationPlacementManager interface {
	// intents
	AddIntent(ctx context.Context, a migrationModel.DeploymentMigrationIntent, p string, ca string, v string, di string, exists bool) (migrationModel.DeploymentMigrationIntent, error)
	GetIntent(ctx context.Context, i string, p string, ca string, v string, di string) (migrationModel.DeploymentMigrationIntent, bool, error)
	GetAllIntents(ctx context.Context, p, ca, v, di string) ([]migrationModel.DeploymentMigrationIntent, error)
	GetAllIntentsByApp(ctx context.Context, app, p, ca, v, di string) ([]migrationModel.DeploymentMigrationIntent, error)
	GetIntentByName(ctx context.Context, i, p, ca, v, di string) (migrationModel.DeploymentMigrationIntent, error)
	DeleteIntent(ctx context.Context, i string, p string, ca string, v string, di string) error

	UpdateIntent(ctx context.Context, migrationIntentName string, p string, ca string, v string, di string, fieldupdate string) (/*migrationModel.DeploymentMigrationIntent,*/ error)

	// MigrationPolicy
	AddPolicy(ctx context.Context, a migrationModel.MigrationPolicy, p string, exists bool) (migrationModel.MigrationPolicy, error)
	GetPolicy(ctx context.Context, pn string, p string) (migrationModel.MigrationPolicy, bool, error)
	GetAllPolicies(ctx context.Context, p string) ([]migrationModel.MigrationPolicy, error)
	GetPolicyByName(ctx context.Context, pn, p string) (migrationModel.MigrationPolicy, error)
	DeletePolicy(ctx context.Context, pn string, p string) error


	// AppIntent
	AddAppIntent(ctx context.Context, a migrationModel.MigrationAppIntent, p string, ca string, v string, di string, i string, exists bool) (migrationModel.MigrationAppIntent, error)
	GetAppIntent(ctx context.Context, cn string, p string, ca string, v string, di string, i string) (migrationModel.MigrationAppIntent, bool, error)
	GetAllAppIntents(ctx context.Context, p, ca, v, di, i string) ([]migrationModel.MigrationAppIntent, error)
	GetAppIntentByName(ctx context.Context, cn, p, ca, v, di, i string) (migrationModel.MigrationAppIntent, error)
	DeleteAppIntent(ctx context.Context, cn, p string, ca string, v string, di string, i string) error
}

// MigrationClient implements the MigrationPlacementManager interface
type MigrationClient struct {
	db migrationModel.ClientDBInfo
}

// NewMigrationClient returns an instance of the MigrationClient
func NewMigrationClient() *MigrationClient {
	return &MigrationClient{
		db: migrationModel.ClientDBInfo{
			StoreName:   "resources",
			TagMetaData: "data",
			TagContent:  "HpaPlacementControllerContent",
			TagState:    "HpaPlacementControllerStateInfo",
		},
	}
}

// MigrationIntentKey ... consists of intent name, Project name, CompositeApp name,
// CompositeApp version, deployment intent group
type MigrationIntentKey struct {
	IntentName            string `json:"migrationIntent"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeApp"`
	Version               string `json:"compositeAppVersion"`
	DeploymentIntentGroup string `json:"deploymentIntentGroup"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ik MigrationIntentKey) String() string {
	out, err := json.Marshal(ik)
	if err != nil {
		return ""
	}

	return string(out)
}

// MigrationPolicyKey ... consists of policy name, Project name
type MigrationPolicyKey struct {
	PolicyName            string `json:"migrationPolicy"`
	Project               string `json:"project"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ik MigrationPolicyKey) String() string {
	out, err := json.Marshal(ik)
	if err != nil {
		return ""
	}

	return string(out)
}

// MigrationAppIntentKey ... consists of Name if the Consumer name, Project name, CompositeApp name,
// CompositeApp version, Deployment intent group, Intent name
type MigrationAppIntentKey struct {
	AppIntentName         string `json:"migrationAppIntent"`
	IntentName            string `json:"migrationIntent"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeApp"`
	Version               string `json:"compositeAppVersion"`
	DeploymentIntentGroup string `json:"deploymentIntentGroup"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ck MigrationAppIntentKey) String() string {
	out, err := json.Marshal(ck)
	if err != nil {
		return ""
	}

	return string(out)
}

// HpaResourceKey ... consists of Name of the Resource name, Project name, CompositeApp name,
// CompositeApp version, Deployment intent group, Intent name, Consumer name
type HpaResourceKey struct {
	ResourceName          string `json:"hpaResource"`
	AppIntentName         string `json:"migrationAppIntent"`
	IntentName            string `json:"migrationIntent"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeApp"`
	Version               string `json:"compositeAppVersion"`
	DeploymentIntentGroup string `json:"deploymentIntentGroup"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (rk HpaResourceKey) String() string {
	out, err := json.Marshal(rk)
	if err != nil {
		return ""
	}

	return string(out)
}
