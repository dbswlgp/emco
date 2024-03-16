// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package event

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

type Provider interface {
	listen()
}

type AgentSpec struct {
	Id       string `json:"id,omitempty"`
	EndPoint string `json:"addr"`
}

type AgentKey struct {
	PolicyModule string `json:"policyModule"`
	AgentID      string `json:"agent"`
}

func (c Client) RegisterAgent(ctx context.Context, id string, spec AgentSpec) (*AgentSpec, error) {
	key := AgentKey{
		"Agent",
		id,
	}
	err := c.db.Insert(ctx, c.storeName, key, nil, c.tag, spec)
	if err != nil {
		return nil, errors.Wrap(err, "Agent Registration failed")
	}
	c.agentStream <- StreamAgentData{
		Spec:      spec,
		Operation: "APPEND",
	}
	return &spec, nil
}

func (c Client) GetAllAgents(ctx context.Context) ([]AgentSpec, error) {
	var (
		agents []AgentSpec
	)
	key := struct {
		PolicyModule string `json:"policyModule"`
		AgentID      bson.M `json:"agent"`
	}{"Agent", bson.M{"$exists": true}}
	value, err := c.db.Find(ctx, c.storeName, key, c.tag)
	if err != nil {
		return nil, errors.Wrap(err, "GetAgents failed")
	}
	if value == nil || len(value) == 0 {
		return nil, nil
	}
	for _, v := range value {
		agent := new(AgentSpec)
		if err := c.db.Unmarshal(v, agent); err != nil {
			return agents, errors.Wrap(err, "GetAgents failed")
		}
		agents = append(agents, *agent)
	}
	return agents, nil
}

func (c Client) GetAgent(ctx context.Context, id string) (*AgentSpec, error) {
	agent := new(AgentSpec)
	key := AgentKey{
		"Agent",
		id,
	}
	value, err := c.db.Find(ctx, c.storeName, key, c.tag)

	if err != nil {
		return nil, errors.Wrap(err, "GetAgents failed")
	}
	if value == nil || len(value) == 0 {
		return nil, errors.Errorf("Agent with id: %v not found", id)
	}
	if err := c.db.Unmarshal(value[0], agent); err != nil {
		return nil, err
	}
	return agent, nil
}

func (c Client) DeleteAgent(ctx context.Context, id string) error {
	key := AgentKey{
		"Agent",
		id,
	}
	err := c.db.Remove(ctx, c.storeName, key)
	if err != nil {
		return errors.Wrap(err, "DeleteAgent failed")
	}
	c.agentStream <- StreamAgentData{
		Spec: AgentSpec{
			Id:       id,
			EndPoint: "",
		},
		Operation: "DELETE",
	}
	return nil
}
