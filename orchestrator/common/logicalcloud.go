// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2022 Intel Corporation

package common

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

// LogicalCloud contains the parameters needed for a Logical Cloud
type LogicalCloud struct {
	MetaData      types.Metadata `json:"metadata"`
	Specification Spec           `json:"spec"`
}

// Spec contains the parameters needed for spec
type Spec struct {
	NameSpace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
	Level     string            `json:"level"`
	User      UserData          `json:"user"`
}

// UserData contains the parameters needed for user
type UserData struct {
	UserName string `json:"userName"`
	Type     string `json:"type"`
}

// LogicalCloudKey is the key structure that is used in the database
type LogicalCloudKey struct {
	Project          string `json:"project"`
	LogicalCloudName string `json:"logicalCloud"`
}

// PrivateKey is the key structure that is used in the database
type PrivateKey struct {
	KeyValue string `json:"key" encrypted:""`
}
