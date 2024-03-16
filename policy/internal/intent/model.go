// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package intent

import (
	"encoding/json"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

type Client struct {
	db           db.Store
	tag          string
	storeName    string
	updateStream chan StreamData
}

type Config struct {
	Db           db.Store
	Tag          string
	StoreName    string
	UpdateStream chan StreamData
}

type PolicySpec struct {
	EngineUrl  string `json:"engineUrl"`
	PolicyName string `json:"policyName"`
}

type Spec struct {
	PolicyIntentID        string           `json:"policyIntentID"`
	Project               string           `json:"project"`
	CompositeApp          string           `json:"compositeApp"`
	CompositeAppVersion   string           `json:"compositeAppVersion"`
	DeploymentIntentGroup string           `json:"deploymentIntentGroup"`
	Policy                PolicySpec       `json:"policy"`
	Actor                 string           `json:"actor"`
	ActorArg              *json.RawMessage `json:"actorArg,omitempty"`
	Event                 Event            `json:"event"`
	SupportingEvents      []Event          `json:"supportingEvent,omitempty"`
}

// Event defines an event/metrics. Controller identifies using this id.
// Redefining events here to avoid import cycle. Originally defined in Events package
// TODO: Refactor this to avoid import cycle and multiple definitions
type Event struct {
	Id      string `json:"id"`
	AgentID string `json:"agent,omitempty"`
}

type Intent struct {
	Metadata Metadata `json:"metadata"`
	Spec     Spec     `json:"spec"`
}

type Metadata struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"-"`
	UserData1   string `json:"userData1" yaml:"-"`
	UserData2   string `json:"userData2" yaml:"-"`
}

type Request struct {
	Project               string
	CompositeApp          string
	CompositeAppVersion   string
	DeploymentIntentGroup string
	PolicyIntentId        string
	IntentData            *Intent
}

type Key struct {
	PolicyIntent        string `json:"policyIntent"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	DigName             string `json:"deploymentIntentGroup"`
}

type StreamData struct {
	Operation string
	Intent    Intent
}
