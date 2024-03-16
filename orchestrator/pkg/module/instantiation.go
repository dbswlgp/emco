// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2022 Intel Corporation

package module

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	gpic "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/gpic"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/utils/helm"
	"go.opentelemetry.io/otel/trace"
)

// ManifestFileName is the name given to the manifest file in the profile package
const ManifestFileName = "manifest.yaml"

// GenericPlacementIntentName denotes the generic placement intent name
const GenericPlacementIntentName = "genericPlacementIntent"

// SEPARATOR used while creating clusternames to store in etcd
const SEPARATOR = "+"

// InstantiationClient implements the InstantiationManager
type InstantiationClient struct {
	db InstantiationClientDbInfo
}

// DeploymentStatus is the structure used to return general status results
// for the Deployment Intent Group
type DeploymentStatus struct {
	Project              string `json:"project,omitempty"`
	CompositeAppName     string `json:"compositeApp,omitempty"`
	CompositeAppVersion  string `json:"compositeAppVersion,omitempty"`
	CompositeProfileName string `json:"compositeProfile,omitempty"`
	status.StatusResult  `json:",inline"`
}

// DeploymentAppsListStatus is the structure used to return the list of Apps
// that have been/were deployed for the DeploymentIntentGroup
type DeploymentAppsListStatus struct {
	Project               string `json:"project,omitempty"`
	CompositeAppName      string `json:"compositeApp,omitempty"`
	CompositeAppVersion   string `json:"compositeAppVersion,omitempty"`
	CompositeProfileName  string `json:"compositeProfile,omitempty"`
	status.AppsListResult `json:",inline"`
}

// DeploymentClustersByAppStatus is the structure used to return the list of Apps
// that have been/were deployed for the DeploymentIntentGroup
type DeploymentClustersByAppStatus struct {
	Project                    string `json:"project,omitempty"`
	CompositeAppName           string `json:"compositeApp,omitempty"`
	CompositeAppVersion        string `json:"compositeAppVersion,omitempty"`
	CompositeProfileName       string `json:"compositeProfile,omitempty"`
	status.ClustersByAppResult `json:",inline"`
}

// DeploymentResourcesByAppStatus is the structure used to return the list of Apps
// that have been/were deployed for the DeploymentIntentGroup
type DeploymentResourcesByAppStatus struct {
	Project                     string `json:"project,omitempty"`
	CompositeAppName            string `json:"compositeApp,omitempty"`
	CompositeAppVersion         string `json:"compositeAppVersion,omitempty"`
	CompositeProfileName        string `json:"compositeProfile,omitempty"`
	status.ResourcesByAppResult `json:",inline"`
}

/*
InstantiationKey used in storing the contextid in the momgodb
It consists of
ProjectName,
CompositeAppName,
CompositeAppVersion,
DeploymentIntentGroup
*/
type InstantiationKey struct {
	Project               string
	CompositeApp          string
	Version               string
	DeploymentIntentGroup string
}

