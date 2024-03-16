// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// These test cases are to validate the route handler functionalities
package api_test

import (
	"context"
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/api"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
)

type mockClusterProviderClusterManager struct {
	Items []module.ClusterGroup
	Err   error
}

func init() {
	api.ClusterSchemaJson = "../../json-schemas/cluster.json"
}

// CreateClusterGroup
func (m *mockClusterProviderClusterManager) CreateClusterGroup(ctx context.Context, cluster module.ClusterGroup, cert, clusterProvider string, failIfExists bool) (module.ClusterGroup, bool, error) {
	if m.Err != nil {
		return module.ClusterGroup{}, false, m.Err
	}

	iExists := false
	index := 0

	for i, item := range m.Items {
		if item.MetaData.Name == cluster.MetaData.Name {
			iExists = true
			index = i
			break
		}
	}

	if iExists && failIfExists { // clusterGroup already exists
		return module.ClusterGroup{}, iExists, emcoerror.NewEmcoError(
			module.CaCertClusterGroupAlreadyExists,
			emcoerror.Conflict,
		)
	}

	if iExists && !failIfExists { // clusterGroup already exists. update the clusterGroup
		m.Items[index] = cluster
		return m.Items[index], iExists, nil
	}

	m.Items = append(m.Items, cluster) // create the clusterGroup

	return m.Items[len(m.Items)-1], iExists, nil
}

// DeleteClusterGroup
func (m *mockClusterProviderClusterManager) DeleteClusterGroup(ctx context.Context, cert, cluster, clusterProvider string) error {
	if m.Err != nil {
		return m.Err
	}

	for k, item := range m.Items {
		if item.MetaData.Name == cluster { // clusterGroup exist
			m.Items[k] = m.Items[len(m.Items)-1]
			m.Items[len(m.Items)-1] = module.ClusterGroup{}
			m.Items = m.Items[:len(m.Items)-1]
			return nil
		}
	}

	return emcoerror.NewEmcoError(
		"The requested resource not found",
		emcoerror.NotFound,
	) // clusterGroup does not exist
}

// GetAllClusterGroups
func (m *mockClusterProviderClusterManager) GetAllClusterGroups(ctx context.Context, cert, clusterProvider string) ([]module.ClusterGroup, error) {
	if m.Err != nil {
		return []module.ClusterGroup{}, m.Err
	}

	var clusterGroups []module.ClusterGroup
	for _, item := range m.Items {
		c := item
		clusterGroups = append(clusterGroups, c)
	}

	return clusterGroups, nil
}

// GetClusterGroup
func (m *mockClusterProviderClusterManager) GetClusterGroup(ctx context.Context, cert, cluster, clusterProvider string) (module.ClusterGroup, error) {
	if m.Err != nil {
		return module.ClusterGroup{}, m.Err
	}

	for _, item := range m.Items {
		if item.MetaData.Name == cluster {
			return item, nil
		}
	}

	return module.ClusterGroup{}, emcoerror.NewEmcoError(
		module.CaCertClusterGroupNotFound,
		emcoerror.NotFound,
	)
}

