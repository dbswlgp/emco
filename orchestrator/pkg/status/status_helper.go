// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package status

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	pkgerrors "github.com/pkg/errors"
	rb "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"gitlab.com/project-emco/core/emco-base/src/monitor/pkg/client/clientset/versioned/scheme"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/utils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/resourcestatus"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// decodeYAML reads a YAMl []byte to extract the Kubernetes object definition
func decodeYAML(y []byte, into runtime.Object) (runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(y, nil, into)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize YAML error")
	}

	return obj, nil
}

func getUnstruct(y []byte) (unstructured.Unstructured, error) {
	//Decode the yaml file to create a runtime.Object
	unstruct := unstructured.Unstructured{}
	//Ignore the returned obj as we expect the data in unstruct
	_, err := decodeYAML(y, &unstruct)
	if err != nil {
		log.Info(":: Error decoding YAML ::", log.Fields{"object": y, "error": err})
		return unstructured.Unstructured{}, pkgerrors.Wrap(err, "Decode object error")
	}

	return unstruct, nil
}

func updateReadyCounts(ready bool, cnts map[string]int) string {
	if ready {
		cnt := cnts["Ready"]
		cnts["Ready"] = cnt + 1
		return "Ready"
	} else {
		cnt := cnts["NotReady"]
		cnts["NotReady"] = cnt + 1
		return "NotReady"
	}
}

func updateNotPresentCount(increment bool, cnts map[string]int) {
	cnt := cnts["NotPresent"]
	if increment {
		cnts["NotPresent"] = cnt + 1
	} else {
		if cnt-1 == 0 {
			delete(cnts, "NotPresent")
		} else {
			cnts["NotPresent"] = cnt - 1
		}
	}
}

// return true if resource is added, false if already present and just updated
func updateResourceList(resourceList *[]ResourceStatus, r ResourceStatus, qType string) bool {
	// see if resource is already in the list - then just update it
	for i, re := range *resourceList {
		if re.Name == r.Name &&
			re.Gvk.Group == r.Gvk.Group &&
			re.Gvk.Version == r.Gvk.Version &&
			re.Gvk.Kind == r.Gvk.Kind {
			(*resourceList)[i].Detail = r.Detail
			if qType == "cluster" {
				(*resourceList)[i].ClusterStatus = r.ClusterStatus
			} else {
				(*resourceList)[i].ReadyStatus = r.ReadyStatus
			}
			return false
		}
	}

	*resourceList = append(*resourceList, r)
	return true
}

