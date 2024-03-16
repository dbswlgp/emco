// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package istioservice

// ProxyConfig holds the proxyConfig data
type ProxyConfig struct {
	APIVersion string          `yaml:"apiVersion" json:"apiVersion"`
	Kind       string          `yaml:"kind" json:"kind"`
	MetaData   Metadata        `yaml:"metadata" json:"metadata"`
	Spec       ProxyConfigSpec `yaml:"spec" json:"spec"`
}

// ProxyConfigSpec holds the cert signer details
type ProxyConfigSpec struct {
	EnvironmentVariables map[string]string `yaml:"environmentVariables" json:"environmentVariables"`
}

// MetaData holds the data
type Metadata struct {
	Name        string `json:"name" yaml:"name"`
	Namespace   string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	UserData1   string `json:"userData1,omitempty" yaml:"userData1,omitempty"`
	UserData2   string `json:"userData2,omitempty" yaml:"userData2,omitempty"`
}