var _ = Describe("Test create cluster handler",
	func() {
		DescribeTable("Create ClusterGroup",
			func(t test) {
				client := t.client.(*mockClusterProviderClusterManager)
				res := executeRequest(http.MethodPost, "/{caCert}/clusters", clusterProviderCertURL, client, t.input)
				validateClusterGroupResponse(res, t)
			},
			Entry("request body validation",
				test{
					entry:      "request body validation",
					input:      clusterGroupInput(""), // create an empty clusterGroup payload
					result:     module.ClusterGroup{},
					err:        errors.New("clusterGroup name may not be empty"),
					statusCode: http.StatusBadRequest,
					client: &mockClusterProviderClusterManager{
						Err:   nil,
						Items: populateClusterGroupTestData(),
					},
				},
			),
			Entry("successful create",
				test{
					entry:      "successful create",
					input:      clusterGroupInput("testClusterGroup"),
					result:     clusterGroupResult("testClusterGroup"),
					err:        nil,
					statusCode: http.StatusCreated,
					client: &mockClusterProviderClusterManager{
						Err:   nil,
						Items: populateClusterGroupTestData(),
					},
				},
			),
			Entry("clusterGroup already exists",
				test{
					entry:  "clusterGroup already exists",
					input:  clusterGroupInput("testClusterGroup-1"),
					result: module.ClusterGroup{},
					err: emcoerror.NewEmcoError(
						module.CaCertClusterGroupAlreadyExists,
						emcoerror.Conflict,
					),
					statusCode: http.StatusConflict,
					client: &mockClusterProviderClusterManager{
						Err:   nil,
						Items: populateClusterGroupTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test get clusterGroup handler",
	func() {
		DescribeTable("Get ClusterGroup",
			func(t test) {
				client := t.client.(*mockClusterProviderClusterManager)
				res := executeRequest(http.MethodGet, "/{caCert}/clusters/"+t.name, clusterProviderCertURL, client, nil)
				validateClusterGroupResponse(res, t)
			},
			Entry("successful get",
				test{
					name:       "testClusterGroup-1",
					statusCode: http.StatusOK,
					err:        nil,
					result:     clusterGroupResult("testClusterGroup-1"),
					client: &mockClusterProviderClusterManager{
						Err:   nil,
						Items: populateClusterGroupTestData(),
					},
				},
			),
			Entry("clusterGroup not found",
				test{
					name:       "nonExistingClusterGroup",
					statusCode: http.StatusNotFound,
					err: emcoerror.NewEmcoError(
						module.CaCertClusterGroupNotFound,
						emcoerror.NotFound,
					),
					result: module.ClusterGroup{},
					client: &mockClusterProviderClusterManager{
						Err:   nil,
						Items: populateClusterGroupTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test update clusterGroup handler",
	func() {
		DescribeTable("Update ClusterGroup",
			func(t test) {
				client := t.client.(*mockClusterProviderClusterManager)
				res := executeRequest(http.MethodPut, "/{caCert}/clusters/"+t.name, clusterProviderCertURL, client, t.input)
				validateClusterGroupResponse(res, t)
			},
			Entry("request body validation",
				test{
					entry:      "request body validation",
					name:       "testClusterGroup",
					input:      clusterGroupInput(""), // create an empty clusterGroup payload
					result:     module.ClusterGroup{},
					err:        errors.New("clusterGroup name may not be empty"),
					statusCode: http.StatusBadRequest,
					client: &mockClusterProviderClusterManager{
						Err:   nil,
						Items: populateClusterGroupTestData(),
					},
				},
			),
			Entry("successful update",
				test{
					entry:      "successful update",
					name:       "testClusterGroup",
					input:      clusterGroupInput("testClusterGroup"),
					result:     clusterGroupResult("testClusterGroup"),
					err:        nil,
					statusCode: http.StatusCreated,
					client: &mockClusterProviderClusterManager{
						Err:   nil,
						Items: populateClusterGroupTestData(),
					},
				},
			),
			Entry("clusterGroup already exists",
				test{
					entry:      "clusterGroup already exists",
					name:       "testClusterGroup-4",
					input:      clusterGroupInput("testClusterGroup-4"),
					result:     clusterGroupResult("testClusterGroup-4"),
					err:        nil,
					statusCode: http.StatusOK,
					client: &mockClusterProviderClusterManager{
						Err:   nil,
						Items: populateClusterGroupTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test delete clusterGroup handler",
	func() {
		DescribeTable("Delete ClusterGroup",
			func(t test) {
				client := t.client.(*mockClusterProviderClusterManager)
				res := executeRequest(http.MethodDelete, "/{caCert}/clusters/"+t.name, clusterProviderCertURL, client, nil)
				validateClusterGroupResponse(res, t)
			},
			Entry("successful delete",
				test{
					entry:      "successful delete",
					name:       "testClusterGroup-1",
					statusCode: http.StatusNoContent,
					err:        nil,
					result:     module.ClusterGroup{},
					client: &mockClusterProviderClusterManager{
						Err:   nil,
						Items: populateClusterGroupTestData(),
					},
				},
			),
			Entry("db remove clusterGroup not found",
				test{
					entry:      "db remove clusterGroup not found",
					name:       "nonExistingClusterGroup",
					statusCode: http.StatusNotFound,
					err: emcoerror.NewEmcoError(
						"The requested resource not found",
						emcoerror.NotFound,
					),
					result: module.ClusterGroup{},
					client: &mockClusterProviderClusterManager{
						Err:   nil,
						Items: populateClusterGroupTestData(),
					},
				},
			),
		)
	},
)
