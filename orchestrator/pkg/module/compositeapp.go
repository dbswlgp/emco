// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module

import (
	"context"
	"encoding/json"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// CompositeApp contains metadata and spec for CompositeApps
type CompositeApp struct {
	Metadata CompositeAppMetaData `json:"metadata"`
	Spec     CompositeAppSpec     `json:"spec"`
}

//CompositeAppMetaData contains the parameters needed for CompositeApps
type CompositeAppMetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

//CompositeAppSpec contains the Version of the CompositeApp
type CompositeAppSpec struct {
	Version string `json:"compositeAppVersion"`
}

// CompositeAppKey is the key structure that is used in the database
type CompositeAppKey struct {
	CompositeAppName string `json:"compositeApp"`
	Version          string `json:"compositeAppVersion"`
	Project          string `json:"project"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (cK CompositeAppKey) String() string {
	out, err := json.Marshal(cK)
	if err != nil {
		return ""
	}
	return string(out)
}

// CompositeAppManager is an interface exposes the CompositeApp functionality
type CompositeAppManager interface {
	CreateCompositeApp(ctx context.Context, c CompositeApp, p string, exists bool) (CompositeApp, error)
	GetCompositeApp(ctx context.Context, name string, version string, p string) (CompositeApp, error)
	GetAllCompositeApps(ctx context.Context, p string) ([]CompositeApp, error)
	DeleteCompositeApp(ctx context.Context, name string, version string, p string) error
}

// CompositeAppClient implements the CompositeAppManager
// It will also be used to maintain some localized state
type CompositeAppClient struct {
	storeName string
	tagMeta   string
}

// NewCompositeAppClient returns an instance of the CompositeAppClient
// which implements the CompositeAppManager
func NewCompositeAppClient() *CompositeAppClient {
	return &CompositeAppClient{
		storeName: "resources",
		tagMeta:   "data",
	}
}

// CreateCompositeApp creates a new collection based on the CompositeApp
func (v *CompositeAppClient) CreateCompositeApp(ctx context.Context, c CompositeApp, p string, exists bool) (CompositeApp, error) {

	//Construct the composite key to select the entry
	key := CompositeAppKey{
		CompositeAppName: c.Metadata.Name,
		Version:          c.Spec.Version,
		Project:          p,
	}

	//Check if this CompositeApp already exists
	_, err := v.GetCompositeApp(ctx, c.Metadata.Name, c.Spec.Version, p)
	if err == nil && !exists {
		return CompositeApp{}, pkgerrors.New("CompositeApp already exists")
	}

	err = db.DBconn.Insert(ctx, v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return CompositeApp{}, pkgerrors.Wrap(err, "Create DB entry error")
	}

	return c, nil
}

// GetCompositeApp returns the CompositeApp for corresponding name
func (v *CompositeAppClient) GetCompositeApp(ctx context.Context, name string, version string, p string) (CompositeApp, error) {

	//Construct the composite key to select the entry
	key := CompositeAppKey{
		CompositeAppName: name,
		Version:          version,
		Project:          p,
	}
	value, err := db.DBconn.Find(ctx, v.storeName, key, v.tagMeta)
	if err != nil {
		return CompositeApp{}, err
	} else if len(value) == 0 {
		return CompositeApp{}, pkgerrors.New("CompositeApp not found")
	}

	//value is a byte array
	if value != nil {
		compApp := CompositeApp{}
		err = db.DBconn.Unmarshal(value[0], &compApp)
		if err != nil {
			return CompositeApp{}, err
		}
		return compApp, nil
	}

	return CompositeApp{}, pkgerrors.New("Unknown Error")
}

// GetAllCompositeApps returns all the compositeApp for a given project
func (v *CompositeAppClient) GetAllCompositeApps(ctx context.Context, p string) ([]CompositeApp, error) {

	_, err := NewProjectClient().GetProject(ctx, p)
	if err != nil {
		return []CompositeApp{}, pkgerrors.Wrap(err, "Project not found")
	}

	key := CompositeAppKey{
		CompositeAppName: "",
		Version:          "",
		Project:          p,
	}

	var caList []CompositeApp
	values, err := db.DBconn.Find(ctx, v.storeName, key, v.tagMeta)
	if err != nil {
		return []CompositeApp{}, err
	}

	for _, value := range values {
		ca := CompositeApp{}
		err = db.DBconn.Unmarshal(value, &ca)
		if err != nil {
			return []CompositeApp{}, err
		}
		caList = append(caList, ca)
	}

	return caList, nil
}

// DeleteCompositeApp deletes the  CompositeApp from database
func (v *CompositeAppClient) DeleteCompositeApp(ctx context.Context, name string, version string, p string) error {

	//Construct the composite key to select the entry
	key := CompositeAppKey{
		CompositeAppName: name,
		Version:          version,
		Project:          p,
	}
	err := db.DBconn.Remove(ctx, v.storeName, key)
	return err
}
