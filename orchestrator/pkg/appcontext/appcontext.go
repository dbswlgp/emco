// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package appcontext

import (
	"context"
	"fmt"
	"strings"

	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/rtcontext"
)

// metaPrefix used for denoting clusterMeta level
const metaGrpPREFIX = "!@#metaGrp"

// Separator is a constant used to create concatenated names for items like cluster or resource names
const Separator = "+"

// OrderInstruction type constants
const OrderInstruction = "order"

// DependencyInstruction type constants
const DependencyInstruction = "dependency"

// Level constant names
const ResourceLevel = "resource"
const AppLevel = "app"

type AppContext struct {
	initDone bool
	rtcObj   rtcontext.RunTimeContext
	rtc      rtcontext.Rtcontext
}

// AppContextStatus represents the current status of the appcontext
//	Instantiating - instantiate has been invoked and is still in progress
//	Instantiated - instantiate has completed
//	Terminating - terminate has been invoked and is still in progress
//	Terminated - terminate has completed
//	InstantiateFailed - the instantiate action has failed
//	TerminateFailed - the terminate action has failed
type AppContextStatus struct {
	Status StatusValue
}
type StatusValue string
type statuses struct {
	Instantiating     StatusValue
	Instantiated      StatusValue
	Terminating       StatusValue
	Terminated        StatusValue
	InstantiateFailed StatusValue
	TerminateFailed   StatusValue
	Created           StatusValue
	Updating          StatusValue
	Updated           StatusValue
	UpdateFailed      StatusValue
}

var AppContextStatusEnum = &statuses{
	Instantiating:     "Instantiating",
	Instantiated:      "Instantiated",
	Terminating:       "Terminating",
	Terminated:        "Terminated",
	InstantiateFailed: "InstantiateFailed",
	TerminateFailed:   "TerminateFailed",
	Created:           "Created",
	Updating:          "Updating",
	Updated:           "Updated",
	UpdateFailed:      "UpdatedFailed",
}

type clusterStatuses struct {
	Unknown   StatusValue
	Available StatusValue
	Retrying  StatusValue
}

var ClusterReadyStatusEnum = &clusterStatuses{
	Unknown:   "Unknown",
	Available: "Available",
	Retrying:  "Retrying",
}

// CompositeAppMeta contains all the possible attributes an
// appcontext /meta handle of a Composite App or Logical Cloud, may have.
// Note: only some of these fields will be used in each for each of the types above:
type CompositeAppMeta struct {
	Project               string   `json:"Project"`
	CompositeApp          string   `json:"CompositeApp"`
	Version               string   `json:"Version"`
	Release               string   `json:"Release"`
	DeploymentIntentGroup string   `json:"DeploymentIntentGroup"`
	Namespace             string   `json:"Namespace"`
	Level                 string   `json:"Level"`
	ChildContextIDs       []string `json:"ChildContextIDs"`
	LogicalCloud          string   `json:"LogicalCloud"`
	LogicalCloudNamespace string   `json:"LogicalCloudNamespace"`
	LogicalCloudLevel     string   `json:"LogicalCloudLevel"`
}

// Init app context
func (ac *AppContext) InitAppContext() (interface{}, error) {
	ac.rtcObj = rtcontext.RunTimeContext{}
	ac.rtc = &ac.rtcObj
	return ac.rtc.RtcInit()
}

// Init app context
func (ac *AppContext) InitAppContextWithValue(cid interface{}) (interface{}, error) {
	ac.rtcObj = rtcontext.RunTimeContext{}
	ac.rtc = &ac.rtcObj
	return ac.rtc.RtcInitWithValue(cid)
}

// Load app context that was previously created
func (ac *AppContext) LoadAppContext(ctx context.Context, cid interface{}) (interface{}, error) {
	ac.rtcObj = rtcontext.RunTimeContext{}
	ac.rtc = &ac.rtcObj
	return ac.rtc.RtcLoad(ctx, cid)
}

