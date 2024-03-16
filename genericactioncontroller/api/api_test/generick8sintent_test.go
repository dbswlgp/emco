// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// These test cases are to validate the route handler functionalities
package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/api"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

type mockGenericK8sIntentManager struct {
	Items []module.GenericK8sIntent
	Err   error
}

func init() {
	api.GenericK8sIntentSchemaJson = "../../json-schemas/genericK8sIntent.json"
}

func (m *mockGenericK8sIntentManager) CreateGenericK8sIntent(ctx context.Context, gki module.GenericK8sIntent,
	project, compositeApp, compositeAppVersion, deploymentIntentGroup string,
	failIfExists bool) (module.GenericK8sIntent, bool, error) {
	iExists := false
	index := 0

	if m.Err != nil {
		return module.GenericK8sIntent{}, iExists, m.Err
	}

	for i, item := range m.Items {
		if item.Metadata.Name == gki.Metadata.Name {
			iExists = true
			index = i
			break
		}
	}

	if iExists && failIfExists { // genericK8sIntent already exists
		return module.GenericK8sIntent{}, iExists, emcoerror.NewEmcoError(
			module.GenericK8sIntentAlreadyExists,
			emcoerror.Conflict,
		)
	}

	if iExists && !failIfExists { // genericK8sIntent already exists. update the genericK8sIntent
		m.Items[index] = gki
		return m.Items[index], iExists, nil
	}

	m.Items = append(m.Items, gki) // create the genericK8sIntent

	return m.Items[len(m.Items)-1], iExists, nil
}

func (m *mockGenericK8sIntentManager) DeleteGenericK8sIntent(ctx context.Context, intent, project, compositeApp, compositeAppVersion, deploymentIntentGroup string) error {
	if m.Err != nil {
		return m.Err
	}

	for k, item := range m.Items {
		if item.Metadata.Name == intent { // genericK8sIntent exist
			m.Items[k] = m.Items[len(m.Items)-1]
			m.Items[len(m.Items)-1] = module.GenericK8sIntent{}
			m.Items = m.Items[:len(m.Items)-1]
			return nil
		}
	}

	return emcoerror.NewEmcoError(
		"the requested resource not found",
		emcoerror.NotFound,
	) // genericK8sIntent does not exist
}

func (m *mockGenericK8sIntentManager) GetAllGenericK8sIntents(ctx context.Context, project, compositeApp, compositeAppVersion, deploymentIntentGroup string) ([]module.GenericK8sIntent, error) {

	if m.Err != nil {
		return []module.GenericK8sIntent{}, m.Err
	}

	var intents []module.GenericK8sIntent
	intents = append(intents, m.Items...)

	return intents, nil
}

func (m *mockGenericK8sIntentManager) GetGenericK8sIntent(ctx context.Context, intent, project, compositeApp, compositeAppVersion, deploymentIntentGroup string) (module.GenericK8sIntent, error) {

	if m.Err != nil {
		return module.GenericK8sIntent{}, m.Err
	}

	for _, item := range m.Items {
		if item.Metadata.Name == intent {
			return item, nil
		}
	}

	return module.GenericK8sIntent{}, emcoerror.NewEmcoError(
		module.GenericK8sIntentNotFound,
		emcoerror.NotFound,
	)
}

