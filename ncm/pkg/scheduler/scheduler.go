// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package scheduler

import (
	"encoding/json"
	"fmt"
	"time"

	clusterPkg "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	oc "gitlab.com/project-emco/core/emco-base/src/ncm/internal/ovncontroller"
	ncmtypes "gitlab.com/project-emco/core/emco-base/src/ncm/pkg/module/types"
	nettypes "gitlab.com/project-emco/core/emco-base/src/ncm/pkg/networkintents/types"
	appcontext "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/installappclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"

	"context"

	pkgerrors "github.com/pkg/errors"
)

// rsyncName denotes the name of the rsync controller
const rsyncName = "rsync"

// ClusterManager is an interface exposes the Cluster functionality
type SchedulerManager interface {
	ApplyNetworkIntents(ctx context.Context, clusterProvider, cluster string) error
	NetworkIntentsStatus(ctx context.Context, clusterProvider, cluster, qInstance, qType, qOutput string, qApps, qClusters, qResources []string) (ClusterStatus, error)
	GenericNetworkIntentsStatus(ctx context.Context, clusterProvider, cluster, qInstance, qType, qOutput string, qApps, qClusters, qResources []string) (status.StatusResult, error)
	TerminateNetworkIntents(ctx context.Context, clusterProvider, cluster string) error
	StopNetworkIntents(ctx context.Context, clusterProvider, cluster string) error
}

// ClusterClient implements the Manager
// It will also be used to maintain some localized state
type SchedulerClient struct {
	db ncmtypes.ClientDbInfo
}

// NewSchedulerClient returns an instance of the SchedulerClient
// which implements the Manager
func NewSchedulerClient() *SchedulerClient {
	return &SchedulerClient{
		db: ncmtypes.ClientDbInfo{
			StoreName:  "resources",
			TagMeta:    "data",
			TagContent: "clustercontent",
			TagState:   "stateInfo",
		},
	}
}

// ClusterStatus holds the status data prepared for cluster network intent status queries
type ClusterStatus struct {
	status.ClusterStatusResult `json:",inline"`
}

func deleteAppContext(ctx context.Context, ac appcontext.AppContext) {
	err := ac.DeleteCompositeApp(ctx)
	if err != nil {
		log.Warn(":: Error deleting AppContext ::", log.Fields{"Error": err})
	}
}