// CreateCompositeApp method returns composite app handle as interface.
func (ac *AppContext) CreateCompositeApp(ctx context.Context) (interface{}, error) {
	h, err := ac.rtc.RtcCreate(ctx)
	if err != nil {
		return nil, err
	}
	log.Info(":: CreateCompositeApp ::", log.Fields{"CompositeAppHandle": h})
	return h, nil
}

// AddCompositeAppMeta adds the meta data associated with a composite app
func (ac *AppContext) AddCompositeAppMeta(ctx context.Context, meta interface{}) error {
	err := ac.rtc.RtcAddMeta(ctx, meta)
	if err != nil {
		return err
	}
	return nil
}

// Deletes the entire context
func (ac *AppContext) DeleteCompositeApp(ctx context.Context) error {
	h, err := ac.rtc.RtcGet(ctx)

	if err != nil {
		return err
	}
	err = ac.rtc.RtcDeletePrefix(ctx, h)
	if err != nil {
		return err
	}
	return nil
}

//Returns the handles for a given composite app context
func (ac *AppContext) GetCompositeAppHandle(ctx context.Context) (interface{}, error) {
	h, err := ac.rtc.RtcGet(ctx)
	if err != nil {
		return nil, err
	}
	return h, nil
}

// GetLevelHandle returns the handle for the supplied level at the given handle.
// For example, to get the handle of the 'status' level at a given handle.
func (ac *AppContext) GetLevelHandle(ctx context.Context, handle interface{}, level string) (interface{}, error) {
	ach := fmt.Sprintf("%v%v/", handle, level)
	hs, err := ac.rtc.RtcGetHandles(ctx, ach)

	//fmt.Println("\n\n")
	//fmt.Println("hs: ",hs)

	if err != nil {
		return nil, err
	}
	for _, v := range hs {
		//fmt.Println("v: ",v)
		//fmt.Println("\n\n")
		if v == ach {
			return v, nil
		}
	}
	return nil, pkgerrors.Errorf("No handle was found for level %v", level)
}

//Add app to the context under composite app
func (ac *AppContext) AddApp(ctx context.Context, handle interface{}, appname string) (interface{}, error) {
	h, err := ac.rtc.RtcAddLevel(ctx, handle, "app", appname)
	if err != nil {
		return nil, err
	}
	log.Info(":: Added app handle ::", log.Fields{"AppHandle": h})
	return h, nil
}

//Delete app from the context and everything underneth
func (ac *AppContext) DeleteApp(ctx context.Context, handle interface{}) error {
	err := ac.rtc.RtcDeletePrefix(ctx, handle)
	if err != nil {
		return err
	}
	return nil
}

//Returns the handle for a given app
func (ac *AppContext) GetAppHandle(ctx context.Context, appname string) (interface{}, error) {
	if appname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context app name")
	}

	rh, err := ac.rtc.RtcGet(ctx)
	if err != nil {
		return nil, err
	}

	apph := fmt.Sprintf("%v", rh) + "app/" + appname + "/"
	hs, err := ac.rtc.RtcGetHandles(ctx, apph)
	if err != nil {
		return nil, err
	}
	for _, v := range hs {
		if v == apph {
			return v, nil
		}
	}
	return nil, pkgerrors.Errorf("No handle was found for the given app")
}

// AddCluster helps to add cluster to the context under app. It takes in the app handle and clusterName as value.
func (ac *AppContext) AddCluster(ctx context.Context, handle interface{}, clustername string) (interface{}, error) {
	h, err := ac.rtc.RtcAddLevel(ctx, handle, "cluster", clustername)
	if err != nil {
		return nil, err
	}
	log.Info(":: Added cluster handle ::", log.Fields{"ClusterHandler": h})
	return h, nil
}

// AddClusterMetaGrp adds the meta info of groupNumber to which a cluster belongs.
// It takes in cluster handle and groupNumber as arguments
func (ac *AppContext) AddClusterMetaGrp(ctx context.Context, ch interface{}, gn string) error {
	mh, err := ac.rtc.RtcAddOneLevel(ctx, ch, metaGrpPREFIX, gn)
	if err != nil {
		return err
	}
	log.Info(":: Added cluster meta handle ::", log.Fields{"ClusterMetaHandler": mh})
	return nil
}

