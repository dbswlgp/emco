// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/resourcestatus"
)

type AppContextReference struct {
	acID string
	ac   appcontext.AppContext
}

func NewAppContextReference(ctx context.Context, acID string) (AppContextReference, error) {
	ac := appcontext.AppContext{}
	if len(acID) == 0 {
		log.Error("Error loading AppContext - appContexID is nil", log.Fields{})
		return AppContextReference{}, pkgerrors.Errorf("appContexID is nil")
	}
	_, err := ac.LoadAppContext(ctx, acID)
	if err != nil {
		log.Error("Error loading AppContext", log.Fields{"err": err, "acID": acID})
		return AppContextReference{}, err
	}
	return AppContextReference{ac: ac, acID: acID}, nil
}
func (a *AppContextReference) GetAppContextHandle() appcontext.AppContext {
	return a.ac
}

//GetAppContextFlag gets the stop flag
func (a *AppContextReference) GetAppContextFlag(ctx context.Context, key string) (bool, error) {
	h, err := a.ac.GetCompositeAppHandle(ctx)
	if err != nil {
		// Treat an error as stop
		return true, err
	}
	sh, err := a.ac.GetLevelHandle(ctx, h, key)
	if sh != nil {
		if v, err := a.ac.GetValue(ctx, sh); err == nil {
			return v.(bool), nil
		}
	}
	return true, err
}

//UpdateAppContextFlag to update flags
func (a *AppContextReference) UpdateAppContextFlag(ctx context.Context, key string, b bool) error {
	h, err := a.ac.GetCompositeAppHandle(ctx)
	if err != nil {
		log.Error("Error UpdateAppContextFlag", log.Fields{"err": err})
		return err
	}
	sh, err := a.ac.GetLevelHandle(ctx, h, key)
	if sh == nil {
		_, err = a.ac.AddLevelValue(ctx, h, key, b)
	} else {
		err = a.ac.UpdateValue(ctx, sh, b)
	}
	if err != nil {
		log.Error("Error UpdateAppContextFlag", log.Fields{"err": err})
	}
	return err

}

//UpdateAppContextStatus updates a field in AppContext
func (a *AppContextReference) UpdateAppContextStatus(ctx context.Context, key string, status interface{}) error {
	//var acStatus appcontext.AppContextStatus = appcontext.AppContextStatus{}
	hc, err := a.ac.GetCompositeAppHandle(ctx)
	if err != nil {
		log.Error("Error UpdateAppContextStatus", log.Fields{"err": err})
		return err
	}
	dsh, err := a.ac.GetLevelHandle(ctx, hc, key)
	if dsh == nil {
		_, err = a.ac.AddLevelValue(ctx, hc, key, status)
	} else {
		err = a.ac.UpdateValue(ctx, dsh, status)
	}
	if err != nil {
		log.Error("Error UpdateAppContextStatus", log.Fields{"err": err})
	}
	return err

}

//GetAppContextStatus gets the status
func (a *AppContextReference) GetAppContextStatus(ctx context.Context, key string) (appcontext.AppContextStatus, error) {
	var acStatus appcontext.AppContextStatus = appcontext.AppContextStatus{}

	hc, err := a.ac.GetCompositeAppHandle(ctx)
	if err != nil {
		log.Error("Error GetAppContextStatus", log.Fields{"err": err})
		return acStatus, err
	}
	dsh, err := a.ac.GetLevelHandle(ctx, hc, key)
	if dsh != nil {
		v, err := a.ac.GetValue(ctx, dsh)
		if err != nil {
			log.Error("Error GetAppContextStatus", log.Fields{"err": err})
			return acStatus, err
		}
		//s := fmt.Sprintf("%v", v)
		//acStatus.Status = appcontext.StatusValue(s)
		acStatus = appcontext.AppContextStatus{}
		js, err := json.Marshal(v)
		if err != nil {
			log.Error("Error GetAppContextStatus", log.Fields{"err": err})
			return acStatus, err
		}
		err = json.Unmarshal(js, &acStatus)
		if err != nil {
			log.Error("Error GetAppContextStatus", log.Fields{"err": err})
			return acStatus, err
		}
	}
	return acStatus, err
}

