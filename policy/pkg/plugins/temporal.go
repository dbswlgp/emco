// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package plugins

import (
	"bytes"
	"emcopolicy/internal/intent"
	"encoding/json"
	"github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"net/http"
)

type TemporalArgs struct {
	WorkFlowMgr  string `json:"workFlowMgr"`
	WorkFlowName string `json:"workFlowName"`
}

type Evaluation struct {
	Result EvaluationNameSpace `json:"result"`
}
type EvaluationNameSpace struct {
	Namespace EvaluationResult `json:"emco"`
}

type EvaluationResult struct {
	ActionRequired bool   `json:"actionRequired"`
	WorkflowName   string `json:"workflowName"`
}

type TemporalActor struct {
	WorkFlowMgrUrl string
}

func (t *TemporalActor) Execute(inputJson []byte, intentSpec []byte, _ []byte) error {
	//log.Info("In Temporal actor in plugin", log.Fields{"input": string(input),
	//	"intent": string(intentSpec), "agent": string(agentSpec)})
	log.Info("---In Temporal actor in plugin--", log.Fields{})
	log.Info("OPA Evaluation result", log.Fields{"result": string(inputJson)})
	log.Info("Intent Details", log.Fields{"Intent": string(intentSpec)})
	input := Evaluation{}
	err := json.Unmarshal(inputJson, &input)
	if !input.Result.Namespace.ActionRequired {
		log.Info("No Action required", log.Fields{})
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "Temporal Workflow Execution failed")
	}
	policyIntent := intent.Spec{}
	err = json.Unmarshal(intentSpec, &policyIntent)
	if err != nil {
		return errors.Wrap(err, "Temporal Workflow Execution failed")
	}
	workflowArgs := TemporalArgs{}
	err = json.Unmarshal(*policyIntent.ActorArg, &workflowArgs)
	if err != nil {
		return errors.Wrap(err, "Temporal Workflow Execution failed")
	}
	postBody := bytes.NewBuffer([]byte{})
	wfName := workflowArgs.WorkFlowName
	workFlowMgrUrl := workflowArgs.WorkFlowMgr
	if len(wfName) == 0 {
		return errors.Errorf("Workflow execution failed: Temporal workflow name is missing(Provide in policy Intent)")
	}
	if len(input.Result.Namespace.WorkflowName) > 0 {
		wfName = input.Result.Namespace.WorkflowName
	}
	dig := "/v2/projects/" + policyIntent.Project +
		"/composite-apps/" + policyIntent.CompositeApp + "/" + policyIntent.CompositeAppVersion +
		"/deployment-intent-groups/" + policyIntent.DeploymentIntentGroup
	startUrl := "http://" + workFlowMgrUrl + dig + "/temporal-workflow-intents/" + wfName + "/start"
	log.Info("Sending request to workflow manager", log.Fields{"startUrl": startUrl})
	response, err := http.Post(startUrl, "application/json", postBody)
	if err != nil {
		return errors.Wrap(err, "Temporal Workflow Execution failed")
	}
	if response.StatusCode != http.StatusCreated {
		return errors.Errorf("Temporal Workflow Execution failed. Couldn't start workflow %s", response.Status)
	}
	return nil
}
