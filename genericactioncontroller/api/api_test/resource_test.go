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

type mockResourceManager struct {
	Items []mockResource
	Err   error
}

type mockResource struct {
	Resource module.Resource
	Content  module.ResourceContent
}

func init() {
	api.ResourceSchemaJson = "../../json-schemas/resource.json"
}

func (m *mockResourceManager) CreateResource(ctx context.Context, res module.Resource, resContent module.ResourceContent,
	project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string,
	failIfExists bool) (module.Resource, bool, error) {
	iExists := false
	index := 0

	if m.Err != nil {
		return module.Resource{}, iExists, m.Err
	}

	for i, item := range m.Items {
		if item.Resource.Metadata.Name == res.Metadata.Name {
			iExists = true
			index = i
			break
		}
	}

	if iExists && failIfExists { // resource already exists
		return module.Resource{}, iExists, emcoerror.NewEmcoError(
			module.ResourceAlreadyExists,
			emcoerror.Conflict,
		)
	}

	if iExists && !failIfExists { // resource already exists. update the resource
		m.Items[index].Resource = res
		return m.Items[index].Resource, iExists, nil
	}
	mr := mockResource{
		Resource: res,
	}

	m.Items = append(m.Items, mr) // create the resource

	return m.Items[len(m.Items)-1].Resource, iExists, nil

}

func (m *mockResourceManager) GetResource(ctx context.Context, resource, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string) (module.Resource, error) {

	if m.Err != nil {
		return module.Resource{}, m.Err
	}

	for _, item := range m.Items {
		if item.Resource.Metadata.Name == resource {
			return item.Resource, nil
		}
	}

	return module.Resource{}, emcoerror.NewEmcoError(
		module.ResourceNotFound,
		emcoerror.NotFound,
	)
}

func (m *mockResourceManager) GetResourceContent(ctx context.Context, resource, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string) (module.ResourceContent, error) {

	if m.Err != nil {
		return module.ResourceContent{}, m.Err
	}

	for _, item := range m.Items {
		if item.Resource.Metadata.Name == resource {
			return item.Content, nil
		}
	}

	return module.ResourceContent{}, nil
}

func (m *mockResourceManager) GetAllResources(ctx context.Context, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string) ([]module.Resource, error) {
	if m.Err != nil {
		return []module.Resource{}, m.Err
	}

	var resources []module.Resource
	for _, item := range m.Items {
		c := item.Resource
		resources = append(resources, c)
	}

	return resources, nil
}

func (m *mockResourceManager) DeleteResource(ctx context.Context, resource, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string) error {
	if m.Err != nil {
		return m.Err
	}

	for k, item := range m.Items {
		if item.Resource.Metadata.Name == resource { // resource exist
			m.Items[k] = m.Items[len(m.Items)-1]
			m.Items[len(m.Items)-1].Resource = module.Resource{}
			m.Items = m.Items[:len(m.Items)-1]
			return nil
		}
	}

	return emcoerror.NewEmcoError(
		"the requested resource not found",
		emcoerror.NotFound,
	) // resource does not exist

}

