// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package anthos

import (
	"context"
	"encoding/json"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// version of ResourceBundleState from monitor that allows emptying the Status field (Anthos compatibility)
type ResourceBundleStateV2 struct {
	APIVersion json.RawMessage `json:"apiVersion,inline"`
	Kind       json.RawMessage `json:"kind,inline"`
	Meta       json.RawMessage `json:"metadata,omitempty"`
	Spec       json.RawMessage `json:"spec,omitempty"`
	Status     json.RawMessage `json:"status,omitempty"`
}

// StartClusterWatcher watches for CR changes in git location
func (p *AnthosProvider) StartClusterWatcher(ctx context.Context) error {
	p.gitProvider.StartClusterWatcher(ctx)
	return nil
}

// ApplyStatusCR applies status CR
func (p *AnthosProvider) ApplyStatusCR(ctx context.Context, name string, content []byte) error {

	// Decode the yaml to create a runtime.Object
	unstruct := &unstructured.Unstructured{}
	_, err := utils.DecodeYAMLData(string(content), unstruct)
	if err != nil {
		return err
	}
	log.Debug("DecodeYAMLData", log.Fields{"unstruct": unstruct})
	// Set Namespace
	unstruct.SetNamespace(p.gitProvider.Namespace)

	rbsJson, err := json.Marshal(unstruct)
	if err != nil {
		return err
	}
	log.Debug("json.Marshal(unstruct)", log.Fields{"rbsJson": rbsJson})

	// Anthos doesn't allow resources with a pre-defined Status field to be added.
	// Currently EMCO sets a Status field for ResourceBundleState, so patch this.

	rbs := &ResourceBundleStateV2{}
	// But first convert from Unstructured to our own ResourceBundleStateV2
	err = json.Unmarshal(rbsJson, rbs)
	if err != nil {
		return err
	}
	log.Debug("json.Unmarshal", log.Fields{"rbs": rbs})

	// Remove Status field
	rbs.Status = nil

	rbsJson, err = json.Marshal(rbs)
	if err != nil {
		return err
	}

	path := p.GetPath("context") + name + ".yaml"
	ref, err := p.gitProvider.Apply(path, nil, rbsJson)
	if err != nil {
		return err
	}
	p.gitProvider.Commit(ctx, ref)
	return err
}

// DeleteStatusCR deletes status CR
func (p *AnthosProvider) DeleteStatusCR(ctx context.Context, name string, content []byte) error {
	path := p.GetPath("context") + name + ".yaml"
	ref, err := p.gitProvider.Delete(path, nil, content)
	if err != nil {
		return err
	}
	p.gitProvider.Commit(ctx, ref)
	return err
}
