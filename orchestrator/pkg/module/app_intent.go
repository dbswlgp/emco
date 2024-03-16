// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

/*
This file deals with the backend implementation of
Adding/Querying AppIntents for each application in the composite-app
*/

import (
	"context"
	"encoding/json"
	"reflect"

	pkgerrors "github.com/pkg/errors"
	gpic "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/gpic"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// AppIntent has two components - metadata, spec
type AppIntent struct {
	MetaData MetaData `json:"metadata,omitempty"`
	Spec     SpecData `json:"spec,omitempty"`
}

// MetaData has - name, description, userdata1, userdata2
type MetaData struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	UserData1   string `json:"userData1,omitempty"`
	UserData2   string `json:"userData2,omitempty"`
}

// SpecData consists of appName and intent
type SpecData struct {
	AppName string           `json:"app,omitempty"`
	Intent  gpic.IntentStruc `json:"intent,omitempty"`
}

// AppIntentManager is an interface which exposes the
// AppIntentManager functionalities
type AppIntentManager interface {
	CreateAppIntent(ctx context.Context, a AppIntent, p string, ca string, v string, i string, digName string, failIfExists bool) (AppIntent, bool, error)
	GetAppIntent(ctx context.Context, ai string, p string, ca string, v string, i string, digName string) (AppIntent, error)
	GetAllIntentsByApp(ctx context.Context, aN, p, ca, v, i, digName string) (SpecData, error)
	GetAllAppIntents(ctx context.Context, p, ca, v, i, digName string) ([]AppIntent, error)
	DeleteAppIntent(ctx context.Context, ai string, p string, ca string, v string, i string, digName string) error
}

//AppIntentQueryKey required for query
type AppIntentQueryKey struct {
	AppName string `json:"app"`
}

// AppIntentKey is used as primary key
type AppIntentKey struct {
	Name                      string `json:"genericAppPlacementIntent"`
	Project                   string `json:"project"`
	CompositeApp              string `json:"compositeApp"`
	Version                   string `json:"compositeAppVersion"`
	Intent                    string `json:"genericPlacementIntent"`
	DeploymentIntentGroupName string `json:"deploymentIntentGroup"`
}

// AppIntentFindByAppKey required for query
type AppIntentFindByAppKey struct {
	Project                   string `json:"project"`
	CompositeApp              string `json:"compositeApp"`
	CompositeAppVersion       string `json:"compositeAppVersion"`
	Intent                    string `json:"genericPlacementIntent"`
	DeploymentIntentGroupName string `json:"deploymentIntentGroup"`
	AppName                   string `json:"app"`
}

// ApplicationsAndClusterInfo type represents the list of
type ApplicationsAndClusterInfo struct {
	ArrayOfAppClusterInfo []AppClusterInfo `json:"applications"`
}