// getClusterResources takes in a ResourceBundleStateStatus CR and returns a list of ResourceStatus elments
func getClusterResources(rbData rb.ResourceBundleStateStatus, qType, qOutput string, fResources []string,
	resourceList *[]ResourceStatus, cnts map[string]int) (int, error) {

	readyChecker := status.NewReadyChecker(status.PausedAsReady(true), status.CheckJobs(true))

	count := 0

	for _, p := range rbData.PodStatuses {
		if !keepResource(p.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = p.Name
		r.Gvk = (&p.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = p
		}
		labels := p.GetLabels()
		_, job := labels["job-name"]
		if qType == "cluster" {
			if job {
				r.ClusterStatus = updateReadyCounts(readyChecker.PodSuccess(&p), cnts)
			} else {
				r.ClusterStatus = updateReadyCounts(readyChecker.PodReady(&p), cnts)
			}
		} else {
			if job {
				r.ReadyStatus = updateReadyCounts(readyChecker.PodSuccess(&p), cnts)
			} else {
				r.ReadyStatus = updateReadyCounts(readyChecker.PodReady(&p), cnts)
			}
		}
		if updateResourceList(resourceList, r, qType) {
			count++
		} else if markAsNotPresent(r.Gvk) {
			updateNotPresentCount(false, cnts)
		}
	}

	for _, s := range rbData.ServiceStatuses {
		if !keepResource(s.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = s.Name
		r.Gvk = (&s.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = s
		}
		if qType == "cluster" {
			r.ClusterStatus = updateReadyCounts(readyChecker.ServiceReady(&s), cnts)
		} else {
			r.ReadyStatus = updateReadyCounts(readyChecker.ServiceReady(&s), cnts)
		}
		if updateResourceList(resourceList, r, qType) {
			count++
		} else if markAsNotPresent(r.Gvk) {
			updateNotPresentCount(false, cnts)
		}
	}

	for _, d := range rbData.DeploymentStatuses {
		if !keepResource(d.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = d.Name
		r.Gvk = (&d.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = d
		}
		if qType == "cluster" {
			r.ClusterStatus = updateReadyCounts(readyChecker.DeploymentReady(&d), cnts)
		} else {
			r.ReadyStatus = updateReadyCounts(readyChecker.DeploymentReady(&d), cnts)
		}
		if updateResourceList(resourceList, r, qType) {
			count++
		} else if markAsNotPresent(r.Gvk) {
			updateNotPresentCount(false, cnts)
		}
	}

	for _, c := range rbData.ConfigMapStatuses {
		if !keepResource(c.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = c.Name
		r.Gvk = (&c.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = c
		}
		if qType == "cluster" {
			r.ClusterStatus = updateReadyCounts(true, cnts)
		} else {
			r.ReadyStatus = updateReadyCounts(true, cnts)
		}
		if updateResourceList(resourceList, r, qType) {
			count++
		} else if markAsNotPresent(r.Gvk) {
			updateNotPresentCount(false, cnts)
		}
	}

	for _, d := range rbData.DaemonSetStatuses {
		if !keepResource(d.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = d.Name
		r.Gvk = (&d.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = d
		}
		if qType == "cluster" {
			r.ClusterStatus = updateReadyCounts(readyChecker.DaemonSetReady(&d), cnts)
		} else {
			r.ReadyStatus = updateReadyCounts(readyChecker.DaemonSetReady(&d), cnts)
		}
		if updateResourceList(resourceList, r, qType) {
			count++
		} else if markAsNotPresent(r.Gvk) {
			updateNotPresentCount(false, cnts)
		}
	}

	for _, j := range rbData.JobStatuses {
		if !keepResource(j.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = j.Name
		r.Gvk = (&j.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = j
		}
		if qType == "cluster" {
			r.ClusterStatus = updateReadyCounts(readyChecker.JobReady(&j), cnts)
		} else {
			r.ReadyStatus = updateReadyCounts(readyChecker.JobReady(&j), cnts)
		}
		if updateResourceList(resourceList, r, qType) {
			count++
		} else if markAsNotPresent(r.Gvk) {
			updateNotPresentCount(false, cnts)
		}
	}

	for _, s := range rbData.StatefulSetStatuses {
		if !keepResource(s.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = s.Name
		r.Gvk = (&s.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = s
		}
		if qType == "cluster" {
			r.ClusterStatus = updateReadyCounts(readyChecker.StatefulSetReady(&s), cnts)
		} else {
			r.ReadyStatus = updateReadyCounts(readyChecker.StatefulSetReady(&s), cnts)
		}
		if updateResourceList(resourceList, r, qType) {
			count++
		} else if markAsNotPresent(r.Gvk) {
			updateNotPresentCount(false, cnts)
		}
	}

	for _, s := range rbData.CsrStatuses {
		if !keepResource(s.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = s.Name
		r.Gvk = (&s.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = s
		}
		if qType == "cluster" {
			r.ClusterStatus = updateReadyCounts(true, cnts)
		} else {
			r.ReadyStatus = updateReadyCounts(true, cnts)
		}
		if updateResourceList(resourceList, r, qType) {
			count++
		} else if markAsNotPresent(r.Gvk) {
			updateNotPresentCount(false, cnts)
		}
	}

	for _, s := range rbData.ResourceStatuses {
		if !keepResource(s.Name, fResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = s.Name
		r.Gvk.Group = s.Group
		r.Gvk.Version = s.Version
		r.Gvk.Kind = s.Kind
		if qOutput == "detail" {
			r.Detail = json.RawMessage(string(s.Res))
		}
		if qType == "cluster" {
			r.ClusterStatus = updateReadyCounts(true, cnts)
		} else {
			r.ReadyStatus = updateReadyCounts(true, cnts)
		}
		if updateResourceList(resourceList, r, qType) {
			count++
		} else if markAsNotPresent(r.Gvk) {
			updateNotPresentCount(false, cnts)
		}
	}

	return count, nil
}

// isResourceHandle takes a cluster handle and determines if the other handle parameter is a resource handle for this cluster
// handle.  It does this by verifying that the cluster handle is a prefix of the handle and that the remainder of the handle
// is a value that matches to a resource format:  "resource/<name>+<type>/"
// Example cluster handle:
// /context/6385596659306465421/app/network-intents/cluster/vfw-cluster-provider+edge01/
// Example resource handle:
// /context/6385596659306465421/app/network-intents/cluster/vfw-cluster-provider+edge01/resource/emco-private-net+ProviderNetwork/
func isResourceHandle(ch, h interface{}) bool {
	clusterHandle := fmt.Sprintf("%v", ch)
	handle := fmt.Sprintf("%v", h)
	diff := strings.Split(handle, clusterHandle)

	if len(diff) != 2 && diff[0] != "" {
		return false
	}

	parts := strings.Split(diff[1], "/")

	if len(parts) == 3 &&
		parts[0] == "resource" &&
		len(strings.Split(parts[1], "+")) == 2 &&
		parts[2] == "" {
		return true
	} else {
		return false
	}
}

// getParallelHandle - return the handle h with the original contextId replaced with ac's contextId
func getParallelHandle(ctx context.Context, h interface{}, ac appcontext.AppContext) string {
	ph := fmt.Sprintf("%v", h)
	oldHs := strings.SplitN(ph, "/", -1)
	if len(oldHs) < 3 {
		return ""
	}

	ch, err := ac.GetCompositeAppHandle(ctx)
	if err != nil {
		return ""
	}

	newHs := strings.SplitN(fmt.Sprintf("%v", ch), "/", -1)
	if len(newHs) < 3 {
		return ""
	}

	return strings.Replace(ph, oldHs[2], newHs[2], 1)
}

// keepResource keeps a resource if the filter list is empty or if the resource is part of the list
func keepResource(r string, rList []string) bool {
	if len(rList) == 0 {
		return true
	}
	for _, res := range rList {
		if r == res {
			return true
		}
	}
	return false
}

// markAsNotPresent - only identify key resources that are tracked as NotPresent
// This will help to avoid false NotReady status queries due to some resources not
// being tracked by the monitor process.
func markAsNotPresent(gvk schema.GroupVersionKind) bool {
	switch gvk {
	case schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
		schema.GroupVersionKind{Version: "v1", Kind: "Service"},
		schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"},
		schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"},
		schema.GroupVersionKind{Version: "v1", Kind: "Pod"},
		schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
		schema.GroupVersionKind{Group: "certificates.k8s.io", Version: "v1", Kind: "CertificateSigningRequest"}:
		return true
	}
	return false
}

// For status get the resource from reference appContext
// In case of no update it'll refer back to the same appContext
// 'ac' is the target appcontext, 'sac' is the status appcontext
// and 'h' is the resource handle in the StatusAppContext
func getResourceFromReference(ctx context.Context, ac, sac appcontext.AppContext, h interface{}, appName, cluster string) (unstructured.Unstructured, error) {

	var val = ""
	var err error
	// Read reference appContext value
	sh, err := sac.GetLevelHandle(ctx, h, "reference")
	if err == nil {
		s, err := sac.GetValue(ctx, sh)
		if err == nil {
			js, err := json.Marshal(s)
			if err == nil {
				json.Unmarshal(js, &val)
			}
		}
	}

	handle := fmt.Sprintf("%v", h)
	// Assuming the handle is already verified as a resource handle
	parts := strings.Split(handle, "/")
	// resource name is the last element
	if len(parts) < 2 {
		log.Error(":: Error getting resource handle ::", log.Fields{"app": appName, "cluster": cluster, "handle": handle})
		return unstructured.Unstructured{}, err
	}
	resource := parts[len(parts)-2]

	if err != nil {
		// If no reference key, then some early error has occurred in rsync, proceed without reference
		log.Info(":: No reference key for resource yet ::", log.Fields{"appContextId": val})

		// Get Resource from status AppContext
		var res interface{}
		rh, err := sac.GetResourceHandle(ctx, appName, cluster, resource)
		if err != nil {
			log.Error(":: Error getting resource handle ::", log.Fields{"app": appName, "cluster": cluster, "resource": resource})
			// if the resource is not in status appcontext and does not have a reference, then look in the appcontext
			rh, err := ac.GetResourceHandle(ctx, appName, cluster, resource)
			if err != nil {
				log.Error(":: Error getting resource handle ::", log.Fields{"app": appName, "cluster": cluster, "resource": resource})
				return unstructured.Unstructured{}, err
			}
			res, err = ac.GetValue(ctx, rh)
			if err != nil {
				log.Error(":: Error getting resource value ::", log.Fields{"app": appName, "cluster": cluster, "resource": resource})
				return unstructured.Unstructured{}, err
			}
		} else {
			res, err = sac.GetValue(ctx, rh)
			if err != nil {
				log.Error(":: Error getting resource value ::", log.Fields{"app": appName, "cluster": cluster, "resource": resource})
				return unstructured.Unstructured{}, err
			}
		}

		// Get the unstructured object
		unstruct, err := getUnstruct([]byte(res.(string)))
		if err != nil {
			log.Error(":: Error getting GVK ::", log.Fields{"Resource": res, "error": err})
			return unstructured.Unstructured{}, err
		}
		return unstruct, nil
	} else {
		// Load the reference appContext
		ref := appcontext.AppContext{}


		fmt.Println("\n ref: ",ref)
		fmt.Println("val: ",val,"\n")

		_, err = ref.LoadAppContext(ctx, val)
		if err != nil {
			return unstructured.Unstructured{}, err
		}
		// Get Resource from reference AppContext
		rh, err := ref.GetResourceHandle(ctx, appName, cluster, resource)
		if err != nil {
			log.Error(":: Error getting resource handle ::", log.Fields{"app": appName, "cluster": cluster, "resource": resource})
			return unstructured.Unstructured{}, err
		}
		res, err := ref.GetValue(ctx, rh)
		if err != nil {
			log.Error(":: Error getting resource value ::", log.Fields{"app": appName, "cluster": cluster, "resource": resource})
			return unstructured.Unstructured{}, err
		}

		// If ac == sac and ref != ac - then no error, but return empty resource
		acH, _ := ac.GetCompositeAppHandle(ctx)
		sacH, _ := sac.GetCompositeAppHandle(ctx)
		refH, _ := ref.GetCompositeAppHandle(ctx)
		if acH == sacH && refH != acH {
			return unstructured.Unstructured{}, nil
		}
		// Get the unstructured object
		unstruct, err := getUnstruct([]byte(res.(string)))
		if err != nil {
			log.Error(":: Error getting GVK ::", log.Fields{"Resource": res, "error": err})
			return unstructured.Unstructured{}, err
		}
		return unstruct, nil
	}
}

// getAppContextResources collects the resource status of all resources in an AppContext subject to the filter parameters
func getAppContextResources(ctx context.Context, ac, sac appcontext.AppContext, ch interface{}, qOutput, qType string, fResources []string, resourceList *[]ResourceStatus, statusCnts map[string]int, clusterStatusCnts map[string]int, app, cluster string) (int, error) {
	count := 0

	// Get all Resources for the Cluster
	hs, err := ac.GetAllHandles(ctx, ch)
	if err != nil {
		log.Info(":: Error getting all handles ::", log.Fields{"handles": ch, "error": err})
		return 0, err
	}

	for _, h := range hs {
		// skip any handles that are not resource handles
		if !isResourceHandle(ch, h) {
			continue
		}

		// Get the resource handle for the status appcontext
		statusH := getParallelHandle(ctx, h, sac)
		if statusH == "" {
			log.Error(":: Error getting status app context handle::",
				log.Fields{
					"handle":  fmt.Sprintf("%v", h),
					"app":     app,
					"cluster": cluster})
			return 0, pkgerrors.Errorf("Error getting status handle for resource handle: %v", fmt.Sprintf("%v", h))
		}

		// Get Resource Status from status AppContext
		// Default to "Pending" if this key does not yet exist (or any other error occurs)
		rstatus := resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Pending}
		sh, err := sac.GetLevelHandle(ctx, statusH, "status")
		if err == nil {
			s, err := sac.GetValue(ctx, sh)
			if err == nil {
				js, err := json.Marshal(s)
				if err == nil {
					json.Unmarshal(js, &rstatus)
				}
			}
		}
		unstruct, err := getResourceFromReference(ctx, ac, sac, statusH, app, cluster)
		if err != nil {
			log.Error(":: Error getting GVK ::", log.Fields{"error": err})
			continue
		}
		if len(unstruct.Object) == 0 {
			continue
		}
		if !keepResource(unstruct.GetName(), fResources) {
			continue
		}
		// Make and fill out a ResourceStatus structure
		r := ResourceStatus{}
		r.Gvk = unstruct.GroupVersionKind()
		r.Name = unstruct.GetName()
		if qOutput == "detail" {
			r.Detail = unstruct.Object
		}
		if qType == "rsync" { // deprecated
			r.RsyncStatus = fmt.Sprintf("%v", rstatus.Status)
			cnt := statusCnts[rstatus.Status]
			statusCnts[rstatus.Status] = cnt + 1
		} else if qType == "deployed" {
			r.DeployedStatus = fmt.Sprintf("%v", rstatus.Status)
			cnt := statusCnts[rstatus.Status]
			statusCnts[rstatus.Status] = cnt + 1
		} else if qType == "ready" && markAsNotPresent(r.Gvk) {
			r.ReadyStatus = "NotPresent"
			updateNotPresentCount(true, clusterStatusCnts)
		} else if markAsNotPresent(r.Gvk) { // qType is "cluster" - deprecated
			r.ClusterStatus = "NotPresent"
			updateNotPresentCount(true, clusterStatusCnts)
		}
		*resourceList = append(*resourceList, r)
		count++
	}

	return count, nil
}

// getListOfApps gets the list of apps from the app context
func getListOfApps(ctx context.Context, ac appcontext.AppContext) []string {
	ch, err := ac.GetCompositeAppHandle(ctx)

	apps := make([]string, 0)

	// Get all handles
	hs, err := ac.GetAllHandles(ctx, ch)
	if err != nil {
		log.Info(":: Error getting all handles ::", log.Fields{"handles": ch, "error": err})
		return apps
	}

	for _, h := range hs {
		contextHandle := fmt.Sprintf("%v", ch)
		handle := fmt.Sprintf("%v", h)
		diff := strings.Split(handle, contextHandle)

		if len(diff) != 2 && diff[0] != "" {
			continue
		}

		parts := strings.Split(diff[1], "/")

		if len(parts) == 3 && parts[0] == "app" {
			apps = append(apps, parts[1])
		}
	}

	return apps
}

// types of status queries
const ClusterStatusQuery = "clusterStatus"
const DeploymentIntentGroupStatusQuery = "digStatus"
const LcStatusQuery = "lcStatus"
const CaCertQuery = "CaCertStatus"

// PrepareClusterStatusResult takes in a resource stateInfo object, the list of apps and the query parameters.
// It then fills out the StatusResult structure appropriately from information in the AppContext
func PrepareClusterStatusResult(ctx context.Context, stateInfo state.StateInfo, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (ClusterStatusResult, error) {
	status, err := prepareStatusResult(ctx, ClusterStatusQuery, stateInfo, qInstance, qType, qOutput, fApps, fClusters, fResources)
	if err != nil {
		return ClusterStatusResult{}, err
	} else {
		var rval ClusterStatusResult
		// "rsync" and "cluster" variants are deprecated
		if qType == "rsync" || qType == "cluster" {
			rval = ClusterStatusResult{
				Name:          status.Name,
				State:         status.State,
				Status:        status.Status,
				RsyncStatus:   status.RsyncStatus,
				ClusterStatus: status.ClusterStatus,
			}
		} else {
			rval = ClusterStatusResult{
				Name:           status.Name,
				State:          status.State,
				DeployedStatus: status.DeployedStatus,
				ReadyStatus:    status.ReadyStatus,
				DeployedCounts: status.DeployedCounts,
				ReadyCounts:    status.ReadyCounts,
			}
		}
		if len(status.Apps) > 0 && len(status.Apps[0].Clusters) > 0 {
			rval.Cluster = status.Apps[0].Clusters[0]
		}
		return rval, nil
	}
}

// GenericPrepareStatusResult takes in a resource stateInfo object, the list of apps and the query parameters.
// It then fills out the StatusResult structure appropriately from information in the AppContext
func GenericPrepareStatusResult(ctx context.Context, statusType string, stateInfo state.StateInfo, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (StatusResult, error) {
	return prepareStatusResult(ctx, statusType, stateInfo, qInstance, qType, qOutput, fApps, fClusters, fResources)
}

// PrepareStatusResult takes in a resource stateInfo object, the list of apps and the query parameters.
// It then fills out the StatusResult structure appropriately from information in the AppContext
func PrepareStatusResult(ctx context.Context, stateInfo state.StateInfo, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (StatusResult, error) {
	return prepareStatusResult(ctx, DeploymentIntentGroupStatusQuery, stateInfo, qInstance, qType, qOutput, fApps, fClusters, fResources)
}

// covenience fn that ignores the index returned by GetSliceContains
func isNameInList(name string, namesList []string) bool {
	_, ok := utils.GetSliceContains(namesList, name)
	return ok
}

// PrepareCaCertStatusResult takes in a caCert stateInfo object, the list of apps and the query parameters.
// It then fills out the StatusResult structure appropriately from information in the AppContext
func PrepareCaCertStatusResult(ctx context.Context, stateInfo state.StateInfo, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (CaCertStatusResult, error) {
	status, err := prepareStatusResult(ctx, CaCertQuery, stateInfo, qInstance, qType, qOutput, fApps, fClusters, fResources)
	if err != nil {
		return CaCertStatusResult{}, err
	} else {
		rval := CaCertStatusResult{
			Name:           status.Name,
			State:          status.State,
			DeployedStatus: status.DeployedStatus,
			ReadyStatus:    status.ReadyStatus,
			DeployedCounts: status.DeployedCounts,
			ReadyCounts:    status.ReadyCounts,
		}
		//if len(status.Apps) > 0 && len(status.Apps[0].Clusters) > 0 && qOutput != "summary" {
		if len(status.Apps) > 0 {
			rval.Clusters = status.Apps[0].Clusters
		}
		return rval, nil
	}
}

// prepareStatusResult takes in a resource stateInfo object, the list of apps and the query parameters.
// It then fills out the StatusResult structure appropriately from information in the AppContext
func prepareStatusResult(ctx context.Context, statusType string, stateInfo state.StateInfo, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (StatusResult, error) {

	statusResult := StatusResult{}

	statusResult.Apps = make([]AppStatus, 0)
	statusResult.State = stateInfo

	var currentCtxId, statusCtxId string
	if qInstance != "" {
		var err error
		statusCtxId, err = state.GetStatusContextIdForContextId(stateInfo, qInstance)
		if err != nil {
			return StatusResult{}, err
		}
		currentCtxId = qInstance
	} else {
		currentCtxId = state.GetLastContextIdFromStateInfo(stateInfo)
		statusCtxId = state.GetStatusContextIdFromStateInfo(stateInfo)
	}

	// If currentCtxId is still an empty string, an AppContext has not yet been
	// created for this resource.  Just return the statusResult with the stateInfo.
	if currentCtxId == "" {
		return statusResult, nil
	}

	ac, err := state.GetAppContextFromId(ctx, currentCtxId)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "AppContext for status query not found")
	}

	// get the appcontext status value
	h, err := ac.GetCompositeAppHandle(ctx)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "AppContext handle not found")
	}
	sh, err := ac.GetLevelHandle(ctx, h, "status")
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "AppContext status handle not found")
	}
	statusVal, err := ac.GetValue(ctx, sh)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "AppContext status value not found")
	}
	acStatus := appcontext.AppContextStatus{}
	js, err := json.Marshal(statusVal)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "Invalid AppContext status value format")
	}
	err = json.Unmarshal(js, &acStatus)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "Invalid AppContext status value format")
	}

	if qType == "rsync" || qType == "cluster" {
		statusResult.Status = acStatus.Status
	} else {
		statusResult.DeployedStatus = acStatus.Status
	}
	// Get the StatusAppContext
	sac, err := state.GetAppContextFromId(ctx, statusCtxId)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "AppContext for status query not found")
	}

	// Get the composite app meta
	caMeta, err := sac.GetCompositeAppMeta(ctx)

	if statusType != LcStatusQuery && statusType != ClusterStatusQuery && statusType != CaCertQuery {
		if err != nil {
			return StatusResult{}, pkgerrors.Wrap(err, "Error getting CompositeAppMeta")
		}
		if len(caMeta.ChildContextIDs) > 0 {
			// Add the child context IDs to status result
			statusResult.ChildContextIDs = caMeta.ChildContextIDs
		}
	}

	rsyncStatusCnts := make(map[string]int)
	clusterStatusCnts := make(map[string]int)

	// Get the list of apps from the app context
	apps := getListOfApps(ctx, ac)

	// If filter-apps list is provided, ensure that every app to be
	// filtered is part of this composite app
	for _, fApp := range fApps {
		if !isNameInList(fApp, apps) {
			return StatusResult{},
				fmt.Errorf("Filter app %s not in list of apps for composite app %s",
					fApp, caMeta.CompositeApp)
		}
	}

	// Loop through each app and get the status data for each cluster in the app
	for _, app := range apps {
		appCount := 0
		if len(fApps) > 0 && !isNameInList(app, fApps) {
			continue
		}
		// Get the clusters in the appcontext for this app
		clusters, err := ac.GetClusterNames(ctx, app)
		if err != nil {
			continue
		}
		var appStatus AppStatus
		appStatus.Name = app
		appStatus.Clusters = make([]ClusterStatus, 0)

		for _, cluster := range clusters {
			clusterCount := 0
			if len(fClusters) > 0 && !isNameInList(cluster, fClusters) {
				continue
			}

			var clusterStatus ClusterStatus
			pc := strings.Split(cluster, "+")
			clusterStatus.ClusterProvider = pc[0]
			clusterStatus.Cluster = pc[1]
			connectivity := getClusterReadyStatus(ctx, sac, app, cluster)
			if connectivity == no_cluster_handle {
				// try the current appcontext in case rsync hasn't updated status appcontext
				connectivity = getClusterReadyStatus(ctx, ac, app, cluster)
				if connectivity == no_cluster_handle {
					connectivity = string(appcontext.ClusterReadyStatusEnum.Unknown)
				}
			}
			if qType == "rsync" || qType == "cluster" {
				clusterStatus.ReadyStatus = connectivity
			} else {
				clusterStatus.Connectivity = connectivity
			}

			ch, err := ac.GetClusterHandle(ctx, app, cluster)
			if err != nil {
				log.Error(":: No handle for cluster, app ::",
					log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
				continue
			}

			clusterStatus.Resources = make([]ResourceStatus, 0)
			cnt, err := getAppContextResources(ctx, ac, sac, ch, qOutput, qType, fResources, &clusterStatus.Resources, rsyncStatusCnts, clusterStatusCnts, app, cluster)
			if err != nil {
				log.Info(":: Error gathering appcontext resources for cluster, app ::",
					log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
				continue
			}
			log.Info(":: Found some rsync appcontext resources ", log.Fields{"cnt": cnt, "app": app, "cluster": cluster})
			appCount += cnt
			clusterCount += cnt

			if qType == "cluster" || qType == "ready" {
				rbValue, err := getResourceBundleStateStatus(ctx, sac, app, cluster)
				if err != nil {
					continue
				}

				cnt, err := getClusterResources(rbValue, qType, qOutput, fResources, &clusterStatus.Resources, clusterStatusCnts)
				if err != nil {
					log.Info(":: Error gathering cluster resources for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}
				log.Info(":: Found cluster status appcontext resources ", log.Fields{"cnt": cnt, "app": app, "cluster": cluster})

				appCount += cnt
				clusterCount += cnt
			} else if qType != "rsync" && qType != "deployed" {
				log.Info(":: Invalid status type ::", log.Fields{"Status Type": qType})
				continue
			}

			if clusterCount > 0 {
				appStatus.Clusters = append(appStatus.Clusters, clusterStatus)
			}
		}
		if appCount > 0 && qOutput != "summary" {
			statusResult.Apps = append(statusResult.Apps, appStatus)
		}
	}

	if qType == "rsync" || qType == "cluster" {
		statusResult.RsyncStatus = rsyncStatusCnts
		statusResult.ClusterStatus = clusterStatusCnts
	} else {
		statusResult.DeployedCounts = rsyncStatusCnts
		statusResult.ReadyCounts = clusterStatusCnts
	}

	if cnt, ok := clusterStatusCnts["NotPresent"]; ok && cnt > 0 {
		statusResult.ReadyStatus = "NotReady"
	} else if cnt, ok := clusterStatusCnts["NotReady"]; ok && cnt > 0 {
		statusResult.ReadyStatus = "NotReady"
	} else if qType != "rsync" && qType != "deployed" {
		statusResult.ReadyStatus = "Ready"
	}

	return statusResult, nil
}

func getResourceBundleStateStatus(ctx context.Context, ac appcontext.AppContext, app, cluster string) (rb.ResourceBundleStateStatus, error) {
	csh, err := ac.GetClusterStatusHandle(ctx, app, cluster)
	if err != nil {
		log.Info(":: No cluster status handle for cluster, app ::",
			log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
		// cluster status has not been created yet, return an empty status value
		return rb.ResourceBundleStateStatus{}, nil
	}
	clusterRbValue, err := ac.GetValue(ctx, csh)
	if err != nil {
		log.Info(":: No cluster status value for cluster, app ::",
			log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
		return rb.ResourceBundleStateStatus{}, err
	}
	var rbValue rb.ResourceBundleStateStatus
	err = json.Unmarshal([]byte(clusterRbValue.(string)), &rbValue)
	if err != nil {
		log.Error(":: Error unmarshalling cluster status value for cluster, app ::",
			log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
		return rb.ResourceBundleStateStatus{}, err
	}

	return rbValue, nil
}

// PrepareAppsListStatusResult takes in a resource stateInfo object, the list of apps and the query parameters.
// It then fills out the StatusResult structure appropriately from information in the AppContext
func PrepareAppsListStatusResult(ctx context.Context, stateInfo state.StateInfo, qInstance string) (AppsListResult, error) {
	statusResult := AppsListResult{}

	var currentCtxId string
	if qInstance != "" {
		// verifies that qInstance is a valid context id
		_, err := state.GetStatusContextIdForContextId(stateInfo, qInstance)
		if err != nil {
			return statusResult, err
		}
		currentCtxId = qInstance
	} else {
		currentCtxId = state.GetLastContextIdFromStateInfo(stateInfo)
	}

	// If currentCtxId is still an empty string, an AppContext has not yet been
	// created for this resource.  So, no Apps list can be returned from the AppContext.
	if currentCtxId == "" {
		statusResult.Apps = make([]string, 0)
		return statusResult, nil
	}

	ac, err := state.GetAppContextFromId(ctx, currentCtxId)
	if err != nil {
		return AppsListResult{}, pkgerrors.Wrap(err, "AppContext for Apps List status query not found")
	}

	// Get the list of apps from the app context
	statusResult.Apps = getListOfApps(ctx, ac)

	return statusResult, nil
}

// prepareStatusResult takes in a resource stateInfo object, the list of apps and the query parameters.
// It then fills out the StatusResult structure appropriately from information in the AppContext
func PrepareClustersByAppStatusResult(ctx context.Context, stateInfo state.StateInfo, qInstance string, fApps []string) (ClustersByAppResult, error) {
	statusResult := ClustersByAppResult{}

	statusResult.ClustersByApp = make([]ClustersByAppEntry, 0)

	var currentCtxId string
	if qInstance != "" {
		// verifies that qInstance is a valid context id
		_, err := state.GetStatusContextIdForContextId(stateInfo, qInstance)
		if err != nil {
			return statusResult, err
		}
		currentCtxId = qInstance
	} else {
		currentCtxId = state.GetLastContextIdFromStateInfo(stateInfo)
	}

	// If currentCtxId is still an empty string, an AppContext has not yet been
	// created for this resource.  Just return the statusResult with the stateInfo.
	if currentCtxId == "" {
		return statusResult, nil
	}

	ac, err := state.GetAppContextFromId(ctx, currentCtxId)
	if err != nil {
		return ClustersByAppResult{}, pkgerrors.Wrap(err, "AppContext for status query not found")
	}

	// Get the list of apps from the app context
	apps := getListOfApps(ctx, ac)

	// Loop through each app and get the clusters for the app
	for _, app := range apps {
		// apply app filter if provided
		if len(fApps) > 0 {
			found := false
			for _, a := range fApps {
				if a == app {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		// add app to output structure
		entry := ClustersByAppEntry{
			App: app,
		}

		// Get the clusters in the appcontext for this app
		entry.Clusters = make([]ClusterEntry, 0)
		clusters, err := ac.GetClusterNames(ctx, app)
		if err != nil {
		} else {
			for _, cl := range clusters {
				pc := strings.Split(cl, "+")
				entry.Clusters = append(entry.Clusters, ClusterEntry{
					ClusterProvider: pc[0],
					Cluster:         pc[1],
				})
			}
		}

		statusResult.ClustersByApp = append(statusResult.ClustersByApp, entry)
	}

	return statusResult, nil
}

// PrepareResourcesByAppStatusResult takes in a resource stateInfo object, the list of apps and the query parameters.
// It then fills out the ResourcesByAppStatusResult structure appropriately from information in the AppContext
func PrepareResourcesByAppStatusResult(ctx context.Context, stateInfo state.StateInfo, qInstance, qType string, fApps, fClusters []string) (ResourcesByAppResult, error) {

	var currentCtxId, statusCtxId string
	if qInstance != "" {
		var err error
		statusCtxId, err = state.GetStatusContextIdForContextId(stateInfo, qInstance)
		if err != nil {
			statusResult := ResourcesByAppResult{}
			statusResult.ResourcesByApp = make([]ResourcesByAppEntry, 0)
			return statusResult, err
		}
		currentCtxId = qInstance
	} else {
		currentCtxId = state.GetLastContextIdFromStateInfo(stateInfo)
		statusCtxId = state.GetStatusContextIdFromStateInfo(stateInfo)
	}

	// If currentCtxId is still an empty string, an AppContext has not yet been
	// created for this resource.  Just an empty status result
	if currentCtxId == "" {
		statusResult := ResourcesByAppResult{}
		statusResult.ResourcesByApp = make([]ResourcesByAppEntry, 0)
		return statusResult, nil
	}

	ac, err := state.GetAppContextFromId(ctx, currentCtxId)
	if err != nil {
		return ResourcesByAppResult{}, pkgerrors.Wrap(err, "AppContext for status query not found")
	}

	sac, err := state.GetAppContextFromId(ctx, statusCtxId)
	if err != nil {
		return ResourcesByAppResult{}, pkgerrors.Wrap(err, "Status AppContext for status query not found")
	}

	return prepareResourcesByAppStatusResult(ctx, ac, sac, qType, fApps, fClusters)
}

func prepareResourcesByAppStatusResult(ctx context.Context, ac, sac appcontext.AppContext, qType string, fApps, fClusters []string) (ResourcesByAppResult, error) {
	statusResult := ResourcesByAppResult{}
	statusResult.ResourcesByApp = make([]ResourcesByAppEntry, 0)

	// Get the list of apps from the app context
	apps := getListOfApps(ctx, ac)

	// Loop through each app and get the status data for each cluster in the app
	for _, app := range apps {
		if len(fApps) > 0 {
			found := false
			for _, a := range fApps {
				if a == app {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Get the clusters in the appcontext for this app
		clusters, err := ac.GetClusterNames(ctx, app)
		if err != nil {
			continue
		}
		var appStatus AppStatus
		appStatus.Name = app
		appStatus.Clusters = make([]ClusterStatus, 0)

		for _, cluster := range clusters {

			// apply the cluster filter
			if len(fClusters) > 0 {
				found := false
				for _, c := range fClusters {
					if c == cluster {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			rsyncStatusCnts := make(map[string]int)
			clusterStatusCnts := make(map[string]int)

			pc := strings.Split(cluster, "+")
			resourcesByAppEntry := ResourcesByAppEntry{
				ClusterProvider: pc[0],
				Cluster:         pc[1],
				App:             app,
			}
			resourcesByAppEntry.Resources = make([]ResourceEntry, 0)

			ch, err := ac.GetClusterHandle(ctx, app, cluster)
			if err != nil {
				log.Error(":: No handle for cluster, app ::",
					log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
				continue
			}

			resources := make([]ResourceStatus, 0)
			// Get all resources from the appcontext for the given app/cluster
			_, err = getAppContextResources(ctx, ac, sac, ch, "all", qType, make([]string, 0), &resources, rsyncStatusCnts, clusterStatusCnts, app, cluster)
			if err != nil {
				log.Info(":: Error gathering appcontext resources for cluster, app ::",
					log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
				continue
			}

			// find any additional resources on the cluster
			if qType == "cluster" || qType == "ready" {
				rbValue, err := getResourceBundleStateStatus(ctx, sac, app, cluster)
				if err != nil {
					continue
				}

				_, err = getClusterResources(rbValue, qType, "all", make([]string, 0), &resources, clusterStatusCnts)
				if err != nil {
					log.Info(":: Error gathering cluster resources for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}
			} else if qType != "rsync" && qType != "deployed" {
				log.Info(":: Invalid status type ::", log.Fields{"Status Type": qType})
				continue
			}

			for _, r := range resources {
				resourcesByAppEntry.Resources = append(resourcesByAppEntry.Resources, ResourceEntry{
					Name: r.Name,
					Gvk:  r.Gvk,
				})
			}
			statusResult.ResourcesByApp = append(statusResult.ResourcesByApp, resourcesByAppEntry)
		}
	}

	return statusResult, nil
}

const no_cluster_handle = "cluster handle not found"

// Read readystatus from reference
func getClusterReadyStatus(ctx context.Context, ac appcontext.AppContext, app, cluster string) string {

	// the appcontext here is the status appcontext - follow the reference to get the readystatus
	ch, err := ac.GetClusterHandle(ctx, app, cluster)
	if err != nil {
		log.Error("Cluster handle not found", log.Fields{"cluster": cluster})
		// return special string to trigger requery with current appcontext
		return no_cluster_handle
	}
	var val = ""
	// Read reference appContext value
	sh, err := ac.GetLevelHandle(ctx, ch, "reference")
	if err == nil {
		s, err := ac.GetValue(ctx, sh)
		if err == nil {
			js, err := json.Marshal(s)
			if err == nil {
				json.Unmarshal(js, &val)
			}
		}
	}
	if err != nil {
		// Cluster may be unreachable, and reference key not created yet
		log.Info("Reference not found for cluster status", log.Fields{"cluster": cluster})

		rsh, _ := ac.GetLevelHandle(ctx, ch, "readystatus")

		if rsh != nil {
			status, err := ac.GetValue(ctx, rsh)
			if err != nil {
				log.Error("Error getting cluster readystatus", log.Fields{"cluster": cluster, "error": err})
				return string(appcontext.ClusterReadyStatusEnum.Unknown)
			}
			return status.(string)
		}
		return string(appcontext.ClusterReadyStatusEnum.Unknown)
	} else {
		// Load the reference appContext
		ref := appcontext.AppContext{}
		_, err = ref.LoadAppContext(ctx, val)
		if err != nil {
			log.Error(":: Error loading the app context::", log.Fields{"appContextId": val, "error": err})
			return string(appcontext.ClusterReadyStatusEnum.Unknown)
		}
		rlh, err := ref.GetClusterHandle(ctx, app, cluster)
		if err != nil {
			log.Error("Error getting cluster handle for Reference", log.Fields{"cluster": cluster, "error": err})
			return string(appcontext.ClusterReadyStatusEnum.Unknown)
		}
		rsh, err := ref.GetLevelHandle(ctx, rlh, "readystatus")

		if rsh != nil {
			status, err := ref.GetValue(ctx, rsh)
			if err != nil {
				log.Error("Error getting readystatus from Reference", log.Fields{"cluster": cluster, "error": err})
				return string(appcontext.ClusterReadyStatusEnum.Unknown)
			}
			return status.(string)
		}
		return string(appcontext.ClusterReadyStatusEnum.Unknown)
	}
}