// InstantiationManager is an interface which exposes the
// InstantiationManager functionalities
type InstantiationManager interface {
	Approve(ctx context.Context, p string, ca string, v string, di string) error
	Instantiate(ctx context.Context, p string, ca string, v string, di string) error
	Status(ctx context.Context, p, ca, v, di, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (DeploymentStatus, error)
	GenericStatus(ctx context.Context, p, ca, v, di, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (status.StatusResult, error)
	StatusAppsList(ctx context.Context, p, ca, v, di, qInstance string) (DeploymentAppsListStatus, error)
	StatusClustersByApp(ctx context.Context, p, ca, v, di, qInstance string, fApps []string) (DeploymentClustersByAppStatus, error)
	StatusResourcesByApp(ctx context.Context, p, ca, v, di, qInstance, qType string, fApps, fClusters []string) (DeploymentResourcesByAppStatus, error)
	Terminate(ctx context.Context, p string, ca string, v string, di string) error
	Stop(ctx context.Context, p string, ca string, v string, di string) error
	Migrate(ctx context.Context, p string, ca string, v string, tCav string, di string, tDi string) error
	Update(ctx context.Context, p string, ca string, v string, di string) (int64, error)
	Rollback(ctx context.Context, p string, ca string, v string, di string, rbRev string) error
}

// InstantiationClientDbInfo consists of storeName and tagState
type InstantiationClientDbInfo struct {
	storeName string // name of the mongodb collection to use for Instantiationclient documents
	tagState  string // attribute key name for context object in App Context
}

// NewInstantiationClient returns an instance of InstantiationClient
func NewInstantiationClient() *InstantiationClient {
	return &InstantiationClient{
		db: InstantiationClientDbInfo{
			storeName: "resources",
			tagState:  "stateInfo",
		},
	}
}

//Approve approves an instantiation
func (c InstantiationClient) Approve(ctx context.Context, p string, ca string, v string, di string) error {
	s, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(ctx, di, p, ca, v)
	if err != nil {
		log.Info("DeploymentIntentGroup has no state info ", log.Fields{"DeploymentIntentGroup: ": di})
		return pkgerrors.Wrap(err, "DeploymentIntentGroup has no state info: "+di)
	}
	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		log.Info("Error getting current state from DeploymentIntentGroup stateInfo", log.Fields{"DeploymentIntentGroup ": di})
		return pkgerrors.Wrap(err, "Error getting current state from DeploymentIntentGroup stateInfo: "+di)
	}
	switch stateVal {
	case state.StateEnum.Approved:
		return nil
	case state.StateEnum.Terminated:
		break
	case state.StateEnum.Created:
		break
	case state.StateEnum.Updated:
		break
	case state.StateEnum.Applied:
		return pkgerrors.Errorf("DeploymentIntentGroup is in an invalid state" + stateVal)
	case state.StateEnum.Instantiated:
		return pkgerrors.Errorf("DeploymentIntentGroup has already been instantiated" + di)
	default:
		return pkgerrors.Errorf("DeploymentIntentGroup is in an unknown state" + stateVal)
	}

	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}
	a := state.ActionEntry{
		State:     state.StateEnum.Approved,
		ContextId: "",
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)

	err = db.DBconn.Insert(ctx, c.db.storeName, key, nil, c.db.tagState, s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of the DeploymentIntentGroup: "+di)
	}

	return nil
}

func getOverrideValuesByAppName(ov []OverrideValues, a string) map[string]string {
	for _, eachOverrideVal := range ov {
		if eachOverrideVal.AppName == a {
			return eachOverrideVal.ValuesObj
		}
	}
	return map[string]string{}
}

/*
	findGenericPlacementIntent takes in projectName, CompositeAppName, CompositeAppVersion, DeploymentIntentName
	and returns the name of the genericPlacementIntentName. Returns empty value if string not found.
*/
func findGenericPlacementIntent(ctx context.Context, p, ca, v, di string) (string, error) {
	var gi string
	iList, err := NewIntentClient().GetAllIntents(ctx, p, ca, v, di)
	if err != nil {
		return gi, err
	}
	for _, eachMap := range iList.ListOfIntents {
		if gi, found := eachMap[GenericPlacementIntentName]; found {
			log.Info(":: Name of the generic-placement-intent found ::", log.Fields{"GenPlmtIntent": gi})
			return gi, nil
		}
	}
	log.Info(":: generic-placement-intent not found ! ::", log.Fields{"Searched for GenPlmtIntent": GenericPlacementIntentName})
	return gi, pkgerrors.New("GenericPlacementIntent not found")
}

