// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/model"

	"context"
	pkgerrors "github.com/pkg/errors"
)

// CreateSfcLinkIntent - create a new SfcLinkIntent
func (v *SfcLinkIntentClient) CreateSfcLinkIntent(ctx context.Context, intent model.SfcLinkIntent, pr, ca, caver, dig, sfcIntent string, exists bool) (model.SfcLinkIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcLinkIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		SfcIntent:           sfcIntent,
		SfcLinkIntent:       intent.Metadata.Name,
	}

	//Check if this SFC Link Intent already exists
	_, err := v.GetSfcLinkIntent(ctx, intent.Metadata.Name, pr, ca, caver, dig, sfcIntent)
	if err == nil && !exists {
		return model.SfcLinkIntent{}, pkgerrors.New("SFC Link Intent already exists")
	}

	err = db.DBconn.Insert(ctx, v.db.storeName, key, nil, v.db.tagMeta, intent)
	if err != nil {
		return model.SfcLinkIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return intent, nil
}

// GetSfcLinkIntent returns the SfcLinkIntent for corresponding name
func (v *SfcLinkIntentClient) GetSfcLinkIntent(ctx context.Context, name, pr, ca, caver, dig, sfcIntent string) (model.SfcLinkIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcLinkIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		SfcIntent:           sfcIntent,
		SfcLinkIntent:       name,
	}

	value, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return model.SfcLinkIntent{}, err
	} else if len(value) == 0 {
		return model.SfcLinkIntent{}, pkgerrors.New("SFC Link Intent not found")
	}

	//value is a byte array
	if value != nil {
		intent := model.SfcLinkIntent{}
		err = db.DBconn.Unmarshal(value[0], &intent)
		if err != nil {
			return model.SfcLinkIntent{}, err
		}
		return intent, nil
	}

	return model.SfcLinkIntent{}, pkgerrors.New("Unknown Error")
}

// GetAllSfcLinkIntent returns all of the SFC Intents for for the given Deployment Intent Group
func (v *SfcLinkIntentClient) GetAllSfcLinkIntents(ctx context.Context, pr, ca, caver, dig, sfcIntent string) ([]model.SfcLinkIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcLinkIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		SfcIntent:           sfcIntent,
		SfcLinkIntent:       "",
	}

	resp := make([]model.SfcLinkIntent, 0)

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
		cp := model.SfcLinkIntent{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return resp, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteSfcLinkIntent deletes the SfcLinkIntent from the database
func (v *SfcLinkIntentClient) DeleteSfcLinkIntent(ctx context.Context, name, pr, ca, caver, dig, sfcIntent string) error {

	//Construct key and tag to select the entry
	key := model.SfcLinkIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		SfcIntent:           sfcIntent,
		SfcLinkIntent:       name,
	}

	err := db.DBconn.Remove(ctx, v.db.storeName, key)
	return err
}
