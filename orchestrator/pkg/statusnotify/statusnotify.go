// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package statusnotify

import (
	"context"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	statusnotifypb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/statusnotify"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/statusnotifyserver"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
)

type digHelpers struct{}

func getDigKeyValues(reg *statusnotifypb.StatusRegistration) (string, string, string, string, error) {
	key := reg.GetKey()

	var digKey *statusnotifypb.DigKey

	switch key.(type) {
	case *statusnotifypb.StatusRegistration_DigKey:
		digKey = reg.GetDigKey()
	default:
		return "", "", "", "", pkgerrors.New("Status Notification Registration - Key is not a Deployment Intent Group key")
	}

	return digKey.GetProject(), digKey.GetCompositeApp(), digKey.GetCompositeAppVersion(), digKey.GetDeploymentIntentGroup(), nil
}

func (d digHelpers) GetAppContextId(ctx context.Context, reg *statusnotifypb.StatusRegistration) (string, error) {
	p, ca, v, di, err := getDigKeyValues(reg)
	if err != nil {
		return "", err
	}
	log.Trace("[StatusNotify] Deployment Intent Group Key",
		log.Fields{"project": p, "compositeApp": ca, "caVersion": v, "dig": di})

	si, err := module.NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(ctx, di, p, ca, v)
	if err != nil {
		log.Info("[StatusNotify] Deployment Intent Group Not Found",
			log.Fields{"Error": err})
		return "", pkgerrors.Wrap(err, "DeploymentIntentGroup state not found: "+di)
	}

	return state.GetStatusContextIdFromStateInfo(si), nil
}

func (d digHelpers) StatusQuery(ctx context.Context, reg *statusnotifypb.StatusRegistration, qStatusInstance, qType, qOutput string, qApps, qClusters, qResources []string) status.StatusResult {

	p, ca, v, di, err := getDigKeyValues(reg)
	if err != nil {
		return status.StatusResult{}
	}

	statusResult, err := module.NewInstantiationClient().GenericStatus(ctx, p, ca, v, di, qStatusInstance, qType, qOutput, qApps, qClusters, qResources)
	if err != nil {
		return status.StatusResult{}
	}
	return statusResult
}

// PrepareStatusNotification invokes the Status() function with the registered parameters and filters.
// The status result is used to populate and return a StatusNotification.
func (d digHelpers) PrepareStatusNotification(reg *statusnotifypb.StatusRegistration, statusResult status.StatusResult) *statusnotifypb.StatusNotification {
	n := new(statusnotifypb.StatusNotification)

	log.Trace("[StatusNotify] Preparing Notification",
		log.Fields{"statusResult": statusResult})

	if statusResult.DeployedStatus == appcontext.AppContextStatusEnum.Instantiated ||
		statusResult.DeployedStatus == appcontext.AppContextStatusEnum.Updated {
		switch reg.StatusType {
		case statusnotifypb.StatusValue_DEPLOYED:
			n.StatusValue = statusnotifypb.StatusValue_DEPLOYED
		case statusnotifypb.StatusValue_READY:
			if statusResult.ReadyStatus == "Ready" {
				n.StatusValue = statusnotifypb.StatusValue_READY
			} else {
				n.StatusValue = statusnotifypb.StatusValue_NOT_READY
			}
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
			appStatus := statusnotifypb.AppStatus{}
			appStatus.App = app.Name
			clusters := make([]*statusnotifypb.ClusterStatus, 0)
			appDeployed := true
			appReady := true
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
						appDeployed = false
					}
				case statusnotifypb.StatusValue_READY:
					if clusterReady {
						clusterStatus.StatusValue = statusnotifypb.StatusValue_READY
					} else {
						clusterStatus.StatusValue = statusnotifypb.StatusValue_NOT_READY
						appReady = false
					}
				}
				clusterStatus.Resources = resources
				clusters = append(clusters, &clusterStatus)
			}
			switch reg.StatusType {
			case statusnotifypb.StatusValue_DEPLOYED:
				if appDeployed {
					appStatus.StatusValue = statusnotifypb.StatusValue_DEPLOYED
				} else {
					appStatus.StatusValue = statusnotifypb.StatusValue_NOT_DEPLOYED
				}
			case statusnotifypb.StatusValue_READY:
				if appReady {
					appStatus.StatusValue = statusnotifypb.StatusValue_READY
				} else {
					appStatus.StatusValue = statusnotifypb.StatusValue_NOT_READY
				}
			}
			appStatus.Clusters = clusters
			details = append(details, &statusnotifypb.StatusDetail{StatusDetail: &statusnotifypb.StatusDetail_App{App: &appStatus}})
		}
		n.Details = details

	}

	return n
}

func StartStatusNotifyServer() *statusnotifyserver.StatusNotifyServer {
	return statusnotifyserver.NewStatusNotifyServer("digStatus", digHelpers{})
}
