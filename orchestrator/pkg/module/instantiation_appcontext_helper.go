// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

/*
This file deals with the interaction of instantiation flow and etcd.
It contains methods for creating appContext, saving cluster and resource details to etcd.

*/
import (
	"context"
	"encoding/json"
	"io/ioutil"
	"time"

	pkgerrors "github.com/pkg/errors"
	"github.com/yourbasic/graph"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	gpic "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/gpic"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/utils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/utils/helm"
)

// resource consists of name of reource
type resource struct {
	name        string
	filecontent string
}

type contextForCompositeApp struct {
	context            appcontext.AppContext
	ctxval             interface{}
	compositeAppHandle interface{}
}

// TODO move into a better place or reuse existing struct
type K8sResource struct {
	Metadata MetadataList `yaml:"metadata"`
}

// TODO move into a better place or reuse existing struct
type MetadataList struct {
	Namespace string `yaml:"namespace"`
}

type appOrderInstr struct {
	Apporder []string `json:"apporder"`
}

type appDepInstr struct {
	AppDepMap map[string]string `json:"appdependency"`
}

type AppHandler struct {
	appName    string
	namespace  string
	clusters   gpic.ClusterList
	ht         []helm.KubernetesResourceTemplate
	hk         []*helm.Hook
	dependency []AdSpecData
}

// deleteAppContext removes an appcontext
func deleteAppContext(ctx context.Context, ct appcontext.AppContext) error {
	err := ct.DeleteCompositeApp(ctx)
	if err != nil {
		log.Warn(":: Error deleting AppContext ::", log.Fields{"Error": err})
		return pkgerrors.Wrapf(err, "Error Deleteing AppContext")
	}
	return nil
}

// getResources shall take in the sorted templates and output the resources
// which consists of name(name+kind) and filecontent
// Returns regular resources and crd resources in separate arrays
func getResources(st []helm.KubernetesResourceTemplate) ([]resource, []resource, error) {
	var resources []resource
	var crdResources []resource
	for _, t := range st {
		yamlStruct, err := utils.ExtractYamlParameters(t.FilePath)
		yamlFile, err := ioutil.ReadFile(t.FilePath)
		if err != nil {
			return nil, nil, pkgerrors.Wrap(err, "Failed to get the resources")
		}
		n := yamlStruct.Metadata.Name + SEPARATOR + yamlStruct.Kind
		// This might happen when the rendered file just has some comments inside, no real k8s object.
		if n == SEPARATOR {
			log.Info(":: Ignoring, Unable to render the template ::", log.Fields{"YAML PATH": t.FilePath})
			continue
		}
		if yamlStruct.Kind == "CustomResourceDefinition" {
			crdResources = append(crdResources, resource{name: n, filecontent: string(yamlFile)})
		} else {
			resources = append(resources, resource{name: n, filecontent: string(yamlFile)})
		}
		log.Info(":: Added resource into resource-order ::", log.Fields{"ResourceName": n})
	}
	return crdResources, resources, nil
}

// getHookResources returns hooks in resource format
func getHookResources(hk []*helm.Hook) (map[string][]resource, error) {
	resources := make(map[string][]resource)
	r, err := helm.GetHooksByEvent(hk)
	if err != nil {
		return resources, err
	}
	for hookName, t := range r {
		for _, res := range t {
			yamlStruct, err := utils.ExtractYamlParameters(res.KRT.FilePath)
			if err != nil {
				return resources, pkgerrors.Wrap(err, "Failed to extract file path")
			}
			yamlFile, err := ioutil.ReadFile(res.KRT.FilePath)
			if err != nil {
				return resources, pkgerrors.Wrap(err, "Failed to read file")
			}
			n := yamlStruct.Metadata.Name + SEPARATOR + yamlStruct.Kind
			// This might happen when the rendered file just has some comments inside, no real k8s object.
			if n == SEPARATOR {
				log.Info(":: Ignoring, Unable to render the template ::", log.Fields{"YAML PATH": res.KRT.FilePath})
				continue
			}
			resources[hookName] = append(resources[hookName], resource{name: n, filecontent: string(yamlFile)})
		}
	}
	return resources, nil
}

func (ah *AppHandler) addResourcesToCluster(ctx context.Context, ct appcontext.AppContext, ch interface{}) ([]resource, error) {

	crdResources, resources, err := getResources(ah.ht)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "Unable to get the resources")
	}

	var resOrderInstr struct {
		Resorder []string `json:"resorder"`
	}

	for _, resource := range resources {
		resOrderInstr.Resorder = append(resOrderInstr.Resorder, resource.name)
		_, err := ct.AddResource(ctx, ch, resource.name, resource.filecontent)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "Error adding resource ::%s to AppContext", resource.name)
		}
	}
	// Add resource order for the cluster
	jresOrderInstr, _ := json.Marshal(resOrderInstr)
	_, err = ct.AddInstruction(ctx, ch, "resource", "order", string(jresOrderInstr))
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "Error adding instruction for resource order")
	}
	return crdResources, nil
}

