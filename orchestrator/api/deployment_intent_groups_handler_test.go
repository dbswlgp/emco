// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	pkgerrors "github.com/pkg/errors"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
)

// mockDeploymentIntentGroupManager allows us to mock the DeploymentIntentGroupManager functionalities and the database connections
type mockDeploymentIntentGroupManager struct {
	Err       error
	Items     []moduleLib.DeploymentIntentGroup
	StateInfo state.StateInfo
}

func (digm *mockDeploymentIntentGroupManager) GetDeploymentIntentGroup(ctx context.Context, deploymentIntentGroup, project, compositeApp, version string) (moduleLib.DeploymentIntentGroup, error) {
	if digm.Err != nil {
		return moduleLib.DeploymentIntentGroup{}, digm.Err
	}

	for _, item := range digm.Items {
		if item.MetaData.Name == deploymentIntentGroup {
			return item, nil
		}
	}

	return moduleLib.DeploymentIntentGroup{}, pkgerrors.New("DeploymentIntentGroup not found")
}

func (digm *mockDeploymentIntentGroupManager) GetAllDeploymentIntentGroups(ctx context.Context, project, compositeApp, version string) ([]moduleLib.DeploymentIntentGroup, error) {
	if digm.Err != nil {
		return []moduleLib.DeploymentIntentGroup{}, digm.Err
	}

	if len(digm.Items) > 0 {
		return digm.Items, nil
	}

	return []moduleLib.DeploymentIntentGroup{}, nil
}

func (digm *mockDeploymentIntentGroupManager) CreateDeploymentIntentGroup(ctx context.Context, d moduleLib.DeploymentIntentGroup, project, compositeApp, version string, failIfExists bool) (moduleLib.DeploymentIntentGroup, bool, error) {
	digExists := false
	index := 0

	if digm.Err != nil {
		return moduleLib.DeploymentIntentGroup{}, digExists, digm.Err
	}

	for i, item := range digm.Items {
		if item.MetaData.Name == d.MetaData.Name {
			digExists = true
			index = i
			break
		}
	}

	if digExists && failIfExists { // resource already exists
		return moduleLib.DeploymentIntentGroup{}, digExists, pkgerrors.New("DeploymentIntent already exists")
	}

	if digExists && !failIfExists { // resource already exists. update the resource
		digm.Items[index] = d
		return digm.Items[index], digExists, nil
	}

	digm.Items = append(digm.Items, d) // create the resource

	return digm.Items[len(digm.Items)-1], digExists, nil

}

func (digm *mockDeploymentIntentGroupManager) DeleteDeploymentIntentGroup(ctx context.Context, deploymentIntentGroup, project, compositeApp, version string) error {
	if digm.Err != nil {
		return digm.Err
	}

	for i, item := range digm.Items {
		if item.MetaData.Name == deploymentIntentGroup { // resource exist
			digm.Items[i] = digm.Items[len(digm.Items)-1]
			digm.Items[len(digm.Items)-1] = moduleLib.DeploymentIntentGroup{}
			digm.Items = digm.Items[:len(digm.Items)-1]
			return nil
		}
	}

	return pkgerrors.New("db Remove resource not found") // resource does not exist
}

func (digm *mockDeploymentIntentGroupManager) GetDeploymentIntentGroupState(ctx context.Context, deploymentIntentGroup, project, compositeApp, version string) (state.StateInfo, error) {
	if digm.Err != nil {
		return state.StateInfo{}, digm.Err
	}

	for _, item := range digm.Items {
		if item.MetaData.Name == deploymentIntentGroup { // resource exist

			return state.StateInfo{}, nil
		}
	}

	return state.StateInfo{}, pkgerrors.New("DeploymentIntentGroup StateInfo not foundd") // resource does not exist
}

func init() {
	dpiJSONFile = "../json-schemas/deployment-group-intent.json"
}

