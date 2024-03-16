// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package fluxv2

import (
	"context"
	"time"

	kustomize "github.com/fluxcd/kustomize-controller/api/v1beta2"
	fluxsc "github.com/fluxcd/source-controller/api/v1beta1"
	yaml "github.com/ghodss/yaml"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create GitRepository and Kustomization CR's for Flux
func (p *Fluxv2Provider) ApplyConfig(ctx context.Context, config interface{}) error {

	var sa string
	var files interface{}

	acUtils, err := utils.NewAppContextReference(ctx, p.gitProvider.Cid)
	if err != nil {
		return nil
	}
	_, level := acUtils.GetNamespace(ctx)
	_, _, _, lcn, _, err := acUtils.GetLogicalCloudInfo(ctx)
	if err != nil {
		return err
	}
	// INFO: the below is always false since level is always 0, as it stands
	if level == "1" {
		sa = lcn + "-sa"
	}
	var namespace, kName string
	var skip bool

	// Special case creating a logical cloud
	if level == "0" && lcn != "" {
		namespace = "flux-system"
		kName = "flux-system"
		skip = true
	} else {
		namespace = p.gitProvider.Namespace
		skip = false
	}
	if !skip {
		// Create Source CR and KustomizeCR
		gr := fluxsc.GitRepository{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "source.toolkit.fluxcd.io/v1beta1",
				Kind:       "GitRepository",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      p.gitProvider.Cid,
				Namespace: p.gitProvider.Namespace,
			},
			Spec: fluxsc.GitRepositorySpec{
				URL:       p.gitProvider.Url,
				Interval:  metav1.Duration{Duration: time.Second * 30},
				Reference: &fluxsc.GitRepositoryRef{Branch: p.gitProvider.Branch},
			},
		}
		x, err := yaml.Marshal(&gr)
		if err != nil {
			log.Error("ApplyConfig:: Marshal err", log.Fields{"err": err, "gr": gr})
			return err
		}
		path := "clusters/" + p.gitProvider.Cluster + "/" + gr.Name + ".yaml"
		files, err = p.gitProvider.Apply(path, files, x)

		kName = gr.Name
	}
	kc := kustomize.Kustomization{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kustomize.toolkit.fluxcd.io/v1beta2",
			Kind:       "Kustomization",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kust" + p.gitProvider.Cid,
			Namespace: namespace,
		},
		Spec: kustomize.KustomizationSpec{
			Interval:      metav1.Duration{Duration: time.Second * time.Duration(p.syncInterval)},
			RetryInterval: &metav1.Duration{Duration: time.Second * time.Duration(p.retryInterval)},
			Timeout:       &metav1.Duration{Duration: time.Second * time.Duration(p.timeOut)},
			Path:          "clusters/" + p.gitProvider.Cluster + "/context/" + p.gitProvider.Cid,
			Prune:         true,
			SourceRef: kustomize.CrossNamespaceSourceReference{
				Kind: "GitRepository",
				Name: kName,
			},
			ServiceAccountName: sa,
		},
	}
	y, err := yaml.Marshal(&kc)
	if err != nil {
		log.Error("ApplyConfig:: Marshal err", log.Fields{"err": err, "kc": kc})
		return err
	}
	path := "clusters/" + p.gitProvider.Cluster + "/" + kc.Name + ".yaml"

	p.gitProvider.GetPath("s")
	files, err = p.gitProvider.Apply(path, files, y)

	// Commit
	err = p.gitProvider.Commit(context.Background(), files)

	if err != nil {
		log.Error("ApplyConfig:: Commit files err", log.Fields{"err": err, "files": files})
	}
	return err
	//}

	return nil
}

// Delete GitRepository and Kustomization CR's for Flux
func (p *Fluxv2Provider) DeleteConfig(ctx context.Context, config interface{}) error {

	var files interface{}

	path := "clusters/" + p.gitProvider.Cluster + "/" + p.gitProvider.Cid + ".yaml"
	files, err := p.gitProvider.Delete(path, files, nil)

	path = "clusters/" + p.gitProvider.Cluster + "/" + "kust" + p.gitProvider.Cid + ".yaml"
	files, err = p.gitProvider.Delete(path, files, nil)

	err = p.gitProvider.Commit(context.Background(), files)
	if err != nil {
		log.Error("DeleteConfig:: Commit files err", log.Fields{"err": err, "files": files})
	}
	return err
}
