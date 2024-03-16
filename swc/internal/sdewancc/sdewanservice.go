// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package sdewancc

import (
	"gopkg.in/yaml.v2"
)

const NETWORKING_SDEWAN_KIND_SERVICE = "CNFService"

type SdewanServiceResource struct {
	ApiVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   Metadata          `yaml:"metadata"`
	Spec       SdewanServiceSpec `yaml:"spec"`
}

type SdewanServiceSpec struct {
	FullName string `yaml:"fullname"`
	Port     string `yaml:"port"`
	DPort    string `yaml:"dport"`
	CIDR     string `yaml:"cidr"`
}

func createSdewanServiceResource(meta Metadata, spec SdewanServiceSpec) ([]byte, error) {
	var ro = SdewanServiceResource{
		ApiVersion: NETWORKING_SDEWAN_APIVERSION,
		Kind:       NETWORKING_SDEWAN_KIND_SERVICE,
		Metadata:   meta,
		Spec:       spec,
	}

	y, err := yaml.Marshal(&ro)
	return y, err
}

func createSdewanServiceSpec(fullname string, port string, dport string, cidr string) SdewanServiceSpec {
	var ssspec = SdewanServiceSpec{
		FullName: fullname,
		Port:     port,
		DPort:    dport,
		CIDR:     cidr,
	}
	return ssspec
}
