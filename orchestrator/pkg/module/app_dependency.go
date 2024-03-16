// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"encoding/json"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// AppDependency contains the metaData for AppDependency
type AppDependency struct {
	MetaData AdMetaData `json:"metadata"`
	Spec     AdSpecData `json:"spec,omitempty"`
}

// AppDependencyMetaData contains the parameters for creating a AppDependency
type AdMetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// SpecData consists of appName dependent
type AdSpecData struct {
	AppName string `json:"app,omitempty"`
	// Ready/Deployed
	OpStatus string `json:"opStatus,omitempty"`
	// Wait time in seconds
	Wait int `json:"wait,omitempty"`
}

// AppDependencyKey is the key structure that is used in the database
type AppDependencyKey struct {
	Name         string `json:"appDependency"`
	AppName      string `json:"app"`
	Project      string `json:"project"`
	CompositeApp string `json:"compositeApp"`
	Version      string `json:"compositeAppVersion"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (pk AppDependency) String() string {
	out, err := json.Marshal(pk)
	if err != nil {
		return ""
	}

	return string(out)
}

// AppDependencyManager is an interface exposes the AppDependency functionality
type AppDependencyManager interface {
	CreateAppDependency(ctx context.Context, dep AppDependency, p string, ca string, v string, app string, exists bool) (AppDependency, error)
	GetAppDependency(ctx context.Context, dep string, p string, ca string, v string, app string) (AppDependency, error)
	DeleteAppDependency(ctx context.Context, dep string, p string, ca string, v string, app string) error
	GetAllAppDependency(ctx context.Context, p string, ca string, v string, app string) ([]AppDependency, error)
}

// AppDependencyClient implements the AppDependencyManager
// It will also be used to maintain some localized state
type AppDependencyClient struct {
	storeName string
	tagMeta   string
}

// NewAppDependencyClient returns an instance of the AppDependencyClient
// which implements the AppDependencyManager
func NewAppDependencyClient() *AppDependencyClient {
	return &AppDependencyClient{
		storeName: "resources",
		tagMeta:   "data",
	}
}

// CreateAppDependency a new collection based on the AppDependency
func (d *AppDependencyClient) CreateAppDependency(ctx context.Context, dep AppDependency, p string, ca string, v string, app string, exists bool) (AppDependency, error) {

	//Construct the key to select the entry
	key := AppDependencyKey{
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		AppName:      app,
		Name:         dep.MetaData.Name,
	}

	//Check if this AppDependency already exists
	_, err := d.GetAppDependency(ctx, dep.MetaData.Name, p, ca, v, app)
	if err == nil && !exists {
		return AppDependency{}, pkgerrors.New("AppDependency already exists")
	}

	err = db.DBconn.Insert(ctx, d.storeName, key, nil, d.tagMeta, dep)
	if err != nil {
		return AppDependency{}, pkgerrors.Wrap(err, "Create DB entry error")
	}

	return dep, nil
}

// GetAppDependency returns the AppDependency for corresponding name
func (d *AppDependencyClient) GetAppDependency(ctx context.Context, dep string, p string, ca string, v string, app string) (AppDependency, error) {

	//Construct the composite key to select the entry
	key := AppDependencyKey{
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		AppName:      app,
		Name:         dep,
	}
	value, err := db.DBconn.Find(ctx, d.storeName, key, d.tagMeta)
	if err != nil {
		return AppDependency{}, err
	} else if len(value) == 0 {
		return AppDependency{}, pkgerrors.New("AppDependency not found")
	}

	//value is a byte array
	if value != nil {
		proj := AppDependency{}
		err = db.DBconn.Unmarshal(value[0], &proj)
		if err != nil {
			return AppDependency{}, err
		}
		return proj, nil
	}

	return AppDependency{}, pkgerrors.New("Unknown Error")
}

// GetAllAppDependency returns all the AppDependencys
func (d *AppDependencyClient) GetAllAppDependency(ctx context.Context, p string, ca string, v string, app string) ([]AppDependency, error) {
	key := AppDependencyKey{
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		AppName:      app,
		Name:         "",
	}

	res := make([]AppDependency, 0)
	values, err := db.DBconn.Find(ctx, d.storeName, key, d.tagMeta)
	if err != nil {
		return []AppDependency{}, err
	}

	for _, value := range values {
		p := AppDependency{}
		err = db.DBconn.Unmarshal(value, &p)
		if err != nil {
			return []AppDependency{}, err
		}
		res = append(res, p)
	}
	return res, nil
}

// DeleteAppDependency the  AppDependency from database
func (d *AppDependencyClient) DeleteAppDependency(ctx context.Context, name string, p string, ca string, v string, app string) error {

	//Construct the composite key to select the entry
	key := AppDependencyKey{
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		AppName:      app,
		Name:         name,
	}
	err := db.DBconn.Remove(ctx, d.storeName, key)
	return err
}

// GetAllAppDependency returns all the AppDependencys
func (d *AppDependencyClient) GetAllSpecAppDependency(ctx context.Context, p string, ca string, v string, app string) ([]AdSpecData, error) {
	key := AppDependencyKey{
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		AppName:      app,
		Name:         "",
	}

	var res []AdSpecData
	values, err := db.DBconn.Find(ctx, d.storeName, key, d.tagMeta)
	if err != nil {
		return []AdSpecData{}, err
	}

	for _, value := range values {
		p := AppDependency{}
		err = db.DBconn.Unmarshal(value, &p)
		if err != nil {
			return []AdSpecData{}, err
		}
		res = append(res, p.Spec)
	}
	return res, nil
}
