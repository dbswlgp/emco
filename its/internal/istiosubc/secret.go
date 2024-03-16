// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package istiosubc

import (
	"gopkg.in/yaml.v2"
)

const K8S_KIND_SECRET = "Secret"
const K8S_APIVERSION =  "v1"

type K8sSecretResource struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Type       string   `yaml:"type"`
	Data       SEData   `yaml:"data"`
}
type SEData struct {
	CaCrt  string `yaml:"ca.crt"`
	TlsCrt string `yaml:"tls.crt"`
	TlsKey string `yaml:"tls.key"`
}

func createSecretResource(meta Metadata, ty string, data SEData) ([]byte, error) {

	var sero = K8sSecretResource {
		ApiVersion: K8S_APIVERSION,
		Kind: K8S_KIND_SECRET,
		Metadata: meta,
		Type: ty,
		Data: data,
	}
	y, err := yaml.Marshal(&sero)
	return y, err
}

func createSecretData(cac string, tlsc string, tlsk string) (SEData){
	var secdata = SEData {
		CaCrt:  cac,
		TlsCrt: tlsc,
		TlsKey: tlsk,
	}
	return secdata
}
