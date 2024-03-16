// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package controller

import (
	event "emcopolicy/internal/events"
	"emcopolicy/internal/intent"
	events "emcopolicy/pkg/grpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"golang.org/x/net/context"
	"sync"
)

type Controller struct {
	policyClient    *intent.Client
	eventClient     *event.Client
	db              db.Store
	tag             string
	storeName       string
	reverseMap      *ReverseMap
	agentMap        *AgentMap
	eventStream     chan event.Event
	agentStream     chan event.StreamAgentData
	updateStream    chan intent.StreamData
	requireRecovery chan interface{}
	eventsQueue     chan *events.Event
	actors          map[string]event.Actor
	contextDb       contextdb.ContextDb
}

type AgentID string

// ReverseMap contains Events to Policy Intent mapping
// This is the core in-memory data structure of policy controller
// This mapping helps to easily loop through all policy intents that is registered
// for a particular event
//
// Appends to both eventMap and Intent list is not thread-safe
// But updates to these data structures will be rare compared to the reads.
// Explicit Locking could be efficient than sync.Map
type ReverseMap struct {
	eventMap map[intent.Event][]intent.Intent
	sync.RWMutex
}

type AgentMap struct {
	runtime map[AgentID]*AgentRuntime
	sync.RWMutex
}

type AgentRuntime struct {
	spec      event.AgentSpec
	cancel    context.CancelFunc
	isRunning bool
}

type Module struct {
	PolicyModule string `json:"policyModule"`
}

type ContextMeta struct {
	Project               string `json:"Project"`
	CompositeApp          string `json:"CompositeApp"`
	Version               string `json:"Version"`
	DeploymentIntentGroup string `json:"DeploymentIntentGroup"`
}
