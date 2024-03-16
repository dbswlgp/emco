// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package controller

import (
	"context"
	event "emcopolicy/internal/events"
	"emcopolicy/internal/intent"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"time"
)

// BuildReverseMap builds a reverse map of events to policy.
// Should be called only during boot up, when no other controller threads are running
func (c *Controller) BuildReverseMap(ctx context.Context) error {
	policyIntents, err := c.getAllPolicyIntents(ctx)
	if err != nil {
		return errors.Wrap(err, "Building In-memory ReverseMap failed")
	}
	c.reverseMap.Lock()
	defer c.reverseMap.Unlock()
	for _, policyIntent := range policyIntents {
		if c.reverseMap.eventMap[policyIntent.Spec.Event] == nil {
			c.reverseMap.eventMap[policyIntent.Spec.Event] = []intent.Intent{}
		}
		c.reverseMap.eventMap[policyIntent.Spec.Event] = append(c.reverseMap.eventMap[policyIntent.Spec.Event], policyIntent)
	}
	return nil
}

// BuildAgentMap builds the runtime map of agents from db.
// AgentManager should be started before calling this method, otherwise this will block
func (c *Controller) BuildAgentMap(ctx context.Context) error {
	agents, err := c.getAllAgents(ctx)
	if err != nil {
		return errors.Wrap(err, "BuildAgentMap failed")
	}
	if agents == nil {
		log.Warn("No Agents found in DB", log.Fields{})
		return nil
	}
	c.agentMap.Lock()
	defer c.agentMap.Unlock()
	for _, agent := range agents {
		runtime := &AgentRuntime{
			spec: agent,
		}
		c.agentMap.runtime[AgentID(agent.Id)] = runtime
	}
	c.requireRecovery <- struct{}{}
	return nil
}

// AddPolicyIntent adds an intent from event's intentList in eventMap
func (m *ReverseMap) AddPolicyIntent(e intent.Event, p intent.Intent) {
	var index int
	m.Lock()
	defer m.Unlock()
	// Remove Intent from list if this is an update.
	// Deletion can be handled by  RemovePolicyIntent, but it's safer to do while lock is holding
	// This lock could starve other appends and reads, but it is unavoidable in current design.
	// This delay won't be user visible since this is a background activity
	// Also appends will be less often compared to reads
	for index < len(m.eventMap[e]) && !isEqual(m.eventMap[e][index], p) {
		index++
	}
	if index < len(m.eventMap[e]) {
		m.eventMap[e] = append(m.eventMap[e][:index], m.eventMap[e][index+1:]...)
	}
	if _, ok := m.eventMap[e]; !ok {
		m.eventMap[e] = []intent.Intent{}
	}
	m.eventMap[e] = append(m.eventMap[e], p)
	log.Info("Added Intent to in-memory event list", log.Fields{"Event": e.Id, "Intent": p.Spec.PolicyIntentID})
}

// RemovePolicyIntent deletes an intent from event's intentList in eventMap
func (m *ReverseMap) RemovePolicyIntent(e intent.Event, p intent.Intent) error {
	var index int
	m.Lock()
	defer m.Unlock()
	for index < len(m.eventMap[e]) && !isEqual(m.eventMap[e][index], p) {
		index++
	}
	if index == len(m.eventMap[e]) {
		i := p.Spec.Project + "/" + p.Spec.CompositeApp + "/" + p.Spec.CompositeAppVersion + "/" +
			p.Spec.DeploymentIntentGroup + "/" + p.Spec.PolicyIntentID
		return errors.Errorf("RemoveIntent failed. No Policy Intent "+
			"(%s) found in the IntentList of event %s", i, e.Id)
	}
	m.eventMap[e] = append(m.eventMap[e][:index], m.eventMap[e][index+1:]...)
	log.Info("Removed Intent from in-memory event list", log.Fields{"Event": e.Id, "Intent": p.Spec.PolicyIntentID})
	return nil
}

// GetPolicyIntentList returns intentList of an event from eventMap
func (m *ReverseMap) GetPolicyIntentList(e intent.Event) (error, []intent.Intent) {
	m.RLock()
	defer m.RUnlock()
	if _, ok := m.eventMap[e]; !ok {
		// TODO error
		return errors.Errorf("GetPolicyIntentList failed. Event %s not found", e.Id), nil
	}
	intentCopy := make([]intent.Intent, len(m.eventMap[e]))
	// slice can get modified from other threads
	// Hence returning a snapshot
	copy(intentCopy, m.eventMap[e])
	return nil, intentCopy
}

func (m *ReverseMap) GetIntents(event intent.Event, contextSpec ContextMeta) ([][]byte, error) {
	var (
		intents [][]byte
	)
	m.RLock()
	defer m.RUnlock()
	for _, i := range m.eventMap[event] {
		if !isSameContext(i.Spec, contextSpec) {
			continue
		}
		data, err := json.Marshal(i.Spec)
		intents = append(intents, data)
		if err != nil {
			return [][]byte{}, errors.Wrap(err, "GetIntents failed")
		}
	}
	return intents, nil
}

// isSameContext compare the intent with context.
func isSameContext(intentSpec intent.Spec, contextSpec ContextMeta) bool {
	return intentSpec.Project == contextSpec.Project &&
		intentSpec.CompositeApp == contextSpec.CompositeApp &&
		intentSpec.CompositeAppVersion == contextSpec.Version &&
		intentSpec.DeploymentIntentGroup == contextSpec.DeploymentIntentGroup
}

// Cancel method cancels the context.
// Required when deleting the agent, to stop the thread.
func (m *AgentMap) Cancel(id AgentID) {
	m.RLock()
	defer m.RUnlock()
	if _, ok := m.runtime[id]; ok {
		m.runtime[id].cancel()
	}
}

func (m *AgentMap) IsAgentExists(id AgentID) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.runtime[id]
	return ok
}

func (m *AgentMap) DeleteAgent(id AgentID) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.runtime[id]; ok {
		delete(m.runtime, id)
	}
}

func (m *AgentMap) VerifySpec(id AgentID, spec event.AgentSpec) bool {
	m.RLock()
	defer m.RUnlock()
	val, ok := m.runtime[id]
	return ok && cmp.Equal(spec, val.spec)
}

func (m *AgentMap) UpdateSpec(id AgentID, spec event.AgentSpec) {
	m.Lock()
	defer m.Unlock()
	m.runtime[id] = &AgentRuntime{
		spec: spec,
	}
}

// MarkForRecovery is the cancel function passed to the agent threads
// It marks the runtime data structure as thread exited
func (m *AgentMap) MarkForRecovery(id AgentID) {
	m.Lock()
	defer m.Unlock()
	if runtime, ok := m.runtime[id]; ok {
		// A small delay to avoid busy looping in case of agent errors
		// We can move an exponential algorithm in the future.
		time.Sleep(time.Second * 5)
		runtime.isRunning = false
	}
}

func isEqual(p, q intent.Intent) bool {
	return p.Spec.Project == q.Spec.Project &&
		p.Spec.CompositeApp == q.Spec.CompositeApp &&
		p.Spec.CompositeAppVersion == q.Spec.CompositeAppVersion &&
		p.Spec.DeploymentIntentGroup == q.Spec.DeploymentIntentGroup &&
		p.Spec.PolicyIntentID == q.Spec.PolicyIntentID
}
