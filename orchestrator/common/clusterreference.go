// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2022 Intel Corporation

package common

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

// Cluster contains the parameters needed for a Cluster
type Cluster struct {
	MetaData      types.Metadata `json:"metadata"`
	Specification ClusterSpec    `json:"spec"`
}

type ClusterSpec struct {
	ClusterProvider string `json:"clusterProvider"`
	ClusterName     string `json:"cluster"`
	LoadBalancerIP  string `json:"loadBalancerIP"`
	Certificate     string `json:"certificate"`
}

type KubeConfig struct {
	ApiVersion     string            `yaml:"apiVersion"`
	Kind           string            `yaml:"kind"`
	Clusters       []KubeCluster     `yaml:"clusters"`
	Contexts       []KubeContext     `yaml:"contexts"`
	CurrentContext string            `yaml:"current-context"`
	Preferences    map[string]string `yaml:"preferences"`
	Users          []KubeUser        `yaml:"users"`
}

type KubeCluster struct {
	ClusterDef  KubeClusterDef `yaml:"cluster"`
	ClusterName string         `yaml:"name"`
}

type KubeClusterDef struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}

type KubeContext struct {
	ContextDef  KubeContextDef `yaml:"context"`
	ContextName string         `yaml:"name"`
}

type KubeContextDef struct {
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace,omitempty"`
	User      string `yaml:"user"`
}

type KubeUser struct {
	UserName string      `yaml:"name"`
	UserDef  KubeUserDef `yaml:"user"`
}

type KubeUserDef struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
	// client-certificate and client-key are NOT implemented
}

type ClusterKey struct {
	Project          string `json:"project"`
	LogicalCloudName string `json:"logicalCloud"`
	ClusterReference string `json:"clusterReference"`
}
