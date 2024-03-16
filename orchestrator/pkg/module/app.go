// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"encoding/json"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// App contains metadata for Apps
type App struct {
	Metadata AppMetaData `json:"metadata"`
}

//AppMetaData contains the parameters needed for Apps
type AppMetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

//AppContent contains fileContent
type AppContent struct {
	FileContent string
}

// AppKey is the key structure that is used in the database
type AppKey struct {
	App                 string `json:"app"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (aK AppKey) String() string {
	out, err := json.Marshal(aK)
	if err != nil {
		return ""
	}
	return string(out)
}

// AppManager is an interface exposes the App functionality
type AppManager interface {
	CreateApp(ctx context.Context, a App, ac AppContent, p string, cN string, cV string, exists bool) (App, error)
	GetApp(ctx context.Context, name string, p string, cN string, cV string) (App, error)
	GetAppContent(ctx context.Context, name string, p string, cN string, cV string) (AppContent, error)
	GetApps(ctx context.Context, p string, cN string, cV string) ([]App, error)
	DeleteApp(ctx context.Context, name string, p string, cN string, cV string) error
}

// AppClient implements the AppManager
// It will also be used to maintain some localized state
type AppClient struct {
	storeName           string
	tagMeta, tagContent string
}

// NewAppClient returns an instance of the AppClient
// which implements the AppManager
func NewAppClient() *AppClient {
	return &AppClient{
		storeName:  "resources",
		tagMeta:    "data",
		tagContent: "appcontent",
	}
}

// CreateApp creates a new collection based on the App
func (v *AppClient) CreateApp(ctx context.Context, a App, ac AppContent, p string, cN string, cV string, exists bool) (App, error) {

	//Construct the composite key to select the entry
	key := AppKey{
		App:                 a.Metadata.Name,
		Project:             p,
		CompositeApp:        cN,
		CompositeAppVersion: cV,
	}

	//Check if this App already exists
	_, err := v.GetApp(ctx, a.Metadata.Name, p, cN, cV)
	if err == nil && !exists {
		return App{}, pkgerrors.New("App already exists")
	}

	err = db.DBconn.Insert(ctx, v.storeName, key, nil, v.tagMeta, a)
	if err != nil {
		return App{}, pkgerrors.Wrap(err, "Create DB entry error")
	}

	err = db.DBconn.Insert(ctx, v.storeName, key, nil, v.tagContent, ac)
	if err != nil {
		return App{}, pkgerrors.Wrap(err, "Create DB entry error")
	}

	return a, nil
}

// GetApp returns the App for corresponding name
func (v *AppClient) GetApp(ctx context.Context, name string, p string, cN string, cV string) (App, error) {

	//Construct the composite key to select the entry
	key := AppKey{
		App:                 name,
		Project:             p,
		CompositeApp:        cN,
		CompositeAppVersion: cV,
	}
	value, err := db.DBconn.Find(ctx, v.storeName, key, v.tagMeta)
	if err != nil {
		return App{}, err
	} else if len(value) == 0 {
		return App{}, pkgerrors.New("App not found")
	}

	//value is a byte array
	if value != nil {
		app := App{}
		err = db.DBconn.Unmarshal(value[0], &app)
		if err != nil {
			return App{}, err
		}
		return app, nil
	}

	return App{}, pkgerrors.New("Unknown Error")
}

// GetAppContent returns content for corresponding app
func (v *AppClient) GetAppContent(ctx context.Context, name string, p string, cN string, cV string) (AppContent, error) {

	//Construct the composite key to select the entry
	key := AppKey{
		App:                 name,
		Project:             p,
		CompositeApp:        cN,
		CompositeAppVersion: cV,
	}
	value, err := db.DBconn.Find(ctx, v.storeName, key, v.tagContent)
	if err != nil {
		return AppContent{}, err
	} else if len(value) == 0 {
		return AppContent{}, pkgerrors.New("AppContent not found")
	}

	//value is a byte array
	if value != nil {
		ac := AppContent{}
		err = db.DBconn.Unmarshal(value[0], &ac)
		if err != nil {
			return AppContent{}, err
		}
		return ac, nil
	}

	return AppContent{}, pkgerrors.New("Unknown Error")
}

// GetApps returns all Apps for given composite App
func (v *AppClient) GetApps(ctx context.Context, project, compositeApp, compositeAppVersion string) ([]App, error) {

	key := AppKey{
		App:                 "",
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
	}

	var resp []App
	values, err := db.DBconn.Find(ctx, v.storeName, key, v.tagMeta)
	if err != nil {
		return []App{}, err
	}

	for _, value := range values {
		a := App{}
		err = db.DBconn.Unmarshal(value, &a)
		if err != nil {
			return []App{}, err
		}
		resp = append(resp, a)
	}

	return resp, nil
}

// DeleteApp deletes the  App from database
func (v *AppClient) DeleteApp(ctx context.Context, name string, p string, cN string, cV string) error {

	//Construct the composite key to select the entry
	key := AppKey{
		App:                 name,
		Project:             p,
		CompositeApp:        cN,
		CompositeAppVersion: cV,
	}
	err := db.DBconn.Remove(ctx, v.storeName, key)
	return err
}
