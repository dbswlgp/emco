// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
)

// ClusterGroup holds the caCert clusterGroup details
type ClusterGroup struct {
	MetaData types.Metadata   `json:"metadata"`
	Spec     ClusterGroupSpec `json:"spec"`
}

// ClusterGroupSpec holds the cluster details
type ClusterGroupSpec struct {
	Label    string `json:"label,omitempty"`   // define the set of cluster(s)
	Cluster  string `json:"cluster,omitempty"` // define the specific cluster
	Provider string `json:"clusterProvider"`   // define the clusterProvider
	Scope    string `json:"scope"`             // specifies which field should be used to identify the cluster(s)
}

// CaCertStatus holds the caCert status details
type CaCertStatus struct {
	ClusterProvider           string `json:"clusterProvider,omitempty"`
	Project                   string `json:"project,omitempty"`
	status.CaCertStatusResult `json:",inline"`
}

// CaCert holds the caCert details
type CaCert struct {
	MetaData types.Metadata `json:"metadata"`
	Spec     CaCertSpec     `json:"spec"`
}

// CertSpec holds the caCert signing details
type CaCertSpec struct {
	Certificate            Certificate            `json:"certificate" yaml:"certificate"`       // represent the certificate info
	CertificateSigningInfo CertificateSigningInfo `json:"csrInfo" yaml:"csrInfo"`               // represent the certificate signining request(CSR) info
	Duration               string                 `json:"duration,omitempty"`                   // duration of the certificate
	IsCA                   bool                   `json:"isCA,omitempty" yaml:"isCA,omitempty"` // specifies the cert is a CA or not
	IssuerRef              certissuer.IssuerRef   `json:"issuerRef"`                            // the details of the issuer for signing the certificate request
	IssuingCluster         IssuingClusterInfo     `json:"issuingCluster"`                       // the details of the issuing cluster
}

// CertificateSigningInfo holds the csr data
type CertificateSigningInfo struct {
	KeySize        int       `json:"keySize,omitempty"`
	Version        int       `json:"version,omitempty"`
	DNSNames       []string  `json:"dnsNames,omitempty"`
	EmailAddresses []string  `json:"emailAddresses,omitempty"`
	KeyUsages      []string  `json:"keyUsages,omitempty"` // certificate usages
	Algorithm      Algorithm `json:"algorithm"`
	Subject        Subject   `json:"subject"`
}

// Subject holds the caCert subject details
type Subject struct {
	Locale       Locale       `json:"locale"`
	Names        Names        `json:"names"`
	Organization Organization `json:"organization"`
}

// Names holds the caCert name details
type Names struct {
	CommonNamePrefix string `json:"commonNamePrefix"`
	CommonName       string
}

// Locale holds the caCert location details
type Locale struct {
	Country       []string `json:"country,omitempty"`
	Locality      []string `json:"locality,omitempty"`
	PostalCode    []string `json:"postalCode,omitempty"`
	Province      []string `json:"province,omitempty"`
	StreetAddress []string `json:"streetAddress,omitempty"`
}

// Organization holds the caCert organization details
type Organization struct {
	Names []string `json:"names,omitempty"`
	Units []string `json:"units,omitempty"`
}

// Algorithm holds the caCert algorithm details
type Algorithm struct {
	PublicKeyAlgorithm string `json:"publicKeyAlgorithm,omitempty"`
	SignatureAlgorithm string `json:"signatureAlgorithm,omitempty"`
}

// IssuingClusterInfo holds the certificate issuing cluster details
type IssuingClusterInfo struct {
	Cluster         string `json:"cluster"`         // name of the cluster
	ClusterProvider string `json:"clusterProvider"` // name of the clusterProvider
}

// Key holds the private keydetails
type Key struct {
	Name string
	Val  string `encrypted:""`
}

// Certificate holds the certificate details
type Certificate struct {
	CommonName string `json:"commonName,omitempty"` // common name to be used on the certificate
}
