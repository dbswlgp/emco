// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package db

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/net/context"

	"github.com/tidwall/gjson"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	utils "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/utils"

	pkgerrors "github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

// MongoCollection defines the a subset of MongoDB operations
// Note: This interface is defined mainly for mock testing
type MongoCollection interface {
	InsertOne(ctx context.Context, document interface{},
		opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{},
		opts ...*options.FindOneOptions) *mongo.SingleResult
	FindOneAndUpdate(ctx context.Context, filter interface{},
		update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult
	DeleteOne(ctx context.Context, filter interface{},
		opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	DeleteMany(ctx context.Context, filter interface{},
		opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	Find(ctx context.Context, filter interface{},
		opts ...*options.FindOptions) (*mongo.Cursor, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{},
		opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	CountDocuments(ctx context.Context, filter interface{},
		opts ...*options.CountOptions) (int64, error)
}

// MongoStore is an implementation of the db.Store interface
type MongoStore struct {
	db *mongo.Database
}

// This exists only for allowing us to mock the collection object
// for testing purposes
var getCollection = func(coll string, m *MongoStore) MongoCollection {
	return m.db.Collection(coll)
}

// This exists only for allowing us to mock the DecodeBytes function
// Mainly because we cannot construct a SingleResult struct from our
// tests. All fields in that struct are private.
var decodeBytes = func(sr *mongo.SingleResult) (bson.Raw, error) {
	return sr.DecodeBytes()
}

// These exists only for allowing us to mock the cursor.Next function
// Mainly because we cannot construct a mongo.Cursor struct from our
// tests. All fields in that struct are private and there is no public
// constructor method.
var cursorNext = func(ctx context.Context, cursor *mongo.Cursor) bool {
	return cursor.Next(ctx)
}
var cursorClose = func(ctx context.Context, cursor *mongo.Cursor) error {
	return cursor.Close(ctx)
}

// NewMongoStore initializes a Mongo Database with the name provided
// If a database with that name exists, it will be returned
func NewMongoStore(ctx context.Context, name string, store *mongo.Database) (Store, error) {
	if store == nil {
		ip := "mongodb://" + net.JoinHostPort(config.GetConfiguration().DatabaseIP, "31509")
		clientOptions := options.Client()
		clientOptions.Monitor = otelmongo.NewMonitor()
		clientOptions.ApplyURI(ip)
		if len(os.Getenv("DB_EMCO_USERNAME")) > 0 && len(os.Getenv("DB_EMCO_PASSWORD")) > 0 {
			clientOptions.SetAuth(options.Credential{
				AuthMechanism: "SCRAM-SHA-256",
				AuthSource:    "emco",
				Username:      os.Getenv("DB_EMCO_USERNAME"),
				Password:      os.Getenv("DB_EMCO_PASSWORD")})
		}
		mongoClient, err := mongo.NewClient(clientOptions)
		if err != nil {
			return nil, err
		}

		err = mongoClient.Connect(ctx)
		if err != nil {
			return nil, err
		}
		store = mongoClient.Database(name)
	}

	// make the MongoStore struct hear and then call schema stuff here
	mongoStore := &MongoStore{
		db: store,
	}

	go mongoStore.ReadRefSchema(ctx)

	return mongoStore, nil
}

// HealthCheck verifies if the database is up and running
func (m *MongoStore) HealthCheck(ctx context.Context) error {

	_, err := decodeBytes(m.db.RunCommand(ctx, bson.D{{"serverStatus", 1}}))
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting server status")
	}

	return nil
}

// findReferencedBys will search to see if this resource (identified by the key) is
// referenced by any other resources.
func (m *MongoStore) findReferencedBys(ctx context.Context, c MongoCollection, key Key) (int64, error) {

	// Create the key tag value for this resource (i.e. resource identifier)
	keyId, err := m.createKeyIdField(key)
	if err != nil {
		return 0, err
	}

	// set up the filter to search for this resource in references arrays in other documents
	filter, err := m.findRefByFilter(keyId, key)
	if err != nil {
		return 0, err
	}

	// search for documents with this resource in their resources list
	count, err := c.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return count, nil
	}

	return 0, nil
}

// findMapKeyValues will create a list of keys (with elements defined in "inKey") where the key elements that
// match "resName" will be filled in with the values of the keys from the map `mapName` in the "spec" object of "data"
func findMapKeyValues(inKey map[string]struct{}, mapName, resName string, data interface{}) ([]map[string]string, error) {
	var iterateSpec func(key, value gjson.Result) bool

	var targetMap gjson.Result

	iterateSpec = func(key, value gjson.Result) bool {
		if value.Type == gjson.JSON {
			if key.String() == mapName {
				targetMap = value
				return false
			}
		}
		return true
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Error marshalling data for key value search")
	}
	spec := gjson.GetBytes(jsonData, "spec")
	spec.ForEach(iterateSpec)

	results := make([]map[string]string, 0)
	for k, _ := range targetMap.Map() {
		m := make(map[string]string)
		for ki, _ := range inKey {
			if ki == resName {
				m[ki] = k
			} else {
				m[ki] = ""
			}
		}
		results = append(results, m)
	}
	return results, nil
}

// findManyKeyValues will scan the "spec" object inside "data" and return a list of
// keys (with elements defined in "inKey").  At each nesting level of the "spec" object,
// all string elements which elements in "inKey" will be used to create a key instance.
func findManyKeyValues(inKey map[string]struct{}, data interface{}) ([]map[string]string, error) {
	var iterateSpec func(key, value gjson.Result) bool

	maps := make(map[int]map[string]string)

	jsonObj := 0
	iterateSpec = func(key, value gjson.Result) bool {
		if value.Type == gjson.String {
			if _, ok := inKey[key.String()]; ok {
				var m map[string]string
				if m, ok = maps[jsonObj]; !ok {
					m = make(map[string]string)
				}
				m[key.String()] = value.String()
				maps[jsonObj] = m
			}
		}
		if value.Type == gjson.JSON {
			jsonObj++
			value.ForEach(iterateSpec)
		}
		return true
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Error marshalling data for key value search")
	}
	spec := gjson.GetBytes(jsonData, "spec")
	spec.ForEach(iterateSpec)

	results := make([]map[string]string, 0)
	for _, m := range maps {
		// fill out rest of the each result map with empty string values
		for k, _ := range inKey {
			if _, ok := m[k]; !ok {
				m[k] = ""
			}
		}
		results = append(results, m)
	}
	return results, nil
}

// findKeyValues will scan the "spec" object of "data" and create a key instance.
// Any element that matches an element in "inKey" and is also not present in "filterKey"
// will be added to the key.
func findKeyValues(inKey, filterKey map[string]struct{}, data interface{}) (map[string]string, error) {
	var iterateSpec func(key, value gjson.Result) bool

	result := make(map[string]string)

	iterateSpec = func(key, value gjson.Result) bool {
		if value.Type == gjson.String {
			if _, ok := inKey[key.String()]; ok {
				if _, fok := filterKey[key.String()]; !fok {
					result[key.String()] = value.String()
				}
			}
		}
		if value.Type == gjson.JSON {
			value.ForEach(iterateSpec)
		}
		return true
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Error marshalling data for key value search")
	}
	spec := gjson.GetBytes(jsonData, "spec")
	spec.ForEach(iterateSpec)

	// fill out rest of the new result key with empty string values
	for k, _ := range inKey {
		if _, ok := result[k]; !ok {
			result[k] = ""
		}
	}
	return result, nil
}

// verifyReferences checks that all references for a resource exist.
// 1. The parent resource, as defined by the schema, is checked.
// 2. The keys for other references, as identified for the schema, are found
//    by searching the "spec" object of the resource "data".
//    These references are then verified to exist.
func (m *MongoStore) verifyReferences(ctx context.Context, coll string, key Key, keyId string, data interface{}) ([]ReferenceEntry, error) {

	// make a references slice to store keys of any references found
	refs := make([]ReferenceEntry, 0)

	schemaLock.Lock()

	// Check if this item is present in the referential schema
	name, ok := refKeyMap[keyId]
//	fmt.Println(refKeyMap)
//	fmt.Println(keyId)
	fmt.Println(name)
	if !ok {
		schemaLock.Unlock()
		log.Info("Resource key ID is not present in referential schema", log.Fields{"keyId": keyId})
		return refs, pkgerrors.Errorf("Resource key ID is not present in referential schema. KeyID: %s, Key: %T %v", keyId, key, key)
	}

	resEntry, ok := refSchemaMap[name]
	if !ok {
		schemaLock.Unlock()
		log.Info("Resource is not present in referential schema", log.Fields{"name": name})
		return refs, pkgerrors.Errorf("Resource is not present in referential schema. Name: %s, KeyID: %s, Key: %T %v", name, keyId, key, key)
	}

	schemaLock.Unlock()

	// make a map[string]string copy of the key
	var rKey map[string]string
	st, err := json.Marshal(key)
	if err != nil {
		return refs, pkgerrors.Wrapf(err, "Error Marshalling key: %T %v", key, key)
	}

	err = json.Unmarshal([]byte(st), &rKey)
	if err != nil {
		return refs, pkgerrors.Wrapf(err, "Error Unmarshalling key to map. Key: %T %v", key, key)
	}

	// Check parent resource reference (if the resource has a parent)
	if len(resEntry.parent) > 0 {
		parentKey := make(map[string]string)

		// make the parent key
		for k, v := range rKey {
			if k == name {
				continue
			}
			if resEntry.trimParent && k == resEntry.parent {
				continue
			}
			parentKey[k] = v
		}

		// if no parent key is left, then no need to check for parent resource
		if len(parentKey) > 0 {
			// All resources should have a "data" element, so search for the parents "data"
			result, err := m.Find(ctx, coll, parentKey, "data")
			if err != nil {
				return refs, pkgerrors.Wrapf(err, "Error finding parent resource for %s. Parent: %T %v", name, parentKey, parentKey)
			}

			if len(result) == 0 {
				return refs, pkgerrors.Errorf("Parent resource not found for %s.  Parent: %T %v KeyID: %s, Key: %T %v", name, parentKey, parentKey, keyId, key, key)
			}
		}
	}

	// Collect the list of referenced resources
	for _, r := range resEntry.references {
		keys := make([]Key, 0)
		refKey := refSchemaMap[r.Name].keyMap

		switch r.Type {
		case "map":
			manyKeys, err := findMapKeyValues(refKey, r.Map, r.Name, data)
			if err != nil {
				return refs, err
			}
			for _, nk := range manyKeys {
				// fill in any fixed entries as defined by the referential schema
				for k, v := range r.FixedKv {
					nk[k] = v
				}
				// check if this key should be filtered
				filter := false
				for _, f := range r.FilterKeys {
					for k, v := range nk {
						if k == r.Name && v == f {
							filter = true
							break
						}
					}
				}
				if !filter {
					keys = append(keys, nk)
				}
			}
		case "many":
			manyKeys, err := findManyKeyValues(refKey, data)
			if err != nil {
				return refs, err
			}
			for _, nk := range manyKeys {
				// fill in any fixed entries as defined by the referential schema
				for k, v := range r.FixedKv {
					nk[k] = v
				}
				// fill out rest of key with this resource key
				// if items in key are not found, then drop this reference key
				// (only keep fully populated keys)
				fullKey := true
				for k, v := range nk {
					if v == "" {
						if _, ok := rKey[k]; !ok {
							fullKey = false
						}
						nk[k] = rKey[k]
					}
				}
				if fullKey {
					// check if nk is already in the list (prevent duplicates)
					found := false
					for _, m := range keys {
						found = reflect.DeepEqual(m, nk)
						if found {
							break
						}
					}
					if !found {
						keys = append(keys, nk)
					}
				}
			}
		default:
			// prepare a filter key (items to not fill out if found in the "spec" object)
			var filterKey map[string]struct{}
			if cResEntry, ok := refSchemaMap[r.CommonKey]; ok {
				filterKey = cResEntry.keyMap
			} else {
				filterKey = make(map[string]struct{})
			}

			nk, err := findKeyValues(refKey, filterKey, data)
			if err != nil {
				return refs, err
			}
			// fill in any fixed entries as defined by the referential schema
			for k, v := range r.FixedKv {
				nk[k] = v
			}
			// fill out rest of the reference key with elements from the current resource key
			fullKey := true
			for k, v := range nk {
				if v == "" {
					if _, ok := rKey[k]; !ok {
						fullKey = false
						log.Info("Reference key element not found", log.Fields{"resource": name, "key": nk, "element": k})
					}
					nk[k] = rKey[k]
				}
			}
			if fullKey {
				keys = append(keys, nk)
			}
		}

		for _, k := range keys {
			ref := ReferenceEntry{Key: k, KeyId: refSchemaMap[r.Name].keyId}
			refs = append(refs, ref)
		}
	}

	// Verify that referenced resources exist
	for _, ref := range refs {
		result, err := m.Find(ctx, coll, ref.Key, "data")
		if err != nil {
			log.Warn("Error finding resource reference", log.Fields{"resource": name, "referenceKey": ref.Key})
			/* For now, just log a warning if there was an error finding the referenced resource.
			 * return refs, pkgerrors.Errorf("Error finding referenced resource: [%v] for [%s]", ref.KeyId, name)
			 */
		} else if len(result) == 0 {
			log.Warn("Resource reference not found", log.Fields{"resource": name, "referenceKey": ref.Key})
			/* For now, just log a warning if the referenced resource does not exist.
			 * return refs, pkgerrors.New("Referenced resource not found: [" + ref.KeyId + "] for [" + name + "]")
			 */
		}
	}

	return refs, nil
}

// validateParams checks to see if any parameters are empty
func (m *MongoStore) validateParams(args ...interface{}) bool {
	for _, v := range args {
		val, ok := v.(string)
		if ok {
			if val == "" {
				return false
			}
		} else {
			if v == nil {
				return false
			}
		}
	}

	return true
}

// Unmarshal implements an unmarshaler for bson data that
// is produced from the mongo database
func (m *MongoStore) Unmarshal(inp []byte, out interface{}) error {
	err := bson.Unmarshal(inp, out)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error Unmarshalling bson data to %T", out)
	}

	// Decrypt data if required
	oe := utils.GetObjectEncryptor("emco")
	if oe != nil {
		_, err := oe.DecryptObject(out)
		if err != nil {
			log.Warn("Error to decrypt object", log.Fields{"error": err.Error()})
		}
	}
	return nil
}

func (m *MongoStore) findFilter(key Key) (primitive.M, error) {

	var bsonMap bson.M
	st, err := json.Marshal(key)
	if err != nil {
		return primitive.M{}, pkgerrors.Errorf("Error Marshalling key: %s", err.Error())
	}
	err = json.Unmarshal([]byte(st), &bsonMap)
	if err != nil {
		return primitive.M{}, pkgerrors.Errorf("Error Unmarshalling key to Bson Map: %s", err.Error())
	}
	filter := bson.M{
		"$and": []bson.M{bsonMap},
	}
	return filter, nil
}

// findRefByFilter creates a filter based on the key and keyId of a resource
// that can match an element in the "references" list.
func (m *MongoStore) findRefByFilter(keyId string, key Key) (primitive.M, error) {

	var bsonMap bson.M
	var bsonMapFinal bson.M
	st, err := json.Marshal(key)
	if err != nil {
		return primitive.M{}, pkgerrors.Errorf("Error Marshalling key: %s", err.Error())
	}
	err = json.Unmarshal([]byte(st), &bsonMap)
	if err != nil {
		return primitive.M{}, pkgerrors.Errorf("Error Unmarshalling key to Bson Map: %s", err.Error())
	}
	bsonMapFinal = make(bson.M)
	for k, v := range bsonMap {
		if v != "" {
			bsonMapFinal["key."+k] = v
		}
	}
	bsonMapFinal["keyid"] = keyId
	filter := bson.M{"references": bson.M{"$elemMatch": bsonMapFinal}}
	return filter, nil
}

func (m *MongoStore) findFilterWithKey(key Key) (primitive.M, error) {

	var bsonMap bson.M
	var bsonMapFinal bson.M
	st, err := json.Marshal(key)
	if err != nil {
		return primitive.M{}, pkgerrors.Errorf("Error Marshalling key: %s", err.Error())
	}
	err = json.Unmarshal([]byte(st), &bsonMap)
	if err != nil {
		return primitive.M{}, pkgerrors.Errorf("Error Unmarshalling key to Bson Map: %s", err.Error())
	}
	bsonMapFinal = make(bson.M)
	for k, v := range bsonMap {
		if v == "" {
			if _, ok := bsonMapFinal["keyId"]; !ok {
				// add type of key to filter
				keyId, err := m.createKeyIdField(key)
				if err != nil {
					return primitive.M{}, err
				}
				bsonMapFinal["keyId"] = keyId
			}
		} else {
			bsonMapFinal[k] = v
		}
	}
	filter := bson.M{
		"$and": []bson.M{bsonMapFinal},
	}
	return filter, nil
}

func (m *MongoStore) updateFilter(key interface{}) (primitive.M, error) {

	var n map[string]string

	st, err := json.Marshal(key)
	if err != nil {
		return primitive.M{}, pkgerrors.Wrapf(err, "Error Marshalling key: %T %v", key, key)
	}

	err = json.Unmarshal([]byte(st), &n)
	if err != nil {
		return primitive.M{}, pkgerrors.Wrapf(err, "Error Unmarshalling key to Bson Map. Key: %T %v", key, key)
	}

	p := make(bson.M, len(n))
	for k, v := range n {
		p[k] = v
	}

	filter := bson.M{
		"$set": p,
	}
	return filter, nil
}

func (m *MongoStore) createKeyIdField(key interface{}) (string, error) {

	var n map[string]string
	st, err := json.Marshal(key)
	if err != nil {
		return "", pkgerrors.Errorf("Error Marshalling key: %s", err.Error())
	}
	err = json.Unmarshal([]byte(st), &n)
	if err != nil {
		return "", pkgerrors.Errorf("Error Unmarshalling key to Bson Map: %s", err.Error())
	}
	var keys []string
	for k := range n {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return fmt.Sprintf("{%s}", strings.Join(keys, ",")), nil
}

// migration 시에 특정 필드 값을 변경하기 위해 추가한 함수
func (m *MongoStore) UpdateIntentField(ctx context.Context, coll string, key Key, fieldToUpdate string, newValue string) error {
        c := getCollection(coll, m)

        filter, err := m.findFilter(key)
	fmt.Println("\n")
	fmt.Println("UpdateIntentField.filter: ",filter)
	fmt.Println("\n")
        if err != nil {
                return pkgerrors.Wrapf(err, "db UpdateField error: Error finding filter with key %T %v", key, key)
        }


	var result bson.M
	c.FindOne(context.TODO(), filter).Decode(&result)


	existvalue, _ := getNestedField(result, "data", "metadata", "description")

	values := strings.Split(existvalue.(string), ",")

	var updatevalue string

	if newValue == "" {
		for i, v := range values[:len(values)-1] {
			updatevalue = updatevalue + v
			if i != len(values)-2 {
				updatevalue = updatevalue + ","
			}
		}
	} else {

		// newValue가 values 배열에 있는지 확인
		found := false
		for _, v := range values {
			if v == newValue {
				found = true
				break
			}
		}

		if found == false {
			updatevalue = existvalue.(string) + "," + newValue
		} else {
			updatevalue = existvalue.(string)
		}
	}


	// Update field with the new value
	update := bson.D{{"$set", bson.D{{fieldToUpdate, updatevalue}}}}
	fmt.Println("UpdateIntentField.update: ",update)

	_, err = c.UpdateOne(ctx, filter, update)
	if err != nil {
		return pkgerrors.Wrapf(err, "db UpdateField error: Error updating field %s in collection %s", fieldToUpdate, coll)
	}

        return nil
}

func getNestedField(data bson.M, keys ...string) (interface{}, error) {
	var current interface{} = data

	for _, key := range keys {
		switch currentValue := current.(type) {
		case bson.M:
			val, found := currentValue[key]
			if !found {
				return nil, fmt.Errorf("field '%s' not found", key)
			}
			current = val
		default:
			return nil, fmt.Errorf("unexpected type at field '%s'", key)
		}
	}

	return current, nil
}


// Insert is used to insert/add element to a document
func (m *MongoStore) Insert(ctx context.Context, coll string, key Key, query interface{}, tag string, data interface{}) error {
/*
	fmt.Println("\n")
	fmt.Println("orchestrator/pkg/infra/db/mongo.go")
	fmt.Println("coll: ",coll)
	fmt.Println("key: ",key)
	fmt.Println("query: ",query)
	fmt.Println("tag: ",tag)
	fmt.Println("data: ",data)
	fmt.Println("\n")
*/
	if data == nil {
		return pkgerrors.Errorf("db Insert error: No data to store")
	}

	if !m.validateParams(coll, key, tag) {
		return pkgerrors.Errorf("db Insert error: Mandatory fields are missing. Collection: %s, Key: %T %v, Tag: %s", coll, key, key, tag)
	}

	c := getCollection(coll, m)

	filter, err := m.findFilter(key)
	if err != nil {
		return pkgerrors.Wrapf(err, "db Insert error: Error finding filter with key %T %v", key, key)
	}

	// Create and add keyId tag
	keyId, err := m.createKeyIdField(key)
	if err != nil {
		return pkgerrors.Wrapf(err, "db Insert error: Error creating KeyID with key %T %v", key, key)
	}

	// Encrypt data if required
	oe := utils.GetObjectEncryptor("emco")
	if oe != nil {
		var edata interface{}
		if reflect.TypeOf(data).Kind() == reflect.Ptr {
			// avoid changing data's field value during encryption
			edata, err = oe.EncryptObject(reflect.ValueOf(data).Elem().Interface())
		} else {
			edata, err = oe.EncryptObject(data)
		}

		if err == nil {
			data = edata
		} else {
			log.Warn("Error to encrypt object", log.Fields{"collection": coll, "tag": tag})
		}
	}

	// verify references for Inserts with the "data" tag
	refs := make([]ReferenceEntry, 0)

	if tag == "data" {
		refs, err = m.verifyReferences(ctx, coll, key, keyId, data)
		if err != nil {
			if strings.Contains(err.Error(), "Parent resource not found") {
				// these errors should be handled separately, not as an internal server error
				return pkgerrors.Wrapf(err, "db Insert parent resource not found")
			}

			if strings.Contains(err.Error(), "is not present in referential schema") {
				// these errors should be handled separately, not as an internal server error
				return pkgerrors.Wrapf(err, "db Insert referential schema missing")

			}

			return pkgerrors.Wrapf(err, "db Insert error: Error verifying the references. Collection: %s, Key: %T %v, KeyID: %s", coll, key, key, keyId)
		}

		_, err = decodeBytes(
			c.FindOneAndUpdate(
				ctx,
				filter,
				bson.D{
					{"$set", bson.D{
						{tag, data},
						{"keyId", keyId},
						{"references", refs},
					}},
				},
				options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)))
	} else {
		_, err = decodeBytes(
			c.FindOneAndUpdate(
				ctx,
				filter,
				bson.D{
					{"$set", bson.D{
						{tag, data},
						{"keyId", keyId},
					}},
				},
				options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)))
	}

	if err != nil {
		return pkgerrors.Wrapf(err, "db Insert error")
	}

	if query == nil {
		return nil
	}

	// Update to add Query fields
	update, err := m.updateFilter(query)
	if err != nil {
		return pkgerrors.Wrapf(err, "db Insert error: Error updating filter with query %T %v", query, query)
	}

	_, err = c.UpdateOne(
		ctx,
		filter,
		update)
	if err != nil {
		return pkgerrors.Wrapf(err, "db Insert error")
	}

	return nil
}

