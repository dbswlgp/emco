// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/model"

	"context"
	pkgerrors "github.com/pkg/errors"
)

// CreateSfcClientSelectorIntent - create a new SfcClientSelectorIntent
func (v *SfcClientSelectorIntentClient) CreateSfcClientSelectorIntent(ctx context.Context, intent model.SfcClientSelectorIntent, pr, ca, caver, dig, sfcIntent string, exists bool) (model.SfcClientSelectorIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcClientSelectorIntentKey{
		Project:                 pr,
		CompositeApp:            ca,
		CompositeAppVersion:     caver,
		DigName:                 dig,
		SfcIntent:               sfcIntent,
		SfcClientSelectorIntent: intent.Metadata.Name,
	}

	endKey := model.SfcEndKey{
		ChainEnd: intent.Spec.ChainEnd,
	}

	//Check if this SFC Client Selector Intent already exists
	_, err := v.GetSfcClientSelectorIntent(ctx, intent.Metadata.Name, pr, ca, caver, dig, sfcIntent)
	if err == nil && !exists {
		return model.SfcClientSelectorIntent{}, pkgerrors.New("SFC Client Selector Intent already exists")
	}

	err = db.DBconn.Insert(ctx, v.db.storeName, key, endKey, v.db.tagMeta, intent)
	if err != nil {
		return model.SfcClientSelectorIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return intent, nil
}

// GetSfcClientSelectorIntent returns the SfcClientSelectorIntent for corresponding name
func (v *SfcClientSelectorIntentClient) GetSfcClientSelectorIntent(ctx context.Context, name, pr, ca, caver, dig, sfcIntent string) (model.SfcClientSelectorIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcClientSelectorIntentKey{
		Project:                 pr,
		CompositeApp:            ca,
		CompositeAppVersion:     caver,
		DigName:                 dig,
		SfcIntent:               sfcIntent,
		SfcClientSelectorIntent: name,
	}

	value, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return model.SfcClientSelectorIntent{}, err
	} else if len(value) == 0 {
		return model.SfcClientSelectorIntent{}, pkgerrors.New("SFC Client Selector Intent not found")
	}

	//value is a byte array
	if value != nil {
		intent := model.SfcClientSelectorIntent{}
		err = db.DBconn.Unmarshal(value[0], &intent)
		if err != nil {
			return model.SfcClientSelectorIntent{}, err
		}
		return intent, nil
	}

	return model.SfcClientSelectorIntent{}, pkgerrors.New("Unknown Error")
}

// GetAllSfcClientSelectorIntent returns all of the SFC Intents for for the given Deployment Intent Group
func (v *SfcClientSelectorIntentClient) GetAllSfcClientSelectorIntents(ctx context.Context, pr, ca, caver, dig, sfcIntent string) ([]model.SfcClientSelectorIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcClientSelectorIntentKey{
		Project:                 pr,
		CompositeApp:            ca,
		CompositeAppVersion:     caver,
		DigName:                 dig,
		SfcIntent:               sfcIntent,
		SfcClientSelectorIntent: "",
	}

	resp := make([]model.SfcClientSelectorIntent, 0)

	// Verify the SFC intent exists
	_, err := NewSfcIntentClient().GetSfcIntent(ctx, sfcIntent, pr, ca, caver, dig)
	if err != nil {
		return resp, err
	}

	values, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return resp, err
	}

	for _, value := range values {
		cp := model.SfcClientSelectorIntent{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return resp, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// GetSfcClientSelectorIntentsByEnd returns all of the SFC Client Selector Intents for for the given Deployment Intent Group
// and specified end of the chain
func (v *SfcClientSelectorIntentClient) GetSfcClientSelectorIntentsByEnd(ctx context.Context, pr, ca, caver, dig, sfcIntent, chainEnd string) ([]model.SfcClientSelectorIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcClientSelectorIntentByEndKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		SfcIntent:           sfcIntent,
		ChainEnd:            chainEnd,
	}

	resp := make([]model.SfcClientSelectorIntent, 0)

	// Verify the SFC intent exists
	_, err := NewSfcIntentClient().GetSfcIntent(ctx, sfcIntent, pr, ca, caver, dig)
	if err != nil {
		return resp, err
	}

	values, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return resp, err
	}

	for _, value := range values {
		cp := model.SfcClientSelectorIntent{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return resp, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteSfcClientSelectorIntent deletes the SfcClientSelectorIntent from the database
func (v *SfcClientSelectorIntentClient) DeleteSfcClientSelectorIntent(ctx context.Context, name, pr, ca, caver, dig, sfcIntent string) error {

	//Construct key and tag to select the entry
	key := model.SfcClientSelectorIntentKey{
		Project:                 pr,
		CompositeApp:            ca,
		CompositeAppVersion:     caver,
		DigName:                 dig,
		SfcIntent:               sfcIntent,
		SfcClientSelectorIntent: name,
	}

	err := db.DBconn.Remove(ctx, v.db.storeName, key)
	return err
}
