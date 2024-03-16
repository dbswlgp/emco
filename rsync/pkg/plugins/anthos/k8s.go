// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package anthos

import (
	"context"

	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	gitsupport "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/gitsupport"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
)

// Connection is for a cluster
type AnthosProvider struct {
	gitProvider gitsupport.GitProvider
}

func NewAnthosProvider(ctx context.Context, cid, app, cluster, level, namespace string) (*AnthosProvider, error) {

	c, err := utils.GetGitOpsConfig(ctx, cluster, level, namespace)
	if err != nil {
		return nil, err
	}
	if c.Props.GitOpsType != "anthos" {
		log.Error("Invalid GitOps type:", log.Fields{})
		return nil, pkgerrors.Errorf("Invalid GitOps type: " + c.Props.GitOpsType)
	}

	gitProvider, err := gitsupport.NewGitProvider(ctx, cid, app, cluster, level, namespace)
	if err != nil {
		return nil, err
	}
	p := AnthosProvider{
		gitProvider: *gitProvider,
	}
	return &p, nil
}

func (p *AnthosProvider) CleanClientProvider() error {
	return nil
}
