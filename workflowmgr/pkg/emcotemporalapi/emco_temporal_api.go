// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package emcotemporalapi

// This package represents the API exported by EMCO for 3rd party workflows.
// A 3rd party workflow is expected to import this package + Temporal SDK.
// See docs/user/Temporal_Workflows_In_EMCO.md .

// TODO What about Temporal workflows in languages other than Go?
// TODO Version this API?

import (
	cl "go.temporal.io/sdk/client"
	wf "go.temporal.io/sdk/workflow"
)

// WfTemporalSpec is the specification needed to start a workflow.
// It is part of the EMCO workflow intent (see WorkflowIntentSpec in
// workflowmgr).
type WfTemporalSpec struct {
	// Name of the workflow client to invoke. Required.
	WfClientName string `json:"workflowClientName"`
	// Options needed by wf client to start a workflow. Workflow ID is required.
	WfStartOpts cl.StartWorkflowOptions `json:"workflowStartOptions"`
	// Parameters that the wf client needs to pass to the workflow. Optional.
	WfParams WorkflowParams `json:"workflowParams,omitempty"`
}

// WorkflowParams are the per-activity data that the wf client passes to a workflow.
type WorkflowParams struct {
	// map of Temporal activity options indexed by activity name
	ActivityOpts map[string]wf.ActivityOptions `json:"activityOptions,omitempty"`
	// map of wf-specific key-value pairs indexed by activity name
	ActivityParams map[string]map[string]string `json:"activityParams,omitempty"`
}
