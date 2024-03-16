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

type mockCustomization struct {
	Customization module.Customization
	Content       module.CustomizationContent
}

// mockCustomizationManager implements the mock services for the CustomizationManager
type mockCustomizationManager struct {
	Items []mockCustomization
	Err   error
}

func init() {
	api.CustomizationSchemaJson = "../../json-schemas/customization.json"
}

func (m *mockCustomizationManager) CreateCustomization(ctx context.Context, customization module.Customization, content module.CustomizationContent,
	project, compositeApp, version, deploymentIntentGroup, intent, resource string,
	failIfExists bool) (module.Customization, bool, error) {

	if m.Err != nil {
		return module.Customization{}, false, m.Err
	}

	iExists := false
	index := 0

	for i, item := range m.Items {
		if item.Customization.Metadata.Name == customization.Metadata.Name {
			iExists = true
			index = i
			break
		}
	}

	if iExists && failIfExists { // customization already exists
		return module.Customization{}, iExists, emcoerror.NewEmcoError(
			module.CustomizationAlreadyExists,
			emcoerror.Conflict,
		)
	}

	if iExists && !failIfExists { // customization already exists. update the customization
		m.Items[index].Customization = customization
		return m.Items[index].Customization, iExists, nil
	}

	mc := mockCustomization{
		Customization: customization,
	}

	m.Items = append(m.Items, mc) // create the customization

	return m.Items[len(m.Items)-1].Customization, iExists, nil
}

func (m *mockCustomizationManager) DeleteCustomization(ctx context.Context,
	customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) error {

	if m.Err != nil {
		return m.Err
	}

	for k, item := range m.Items {
		if item.Customization.Metadata.Name == customization { // customization exist
			m.Items[k] = m.Items[len(m.Items)-1]
			m.Items[len(m.Items)-1].Customization = module.Customization{}
			m.Items = m.Items[:len(m.Items)-1]
			return nil
		}
	}

	return emcoerror.NewEmcoError(
		"the requested resource not found",
		emcoerror.NotFound,
	) // customization does not exist
}

func (m *mockCustomizationManager) GetAllCustomization(ctx context.Context,
	project, compositeApp, version, deploymentIntentGroup, intent, resource string) ([]module.Customization, error) {

	if m.Err != nil {
		return []module.Customization{}, m.Err
	}

	var customizations []module.Customization
	for _, item := range m.Items {
		c := item.Customization
		customizations = append(customizations, c)
	}

	return customizations, nil
}

func (m *mockCustomizationManager) GetCustomization(ctx context.Context,
	customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) (module.Customization, error) {

	if m.Err != nil {
		return module.Customization{}, m.Err
	}

	for _, item := range m.Items {
		if item.Customization.Metadata.Name == customization {
			return item.Customization, nil
		}
	}

	return module.Customization{}, emcoerror.NewEmcoError(
		module.CustomizationNotFound,
		emcoerror.NotFound,
	)
}

func (m *mockCustomizationManager) GetCustomizationContent(ctx context.Context,
	customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) (module.CustomizationContent, error) {

	if m.Err != nil {
		return module.CustomizationContent{}, m.Err
	}

	for _, item := range m.Items {
		if item.Customization.Metadata.Name == customization {
			return item.Content, nil
		}
	}

	return module.CustomizationContent{}, nil
}