// DeleteClusterMetaGrpHandle deletes the group number to which the cluster belongs, it takes in the cluster handle.
func (ac *AppContext) DeleteClusterMetaGrpHandle(ctx context.Context, ch interface{}) error {
	err := ac.rtc.RtcDeletePrefix(ctx, ch)
	if err != nil {
		return err
	}
	log.Info(":: Deleted cluster meta handle ::", log.Fields{"ClusterMetaHandler": ch})
	return nil
}

/*
GetClusterMetaHandle takes in appName and ClusterName as string arguments and return the ClusterMetaHandle as string
*/
func (ac *AppContext) GetClusterMetaHandle(ctx context.Context, app string, cluster string) (string, error) {
	if app == "" {
		return "", pkgerrors.Errorf("Not a valid run time context app name")
	}
	if cluster == "" {
		return "", pkgerrors.Errorf("Not a valid run time context cluster name")
	}

	ch, err := ac.GetClusterHandle(ctx, app, cluster)
	if err != nil {
		return "", err
	}
	cmh := fmt.Sprintf("%v", ch) + metaGrpPREFIX + "/"
	return cmh, nil

}

/*
GetClusterGroupMap shall take in appName and return a map showing the grouping among the clusters.
sample output of "GroupMap" :{"1":["cluster_provider1+clusterName3","cluster_provider1+clusterName5"],"2":["cluster_provider2+clusterName4","cluster_provider2+clusterName6"]}
*/
func (ac *AppContext) GetClusterGroupMap(ctx context.Context, an string) (map[string][]string, error) {
	cl, err := ac.GetClusterNames(ctx, an)

	if err != nil {
		log.Info(":: Unable to fetch clusterList for app ::", log.Fields{"AppName ": an})
		return nil, err
	}
	rh, err := ac.rtc.RtcGet(ctx)
	if err != nil {
		return nil, err
	}

	var gmap = make(map[string][]string)
	for _, cn := range cl {
		s := fmt.Sprintf("%v", rh) + "app/" + an + "/cluster/" + cn + "/" + metaGrpPREFIX + "/"
		var v string
		err = ac.rtc.RtcGetValue(ctx, s, &v)
		if err != nil {
			log.Info(":: No group number for cluster  ::", log.Fields{"cluster": cn, "Reason": err})
			continue
		}
		gn := fmt.Sprintf("%v", v)
		log.Info(":: GroupNumber retrieved  ::", log.Fields{"GroupNumber": gn})

		cl, found := gmap[gn]
		if found == false {
			cl = make([]string, 0)
		}
		cl = append(cl, cn)
		gmap[gn] = cl

	}
	return gmap, nil
}

//Delete cluster from the context and everything underneth
func (ac *AppContext) DeleteCluster(ctx context.Context, handle interface{}) error {
	err := ac.rtc.RtcDeletePrefix(ctx, handle)
	if err != nil {
		return err
	}
	return nil
}

//Returns the handle for a given app and cluster
func (ac *AppContext) GetClusterHandle(ctx context.Context, appname string, clustername string) (interface{}, error) {
	if appname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context app name")
	}
	if clustername == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context cluster name")
	}

	rh, err := ac.rtc.RtcGet(ctx)
	if err != nil {
		return nil, err
	}

	ach := fmt.Sprintf("%v", rh) + "app/" + appname + "/cluster/" + clustername + "/"
	hs, err := ac.rtc.RtcGetHandles(ctx, ach)

	if err != nil {
		return nil, err
	}
	for _, v := range hs {
		if v == ach {
			return v, nil
		}
	}
	return nil, pkgerrors.Errorf("No handle was found for the given cluster")
}

