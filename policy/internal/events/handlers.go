// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package event

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"net/http"
)

func (c Client) RegisterAgentHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	spec := new(AgentSpec)
	if err := json.NewDecoder(r.Body).Decode(spec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	spec.Id = v["id"]
	agent, err := c.RegisterAgent(ctx, v["id"], *spec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(agent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("Registering Agent processed successfully", log.Fields{"Agent ID": spec.Id})
}

func (c Client) GetAllAgentHandler(ctx context.Context, w http.ResponseWriter, _ *http.Request) {
	agents, err := c.GetAllAgents(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if agents == nil {
		agents = []AgentSpec{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(agents); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c Client) GetAgentHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	agent, err := c.GetAgent(ctx, v["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(agent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c Client) DeleteAgentHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	err := c.DeleteAgent(ctx, v["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
