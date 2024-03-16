// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"context"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

// getCertificate retrieves the caCert from db
func getCertificate(ctx context.Context, cert, project string) (module.CaCert, error) {
	caCert, err := NewCaCertClient().GetCert(ctx, cert, project)
	if err != nil {
		logutils.Error("Failed to retrieve the caCert", logutils.Fields{
			"Cert":    cert,
			"Project": project,
			"Error":   err.Error()})
		return module.CaCert{}, err
	}
	return caCert, nil
}

// getAllLogicalClouds retrieves the logicalCloud(s) from db
func getAllLogicalClouds(ctx context.Context, cert, project string) ([]CaCertLogicalCloud, error) {
	// get all the logicalCloud(s) within the caCert
	lcs, err := NewCaCertLogicalCloudClient().GetAllLogicalClouds(ctx, cert, project)
	if err != nil {
		logutils.Error("Failed to retrieve the logicalCloud(s)", logutils.Fields{
			"Cert":    cert,
			"Project": project,
			"Error":   err.Error()})
		return []CaCertLogicalCloud{}, err
	}
	return lcs, nil
}

// getAllClusterGroup retrieves the clusterGroup(s) from db
func getAllClusterGroup(ctx context.Context, logicalCloud, cert, project string) ([]module.ClusterGroup, error) {
	// get all the clusterGroup(s) within the caCert and logicalCloud
	clusters, err := NewClusterGroupClient().GetAllClusterGroups(ctx, logicalCloud, cert, project)
	if err != nil {
		logutils.Error("Failed to retrieve the clusterGroup(s)", logutils.Fields{
			"Cert":         cert,
			"LogicalCloud": logicalCloud,
			"Project":      project,
			"Error":        err.Error()})
		return []module.ClusterGroup{}, err
	}

	return clusters, nil
}
