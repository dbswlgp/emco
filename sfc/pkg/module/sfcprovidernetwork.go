// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/model"

	"context"
	pkgerrors "github.com/pkg/errors"
)

// CreateSfcProviderNetworkIntent - create a new SfcProviderNetworkIntent
func (v *SfcProviderNetworkIntentClient) CreateSfcProviderNetworkIntent(ctx context.Context, intent model.SfcProviderNetworkIntent, pr, ca, caver, dig, sfcIntent string, exists bool) (model.SfcProviderNetworkIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcProviderNetworkIntentKey{
		Project:                  pr,
		CompositeApp:             ca,
		CompositeAppVersion:      caver,
		DigName:                  dig,
		SfcIntent:                sfcIntent,
		SfcProviderNetworkIntent: intent.Metadata.Name,
	}

	//Check if this SFC Provider Network Intent already exists
	_, err := v.GetSfcProviderNetworkIntent(ctx, intent.Metadata.Name, pr, ca, caver, dig, sfcIntent)
	if err == nil && !exists {
		return model.SfcProviderNetworkIntent{}, pkgerrors.New("SFC Provider Network Intent already exists")
	}

	err = db.DBconn.Insert(ctx, v.db.storeName, key, nil, v.db.tagMeta, intent)
	if err != nil {
		return model.SfcProviderNetworkIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return intent, nil
}

// GetSfcProviderNetworkIntent returns the SfcProviderNetworkIntent for corresponding name
func (v *SfcProviderNetworkIntentClient) GetSfcProviderNetworkIntent(ctx context.Context, name, pr, ca, caver, dig, sfcIntent string) (model.SfcProviderNetworkIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcProviderNetworkIntentKey{
		Project:                  pr,
		CompositeApp:             ca,
		CompositeAppVersion:      caver,
		DigName:                  dig,
		SfcIntent:                sfcIntent,
		SfcProviderNetworkIntent: name,
	}

	value, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return model.SfcProviderNetworkIntent{}, err
	} else if len(value) == 0 {
		return model.SfcProviderNetworkIntent{}, pkgerrors.New("SFC Provider Network Intent not found")
	}

	//value is a byte array
	if value != nil {
		intent := model.SfcProviderNetworkIntent{}
		err = db.DBconn.Unmarshal(value[0], &intent)
		if err != nil {
			return model.SfcProviderNetworkIntent{}, err
		}
		return intent, nil
	}

	return model.SfcProviderNetworkIntent{}, pkgerrors.New("Unknown Error")
}

// GetAllSfcProviderNetworkIntent returns all of the SFC Intents for for the given Deployment Intent Group
func (v *SfcProviderNetworkIntentClient) GetAllSfcProviderNetworkIntents(ctx context.Context, pr, ca, caver, dig, sfcIntent string) ([]model.SfcProviderNetworkIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcProviderNetworkIntentKey{
		Project:                  pr,
		CompositeApp:             ca,
		CompositeAppVersion:      caver,
		DigName:                  dig,
		SfcIntent:                sfcIntent,
		SfcProviderNetworkIntent: "",
	}

	resp := make([]model.SfcProviderNetworkIntent, 0)

	// verify SFC Intent exists
	_, err := NewSfcIntentClient().GetSfcIntent(ctx, sfcIntent, pr, ca, caver, dig)
	if err != nil {
		return resp, err
	}

	values, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return resp, err
	}

	for _, value := range values {
		cp := model.SfcProviderNetworkIntent{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return resp, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// GetSfcProviderNetworkIntentByEnd returns all of the SFC Provider Network Intents for for the given Deployment Intent Group
func (v *SfcProviderNetworkIntentClient) GetSfcProviderNetworkIntentsByEnd(ctx context.Context, pr, ca, caver, dig, sfcIntent, chainEnd string) ([]model.SfcProviderNetworkIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcProviderNetworkIntentByEndKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		SfcIntent:           sfcIntent,
		ChainEnd:            chainEnd,
	}

	resp := make([]model.SfcProviderNetworkIntent, 0)

	// verify SFC Intent exists
	_, err := NewSfcIntentClient().GetSfcIntent(ctx, sfcIntent, pr, ca, caver, dig)
	if err != nil {
		return resp, err
	}

	values, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return resp, err
	}

	for _, value := range values {
		cp := model.SfcProviderNetworkIntent{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return resp, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteSfcProviderNetworkIntent deletes the SfcProviderNetworkIntent from the database
func (v *SfcProviderNetworkIntentClient) DeleteSfcProviderNetworkIntent(ctx context.Context, name, pr, ca, caver, dig, sfcIntent string) error {

	//Construct key and tag to select the entry
	key := model.SfcProviderNetworkIntentKey{
		Project:                  pr,
		CompositeApp:             ca,
		CompositeAppVersion:      caver,
		DigName:                  dig,
		SfcIntent:                sfcIntent,
		SfcProviderNetworkIntent: name,
	}

	err := db.DBconn.Remove(ctx, v.db.storeName, key)
	return err
}
