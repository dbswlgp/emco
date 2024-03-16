// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package sdewancc

const NETWORKING_SDEWAN_APIVERSION = "batch.sdewan.akraino.org/v1alpha1"

type Metadata struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Description string            `yaml:"description,omitempty"`
}

func createGenericMetadata(name string, namespace string, description string)(Metadata) {
	labelmap := make(map[string]string)
	labelmap["sdewanPurpose"] = "base"
	var meta = Metadata {
		Name:        name,
		Namespace:   namespace,
		Labels:      labelmap,
		Description: description,
	}
	return meta
}
