// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package fluxv2

import (
	"context"
	"fmt"
	"strings"

	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"
	gitsupport "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/gitsupport"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
)

// Connection is for a cluster
type Fluxv2Provider struct {
	gitProvider   gitsupport.GitProvider
	timeOut       int
	syncInterval  int
	retryInterval int
}

func NewFluxv2Provider(ctx context.Context, cid, app, cluster, level, namespace string) (*Fluxv2Provider, error) {

	result := strings.SplitN(cluster, "+", 2)

	c, err := utils.GetGitOpsConfig(ctx, cluster, "0", "default")

	if err != nil {
		return nil, err
	}
	if c.Props.GitOpsType != "fluxcd" {
		log.Error("Invalid GitOps type:", log.Fields{})
		return nil, pkgerrors.Errorf("Invalid GitOps type: " + c.Props.GitOpsType)
	}

	// Read from database
	ccc := db.NewCloudConfigClient()

	gitProvider, err := gitsupport.NewGitProvider(ctx, cid, app, cluster, level, namespace)
	if err != nil {
		return nil, err
	}

	resObject, err := ccc.GetClusterSyncObjects(ctx, result[0], c.Props.GitOpsResourceObject)
	if err != nil {
		log.Error("Invalid resObject :", log.Fields{"resObj": c.Props.GitOpsResourceObject})
		return nil, pkgerrors.Errorf("Invalid resObject: " + c.Props.GitOpsResourceObject)
	}

	kvRes := resObject.Spec.Kv

	var timeOutStr, syncIntervalStr, retryIntervalStr string

	timeOutStr = "60"
	syncIntervalStr = "60"
	retryIntervalStr = "60"

	for _, kvpair := range kvRes {
		log.Info("kvpair", log.Fields{"kvpair": kvpair})
		v, ok := kvpair["timeOut"]
		if ok {
			timeOutStr = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["syncInterval"]
		if ok {
			syncIntervalStr = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["retryInterval"]
		if ok {
			retryIntervalStr = fmt.Sprintf("%v", v)
			continue
		}
	}

	var timeOut, syncInterval, retryInterval int

	_, err = fmt.Sscan(timeOutStr, &timeOut)

	if err != nil {
		log.Error("Invalid time out value", log.Fields{"timeOutStr": timeOutStr, "err": err})
		return nil, err
	}

	_, err = fmt.Sscan(syncIntervalStr, &syncInterval)

	if err != nil {
		log.Error("Invalid sync interval value", log.Fields{"syncIntervalStr": syncIntervalStr, "err": err})
		return nil, err
	}

	_, err = fmt.Sscan(retryIntervalStr, &retryInterval)

	if err != nil {
		log.Error("Invalid retry interval value", log.Fields{"retryIntervalStr": retryIntervalStr, "err": err})
		return nil, err
	}

	p := Fluxv2Provider{
		gitProvider:   *gitProvider,
		timeOut:       timeOut,
		syncInterval:  syncInterval,
		retryInterval: retryInterval,
	}
	return &p, nil
}

func (p *Fluxv2Provider) CleanClientProvider() error {
	return nil
}
