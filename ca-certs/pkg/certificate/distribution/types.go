// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package distribution

import (
	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	tcsv1 "github.com/intel/trusted-certificate-issuer/api/v1alpha1"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/service/istioservice"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/service/knccservice"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	v1 "k8s.io/api/core/v1"
)

// DistributionContext holds the caCert distribution details
type DistributionContext struct {
	AppContext          appcontext.AppContext
	AppHandle           interface{}
	CaCert              module.CaCert // CA
	Project             string
	ContextID           string
	ResOrder            []string
	EnrollmentContextID string
	Certificates        []cmv1.Certificate
	CertificateRequests []cmv1.CertificateRequest
	Secrets             []v1.Secret
	Resources           DistributionResource
	ClusterGroups       []module.ClusterGroup
	ClusterGroup        module.ClusterGroup
	Namespace           string
	ClientName          string
	Cluster             string
	ClusterHandle       interface{}
	LogicalCloud        string
}

// DistributionResource holds the resources created for the caCert distribution
type DistributionResource struct {
	ClusterIssuer map[string]*cmv1.ClusterIssuer
	ProxyConfig   map[string]*istioservice.ProxyConfig
	Secret        map[string]*v1.Secret
	KnccConfig    map[string]*knccservice.Config
	TCSIssuer     map[string]*tcsv1.TCSIssuer
}
