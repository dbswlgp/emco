// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package intent

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

func NewClient(config Config) *Client {
	return &Client{
		db:           config.Db,
		tag:          config.Tag,
		storeName:    config.StoreName,
		updateStream: config.UpdateStream,
	}
}

func (c Client) CreateIntent(ctx context.Context, request *Request) (*Intent, error) {
	key := Key{
		PolicyIntent:        request.PolicyIntentId,
		Project:             request.Project,
		CompositeApp:        request.CompositeApp,
		CompositeAppVersion: request.CompositeAppVersion,
		DigName:             request.DeploymentIntentGroup,
	}
	intent := *request.IntentData
	value, err := c.db.Find(ctx, c.storeName, key, c.tag)
	if err == nil && len(value) > 0 {
		// Remove from in-memory if events are different
		data := new(Intent)
		if err := c.db.Unmarshal(value[0], data); err != nil {
			return nil, err
		}
		if !cmp.Equal(data.Spec.Event, intent.Spec.Event) {
			c.updateStream <- StreamData{
				Operation: "DELETE",
				Intent:    *data,
			}
		}
	}
	if err := c.db.Insert(ctx, c.storeName, key, nil, c.tag, intent); err != nil {
		return nil, err
	}
	// Mark for appending to the in-memory list
	c.updateStream <- StreamData{
		Operation: "APPEND",
		Intent:    intent,
	}
	return &intent, nil
}

func (c Client) DeleteIntent(ctx context.Context, request *Request) error {
	key := Key{
		PolicyIntent:        request.PolicyIntentId,
		Project:             request.Project,
		CompositeApp:        request.CompositeApp,
		CompositeAppVersion: request.CompositeAppVersion,
		DigName:             request.DeploymentIntentGroup,
	}
	intent, err := c.GetIntent(ctx, request)
	if err != nil {
		return err
	}
	if intent == nil {
		return errors.Errorf("Policy Intent not found")
	}
	// Deleting from in memory list can be a time-consuming operation.
	// Hence, we will just mark for deletion and proceed
	if err := c.db.Remove(ctx, c.storeName, key); err != nil {
		return err
	}
	c.updateStream <- StreamData{
		Operation: "DELETE",
		Intent:    *intent,
	}
	return nil

}

func (c Client) GetIntent(ctx context.Context, request *Request) (*Intent, error) {
	key := Key{
		PolicyIntent:        request.PolicyIntentId,
		Project:             request.Project,
		CompositeApp:        request.CompositeApp,
		CompositeAppVersion: request.CompositeAppVersion,
		DigName:             request.DeploymentIntentGroup,
	}
	value, err := c.db.Find(ctx, c.storeName, key, c.tag)
	if err != nil || len(value) == 0 {
		return nil, err
	}
	data := new(Intent)
	if err := c.db.Unmarshal(value[0], data); err != nil {
		return nil, err
	}
	return data, nil
}
