// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// These test cases are to validate the route handler functionalities
package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/api"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

type mockLogicalCloudManager struct {
	Items []logicalcloud.CaCertLogicalCloud
	Err   error
}

func init() {
	api.LogicalCloudSchemaJson = "../../json-schemas/logicalCloud.json"
}

func (m *mockLogicalCloudManager) CreateLogicalCloud(ctx context.Context, logicalCloud logicalcloud.CaCertLogicalCloud, cert, project string, failIfExists bool) (logicalcloud.CaCertLogicalCloud, bool, error) {
	iExists := false
	index := 0

	if m.Err != nil {
		return logicalcloud.CaCertLogicalCloud{}, iExists, m.Err
	}

	for i, item := range m.Items {
		if item.MetaData.Name == logicalCloud.MetaData.Name {
			iExists = true
			index = i
			break
		}
	}

	if iExists && failIfExists { // logicalCloud already exists
		return logicalcloud.CaCertLogicalCloud{}, iExists,
			emcoerror.NewEmcoError(
				module.CaCertLogicalCloudAlreadyExists,
				emcoerror.Conflict,
			)
	}

	if iExists && !failIfExists { // logicalCloud already exists. update the logicalCloud
		m.Items[index] = logicalCloud
		return m.Items[index], iExists, nil
	}

	m.Items = append(m.Items, logicalCloud) // create the logicalCloud

	return m.Items[len(m.Items)-1], iExists, nil

}
func (m *mockLogicalCloudManager) DeleteLogicalCloud(ctx context.Context, logicalCloud, cert, project string) error {
	if m.Err != nil {
		return m.Err
	}

	for k, item := range m.Items {
		if item.MetaData.Name == logicalCloud { // logicalCloud exist
			m.Items[k] = m.Items[len(m.Items)-1]
			m.Items[len(m.Items)-1] = logicalcloud.CaCertLogicalCloud{}
			m.Items = m.Items[:len(m.Items)-1]
			return nil
		}
	}

	return emcoerror.NewEmcoError(
		"The requested resource not found",
		emcoerror.NotFound,
	) // logicalCloud does not exist

}

func (m *mockLogicalCloudManager) GetAllLogicalClouds(ctx context.Context, cert, project string) ([]logicalcloud.CaCertLogicalCloud, error) {
	if m.Err != nil {
		return []logicalcloud.CaCertLogicalCloud{}, m.Err
	}

	var certs []logicalcloud.CaCertLogicalCloud
	certs = append(certs, m.Items...)

	return certs, nil

}
func (m *mockLogicalCloudManager) GetLogicalCloud(ctx context.Context, logicalCloud, cert, project string) (logicalcloud.CaCertLogicalCloud, error) {
	if m.Err != nil {
		return logicalcloud.CaCertLogicalCloud{}, m.Err
	}

	for _, item := range m.Items {
		if item.MetaData.Name == logicalCloud {
			return item, nil
		}
	}

	return logicalcloud.CaCertLogicalCloud{}, emcoerror.NewEmcoError(
		module.CaCertLogicalCloudNotFound,
		emcoerror.NotFound,
	)
}