func TestGetDeploymentIntentGroupHandler(t *testing.T) {
	testCases := []struct {
		err, label, name string
		client           *mockDeploymentIntentGroupManager
		code             int
		result           moduleLib.DeploymentIntentGroup
	}{
		{
			label: "Get DeploymentIntentGroup",
			code:  http.StatusOK,
			result: moduleLib.DeploymentIntentGroup{
				MetaData: moduleLib.DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "Test DeploymentIntentGroup used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.DepSpecData{
					Profile: "testCompositeProfile",
					Version: "v1",
					OverrideValuesObj: []moduleLib.OverrideValues{
						{
							AppName: "testApp",
							ValuesObj: map[string]string{
								"": "",
							},
						},
					},
				},
			},
			name: "testDeploymentIntentGroup",
			client: &mockDeploymentIntentGroupManager{
				Items: []moduleLib.DeploymentIntentGroup{
					{
						MetaData: moduleLib.DepMetaData{
							Name:        "testDeploymentIntentGroup_1",
							Description: "Test DeploymentIntentGroup_1 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.DepSpecData{
							Profile: "testCompositeProfile_1",
							Version: "v1",
							OverrideValuesObj: []moduleLib.OverrideValues{
								{
									AppName: "testApp_1",
									ValuesObj: map[string]string{
										"": "",
									},
								},
							},
						},
					},
					{
						MetaData: moduleLib.DepMetaData{
							Name:        "testDeploymentIntentGroup",
							Description: "Test DeploymentIntentGroup used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.DepSpecData{
							Profile: "testCompositeProfile",
							Version: "v1",
							OverrideValuesObj: []moduleLib.OverrideValues{
								{
									AppName: "testApp",
									ValuesObj: map[string]string{
										"": "",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			label: "DeploymentIntentGroup Not Found",
			code:  http.StatusNotFound,
			err:   "DeploymentIntentGroup not found",
			name:  "nonExistingDeploymentIntentGroup",
			client: &mockDeploymentIntentGroupManager{
				Items: []moduleLib.DeploymentIntentGroup{
					{
						MetaData: moduleLib.DepMetaData{
							Name:        "testDeploymentIntentGroup_1",
							Description: "Test DeploymentIntentGroup_1 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.DepSpecData{
							Profile: "testCompositeProfile_1",
							Version: "v1",
							OverrideValuesObj: []moduleLib.OverrideValues{
								{
									AppName: "testApp_1",
									ValuesObj: map[string]string{
										"": "",
									},
								},
							},
						},
					},
					{
						MetaData: moduleLib.DepMetaData{
							Name:        "testDeploymentIntentGroup",
							Description: "Test DeploymentIntentGroup used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.DepSpecData{
							Profile: "testCompositeProfile",
							Version: "v1",
							OverrideValuesObj: []moduleLib.OverrideValues{
								{
									AppName: "testApp",
									ValuesObj: map[string]string{
										"": "",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/"+test.name, nil)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("getDeploymentIntentGroupHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if resp.Code == http.StatusOK {
				dig := moduleLib.DeploymentIntentGroup{}
				json.NewDecoder(resp.Body).Decode(&dig)
				if reflect.DeepEqual(test.result, dig) == false {
					t.Fatalf("getDeploymentIntentGroupHandler returned an unexpected body. Expected %v; Got: %v", test.result, dig)
				}
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("getDeploymentIntentGroupHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}

func TestGetAllDeploymentIntentGroupsHandler(t *testing.T) {
	testCases := []struct {
		err, label string
		client     *mockDeploymentIntentGroupManager
		code       int
		result     []moduleLib.DeploymentIntentGroup
	}{
		{
			label: "Get All DeploymentIntentGroups",
			code:  http.StatusOK,
			result: []moduleLib.DeploymentIntentGroup{
				{
					MetaData: moduleLib.DepMetaData{
						Name:        "testDeploymentIntentGroup_1",
						Description: "Test DeploymentIntentGroup_1 used for unit testing",
						UserData1:   "data1",
						UserData2:   "data2",
					},
					Spec: moduleLib.DepSpecData{
						Profile: "testCompositeProfile_1",
						Version: "v1",
						OverrideValuesObj: []moduleLib.OverrideValues{
							{
								AppName: "testApp_1",
								ValuesObj: map[string]string{
									"": "",
								},
							},
						},
					},
				},
				{
					MetaData: moduleLib.DepMetaData{
						Name:        "testDeploymentIntentGroup_2",
						Description: "Test DeploymentIntentGroup_2 used for unit testing",
						UserData1:   "data1",
						UserData2:   "data2",
					},
					Spec: moduleLib.DepSpecData{
						Profile: "testCompositeProfile_2",
						Version: "v1",
						OverrideValuesObj: []moduleLib.OverrideValues{
							{
								AppName: "testApp_2",
								ValuesObj: map[string]string{
									"": "",
								},
							},
						},
					},
				},
			},
			client: &mockDeploymentIntentGroupManager{
				Items: []moduleLib.DeploymentIntentGroup{
					{
						MetaData: moduleLib.DepMetaData{
							Name:        "testDeploymentIntentGroup_1",
							Description: "Test DeploymentIntentGroup_1 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.DepSpecData{
							Profile: "testCompositeProfile_1",
							Version: "v1",
							OverrideValuesObj: []moduleLib.OverrideValues{
								{
									AppName: "testApp_1",
									ValuesObj: map[string]string{
										"": "",
									},
								},
							},
						},
					},
					{
						MetaData: moduleLib.DepMetaData{
							Name:        "testDeploymentIntentGroup_2",
							Description: "Test DeploymentIntentGroup_2 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.DepSpecData{
							Profile: "testCompositeProfile_2",
							Version: "v1",
							OverrideValuesObj: []moduleLib.OverrideValues{
								{
									AppName: "testApp_2",
									ValuesObj: map[string]string{
										"": "",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			label:  "Get All DeploymentIntentGroups Not Exists",
			code:   http.StatusOK,
			result: []moduleLib.DeploymentIntentGroup{},
			client: &mockDeploymentIntentGroupManager{
				Items: []moduleLib.DeploymentIntentGroup{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil))
			if resp.StatusCode != test.code {
				t.Fatalf("getAllDeploymentIntentGroupsHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.StatusCode)
			}

			if resp.StatusCode == http.StatusOK {
				dig := []moduleLib.DeploymentIntentGroup{}
				json.NewDecoder(resp.Body).Decode(&dig)
				if reflect.DeepEqual(test.result, dig) == false {
					t.Fatalf("getAllDeploymentIntentGroupsHandler returned an unexpected body. Expected %v; Got: %v", test.result, dig)
				}
			}
		})
	}
}

func TestCreateDeploymentIntentGroupHandler(t *testing.T) {
	testCases := []struct {
		err, label string
		client     *mockDeploymentIntentGroupManager
		code       int
		result     moduleLib.DeploymentIntentGroup
		reader     io.Reader
	}{
		{
			label:  "Empty Request Body",
			code:   http.StatusBadRequest,
			client: &mockDeploymentIntentGroupManager{},
			err:    "Empty body",
		},
		{
			label: "Invalid Input. Missing DeploymentIntentGroup Name",
			reader: bytes.NewBuffer([]byte(`{
				"description":"test description"
				}`)),
			code:   http.StatusBadRequest,
			client: &mockDeploymentIntentGroupManager{},
			err:    "Invalid Input",
		},
		{
			label: "Create DeploymentIntentGroup",
			code:  http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testDeploymentIntentGroup",
    				"description": "Test DeploymentIntentGroup used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec": {
					"compositeProfile": "testCompositeProfile",
					"version": "v1",
					"logicalCloud": "testLogicalCloud",
					"overrideValues": [
						{
							"app": "testApp",
							"values": {
								"testPort": "12345"
							}
						}
					]
				}
			}`)),
			result: moduleLib.DeploymentIntentGroup{
				MetaData: moduleLib.DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "Test DeploymentIntentGroup used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.DepSpecData{
					Profile:      "testCompositeProfile",
					Version:      "v1",
					LogicalCloud: "testLogicalCloud",
					OverrideValuesObj: []moduleLib.OverrideValues{
						{
							AppName: "testApp",
							ValuesObj: map[string]string{
								"testPort": "12345",
							},
						},
					},
				},
			},
			client: &mockDeploymentIntentGroupManager{
				Items: []moduleLib.DeploymentIntentGroup{},
			},
		},
		{
			label: "DeploymentIntentGroup Already Exists",
			code:  http.StatusConflict,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testDeploymentIntentGroup",
    				"description": "Test DeploymentIntentGroup used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec": {
					"compositeProfile": "testCompositeProfile",
					"version": "v1",
					"logicalCloud": "testLogicalCloud",
					"overrideValues": [
						{
							"app": "testApp",
							"values": {
								"testPort": "12345"
							}
						}
					]
				}
			}`)),
			result: moduleLib.DeploymentIntentGroup{},
			err:    "Intent already exists",
			client: &mockDeploymentIntentGroupManager{
				Items: []moduleLib.DeploymentIntentGroup{
					{
						MetaData: moduleLib.DepMetaData{
							Name:        "testDeploymentIntentGroup",
							Description: "Test DeploymentIntentGroup used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.DepSpecData{
							Profile:      "testCompositeProfile",
							Version:      "v1",
							LogicalCloud: "testLogicalCloud",
							OverrideValuesObj: []moduleLib.OverrideValues{
								{
									AppName: "testApp",
									ValuesObj: map[string]string{
										"testPort": "12345",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups", test.reader)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("createDeploymentIntentGroupHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if resp.Code == http.StatusCreated {
				dig := moduleLib.DeploymentIntentGroup{}
				json.NewDecoder(resp.Body).Decode(&dig)
				if reflect.DeepEqual(test.result, dig) == false {
					t.Fatalf("createDeploymentIntentGroupHandler returned an unexpected body. Expected %v; Got: %v", test.result, dig)
				}
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("createDeploymentIntentGroupHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}

func TestUpdateDeploymentIntentGroupHandler(t *testing.T) {
	testCases := []struct {
		err, label, name string
		client           *mockDeploymentIntentGroupManager
		code             int
		result           moduleLib.DeploymentIntentGroup
		reader           io.Reader
	}{
		{
			label:  "Empty Request Body",
			name:   "testDeploymentIntentGroup",
			code:   http.StatusBadRequest,
			client: &mockDeploymentIntentGroupManager{},
			err:    "Empty body",
		},
		{
			label: "Invalid Input. Missing DeploymentIntentGroup Name",
			name:  "testDeploymentIntentGroup",
			reader: bytes.NewBuffer([]byte(`{
				"description":"test description"
				}`)),
			code:   http.StatusBadRequest,
			client: &mockDeploymentIntentGroupManager{},
			err:    "Invalid Input",
		},
		{
			label: "Update Existing DeploymentIntentGroup",
			name:  "testDeploymentIntentGroup",
			code:  http.StatusOK,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testDeploymentIntentGroup",
    				"description": "Test DeploymentIntentGroup updated for unit testing",
    				"userData1": "data1_new",
    				"userData2": "data2_new"
				},
				"spec": {
					"compositeProfile": "testCompositeProfile",
					"version": "v1",
					"logicalCloud": "testLogicalCloud",
					"overrideValues": [
						{
							"app": "testApp_New",
							"values": {
								"testPort": "6789"
							}
						}
					]
				}
			}`)),
			result: moduleLib.DeploymentIntentGroup{
				MetaData: moduleLib.DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "Test DeploymentIntentGroup updated for unit testing",
					UserData1:   "data1_new",
					UserData2:   "data2_new",
				},
				Spec: moduleLib.DepSpecData{
					Profile:      "testCompositeProfile",
					Version:      "v1",
					LogicalCloud: "testLogicalCloud",
					OverrideValuesObj: []moduleLib.OverrideValues{
						{
							AppName: "testApp_New",
							ValuesObj: map[string]string{
								"testPort": "6789",
							},
						},
					},
				},
			},
			client: &mockDeploymentIntentGroupManager{
				Items: []moduleLib.DeploymentIntentGroup{
					{
						MetaData: moduleLib.DepMetaData{
							Name:        "testDeploymentIntentGroup_1",
							Description: "Test DeploymentIntentGroup_1 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.DepSpecData{
							Profile:      "testCompositeProfile_1",
							Version:      "v1",
							LogicalCloud: "testLogicalCloud",
							OverrideValuesObj: []moduleLib.OverrideValues{
								{
									AppName: "testApp_1",
									ValuesObj: map[string]string{
										"testPort": "6789",
									},
								},
							},
						},
					},
					{
						MetaData: moduleLib.DepMetaData{
							Name:        "testDeploymentIntentGroup",
							Description: "Test DeploymentIntentGroup used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.DepSpecData{
							Profile:      "testCompositeProfile",
							Version:      "v1",
							LogicalCloud: "testLogicalCloud",
							OverrideValuesObj: []moduleLib.OverrideValues{
								{
									AppName: "testApp_New",
									ValuesObj: map[string]string{
										"testPort": "6789",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			label: "Update Non Existing DeploymentIntentGroup",
			code:  http.StatusCreated,
			name:  "nonExistingDeploymentIntentGroup",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "nonExistingDeploymentIntentGroup",
    				"description": "Test DeploymentIntentGroup used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec": {
					"compositeProfile": "testCompositeProfile",
					"version": "v1",
					"logicalCloud": "testLogicalCloud",
					"overrideValues": [
						{
							"app": "testApp",
							"values": {
								"testPort": "6789"
							}
						}
					]
				}
			}`)),
			result: moduleLib.DeploymentIntentGroup{
				MetaData: moduleLib.DepMetaData{
					Name:        "nonExistingDeploymentIntentGroup",
					Description: "Test DeploymentIntentGroup used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.DepSpecData{
					Profile:      "testCompositeProfile",
					Version:      "v1",
					LogicalCloud: "testLogicalCloud",
					OverrideValuesObj: []moduleLib.OverrideValues{
						{
							AppName: "testApp",
							ValuesObj: map[string]string{
								"testPort": "6789",
							},
						},
					},
				},
			},
			client: &mockDeploymentIntentGroupManager{
				Items: []moduleLib.DeploymentIntentGroup{
					{
						MetaData: moduleLib.DepMetaData{
							Name:        "testDeploymentIntentGroup",
							Description: "Test DeploymentIntentGroup used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.DepSpecData{
							Profile:      "testCompositeProfile",
							Version:      "v1",
							LogicalCloud: "testLogicalCloud",
							OverrideValuesObj: []moduleLib.OverrideValues{
								{
									AppName: "testApp",
									ValuesObj: map[string]string{
										"testPort": "6789",
									},
								},
							},
						},
					},
					{
						MetaData: moduleLib.DepMetaData{
							Name:        "testDeploymentIntentGroup_1",
							Description: "Test DeploymentIntentGroup used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.DepSpecData{
							Profile:      "testCompositeProfile_1",
							Version:      "v1",
							LogicalCloud: "testLogicalCloud",
							OverrideValuesObj: []moduleLib.OverrideValues{
								{
									AppName: "testApp_1",
									ValuesObj: map[string]string{
										"testPort": "6789",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("PUT", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/"+test.name, test.reader)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("putDeploymentIntentGroupHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if resp.Code == http.StatusCreated || resp.Code == http.StatusOK {
				dig := moduleLib.DeploymentIntentGroup{}
				json.NewDecoder(resp.Body).Decode(&dig)
				if reflect.DeepEqual(test.result, dig) == false {
					t.Fatalf("putDeploymentIntentGroupHandler returned an unexpected body. Expected %v; Got: %v", test.result, dig)
				}
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("putDeploymentIntentGroupHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}

func TestDeleteDeploymentIntentGroupHandler(t *testing.T) {
	testCases := []struct {
		err, label, name string
		client           *mockDeploymentIntentGroupManager
		code             int
		result           moduleLib.DeploymentIntentGroup
	}{
		{
			label: "Delete DeploymentIntentGroup",
			code:  http.StatusNoContent,
			name:  "testDeploymentIntentGroup",
			client: &mockDeploymentIntentGroupManager{
				Items: []moduleLib.DeploymentIntentGroup{
					{
						MetaData: moduleLib.DepMetaData{
							Name:        "testDeploymentIntentGroup_1",
							Description: "Test testDeploymentIntentGroup used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.DepSpecData{
							Profile:      "testCompositeProfile_1",
							Version:      "v1",
							LogicalCloud: "testLogicalCloud",
							OverrideValuesObj: []moduleLib.OverrideValues{
								{
									AppName: "testApp_1",
									ValuesObj: map[string]string{
										"testPort": "6789",
									},
								},
							},
						},
					},
					{
						MetaData: moduleLib.DepMetaData{
							Name:        "testDeploymentIntentGroup",
							Description: "Test DeploymentIntentGroup used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.DepSpecData{
							Profile:      "testCompositeProfile",
							Version:      "v1",
							LogicalCloud: "testLogicalCloud",
							OverrideValuesObj: []moduleLib.OverrideValues{
								{
									AppName: "testApp",
									ValuesObj: map[string]string{
										"testPort": "6789",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			label: "Delete Non Existing DeploymentIntentGroup",
			code:  http.StatusNotFound,
			err:   "The requested resource not found",
			name:  "nonExistingDeploymentIntentGroup",
			client: &mockDeploymentIntentGroupManager{
				Items: []moduleLib.DeploymentIntentGroup{
					{
						MetaData: moduleLib.DepMetaData{
							Name:        "testDeploymentIntentGroup",
							Description: "Test DeploymentIntentGroup used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.DepSpecData{
							Profile:      "testCompositeProfile",
							Version:      "v1",
							LogicalCloud: "testLogicalCloud",
							OverrideValuesObj: []moduleLib.OverrideValues{
								{
									AppName: "testApp",
									ValuesObj: map[string]string{
										"testPort": "6789",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/"+test.name, nil)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("deleteDeploymentIntentGroupHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("deleteDeploymentIntentGroupHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}
