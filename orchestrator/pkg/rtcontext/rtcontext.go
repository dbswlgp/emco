// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package rtcontext

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

const maxrand = 0x7fffffffffffffff
const prefix string = "/context/"

type RunTimeContext struct {
	cid  interface{}
	meta interface{}
}

type Rtcontext interface {
	RtcInit() (interface{}, error)
	RtcInitWithValue(value interface{}) (interface{}, error)
	RtcLoad(ctx context.Context, id interface{}) (interface{}, error)
	RtcCreate(ctx context.Context) (interface{}, error)
	RtcAddMeta(ctx context.Context, meta interface{}) error
	RtcGet(ctx context.Context) (interface{}, error)
	RtcAddLevel(ctx context.Context, handle interface{}, level string, value string) (interface{}, error)
	RtcAddResource(ctx context.Context, handle interface{}, resname string, value interface{}) (interface{}, error)
	RtcAddInstruction(ctx context.Context, handle interface{}, level string, insttype string, value interface{}) (interface{}, error)
	RtcDeletePair(ctx context.Context, handle interface{}) error
	RtcDeletePrefix(ctx context.Context, handle interface{}) error
	RtcGetHandles(ctx context.Context, handle interface{}) ([]interface{}, error)
	RtcGetValue(ctx context.Context, handle interface{}, value interface{}) error
	RtcUpdateValue(ctx context.Context, handle interface{}, value interface{}) error
	RtcGetMeta(ctx context.Context) (interface{}, error)
	RtcAddOneLevel(ctx context.Context, pl interface{}, level string, value interface{}) (interface{}, error)
}

//Intialize context by assiging a new id
func (rtc *RunTimeContext) RtcInit() (interface{}, error) {
	if rtc.cid != nil {
		return nil, pkgerrors.Errorf("Error, context already initialized")
	}
	ra := rand.New(rand.NewSource(time.Now().UnixNano()))
	rn := ra.Int63n(maxrand)
	id := fmt.Sprintf("%v", rn)
	cid := (prefix + id + "/")
	rtc.cid = interface{}(cid)
	return interface{}(id), nil
}

//Intialize context by assiging a new id
func (rtc *RunTimeContext) RtcInitWithValue(value interface{}) (interface{}, error) {
	if rtc.cid != nil {
		return nil, pkgerrors.Errorf("Error, context already initialized")
	}
	//Check if input is of type int or int64
	xType := fmt.Sprintf("%T", value)
	if xType != "int64" {
		return nil, pkgerrors.Errorf("Valid values are of type int64: Not valid type %v", xType)
	}

	id := fmt.Sprintf("%v", value)
	cid := (prefix + id + "/")
	rtc.cid = interface{}(cid)

	return interface{}(id), nil
}

//Load context using the given id
func (rtc *RunTimeContext) RtcLoad(ctx context.Context, id interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", id)
	if str == "" {
		return nil, pkgerrors.Errorf("Not a valid context id")
	}
	cid := (prefix + str + "/")
	rtc.cid = interface{}(cid)
	handle, err := rtc.RtcGet(ctx)
	if err != nil {
		return nil, pkgerrors.Errorf("Error finding the context id: %s", err.Error())
	}
	return handle, nil
}

func (rtc *RunTimeContext) RtcCreate(ctx context.Context) (interface{}, error) {
	cid := fmt.Sprintf("%v", rtc.cid)
	if cid == "" {
		return nil, pkgerrors.Errorf("Error, context not intialized")
	}
	if !strings.HasPrefix(cid, prefix) {
		return nil, pkgerrors.Errorf("Not a valid run time context prefix")
	}
	id := strings.SplitN(cid, "/", 4)[2]
	// Create context only if context doesn't exist
	// Returns error if id is in database
	err := contextdb.Db.PutWithCheck(ctx, cid, id)
	if err != nil {
		return nil, pkgerrors.Errorf("Error creating run time context: %s", err.Error())
	}

	return rtc.cid, nil
}

//RtcAddMeta is used for saving meta data of appContext into ETCD.
func (rtc *RunTimeContext) RtcAddMeta(ctx context.Context, meta interface{}) error {
	cid := fmt.Sprintf("%v", rtc.cid)
	if cid == "" {
		return pkgerrors.Errorf("Error, context not intialized")
	}
	if !strings.HasPrefix(cid, prefix) {
		return pkgerrors.Errorf("Not a valid run time context prefix")
	}

	rtc.meta = meta
	k := cid + "meta" + "/"
	err := contextdb.Db.Put(ctx, k, rtc.meta)
	if err != nil {
		return pkgerrors.Errorf("Error saving metadata in run time context: %s", err.Error())
	}

	return nil
}

//Get the root handle
func (rtc *RunTimeContext) RtcGet(ctx context.Context) (interface{}, error) {
	str := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, prefix) {
		return nil, pkgerrors.Errorf("Not a valid run time context")
	}

	var value string
	err := contextdb.Db.Get(ctx, str, &value)
	if err != nil {
		return nil, pkgerrors.Errorf("Error getting run time context metadata: %s", err.Error())
	}
	if !strings.Contains(str, value) {
		return nil, pkgerrors.Errorf("Error matching run time context metadata")
	}

	return rtc.cid, nil
}

// RtcGetMeta method fetches the meta data of the rtc object and returns it.
func (rtc *RunTimeContext) RtcGetMeta(ctx context.Context) (interface{}, error) {
	str := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, prefix) {
		return nil, pkgerrors.Errorf("Not a valid run time context")
	}

	var value interface{}
	k := str + "meta" + "/"

	err := contextdb.Db.Get(ctx, k, &value)

	if err != nil {
		return nil, pkgerrors.Errorf("Error getting run time context metadata: %s", err.Error())
	}
	return value, nil

}

