// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package model

import (
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

// ClientDBInfo ... to save hpa data to db
type ClientDBInfo struct {
	StoreName   string // name of the mongodb collection to use for client documents
	TagMetaData string // attribute key name for the json data of a client document
	TagContent  string // attribute key name for the file data of a client document
	TagState    string // attribute key name for StateInfo object in the cluster
}

// DeploymentHpaIntent ..
type DeploymentMigrationIntent struct {
	// Intent Metadata
	MetaData mtypes.Metadata `json:"metadata,omitempty"`
	Status DeploymentMigrationIntentStatus `json:"status,omitempty"`
}

type DeploymentMigrationIntentStatus struct {
	Selected bool `json:"selected,omitempty"`
}

// HpaResourceConsumerSpec .. HpaIntent ResourceConsumer spec
type MigrationAppIntentSpec struct {
	// app name
	App string `json:"app,omitempty"`
	// migration
	Migration bool `json:"migration,omitempty"`
	// priority
	Priority int64 `json:"priority,omitempty"`
}

type MigrationAppIntentStatus struct {

	SelectedApp bool `json:"selectedApp,omitempty"`
	DeployedCluster string `json:"deployedCluster,omitempty"`

}

// HpaResourceConsumer .. Intent mapping to K8s
type MigrationAppIntent struct {
	// Intent Metadata
	MetaData mtypes.Metadata `json:"metadata,omitempty"`

	// Intent Spec
	Spec MigrationAppIntentSpec `json:"spec,omitempty"`

	Status MigrationAppIntentStatus `json:"status,omitempty"`
}

type MigrationPolicy struct {
	MetaData mtypes.Metadata `json:"metadata,omitempty"`
	Spec MigrationPolicySpec `json:"spec,omitempty"`
}

type MigrationPolicySpec struct {
	Considerations MigrationPolicyConsiderations `json:"considerations,omitempty"`
}

type MigrationPolicyConsiderations struct {
//	Priority ConsiderationSpec `json:"priority,omitempty"`
	CPUUtilization ConsiderationSpec `json:"cpuUtilization,omitempty"`
	Dispersion ConsiderationSpec `json:"dispersion,omitempty"`
}

type ConsiderationSpec struct {
	Active bool `json:"active,omitempty"`
	Order int64 `json:"order,omitempty"`
}


