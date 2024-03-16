// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"encoding/json"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// GenericPlacementIntent shall have 2 fields - metadata and spec
type GenericPlacementIntent struct {
	MetaData GenIntentMetaData `json:"metadata"`
}

// GenIntentMetaData has name, description, userdata1, userdata2
type GenIntentMetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// GenericPlacementIntentManager is an interface which exposes the GenericPlacementIntentManager functionality
type GenericPlacementIntentManager interface {
	CreateGenericPlacementIntent(ctx context.Context, g GenericPlacementIntent, p string, ca string, v string, digName string, failIfExists bool) (GenericPlacementIntent, bool, error)
	GetGenericPlacementIntent(ctx context.Context, intentName string, projectName string,
		compositeAppName string, version string, digName string) (GenericPlacementIntent, error)
	DeleteGenericPlacementIntent(ctx context.Context, intentName string, projectName string,
		compositeAppName string, version string, digName string) error

	GetAllGenericPlacementIntents(ctx context.Context, p string, ca string, v string, digName string) ([]GenericPlacementIntent, error)
}

// GenericPlacementIntentKey is used as the primary key
type GenericPlacementIntentKey struct {
	Name         string `json:"genericPlacementIntent"`
	Project      string `json:"project"`
	CompositeApp string `json:"compositeApp"`
	Version      string `json:"compositeAppVersion"`
	DigName      string `json:"deploymentIntentGroup"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (gk GenericPlacementIntentKey) String() string {
	out, err := json.Marshal(gk)
	if err != nil {
		return ""
	}
	return string(out)
}

// GenericPlacementIntentClient implements the GenericPlacementIntentManager interface
type GenericPlacementIntentClient struct {
	storeName   string
	tagMetaData string
}

// NewGenericPlacementIntentClient return an instance of GenericPlacementIntentClient which implements GenericPlacementIntentManager
func NewGenericPlacementIntentClient() *GenericPlacementIntentClient {
	return &GenericPlacementIntentClient{
		storeName:   "resources",
		tagMetaData: "data",
	}
}

// CreateGenericPlacementIntent creates an entry for GenericPlacementIntent in the database.
// Other Input parameters for it - projectName, compositeAppName, version and deploymentIntentGroupName
// failIfExists - indicates the request is POST=true or PUT=false
func (c *GenericPlacementIntentClient) CreateGenericPlacementIntent(ctx context.Context, g GenericPlacementIntent, p string, ca string, v string, digName string, failIfExists bool) (GenericPlacementIntent, bool, error) {
	gpiExists := false

	// check if the genericPlacement already exists.
	res, err := c.GetGenericPlacementIntent(ctx, g.MetaData.Name, p, ca, v, digName)
	if err == nil && res != (GenericPlacementIntent{}) {
		gpiExists = true
	}

	if gpiExists && failIfExists {
		return GenericPlacementIntent{}, gpiExists, pkgerrors.New("Intent already exists")
	}

	gkey := GenericPlacementIntentKey{
		Name:         g.MetaData.Name,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		DigName:      digName,
	}

	err = db.DBconn.Insert(ctx, c.storeName, gkey, nil, c.tagMetaData, g)
	if err != nil {
		return GenericPlacementIntent{}, gpiExists, err
	}

	return g, gpiExists, nil
}

// GetGenericPlacementIntent shall take arguments - name of the intent, name of the project, name of the composite app, version of the composite app and deploymentIntentGroupName. It shall return the genericPlacementIntent if its present.
func (c *GenericPlacementIntentClient) GetGenericPlacementIntent(ctx context.Context, i string, p string, ca string, v string, digName string) (GenericPlacementIntent, error) {
	key := GenericPlacementIntentKey{
		Name:         i,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		DigName:      digName,
	}

	result, err := db.DBconn.Find(ctx, c.storeName, key, c.tagMetaData)
	if err != nil {
		return GenericPlacementIntent{}, err
	}

	if len(result) == 0 {
		return GenericPlacementIntent{}, pkgerrors.New("GenericPlacementIntent not found")
	}

	if result != nil {
		g := GenericPlacementIntent{}
		err = db.DBconn.Unmarshal(result[0], &g)
		if err != nil {
			return GenericPlacementIntent{}, err
		}
		return g, nil
	}

	return GenericPlacementIntent{}, pkgerrors.New("Unknown Error")

}

// GetAllGenericPlacementIntents returns all the generic placement intents for a given compsoite app name, composite app version, project and deploymentIntentGroupName
func (c *GenericPlacementIntentClient) GetAllGenericPlacementIntents(ctx context.Context, p string, ca string, v string, digName string) ([]GenericPlacementIntent, error) {

	//Check if project exists
	_, err := NewProjectClient().GetProject(ctx, p)
	if err != nil {
		return []GenericPlacementIntent{}, pkgerrors.Wrap(err, "Project not found")
	}

	// check if compositeApp exists
	_, err = NewCompositeAppClient().GetCompositeApp(ctx, ca, v, p)
	if err != nil {
		return []GenericPlacementIntent{}, err
	}

	key := GenericPlacementIntentKey{
		Name:         "",
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		DigName:      digName,
	}

	var gpList []GenericPlacementIntent
	values, err := db.DBconn.Find(ctx, c.storeName, key, c.tagMetaData)
	if err != nil {
		return []GenericPlacementIntent{}, err
	}

	for _, value := range values {
		gp := GenericPlacementIntent{}
		err = db.DBconn.Unmarshal(value, &gp)
		if err != nil {
			return []GenericPlacementIntent{}, err
		}
		gpList = append(gpList, gp)
	}

	return gpList, nil

}

// DeleteGenericPlacementIntent the intent from the database
func (c *GenericPlacementIntentClient) DeleteGenericPlacementIntent(ctx context.Context, i string, p string, ca string, v string, digName string) error {
	key := GenericPlacementIntentKey{
		Name:         i,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		DigName:      digName,
	}

	err := db.DBconn.Remove(ctx, c.storeName, key)
	return err
}
