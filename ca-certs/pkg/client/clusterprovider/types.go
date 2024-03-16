// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

// EnrollmentKey represents the resources associated with a caCert enrollment
type EnrollmentKey struct {
	Cert            string `json:"caCert"`
	ClusterProvider string `json:"clusterProvider"`
	Enrollment      string `json:"caCertEnrollment"`
}

// DistributionKey represents the resources associated with a caCert distribution
type DistributionKey struct {
	Cert            string `json:"caCert"`
	ClusterProvider string `json:"clusterProvider"`
	Distribution    string `json:"caCertDistribution"`
}
