// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"

	pkgerrors "github.com/pkg/errors"
)

// DeploymentIntentGroup shall have 2 fields - MetaData and Spec
type DeploymentIntentGroup struct {
	MetaData DepMetaData `json:"metadata"`
	Spec     DepSpecData `json:"spec"`
}

// DepMetaData has Name, description, userdata1, userdata2
type DepMetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// DepSpecData has profile, version, OverrideValuesObj
type DepSpecData struct {
	Profile           string           `json:"compositeProfile"`
	Version           string           `json:"version"`
	OverrideValuesObj []OverrideValues `json:"overrideValues"`
	LogicalCloud      string           `json:"logicalCloud"`
}

// OverrideValues has appName and ValuesObj
type OverrideValues struct {
	AppName   string            `json:"app"`
	ValuesObj map[string]string `json:"values"`
}

// Values has ImageRepository
// type Values struct {
// 	ImageRepository string `json:"imageRepository"`
// }

// DeploymentIntentGroupManager is an interface which exposes the DeploymentIntentGroupManager functionality
type DeploymentIntentGroupManager interface {
	CreateDeploymentIntentGroup(ctx context.Context, d DeploymentIntentGroup, p string, ca string, v string, failIfExists bool) (DeploymentIntentGroup, bool, error)
	GetDeploymentIntentGroup(ctx context.Context, di string, p string, ca string, v string) (DeploymentIntentGroup, error)
	GetDeploymentIntentGroupState(ctx context.Context, di string, p string, ca string, v string) (state.StateInfo, error)
	DeleteDeploymentIntentGroup(ctx context.Context, di string, p string, ca string, v string) error
	GetAllDeploymentIntentGroups(ctx context.Context, p string, ca string, v string) ([]DeploymentIntentGroup, error)
	GetAllExistDeploymentIntentGroups(ctx context.Context, p string) ([]DeploymentIntentGroup, error)
}