// SetClusterAvailableStatus sets the cluster available status
func (a *AppContextReference) SetClusterAvailableStatus(ctx context.Context, app, cluster string, status appcontext.StatusValue) {
	ch, err := a.ac.GetClusterHandle(ctx, app, cluster)
	if err != nil {
		return
	}
	rsh, _ := a.ac.GetLevelHandle(ctx, ch, "readystatus")
	// If readystatus handle was not found, then create it
	if rsh == nil {
		a.ac.AddLevelValue(ctx, ch, "readystatus", status)
	} else {
		a.ac.UpdateStatusValue(ctx, rsh, status)
	}
}

// GetClusterAvailableStatus sets the cluster ready status
// does not return an error, just a status of Unknown if the cluster readystatus key does
// not exist or any other error occurs.
func (a *AppContextReference) GetClusterAvailableStatus(ctx context.Context, app, cluster string) appcontext.StatusValue {
	ch, err := a.ac.GetClusterHandle(ctx, app, cluster)
	if err != nil {
		return appcontext.ClusterReadyStatusEnum.Unknown
	}
	rsh, _ := a.ac.GetLevelHandle(ctx, ch, "readystatus")
	if rsh != nil {
		status, err := a.ac.GetValue(ctx, rsh)
		if err != nil {
			return appcontext.ClusterReadyStatusEnum.Unknown
		}
		return status.(appcontext.StatusValue)
	}

	return appcontext.ClusterReadyStatusEnum.Unknown
}

// GetRes Reads resource
func (a *AppContextReference) GetRes(ctx context.Context, name string, app string, cluster string) ([]byte, interface{}, error) {
	var byteRes []byte

	rh, err := a.ac.GetResourceHandle(ctx, app, cluster, name)
	if err != nil {
		log.Error("Error GetRes", log.Fields{"err": err})
		return nil, nil, err
	}
	sh, err := a.ac.GetLevelHandle(ctx, rh, "status")
	if err != nil {
		statusPending := resourcestatus.ResourceStatus{
			Status: resourcestatus.RsyncStatusEnum.Pending,
		}
		sh, err = a.ac.AddLevelValue(ctx, rh, "status", statusPending)
		if err != nil {
			log.Error("Error GetRes", log.Fields{"err": err})
			return nil, nil, err
		}
	}
	resval, err := a.ac.GetValue(ctx, rh)
	if err != nil {
		log.Error("Error GetRes", log.Fields{"err": err})
		return nil, sh, err
	}
	if resval != "" {
		result := strings.Split(name, "+")
		if result[0] == "" {
			log.Error("Error GetRes, Resource name is nil", log.Fields{})
			return nil, sh, pkgerrors.Errorf("Resource name is nil %s:", name)
		}
		byteRes = []byte(fmt.Sprintf("%v", resval.(interface{})))
	} else {
		log.Error("Error GetRes, Resource name is nil", log.Fields{})
		return nil, sh, pkgerrors.Errorf("Resource value is nil %s", name)
	}
	return byteRes, sh, nil
}

// INFO: the function below always returns level=0 as it stands
//GetNamespace reads namespace from metadata
func (a *AppContextReference) GetNamespace(ctx context.Context) (string, string) {
	namespace := "default"
	level := "0"
	appmeta, err := a.ac.GetCompositeAppMeta(ctx)
	if err == nil {
		namespace = appmeta.Namespace
		level = appmeta.Level
	}
	log.Info("CloudConfig for this app will be looked up using level and namespace specified", log.Fields{
		"level":     level,
		"namespace": namespace,
	})
	return namespace, level
}

//GetLogicalCloudInfo reads logical cloud related info from metadata
func (a *AppContextReference) GetLogicalCloudInfo(ctx context.Context) (string, string, string, string, string, error) {

	appmeta, err := a.ac.GetCompositeAppMeta(ctx)
	if err != nil {
		log.Error("Error GetLogicalCloudInfo", log.Fields{"err": err})
		return "", "", "", "", "", err
	}
	return appmeta.Project, appmeta.LogicalCloud, appmeta.Level, appmeta.LogicalCloudNamespace, appmeta.LogicalCloudLevel, nil
}

// PutRes copies resource into appContext
func (a *AppContextReference) PutRes(ctx context.Context, name string, app string, cluster string, data []byte) error {
	rh, err := a.ac.GetResourceHandle(ctx, app, cluster, name)
	if err != nil {
		log.Error("Error GetResourceHandle", log.Fields{"err": err})
		return err
	}
	handle, _ := a.ac.GetLevelHandle(ctx, rh, "definition")
	// If definition handle was not found, then create it
	if handle == nil {
		a.ac.AddLevelValue(ctx, rh, "definition", string(data))
	} else {
		a.ac.UpdateStatusValue(ctx, handle, string(data))
	}
	return nil
}