// Find method returns the data stored for this key and for this particular tag
func (m *MongoStore) Find(ctx context.Context, coll string, key Key, tag string) ([][]byte, error) {

	//result, err := m.findInternal(coll, key, tag, "")
	//return result, err
	if !m.validateParams(coll, key, tag) {
		return nil, pkgerrors.Errorf("db Find error: Mandatory fields are missing. Collection: %s, Key: %T %v, Tag: %s", coll, key, key, tag)
	}

	c := getCollection(coll, m)

	filter, err := m.findFilterWithKey(key)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "db Find error: Error finding filter with key %T %v", key, key)
	}
	// Find only the field requested
	projection := bson.D{
		{tag, 1},
		{"_id", 0},
	}

	cursor, err := c.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, pkgerrors.Wrap(err, "db Find error")
	}
	defer cursorClose(ctx, cursor)
	var data []byte
	var result [][]byte
	for cursorNext(ctx, cursor) {
		d := cursor.Current
		switch d.Lookup(tag).Type {
		case bson.TypeString:
			data = []byte(d.Lookup(tag).StringValue())
		default:
			r, err := d.LookupErr(tag)
			if err != nil {
				// Throw error if not found
				pkgerrors.New("db Find error: Unable to read data")
			}
			data = r.Value
		}
		result = append(result, data)
	}
	return result, nil
}

