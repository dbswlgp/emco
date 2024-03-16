// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// KeyValue contains the parameters needed for a key value
type KeyValue struct {
	MetaData      KVMetaDataList `json:"metadata"`
	Specification KVSpec         `json:"spec"`
}

// MetaData contains the parameters needed for metadata
type KVMetaDataList struct {
	KeyValueName string `json:"name"`
	Description  string `json:"description"`
	UserData1    string `json:"userData1"`
	UserData2    string `json:"userData2"`
}

// Spec contains the parameters needed for spec
type KVSpec struct {
	Kv []map[string]interface{} `json:"kv"`
}

// KeyValueKey is the key structure that is used in the database
type KeyValueKey struct {
	Project          string `json:"project"`
	LogicalCloudName string `json:"logicalCloud"`
	KeyValueName     string `json:"logicalCloudKv"`
}

// KeyValueManager is an interface that exposes the connection
// functionality
type KeyValueManager interface {
	CreateKVPair(ctx context.Context, project, logicalCloud string, c KeyValue) (KeyValue, error)
	GetKVPair(ctx context.Context, project, logicalCloud, name string) (KeyValue, error)
	GetAllKVPairs(ctx context.Context, project, logicalCloud string) ([]KeyValue, error)
	DeleteKVPair(ctx context.Context, project, logicalCloud, name string) error
	UpdateKVPair(ctx context.Context, project, logicalCloud, name string, c KeyValue) (KeyValue, error)
}

// KeyValueClient implements the KeyValueManager
// It will also be used to maintain some localized state
type KeyValueClient struct {
	storeName string
	tagMeta   string
}

// KeyValueClient returns an instance of the KeyValueClient
// which implements the KeyValueManager
func NewKeyValueClient() *KeyValueClient {
	return &KeyValueClient{
		storeName: "resources",
		tagMeta:   "data",
	}
}

// Create entry for the key value resource in the database
func (v *KeyValueClient) CreateKVPair(ctx context.Context, project, logicalCloud string, c KeyValue) (KeyValue, error) {

	//Construct key consisting of name
	key := KeyValueKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		KeyValueName:     c.MetaData.KeyValueName,
	}

	//Check if this Key Value already exists
	_, err := v.GetKVPair(ctx, project, logicalCloud, c.MetaData.KeyValueName)
	if err == nil {
		return KeyValue{}, pkgerrors.New("Key Value already exists")
	}

	err = db.DBconn.Insert(ctx, v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return KeyValue{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return c, nil
}

// Get returns Key Value for correspondin name
func (v *KeyValueClient) GetKVPair(ctx context.Context, project, logicalCloud, kvPairName string) (KeyValue, error) {

	//Construct the composite key to select the entry
	key := KeyValueKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		KeyValueName:     kvPairName,
	}
	value, err := db.DBconn.Find(ctx, v.storeName, key, v.tagMeta)
	if err != nil {
		return KeyValue{}, err
	}

	if len(value) == 0 {
		return KeyValue{}, pkgerrors.New("Key Value not found")
	}

	//value is a byte array
	if value != nil {
		kv := KeyValue{}
		err = db.DBconn.Unmarshal(value[0], &kv)
		if err != nil {
			return KeyValue{}, err
		}
		return kv, nil
	}

	return KeyValue{}, pkgerrors.New("Unknown Error")
}

// Get All lists all key value pairs
func (v *KeyValueClient) GetAllKVPairs(ctx context.Context, project, logicalCloud string) ([]KeyValue, error) {

	//Construct the composite key to select the entry
	key := KeyValueKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		KeyValueName:     "",
	}
	var resp []KeyValue
	values, err := db.DBconn.Find(ctx, v.storeName, key, v.tagMeta)
	if err != nil {
		return []KeyValue{}, err
	}

	for _, value := range values {
		kv := KeyValue{}
		err = db.DBconn.Unmarshal(value, &kv)
		if err != nil {
			return []KeyValue{}, err
		}
		resp = append(resp, kv)
	}

	return resp, nil
}

// Delete the Key Value entry from database
func (v *KeyValueClient) DeleteKVPair(ctx context.Context, project, logicalCloud, kvPairName string) error {

	//Construct the composite key to select the entry
	key := KeyValueKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		KeyValueName:     kvPairName,
	}
	err := db.DBconn.Remove(ctx, v.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Key Value")
	}
	return nil
}

// Update an entry for the Key Value in the database
func (v *KeyValueClient) UpdateKVPair(ctx context.Context, project, logicalCloud, kvPairName string, c KeyValue) (KeyValue, error) {

	key := KeyValueKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		KeyValueName:     kvPairName,
	}
	//Check if KV pair URl name is the same name in json
	if c.MetaData.KeyValueName != kvPairName {
		return KeyValue{}, pkgerrors.New("Update Error - KV pair name mismatch")
	}
	//Check if this Key Value exists
	_, err := v.GetKVPair(ctx, project, logicalCloud, kvPairName)
	if err != nil {
		return KeyValue{}, err
	}
	err = db.DBconn.Insert(ctx, v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return KeyValue{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}
