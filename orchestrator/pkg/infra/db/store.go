// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package db

import (
	"encoding/json"
	"reflect"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	"golang.org/x/net/context"

	pkgerrors "github.com/pkg/errors"
)

// DBconn interface used to talk a concrete Database connection
var DBconn Store

// Key is an interface that will be implemented by anypackage
// that wants to use the Store interface. This allows various
// db backends and key types.
type Key interface {
}

// Store is an interface for accessing the database
type Store interface {
	// Returns nil if db health is good
	HealthCheck(ctx context.Context) error

	// Unmarshal implements any unmarshalling needed for the database
	Unmarshal(inp []byte, out interface{}) error

	// Inserts and Updates a tag with key and also adds query fields if provided
	Insert(ctx context.Context, coll string, key Key, query interface{}, tag string, data interface{}) error

	// Find the document(s) with key and get the tag values from the document(s)
	Find(ctx context.Context, coll string, key Key, tag string) ([][]byte, error)

	// Removes the document(s) matching the key if no child reference in collection
	Remove(ctx context.Context, coll string, key Key) error

	// Remove all the document(s) matching the key
	RemoveAll(ctx context.Context, coll string, key Key) error

	// Remove the specifiec tag from the document matching the key
	RemoveTag(ctx context.Context, coll string, key Key, tag string) error

	UpdateIntentField(ctx context.Context, coll string, key Key, fieldToUpdate string, newValue string) error
}

// CreateDBClient creates the DB client
func createDBClient(ctx context.Context, dbType string, dbName string) error {
	var err error
	switch dbType {
	case "mongo":
		// create a mongodb database with orchestrator as the name
		DBconn, err = NewMongoStore(ctx, dbName, nil)
	default:
		return pkgerrors.New(dbType + "DB not supported")
	}
	return err
}

// Serialize converts given data into a JSON string
func Serialize(v interface{}) (string, error) {
	out, err := json.Marshal(v)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error serializing "+reflect.TypeOf(v).String())
	}
	return string(out), nil
}

// DeSerialize converts string to a json object specified by type
func DeSerialize(str string, v interface{}) error {
	err := json.Unmarshal([]byte(str), &v)
	if err != nil {
		return pkgerrors.Wrap(err, "Error deSerializing "+str)
	}
	return nil
}

// InitializeDatabaseConnection sets up the connection to the
// configured database to allow the application to talk to it.
func InitializeDatabaseConnection(ctx context.Context, dbName string) error {
	err := createDBClient(ctx, config.GetConfiguration().DatabaseType, dbName)
	if err != nil {
		return pkgerrors.Cause(err)
	}

	err = DBconn.HealthCheck(ctx)
	if err != nil {
		return pkgerrors.Cause(err)
	}

	return nil
}