//Returns a list of all clusters for a given app
func (ac *AppContext) GetClusterNames(ctx context.Context, appname string) ([]string, error) {
	if appname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context app name")
	}

	rh, err := ac.rtc.RtcGet(ctx)
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("%v", rh) + "app/" + appname + "/cluster/"

	hs, err := ac.rtc.RtcGetHandles(ctx, prefix)

	if err != nil {
		return nil, pkgerrors.Errorf("Error getting handles for %v", prefix)
	}
	var cs []string
	for _, h := range hs {
		hstr := fmt.Sprintf("%v", h)
		ks := strings.Split(hstr, prefix)
		for _, k := range ks {
			ck := strings.Split(k, "/")
			if len(ck) == 2 && ck[1] == "" {
				cs = append(cs, ck[0])
			}
		}
	}

	if len(cs) == 0 {
		err = pkgerrors.New("Cluster list is empty")
		log.Error("Cluster list is empty",
			log.Fields{"clusters": cs})
		return cs, err
	}
	return cs, nil
}

//Add resource under app and cluster
func (ac *AppContext) AddResource(ctx context.Context, handle interface{}, resname string, value interface{}) (interface{}, error) {
	h, err := ac.rtc.RtcAddResource(ctx, handle, resname, value)
	if err != nil {
		return nil, err
	}
	log.Info(":: Added resource handle ::", log.Fields{"ResourceHandler": h})

	return h, nil
}

//Return the handle for given app, cluster and resource name
func (ac *AppContext) GetResourceHandle(ctx context.Context, appname string, clustername string, resname string) (interface{}, error) {
	if appname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context app name")
	}
	if clustername == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context cluster name")
	}

	rh, err := ac.rtc.RtcGet(ctx)
	if err != nil {
		return nil, err
	}

	acrh := fmt.Sprintf("%v", rh) + "app/" + appname + "/cluster/" + clustername + "/resource/" + resname + "/"
	hs, err := ac.rtc.RtcGetHandles(ctx, acrh)
	if err != nil {
		return nil, err
	}
	for _, v := range hs {
		if v == acrh {
			return v, nil
		}
	}
	return nil, pkgerrors.Errorf("No handle was found for the given resource")
}

//Update the resource value using the given handle
func (ac *AppContext) UpdateResourceValue(ctx context.Context, handle interface{}, value interface{}) error {
	return ac.rtc.RtcUpdateValue(ctx, handle, value)
}

//Return the handle for given app, cluster and resource name
func (ac *AppContext) GetResourceStatusHandle(ctx context.Context, appname string, clustername string, resname string) (interface{}, error) {
	if appname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context app name")
	}
	if clustername == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context cluster name")
	}
	if resname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context resource name")
	}

	rh, err := ac.rtc.RtcGet(ctx)
	if err != nil {
		return nil, err
	}

	acrh := fmt.Sprintf("%v", rh) + "app/" + appname + "/cluster/" + clustername + "/resource/" + resname + "/status/"
	hs, err := ac.rtc.RtcGetHandles(ctx, acrh)
	if err != nil {
		return nil, err
	}
	for _, v := range hs {
		if v == acrh {
			return v, nil
		}
	}
	return nil, pkgerrors.Errorf("No handle was found for the given resource")
}

//GetResourceNames ... Returns a list of all resource names for a given app
func (ac *AppContext) GetResourceNames(ctx context.Context, appname string, clustername string) ([]string, error) {
	if appname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context app name")
	}
	if clustername == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context cluster name")
	}

	rh, err := ac.rtc.RtcGet(ctx)
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("%v", rh) + "app/" + appname + "/cluster/" + clustername + "/resource/"
	hs, err := ac.rtc.RtcGetHandles(ctx, prefix)
	if err != nil {
		return nil, pkgerrors.Errorf("Error getting handles for %v", prefix)
	}
	var cs []string
	for _, h := range hs {
		hstr := fmt.Sprintf("%v", h)
		ks := strings.Split(hstr, prefix)
		for _, k := range ks {
			ck := strings.Split(k, "/")
			if len(ck) == 2 && ck[1] == "" {
				cs = append(cs, ck[0])
			}
		}
	}
	return cs, nil
}