/*
queryDBAndSetRsyncInfo queries the MCO db to find the record the sync controller
and then sets the RsyncInfo global variable.
*/
func queryDBAndSetRsyncInfo(ctx context.Context) (installappclient.RsyncInfo, error) {
	client := controller.NewControllerClient("resources", "data", "orchestrator")
	vals, _ := client.GetControllers(ctx)
	for _, v := range vals {
		if v.Metadata.Name == rsyncName {
			log.Info("Initializing RPC connection to resource synchronizer", log.Fields{
				"Controller": v.Metadata.Name,
			})
			rsyncInfo := installappclient.NewRsyncInfo(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			return rsyncInfo, nil
		}
	}
	return installappclient.RsyncInfo{}, pkgerrors.Errorf("queryRsyncInfoInMCODB Failed - Could not get find rsync by name : %v", rsyncName)
}

/*
callRsyncInstall method shall take in the app context id and invokes the rsync service via grpc
*/
func callRsyncInstall(ctx context.Context, contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo(ctx)
	log.Info("Calling the Rsync ", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = installappclient.InvokeInstallApp(ctx, appContextID)
	if err != nil {
		return err
	}
	return nil
}

/*
callRsyncUninstall method shall take in the app context id and invokes the rsync service via grpc
*/
func callRsyncUninstall(ctx context.Context, contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo(ctx)
	log.Info("Calling the Rsync ", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = installappclient.InvokeUninstallApp(ctx, appContextID)
	if err != nil {
		return err
	}
	return nil
}

// Apply Network Intents associated with a cluster
func (v *SchedulerClient) ApplyNetworkIntents(ctx context.Context, clusterProvider, cluster string) error {

	s, err := clusterPkg.NewClusterClient().GetClusterState(ctx, clusterProvider, cluster)
	if err != nil {
		return err
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting current state from Cluster stateInfo: "+cluster)
	}
	switch stateVal {
	case state.StateEnum.Approved:
		return pkgerrors.New("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Approved)
	case state.StateEnum.Terminated:
		break
	case state.StateEnum.TerminateStopped:
		break
	case state.StateEnum.Created:
		break
	case state.StateEnum.Applied:
		return nil
	case state.StateEnum.InstantiateStopped: // applies to Applied as well
		return nil
	case state.StateEnum.Instantiated:
		return pkgerrors.New("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Instantiated)
	default:
		return pkgerrors.New("Cluster is in an invalid state: " + cluster + " " + stateVal)
	}

	// Make an app context for the network intent resources
	ac := appcontext.AppContext{}
	ctxVal, err := ac.InitAppContext()
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating AppContext")
	}
	handle, err := ac.CreateCompositeApp(ctx)
	if err != nil {
		deleteAppContext(ctx, ac)
		return pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}

	// Add an app (fixed value) to the app context
	apphandle, err := ac.AddApp(ctx, handle, nettypes.CONTEXT_CLUSTER_APP)
	if err != nil {
		deleteAppContext(ctx, ac)
		return pkgerrors.Wrap(err, "Error adding App to AppContext")
	}

	// Add an app order instruction
	appinstr := struct {
		Apporder []string `json:"apporder"`
	}{
		[]string{nettypes.CONTEXT_CLUSTER_APP},
	}
	jinstr, err := json.Marshal(appinstr)
	if err != nil {
		deleteAppContext(ctx, ac)
		return pkgerrors.Wrap(err, "Error marshalling network intent app order instruction")
	}

	appdepinstr := struct {
		Appdep map[string]string `json:"appdependency"`
	}{
		map[string]string{nettypes.CONTEXT_CLUSTER_APP: "go"},
	}
	jdep, err := json.Marshal(appdepinstr)
	if err != nil {
		deleteAppContext(ctx, ac)
		return pkgerrors.Wrap(err, "Error marshalling network intent app dependency instruction")
	}

	_, err = ac.AddInstruction(ctx, handle, "app", "order", string(jinstr))
	if err != nil {
		deleteAppContext(ctx, ac)
		return pkgerrors.Wrap(err, "Error adding network intent app order instruction")
	}
	_, err = ac.AddInstruction(ctx, handle, "app", "dependency", string(jdep))
	if err != nil {
		deleteAppContext(ctx, ac)
		return pkgerrors.Wrap(err, "Error adding network intent app dependency instruction")
	}

	// Add a cluster to the app
	_, err = ac.AddCluster(ctx, apphandle, clusterProvider+nettypes.SEPARATOR+cluster)
	if err != nil {
		deleteAppContext(ctx, ac)
		return pkgerrors.Wrap(err, "Error adding Cluster to AppContext")
	}

	// Pass the appCtx to the appropriate controller (just default ovncontroller now)
	// for internal controller - pass the appcontext, cluster provider and cluster names in directly
	// external controllers will be given the appcontext id and wiil have to recontstruct
	// their own context
	err = oc.Apply(ctx, ctxVal, clusterProvider, cluster)
	if err != nil {
		deleteAppContext(ctx, ac)
		return pkgerrors.Wrap(err, "Error adding Cluster to AppContext")
	}

	// call resource synchronizer to instantiate the CRs in the cluster
	err = callRsyncInstall(ctx, ctxVal)
	if err != nil {
		deleteAppContext(ctx, ac)
		return err
	}

	// update the StateInfo in the cluster db record
	key := clusterPkg.ClusterKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
	}
	a := state.ActionEntry{
		State:     state.StateEnum.Applied,
		ContextId: ctxVal.(string),
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)
	s.StatusContextId = ctxVal.(string)

	err = db.DBconn.Insert(ctx, v.db.StoreName, key, nil, v.db.TagState, s)
	if err != nil {
		log.Warn(":: Error updating Cluster state in DB ::", log.Fields{"Error": err.Error(), "cluster": cluster, "cluster provider": clusterProvider, "AppContext": ctxVal.(string)})
		return pkgerrors.Wrap(err, "Error updating the stateInfo of cluster after Apply on network intents: "+cluster)
	}

	return nil
}

// Terminate Network Intents associated with a cluster
func (v *SchedulerClient) TerminateNetworkIntents(ctx context.Context, clusterProvider, cluster string) error {
	s, err := clusterPkg.NewClusterClient().GetClusterState(ctx, clusterProvider, cluster)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error finding StateInfo for cluster: %v, %v", clusterProvider, cluster)
	}
	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting current state from Cluster stateInfo: "+cluster)
	}
	switch stateVal {
	case state.StateEnum.Approved:
		return pkgerrors.New("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Approved)
	case state.StateEnum.Terminated:
		return nil
	case state.StateEnum.TerminateStopped:
		return nil
	case state.StateEnum.InstantiateStopped:
		break
	case state.StateEnum.Created:
		return pkgerrors.New("Cluster network intents have not been applied: " + cluster)
	case state.StateEnum.Applied:
		break
	case state.StateEnum.Instantiated:
		return pkgerrors.New("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Instantiated)
	default:
		return pkgerrors.New("Cluster is in an invalid state: " + cluster + " " + stateVal)
	}

	// call resource synchronizer to terminate the CRs in the cluster
	contextId := state.GetLastContextIdFromStateInfo(s)
	if stateVal == state.StateEnum.InstantiateStopped {
		err = state.UpdateAppContextStopFlag(ctx, contextId, false)
		if err != nil {
			return err
		}
	}
	err = callRsyncUninstall(ctx, contextId)
	if err != nil {
		return err
	}

	// update StateInfo
	key := clusterPkg.ClusterKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
	}
	a := state.ActionEntry{
		State:     state.StateEnum.Terminated,
		ContextId: contextId,
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)
	err = db.DBconn.Insert(ctx, v.db.StoreName, key, nil, v.db.TagState, s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of cluster: "+cluster)
	}

	return nil
}

// Stop Network Intent instantiation or termination associated with a cluster
func (v *SchedulerClient) StopNetworkIntents(ctx context.Context, clusterProvider, cluster string) error {
	s, err := clusterPkg.NewClusterClient().GetClusterState(ctx, clusterProvider, cluster)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error finding StateInfo for cluster: %v, %v", clusterProvider, cluster)
	}
	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting current state from Cluster stateInfo: "+cluster)
	}
	stopState := state.StateEnum.Undefined
	switch stateVal {
	case state.StateEnum.Approved:
		return pkgerrors.New("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Approved)
	case state.StateEnum.Terminated:
		stopState = state.StateEnum.TerminateStopped
		break
	case state.StateEnum.Applied:
		stopState = state.StateEnum.InstantiateStopped
		break
	case state.StateEnum.TerminateStopped:
		return pkgerrors.New("Cluster network intents termination already stopped: " + cluster)
	case state.StateEnum.InstantiateStopped:
		return pkgerrors.New("Cluster network intents instantiation already stopped: " + cluster)
	case state.StateEnum.Created:
		return pkgerrors.New("Cluster network intents have not been applied: " + cluster)
	case state.StateEnum.Instantiated:
		return pkgerrors.New("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Instantiated)
	default:
		return pkgerrors.New("Cluster is in an invalid state: " + cluster + " " + stateVal)
	}

	// call resource synchronizer to terminate the CRs in the cluster
	contextId := state.GetLastContextIdFromStateInfo(s)
	acStatus, err := state.GetAppContextStatus(ctx, contextId)
	if err != nil {
		return err
	}
	if acStatus.Status != appcontext.AppContextStatusEnum.Instantiating &&
		acStatus.Status != appcontext.AppContextStatusEnum.Terminating {
		return pkgerrors.Errorf("Cluster network intents are not instantiating or terminating:" + cluster)
	}
	err = state.UpdateAppContextStopFlag(ctx, contextId, true)
	if err != nil {
		return err
	}

	// update StateInfo
	key := clusterPkg.ClusterKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
	}
	a := state.ActionEntry{
		State:     stopState,
		ContextId: contextId,
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)
	err = db.DBconn.Insert(ctx, v.db.StoreName, key, nil, v.db.TagState, s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of cluster: "+cluster)
	}

	return nil
}

/*
NetworkIntentsStatus takes in cluster provider, cluster and query parameters.
This method is responsible obtaining the status of
the cluster network intents, which is made available in the appcontext
*/
func (c SchedulerClient) NetworkIntentsStatus(ctx context.Context, clusterProvider, cluster, qInstance, qType, qOutput string, qApps, qClusters, qResources []string) (ClusterStatus, error) {

	s, err := clusterPkg.NewClusterClient().GetClusterState(ctx, clusterProvider, cluster)
	if err != nil {
		return ClusterStatus{}, pkgerrors.Wrap(err, "Cluster state not found")
	}

	statusResponse, err := status.PrepareClusterStatusResult(ctx, s, qInstance, qType, qOutput, qApps, qClusters, qResources)
	if err != nil {
		return ClusterStatus{}, err
	}
	statusResponse.Name = clusterProvider + "+" + cluster
	clStatus := ClusterStatus{
		ClusterStatusResult: statusResponse,
	}

	return clStatus, nil
}

/*
GenericNetworkIntentsStatus takes in cluster provider, cluster and query parameters.
This method is responsible obtaining the status of
the cluster network intents, which is made available in the appcontext
It returns the full StatusResult structure so it can be used with the status notification framework.
*/
func (c SchedulerClient) GenericNetworkIntentsStatus(ctx context.Context, clusterProvider, cluster, qStatusInstance, qType, qOutput string, qApps, qClusters, qResources []string) (status.StatusResult, error) {

	s, err := clusterPkg.NewClusterClient().GetClusterState(ctx, clusterProvider, cluster)
	if err != nil {
		return status.StatusResult{}, pkgerrors.Wrap(err, "Cluster state not found")
	}

	qInstance, err := state.GetContextIdForStatusContextId(s, qStatusInstance)
	if err != nil {
		return status.StatusResult{}, err
	}

	statusResponse, err := status.GenericPrepareStatusResult(ctx, status.ClusterStatusQuery, s, qInstance, qType, qOutput, qApps, qClusters, qResources)
	if err != nil {
		return status.StatusResult{}, err
	}
	statusResponse.Name = clusterProvider + "+" + cluster

	return statusResponse, nil
}
