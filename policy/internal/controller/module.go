// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package controller

import (
	"context"
	event "emcopolicy/internal/events"
	"emcopolicy/internal/intent"
	events "emcopolicy/pkg/grpc"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"time"
)

// OperationalScheduler listens to the different channels for management updates  and invoke
// corresponding  goroutines for handling them.
// Handle functions like handleIntentUpdate runs in OperationalScheduler context, not in
// the http request context, since those request will be completed once channel is read.
// There's possible corner case in current implementation, that if handling is failed,
// the updates will not get reflected in in-memory structure till next application boot.
func (c *Controller) OperationalScheduler(ctx context.Context) {
	var (
		p intent.StreamData
		a event.StreamAgentData
	)
	log.Info("Starting Scheduler. Listening to events", log.Fields{})
	for {
		select {
		case p = <-c.updateStream:
			go c.handleIntentUpdate(ctx, p)
		case a = <-c.agentStream:
			go c.handleAgentUpdate(ctx, a)
		case <-ctx.Done():
			log.Warn("OperationalScheduler exiting", log.Fields{"Content Err": ctx.Err()})
			return
		}
	}
}

// AgentManager manages the agent threads.  It runs one thread per registered
// agent. These threads start a rpc connection with agent it is managing.
// This method watch the status of these threads, and restart if required.
func (c *Controller) AgentManager(ctx context.Context) {
	log.Info("Starting Agent Manager", log.Fields{})
	for {
		var notRunning []AgentID
		checkAgents := func() {
			c.agentMap.RLock()
			defer c.agentMap.RUnlock()
			for id, runtime := range c.agentMap.runtime {
				if !runtime.isRunning {
					notRunning = append(notRunning, id)
				}
			}
		}
		select {
		case <-c.requireRecovery:
			log.Info("Require agent recovery. Checking agents", log.Fields{})
			checkAgents()
		case <-time.After(time.Minute * 5):
			log.Info("Scheduled Health Check of Agents", log.Fields{})
			checkAgents()
		}
		// Why this step of restart of agent not done along with checkAgents?
		// Number of agents to be started will a small number compared to total number of agents
		// In checkAgents, we traverse all the agents to identify which agents require attention.
		// This process require only read-lock, while update of agent require a write-lock.
		// Hence, to avoid holding write-lock for longer period, updating agents as a separate step
		// Updates that can happen in between these two critical session can be ignored, since next run
		// will take care of them.
		func() {
			c.agentMap.Lock()
			defer c.agentMap.Unlock()
			for _, id := range notRunning {
				var newCtx context.Context
				newCtx, c.agentMap.runtime[id].cancel = context.WithCancel(ctx)
				c.agentMap.runtime[id].isRunning = true
				go event.ListenOne(newCtx, c.agentMap.runtime[id].spec.EndPoint, func() {
					c.markForRecovery(id)
				}, c.eventsQueue)
			}
		}()
	}
}

// EventsManager continuously watch for new events from agents.
// New events are queued to eventsQueue by the agent threads.
// EventsManager dequeue them and process the event.
func (c *Controller) EventsManager(_ context.Context) {
	log.Info("Starting Events manager", log.Fields{})
	for e := range c.eventsQueue {
		go c.processEvent(e)
	}
}

// markForRecovery is  passed as cancellation function for agent threads
// It mark that thread for this agentID is closed.
// AgentManager will restart the thread in its next run.
// Major reason for agent thread closing is the connection getting closed
// from agent side.
func (c *Controller) markForRecovery(id AgentID) {
	c.agentMap.MarkForRecovery(id)
	c.requireRecovery <- struct{}{}
}

// handleAgentUpdate Append/Delete the agent update to the runtime data structure.
// This method should be called only after the updates are persisted.
func (c *Controller) handleAgentUpdate(_ context.Context, data event.StreamAgentData) {
	if c.agentMap == nil || c.agentMap.runtime == nil {
		log.Fatal("AgentMap or runtime is nil", log.Fields{})
	}
	switch data.Operation {
	case "DELETE":
		c.deleteAgent(data.Spec)
	case "APPEND":
		c.appendAgent(data.Spec)
	}

}

// handleIntentUpdate Append/Delete the intent update to the runtime data structure.
// This method should be called only after the updates are persisted.
func (c *Controller) handleIntentUpdate(ctx context.Context, data intent.StreamData) {
	switch data.Operation {
	case "DELETE":
		c.deleteIntent(ctx, data.Intent)
	case "APPEND":
		c.addIntent(ctx, data.Intent)
	}
}

func (c *Controller) deleteIntent(_ context.Context, p intent.Intent) {
	err := c.reverseMap.RemovePolicyIntent(p.Spec.Event, p)
	if err != nil {
		log.Error("Delete intent failed", log.Fields{"error": err})
	}
}

func (c *Controller) addIntent(_ context.Context, p intent.Intent) {
	c.reverseMap.AddPolicyIntent(p.Spec.Event, p)
}

// getAllPolicyIntents is helper function for getting all the policy intents from DB
// If there is no policy intent, it will return  ([]intent.Intent{}, nil)
func (c *Controller) getAllPolicyIntents(ctx context.Context) ([]intent.Intent, error) {
	var intents []intent.Intent
	key := struct {
		PolicyIntent bson.M `json:"policyIntent"`
	}{bson.M{"$exists": true}}
	value, err := c.db.Find(ctx, c.storeName, key, c.tag)
	if err != nil {
		return intents, err
	}
	for _, v := range value {
		item := new(intent.Intent)
		if err := c.db.Unmarshal(v, item); err != nil {
			return nil, err
		}
		intents = append(intents, *item)
	}

	return intents, nil
}