//Add instruction under given handle and type
func (ac *AppContext) AddInstruction(ctx context.Context, handle interface{}, level string, insttype string, value interface{}) (interface{}, error) {
	if !(insttype == OrderInstruction || insttype == DependencyInstruction) {
		log.Error("Not a valid app context instruction type", log.Fields{})
		return nil, pkgerrors.Errorf("Not a valid app context instruction type")
	}
	if !(level == "app" || level == "resource" || level == "subresource") {
		log.Error("Not a valid app context instruction level", log.Fields{})
		return nil, pkgerrors.Errorf("Not a valid app context instruction level")
	}
	h, err := ac.rtc.RtcAddInstruction(ctx, handle, level, insttype, value)
	if err != nil {
		log.Error("ac.rtc.RtcAddInstruction(handle, level, insttype, value)", log.Fields{"err": err})
		return nil, err
	}
	log.Info(":: Added instruction handle ::", log.Fields{"InstructionHandler": h})
	return h, nil
}

//Returns the resource instruction for a given instruction type per app
func (ac *AppContext) GetAppLevelInstruction(ctx context.Context, appname, insttype string) (interface{}, error) {
	if !(insttype == DependencyInstruction) {
		log.Error("Not a valid app context instruction type", log.Fields{})
		return nil, pkgerrors.Errorf("Not a valid app context instruction type")
	}

	rh, err := ac.rtc.RtcGet(ctx)
	if err != nil {
		log.Error("ac.rtc.RtcGet()", log.Fields{"err": err})
		return nil, err
	}
	s := fmt.Sprintf("%v", rh) + "app/" + appname + "/instruction/" + insttype + "/"
	log.Info("Getting app instruction", log.Fields{"s": s})
	var v string
	err = ac.rtc.RtcGetValue(ctx, s, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

//Delete instruction under given handle
func (ac *AppContext) DeleteInstruction(ctx context.Context, handle interface{}) error {
	err := ac.rtc.RtcDeletePair(ctx, handle)
	if err != nil {
		return err
	}
	return nil
}

//Returns the app instruction for a given instruction type
func (ac *AppContext) GetAppInstruction(ctx context.Context, insttype string) (interface{}, error) {
	if !(insttype == OrderInstruction || insttype == DependencyInstruction) {
		log.Error("Not a valid app context instruction type", log.Fields{})
		return nil, pkgerrors.Errorf("Not a valid app context instruction type")
	}
	rh, err := ac.rtc.RtcGet(ctx)
	if err != nil {
		log.Error("ac.rtc.RtcGet()", log.Fields{"err": err})
		return nil, err
	}
	s := fmt.Sprintf("%v", rh) + "app/" + "instruction/" + insttype + "/"
	log.Info("Getting app instruction", log.Fields{"s": s})
	var v string
	err = ac.rtc.RtcGetValue(ctx, s, &v)
	if err != nil {
		log.Error("ac.rtc.RtcGetValue(s, &v)", log.Fields{"err": err})
		return nil, err
	}
	return v, nil
}

//Update the instruction usign the given handle
func (ac *AppContext) UpdateInstructionValue(ctx context.Context, handle interface{}, value interface{}) error {
	return ac.rtc.RtcUpdateValue(ctx, handle, value)
}

//Returns the resource instruction for a given instruction type
func (ac *AppContext) GetResourceInstruction(ctx context.Context, appname string, clustername string, insttype string) (interface{}, error) {
	if !(insttype == OrderInstruction || insttype == DependencyInstruction) {
		log.Error("Not a valid app context instruction type", log.Fields{})
		return nil, pkgerrors.Errorf("Not a valid app context instruction type")
	}
	rh, err := ac.rtc.RtcGet(ctx)
	if err != nil {
		log.Error("ac.rtc.RtcGet()", log.Fields{"err": err})
		return nil, err
	}
	s := fmt.Sprintf("%v", rh) + "app/" + appname + "/cluster/" + clustername + "/resource/instruction/" + insttype + "/"
	var v string
	err = ac.rtc.RtcGetValue(ctx, s, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// AddLevelValue for holding a state object at a given level
// will make a handle with an appended "<level>/" to the key
func (ac *AppContext) AddLevelValue(ctx context.Context, handle interface{}, level string, value interface{}) (interface{}, error) {
	h, err := ac.rtc.RtcAddOneLevel(ctx, handle, level, value)
	if err != nil {
		return nil, err
	}
	log.Debug(":: Added handle ::", log.Fields{"Handle": h})

	return h, nil
}

// GetClusterStatusHandle returns the handle for cluster status for a given app and cluster
func (ac *AppContext) GetClusterStatusHandle(ctx context.Context, appname string, clustername string) (interface{}, error) {
	if appname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context app name")
	}
	if clustername == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context cluster name")
	}

	rh, err := ac.rtc.RtcGet(ctx)
	if err != nil {
		return nil, err
	}

	acrh := fmt.Sprintf("%v", rh) + "app/" + appname + "/cluster/" + clustername + "/status/"
	hs, err := ac.rtc.RtcGetHandles(ctx, acrh)
	if err != nil {
		return nil, err
	}
	for _, v := range hs {
		if v == acrh {
			return v, nil
		}
	}
	return nil, pkgerrors.Errorf("No handle was found for the given resource")
}

//UpdateStatusValue updates the status value with the given handle
func (ac *AppContext) UpdateStatusValue(ctx context.Context, handle interface{}, value interface{}) error {
	return ac.rtc.RtcUpdateValue(ctx, handle, value)
}

//UpdateValue updates the state value with the given handle
func (ac *AppContext) UpdateValue(ctx context.Context, handle interface{}, value interface{}) error {
	return ac.rtc.RtcUpdateValue(ctx, handle, value)
}

//Return all the handles under the composite app
func (ac *AppContext) GetAllHandles(ctx context.Context, handle interface{}) ([]interface{}, error) {
	hs, err := ac.rtc.RtcGetHandles(ctx, handle)
	if err != nil {
		return nil, err
	}
	return hs, nil
}

//Returns the value for a given handle
func (ac *AppContext) GetValue(ctx context.Context, handle interface{}) (interface{}, error) {
	var v interface{}
	err := ac.rtc.RtcGetValue(ctx, handle, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// GetCompositeAppMeta returns the meta data associated with the compositeApp
// Its return type is CompositeAppMeta
func (ac *AppContext) GetCompositeAppMeta(ctx context.Context) (CompositeAppMeta, error) {
	mi, err := ac.rtcObj.RtcGetMeta(ctx)

	if err != nil {
		return CompositeAppMeta{}, pkgerrors.Errorf("Failed to get compositeApp meta")
	}
	datamap, ok := mi.(map[string]interface{})

	if ok == false {
		return CompositeAppMeta{}, pkgerrors.Errorf("Failed to cast meta interface to compositeApp meta")
	}

	p := fmt.Sprintf("%v", datamap["Project"])
	ca := fmt.Sprintf("%v", datamap["CompositeApp"])
	v := fmt.Sprintf("%v", datamap["Version"])
	rn := fmt.Sprintf("%v", datamap["Release"])
	dig := fmt.Sprintf("%v", datamap["DeploymentIntentGroup"])
	namespace := fmt.Sprintf("%v", datamap["Namespace"])
	level := fmt.Sprintf("%v", datamap["Level"])
	var childInterface []interface{}
	childCtxs := make([]string, len(childInterface))
	if datamap["ChildContextIDs"] != nil {
		childInterface = datamap["ChildContextIDs"].([]interface{})
		for _, v := range childInterface {
			childCtxs = append(childCtxs, v.(string))
		}
	}
	lc := fmt.Sprintf("%v", datamap["LogicalCloud"])
	lcn := fmt.Sprintf("%v", datamap["LogicalCloudNamespace"])
	// user-intended level of logical cloud, not level of app itself (which a logical cloud can be):
	lclevel := fmt.Sprintf("%v", datamap["LogicalCloudLevel"])

	return CompositeAppMeta{Project: p, CompositeApp: ca, Version: v, Release: rn, DeploymentIntentGroup: dig, Namespace: namespace, Level: level, ChildContextIDs: childCtxs, LogicalCloud: lc, LogicalCloudNamespace: lcn, LogicalCloudLevel: lclevel}, nil
}