// GetSortedTemplateForApp returns the sorted templates.
//It takes in arguments - appName, project, compositeAppName, releaseName, compositeProfileName, array of override values
func GetSortedTemplateForApp(ctx context.Context, appName, p, ca, v, rName, cp, namespace string, overrideValues []OverrideValues) ([]helm.KubernetesResourceTemplate, []*helm.Hook, error) {

	log.Info(":: Processing App ::", log.Fields{"appName": appName})

	var sortedTemplates []helm.KubernetesResourceTemplate
	var hookList []*helm.Hook

	aC, err := NewAppClient().GetAppContent(ctx, appName, p, ca, v)
	if err != nil {
		return sortedTemplates, hookList, pkgerrors.Wrap(err, fmt.Sprint("AppContent not found for:: ", appName))
	}
	appContent, err := base64.StdEncoding.DecodeString(aC.FileContent)
	if err != nil {
		return sortedTemplates, hookList, pkgerrors.Wrap(err, "Fail to convert to byte array")
	}

	log.Info(":: Got the app content.. ::", log.Fields{"appName": appName})

	appPC, err := NewAppProfileClient().GetAppProfileContentByApp(ctx, p, ca, v, cp, appName)
	if err != nil {
		return sortedTemplates, hookList, pkgerrors.Wrap(err, fmt.Sprintf("AppProfileContent not found for:: %s", appName))
	}
	appProfileContent, err := base64.StdEncoding.DecodeString(appPC.Profile)
	if err != nil {
		return sortedTemplates, hookList, pkgerrors.Wrap(err, "Fail to convert to byte array")
	}

	log.Info(":: Got the app Profile content .. ::", log.Fields{"appName": appName})

	overrideValuesOfApp := getOverrideValuesByAppName(overrideValues, appName)
	//Convert override values from map to array of strings of the following format
	//foo=bar
	overrideValuesOfAppStr := []string{}
	if overrideValuesOfApp != nil {
		for k, v := range overrideValuesOfApp {
			overrideValuesOfAppStr = append(overrideValuesOfAppStr, k+"="+v)
		}
	}

	sortedTemplates, hookList, err = helm.NewTemplateClient("", namespace, rName,
		ManifestFileName).Resolve(appContent,
		appProfileContent, overrideValuesOfAppStr,
		appName)

	log.Debug(":: Total no. of sorted templates ::", log.Fields{"len(sortedTemplates):": len(sortedTemplates)})

	return sortedTemplates, hookList, err
}

func calculateDirPath(fp string) string {
	sa := strings.Split(fp, "/")
	return "/" + sa[1] + "/" + sa[2] + "/"
}

func cleanTmpfiles(sortedTemplates []helm.KubernetesResourceTemplate) error {
	dp := calculateDirPath(sortedTemplates[0].FilePath)
	for _, st := range sortedTemplates {
		log.Info("Clean up ::", log.Fields{"file: ": st.FilePath})
		err := os.Remove(st.FilePath)
		if err != nil {
			log.Error("Error while deleting file", log.Fields{"file: ": st.FilePath})
			return err
		}
	}
	err := os.RemoveAll(dp)
	if err != nil {
		log.Error("Error while deleting dir", log.Fields{"Dir: ": dp})
		return err
	}
	log.Info("Clean up temp-dir::", log.Fields{"Dir: ": dp})
	return nil
}

