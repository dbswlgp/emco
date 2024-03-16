// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package azurearcv2

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

type AzureArcV2Provider struct {
	gitProvider      gitsupport.GitProvider
	clientID         string
	tenantID         string
	clientSecret     string
	subscriptionID   string
	arcCluster       string
	arcResourceGroup string
	timeOut          int
	syncInterval     int
	retryInterval    int
}

func NewAzureArcProvider(ctx context.Context, cid, app, cluster, level, namespace string) (*AzureArcV2Provider, error) {

	result := strings.SplitN(cluster, "+", 2)

	c, err := utils.GetGitOpsConfig(ctx, cluster, "0", "default")

	if err != nil {
		return nil, err
	}
	if c.Props.GitOpsType != "azureArcV2" {
		log.Error("Invalid GitOps type:", log.Fields{})
		return nil, pkgerrors.Errorf("Invalid GitOps type: " + c.Props.GitOpsType)
	}

	// Read from database
	ccc := db.NewCloudConfigClient()

	gitProvider, err := gitsupport.NewGitProvider(ctx, cid, app, cluster, level, namespace)
	if err != nil {
		log.Error("Error creating git provider", log.Fields{"err": err, "gitProvider": gitProvider})
		return nil, err
	}

	resObject, err := ccc.GetClusterSyncObjects(ctx, result[0], c.Props.GitOpsResourceObject)
	if err != nil {
		log.Error("Invalid resObject :", log.Fields{"resObj": c.Props.GitOpsResourceObject})
		return nil, pkgerrors.Errorf("Invalid resObject: " + c.Props.GitOpsResourceObject)
	}

	kvRes := resObject.Spec.Kv

	var clientID, tenantID, clientSecret, subscriptionID, arcCluster, arcResourceGroup, timeOutStr, syncIntervalStr, retryIntervalStr string

	timeOutStr = "60"
	syncIntervalStr = "60"
	retryIntervalStr = "60"

	for _, kvpair := range kvRes {
		log.Info("kvpair", log.Fields{"kvpair": kvpair})
		v, ok := kvpair["clientID"]
		if ok {
			clientID = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["tenantID"]
		if ok {
			tenantID = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["clientSecret"]
		if ok {
			clientSecret = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["subscriptionID"]
		if ok {
			subscriptionID = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["arcCluster"]
		if ok {
			arcCluster = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["arcResourceGroup"]
		if ok {
			arcResourceGroup = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["timeOut"]
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
	if len(clientID) <= 0 || len(tenantID) <= 0 || len(clientSecret) <= 0 || len(subscriptionID) <= 0 || len(arcCluster) <= 0 || len(arcResourceGroup) <= 0 || len(timeOutStr) <= 0 || len(syncIntervalStr) <= 0 || len(retryIntervalStr) <= 0 {
		log.Error("Missing information for Azure Arc", log.Fields{"clientID": clientID, "tenantID": tenantID, "clientSecret": clientSecret, "subscriptionID": subscriptionID,
			"arcCluster": arcCluster, "arcResourceGroup": arcResourceGroup, "timeOut": timeOutStr, "syncInterval": syncIntervalStr, "retryInterval": retryIntervalStr})
		return nil, pkgerrors.Errorf("Missing Information for Azure Arc V2")
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

	p := AzureArcV2Provider{

		gitProvider:      *gitProvider,
		clientID:         clientID,
		tenantID:         tenantID,
		clientSecret:     clientSecret,
		subscriptionID:   subscriptionID,
		arcCluster:       arcCluster,
		arcResourceGroup: arcResourceGroup,
		timeOut:          timeOut,
		syncInterval:     syncInterval,
		retryInterval:    retryInterval,
	}
	return &p, nil
}

func (p *AzureArcV2Provider) CleanClientProvider() error {
	return nil
}
