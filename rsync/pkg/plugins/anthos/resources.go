// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package anthos

import (
	"context"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Creates a new resource if not already existing
func (p *AnthosProvider) Create(name string, ref interface{}, content []byte) (interface{}, error) {

	res, err := p.gitProvider.Create(name, ref, content)
	return res, err
}

// Apply resource to the cluster
func (p *AnthosProvider) Apply(ctx context.Context, name string, ref interface{}, content []byte) (interface{}, error) {

	//Decode the yaml to create a runtime.Object
	unstruct := &unstructured.Unstructured{}
	//Ignore the returned obj as we expect the data in unstruct
	_, err := utils.DecodeYAMLData(string(content), unstruct)
	if err != nil {
		return nil, err
	}

	// Namespaces shouldn't be blindly overwritten.
	// Cluster-scoped resources must not contain a namespace (incompatible with certain distros) and (maybe) resources with an existing namespace should be respected (User Permissions will decide if they can actually be deployed)
	structkind := unstruct.GetKind()
	if structkind != "ClusterRoleBinding" && structkind != "ClusterRole" && structkind != "RepoSync" && structkind != "Namespace" {
		// Set Namespace
		unstruct.SetNamespace(p.gitProvider.Namespace)
	}

	unstructJson, err := unstruct.MarshalJSON()
	if err != nil {
		return nil, err
	}
	path := p.GetPath("context") + name + ".yaml"
	res, err := p.gitProvider.Apply(path, ref, unstructJson)
	return res, err

}

// Delete resource from the cluster
func (p *AnthosProvider) Delete(name string, ref interface{}, content []byte) (interface{}, error) {

	path := p.GetPath("context") + name + ".yaml"
	res, err := p.gitProvider.Delete(path, ref, content)
	return res, err

}

// Get resource from the cluster
func (p *AnthosProvider) Get(ctx context.Context, name string, gvkRes []byte) ([]byte, error) {

	return []byte{}, nil
}

// Commit resources to the cluster
func (p *AnthosProvider) Commit(ctx context.Context, ref interface{}) error {

	err := p.gitProvider.Commit(ctx, ref)
	return err
}

// IsReachable cluster reachablity test
func (p *AnthosProvider) IsReachable() error {
	return nil
}

func (m *AnthosProvider) TagResource(res []byte, label string) ([]byte, error) {
	b, err := status.TagResource(res, label)
	if err != nil {
		log.Error("Error Tag Resource with label:", log.Fields{"err": err, "label": label, "resource": res})
		return nil, err
	}
	return b, nil
}