func validateLogicalCloud(ctx context.Context, p string, lc string, dcmCloudClient *LogicalCloudClient) error {
	// check that specified logical cloud is already instantiated
	logicalCloud, err := dcmCloudClient.Get(ctx, p, lc)
	if err != nil {
		log.Error("Failed to obtain Logical Cloud specified", log.Fields{"error": err.Error()})
		return pkgerrors.Wrap(err, "Failed to obtain Logical Cloud specified")
	}
	log.Info(":: logicalCloud ::", log.Fields{"logicalCloud": logicalCloud})

	// Check if there was a previous context for this logical cloud
	s, err := dcmCloudClient.GetState(ctx, p, lc)
	if err != nil {
		return err
	}
	cid := state.GetLastContextIdFromStateInfo(s)
	if cid == "" {
		log.Error("The Logical Cloud has never been instantiated", log.Fields{"cid": cid})
		return pkgerrors.New("The Logical Cloud has never been instantiated")
	}

	// make sure rsync status for this logical cloud is Instantiated (instantiated),
	// otherwise the cloud isn't ready to receive the application being instantiated
	acStatus, err := state.GetAppContextStatus(ctx, cid) // new from state
	if err != nil {
		return err
	}
	switch acStatus.Status {
	case appcontext.AppContextStatusEnum.Instantiated:
		log.Info("The Logical Cloud is instantiated, proceeding with DIG instantiation.", log.Fields{"logicalcloud": lc})
	case appcontext.AppContextStatusEnum.Terminated:
		log.Error("The Logical Cloud is not currently instantiated (has been terminated).", log.Fields{"logicalcloud": lc})
		return pkgerrors.New("The Logical Cloud is not currently instantiated (has been terminated).")
	case appcontext.AppContextStatusEnum.Instantiating:
		log.Error("The Logical Cloud is still instantiating, please wait and try again.", log.Fields{"logicalcloud": lc})
		return pkgerrors.New("The Logical Cloud is still instantiating, please wait and try again.")
	case appcontext.AppContextStatusEnum.Terminating:
		log.Error("The Logical Cloud is terminating, so it can no longer receive DIGs.", log.Fields{"logicalcloud": lc})
		return pkgerrors.New("The Logical Cloud is terminating, so it can no longer receive DIGs.")
	case appcontext.AppContextStatusEnum.InstantiateFailed:
		log.Error("The Logical Cloud has failed instanting, so it can't receive DIGs.", log.Fields{"logicalcloud": lc})
		return pkgerrors.New("The Logical Cloud has failed instanting, so it can't receive DIGs.")
	case appcontext.AppContextStatusEnum.TerminateFailed:
		log.Error("The Logical Cloud has failed terminating, so for safety it can no longer receive DIGs.", log.Fields{"logicalcloud": lc})
		return pkgerrors.New("The Logical Cloud has failed terminating, so for safety it can no longer receive DIGs.")
	default:
		log.Error("The Logical Cloud isn't in an expected status so not taking any action.", log.Fields{"logicalcloud": lc, "status": acStatus.Status})
		return pkgerrors.New("The Logical Cloud isn't in an expected status so not taking any action.")
	}

	return nil
}

func getLogicalCloudInfo(ctx context.Context, p string, lc string) ([]common.Cluster, string, string, error) {
	dcmCloudClient := NewLogicalCloudClient()
	logicalCloud, _ := dcmCloudClient.Get(ctx, p, lc)
	if err := validateLogicalCloud(ctx, p, lc, dcmCloudClient); err != nil {
		return nil, "", "", err
	}

	// the namespace where the resources of this app are supposed to deployed to
	namespace := logicalCloud.Specification.NameSpace
	log.Info("Namespace for this logical cloud", log.Fields{"namespace": namespace})
	// level of the logical cloud (0 - admin, 1 - custom)
	level := logicalCloud.Specification.Level

	// get all clusters from specified logical cloud (LC)
	// [candidate in the future for emco-lib]
	dcmClusterClient := NewClusterClient()
	dcmClusters, _ := dcmClusterClient.GetAllClusters(ctx, p, lc)
	log.Info(":: dcmClusters ::", log.Fields{"dcmClusters": dcmClusters})
	return dcmClusters, namespace, level, nil
}

func checkClusters(listOfClusters gpic.ClusterList, dcmClusters []common.Cluster) error {
	// make sure LC can support DIG by validating DIG clusters against LC clusters
	var mandatoryClusters []gpic.ClusterWithName

	for _, mc := range listOfClusters.MandatoryClusters {
		for _, c := range mc.Clusters {
			mandatoryClusters = append(mandatoryClusters, c)
		}
	}

	for _, dcluster := range dcmClusters {
		for i, cluster := range mandatoryClusters {
			if cluster.ProviderName == dcluster.Specification.ClusterProvider && cluster.ClusterName == dcluster.Specification.ClusterName {
				// remove the cluster from slice since it's part of the LC
				lastIndex := len(mandatoryClusters) - 1
				mandatoryClusters[i] = mandatoryClusters[lastIndex]
				mandatoryClusters = mandatoryClusters[:lastIndex]
				// we're done checking this DCM cluster
				break
			}
		}
	}
	if len(mandatoryClusters) > 0 {
		log.Error("The specified Logical Cloud doesn't provide the mandatory clusters", log.Fields{"mandatoryClusters": mandatoryClusters})
		return pkgerrors.New("The specified Logical Cloud doesn't provide the mandatory clusters")
	}

	return nil
}