//GetAppContextFlag gets the statusappctxid
func (a *AppContextReference) GetStatusAppContext(ctx context.Context, key string) (string, error) {
	h, err := a.ac.GetCompositeAppHandle(ctx)
	if err != nil {
		log.Error("Error GetAppContextFlag", log.Fields{"err": err})
		return "", err
	}
	sh, err := a.ac.GetLevelHandle(ctx, h, key)
	if sh != nil {
		if v, err := a.ac.GetValue(ctx, sh); err == nil {
			return v.(string), nil
		}
	}
	return "", err
}

// Add resource level for a status
// Function adds any missing levels to AppContext
func (a *AppContextReference) AddResourceStatus(ctx context.Context, name string, app string, cluster string, status interface{}, acID string) error {
	var rh, ch, ah interface{}

	rh, err := a.ac.GetResourceHandle(ctx, app, cluster, name)
	if err != nil {
		// Assume the resource doesn't exist
		h, err := a.ac.GetCompositeAppHandle(ctx)
		if err != nil {
			log.Error("Composite App Handle not found", log.Fields{"err": err})
			return err
		}
		// Check if App exists if not add handle
		ah, err = a.ac.GetAppHandle(ctx, app)
		if err != nil {
			//Add App level
			ah, err = a.ac.AddApp(ctx, h, app)
			if err != nil {
				log.Error("Unable to add application to context for status", log.Fields{"err": err})
				return err
			}
		}
		ch, err = a.ac.GetClusterHandle(ctx, app, cluster)
		if err != nil {
			ch, err = a.ac.AddCluster(ctx, ah, cluster)
			if err != nil {
				log.Error("Unable to add cluster to context for status", log.Fields{"err": err})
				return err
			}
		}
		rh, err = a.ac.AddResource(ctx, ch, name, "nil")
		if err != nil {
			log.Error("Unable to add resource to context for status", log.Fields{"err": err})
			return err
		}
	}
	sh, err := a.ac.GetLevelHandle(ctx, rh, "status")
	if err != nil {
		sh, err = a.ac.AddLevelValue(ctx, rh, "status", status)
		if err != nil {
			log.Error("Error add status to resource", log.Fields{"err": err})
			return err
		}
	} else {
		a.ac.UpdateStatusValue(ctx, sh, status)
	}
	// Create link to the original resource
	link := acID
	lh, err := a.ac.GetLevelHandle(ctx, rh, "reference")
	if err != nil {
		lh, err = a.ac.AddLevelValue(ctx, rh, "reference", link)
		if err != nil {
			log.Error("Error add reference to resource for status", log.Fields{"err": err})
			return err
		}
	} else {
		a.ac.UpdateStatusValue(ctx, lh, link)
	}
	// Create a link to new appContext at the cluster level also for readystatus
	ch, err = a.ac.GetClusterHandle(ctx, app, cluster)
	if err != nil {
		return err
	}
	lch, err := a.ac.GetLevelHandle(ctx, ch, "reference")
	if err != nil {
		lch, err = a.ac.AddLevelValue(ctx, ch, "reference", link)
		if err != nil {
			log.Error("Error add reference to resource for status", log.Fields{"err": err})
			return err
		}
	} else {
		a.ac.UpdateStatusValue(ctx, lch, link)
	}
	return nil
}

// SetClusterResourceReady sets the cluster ready status
func (a *AppContextReference) SetClusterResourcesReady(ctx context.Context, app, cluster string, value bool) error {
	ch, err := a.ac.GetClusterHandle(ctx, app, cluster)
	if err != nil {
		return err
	}
	rsh, _ := a.ac.GetLevelHandle(ctx, ch, "resourcesready")
	// If resource ready handle was not found, then create it
	if rsh == nil {
		a.ac.AddLevelValue(ctx, ch, "resourcesready", value)
	} else {
		a.ac.UpdateStatusValue(ctx, rsh, value)
	}
	return nil
}

