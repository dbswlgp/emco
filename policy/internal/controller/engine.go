// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package controller

import (
	"bytes"
	event "emcopolicy/internal/events"
	"emcopolicy/internal/intent"
	"encoding/json"
	"github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"io/ioutil"
	"net/http"
)

// ExecuteEvent is the core of the Policy evaluation logic
// It evaluates the event by calling the policy engine, as per the intent.
// Evaluation result is passed to the actor. The actor plugin is responsible for
// taking the action, if required
func (c *Controller) ExecuteEvent(intentSpecJson []byte, agentSpec []byte, eventMessage []byte) {
	intentSpec := new(intent.Spec)
	if err := json.Unmarshal(intentSpecJson, intentSpec); err != nil {
		log.Error("ExecuteEvent failed", log.Fields{"err": err})
		return
	}
	type inputSpec struct {
		IntentSpec json.RawMessage `json:"intentSpec"`
		AgentSpec  json.RawMessage `json:"agentSpec"`
		Event      json.RawMessage `json:"event"`
	}
	policyInput := struct {
		Input inputSpec `json:"input"`
	}{inputSpec{
		IntentSpec: intentSpecJson,
		AgentSpec:  agentSpec,
		Event:      eventMessage,
	},
	}
	input, err := json.Marshal(policyInput)
	if err != nil {
		log.Error("ExecuteEvent failed", log.Fields{"err": err})
		return
	}
	policyUrl := "http://" + intentSpec.Policy.EngineUrl + "/" + intentSpec.Policy.PolicyName
	log.Debug("Evalutaing policy:", log.Fields{"input": string(input)})
	response, err := EvaluatePolicy(policyUrl, input)
	if err != nil {
		log.Error("ExecuteEvent failed", log.Fields{"err": err})
		return
	}
	if err := event.DoAction(intentSpec.Actor, response, intentSpecJson, agentSpec, c.actors); err != nil {
		log.Error("ExecuteEvent failed", log.Fields{"err": err})
		return
	}
}

// EvaluatePolicy Call the policy endpoint for evaluation.
// Will move this in a plugin model.
func EvaluatePolicy(url string, input []byte) ([]byte, error) {
	response, err := http.Post(url, "application/json", bytes.NewBuffer(input))
	if err != nil {
		return []byte{}, err
	}
	if response.StatusCode != http.StatusOK {
		log.Error("EvaluatePolicy failed due to http error", log.Fields{"status code": response.StatusCode, "url": url})
		return []byte{}, errors.Errorf("EvaluatePolicy failed due to http error. http Status code: %d", response.StatusCode)
	}
	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte{}, err
	}
	return responseData, nil
}
