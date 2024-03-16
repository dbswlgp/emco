// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package istiosubc


const NETWORKING_ISTIO_APIVERSION = "networking.istio.io/v1beta1"

type Metadata struct {
	Name        string `yaml:"name"`
	Namespace   string `yaml:"namespace,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type Port struct {
	Name        string `yaml:"name"`
	Number      uint32 `yaml:"number"`
	Protocol    string `yaml:"protocol"`
	TargetPort  string `yaml:"targetPort,omitempty"`
}

func createGenericMetadata(name string, namespace string, description string)(Metadata) {
	var meta = Metadata {
		Name:        name,
		Namespace:   namespace,
		Description: description,
	}
	return meta
}

func createGenericPort(name string, number uint32, protocol string, targetPort string)(Port) {
	var port = Port {
		Name:       name,
		Number:     number,
		Protocol:   protocol,
		TargetPort: targetPort,
	}
	return port
}
