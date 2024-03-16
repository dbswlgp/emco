// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"

// EnrollmentKey represents the resources associated with a caCert enrollment
type EnrollmentKey struct {
	Cert       string `json:"caCert"`
	Project    string `json:"project"`
	Enrollment string `json:"caCertEnrollment"`
}

// DistributionKey represents the resources associated with a caCert distribution
type DistributionKey struct {
	Cert         string `json:"caCert"`
	Project      string `json:"project"`
	Distribution string `json:"caCertDistribution"`
}

// CaCertLogicalCloud holds the caCert logicalCloud details
type CaCertLogicalCloud struct {
	MetaData types.Metadata         `json:"metadata"`
	Spec     CaCertLogicalCloudSpec `json:"spec"`
}

// CaCertLogicalCloudSpec holds the logicalCloud details
type CaCertLogicalCloudSpec struct {
	LogicalCloud string `json:"logicalCloud"` // name of the logicalCloud
}
