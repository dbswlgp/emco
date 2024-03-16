// SPDX-License-Identifier: Apache-2.0
// C[]model.SfcClientIntent{}opyright (c) 2021 Intel Corporation

package module

import (
	"context"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	orchmod "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/sfcclient/pkg/model"

	pkgerrors "github.com/pkg/errors"
)

// CreateSfcClientIntent - create a new SfcClientIntent
func (v *SfcClient) CreateSfcClientIntent(ctx context.Context, intent model.SfcClientIntent, pr, ca, caver, dig string, exists bool) (model.SfcClientIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcClientIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		SfcClientIntent:     intent.Metadata.Name,
	}

	//Check if this SFC Client Intent already exists
	_, err := v.GetSfcClientIntent(ctx, intent.Metadata.Name, pr, ca, caver, dig)
	if err == nil && !exists {
		return model.SfcClientIntent{}, pkgerrors.New("SFC Client Intent already exists")
	}

	err = db.DBconn.Insert(ctx, v.db.storeName, key, nil, v.db.tagMeta, intent)
	if err != nil {
		return model.SfcClientIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return intent, nil
}

// GetSfcClientIntent returns the SfcClientIntent for corresponding name
func (v *SfcClient) GetSfcClientIntent(ctx context.Context, name, pr, ca, caver, dig string) (model.SfcClientIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcClientIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		SfcClientIntent:     name,
	}

	value, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return model.SfcClientIntent{}, err
	} else if len(value) == 0 {
		return model.SfcClientIntent{}, pkgerrors.New("SFC Client Intent not found")
	}

	//value is a byte array
	if value != nil {
		intent := model.SfcClientIntent{}
		err = db.DBconn.Unmarshal(value[0], &intent)
		if err != nil {
			return model.SfcClientIntent{}, err
		}
		return intent, nil
	}

	return model.SfcClientIntent{}, pkgerrors.New("Unknown Error")
}

// GetAllSfcClientIntent returns all of the SFC Client Intents for for the given network control intent
func (v *SfcClient) GetAllSfcClientIntents(ctx context.Context, pr, ca, caver, dig string) ([]model.SfcClientIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcClientIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		SfcClientIntent:     "",
	}

	resp := make([]model.SfcClientIntent, 0)

	// Verify the Deployment Intent Group exists
	_, err := orchmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, dig, pr, ca, caver)
	if err != nil {
		return resp, err
	}

	values, err := db.DBconn.Find(ctx, v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return resp, err
	}

	for _, value := range values {
		cp := model.SfcClientIntent{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return resp, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteSfcClientIntent deletes the SfcClientIntent from the database
func (v *SfcClient) DeleteSfcClientIntent(ctx context.Context, name, pr, ca, caver, dig string) error {

	//Construct key and tag to select the entry
	key := model.SfcClientIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		SfcClientIntent:     name,
	}

	err := db.DBconn.Remove(ctx, v.db.storeName, key)
	return err
}
