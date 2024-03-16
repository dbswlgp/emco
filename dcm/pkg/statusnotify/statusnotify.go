// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package statusnotify

import (
	"context"

	pkgerrors "github.com/pkg/errors"
	dcm "gitlab.com/project-emco/core/emco-base/src/dcm/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	statusnotifypb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/statusnotify"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/statusnotifyserver"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
)

type lcHelpers struct{}

func getLcKeyValues(reg *statusnotifypb.StatusRegistration) (string, string, error) {
	key := reg.GetKey()

	var lcKey *statusnotifypb.LcKey

	switch key.(type) {
	case *statusnotifypb.StatusRegistration_LcKey:
		lcKey = reg.GetLcKey()
	default:
		return "", "", pkgerrors.New("Status Notification Registration - Key is not a Deployment Intent Group key")
	}

	return lcKey.GetProject(), lcKey.GetLogicalCloud(), nil
}

func (d lcHelpers) GetAppContextId(ctx context.Context, reg *statusnotifypb.StatusRegistration) (string, error) {
	p, lc, err := getLcKeyValues(reg)
	if err != nil {
		return "", err
	}

	si, err := dcm.NewLogicalCloudClient().GetState(ctx, p, lc)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Logical cloud state not found: "+lc)
	}

	return state.GetStatusContextIdFromStateInfo(si), nil
}

func (d lcHelpers) StatusQuery(ctx context.Context, reg *statusnotifypb.StatusRegistration, qStatusInstance, qType, qOutput string, fApps, fClusters, fResources []string) status.StatusResult {
	p, lc, err := getLcKeyValues(reg)
	if err != nil {
		return status.StatusResult{}
	}

	statusResult, err := dcm.NewLogicalCloudClient().GenericStatus(ctx, p, lc, qStatusInstance, qType, qOutput, fClusters, fResources)
	if err != nil {
		return status.StatusResult{}
	}
	return statusResult
}

func (d lcHelpers) PrepareStatusNotification(reg *statusnotifypb.StatusRegistration, statusResult status.StatusResult) *statusnotifypb.StatusNotification {
	n := new(statusnotifypb.StatusNotification)
	log.Trace("[StatusNotify] Preparing Notification",
		log.Fields{"statusResult": statusResult})

	if statusResult.DeployedStatus == appcontext.AppContextStatusEnum.Instantiated ||
		statusResult.DeployedStatus == appcontext.AppContextStatusEnum.Updated {
		switch reg.StatusType {
		case statusnotifypb.StatusValue_DEPLOYED:
			n.StatusValue = statusnotifypb.StatusValue_DEPLOYED
		case statusnotifypb.StatusValue_READY:
			n.StatusValue = statusnotifypb.StatusValue_READY
		}
	} else {
		switch reg.StatusType {
		case statusnotifypb.StatusValue_DEPLOYED:
			n.StatusValue = statusnotifypb.StatusValue_NOT_DEPLOYED
		case statusnotifypb.StatusValue_READY:
			n.StatusValue = statusnotifypb.StatusValue_NOT_READY
		}
	}

	if reg.Output == statusnotifypb.OutputType_ALL {
		details := make([]*statusnotifypb.StatusDetail, 0)
		for _, app := range statusResult.Apps {
			clusters := make([]*statusnotifypb.ClusterStatus, 0)
			for _, cluster := range app.Clusters {
				clusterStatus := statusnotifypb.ClusterStatus{}
				clusterStatus.Cluster = cluster.Cluster
				clusterStatus.ClusterProvider = cluster.ClusterProvider
				resources := make([]*statusnotifypb.ResourceStatus, 0)
				clusterDeployed := true
				clusterReady := true
				for _, resource := range cluster.Resources {
					resourceStatus := statusnotifypb.ResourceStatus{}
					resourceStatus.Name = resource.Name
					resourceStatus.Gvk = &statusnotifypb.GVK{Group: resource.Gvk.Group, Version: resource.Gvk.Version, Kind: resource.Gvk.Kind}
					switch reg.StatusType {
					case statusnotifypb.StatusValue_DEPLOYED:
						if resource.DeployedStatus == "Applied" {
							resourceStatus.StatusValue = statusnotifypb.StatusValue_DEPLOYED
						} else {
							resourceStatus.StatusValue = statusnotifypb.StatusValue_NOT_DEPLOYED
							clusterDeployed = false
						}
					case statusnotifypb.StatusValue_READY:
						if resource.ReadyStatus == "Ready" {
							resourceStatus.StatusValue = statusnotifypb.StatusValue_READY
						} else {
							resourceStatus.StatusValue = statusnotifypb.StatusValue_NOT_READY
							clusterReady = false
						}
					}
					resources = append(resources, &resourceStatus)
				}
				switch reg.StatusType {
				case statusnotifypb.StatusValue_DEPLOYED:
					if clusterDeployed {
						clusterStatus.StatusValue = statusnotifypb.StatusValue_DEPLOYED
					} else {
						clusterStatus.StatusValue = statusnotifypb.StatusValue_NOT_DEPLOYED
					}
				case statusnotifypb.StatusValue_READY:
					if clusterReady {
						clusterStatus.StatusValue = statusnotifypb.StatusValue_READY
					} else {
						clusterStatus.StatusValue = statusnotifypb.StatusValue_NOT_READY
					}
				}
				clusterStatus.Resources = resources
				clusters = append(clusters, &clusterStatus)
				details = append(details, &statusnotifypb.StatusDetail{StatusDetail: &statusnotifypb.StatusDetail_Cluster{Cluster: &clusterStatus}})
			}
		}
		n.Details = details

	}

	return n
}

func StartStatusNotifyServer() *statusnotifyserver.StatusNotifyServer {
	return statusnotifyserver.NewStatusNotifyServer("digStatus", lcHelpers{})
}
