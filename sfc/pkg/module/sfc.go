// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	orchmod "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/model"

	"context"
	pkgerrors "github.com/pkg/errors"
)

// CreateSfcIntent - create a new SfcIntent
func (v *SfcIntentClient) CreateSfcIntent(ctx context.Context, intent model.SfcIntent, pr, ca, caver, dig string, exists bool) (model.SfcIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		SfcIntent:           intent.Metadata.Name,
	}

	//Check if this SFC Intent already exists
	_, err := v.GetSfcIntent(ctx, intent.Metadata.Name, pr, ca, caver, dig)
	if err == nil && !exists {
		return model.SfcIntent{}, pkgerrors.New("SFC Intent already exists")
	}

	err = db.DBconn.Insert(ctx, v.db.storeName, key, nil, v.db.tagMeta, intent)
	if err != nil {
		return model.SfcIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return intent, nil
}

// GetSfcIntent returns the SfcIntent for corresponding name
func (v *SfcIntentClient) GetSfcIntent(ctx context.Context, name, pr, ca, caver, dig string) (model.SfcIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		SfcIntent:           name,
	}

	value, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return model.SfcIntent{}, err
	} else if len(value) == 0 {
		return model.SfcIntent{}, pkgerrors.New("SFC Intent not found")
	}

	//value is a byte array
	if value != nil {
		intent := model.SfcIntent{}
		err = db.DBconn.Unmarshal(value[0], &intent)
		if err != nil {
			return model.SfcIntent{}, err
		}
		return intent, nil
	}

	return model.SfcIntent{}, pkgerrors.New("Unknown Error")
}

// GetAllSfcIntent returns all of the SFC Intents for for the given network control intent
func (v *SfcIntentClient) GetAllSfcIntents(ctx context.Context, pr, ca, caver, dig string) ([]model.SfcIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		SfcIntent:           "",
	}

	resp := make([]model.SfcIntent, 0)

	// Verify the Deployment Intent Group exists
	_, err := orchmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, dig, pr, ca, caver)
	if err != nil {
		return resp, err
	}

	values, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []model.SfcIntent{}, err
	}

	for _, value := range values {
		cp := model.SfcIntent{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []model.SfcIntent{}, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteSfcIntent deletes the SfcIntent from the database
func (v *SfcIntentClient) DeleteSfcIntent(ctx context.Context, name, pr, ca, caver, dig string) error {

	//Construct key and tag to select the entry
	key := model.SfcIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		SfcIntent:           name,
	}

	err := db.DBconn.Remove(ctx, v.db.storeName, key)
	return err
}
