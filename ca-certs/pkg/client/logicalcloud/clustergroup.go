// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"context"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
)

// ClusterGroupManager exposes all the caCert clusterGroup functionalities
type ClusterGroupManager interface {
	CreateClusterGroup(ctx context.Context, cluster module.ClusterGroup, logicalCloud, cert, project string, failIfExists bool) (module.ClusterGroup, bool, error)
	DeleteClusterGroup(ctx context.Context, cluster, logicalCloud, cert, project string) error
	GetAllClusterGroups(ctx context.Context, logicalCloud, cert, project string) ([]module.ClusterGroup, error)
	GetClusterGroup(ctx context.Context, cluster, logicalCloud, cert, project string) (module.ClusterGroup, error)
}

// ClusterGroupKey represents the resources associated with a caCert clusterGroup
type ClusterGroupKey struct {
	Cert               string `json:"caCertLc"`
	ClusterGroup       string `json:"caCertClusterGroupLc"`
	CaCertLogicalCloud string `json:"caCertLogicalCloud"`
	Project            string `json:"project"`
}

// ClusterGroupClient holds the client properties
type ClusterGroupClient struct {
}

// NewClusterGroupClient returns an instance of the ClusterGroupClient which implements the Manager
func NewClusterGroupClient() *ClusterGroupClient {
	return &ClusterGroupClient{}
}

// CreateClusterGroup creates a caCert clusterGroup
func (c *ClusterGroupClient) CreateClusterGroup(ctx context.Context, group module.ClusterGroup, caCertLogicalCloud, cert, project string, failIfExists bool) (module.ClusterGroup, bool, error) {
	ck := ClusterGroupKey{
		Cert:               cert,
		ClusterGroup:       group.MetaData.Name,
		CaCertLogicalCloud: caCertLogicalCloud,
		Project:            project}

	return module.NewClusterGroupClient(ck).CreateClusterGroup(ctx, group, failIfExists)
}

// DeleteClusterGroup deletes a caCert clusterGroup
func (c *ClusterGroupClient) DeleteClusterGroup(ctx context.Context, clusterGroup, caCertLogicalCloud, cert, project string) error {
	ck := ClusterGroupKey{
		Cert:               cert,
		ClusterGroup:       clusterGroup,
		CaCertLogicalCloud: caCertLogicalCloud,
		Project:            project}

	return module.NewClusterGroupClient(ck).DeleteClusterGroup(ctx)
}

// GetAllClusterGroups returns all the caCert clusterGroup
func (c *ClusterGroupClient) GetAllClusterGroups(ctx context.Context, caCertLogicalCloud, cert, project string) ([]module.ClusterGroup, error) {
	ck := ClusterGroupKey{
		Cert:               cert,
		CaCertLogicalCloud: caCertLogicalCloud,
		Project:            project}

	return module.NewClusterGroupClient(ck).GetAllClusterGroups(ctx)
}

// GetClusterGroup returns the caCert clusterGroup
func (c *ClusterGroupClient) GetClusterGroup(ctx context.Context, clusterGroup, caCertLogicalCloud, cert, project string) (module.ClusterGroup, error) {
	ck := ClusterGroupKey{
		Cert:               cert,
		ClusterGroup:       clusterGroup,
		CaCertLogicalCloud: caCertLogicalCloud,
		Project:            project}

	return module.NewClusterGroupClient(ck).GetClusterGroup(ctx)
}
