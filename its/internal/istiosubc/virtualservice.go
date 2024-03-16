// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package istiosubc

import (
	"gopkg.in/yaml.v2"
)

const NETWORKING_ISTIO_KIND_VS =       "VirtualService"

type IstioVirtualSericeResource struct {
	ApiVersion string     `yaml:"apiVersion"`
	Kind       string     `yaml:"kind"`
	Metadata   Metadata   `yaml:"metadata"`
	Spec       VSSpec     `yaml:"spec"`
}

type VSSpec struct {
	Hosts      []string    `yaml:"hosts"`
	Gateways   []string    `yaml:"gateways,omitempty"`
	Http       []HTTPRoute `yaml:"http,omitempty"`
}

func createVirtualServieResource(meta Metadata, spec VSSpec) ([]byte, error) {

	var sero = IstioVirtualSericeResource {
		ApiVersion: NETWORKING_ISTIO_APIVERSION,
		Kind:       NETWORKING_ISTIO_KIND_VS,
		Metadata:   meta,
		Spec:       spec,
	}

	y, err := yaml.Marshal(&sero)
	return y, err

}
func createVirtualServiceSpec(hosts []string, gateways []string, route []HTTPRoute) (VSSpec){
	var vsspec = VSSpec {
		Hosts:    hosts,
		Gateways: gateways,
		Http:     route,
	}
	return vsspec

}
