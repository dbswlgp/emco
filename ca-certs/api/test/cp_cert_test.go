// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// These test cases are to validate the route handler functionalities
package api_test

import (
	"context"
	"errors"
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/api"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

type mockClusterProviderCertManager struct {
	Items []module.CaCert
	Err   error
}

func init() {
	api.CertificateSchemaJson = "../../json-schemas/certificate.json"
}

func (m *mockClusterProviderCertManager) CreateCert(ctx context.Context, cert module.CaCert, clusterProvider string, failIfExists bool) (module.CaCert, bool, error) {
	iExists := false
	index := 0

	if m.Err != nil {
		return module.CaCert{}, iExists, m.Err
	}

	for i, item := range m.Items {
		if item.MetaData.Name == cert.MetaData.Name {
			iExists = true
			index = i
			break
		}
	}

	if iExists && failIfExists { // cert already exists
		return module.CaCert{}, iExists,
			emcoerror.NewEmcoError(
				module.CaCertAlreadyExists,
				emcoerror.Conflict,
			)
	}

	if iExists && !failIfExists { // cert already exists. update the cert
		m.Items[index] = cert
		return m.Items[index], iExists, nil
	}

	m.Items = append(m.Items, cert) // create the cert

	return m.Items[len(m.Items)-1], iExists, nil

}
func (m *mockClusterProviderCertManager) DeleteCert(ctx context.Context, cert, clusterProvider string) error {
	if m.Err != nil {
		return m.Err
	}

	for k, item := range m.Items {
		if item.MetaData.Name == cert { // cert exist
			m.Items[k] = m.Items[len(m.Items)-1]
			m.Items[len(m.Items)-1] = module.CaCert{}
			m.Items = m.Items[:len(m.Items)-1]
			return nil
		}
	}

	return emcoerror.NewEmcoError(
		"The requested resource not found",
		emcoerror.NotFound,
	)
}

func (m *mockClusterProviderCertManager) GetAllCert(ctx context.Context, clusterProvider string) ([]module.CaCert, error) {
	if m.Err != nil {
		return []module.CaCert{}, m.Err
	}

	var certs []module.CaCert
	certs = append(certs, m.Items...)

	return certs, nil

}
func (m *mockClusterProviderCertManager) GetCert(ctx context.Context, cert, clusterProvider string) (module.CaCert, error) {
	if m.Err != nil {
		return module.CaCert{}, m.Err
	}

	for _, item := range m.Items {
		if item.MetaData.Name == cert {
			return item, nil
		}
	}

	return module.CaCert{}, emcoerror.NewEmcoError(
		module.CaCertNotFound,
		emcoerror.NotFound,
	)
}

var _ = Describe("Test create cert handler",
	func() {
		DescribeTable("Create Cert",
			func(t test) {
				client := t.client.(*mockClusterProviderCertManager)
				res := executeRequest(http.MethodPost, "", clusterProviderCertURL, client, t.input)
				validateCertResponse(res, t)
			},
			Entry("request body validation",
				test{
					entry:      "request body validation",
					input:      certInput(""), // create an empty cert payload
					result:     module.CaCert{},
					err:        errors.New("caCert name may not be empty"),
					statusCode: http.StatusBadRequest,
					client: &mockClusterProviderCertManager{
						Err:   nil,
						Items: populateCertTestData(),
					},
				},
			),
			Entry("successful create",
				test{
					entry:      "successful create",
					input:      certInput("testCert"),
					result:     certResult("testCert"),
					err:        nil,
					statusCode: http.StatusCreated,
					client: &mockClusterProviderCertManager{
						Err:   nil,
						Items: populateCertTestData(),
					},
				},
			),
			Entry("cert already exists",
				test{
					entry:  "cert already exists",
					input:  certInput("testCert1"),
					result: module.CaCert{},
					err: emcoerror.NewEmcoError(
						module.CaCertAlreadyExists,
						emcoerror.Conflict,
					),
					statusCode: http.StatusConflict,
					client: &mockClusterProviderCertManager{
						Err:   nil,
						Items: populateCertTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test get cert handler",
	func() {
		DescribeTable("Get Cert",
			func(t test) {
				client := t.client.(*mockClusterProviderCertManager)
				res := executeRequest(http.MethodGet, "/"+t.name, clusterProviderCertURL, client, nil)
				validateCertResponse(res, t)
			},
			Entry("successful get",
				test{
					entry:      "successful get",
					name:       "testCert1",
					statusCode: http.StatusOK,
					err:        nil,
					result:     certResult("testCert1"),
					client: &mockClusterProviderCertManager{
						Err:   nil,
						Items: populateCertTestData(),
					},
				},
			),
			Entry("cert not found",
				test{
					entry:      "cert not found",
					name:       "nonExistingCert",
					statusCode: http.StatusNotFound,
					err: emcoerror.NewEmcoError(
						module.CaCertNotFound,
						emcoerror.NotFound,
					),
					result: module.CaCert{},
					client: &mockClusterProviderCertManager{
						Err:   nil,
						Items: populateCertTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test update cert handler",
	func() {
		DescribeTable("Update Cert",
			func(t test) {
				client := t.client.(*mockClusterProviderCertManager)
				res := executeRequest(http.MethodPut, "/"+t.name, clusterProviderCertURL, client, t.input)
				validateCertResponse(res, t)
			},
			Entry("request body validation",
				test{
					entry:      "request body validation",
					name:       "testCert",
					input:      certInput(""), // create an empty cert payload
					result:     module.CaCert{},
					err:        errors.New("caCert name may not be empty"),
					statusCode: http.StatusBadRequest,
					client: &mockClusterProviderCertManager{
						Err:   nil,
						Items: populateCertTestData(),
					},
				},
			),
			Entry("successful update",
				test{
					entry:      "successful update",
					name:       "testCert",
					input:      certInput("testCert"),
					result:     certResult("testCert"),
					err:        nil,
					statusCode: http.StatusCreated,
					client: &mockClusterProviderCertManager{
						Err:   nil,
						Items: populateCertTestData(),
					},
				},
			),
			Entry("cert already exists",
				test{
					entry:      "cert already exists",
					name:       "testCert4",
					input:      certInput("testCert4"),
					result:     certResult("testCert4"),
					err:        nil,
					statusCode: http.StatusOK,
					client: &mockClusterProviderCertManager{
						Err:   nil,
						Items: populateCertTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test delete cert handler",
	func() {
		DescribeTable("Delete Cert",
			func(t test) {
				client := t.client.(*mockClusterProviderCertManager)
				res := executeRequest(http.MethodDelete, "/"+t.name, clusterProviderCertURL, client, nil)
				validateCertResponse(res, t)
			},
			Entry("successful delete",
				test{
					entry:      "successful delete",
					name:       "testCert1",
					statusCode: http.StatusNoContent,
					err:        nil,
					result:     module.CaCert{},
					client: &mockClusterProviderCertManager{
						Err:   nil,
						Items: populateCertTestData(),
					},
				},
			),
			Entry("db remove cert not found",
				test{
					entry:      "db remove cert not found",
					name:       "nonExistingCert",
					statusCode: http.StatusNotFound,
					err: emcoerror.NewEmcoError(
						"The requested resource not found",
						emcoerror.NotFound,
					),
					result: module.CaCert{},
					client: &mockClusterProviderCertManager{
						Err:   nil,
						Items: populateCertTestData(),
					},
				},
			),
		)
	},
)
