// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package sdewancc

import (
	"strings"

	"gopkg.in/yaml.v2"
)

const NETWORKING_SDEWAN_KIND_APP = "SdewanApplication"

type SdewanApplicationResource struct {
	ApiVersion string                `yaml:"apiVersion"`
	Kind       string                `yaml:"kind"`
	Metadata   Metadata              `yaml:"metadata"`
	Spec       SdewanApplicationSpec `yaml:"spec"`
}

type SdewanApplicationSpec struct {
	AppNamespace string      `yaml:"appNamespace"`
	PodSelector  PodSelector `yaml:"podSelector"`
	ServicePort  string      `yaml:"servicePort"`
	CNFPort      string      `yaml:"cnfPort"`
}

type PodSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels"`
}

func createSdewanApplicationResource(meta Metadata, spec SdewanApplicationSpec) ([]byte, error) {
	var ro = SdewanApplicationResource{
		ApiVersion: NETWORKING_SDEWAN_APIVERSION,
		Kind:       NETWORKING_SDEWAN_KIND_APP,
		Metadata:   meta,
		Spec:       spec,
	}

	y, err := yaml.Marshal(&ro)
	return y, err
}

func createSdewanApplicationSpec(appNamespace string, podSelector PodSelector, servicePort, cnfPort string) SdewanApplicationSpec {
	var saspec = SdewanApplicationSpec{
		AppNamespace: appNamespace,
		PodSelector:  podSelector,
		ServicePort:  servicePort,
		CNFPort:      cnfPort,
	}
	return saspec
}

func createPodSelector(podLabels string) PodSelector {
	psmap := make(map[string]string)
	labelList := strings.Split(podLabels, "=")
	psmap[labelList[0]] = labelList[1]
	var ps = PodSelector{
		MatchLabels: psmap,
	}
	return ps
}