// RemoveAll method to removes all the documet matching key
func (m *MongoStore) RemoveAll(ctx context.Context, coll string, key Key) error {
	if !m.validateParams(coll, key) {
		return pkgerrors.Errorf("db Remove error: Mandatory fields are missing. Collection: %s, Key: %T %v", coll, key, key)
	}
	c := getCollection(coll, m)
	filter, err := m.findFilterWithKey(key)
	if err != nil {
		return pkgerrors.Wrapf(err, "db Remove error: Error finding filter with key %T %v", key, key)
	}
	_, err = c.DeleteMany(ctx, filter)
	if err != nil {
		return pkgerrors.Wrapf(err, "db Remove error: Error deleting document(s) from database. Key: %T %v, Filter: %v", key, key, filter)
	}
	return nil
}

// Remove method to remove the documet by key if no child references
func (m *MongoStore) Remove(ctx context.Context, coll string, key Key) error {
	if !m.validateParams(coll, key) {
		return pkgerrors.Errorf("db Remove error: Mandatory fields are missing. Collection: %s, Key: %T %v", coll, key, key)
	}

	// search for child references - assumes all children are part of the
	// same collection
	c := getCollection(coll, m)
	filter, err := m.findFilter(key)
	if err != nil {
		return pkgerrors.Wrapf(err, "db Remove error: Error finding filter with key %T %v", key, key)
	}

	count, err := c.CountDocuments(ctx, filter)
	if err != nil {
		return pkgerrors.Wrap(err, "db Remove error")
	}

	if count == 0 {
		return pkgerrors.Errorf("db Remove resource not found: The requested resource not found. Key: %T %v", key, key)
	}

	if count > 1 {
		return pkgerrors.Errorf("db Remove parent child constraint: Cannot delete parent without deleting child references first. Key: %T %v", key, key)
	}

	// search to see if this document is referenced by any other document
	count, err = m.findReferencedBys(ctx, c, key)
	if err != nil {
		return pkgerrors.Wrapf(err, "db Remove error: Error finding referencing resources for key %T %v", key, key)
	}

	if count > 0 {
		return pkgerrors.Errorf("db Remove referential constraint: Cannot delete without deleting or updating referencing resources first. Key: %T %v", key, key)
	}

	// ok to delete the document
	_, err = c.DeleteOne(ctx, filter)
	if err != nil {
		return pkgerrors.Wrapf(err, "db Remove error: Error deleting document from database. Key: %T %v, Filter: %v", key, key, filter)
	}
	return nil
}

// RemoveTag is used to remove an element from a document
func (m *MongoStore) RemoveTag(ctx context.Context, coll string, key Key, tag string) error {
	c := getCollection(coll, m)

	filter, err := m.findFilter(key)
	if err != nil {
		return err
	}

	_, err = decodeBytes(
		c.FindOneAndUpdate(
			ctx,
			filter,
			bson.D{
				{"$unset", bson.D{
					{tag, ""},
				}},
			},
			options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)))

	if err != nil {
		return pkgerrors.Errorf("Error removing tag: %s", err.Error())
	}

	return nil
}