// GetClusterResourceReady gets the cluster ready status
func (a *AppContextReference) GetClusterResourcesReady(ctx context.Context, app, cluster string) bool {
	ch, err := a.ac.GetClusterHandle(ctx, app, cluster)
	if err != nil {
		return false
	}
	rsh, _ := a.ac.GetLevelHandle(ctx, ch, "resourcesready")
	if rsh != nil {
		status, err := a.ac.GetValue(ctx, rsh)
		if err != nil {
			return false
		}
		return status.(bool)
	}
	return false
}

// SetResourceReadyStatus sets the resource ready status
func (a *AppContextReference) SetResourceReadyStatus(ctx context.Context, app, cluster, res string, readyType string, value bool) error {
	rh, err := a.ac.GetResourceHandle(ctx, app, cluster, res)
	if err != nil {
		return err
	}
	rsh, _ := a.ac.GetLevelHandle(ctx, rh, string(readyType))
	// If resource ready handle was not found, then create it
	if rsh == nil {
		a.ac.AddLevelValue(ctx, rh, string(readyType), value)
	} else {
		a.ac.UpdateStatusValue(ctx, rsh, value)
	}
	return nil
}

// GetClusterResourceReady gets the resources ready status
func (a *AppContextReference) GetResourceReadyStatus(ctx context.Context, app, cluster, res string, readyType string) bool {
	rh, err := a.ac.GetResourceHandle(ctx, app, cluster, res)
	if err != nil {
		return false
	}
	rsh, _ := a.ac.GetLevelHandle(ctx, rh, string(readyType))
	if rsh != nil {
		status, err := a.ac.GetValue(ctx, rsh)
		if err != nil {
			return false
		}
		return status.(bool)
	}
	return false
}

// CheckAppReadyOnAllClusters checks if App is ready on all clusters
func (a *AppContextReference) CheckAppReadyOnAllClusters(ctx context.Context, app string) bool {
	// Check if all the clusters are ready
	cl, err := a.ac.GetClusterNames(ctx, app)
	if err != nil {
		return false
	}
	for _, cn := range cl {
		if !a.GetClusterResourcesReady(ctx, app, cn) {
			// Some cluster is not ready
			return false
		}
	}
	return true
}

func (a *AppContextReference) GetSubResApprove(ctx context.Context, name, app, cluster string) ([]byte, interface{}, error) {
	var byteRes []byte

	rh, err := a.ac.GetResourceHandle(ctx, app, cluster, name)
	if err != nil {
		log.Error("GetSubResApprove - Error getting resource handle", log.Fields{"name": name, "cluster": cluster, "app": app, "error": err})
		return nil, nil, err
	}

	// Look up the subresource approval by following the reference
	var val = ""
	refh, err := a.ac.GetLevelHandle(ctx, rh, "reference")
	if err == nil {
		s, err := a.ac.GetValue(ctx, refh)
		if err == nil {
			js, err := json.Marshal(s)
			if err == nil {
				json.Unmarshal(js, &val)
			}
		}
	}
	if err != nil {
		log.Error("GetSubResApprove - Error getting reference handle and value", log.Fields{"name": name, "cluster": cluster, "app": app, "error": err})
		return nil, nil, err
	}

	// Load the reference appContext
	ref := appcontext.AppContext{}
	_, err = ref.LoadAppContext(ctx, val)
	if err != nil {
		log.Error(":: Error loading the referenced app context::", log.Fields{"reference Cid": val, "error": err})
		return nil, nil, err
	}

	// get referenced resource handle
	rh, err = ref.GetResourceHandle(ctx, app, cluster, name)
	if err != nil {
		return nil, nil, err
	}

	// Check if Subresource defined
	sh, err := ref.GetLevelHandle(ctx, rh, "subresource/approval")
	if err != nil {
		log.Error("GetSubResApprove - Error getting referenced subresource/approval handle", log.Fields{"resource handle": rh, "name": name, "cluster": cluster, "app": app, "error": err})
		return nil, nil, err
	}
	resval, err := ref.GetValue(ctx, sh)
	if err != nil {
		log.Error("GetSubResApprove - Error getting referenced subresource/approval value", log.Fields{"subresource handle": sh, "name": name, "cluster": cluster, "app": app, "error": err})
		return nil, sh, err
	}
	if resval != "" {
		byteRes = []byte(fmt.Sprintf("%v", resval.(interface{})))
	} else {
		log.Error("Error GetSubResApprove, Resource name is nil", log.Fields{})
		return nil, sh, pkgerrors.Errorf("SubResource value is nil %s", name)
	}
	return byteRes, sh, nil
}
