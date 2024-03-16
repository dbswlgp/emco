// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package istiosubc

import (
	"gopkg.in/yaml.v2"
)

const NETWORKING_ISTIO_KIND_DR =       "DestinationRule"

type IstioDestinationRuleResource struct {
	ApiVersion string              `yaml:"apiVersion"`
	Kind       string              `yaml:"kind"`
	Metadata   Metadata            `yaml:"metadata"`
	Spec       DestinationRuleSpec `yaml:"spec"`
}

type DestinationRuleSpec struct {
	Host           string        `yaml:"host"`
	TrafficPolicy  TrafficPolicy `yaml:"trafficPolicy"`
}

type TrafficPolicy struct {
	Tls ClientTLSSettings `yaml:"tls,omitempty"`
}

type ClientTLSSettings struct {
	Mode string `yaml:"mode", default:"ISTIO_MUTUAL"`
}

func createDestinationRuleResource(meta Metadata, spec DestinationRuleSpec) ([]byte, error) {

	var ro = IstioDestinationRuleResource {
		ApiVersion: NETWORKING_ISTIO_APIVERSION,
		Kind:       NETWORKING_ISTIO_KIND_DR,
		Metadata:   meta,
		Spec:       spec,
	}

	y, err := yaml.Marshal(&ro)
	return y, err

}

func createDestinationRuleSpec(host string, trafficPolicy TrafficPolicy) (DestinationRuleSpec,error){
	var drspec = DestinationRuleSpec {
		Host:          host,
		TrafficPolicy: trafficPolicy,
	}
	return drspec, nil

}