func (c *Controller) getAllAgents(ctx context.Context) ([]event.AgentSpec, error) {
	return c.eventClient.GetAllAgents(ctx)
}

func (c *Controller) deleteAgent(agent event.AgentSpec) {
	var (
		agentId = AgentID(agent.Id)
	)
	if !c.agentMap.IsAgentExists(agentId) {
		log.Warn("Delete failed: AgentID not found", log.Fields{"Id": agent.Id})
		return
	}
	c.agentMap.Cancel(agentId)
	c.agentMap.DeleteAgent(agentId)
}

func (c *Controller) appendAgent(agent event.AgentSpec) {
	var (
		agentId = AgentID(agent.Id)
	)
	if c.agentMap.VerifySpec(agentId, agent) {
		log.Warn("appendAgent Specs are same. Not updated", log.Fields{"AgentId": agentId})
		return
	}
	c.agentMap.Cancel(agentId)
	c.agentMap.UpdateSpec(agentId, agent)
	c.requireRecovery <- struct{}{}
}

// processEvent process the events from data. It does  the following steps
//   1. Decodes the events from protobuf format.
//   2. Converts event, intent, agentSpec into json format
//   3. Pass these data for further processing. (Policy evaluation and action)
func (c *Controller) processEvent(e *events.Event) {
	agentId := e.AgentId
	eventId := e.EventId
	contextId := e.ContextId
	log.Debug("Processing event", log.Fields{"id": eventId, "agent": agentId})
	// Convert agent spec from proto format to Json
	agentSpec, err := anypb.UnmarshalNew(e.Spec, proto.UnmarshalOptions{})
	if err != nil {
		log.Error("processEvent failed", log.Fields{"err": err, "id": eventId, "agent": agentId})
		return
	}
	agentSpecJson, err := json.Marshal(agentSpec)
	if err != nil {
		log.Error("processEvent failed", log.Fields{"err": err, "id": eventId, "agent": agentId})
		return
	}
	// Convert event message from proto format to Json
	//eventMessageJson, err := json.Marshal(e.MetricList)
	eventMessageJson := e.MetricList
	if err != nil {
		log.Error("processEvent failed", log.Fields{"err": err, "id": eventId, "agent": agentId})
		return
	}
	contextMeta, err := c.getContextMeta(contextId)
	if err != nil {
		log.Error("processEvent failed", log.Fields{"err": err, "id": eventId, "agent": agentId})
		return
	}
	// Process events with Agent ID mentioned in policy intent
	eventSpec := intent.Event{
		Id:      eventId,
		AgentID: agentId,
	}
	// Deep copying the intent. This allows remaining tasks,
	// which can be time-consuming to do without taking lock
	intentJsons, err := c.reverseMap.GetIntents(eventSpec, contextMeta)
	if err != nil {
		log.Error("processEvent failed", log.Fields{"err": err, "id": eventId, "agent": agentId})
		return
	}
	// Process events without Agent ID mentioned in policy intent
	eventSpec = intent.Event{
		Id:      eventId,
		AgentID: "",
	}

	temp, err := c.reverseMap.GetIntents(eventSpec, contextMeta)
	if err != nil {
		log.Error("processEvent failed", log.Fields{"err": err, "id": eventId, "agent": agentId})
		return
	}
	intentJsons = append(intentJsons, temp...)
	// Execute events for each attached intent
	log.Debug("Processing event", log.Fields{"Event": string(eventMessageJson)})
	fmt.Println("Processing event:", string(eventMessageJson))
	for _, intentJson := range intentJsons {
		go c.ExecuteEvent(intentJson, agentSpecJson, eventMessageJson)
	}
}

func (c *Controller) getContextMeta(contextId string) (ContextMeta, error) {
	contextMetaJson, err := c.getContextMetaFromDB(contextId)
	if err != nil {
		return ContextMeta{}, errors.Wrap(err, "Failed to get context metadata from contextDB")
	}
	var contextMeta ContextMeta
	err = json.Unmarshal(contextMetaJson, &contextMeta)
	if err != nil {
		return ContextMeta{}, errors.Wrap(err, "Failed to parse context metadata")
	}
	return contextMeta, err
}

// getContextMetaFromDB is a workaround for etcd library issue.
// Ideally we should use the ContextDB library provided by EMCO orchestrator
// But the etcd library that ContextDB using is not backward compatible and has conflict with this module.
// Hence, we are going for a separate service for reading from ContextDB
func (c *Controller) getContextMetaFromDB(contextId string) ([]byte, error) {
	key := "/context/" + contextId + "/meta/"
	var value json.RawMessage
	err := c.contextDb.Get(context.TODO(), key, value)
	if err != nil {
		log.Error("Error while getting context db data", log.Fields{"err": err})
		return nil, err
	}
	jsonValue, err := json.Marshal(value)
	if err != nil {
		log.Error("Error while parsing context db data", log.Fields{"err": err})
		return nil, err
	}
	return jsonValue, nil
}