var _ = Describe("Test create resource handler",
	func() {
		DescribeTable("Create Resource",
			func(t test) {
				var buf bytes.Buffer
				ct, err := createMultiPartFormData(t.input, &buf)
				if err != nil {
					Fail(err.Error())
				}
				header := map[string]string{
					"Content-Type": ct,
				}
				client := t.client.(*mockResourceManager)
				res := executeRequest(http.MethodPost, "/test-gki/resources", header, client, &buf)
				validateResourceResponse(res, t)
			},
			Entry("request body validation",
				test{
					input:      resourceInput(""), // create an empty resource payload
					result:     mockResource{},
					err:        errors.New("resource name may not be emptyapp may not be emptynewObject may not be emptyapiVersion may not be emptykind may not be emptyname may not be empty"),
					statusCode: http.StatusBadRequest,
					client: &mockResourceManager{
						Err:   nil,
						Items: populateResourceTestData(),
					},
				},
			),
			Entry("successful create",
				test{
					input:      resourceInput("testResource"),
					result:     resourceResult("testResource"),
					err:        nil,
					statusCode: http.StatusCreated,
					client: &mockResourceManager{
						Err:   nil,
						Items: populateResourceTestData(),
					},
				},
			),
			Entry("resource already exists",
				test{
					input:  resourceInput("testResource-1"),
					result: mockResource{},
					err: emcoerror.NewEmcoError(
						module.ResourceAlreadyExists,
						emcoerror.Conflict,
					),
					statusCode: http.StatusConflict,
					client: &mockResourceManager{
						Err:   nil,
						Items: populateResourceTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test get resource handler",
	func() {
		DescribeTable("Get Resource",
			func(t test) {
				client := t.client.(*mockResourceManager)
				header := map[string]string{
					"Accept": "application/json",
				}
				res := executeRequest(http.MethodGet, "/test-gki/resources/"+t.name, header, client, nil)
				validateResourceResponse(res, t)
			},
			Entry("successful get",
				test{
					name:       "testResource-1",
					statusCode: http.StatusOK,
					err:        nil,
					result:     resourceResult("testResource-1"),
					client: &mockResourceManager{
						Err:   nil,
						Items: populateResourceTestData(),
					},
				},
			),
			Entry("resource not found",
				test{
					name:       "nonExistingResource",
					statusCode: http.StatusNotFound,
					err: emcoerror.NewEmcoError(
						module.ResourceNotFound,
						emcoerror.NotFound,
					),
					result: mockResource{},
					client: &mockResourceManager{
						Err:   nil,
						Items: populateResourceTestData(),
					},
				},
			),
		)
		DescribeTable("Get Multipart Resource",
			func(t test) {
				client := t.client.(*mockResourceManager)
				header := map[string]string{
					"Accept": "multipart/form-data",
				}
				res := executeRequest(http.MethodGet, "/test-gki/resources/"+t.name, header, client, nil)
				validateResourceResponse(res, t)
			},
			Entry("successful get multipart/form-data",
				test{
					name:       "testResource-3",
					statusCode: http.StatusOK,
					err:        nil,
					result:     resourceWithFileContent("testResource-3"),
					client: &mockResourceManager{
						Err:   nil,
						Items: populateResourceTestData(),
					},
				},
			),
		)
		DescribeTable("Get OctetStream Resource",
			func(t test) {
				client := t.client.(*mockResourceManager)
				header := map[string]string{
					"Accept": "application/octet-stream",
				}
				res := executeRequest(http.MethodGet, "/test-gki/resources/"+t.name, header, client, nil)
				validateResourceResponse(res, t)
			},
			Entry("successful get application/octet-stream",
				test{
					name:       "testResource-3",
					statusCode: http.StatusOK,
					err:        nil,
					result:     resourceWithFileContent("testResource-3"),
					client: &mockResourceManager{
						Err:   nil,
						Items: populateResourceTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test update resource handler",
	func() {
		DescribeTable("Update Resource",
			func(t test) {
				var buf bytes.Buffer
				ct, err := createMultiPartFormData(t.input, &buf)
				if err != nil {
					Fail(err.Error())
				}
				header := map[string]string{
					"Content-Type": ct,
				}
				client := t.client.(*mockResourceManager)
				res := executeRequest(http.MethodPut, "/test-gki/resources/"+t.name, header, client, &buf)
				validateResourceResponse(res, t)
			},
			Entry("request body validation",
				test{
					name:       "testResource",
					input:      resourceInput(""), // create an empty resource payload
					result:     mockResource{},
					err:        errors.New("resource name may not be emptyapp may not be emptynewObject may not be emptyapiVersion may not be emptykind may not be emptyname may not be empty"),
					statusCode: http.StatusBadRequest,
					client: &mockResourceManager{
						Err:   nil,
						Items: populateResourceTestData(),
					},
				},
			),
			Entry("successful update",
				test{
					name:       "testResource",
					input:      resourceInput("testResource"),
					result:     resourceResult("testResource"),
					err:        nil,
					statusCode: http.StatusCreated,
					client: &mockResourceManager{
						Err:   nil,
						Items: populateResourceTestData(),
					},
				},
			),
			Entry("resource already exists",
				test{
					name:       "testResource-4",
					input:      resourceInput("testResource-4"),
					result:     resourceResult("testResource-4"),
					err:        nil,
					statusCode: http.StatusOK,
					client: &mockResourceManager{
						Err:   nil,
						Items: populateResourceTestData(),
					},
				},
			),
		)
	},
)

var _ = Describe("Test delete resource handler",
	func() {
		DescribeTable("Delete Resource",
			func(t test) {
				client := t.client.(*mockResourceManager)
				res := executeRequest(http.MethodDelete, "/test-gki/resources/"+t.name, nil, client, nil)
				validateResourceResponse(res, t)
			},
			Entry("successful delete",
				test{
					name:       "testResource-1",
					statusCode: http.StatusNoContent,
					err:        nil,
					result:     mockResource{},
					client: &mockResourceManager{
						Err:   nil,
						Items: populateResourceTestData(),
					},
				},
			),
			Entry("db remove resource not found",
				test{
					entry:      "db remove resource not found",
					name:       "nonExistingResource",
					statusCode: http.StatusNotFound,
					err: emcoerror.NewEmcoError(
						"the requested resource not found",
						emcoerror.NotFound,
					),
					result: mockResource{},
					client: &mockResourceManager{
						Err:   nil,
						Items: populateResourceTestData(),
					},
				},
			),
		)
	},
)

func populateResourceTestData() []mockResource {
	return []mockResource{
		{
			Resource: module.Resource{
				Metadata: types.Metadata{
					Name:        "testResource-1",
					Description: "test resource",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.ResourceSpec{
					AppName:   "operator",
					NewObject: "true",
					ResourceGVK: module.ResourceGVK{
						APIVersion: "v1",
						Kind:       "ConfigMap",
						Name:       "my-cm",
					},
				},
			},
		},
		{
			Resource: module.Resource{
				Metadata: types.Metadata{
					Name:        "testResource-2",
					Description: "test resource",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.ResourceSpec{
					AppName:   "operator",
					NewObject: "true",
					ResourceGVK: module.ResourceGVK{
						APIVersion: "v1",
						Kind:       "Secret",
						Name:       "my-secret",
					},
				},
			},
		},
		{
			Resource: module.Resource{
				Metadata: types.Metadata{
					Name:        "testResource-3",
					Description: "test resource",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.ResourceSpec{
					AppName:   "operator",
					NewObject: "true",
					ResourceGVK: module.ResourceGVK{
						APIVersion: "v1",
						Kind:       "ConfigMap",
						Name:       "my-cm",
					},
				},
			},
			Content: module.ResourceContent{
				Content: "YXBpVmVyc2lvbjogdjEKa2luZDogQ29uZmlnTWFwCm1ldGFkYXRhOgogIG5hbWU6IG15LWNtCmRhdGE6CiAgIyBwcm9wZXJ0eS1saWtlIGtleXM7IGVhY2gga2V5IG1hcHMgdG8gYSBzaW1wbGUgdmFsdWUKICBwbGF5ZXJfaW5pdGlhbF9saXZlczogIjMiCiAgdWlfcHJvcGVydGllc19maWxlX25hbWU6ICJ1c2VyLWludGVyZmFjZS5wcm9wZXJ0aWVzIgoKICAjIGZpbGUtbGlrZSBrZXlzCiAgZ2FtZS5wcm9wZXJ0aWVzOiB8CiAgICBlbmVteS50eXBlcz1hbGllbnMsbW9uc3RlcnMKICAgIHBsYXllci5tYXhpbXVtLWxpdmVzPTUgICAgCiAgdXNlci1pbnRlcmZhY2UucHJvcGVydGllczogfAogICAgY29sb3IuZ29vZD1wdXJwbGUKICAgIGNvbG9yLmJhZD15ZWxsb3cKICAgIGFsbG93LnRleHRtb2RlPXRydWUKCg==",
			},
		},
		{
			Resource: module.Resource{
				Metadata: types.Metadata{
					Name:        "testResource-4",
					Description: "",
					UserData1:   "",
					UserData2:   "",
				},
				Spec: module.ResourceSpec{
					AppName:   "operator",
					NewObject: "true",
					ResourceGVK: module.ResourceGVK{
						APIVersion: "v1",
						Kind:       "ConfigMap",
						Name:       "my-cm",
					},
				},
			},
			Content: module.ResourceContent{},
		},
	}
}

func resourceInput(name string) io.Reader {
	if len(name) == 0 {
		return bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": ""
			},
			"spec": {
				"app": "",
				"newObject": "",
				"resourceGVK": {
					"apiVersion": "",
					"kind": "",
					"name": ""
				}
			}
		}`))
	}

	return bytes.NewBuffer([]byte(`{
		"metadata": {
			"name": "` + name + `",
			"description": "test resource",
			"userData1": "some user data 1",
			"userData2": "some user data 2"
		},
		"spec": {
			"app": "operator",
				"newObject": "true",
				"resourceGVK": {
					"apiVersion": "v1",
					"kind": "ConfigMap",
					"name": "my-cm"
				}
		}
	}`))
}

func resourceResult(name string) mockResource {
	return mockResource{
		Resource: module.Resource{
			Metadata: types.Metadata{
				Name:        name,
				Description: "test resource",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: module.ResourceSpec{
				AppName:   "operator",
				NewObject: "true",
				ResourceGVK: module.ResourceGVK{
					APIVersion: "v1",
					Kind:       "ConfigMap",
					Name:       "my-cm",
				},
			},
		},
		Content: module.ResourceContent{},
	}
}

func resourceWithFileContent(name string) mockResource {
	return mockResource{
		Resource: module.Resource{
			Metadata: types.Metadata{
				Name:        name,
				Description: "test resource",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: module.ResourceSpec{
				AppName:   "operator",
				NewObject: "true",
				ResourceGVK: module.ResourceGVK{
					APIVersion: "v1",
					Kind:       "ConfigMap",
					Name:       "my-cm",
				},
			},
		},
		Content: module.ResourceContent{
			Content: `apiVersion: v1
			kind: ConfigMap
			metadata:
			  name: my-cm
			data:
			  # property-like keys; each key maps to a simple value
			  player_initial_lives: "3"
			  ui_properties_file_name: "user-interface.properties"

			  # file-like keys
			  game.properties: |
				enemy.types=aliens,monsters
				player.maximum-lives=5
			  user-interface.properties: |
				color.good=purple
				color.bad=yellow
				allow.textmode=true`,
		},
	}
}

func validateResourceResponse(res *http.Response, t test) {
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

	result := t.result.(mockResource)

	if len(res.Header) > 0 {
		if strings.Contains(res.Header.Get("Content-Type"), "multipart/form-data") {
			// validate response body
			b := string(data)
			Expect(b).To(ContainSubstring(result.Resource.Metadata.Name))
			return
		}

		switch res.Header.Get("Content-Type") {
		case "application/json":
			r := module.Resource{}
			json.NewDecoder(bytes.NewReader(data)).Decode(&r)
			Expect(r).To(Equal(result.Resource))

		case "application/octet-stream":
			b := string(data)
			// validate response body
			Expect(b).NotTo(ContainSubstring(result.Resource.Metadata.Name))

		default:
			Fail("Invalid response!")
		}
	}
}