// Add Hook resources and add an instruction
func (ah *AppHandler) addHooksToCluster(ctx context.Context, ct appcontext.AppContext, ch interface{}, crdResources []resource) error {
	hk, err := getHookResources(ah.hk)
	if err != nil {
		return err
	}
	// If no hooks present and no crdResources then return without adding dependency instruction
	if len(hk) <= 0 && len(crdResources) <= 0 {
		return nil
	}
	var resDepInstr struct {
		Resdep map[string][]string `json:"resdependency"`
	}
	resdep := make(map[string][]string)

	// Add Hooks resources and add in the dependency instruction
	for name, t := range hk {
		for _, res := range t {
			_, err := ct.AddResource(ctx, ch, res.name, res.filecontent)
			if err != nil {
				return err
			}
			resdep[name] = append(resdep[name], res.name)
		}
	}
	// Add CRD Resources also
	for _, res := range crdResources {
		_, err := ct.AddResource(ctx, ch, res.name, res.filecontent)
		if err != nil {
			return err
		}
		resdep["crd-install"] = append(resdep["crd-install"], res.name)
	}
	log.Info(":: Hook and CRD resources  ::", log.Fields{"Dependency": resdep})
	resDepInstr.Resdep = resdep
	jresDepInstr, _ := json.Marshal(resDepInstr)
	_, err = ct.AddInstruction(ctx, ch, "resource", "dependency", string(jresDepInstr))
	if err != nil {
		return pkgerrors.Wrapf(err, "Error adding instruction for resource to AppContext")
	}
	return nil
}

//addClustersToAppContextHelper helper to add clusters
func (ah *AppHandler) addClustersToAppContextHelper(ctx context.Context, cg []gpic.ClusterGroup, ct appcontext.AppContext, appHandle interface{}) error {
	for _, eachGrp := range cg {
		oc := eachGrp.Clusters
		gn := eachGrp.GroupNumber

		for _, eachCluster := range oc {
			p := eachCluster.ProviderName
			n := eachCluster.ClusterName

			clusterhandle, err := ct.AddCluster(ctx, appHandle, p+SEPARATOR+n)

			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding Resources to Cluster(provider::%s, name::%s and groupName:: %s) to AppContext", p, n, gn)
			}
			log.Info(":: Added cluster ::", log.Fields{"Cluster ": p + SEPARATOR + n})

			err = ct.AddClusterMetaGrp(ctx, clusterhandle, gn)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding Resources to Cluster(provider::%s, name::%s and groupName:: %s) to AppContext", p, n, gn)
			}
			log.Info(":: Added cluster ::", log.Fields{"Cluster ": p + SEPARATOR + n, "GroupNumber ": gn})

			crdResources, err := ah.addResourcesToCluster(ctx, ct, clusterhandle)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding Resources to Cluster(provider::%s, name::%s and groupName:: %s) to AppContext", p, n, gn)
			}
			err = ah.addHooksToCluster(ctx, ct, clusterhandle, crdResources)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding Hooks to Cluster(provider::%s, name::%s and groupName:: %s) to AppContext", p, n, gn)
			}
		}
	}
	return nil
}

func (ah *AppHandler) addAppToAppContext(ctx context.Context, cxtForCApp contextForCompositeApp) error {

	ct := cxtForCApp.context
	appHandle, err := ct.AddApp(ctx, cxtForCApp.compositeAppHandle, ah.appName)
	if err != nil {
		return pkgerrors.Wrap(err, "Error adding App to AppContext")
	}
	if len(ah.dependency) > 0 {
		// Add Dependency Instruction for the App
		dependency, err := json.Marshal(ah.dependency)
		if err != nil {
			return pkgerrors.Wrap(err, "Error Marshalling dependency for app")
		}
		dh, err := ct.AddLevelValue(ctx, appHandle, "instruction/dependency", string(dependency))
		if err != nil {
			return pkgerrors.Wrap(err, "Error adding App dependency to AppContext")
		}
		log.Info(":: appDep ::", log.Fields{"dh": dh, "dep": string(dependency)})
	}

	mClusters := ah.clusters.MandatoryClusters
	oClusters := ah.clusters.OptionalClusters

	err = ah.addClustersToAppContextHelper(ctx, mClusters, ct, appHandle)
	if err != nil {
		return err
	}
	log.Info("::Added mandatory clusters to the AppContext", log.Fields{})

	err = ah.addClustersToAppContextHelper(ctx, oClusters, ct, appHandle)
	if err != nil {
		return err
	}
	log.Info("::Added optional clusters to the AppContext", log.Fields{})
	return nil
}