/*
Instantiate methods takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName. This method is responsible for template resolution, intent
resolution, creation and saving of context for saving into etcd.
*/
func (c InstantiationClient) Instantiate(ctx context.Context, p string, ca string, v string, di string) error {

	log.Info(":: Orchestrator Instantiate ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di})

	span := trace.SpanFromContext(ctx)
	span.AddEvent("retrieve-info")

	// in case of migrate dig comes from JSON body
	dIGrp, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, di, p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "DeploymentIntentGroup not found")
	}

	// handle state info
	s, err := handleStateInfo(ctx, p, ca, v, di)
	if err != nil {
		return pkgerrors.Wrap(err, "Error in handleStateInfo for DeploymentIntent:: "+di)
	}

	// BEGIN : Make app context
	span.AddEvent("create-app-context")
	instantiator := Instantiator{p, ca, v, di, dIGrp}
	cca, err := instantiator.MakeAppContext(ctx)

	if err != nil {
		return pkgerrors.Wrap(err, "Error in making AppContext")
	}
	// END : Make app context

	// BEGIN : callScheduler
	err = callScheduler(ctx, cca.context, cca.ctxval, nil, p, ca, v, di)
	if err != nil {
		return pkgerrors.Wrap(err, "Error in callScheduler")
	}
	// END : callScheduler


	// BEGIN : Rsync code
	err = callRsyncInstall(ctx, cca.ctxval)
	if err != nil {
		deleteAppContext(ctx, cca.context)
		return pkgerrors.Wrap(err, "Error calling rsync")
	}
	// END : Rsync code

	err = storeAppContextIntoMetaDB(ctx, cca.ctxval, c.db.storeName, c.db.tagState, s, p, ca, v, di)

	// Call Post INSTANTIATE Event for all controllers
	//_ = callPostEventScheduler(ctx, cca.ctxval, p, ca, v, di, "INSTANTIATE")

	log.Info(":: Done with instantiation call to rsync... ::", log.Fields{"CompositeAppName": ca})
	return err
}

/*
Status takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName. This method is responsible obtaining the status of
the deployment, which is made available in the appcontext.
*/
func (c InstantiationClient) Status(ctx context.Context, p, ca, v, di, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (DeploymentStatus, error) {
	dIGrp, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, di, p, ca, v)
	if err != nil {
		return DeploymentStatus{}, pkgerrors.Wrap(err, "DeploymentIntentGroup not found")
	}

	diState, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(ctx, di, p, ca, v)
	if err != nil {
		return DeploymentStatus{}, pkgerrors.Wrap(err, "DeploymentIntentGroup state not found: "+di)
	}

	statusResponse, err := status.PrepareStatusResult(ctx, diState, qInstance, qType, qOutput, fApps, fClusters, fResources)
	if err != nil {
		return DeploymentStatus{}, err
	}
	statusResponse.Name = di
	diStatus := DeploymentStatus{
		Project:              p,
		CompositeAppName:     ca,
		CompositeAppVersion:  v,
		CompositeProfileName: dIGrp.Spec.Profile,
		StatusResult:         statusResponse,
	}

	return diStatus, nil
}

/*
GenericStatus takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName. This method is responsible obtaining the status of
the deployment, which is made available in the appcontext.
*/
func (c InstantiationClient) GenericStatus(ctx context.Context, p, ca, v, di, qStatusInstance, qType, qOutput string, fApps, fClusters, fResources []string) (status.StatusResult, error) {
	diState, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(ctx, di, p, ca, v)
	if err != nil {
		return status.StatusResult{}, pkgerrors.Wrap(err, "DeploymentIntentGroup state not found: "+di)
	}

	qInstance, err := state.GetContextIdForStatusContextId(diState, qStatusInstance)
	if err != nil {
		return status.StatusResult{}, err
	}

	statusResponse, err := status.GenericPrepareStatusResult(ctx, status.DeploymentIntentGroupStatusQuery, diState, qInstance, qType, qOutput, fApps, fClusters, fResources)
	if err != nil {
		return status.StatusResult{}, err
	}
	statusResponse.Name = di

	return statusResponse, nil
}