// DeploymentIntentGroupKey consists of Name of the deployment group, project name, CompositeApp name, CompositeApp version
type DeploymentIntentGroupKey struct {
	Name         string `json:"deploymentIntentGroup"`
	Project      string `json:"project"`
	CompositeApp string `json:"compositeApp"`
	Version      string `json:"compositeAppVersion"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (dk DeploymentIntentGroupKey) String() string {
	out, err := json.Marshal(dk)
	if err != nil {
		return ""
	}
	return string(out)
}

// DeploymentIntentGroupClient implements the DeploymentIntentGroupManager interface
type DeploymentIntentGroupClient struct {
	storeName   string
	tagMetaData string
	tagState    string
}

// NewDeploymentIntentGroupClient return an instance of DeploymentIntentGroupClient which implements DeploymentIntentGroupManager
func NewDeploymentIntentGroupClient() *DeploymentIntentGroupClient {
	return &DeploymentIntentGroupClient{
		storeName:   "resources",
		tagMetaData: "data",
		tagState:    "stateInfo",
	}
}

// CreateDeploymentIntentGroup creates an entry for a given  DeploymentIntentGroup in the database. Other Input parameters for it - projectName, compositeAppName, version
func (c *DeploymentIntentGroupClient) CreateDeploymentIntentGroup(ctx context.Context, d DeploymentIntentGroup, p string, ca string, v string, failIfExists bool) (DeploymentIntentGroup, bool, error) {
	digExists := false

	// check if the DeploymentIntentGroup already exists.
	res, err := c.GetDeploymentIntentGroup(ctx, d.MetaData.Name, p, ca, v)
	if err == nil && !reflect.DeepEqual(res, DeploymentIntentGroup{}) {
		digExists = true
	}

	if digExists && failIfExists {
		return DeploymentIntentGroup{}, digExists, pkgerrors.New("DeploymentIntent already exists")
	}

	gkey := DeploymentIntentGroupKey{
		Name:         d.MetaData.Name,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	if digExists {
		// The DeploymentIntentGroup exists. Check the state of the DeploymentIntentGroup
		// Update the DeploymentIntentGroup if the state is "Created"
		stateInfo, err := c.GetDeploymentIntentGroupState(ctx, d.MetaData.Name, p, ca, v)
		if err != nil {
			return DeploymentIntentGroup{}, digExists, err
		}

		currentState, err := state.GetCurrentStateFromStateInfo(stateInfo)
		if err != nil {
			return DeploymentIntentGroup{}, digExists, err
		}

		if currentState == state.StateEnum.Created {
			err := db.DBconn.Insert(ctx, c.storeName, gkey, nil, c.tagMetaData, d)
			if err != nil {
				return DeploymentIntentGroup{}, digExists, err
			}
			return d, digExists, nil
		}

		return DeploymentIntentGroup{}, digExists, pkgerrors.Errorf(
			"The DeploymentIntentGroup is not updated. The DeploymentIntentGroup, %s, is in %s state",
			d.MetaData.Name,
			currentState,
		)
	}

	// The DeploymentIntentGroup does not exists. Create the DeploymentIntentGroup and add the StateInfo details
	err = db.DBconn.Insert(ctx, c.storeName, gkey, nil, c.tagMetaData, d)
	if err != nil {
		return DeploymentIntentGroup{}, digExists, err
	}

	// Add the stateInfo record
	s := state.StateInfo{}
	a := state.ActionEntry{
		State:     state.StateEnum.Created,
		ContextId: "",
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)

	err = db.DBconn.Insert(ctx, c.storeName, gkey, nil, c.tagState, s)
	if err != nil {
		return DeploymentIntentGroup{}, digExists, pkgerrors.Wrapf(err, "Error updating the stateInfo of the DeploymentIntentGroup: %s", d.MetaData.Name)
	}

	return d, digExists, nil
}

// GetDeploymentIntentGroup returns the DeploymentIntentGroup with a given name, project, compositeApp and version of compositeApp
func (c *DeploymentIntentGroupClient) GetDeploymentIntentGroup(ctx context.Context, di string, p string, ca string, v string) (DeploymentIntentGroup, error) {

	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	result, err := db.DBconn.Find(ctx, c.storeName, key, c.tagMetaData)
	if err != nil {
		return DeploymentIntentGroup{}, err
	} else if len(result) == 0 {
		return DeploymentIntentGroup{}, pkgerrors.New("DeploymentIntentGroup not found")
	}

	if result != nil {
		d := DeploymentIntentGroup{}
		err = db.DBconn.Unmarshal(result[0], &d)
		if err != nil {
			return DeploymentIntentGroup{}, err
		}
		return d, nil
	}

	return DeploymentIntentGroup{}, pkgerrors.New("Unknown Error")

}

// GetAllDeploymentIntentGroups returns all the deploymentIntentGroups under a specific project, compositeApp and version
func (c *DeploymentIntentGroupClient) GetAllDeploymentIntentGroups(ctx context.Context, p string, ca string, v string) ([]DeploymentIntentGroup, error) {

	key := DeploymentIntentGroupKey{
		Name:         "",
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	//Check if project exists
	_, err := NewProjectClient().GetProject(ctx, p)
	if err != nil {
		return []DeploymentIntentGroup{}, pkgerrors.Wrap(err, "Project not found")
	}

	//check if compositeApp exists
	_, err = NewCompositeAppClient().GetCompositeApp(ctx, ca, v, p)
	if err != nil {
		return []DeploymentIntentGroup{}, err
	}
	var diList []DeploymentIntentGroup
	result, err := db.DBconn.Find(ctx, c.storeName, key, c.tagMetaData)
	if err != nil {
		return []DeploymentIntentGroup{}, err
	}

	for _, value := range result {
		di := DeploymentIntentGroup{}
		err = db.DBconn.Unmarshal(value, &di)
		if err != nil {
			return []DeploymentIntentGroup{}, err
		}
		diList = append(diList, di)
	}

	return diList, nil

}

// GetAllDeploymentIntentGroups returns all the deploymentIntentGroups under a specific project, compositeApp and version
func (c *DeploymentIntentGroupClient) GetAllExistDeploymentIntentGroups(ctx context.Context, p string) ([]DeploymentIntentGroup, error) {

	key := DeploymentIntentGroupKey{
		Name:         "",
		Project:      p,
		CompositeApp: "",
		Version:      "",
	}

	//Check if project exists
	_, err := NewProjectClient().GetProject(ctx, p)
	if err != nil {
		return []DeploymentIntentGroup{}, pkgerrors.Wrap(err, "Project not found")
	}

	var diList []DeploymentIntentGroup
	result, err := db.DBconn.Find(ctx, c.storeName, key, c.tagMetaData)
	if err != nil {
		return []DeploymentIntentGroup{}, err
	}

	for _, value := range result {
		di := DeploymentIntentGroup{}
		err = db.DBconn.Unmarshal(value, &di)
		if err != nil {
			return []DeploymentIntentGroup{}, err
		}
		diList = append(diList, di)
	}

	return diList, nil

}

// GetDeploymentIntentGroupState returns the DIG-StateInfo with a given DeploymentIntentname, project, compositeAppName and version of compositeApp
func (c *DeploymentIntentGroupClient) GetDeploymentIntentGroupState(ctx context.Context, di string, p string, ca string, v string) (state.StateInfo, error) {

	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	result, err := db.DBconn.Find(ctx, c.storeName, key, c.tagState)
	if err != nil {
		return state.StateInfo{}, err
	}

	if len(result) == 0 {
		return state.StateInfo{}, pkgerrors.New("DeploymentIntentGroup StateInfo not found")
	}

	if result != nil {
		s := state.StateInfo{}
		err = db.DBconn.Unmarshal(result[0], &s)
		if err != nil {
			return state.StateInfo{}, err
		}
		return s, nil
	}

	return state.StateInfo{}, pkgerrors.New("Unknown Error")
}

// DeleteDeploymentIntentGroup deletes a DeploymentIntentGroup
func (c *DeploymentIntentGroupClient) DeleteDeploymentIntentGroup(ctx context.Context, di string, p string, ca string, v string) error {
	k := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}
	s, err := c.GetDeploymentIntentGroupState(ctx, di, p, ca, v)
	if err != nil {
		// If the StateInfo cannot be found, then a proper deployment intent group record is not present.
		// Call the DB delete to clean up any errant record without a StateInfo element that may exist.
		err = db.DBconn.Remove(ctx, c.storeName, k)
		if err != nil {
			return pkgerrors.Wrap(err, "Error deleting DeploymentIntentGroup entry")
		}
		return nil
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting current state from DeploymentIntentGroup stateInfo: "+di)
	}

	if stateVal == state.StateEnum.Instantiated || stateVal == state.StateEnum.InstantiateStopped {
		return pkgerrors.Errorf("DeploymentIntentGroup must be terminated before it can be deleted " + di)
	}

	// remove the app contexts associated with thie Deployment Intent Group
	if stateVal == state.StateEnum.Terminated || stateVal == state.StateEnum.TerminateStopped {
		// Verify that the appcontext has completed terminating
		ctxid := state.GetLastContextIdFromStateInfo(s)
		acStatus, err := state.GetAppContextStatus(ctx, ctxid)
		if err == nil &&
			!(acStatus.Status == appcontext.AppContextStatusEnum.Terminated || acStatus.Status == appcontext.AppContextStatusEnum.TerminateFailed) {
			return pkgerrors.New("DeploymentIntentGroup has not completed terminating: " + di)
		}

		for _, id := range state.GetContextIdsFromStateInfo(s) {
			context, err := state.GetAppContextFromId(ctx, id)
			if err != nil {
				return pkgerrors.Wrap(err, "Error getting appcontext from DeploymentIntentGroup StateInfo")
			}
			err = context.DeleteCompositeApp(ctx)
			if err != nil {
				return pkgerrors.Wrap(err, "Error deleting appcontext for DeploymentIntentGroup")
			}
		}
	}

	err = db.DBconn.Remove(ctx, c.storeName, k)
	return err
}
