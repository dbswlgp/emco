// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	"context"
	pkgerrors "github.com/pkg/errors"
)

// NetControlIntent contains the parameters needed for dynamic networks
type NetControlIntent struct {
	Metadata Metadata `json:"metadata"`
}

// NetControlIntentKey is the key structure that is used in the database
type NetControlIntentKey struct {
	NetControlIntent    string `json:"netControllerIntent"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	DigName             string `json:"deploymentIntentGroup"`
}

// Manager is an interface exposing the NetControlIntent functionality
type NetControlIntentManager interface {
	CreateNetControlIntent(ctx context.Context, nci NetControlIntent, project, compositeapp, compositeappversion, dig string, exists bool) (NetControlIntent, error)
	GetNetControlIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig string) (NetControlIntent, error)
	GetNetControlIntents(ctx context.Context, project, compositeapp, compositeappversion, dig string) ([]NetControlIntent, error)
	DeleteNetControlIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig string) error
}

// NetControlIntentClient implements the Manager
// It will also be used to maintain some localized state
type NetControlIntentClient struct {
	db ClientDbInfo
}

// NewNetControlIntentClient returns an instance of the NetControlIntentClient
// which implements the Manager
func NewNetControlIntentClient() *NetControlIntentClient {
	return &NetControlIntentClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}

// CreateNetControlIntent - create a new NetControlIntent
func (v *NetControlIntentClient) CreateNetControlIntent(ctx context.Context, nci NetControlIntent, project, compositeapp, compositeappversion, dig string, exists bool) (NetControlIntent, error) {

	//Construct key and tag to select the entry
	key := NetControlIntentKey{
		NetControlIntent:    nci.Metadata.Name,
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
	}

	//Check if this NetControlIntent already exists
	_, err := v.GetNetControlIntent(ctx, nci.Metadata.Name, project, compositeapp, compositeappversion, dig)
	if err == nil && !exists {
		return NetControlIntent{}, pkgerrors.New("NetControlIntent already exists")
	}

	err = db.DBconn.Insert(ctx, v.db.storeName, key, nil, v.db.tagMeta, nci)
	if err != nil {
		return NetControlIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return nci, nil
}

// GetNetControlIntent returns the NetControlIntent for corresponding name
func (v *NetControlIntentClient) GetNetControlIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig string) (NetControlIntent, error) {

	//Construct key and tag to select the entry
	key := NetControlIntentKey{
		NetControlIntent:    name,
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
	}

	value, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return NetControlIntent{}, err
	} else if len(value) == 0 {
		return NetControlIntent{}, pkgerrors.New("Net Control Intent not found")
	}

	//value is a byte array
	if value != nil {
		nci := NetControlIntent{}
		err = db.DBconn.Unmarshal(value[0], &nci)
		if err != nil {
			return NetControlIntent{}, err
		}
		return nci, nil
	}

	return NetControlIntent{}, pkgerrors.New("Unknown Error")
}

// GetNetControlIntentList returns all of the NetControlIntent for corresponding name
func (v *NetControlIntentClient) GetNetControlIntents(ctx context.Context, project, compositeapp, compositeappversion, dig string) ([]NetControlIntent, error) {

	//Construct key and tag to select the entry
	key := NetControlIntentKey{
		NetControlIntent:    "",
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
	}

	var resp []NetControlIntent
	values, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []NetControlIntent{}, err
	}

	for _, value := range values {
		nci := NetControlIntent{}
		err = db.DBconn.Unmarshal(value, &nci)
		if err != nil {
			return []NetControlIntent{}, err
		}
		resp = append(resp, nci)
	}

	return resp, nil
}

// Delete the  NetControlIntent from database
func (v *NetControlIntentClient) DeleteNetControlIntent(ctx context.Context, name, project, compositeapp, compositeappversion, dig string) error {

	//Construct key and tag to select the entry
	key := NetControlIntentKey{
		NetControlIntent:    name,
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
	}

	err := db.DBconn.Remove(ctx, v.db.storeName, key)
	return err
}
