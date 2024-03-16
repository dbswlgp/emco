// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"encoding/json"

	hpaModel "gitlab.com/project-emco/core/emco-base/src/hpa-plc/pkg/model"
)

// HpaPlacementManager .. Manager is an interface exposing the HpaPlacementIntent functionality
type HpaPlacementManager interface {
	// intents
	AddIntent(ctx context.Context, a hpaModel.DeploymentHpaIntent, p string, ca string, v string, di string, exists bool) (hpaModel.DeploymentHpaIntent, error)
	GetIntent(ctx context.Context, i string, p string, ca string, v string, di string) (hpaModel.DeploymentHpaIntent, bool, error)
	GetAllIntents(ctx context.Context, p, ca, v, di string) ([]hpaModel.DeploymentHpaIntent, error)
	GetAllIntentsByApp(ctx context.Context, app, p, ca, v, di string) ([]hpaModel.DeploymentHpaIntent, error)
	GetIntentByName(ctx context.Context, i, p, ca, v, di string) (hpaModel.DeploymentHpaIntent, error)
	DeleteIntent(ctx context.Context, i string, p string, ca string, v string, di string) error

	// consumers
	AddConsumer(ctx context.Context, a hpaModel.HpaResourceConsumer, p string, ca string, v string, di string, i string, exists bool) (hpaModel.HpaResourceConsumer, error)
	GetConsumer(ctx context.Context, cn string, p string, ca string, v string, di string, i string) (hpaModel.HpaResourceConsumer, bool, error)
	GetAllConsumers(ctx context.Context, p, ca, v, di, i string) ([]hpaModel.HpaResourceConsumer, error)
	GetConsumerByName(ctx context.Context, cn, p, ca, v, di, i string) (hpaModel.HpaResourceConsumer, error)
	DeleteConsumer(ctx context.Context, cn, p string, ca string, v string, di string, i string) error

	// resources
	AddResource(ctx context.Context, a hpaModel.HpaResourceRequirement, p string, ca string, v string, di string, i string, cn string, exists bool) (hpaModel.HpaResourceRequirement, error)
	GetResource(ctx context.Context, rn string, p string, ca string, v string, di string, i string, cn string) (hpaModel.HpaResourceRequirement, bool, error)
	GetAllResources(ctx context.Context, p, ca, v, di, i, cn string) ([]hpaModel.HpaResourceRequirement, error)
	GetResourceByName(ctx context.Context, rn, p, ca, v, di, i, cn string) (hpaModel.HpaResourceRequirement, error)
	DeleteResource(ctx context.Context, rn string, p string, ca string, v string, di string, i string, cn string) error
}

// HpaPlacementClient implements the HpaPlacementManager interface
type HpaPlacementClient struct {
	db hpaModel.ClientDBInfo
}

// NewHpaPlacementClient returns an instance of the HpaPlacementClient
func NewHpaPlacementClient() *HpaPlacementClient {
	return &HpaPlacementClient{
		db: hpaModel.ClientDBInfo{
			StoreName:   "resources",
			TagMetaData: "data",
			TagContent:  "HpaPlacementControllerContent",
			TagState:    "HpaPlacementControllerStateInfo",
		},
	}
}

// HpaIntentKey ... consists of intent name, Project name, CompositeApp name,
// CompositeApp version, deployment intent group
type HpaIntentKey struct {
	IntentName            string `json:"hpaIntent"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeApp"`
	Version               string `json:"compositeAppVersion"`
	DeploymentIntentGroup string `json:"deploymentIntentGroup"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ik HpaIntentKey) String() string {
	out, err := json.Marshal(ik)
	if err != nil {
		return ""
	}

	return string(out)
}

// HpaConsumerKey ... consists of Name if the Consumer name, Project name, CompositeApp name,
// CompositeApp version, Deployment intent group, Intent name
type HpaConsumerKey struct {
	ConsumerName          string `json:"hpaConsumer"`
	IntentName            string `json:"hpaIntent"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeApp"`
	Version               string `json:"compositeAppVersion"`
	DeploymentIntentGroup string `json:"deploymentIntentGroup"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ck HpaConsumerKey) String() string {
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
	ConsumerName          string `json:"hpaConsumer"`
	IntentName            string `json:"hpaIntent"`
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
