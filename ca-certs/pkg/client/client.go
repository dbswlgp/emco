// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package client

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
)

// Client for using the services
type Client struct {
	ClusterProviderCert             *clusterprovider.CaCertClient
	ClusterProviderCluster          *clusterprovider.ClusterGroupClient
	ClusterProviderCertDistribution *clusterprovider.CaCertDistributionClient
	ClusterProviderCertEnrollment   *clusterprovider.CaCertEnrollmentClient
	LogicalCloud                    *logicalcloud.CaCertLogicalCloudClient
	LogicalCloudCert                *logicalcloud.CaCertClient
	LogicalCloudCluster             *logicalcloud.ClusterGroupClient
	LogicalCloudCertDistribution    *logicalcloud.CaCertDistributionClient
	LogicalCloudCertEnrollment      *logicalcloud.CaCertEnrollmentClient
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.ClusterProviderCert = clusterprovider.NewCaCertClient()
	c.ClusterProviderCluster = clusterprovider.NewClusterGroupClient()
	c.ClusterProviderCertDistribution = clusterprovider.NewCaCertDistributionClient()
	c.ClusterProviderCertEnrollment = clusterprovider.NewCaCertEnrollmentClient()
	c.LogicalCloud = logicalcloud.NewCaCertLogicalCloudClient()
	c.LogicalCloudCert = logicalcloud.NewCaCertClient()
	c.LogicalCloudCluster = logicalcloud.NewClusterGroupClient()
	c.LogicalCloudCertDistribution = logicalcloud.NewCaCertDistributionClient()
	c.LogicalCloudCertEnrollment = logicalcloud.NewCaCertEnrollmentClient()
	return c
}