var _ = Describe("Test create customization handlers",
	func() {
		DescribeTable("Create Customization",
			func(t test) {
				var buf bytes.Buffer
				ct, err := createMultiPartFormData(t.input, &buf)
				if err != nil {
					Fail(err.Error())
				}
				header := map[string]string{
					"Content-Type": ct,
				}
				client := t.client.(*mockCustomizationManager)
				res := executeRequest(http.MethodPost, "/test-gki/resources/test-resource/customizations", header, client, &buf)
				validateCustomizationResponse(res, t)
			},
			Entry("request body validation",
				test{
					entry:      "request body validation",
					input:      customizationInput(""), // create an empty customization payload
					result:     mockCustomization{},
					err:        errors.New("customization name may not be emptycluster specific may not be emptyscope may not be emptycluster provider may not be emptymode may not be empty"),
					statusCode: http.StatusBadRequest,
					client: &mockCustomizationManager{
						Err:   nil,
						Items: populateCustomizationTestData(),
					},
				},
			),
			Entry("successful create",
				test{
					entry:      "successful create",
					input:      customizationInput("testCustomization"),
					result:     customizationResult("testCustomization"),
					err:        nil,
					statusCode: http.StatusCreated,
					client: &mockCustomizationManager{
						Err:   nil,
						Items: populateCustomizationTestData(),
					},
				},
			),
			Entry("customization already exists",
				test{
					entry:  "customization already exists",
					input:  customizationInput("testCustomization-1"),
					result: mockCustomization{},
					err: emcoerror.NewEmcoError(
						module.CustomizationAlreadyExists,
						emcoerror.Conflict,
					),
					statusCode: http.StatusConflict,
					client: &mockCustomizationManager{
						Err:   nil,
						Items: populateCustomizationTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test get customization handlers",
	func() {
		DescribeTable("Get Customization",
			func(t test) {
				client := t.client.(*mockCustomizationManager)
				header := map[string]string{
					"Accept": "application/json",
				}
				res := executeRequest(http.MethodGet, "/test-gki/resources/test-resource/customizations/"+t.name, header, client, nil)
				validateCustomizationResponse(res, t)
			},
			Entry("successful get",
				test{
					entry:      "successful get",
					name:       "testCustomization-1",
					statusCode: http.StatusOK,
					err:        nil,
					result:     customizationResult("testCustomization-1"),
					client: &mockCustomizationManager{
						Err:   nil,
						Items: populateCustomizationTestData(),
					},
				},
			),
			Entry("customization not found",
				test{
					entry:      "customization not found",
					name:       "nonExistingCustomization",
					statusCode: http.StatusNotFound,
					err: emcoerror.NewEmcoError(
						module.CustomizationNotFound,
						emcoerror.NotFound,
					),
					result: mockCustomization{},
					client: &mockCustomizationManager{
						Err:   nil,
						Items: populateCustomizationTestData(),
					},
				},
			),
		)
		DescribeTable("Get Multipart Customization",
			func(t test) {
				client := t.client.(*mockCustomizationManager)
				header := map[string]string{
					"Accept": "multipart/form-data",
				}
				res := executeRequest(http.MethodGet, "/test-gki/resources/test-resource/customizations/"+t.name, header, client, nil)
				validateCustomizationResponse(res, t)
			},
			Entry("successful get multipart/form-data",
				test{
					entry:      "successful get multipart/form-data",
					name:       "testCustomization-3",
					statusCode: http.StatusOK,
					err:        nil,
					result:     customizationWithFileContent("testCustomization-3"),
					client: &mockCustomizationManager{
						Err:   nil,
						Items: populateCustomizationTestData(),
					},
				},
			),
		)
		DescribeTable("Get OctetStream Customization",
			func(t test) {
				client := t.client.(*mockCustomizationManager)
				header := map[string]string{
					"Accept": "application/octet-stream",
				}
				res := executeRequest(http.MethodGet, "/test-gki/resources/test-resource/customizations/"+t.name, header, client, nil)
				validateCustomizationResponse(res, t)
			},
			Entry("successful get application/octet-stream",
				test{
					entry:      "successful get application/octet-stream",
					name:       "testCustomization-3",
					statusCode: http.StatusOK,
					err:        nil,
					result:     customizationWithFileContent("testCustomization-3"),
					client: &mockCustomizationManager{
						Err:   nil,
						Items: populateCustomizationTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test update customization handlers",
	func() {
		DescribeTable("Update Customization",
			func(t test) {
				var buf bytes.Buffer
				ct, err := createMultiPartFormData(t.input, &buf)
				if err != nil {
					Fail(err.Error())
				}
				header := map[string]string{
					"Content-Type": ct,
				}
				client := t.client.(*mockCustomizationManager)
				res := executeRequest(http.MethodPut, "/test-gki/resources/test-resource/customizations/"+t.name, header, client, &buf)
				validateCustomizationResponse(res, t)
			},
			Entry("request body validation",
				test{
					entry:      "request body validation",
					name:       "testCustomization",
					input:      customizationInput(""), // create an empty customization payload
					result:     mockCustomization{},
					err:        errors.New("customization name may not be emptycluster specific may not be emptyscope may not be emptycluster provider may not be emptymode may not be empty"),
					statusCode: http.StatusBadRequest,
					client: &mockCustomizationManager{
						Err:   nil,
						Items: populateCustomizationTestData(),
					},
				},
			),
			Entry("successful update",
				test{
					entry:      "successful update",
					name:       "testCustomization",
					input:      customizationInput("testCustomization"),
					result:     customizationResult("testCustomization"),
					err:        nil,
					statusCode: http.StatusCreated,
					client: &mockCustomizationManager{
						Err:   nil,
						Items: populateCustomizationTestData(),
					},
				},
			),
			Entry("customization already exists",
				test{
					entry:      "customization already exists",
					name:       "testCustomization-4",
					input:      customizationInput("testCustomization-4"),
					result:     customizationResult("testCustomization-4"),
					err:        nil,
					statusCode: http.StatusOK,
					client: &mockCustomizationManager{
						Err:   nil,
						Items: populateCustomizationTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test delete customization handlers",
	func() {
		DescribeTable("Delete Customization",
			func(t test) {

				client := t.client.(*mockCustomizationManager)
				res := executeRequest(http.MethodDelete, "/test-gki/resources/test-resource/customizations/"+t.name, nil, client, nil)
				validateCustomizationResponse(res, t)
			},
			Entry("successful delete",
				test{
					entry:      "successful delete",
					name:       "testCustomization-1",
					statusCode: http.StatusNoContent,
					err:        nil,
					result:     mockCustomization{},
					client: &mockCustomizationManager{
						Err:   nil,
						Items: populateCustomizationTestData(),
					},
				},
			),
			Entry("db remove customization not found",
				test{
					entry:      "db remove customization not found",
					name:       "nonExistingCustomization",
					statusCode: http.StatusNotFound,
					err: emcoerror.NewEmcoError(
						"the requested resource not found",
						emcoerror.NotFound,
					),
					result: mockCustomization{},
					client: &mockCustomizationManager{
						Err:   nil,
						Items: populateCustomizationTestData(),
					},
				},
			),
		)
	},
)

func populateCustomizationTestData() []mockCustomization {
	return []mockCustomization{
		{
			Customization: module.Customization{
				Metadata: types.Metadata{
					Name:        "testCustomization-1",
					Description: "test customization",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.CustomizationSpec{
					ClusterSpecific: "true",
					ClusterInfo: module.ClusterInfo{
						Scope:           "label",
						ClusterProvider: "provider-1",
						ClusterName:     "cluster-1",
						ClusterLabel:    "edge-cluster-1",
						Mode:            "allow",
					},
					PatchType: "json",
					PatchJSON: []map[string]interface{}{
						{
							"op":    "replace",
							"path":  "/spec/replicas",
							"value": "6",
						},
					},
					ConfigMapOptions: module.ConfigMapOptions{
						DataKeyOptions: []module.KeyOptions{
							{
								FileName: "info.json",
								KeyName:  "info",
							},
						},
					},
				},
			},
		},
		{
			Customization: module.Customization{
				Metadata: types.Metadata{
					Name:        "testCustomization-2",
					Description: "test customization",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.CustomizationSpec{
					ClusterSpecific: "true",
					ClusterInfo: module.ClusterInfo{
						Scope:           "label",
						ClusterProvider: "provider-1",
						ClusterName:     "cluster-1",
						ClusterLabel:    "edge-cluster-1",
						Mode:            "allow",
					},
					PatchType: "json",
					PatchJSON: []map[string]interface{}{
						{
							"op":    "replace",
							"path":  "/spec/replicas",
							"value": "6",
						},
					},
					ConfigMapOptions: module.ConfigMapOptions{
						DataKeyOptions: []module.KeyOptions{
							{
								FileName: "info.json",
								KeyName:  "info",
							},
							{
								FileName: "data.json",
								KeyName:  "data",
							},
						},
					},
				},
			},
		},
		{
			Customization: module.Customization{
				Metadata: types.Metadata{
					Name:        "testCustomization-3",
					Description: "test customization",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.CustomizationSpec{
					ClusterSpecific: "true",
					ClusterInfo: module.ClusterInfo{
						Scope:           "label",
						ClusterProvider: "provider-1",
						ClusterName:     "cluster-1",
						ClusterLabel:    "edge-cluster-1",
						Mode:            "allow",
					},
					PatchType: "json",
					PatchJSON: []map[string]interface{}{
						{
							"op":    "replace",
							"path":  "/spec/replicas",
							"value": "6",
						},
					},
					ConfigMapOptions: module.ConfigMapOptions{
						DataKeyOptions: []module.KeyOptions{
							{
								FileName: "info.json",
								KeyName:  "info",
							},
							{
								FileName: "data.json",
								KeyName:  "data",
							},
						},
					},
				},
			},
			Content: module.CustomizationContent{
				Content: []module.Content{
					{
						FileName: "info.json",
						Content:  "ewogICAgImZydWl0IjogImJlcnJ5IiwKICAgICJhbmltYWwiOiAiZG9nIiwKICAgICJjb2xvciI6ICJyZWQiCn0K",
						KeyName:  "info",
					},
					{
						FileName: "data.json",
						Content:  "ewogICAgIjEiOiAiYmVycnkiLAogICAgIjIiOiAiZG9nIiwKICAgICIzIjogInJlZCIKfQo=",
						KeyName:  "data",
					},
				},
			},
		},
		{
			Customization: module.Customization{
				Metadata: types.Metadata{
					Name:        "testCustomization-4",
					Description: "",
					UserData1:   "",
					UserData2:   "",
				},
				Spec: module.CustomizationSpec{
					ClusterSpecific: "true",
					ClusterInfo: module.ClusterInfo{
						Scope:           "label",
						ClusterProvider: "provider-1",
						ClusterName:     "cluster-1",
						ClusterLabel:    "edge-cluster-1",
						Mode:            "allow",
					},
					PatchType: "json",
					PatchJSON: []map[string]interface{}{
						{
							"op":    "replace",
							"path":  "/spec/replicas",
							"value": "6",
						},
					},
					ConfigMapOptions: module.ConfigMapOptions{
						DataKeyOptions: []module.KeyOptions{
							{
								FileName: "info.json",
								KeyName:  "info",
							},
							{
								FileName: "data.json",
								KeyName:  "data",
							},
						},
					},
				},
			},
			Content: module.CustomizationContent{},
		},
	}
}

func customizationInput(name string) io.Reader {
	if len(name) == 0 {
		return bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": ""
			},
			"spec": {
				"clusterSpecific": "",
				"clusterInfo": {
					"scope": "",
					"clusterProvider": "",
					"mode": ""
				}
			}
		}`))
	}

	return bytes.NewBuffer([]byte(`{
		"metadata": {
			"name": "` + name + `",
			"description": "test customization",
			"userData1": "some user data 1",
			"userData2": "some user data 2"
		},
		"spec": {
			"clusterSpecific": "true",
			"clusterInfo": {
				"scope": "label",
				"clusterProvider": "provider-1",
				"cluster": "cluster-1",
				"clusterLabel": "edge-cluster-1",
				"mode": "allow"
			},
			"patchType": "json",
			"patchJson": [
				{
					"op": "replace",
					"path": "/spec/replicas",
					"value": "6"
				}
			],
			"configMapOptions": {
				"dataKeyOptions": [
					{
						"fileName": "info.json",
						"keyName" : "info"
					}
				]
			}
		}
	}`))
}

func customizationResult(name string) mockCustomization {
	return mockCustomization{
		Customization: module.Customization{
			Metadata: types.Metadata{
				Name:        name,
				Description: "test customization",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: module.CustomizationSpec{
				ClusterSpecific: "true",
				ClusterInfo: module.ClusterInfo{
					Scope:           "label",
					ClusterProvider: "provider-1",
					ClusterName:     "cluster-1",
					ClusterLabel:    "edge-cluster-1",
					Mode:            "allow",
				},
				PatchType: "json",
				PatchJSON: []map[string]interface{}{
					{
						"op":    "replace",
						"path":  "/spec/replicas",
						"value": "6",
					},
				},
				ConfigMapOptions: module.ConfigMapOptions{
					DataKeyOptions: []module.KeyOptions{
						{
							FileName: "info.json",
							KeyName:  "info",
						},
					},
				},
			},
		},
		Content: module.CustomizationContent{},
	}
}

func customizationWithFileContent(name string) mockCustomization {
	return mockCustomization{
		Customization: module.Customization{
			Metadata: types.Metadata{
				Name:        name,
				Description: "test customization",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: module.CustomizationSpec{
				ClusterSpecific: "true",
				ClusterInfo: module.ClusterInfo{
					Scope:           "label",
					ClusterProvider: "provider-1",
					ClusterName:     "cluster-1",
					ClusterLabel:    "edge-cluster-1",
					Mode:            "allow",
				},
				PatchType: "json",
				PatchJSON: []map[string]interface{}{
					{
						"op":    "replace",
						"path":  "/spec/replicas",
						"value": "6",
					},
				},
				ConfigMapOptions: module.ConfigMapOptions{
					DataKeyOptions: []module.KeyOptions{
						{
							FileName: "info.json",
							KeyName:  "info",
						},
					},
				},
			},
		},
		Content: module.CustomizationContent{
			Content: []module.Content{
				{
					FileName: "info.json",
					KeyName:  "info",
					Content:  `{ "fruit": "berry", "animal": "dog", "color": "red" }`,
				},
				{
					FileName: "data.json",
					KeyName:  "data",
					Content:  `{ "1": "berry", "2": "dog", "3": "red" }`,
				},
			},
		},
	}
}

func validateCustomizationResponse(res *http.Response, t test) {
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body) // to retain the content
	if err != nil {
		Fail(err.Error())
	}

	Expect(res.StatusCode).To(Equal(t.statusCode))

	if t.err != nil {
		b := strings.Replace(string(data), "\n", "", -1) // replace the new line at the end
		Expect(b).To(Equal(t.err.Error()))
		return
	}

	result := t.result.(mockCustomization)

	if len(res.Header) > 0 {
		if strings.Contains(res.Header.Get("Content-Type"), "multipart/form-data") {
			// validate response body
			b := string(data)
			Expect(b).To(ContainSubstring(result.Customization.Metadata.Name))
			return
		}

		switch res.Header.Get("Content-Type") {
		case "application/json":
			c := module.Customization{}
			json.NewDecoder(bytes.NewReader(data)).Decode(&c)
			Expect(c).To(Equal(result.Customization))

		case "application/octet-stream":
			b := string(data)
			// validate response body
			Expect(b).NotTo(ContainSubstring(result.Customization.Metadata.Name))

		default:
			Fail("Invalid response!")
		}
	}
}