var _ = Describe("Test create logical-cloud handler",
	func() {
		DescribeTable("Create LogicalCloud",
			func(t test) {
				client := t.client.(*mockLogicalCloudManager)
				res := executeRequest(http.MethodPost, "/{caCert}/logical-clouds", logicalCloudCertURL, client, t.input)
				validateLogicalCloudResponse(res, t)
			},
			Entry("request body validation",
				test{
					input:      logicalCLoudInput(""), // create an empty logicalCloud payload
					result:     logicalcloud.CaCertLogicalCloud{},
					err:        errors.New("caCert logicalCloud name may not be empty"),
					statusCode: http.StatusBadRequest,
					client: &mockLogicalCloudManager{
						Err:   nil,
						Items: populateLogicalCloudTestData(),
					},
				},
			),
			Entry("successful create",
				test{
					input:      logicalCLoudInput("testLogicalCloud"),
					result:     logicalCLoudResult("testLogicalCloud"),
					err:        nil,
					statusCode: http.StatusCreated,
					client: &mockLogicalCloudManager{
						Err:   nil,
						Items: populateLogicalCloudTestData(),
					},
				},
			),
			Entry("logicalCloud already exists",
				test{
					input:  logicalCLoudInput("testLogicalCloud-1"),
					result: logicalcloud.CaCertLogicalCloud{},
					err: emcoerror.NewEmcoError(
						module.CaCertLogicalCloudAlreadyExists,
						emcoerror.Conflict,
					),
					statusCode: http.StatusConflict,
					client: &mockLogicalCloudManager{
						Err:   nil,
						Items: populateLogicalCloudTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test get logicalCloud handler",
	func() {
		DescribeTable("Get LogicalCloud",
			func(t test) {
				client := t.client.(*mockLogicalCloudManager)
				res := executeRequest(http.MethodGet, "/{caCert}/logical-clouds/"+t.name, logicalCloudCertURL, client, nil)
				validateLogicalCloudResponse(res, t)
			},
			Entry("successful get",
				test{
					name:       "testLogicalCloud-1",
					statusCode: http.StatusOK,
					err:        nil,
					result:     logicalCLoudResult("testLogicalCloud-1"),
					client: &mockLogicalCloudManager{
						Err:   nil,
						Items: populateLogicalCloudTestData(),
					},
				},
			),
			Entry("logicalCloud not found",
				test{
					name:       "nonExistingLogicalCloud",
					statusCode: http.StatusNotFound,
					err: emcoerror.NewEmcoError(
						module.CaCertLogicalCloudNotFound,
						emcoerror.NotFound,
					),
					result: logicalcloud.CaCertLogicalCloud{},
					client: &mockLogicalCloudManager{
						Err:   nil,
						Items: populateLogicalCloudTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test update logicalCloud handler",
	func() {
		DescribeTable("Update LogicalCloud",
			func(t test) {
				client := t.client.(*mockLogicalCloudManager)
				res := executeRequest(http.MethodPut, "/{caCert}/logical-clouds/"+t.name, logicalCloudCertURL, client, t.input)
				validateLogicalCloudResponse(res, t)
			},
			Entry("request body validation",
				test{
					entry:      "request body validation",
					name:       "testLogicalCloud",
					input:      logicalCLoudInput(""), // create an empty logicalCloud payload
					result:     logicalcloud.CaCertLogicalCloud{},
					err:        errors.New("caCert logicalCloud name may not be empty"),
					statusCode: http.StatusBadRequest,
					client: &mockLogicalCloudManager{
						Err:   nil,
						Items: populateLogicalCloudTestData(),
					},
				},
			),
			Entry("successful update",
				test{
					entry:      "successful update",
					name:       "testLogicalCloud",
					input:      logicalCLoudInput("testLogicalCloud"),
					result:     logicalCLoudResult("testLogicalCloud"),
					err:        nil,
					statusCode: http.StatusCreated,
					client: &mockLogicalCloudManager{
						Err:   nil,
						Items: populateLogicalCloudTestData(),
					},
				},
			),
			Entry("logicalCloud already exists",
				test{
					entry:      "logicalCloud already exists",
					name:       "testLogicalCloud-4",
					input:      logicalCLoudInput("testLogicalCloud-4"),
					result:     logicalCLoudResult("testLogicalCloud-4"),
					err:        nil,
					statusCode: http.StatusOK,
					client: &mockLogicalCloudManager{
						Err:   nil,
						Items: populateLogicalCloudTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test delete logicalCloud handler",
	func() {
		DescribeTable("Delete LogicalCloud",
			func(t test) {
				client := t.client.(*mockLogicalCloudManager)
				res := executeRequest(http.MethodDelete, "/{caCert}/logical-clouds/"+t.name, logicalCloudCertURL, client, nil)
				validateLogicalCloudResponse(res, t)
			},
			Entry("successful delete",
				test{
					entry:      "successful delete",
					name:       "testLogicalCloud-1",
					statusCode: http.StatusNoContent,
					err:        nil,
					result:     logicalcloud.CaCertLogicalCloud{},
					client: &mockLogicalCloudManager{
						Err:   nil,
						Items: populateLogicalCloudTestData(),
					},
				},
			),
			Entry("db remove logicalCloud not found",
				test{
					entry:      "db remove logicalCloud not found",
					name:       "nonExistingLogicalCloud",
					statusCode: http.StatusNotFound,
					err: emcoerror.NewEmcoError(
						"The requested resource not found",
						emcoerror.NotFound,
					),
					result: logicalcloud.CaCertLogicalCloud{},
					client: &mockLogicalCloudManager{
						Err:   nil,
						Items: populateLogicalCloudTestData(),
					},
				},
			),
		)
	},
)

func populateLogicalCloudTestData() []logicalcloud.CaCertLogicalCloud {
	return []logicalcloud.CaCertLogicalCloud{
		{
			MetaData: types.Metadata{
				Name:        "testLogicalCloud-1",
				Description: "test logicalCloud",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: logicalcloud.CaCertLogicalCloudSpec{
				LogicalCloud: "lc1",
			},
		},
		{
			MetaData: types.Metadata{
				Name:        "testLogicalCloud-2",
				Description: "test logicalCloud",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: logicalcloud.CaCertLogicalCloudSpec{
				LogicalCloud: "lc1",
			},
		},
		{
			MetaData: types.Metadata{
				Name:        "testLogicalCloud-3",
				Description: "test logicalCloud",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: logicalcloud.CaCertLogicalCloudSpec{
				LogicalCloud: "lc1",
			},
		},
		{
			MetaData: types.Metadata{
				Name:        "testLogicalCloud-4",
				Description: "",
				UserData1:   "",
				UserData2:   "",
			},
			Spec: logicalcloud.CaCertLogicalCloudSpec{
				LogicalCloud: "lc1",
			},
		},
	}
}

func logicalCLoudInput(name string) io.Reader {
	if len(name) == 0 {
		return bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": ""
			},
			"spec": {
				"logicalCloud": "lc1"
			}
		}`))
	}

	return bytes.NewBuffer([]byte(`{
		"metadata": {
			"name": "` + name + `",
			"description": "test logicalCloud",
			"userData1": "some user data 1",
			"userData2": "some user data 2"
		},
		"spec": {
			"logicalCloud": "lc1"
		}
	}`))
}

func logicalCLoudResult(name string) logicalcloud.CaCertLogicalCloud {
	return logicalcloud.CaCertLogicalCloud{
		MetaData: types.Metadata{
			Name:        name,
			Description: "test logicalCloud",
			UserData1:   "some user data 1",
			UserData2:   "some user data 2",
		},
		Spec: logicalcloud.CaCertLogicalCloudSpec{
			LogicalCloud: "lc1",
		},
	}
}

func validateLogicalCloudResponse(res *http.Response, t test) {
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

	result := t.result.(logicalcloud.CaCertLogicalCloud)

	lc := logicalcloud.CaCertLogicalCloud{}
	json.NewDecoder(bytes.NewReader(data)).Decode(&lc)
	Expect(lc).To(Equal(result))
}
