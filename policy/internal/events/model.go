// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package event

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

type Client struct {
	db          db.Store
	tag         string
	storeName   string
	agentStream chan StreamAgentData
}

type Event struct {
	Id      string `json:"id"`
	AgentID string `json:"agent,omitempty"`
}

func NewClient(config Config) *Client {
	return &Client{
		db:          config.Db,
		tag:         config.Tag,
		storeName:   config.StoreName,
		agentStream: config.AgentStream,
	}
}

type Request struct {
	Dummy string
}

type Config struct {
	Db          db.Store
	Tag         string
	StoreName   string
	AgentStream chan StreamAgentData
}

type StreamAgentData struct {
	Spec      AgentSpec
	Operation string
}

type Actor interface {
	Execute(evaluationResult []byte, intentSpec []byte, agentSpec []byte) error
}
