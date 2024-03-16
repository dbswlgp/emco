// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package enrollment

import (
	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
)

// EnrollmentContext holds the caCert enrollment details
type EnrollmentContext struct {
	AppContext    appcontext.AppContext
	AppHandle     interface{}
	CaCert        module.CaCert // CA
	ContextID     string
	ResOrder      []string
	ClientName    string
	ClusterGroups []module.ClusterGroup
	ClusterGroup  module.ClusterGroup
	IssuerHandle  interface{}
	Cluster       string
	Resources     EnrollmentResource
	Project       string
	Namespace     string
	LogicalCloud  string
}

// EnrollmentResource holds the resources created for the caCert enrollment
type EnrollmentResource struct {
	CertificateRequest map[string]*cmv1.CertificateRequest
	Certificate        map[string]*cmv1.Certificate
}
