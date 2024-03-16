// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module

import (
	"context"

	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/model"
)

// ClientDbInfo structure for storing info about SFC DB
type ClientDbInfo struct {
	storeName  string // name of the mongodb collection to use for client documents
	tagMeta    string // attribute key name for the json data of a client document
	tagContent string // attribute key name for the file data of a client document
	tagContext string // attribute key name for context object in App Context
}

// SfcIntentManager is an interface exposing the SFC Intent functionality
type SfcIntentManager interface {
	// SFC Intent functions
	CreateSfcIntent(ctx context.Context, sfc model.SfcIntent, pr, ca, caver, dig string, exists bool) (model.SfcIntent, error)
	GetSfcIntent(ctx context.Context, name, pr, ca, caver, dig string) (model.SfcIntent, error)
	GetAllSfcIntents(ctx context.Context, pr, ca, caver, dig string) ([]model.SfcIntent, error)
	DeleteSfcIntent(ctx context.Context, name, pr, ca, caver, dig string) error
}

// SfcIntentManager is an interface exposing the SFC Intent functionality
type SfcLinkIntentManager interface {
	// SFC Client Selector Intent functions
	CreateSfcLinkIntent(ctx context.Context, sfc model.SfcLinkIntent, pr, ca, caver, dig, sfcIntent string, exists bool) (model.SfcLinkIntent, error)
	GetSfcLinkIntent(ctx context.Context, name, pr, ca, caver, dig, sfcIntent string) (model.SfcLinkIntent, error)
	GetAllSfcLinkIntents(ctx context.Context, pr, ca, caver, dig, sfcIntent string) ([]model.SfcLinkIntent, error)
	DeleteSfcLinkIntent(ctx context.Context, name, pr, ca, caver, dig, sfcIntent string) error
}

// SfcIntentManager is an interface exposing the SFC Intent functionality
type SfcClientSelectorIntentManager interface {
	// SFC Client Selector Intent functions
	CreateSfcClientSelectorIntent(ctx context.Context, sfc model.SfcClientSelectorIntent, pr, ca, caver, dig, sfcIntent string, exists bool) (model.SfcClientSelectorIntent, error)
	GetSfcClientSelectorIntent(ctx context.Context, name, pr, ca, caver, dig, sfcIntent string) (model.SfcClientSelectorIntent, error)
	GetAllSfcClientSelectorIntents(ctx context.Context, pr, ca, caver, dig, sfcIntent string) ([]model.SfcClientSelectorIntent, error)
	GetSfcClientSelectorIntentsByEnd(ctx context.Context, pr, ca, caver, dig, sfcIntent, chainEnd string) ([]model.SfcClientSelectorIntent, error)
	DeleteSfcClientSelectorIntent(ctx context.Context, name, pr, ca, caver, dig, sfcIntent string) error
}

// SfcIntentManager is an interface exposing the SFC Intent functionality
type SfcProviderNetworkIntentManager interface {
	// SFC Network Provider Intent functions
	CreateSfcProviderNetworkIntent(ctx context.Context, sfc model.SfcProviderNetworkIntent, pr, ca, caver, dig, sfcIntent string, exists bool) (model.SfcProviderNetworkIntent, error)
	GetSfcProviderNetworkIntent(ctx context.Context, name, pr, ca, caver, dig, sfcIntent string) (model.SfcProviderNetworkIntent, error)
	GetAllSfcProviderNetworkIntents(ctx context.Context, pr, ca, caver, dig, sfcIntent string) ([]model.SfcProviderNetworkIntent, error)
	GetSfcProviderNetworkIntentsByEnd(ctx context.Context, pr, ca, caver, dig, sfcIntent, chainEnd string) ([]model.SfcProviderNetworkIntent, error)
	DeleteSfcProviderNetworkIntent(ctx context.Context, name, pr, ca, caver, dig, sfcIntent string) error
}

// Client for using the services in the ncm
type Client struct {
	SfcIntent                *SfcIntentClient
	SfcLinkIntent            *SfcLinkIntentClient
	SfcClientSelectorIntent  *SfcClientSelectorIntentClient
	SfcProviderNetworkIntent *SfcProviderNetworkIntentClient
	// Add Clients for API's here
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.SfcIntent = NewSfcIntentClient()
	c.SfcLinkIntent = NewSfcLinkIntentClient()
	c.SfcClientSelectorIntent = NewSfcClientSelectorIntentClient()
	c.SfcProviderNetworkIntent = NewSfcProviderNetworkIntentClient()
	// Add Client API handlers here
	return c
}

// SfcIntentClient implements the SfcIntentManager
// It will also be used to maintain some localized state
type SfcIntentClient struct {
	db ClientDbInfo
}

// NewSfcIntentClient returns an instance of the SfcIntentClient
// which implements the Manager for SFC Intents
func NewSfcIntentClient() *SfcIntentClient {
	return &SfcIntentClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}

// SfcLinkIntentClient implements the SfcLinkIntentManager
// It will also be used to maintain some localized state
type SfcLinkIntentClient struct {
	db ClientDbInfo
}

// NewSfcLinkIntentClient returns an instance of the SfcIntentClient
// which implements the Manager for SFC Client Selector Intents
func NewSfcLinkIntentClient() *SfcLinkIntentClient {
	return &SfcLinkIntentClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}

// SfcClientSelectorIntentClient implements the SfcClientSelectorIntentManager
// It will also be used to maintain some localized state
type SfcClientSelectorIntentClient struct {
	db ClientDbInfo
}

// NewSfcClientSelectorIntentClient returns an instance of the SfcIntentClient
// which implements the Manager for SFC Client Selector Intents
func NewSfcClientSelectorIntentClient() *SfcClientSelectorIntentClient {
	return &SfcClientSelectorIntentClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}

// SfcProviderNetworkIntentClient implements the SfcProviderNetworkIntentManager
// It will also be used to maintain some localized state
type SfcProviderNetworkIntentClient struct {
	db ClientDbInfo
}

// NewSfcProviderNetworkIntentClient returns an instance of the SfcIntentClient
// which implements the Manager for SFC Provider Network Intents
func NewSfcProviderNetworkIntentClient() *SfcProviderNetworkIntentClient {
	return &SfcProviderNetworkIntentClient{
		db: ClientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
		},
	}
}
