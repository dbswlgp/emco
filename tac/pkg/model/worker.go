// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package model

import mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"

type WorkerIntent struct {
	// Worker Metadata
	Metadata mtypes.Metadata `json:"metadata"`

	// Worker Spec
	Spec WorkerSpec `json:"spec"`
}

type WorkerSpec struct {
	StartToCloseTimeout int    `json:"startToCloseTimeout"`
	DIG                 string `json:"deploymentIntentGroup"`
	CApp                string `json:"compositeApp"`
	CAppVersion         string `json:"compositeAppVersion"`
}

type WorkerKey struct {
	WorkerName          string `json:"workerIntent"`
	WorkflowHook        string `json:"workflowHookIntent"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	DigName             string `json:"deploymentIntentGroup"`
}