/*
StatusAppsList takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName. This method returns the list of apps in use for the given instance
of appcontext of this deployment intent group.
*/
func (c InstantiationClient) StatusAppsList(ctx context.Context, p, ca, v, di, qInstance string) (DeploymentAppsListStatus, error) {
	dIGrp, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, di, p, ca, v)
	if err != nil {
		return DeploymentAppsListStatus{}, pkgerrors.Wrap(err, "DeploymentIntentGroup not found")
	}

	diState, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(ctx, di, p, ca, v)
	if err != nil {
		return DeploymentAppsListStatus{}, pkgerrors.Wrap(err, "DeploymentIntentGroup state not found: "+di)
	}

	statusResponse, err := status.PrepareAppsListStatusResult(ctx, diState, qInstance)
	if err != nil {
		return DeploymentAppsListStatus{}, err
	}
	statusResponse.Name = di
	diStatus := DeploymentAppsListStatus{
		Project:              p,
		CompositeAppName:     ca,
		CompositeAppVersion:  v,
		CompositeProfileName: dIGrp.Spec.Profile,
		AppsListResult:       statusResponse,
	}

	return diStatus, nil
}

/*
StatusClustersByApp takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName. This method returns the list of apps in use for the given instance
of appcontext of this deployment intent group.
*/
func (c InstantiationClient) StatusClustersByApp(ctx context.Context, p, ca, v, di, qInstance string, fApps []string) (DeploymentClustersByAppStatus, error) {
	dIGrp, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, di, p, ca, v)
	if err != nil {
		return DeploymentClustersByAppStatus{}, pkgerrors.Wrap(err, "DeploymentIntentGroup not found")
	}

	diState, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(ctx, di, p, ca, v)
	if err != nil {
		return DeploymentClustersByAppStatus{}, pkgerrors.Wrap(err, "DeploymentIntentGroup state not found: "+di)
	}

	statusResponse, err := status.PrepareClustersByAppStatusResult(ctx, diState, qInstance, fApps)
	if err != nil {
		return DeploymentClustersByAppStatus{}, err
	}
	statusResponse.Name = di
	diStatus := DeploymentClustersByAppStatus{
		Project:              p,
		CompositeAppName:     ca,
		CompositeAppVersion:  v,
		CompositeProfileName: dIGrp.Spec.Profile,
		ClustersByAppResult:  statusResponse,
	}

	return diStatus, nil
}

/*
StatusResourcesByApp takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName. This method returns the list of apps in use for the given instance
of appcontext of this deployment intent group.
*/
func (c InstantiationClient) StatusResourcesByApp(ctx context.Context, p, ca, v, di, qInstance, qType string, fApps, fClusters []string) (DeploymentResourcesByAppStatus, error) {
	dIGrp, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, di, p, ca, v)
	if err != nil {
		return DeploymentResourcesByAppStatus{}, pkgerrors.Wrap(err, "DeploymentIntentGroup not found")
	}

	diState, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(ctx, di, p, ca, v)
	if err != nil {
		return DeploymentResourcesByAppStatus{}, pkgerrors.Wrap(err, "DeploymentIntentGroup state not found: "+di)
	}

	statusResponse, err := status.PrepareResourcesByAppStatusResult(ctx, diState, qInstance, qType, fApps, fClusters)
	if err != nil {
		return DeploymentResourcesByAppStatus{}, err
	}
	statusResponse.Name = di
	diStatus := DeploymentResourcesByAppStatus{
		Project:              p,
		CompositeAppName:     ca,
		CompositeAppVersion:  v,
		CompositeProfileName: dIGrp.Spec.Profile,
		ResourcesByAppResult: statusResponse,
	}

	return diStatus, nil
}