//Add a new level at a given handle and return the new handle
func (rtc *RunTimeContext) RtcAddLevel(ctx context.Context, handle interface{}, level string, value string) (interface{}, error) {
	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return nil, pkgerrors.Errorf("Not a valid run time context handle")
	}

	if level == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context level")
	}
	if value == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context level value")
	}

	key := str + level + "/" + value + "/"
	err := contextdb.Db.Put(ctx, key, value)
	if err != nil {
		return nil, pkgerrors.Errorf("Error adding run time context level: %s", err.Error())
	}

	return (interface{})(key), nil
}

// RtcAddOneLevel adds one more level to the existing context prefix.RtcAddOneLevel. It takes in PreviousContentLevel as inteface, new level to be appended as string and the value to be saved of any type. It returns the updated interface and nil if no error.
//
func (rtc *RunTimeContext) RtcAddOneLevel(ctx context.Context, pl interface{}, level string, value interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", pl)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return nil, pkgerrors.Errorf("Not a valid run time context handle")
	}

	if level == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context level")
	}
	if value == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context level value")
	}

	key := str + level + "/"
	err := contextdb.Db.Put(ctx, key, value)
	if err != nil {
		return nil, pkgerrors.Errorf("Error adding run time context level: %s", err.Error())
	}
	return (interface{})(key), nil
}

// Add a resource under the given level and return new handle
func (rtc *RunTimeContext) RtcAddResource(ctx context.Context, handle interface{}, resname string, value interface{}) (interface{}, error) {

	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return nil, pkgerrors.Errorf("Not a valid run time context handle")
	}
	if resname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context resource name")
	}
	if value == nil {
		return nil, pkgerrors.Errorf("Not a valid run time context resource value")
	}

	k := str + "resource" + "/" + resname + "/"
	err := contextdb.Db.Put(ctx, k, value)
	if err != nil {
		return nil, pkgerrors.Errorf("Error adding run time context resource: %s", err.Error())
	}
	return (interface{})(k), nil
}

// Add instruction at a given level and type, return the new handle
func (rtc *RunTimeContext) RtcAddInstruction(ctx context.Context, handle interface{}, level string, insttype string, value interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return nil, pkgerrors.Errorf("Not a valid run time context handle")
	}

	if level == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context level")
	}
	if insttype == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context instruction type")
	}
	if value == nil {
		return nil, pkgerrors.Errorf("Not a valid run time context instruction value")
	}
	k := str + level + "/" + "instruction" + "/" + insttype + "/"
	err := contextdb.Db.Put(ctx, k, fmt.Sprintf("%v", value))
	if err != nil {
		return nil, pkgerrors.Errorf("Error adding run time context instruction: %s", err.Error())
	}

	return (interface{})(k), nil
}

//Delete the key value pair using given handle
func (rtc *RunTimeContext) RtcDeletePair(ctx context.Context, handle interface{}) error {
	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return pkgerrors.Errorf("Not a valid run time context handle")
	}
	err := contextdb.Db.Delete(ctx, str)
	if err != nil {
		return pkgerrors.Errorf("Error deleting run time context pair: %s", err.Error())
	}

	return nil
}

// Delete all handles underneath the given handle
func (rtc *RunTimeContext) RtcDeletePrefix(ctx context.Context, handle interface{}) error {
	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)

	if !strings.HasPrefix(str, sid) {
		logutils.Error("Not a valid run time context handle", logutils.Fields{"key ::": str, "sid :: ": sid})
		return pkgerrors.Errorf("Not a valid run time context handle")
	}

	err := contextdb.Db.DeleteAll(ctx, str)

	if err != nil {
		return pkgerrors.Errorf("Error deleting run time context with prefix: %s", err.Error())
	}

	return nil
}

// Return the list of handles under the given handle
func (rtc *RunTimeContext) RtcGetHandles(ctx context.Context, handle interface{}) ([]interface{}, error) {
	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return nil, pkgerrors.Errorf("Not a valid run time context handle")
	}

	s, err := contextdb.Db.GetAllKeys(ctx, str)

	//fmt.Println("\n\n\n str:",str,"\n rtc:",rtc,"\n sid:",sid,"\n s:",s)

	if err != nil {

		return nil, pkgerrors.Errorf("Error getting run time context handles: %s", err.Error())
	}
	r := make([]interface{}, len(s))
	for i, v := range s {
		r[i] = v
	}
	return r, nil
}

// Get the value for a given handle
func (rtc *RunTimeContext) RtcGetValue(ctx context.Context, handle interface{}, value interface{}) error {
	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return pkgerrors.Errorf("Not a valid run time context handle")
	}

	err := contextdb.Db.Get(ctx, str, value)
	if err != nil {
		return pkgerrors.Errorf("Error getting run time context value: %s", err.Error())
	}

	return nil
}

// Update the value of a given handle
func (rtc *RunTimeContext) RtcUpdateValue(ctx context.Context, handle interface{}, value interface{}) error {
	str := fmt.Sprintf("%v", handle)
	sid := fmt.Sprintf("%v", rtc.cid)
	if !strings.HasPrefix(str, sid) {
		return pkgerrors.Errorf("Not a valid run time context handle")
	}
	err := contextdb.Db.Put(ctx, str, value)
	if err != nil {
		return pkgerrors.Errorf("Error updating run time context value: %s", err.Error())
	}
	return nil

}
