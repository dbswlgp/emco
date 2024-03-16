// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module

import (
	"context"
	"gitlab.com/project-emco/core/emco-base/src/sfcclient/pkg/model"
)

// ClientDbInfo structure for storing info about SFC DB
type ClientDbInfo struct {
	storeName  string // name of the mongodb collection to use for client documents
	tagMeta    string // attribute key name for the json data of a client document
	tagContent string // attribute key name for the file data of a client document
	tagContext string // attribute key name for context object in App Context
}

// SfcIntentManager is an interface exposing the SFC Intent functionality
type SfcManager interface {
	// SFC Intent functions
	CreateSfcClientIntent(ctx context.Context, sfc model.SfcClientIntent, pr, ca, caver, dig string, exists bool) (model.SfcClientIntent, error)
	GetSfcClientIntent(ctx context.Context, name, pr, ca, caver, dig string) (model.SfcClientIntent, error)
	GetAllSfcClientIntents(ctx context.Context, pr, ca, caver, dig string) ([]model.SfcClientIntent, error)
	DeleteSfcClientIntent(ctx context.Context, name, pr, ca, caver, dig string) error
}

// SfcClient implements the Manager
// It will also be used to maintain some localized state
type SfcClient struct {
	db ClientDbInfo
}

// NewSfcClient returns an instance of the SfcClient
// which implements the Manager
func NewSfcClient() *SfcClient {
	return &SfcClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}