var _ = Describe("Test create genericK8sIntent handler",
	func() {
		DescribeTable("Create GenericK8sIntent",
			func(t test) {
				client := t.client.(*mockGenericK8sIntentManager)
				res := executeRequest(http.MethodPost, "", nil, client, t.input)
				validategenericK8sIntentResponse(res, t)
			},
			Entry("request body validation",
				test{
					input:      genericK8sIntentInput(""), // create an empty genericK8sIntent payload
					result:     module.GenericK8sIntent{},
					err:        errors.New("genericK8sIntent name may not be empty"),
					statusCode: http.StatusBadRequest,
					client: &mockGenericK8sIntentManager{
						Err:   nil,
						Items: populateGenericK8sIntentTestData(),
					},
				},
			),
			Entry("successful create",
				test{
					input:      genericK8sIntentInput("testGenericK8sIntent"),
					result:     genericK8sIntentResult("testGenericK8sIntent"),
					err:        nil,
					statusCode: http.StatusCreated,
					client: &mockGenericK8sIntentManager{
						Err:   nil,
						Items: populateGenericK8sIntentTestData(),
					},
				},
			),
			Entry("genericK8sIntent already exists",
				test{
					input:  genericK8sIntentInput("testGenericK8sIntent-1"),
					result: module.GenericK8sIntent{},
					err: emcoerror.NewEmcoError(
						module.GenericK8sIntentAlreadyExists,
						emcoerror.Conflict,
					),
					statusCode: http.StatusConflict,
					client: &mockGenericK8sIntentManager{
						Err:   nil,
						Items: populateGenericK8sIntentTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test get genericK8sIntent handler",
	func() {
		DescribeTable("Get GenericK8sIntent",
			func(t test) {
				client := t.client.(*mockGenericK8sIntentManager)
				header := map[string]string{
					"Accept": "application/json",
				}
				res := executeRequest(http.MethodGet, "/"+t.name, header, client, nil)
				validategenericK8sIntentResponse(res, t)
			},
			Entry("successful get",
				test{
					name:       "testGenericK8sIntent-1",
					statusCode: http.StatusOK,
					err:        nil,
					result:     genericK8sIntentResult("testGenericK8sIntent-1"),
					client: &mockGenericK8sIntentManager{
						Err:   nil,
						Items: populateGenericK8sIntentTestData(),
					},
				},
			),
			Entry("genericK8sIntent not found",
				test{
					name:       "nonExistingGenericK8sIntent",
					statusCode: http.StatusNotFound,
					err: emcoerror.NewEmcoError(
						module.GenericK8sIntentNotFound,
						emcoerror.NotFound,
					),
					result: module.GenericK8sIntent{},
					client: &mockGenericK8sIntentManager{
						Err:   nil,
						Items: populateGenericK8sIntentTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test update genericK8sIntent handler",
	func() {
		DescribeTable("Update GenericK8sIntent",
			func(t test) {
				client := t.client.(*mockGenericK8sIntentManager)
				res := executeRequest(http.MethodPut, "/"+t.name, nil, client, t.input)
				validategenericK8sIntentResponse(res, t)
			},
			Entry("request body validation",
				test{
					entry:      "request body validation",
					name:       "testGenericK8sIntent",
					input:      genericK8sIntentInput(""), // create an empty genericK8sIntent payload
					result:     module.GenericK8sIntent{},
					err:        errors.New("genericK8sIntent name may not be empty"),
					statusCode: http.StatusBadRequest,
					client: &mockGenericK8sIntentManager{
						Err:   nil,
						Items: populateGenericK8sIntentTestData(),
					},
				},
			),
			Entry("successful update",
				test{
					entry:      "successful update",
					name:       "testGenericK8sIntent",
					input:      genericK8sIntentInput("testGenericK8sIntent"),
					result:     genericK8sIntentResult("testGenericK8sIntent"),
					err:        nil,
					statusCode: http.StatusCreated,
					client: &mockGenericK8sIntentManager{
						Err:   nil,
						Items: populateGenericK8sIntentTestData(),
					},
				},
			),
			Entry("genericK8sIntent already exists",
				test{
					entry:      "generick8sintent already exists",
					name:       "testGenericK8sIntent-4",
					input:      genericK8sIntentInput("testGenericK8sIntent-4"),
					result:     genericK8sIntentResult("testGenericK8sIntent-4"),
					err:        nil,
					statusCode: http.StatusOK,
					client: &mockGenericK8sIntentManager{
						Err:   nil,
						Items: populateGenericK8sIntentTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test delete genericK8sIntent handler",
	func() {
		DescribeTable("Delete GenericK8sIntent",
			func(t test) {
				client := t.client.(*mockGenericK8sIntentManager)
				res := executeRequest(http.MethodDelete, "/"+t.name, nil, client, nil)
				validategenericK8sIntentResponse(res, t)
			},
			Entry("successful delete",
				test{
					entry:      "successful delete",
					name:       "testGenericK8sIntent-1",
					statusCode: http.StatusNoContent,
					err:        nil,
					result:     module.GenericK8sIntent{},
					client: &mockGenericK8sIntentManager{
						Err:   nil,
						Items: populateGenericK8sIntentTestData(),
					},
				},
			),
			Entry("db remove genericK8sIntent not found",
				test{
					entry:      "db remove genericK8sIntent not found",
					name:       "nonExistingGenericK8sIntent",
					statusCode: http.StatusNotFound,
					err: emcoerror.NewEmcoError(
						"the requested resource not found",
						emcoerror.NotFound,
					),
					result: module.GenericK8sIntent{},
					client: &mockGenericK8sIntentManager{
						Err:   nil,
						Items: populateGenericK8sIntentTestData(),
					},
				},
			),
		)
	},
)

func populateGenericK8sIntentTestData() []module.GenericK8sIntent {
	return []module.GenericK8sIntent{
		{
			Metadata: types.Metadata{
				Name:        "testGenericK8sIntent-1",
				Description: "test genericK8sIntent",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
		},
		{
			Metadata: types.Metadata{
				Name:        "testGenericK8sIntent-2",
				Description: "test genericK8sIntent",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
		},
		{
			Metadata: types.Metadata{
				Name:        "testGenericK8sIntent-3",
				Description: "test genericK8sIntent",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
		},
		{
			Metadata: types.Metadata{
				Name:        "testGenericK8sIntent-4",
				Description: "",
				UserData1:   "",
				UserData2:   "",
			},
		},
	}
}

func genericK8sIntentInput(name string) io.Reader {
	if len(name) == 0 {
		return bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": ""
			}
		}`))
	}

	return bytes.NewBuffer([]byte(`{
		"metadata": {
			"name": "` + name + `",
			"description": "test genericK8sIntent",
			"userData1": "some user data 1",
			"userData2": "some user data 2"
		}
	}`))
}

func genericK8sIntentResult(name string) module.GenericK8sIntent {
	return module.GenericK8sIntent{
		Metadata: types.Metadata{
			Name:        name,
			Description: "test genericK8sIntent",
			UserData1:   "some user data 1",
			UserData2:   "some user data 2",
		},
	}
}

func validategenericK8sIntentResponse(res *http.Response, t test) {
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

	result := t.result.(module.GenericK8sIntent)

	gki := module.GenericK8sIntent{}
	json.NewDecoder(bytes.NewReader(data)).Decode(&gki)
	Expect(gki).To(Equal(result))
}