/*
Terminate takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName and calls rsync to terminate.
*/
func (c InstantiationClient) Terminate(ctx context.Context, p string, ca string, v string, di string) error {

	s, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(ctx, di, p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "DeploymentIntentGroup has no state info: "+di)
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting current state from DeploymentIntentGroup stateInfo: "+di)
	}

	if stateVal != state.StateEnum.Instantiated && stateVal != state.StateEnum.InstantiateStopped {
		return pkgerrors.Errorf("DeploymentIntentGroup is not instantiated :" + di)
	}

	currentCtxId := state.GetLastContextIdFromStateInfo(s)

	// BEGIN : callScheduler
	err = callTerminateScheduler(ctx, currentCtxId, p, ca, v, di)
	if err != nil {
		return pkgerrors.Wrap(err, "Error in callTerminateScheduler")
	}
	// END : callScheduler

	if stateVal == state.StateEnum.InstantiateStopped {
		err = state.UpdateAppContextStopFlag(ctx, currentCtxId, false)
		if err != nil {
			return err
		}
	}

	var ac appcontext.AppContext
	_, err = ac.LoadAppContext(ctx, currentCtxId)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting AppContext with Id: %v", currentCtxId)
	}

	// Get the composite app meta
	m, err := ac.GetCompositeAppMeta(ctx)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting CompositeAppMeta")
	}
	if len(m.ChildContextIDs) > 0 {
		// Uninstall the resources associated to the child contexts
		for _, childContextID := range m.ChildContextIDs {
			err = callRsyncUninstall(ctx, childContextID)
			if err != nil {
				log.Warn("Unable to uninstall the resources associated to the child context", log.Fields{"childContext": childContextID})
				continue
			}
		}
	}
	// Uninstall the resources associated to the parent contexts
	err = callRsyncUninstall(ctx, currentCtxId)
	if err != nil {
		return err
	}

	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}
	a := state.ActionEntry{
		State:     state.StateEnum.Terminated,
		ContextId: currentCtxId,
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)

	err = db.DBconn.Insert(ctx, c.db.storeName, key, nil, c.db.tagState, s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of the DeploymentIntentGroup: "+di)
	}
	// Call Post Terminate Event for all controllers
	//_ = callPostEventScheduler(ctx, currentCtxId, p, ca, v, di, "TERMINATE")

	return nil
}

/*
Stop takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName and sets the stopFlag in the associated appContext.
*/
func (c InstantiationClient) Stop(ctx context.Context, p string, ca string, v string, di string) error {

	s, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(ctx, di, p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "DeploymentIntentGroup has no state info: "+di)
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting current state from DeploymentIntentGroup stateInfo: "+di)
	}
	stopState := state.StateEnum.Undefined
	switch stateVal {
	case state.StateEnum.Approved:
		return pkgerrors.Errorf("DeploymentIntentGroup has not been instantiated:" + di)
	case state.StateEnum.Instantiated:
		stopState = state.StateEnum.InstantiateStopped
		break
	case state.StateEnum.Terminated:
		stopState = state.StateEnum.TerminateStopped
		break
	case state.StateEnum.Applied:
		return pkgerrors.New("DeploymentIntentGroup is in an invalid state:" + di)
		break
	case state.StateEnum.TerminateStopped:
		return pkgerrors.New("DeploymentIntentGroup termination already stopped: " + di)
	case state.StateEnum.InstantiateStopped:
		return pkgerrors.New("DeploymentIntentGroup instantiation already stopped: " + di)
	case state.StateEnum.Created:
		return pkgerrors.New("DeploymentIntentGroup have not been approved: " + di)
	default:
		return pkgerrors.New("DeploymentIntentGroup is in an invalid state: " + di + " " + stateVal)
	}

	currentCtxId := state.GetLastContextIdFromStateInfo(s)

	acStatus, err := state.GetAppContextStatus(ctx, currentCtxId)
	if err != nil {
		return err
	}
	if acStatus.Status != appcontext.AppContextStatusEnum.Instantiating &&
		acStatus.Status != appcontext.AppContextStatusEnum.Terminating {
		return pkgerrors.Errorf("DeploymentIntentGroup is not instantiating or terminating:" + di)
	}
	err = state.UpdateAppContextStopFlag(ctx, currentCtxId, true)
	if err != nil {
		return err
	}

	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}
	a := state.ActionEntry{
		State:     stopState,
		ContextId: currentCtxId,
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)

	err = db.DBconn.Insert(ctx, c.db.storeName, key, nil, c.db.tagState, s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of the DeploymentIntentGroup: "+di)
	}

	return nil
}
