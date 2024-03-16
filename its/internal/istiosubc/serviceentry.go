// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package istiosubc

import (
	"gopkg.in/yaml.v2"
)

const NETWORKING_ISTIO_KIND_SE =       "ServiceEntry"

type IstioSericeEntryResource struct {
	ApiVersion string     `yaml:"apiVersion"`
	Kind       string     `yaml:"kind"`
	Metadata   Metadata   `yaml:"metadata"`
	Spec       SESpec     `yaml:"spec"`
}
type SESpec struct {
	Hosts      []string        `yaml:"hosts"`
	Addresses  []string        `yaml:"addresses,omitempty"`
	ExportTo   []string        `yaml:"exportTo,omitempty"`
	Location   string          `yaml:"location",omitempty"`
	Resolution string          `yaml:"resolution", default:DNS`
	Ports      []Port          `yaml:"ports"`
	Endpoints  []WorkloadEntry `yaml:"endpoints,omitempty"`
}
type WorkloadEntry struct {
	Address string            `yaml:"address"`
	Ports   map[string]uint32 `yaml:"ports,omitempty"`
}

func createServieEntryResource(meta Metadata, spec SESpec) ([]byte, error) {

	var sero = IstioSericeEntryResource {
		ApiVersion: NETWORKING_ISTIO_APIVERSION,
		Kind: NETWORKING_ISTIO_KIND_SE,
		Metadata: meta,
		Spec: spec,
	}

	y, err := yaml.Marshal(&sero)
	return y, err

}

func createServiceEntrySpec(hosts []string, addresses []string, exportTo []string, wle []WorkloadEntry, ports []Port, location string, resolution string) (SESpec){
	var vsspec = SESpec {
		Hosts: hosts,
		Addresses: addresses,
		ExportTo: exportTo,
		Location: location,
		Resolution: resolution,
		Ports: ports,
		Endpoints: wle,
	}
	return vsspec

}
