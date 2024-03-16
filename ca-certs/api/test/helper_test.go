// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// These test cases are to validate the route handler functionalities
package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

func populateCertTestData() []module.CaCert {
	return []module.CaCert{
		{
			MetaData: types.Metadata{
				Name:        "testCert1",
				Description: "test cert",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: module.CaCertSpec{
				IsCA: true,
				IssuerRef: certissuer.IssuerRef{
					Name:  "foo",
					Kind:  "ClusterIssuer",
					Group: "cert-manager.io",
				},
				Duration: "8760h",
				IssuingCluster: module.IssuingClusterInfo{
					Cluster:         "issuer1",
					ClusterProvider: "provider1",
				},
				CertificateSigningInfo: module.CertificateSigningInfo{
					KeySize: 4096,
					Version: 1,
					Algorithm: module.Algorithm{
						PublicKeyAlgorithm: "RSA",
						SignatureAlgorithm: "SHA512WithRSA",
					},
					Subject: module.Subject{
						Locale: module.Locale{},
						Names: module.Names{
							CommonName: "foo",
						},
						Organization: module.Organization{},
					},
				},
			},
		},
		{
			MetaData: types.Metadata{
				Name:        "testCert2",
				Description: "test cert",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: module.CaCertSpec{
				IsCA: true,
				IssuerRef: certissuer.IssuerRef{
					Name:  "foo",
					Kind:  "ClusterIssuer",
					Group: "cert-manager.io",
				},
				Duration: "8760h",
				IssuingCluster: module.IssuingClusterInfo{
					Cluster:         "issuer1",
					ClusterProvider: "provider1",
				},
				CertificateSigningInfo: module.CertificateSigningInfo{
					KeySize: 4096,
					Version: 1,
					Algorithm: module.Algorithm{
						PublicKeyAlgorithm: "RSA",
						SignatureAlgorithm: "SHA512WithRSA",
					},
					Subject: module.Subject{
						Locale: module.Locale{},
						Names: module.Names{
							CommonName: "foo",
						},
						Organization: module.Organization{},
					},
				},
			},
		},
		{
			MetaData: types.Metadata{
				Name:        "testCert3",
				Description: "test cert",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: module.CaCertSpec{
				IsCA: true,
				IssuerRef: certissuer.IssuerRef{
					Name:  "foo",
					Kind:  "ClusterIssuer",
					Group: "cert-manager.io",
				},
				Duration: "8760h",
				IssuingCluster: module.IssuingClusterInfo{
					Cluster:         "issuer1",
					ClusterProvider: "provider1",
				},
				CertificateSigningInfo: module.CertificateSigningInfo{
					KeySize: 4096,
					Version: 1,
					Algorithm: module.Algorithm{
						PublicKeyAlgorithm: "RSA",
						SignatureAlgorithm: "SHA512WithRSA",
					},
					Subject: module.Subject{
						Locale: module.Locale{},
						Names: module.Names{
							CommonName: "foo",
						},
						Organization: module.Organization{},
					},
				},
			},
		},
		{
			MetaData: types.Metadata{
				Name:        "testCert4",
				Description: "",
				UserData1:   "",
				UserData2:   "",
			},
			Spec: module.CaCertSpec{
				IsCA: true,
				IssuerRef: certissuer.IssuerRef{
					Name:  "foo",
					Kind:  "ClusterIssuer",
					Group: "cert-manager.io",
				},
				Duration: "8760h",
				IssuingCluster: module.IssuingClusterInfo{
					Cluster:         "issuer1",
					ClusterProvider: "provider1",
				},
				CertificateSigningInfo: module.CertificateSigningInfo{
					KeySize: 4096,
					Version: 1,
					Algorithm: module.Algorithm{
						PublicKeyAlgorithm: "RSA",
						SignatureAlgorithm: "SHA512WithRSA",
					},
					Subject: module.Subject{
						Locale: module.Locale{},
						Names: module.Names{
							CommonName: "foo",
						},
						Organization: module.Organization{},
					},
				},
			},
		},
	}
}

func certInput(name string) io.Reader {
	if len(name) == 0 {
		return bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": "",
				"description": "test cert",
			    "userData1": "some user data 1",
			    "userData2": "some user data 2"
			},
			"spec": {
				"isCA": true,
				"issuerRef": {
				  "name": "foo",
				  "kind": "ClusterIssuer",
				  "group": "cert-manager.io"
				},
				"duration": "8760h",
				"issuingCluster": {
				  "cluster": "issuer1",
				  "clusterProvider": "provider1"
				},
				"csrInfo": {
				  "keySize": 4096,
				  "version": 1,
				  "dnsNames": [],
				  "emailAddresses": [],
				  "keyUsages": [],
				  "algorithm": {
					"publicKeyAlgorithm": "RSA",
					"signatureAlgorithm": "SHA512WithRSA"
				  },
				  "subject": {
					"locale": {
					  "country": [],
					  "locality": [],
					  "postalCode": [],
					  "province": [],
					  "streetAddress": []
					},
					"names": {
					  "commonName": "foo"
					},
					"organization": {
						"names": [],
						"units": []
					}
				  }
				}
			  }
		}`))
	}

	return bytes.NewBuffer([]byte(`{
		"metadata": {
			"name": "` + name + `",
			"description": "test cert",
			"userData1": "some user data 1",
			"userData2": "some user data 2"
		  },
		  "spec": {
			"isCA": true,
			"issuerRef": {
			  "name": "foo",
			  "kind": "ClusterIssuer",
			  "group": "cert-manager.io"
			},
			"duration": "8760h",
			"issuingCluster": {
			  "cluster": "issuer1",
			  "clusterProvider": "provider1"
			},
			"csrInfo": {
			  "keySize": 4096,
			  "version": 1,
			  "dnsNames": [],
			  "emailAddresses": [],
			  "keyUsages": [],
			  "algorithm": {
				"publicKeyAlgorithm": "RSA",
				"signatureAlgorithm": "SHA512WithRSA"
			  },
			  "subject": {
				"locale": {
				  "country": [],
				  "locality": [],
				  "postalCode": [],
				  "province": [],
				  "streetAddress": []
				},
				"names": {
				  "commonName": "foo"
				},
				"organization": {
					"names": [],
					"units": []
				}
			  }
			}
		  }
	}`))
}

func certResult(name string) module.CaCert {
	return module.CaCert{
		MetaData: types.Metadata{
			Name:        name,
			Description: "test cert",
			UserData1:   "some user data 1",
			UserData2:   "some user data 2",
		},
		Spec: module.CaCertSpec{
			IsCA: true,
			IssuerRef: certissuer.IssuerRef{
				Name:  "foo",
				Kind:  "ClusterIssuer",
				Group: "cert-manager.io",
			},
			Duration: "8760h",
			IssuingCluster: module.IssuingClusterInfo{
				Cluster:         "issuer1",
				ClusterProvider: "provider1",
			},
			CertificateSigningInfo: module.CertificateSigningInfo{
				KeySize: 4096,
				Version: 1,
				Algorithm: module.Algorithm{
					PublicKeyAlgorithm: "RSA",
					SignatureAlgorithm: "SHA512WithRSA",
				},
				Subject: module.Subject{
					Locale: module.Locale{},
					Names: module.Names{
						CommonName: "foo",
					},
					Organization: module.Organization{},
				},
			},
		},
	}
}

func validateCertResponse(res *http.Response, t test) {
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body) // to retain the content
	if err != nil {
		Fail(err.Error())
	}

	Expect(res.StatusCode).To(Equal(t.statusCode))

	if t.err != nil {
		b := strings.Replace(string(data), "\n", "", -1) // replace the new line at the end
		Expect(b).To(Equal(t.err.Error()))
	}

	result := t.result.(module.CaCert)

	c := module.CaCert{}
	json.NewDecoder(bytes.NewReader(data)).Decode(&c)
	Expect(c).To(Equal(result))
}

func populateClusterGroupTestData() []module.ClusterGroup {
	return []module.ClusterGroup{
		{
			MetaData: types.Metadata{
				Name:        "testClusterGroup-1",
				Description: "test clusterGroup",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: module.ClusterGroupSpec{
				Label:    "edge-1",
				Cluster:  "edge1",
				Scope:    "label",
				Provider: "provider1",
			},
		},
		{
			MetaData: types.Metadata{
				Name:        "testClusterGroup-2",
				Description: "test clusterGroup",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: module.ClusterGroupSpec{
				Label:    "edge-1",
				Cluster:  "edge1",
				Scope:    "label",
				Provider: "provider1",
			},
		},
		{
			MetaData: types.Metadata{
				Name:        "testClusterGroup-3",
				Description: "test clusterGroup",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: module.ClusterGroupSpec{
				Label:    "edge-1",
				Cluster:  "edge1",
				Scope:    "label",
				Provider: "provider1",
			},
		},
		{
			MetaData: types.Metadata{
				Name:        "testClusterGroup-4",
				Description: "",
				UserData1:   "",
				UserData2:   "",
			},
			Spec: module.ClusterGroupSpec{
				Label:    "edge-1",
				Cluster:  "edge1",
				Scope:    "label",
				Provider: "provider1",
			},
		},
	}
}

func clusterGroupInput(name string) io.Reader {
	if len(name) == 0 {
		return bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": ""
			},
			"spec": {
				"label": "edge-1",
				"cluster": "edge1",
				"scope": "label",
				"clusterProvider": "provider1"
			}
		}`))
	}

	return bytes.NewBuffer([]byte(`{
		"metadata": {
			"name": "` + name + `",
			"description": "test clusterGroup",
			"userData1": "some user data 1",
			"userData2": "some user data 2"
		},
		"spec": {
			"label": "edge-1",
			"cluster": "edge1",
			"scope": "label",
			"clusterProvider": "provider1"
		}
	}`))
}

func clusterGroupResult(name string) module.ClusterGroup {
	return module.ClusterGroup{
		MetaData: types.Metadata{
			Name:        name,
			Description: "test clusterGroup",
			UserData1:   "some user data 1",
			UserData2:   "some user data 2",
		},
		Spec: module.ClusterGroupSpec{
			Label:    "edge-1",
			Cluster:  "edge1",
			Scope:    "label",
			Provider: "provider1",
		},
	}
}

func validateClusterGroupResponse(res *http.Response, t test) {
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body) // to retain the content
	if err != nil {
		Fail(err.Error())
	}

	Expect(res.StatusCode).To(Equal(t.statusCode))

	if t.err != nil {
		b := strings.Replace(string(data), "\n", "", -1) // replace the new line at the end
		Expect(b).To(Equal(t.err.Error()))
	}

	result := t.result.(module.ClusterGroup)

	c := module.ClusterGroup{}
	json.NewDecoder(bytes.NewReader(data)).Decode(&c)
	Expect(c).To(Equal(result))
}