/*
verifyResources method is just to check if the resource handles are correctly saved.
*/
func (ah *AppHandler) verifyResources(ctx context.Context, cxtForCApp contextForCompositeApp) error {

	ct := cxtForCApp.context
	_, resources, err := getResources(ah.ht)
	if err != nil {
		return pkgerrors.Wrapf(err, "Unable to get the resources")
	}
	for _, cg := range ah.clusters.OptionalClusters {
		gn := cg.GroupNumber
		oc := cg.Clusters
		for _, eachCluster := range oc {
			p := eachCluster.ProviderName
			n := eachCluster.ClusterName
			cn := p + SEPARATOR + n

			for _, res := range resources {
				rh, err := ct.GetResourceHandle(ctx, ah.appName, cn, res.name)
				if err != nil {
					return pkgerrors.Wrapf(err, "Error getting resource handle for resource :: %s, app:: %s, cluster :: %s, groupName :: %s", ah.appName, res.name, cn, gn)
				}
				log.Info(":: GetResourceHandle ::", log.Fields{"ResourceHandler": rh, "appName": ah.appName, "Cluster": cn, "Resource": res.name})
			}
		}
		grpMap, err := ct.GetClusterGroupMap(ctx, ah.appName)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error getting GetGroupMap for app:: %s, groupName :: %s", ah.appName, gn)
		}
		log.Info(":: GetGroupMapReults ::", log.Fields{"GroupMap": grpMap})
	}

	for _, mClusters := range ah.clusters.MandatoryClusters {
		for _, mc := range mClusters.Clusters {
			p := mc.ProviderName
			n := mc.ClusterName
			cn := p + SEPARATOR + n
			for _, res := range resources {
				rh, err := ct.GetResourceHandle(ctx, ah.appName, cn, res.name)
				if err != nil {
					return pkgerrors.Wrapf(err, "Error getting resoure handle for resource :: %s, app:: %s, cluster :: %s", ah.appName, res.name, cn)
				}
				log.Info(":: GetResourceHandle ::", log.Fields{"ResourceHandler": rh, "appName": ah.appName, "Cluster": cn, "Resource": res.name})
			}
		}
	}
	return nil
}

func storeAppContextIntoMetaDB(ctx context.Context, ctxval interface{}, storeName string, colName string, s state.StateInfo, p, ca, v, di string) error {

	// BEGIN:: save the context in the orchestrator db record
	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}
	a := state.ActionEntry{
		State:     state.StateEnum.Instantiated,
		ContextId: ctxval.(string),
		TimeStamp: time.Now(),
		Revision:  1,
	}
	s.StatusContextId = ctxval.(string)
	s.Actions = append(s.Actions, a)
	err := db.DBconn.Insert(ctx, storeName, key, nil, colName, s)
	if err != nil {
		log.Warn(":: Error updating DeploymentIntentGroup state in DB ::", log.Fields{"Error": err.Error(), "DeploymentIntentGroup": di, "CompositeApp": ca, "CompositeAppVersion": v, "Project": p, "AppContext": ctxval.(string)})
		return pkgerrors.Wrap(err, "Error adding DeploymentIntentGroup state to DB")
	}
	// END:: save the context in the orchestrator db record
	return nil
}

func handleStateInfo(ctx context.Context, p, ca, v, di string) (state.StateInfo, error) {

	s, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(ctx, di, p, ca, v)
	if err != nil {
		return state.StateInfo{}, pkgerrors.Wrap(err, "Error retrieving DeploymentIntentGroup stateInfo: "+di)
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return state.StateInfo{}, pkgerrors.Errorf("Error getting current state from DeploymentIntentGroup stateInfo: " + di)
	}
	switch stateVal {
	case state.StateEnum.Approved:
		break
	case state.StateEnum.Terminated:
		break // TODO - ideally, should check that all resources have completed being terminated
	case state.StateEnum.TerminateStopped:
		break
	case state.StateEnum.Created:
		return state.StateInfo{}, pkgerrors.Errorf("DeploymentIntentGroup must be Approved before instantiating" + di)
	case state.StateEnum.Applied:
		return state.StateInfo{}, pkgerrors.Errorf("DeploymentIntentGroup is in an invalid state" + di)
	case state.StateEnum.InstantiateStopped:
		return state.StateInfo{}, pkgerrors.Errorf("DeploymentIntentGroup has already been instantiated and stopped" + di)
	case state.StateEnum.Instantiated:
		return state.StateInfo{}, pkgerrors.Errorf("DeploymentIntentGroup has already been instantiated" + di)
	default:
		return state.StateInfo{}, pkgerrors.Errorf("DeploymentIntentGroup is in an unknown state" + stateVal)
	}
	return s, nil
}

// Check if cycles exist in dependency
func checkDependency(ctx context.Context, allApps []App, p, ca, v string) bool {

	gph := make(map[string]int)
	g := graph.New(len(allApps))
	// Check Dependency

	// Assign a number to each app
	for i, eachApp := range allApps {
		gph[eachApp.Metadata.Name] = i
	}
	for _, eachApp := range allApps {
		// Read app dependency
		appDep, err := NewAppDependencyClient().GetAllSpecAppDependency(ctx, p, ca, v, eachApp.Metadata.Name)
		if err == nil && len(appDep) > 0 {
			for _, b := range appDep {
				if item, ok := gph[b.AppName]; ok {
					g.Add(gph[eachApp.Metadata.Name], item)
				} else {
					log.Error("Unknown App in dependency", log.Fields{"app": eachApp.Metadata.Name, "Dependency app": b.AppName})
					return false
				}
			}
		}
	}
	// Empty graph is acyclic
	return graph.Acyclic(g)

}
