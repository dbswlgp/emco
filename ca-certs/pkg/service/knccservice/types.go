// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package knccservice

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Config holds the kncc config data
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ConfigSpec `json:"spec,omitempty"`
}

// ConfigSpec holds the kncc resource and patch data
type ConfigSpec struct {
	Resource Resource `json:"resource,omitempty"`
	Patch    []Patch  `json:"patch,omitempty"`
}

// Resource holds the resource data
type Resource struct {
	Name      string `json:"name,omitempty"`
	NameSpace string `json:"namespace,omitempty"`
}

// Patch holds the patch data
type Patch struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