// AppClusterInfo is a type linking the app and the clusters
// on which they need to be installed.
type AppClusterInfo struct {
	Name       string       `json:"name"`
	AllOfArray []gpic.AllOf `json:"allOf,omitempty"`
	AnyOfArray []gpic.AnyOf `json:"anyOf,omitempty"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ak AppIntentKey) String() string {
	out, err := json.Marshal(ak)
	if err != nil {
		return ""
	}
	return string(out)
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ai AppIntentFindByAppKey) String() string {
	out, err := json.Marshal(ai)
	if err != nil {
		return ""
	}
	return string(out)
}

// AppIntentClient implements the AppIntentManager interface
type AppIntentClient struct {
	storeName   string
	tagMetaData string
}

// NewAppIntentClient returns an instance of AppIntentClient
func NewAppIntentClient() *AppIntentClient {
	return &AppIntentClient{
		storeName:   "resources",
		tagMetaData: "data",
	}
}

// CreateAppIntent creates an entry for AppIntent in the db.
// Other input parameters for it - projectName, compositeAppName, version, intentName and deploymentIntentGroupName.
// failIfExists - indicates the request is POST=true or PUT=false
func (c *AppIntentClient) CreateAppIntent(ctx context.Context, a AppIntent, p string, ca string, v string, i string, digName string, failIfExists bool) (AppIntent, bool, error) {
	aiExists := false

	//Check for the AppIntent already exists here.
	res, err := c.GetAppIntent(ctx, a.MetaData.Name, p, ca, v, i, digName)
	if err == nil && !reflect.DeepEqual(res, AppIntent{}) {
		aiExists = true
	}

	if aiExists && failIfExists {
		return AppIntent{}, aiExists, pkgerrors.New("AppIntent already exists")
	}

	akey := AppIntentKey{
		Name:                      a.MetaData.Name,
		Project:                   p,
		CompositeApp:              ca,
		Version:                   v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
	}

	qkey := AppIntentQueryKey{
		AppName: a.Spec.AppName,
	}

	err = db.DBconn.Insert(ctx, c.storeName, akey, qkey, c.tagMetaData, a)
	if err != nil {
		return AppIntent{}, aiExists, err
	}

	return a, aiExists, nil
}

// GetAppIntent shall take arguments - name of the app intent, name of the project, name of the composite app, version of the composite app,intent name and deploymentIntentGroupName. It shall return the AppIntent
func (c *AppIntentClient) GetAppIntent(ctx context.Context, ai string, p string, ca string, v string, i string, digName string) (AppIntent, error) {

	k := AppIntentKey{
		Name:                      ai,
		Project:                   p,
		CompositeApp:              ca,
		Version:                   v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
	}

	result, err := db.DBconn.Find(ctx, c.storeName, k, c.tagMetaData)
	if err != nil {
		return AppIntent{}, err
	}

	if len(result) == 0 {
		return AppIntent{}, pkgerrors.New("AppIntent not found")
	}

	if result != nil {
		a := AppIntent{}
		err = db.DBconn.Unmarshal(result[0], &a)
		if err != nil {
			return AppIntent{}, err
		}
		return a, nil

	}
	return AppIntent{}, pkgerrors.New("Unknown Error")
}

/*
GetAllIntentsByApp queries intent by AppName, it takes in parameters AppName, CompositeAppName, CompositeNameVersion,
GenericPlacementIntentName & DeploymentIntentGroupName. Returns SpecData which contains
all the intents for the app.
*/
func (c *AppIntentClient) GetAllIntentsByApp(ctx context.Context, aN, p, ca, v, i, digName string) (SpecData, error) {
	k := AppIntentFindByAppKey{
		Project:                   p,
		CompositeApp:              ca,
		CompositeAppVersion:       v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
		AppName:                   aN,
	}
	result, err := db.DBconn.Find(ctx, c.storeName, k, c.tagMetaData)
	if err != nil {
		return SpecData{}, err
	}
	if len(result) == 0 {
		return SpecData{}, nil
	}

	var a AppIntent
	err = db.DBconn.Unmarshal(result[0], &a)
	if err != nil {
		return SpecData{}, err
	}
	return a.Spec, nil

}

/*
GetAllAppIntents takes in paramaters ProjectName, CompositeAppName, CompositeNameVersion
and GenericPlacementIntentName,DeploymentIntentGroupName. Returns an array of AppIntents
*/
func (c *AppIntentClient) GetAllAppIntents(ctx context.Context, p, ca, v, i, digName string) ([]AppIntent, error) {
	k := AppIntentKey{
		Name:                      "",
		Project:                   p,
		CompositeApp:              ca,
		Version:                   v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
	}
	result, err := db.DBconn.Find(ctx, c.storeName, k, c.tagMetaData)
	if err != nil {
		return []AppIntent{}, err
	}

	var appIntents []AppIntent

	if len(result) != 0 {
		for i := range result {
			aI := AppIntent{}
			err = db.DBconn.Unmarshal(result[i], &aI)
			if err != nil {
				return []AppIntent{}, err
			}
			appIntents = append(appIntents, aI)
		}
	}

	return appIntents, err
}

// DeleteAppIntent delete an AppIntent
func (c *AppIntentClient) DeleteAppIntent(ctx context.Context, ai string, p string, ca string, v string, i string, digName string) error {
	k := AppIntentKey{
		Name:                      ai,
		Project:                   p,
		CompositeApp:              ca,
		Version:                   v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
	}

	err := db.DBconn.Remove(ctx, c.storeName, k)
	return err

}
