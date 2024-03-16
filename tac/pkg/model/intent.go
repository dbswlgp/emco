// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package model

import (
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
	tmpl "gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/emcotemporalapi"
	wfMod "gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/module"
)

// WorkflowHookIntent contains the parameters needed to create a workflow hook
type WorkflowHookIntent struct {
	// Intent Metadata
	Metadata mtypes.Metadata `json:"metadata,omitempty"`

	// Workflow Hook Type
	Spec WorkflowHookSpec `json:"spec"`
}

// Workflow Hook specs have the specifications needed to create a hook
type WorkflowHookSpec struct {
	// What kind of hook that this is.
	HookType string `json:"hookType"` // (pre/post)-(install/update/terminate)
	// Network endpoint at which the workflow client resides.
	WfClientSpec wfMod.WfClientSpec `json:"workflowClient"`
	// See emcotemporalapi package.
	WfTemporalSpec tmpl.WfTemporalSpec `json:"temporal"`
}

// WorkflowHookKey is the key structure that is used in the database
type WorkflowHookKey struct {
	WorkflowHook        string `json:"workflowHookIntent"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	DigName             string `json:"deploymentIntentGroup"`
}

type WfhTemporalCancelRequest struct {
	Metadata mtypes.Metadata              `json:"metadata,omitempty"`
	Spec     WfhTemporalCancelRequestSpec `json:"spec"`
}

type WfhTemporalCancelRequestSpec struct {
	// The Temporal server's endpoint. E.g. "temporal.foo.com:7233". Required.
	TemporalServer string `json:"temporalServer"`
	// If WfID is specified, that overrides the one in the workflow intent.
	WfID  string `json:"workflowID,omitempty"`
	RunID string `json:"runID,omitempty"`
	// If Terminate == true, TerminateWorkflow() is called, else CancelWorkflow().
	Terminate bool          `json:"terminate,omitempty"`
	Reason    string        `json:"reason,omitempty"`
	Details   []interface{} `json:"details,omitempty"`
}
