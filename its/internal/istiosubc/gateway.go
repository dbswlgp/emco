// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package istiosubc

import (
	"gopkg.in/yaml.v2"
)

//const NETWORKING_ISTIO_APIVERSION = "networking.istio.io/v1beta1"
const NETWORKING_ISTIO_KIND_GATEWAY =       "Gateway"

type IstioGatewayResource struct {
	ApiVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Metadata   Metadata    `yaml:"metadata"`
	Spec       GatewaySpec `yaml:"spec"`
}

type GatewaySpec struct {
	Servers    []Server          `yaml:"servers"`
	Selector   map[string]string `yaml:"selector"`
}

type Server struct {
	Port Port             `yaml:"port"`
	Bind string           `yaml:"bind,omitempty"`
	Hosts []string        `yaml:"hosts"`
	Tls ServerTLSSettings `yaml:"tls,omitempty"`
	Name string           `yaml:"name,omitempty"`
}

type ServerTLSSettings struct {
	Mode string `yaml:"mode,omitempty"`
	CredentialName string `yaml:"credentialName,omitempty"`
}

func createServerItem(port Port, bind string, hosts []string, tls ServerTLSSettings, name string) (Server){
	var ser = Server {
		Port:  port,
		Bind:  bind,
		Hosts: hosts,
		Tls:   tls,
		Name:  name,
	}
	return ser
}
func createGatewayResource(meta Metadata, spec GatewaySpec) ([]byte, error) {

	var ro = IstioGatewayResource {
		ApiVersion: NETWORKING_ISTIO_APIVERSION,
		Kind:       NETWORKING_ISTIO_KIND_GATEWAY,
		Metadata:   meta,
		Spec:       spec,
	}

	y, err := yaml.Marshal(&ro)
	return y, err

}

func createGatewaySpec(servers []Server, selector map[string]string) (GatewaySpec){
	var gspec = GatewaySpec {
		Servers:  servers,
		Selector: selector,
	}
	return gspec

}
