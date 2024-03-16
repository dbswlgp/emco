// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"context"
	"reflect"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// CaCertLogicalCloudManager exposes all the caCert logicalCloud functionalities
type CaCertLogicalCloudManager interface {
	CreateLogicalCloud(ctx context.Context, logicalCloud CaCertLogicalCloud, cert, project string, failIfExists bool) (CaCertLogicalCloud, bool, error)
	DeleteLogicalCloud(ctx context.Context, logicalCloud, cert, project string) error
	GetAllLogicalClouds(ctx context.Context, cert, project string) ([]CaCertLogicalCloud, error)
	GetLogicalCloud(ctx context.Context, logicalCloud, cert, project string) (CaCertLogicalCloud, error)
}

// CaCertLogicalCloudKey represents the resources associated with a caCert logicalCloud
type CaCertLogicalCloudKey struct {
	Cert               string `json:"caCertLc"`
	CaCertLogicalCloud string `json:"caCertLogicalCloud"`
	Project            string `json:"project"`
}

// CaCertLogicalCloudClient holds the client properties
type CaCertLogicalCloudClient struct {
	dbInfo db.DbInfo
}

// NewCaCertLogicalCloudClient returns an instance of the CaCertLogicalCloudClient which implements the Manager
func NewCaCertLogicalCloudClient() *CaCertLogicalCloudClient {
	return &CaCertLogicalCloudClient{
		dbInfo: db.DbInfo{
			StoreName: "resources",
			TagMeta:   "data"}}
}

// CreateLogicalCloud creates a caCert logicalCloud
func (c *CaCertLogicalCloudClient) CreateLogicalCloud(ctx context.Context, logicalCloud CaCertLogicalCloud, cert, project string, failIfExists bool) (CaCertLogicalCloud, bool, error) {
	lcExists := false
	key := CaCertLogicalCloudKey{
		Cert:               cert,
		Project:            project,
		CaCertLogicalCloud: logicalCloud.MetaData.Name}

	if lc, err := c.GetLogicalCloud(ctx, logicalCloud.MetaData.Name, cert, project); err == nil &&
		!reflect.DeepEqual(lc, CaCertLogicalCloud{}) {
		lcExists = true
	}

	if lcExists &&
		failIfExists {
		return CaCertLogicalCloud{}, lcExists, emcoerror.NewEmcoError(
			module.CaCertLogicalCloudAlreadyExists,
			emcoerror.Conflict,
		)
	}

	if err := db.DBconn.Insert(ctx, c.dbInfo.StoreName, key, nil, c.dbInfo.TagMeta, logicalCloud); err != nil {
		return CaCertLogicalCloud{}, lcExists, err
	}

	return logicalCloud, lcExists, nil
}

// DeleteLogicalCloud deletes a caCert logicalCloud
func (c *CaCertLogicalCloudClient) DeleteLogicalCloud(ctx context.Context, logicalCloud, cert, project string) error {
	key := CaCertLogicalCloudKey{
		Cert:               cert,
		CaCertLogicalCloud: logicalCloud,
		Project:            project}

	return db.DBconn.Remove(ctx, c.dbInfo.StoreName, key)
}

// GetAllLogicalClouds returns all the caCert logicalCloud
func (c *CaCertLogicalCloudClient) GetAllLogicalClouds(ctx context.Context, cert, project string) ([]CaCertLogicalCloud, error) {
	key := CaCertLogicalCloudKey{
		Cert:    cert,
		Project: project}

	values, err := db.DBconn.Find(ctx, c.dbInfo.StoreName, key, c.dbInfo.TagMeta)
	if err != nil {
		return []CaCertLogicalCloud{}, err
	}

	var logicalClouds []CaCertLogicalCloud
	for _, value := range values {
		lc := CaCertLogicalCloud{}
		if err = db.DBconn.Unmarshal(value, &lc); err != nil {
			return []CaCertLogicalCloud{}, err
		}
		logicalClouds = append(logicalClouds, lc)
	}

	return logicalClouds, nil
}

// GetLogicalCloud returns the caCert logicalCloud
func (c *CaCertLogicalCloudClient) GetLogicalCloud(ctx context.Context, logicalCloud, cert, project string) (CaCertLogicalCloud, error) {
	key := CaCertLogicalCloudKey{
		Cert:               cert,
		CaCertLogicalCloud: logicalCloud,
		Project:            project}

	value, err := db.DBconn.Find(ctx, c.dbInfo.StoreName, key, c.dbInfo.TagMeta)
	if err != nil {
		return CaCertLogicalCloud{}, err
	}

	if len(value) == 0 {
		return CaCertLogicalCloud{}, emcoerror.NewEmcoError(
			module.CaCertLogicalCloudNotFound,
			emcoerror.NotFound,
		)
	}

	if value != nil {
		lc := CaCertLogicalCloud{}
		if err = db.DBconn.Unmarshal(value[0], &lc); err != nil {
			return CaCertLogicalCloud{}, err
		}
		return lc, nil
	}

	return CaCertLogicalCloud{}, emcoerror.NewEmcoError(
		emcoerror.UnknownErrorMessage,
		emcoerror.Unknown,
	)
}
