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
)

// mockIntentManager allows us to mock the IntentManager functionalities and the database connections
type mockIntentManager struct {
	Err           error
	Items         []moduleLib.Intent
	ListOfIntents moduleLib.ListOfIntents
}

func (im *mockIntentManager) GetIntent(ctx context.Context, intent, project, compositeApp, version, deploymentIntentGroup string) (moduleLib.Intent, error) {
	if im.Err != nil {
		return moduleLib.Intent{}, im.Err
	}

	for _, item := range im.Items {
		if item.MetaData.Name == intent {
			return item, nil
		}
	}

	return moduleLib.Intent{}, pkgerrors.New("Intent not found")
}

func (im *mockIntentManager) GetIntentByName(ctx context.Context, intent, project, compositeApp, version, deploymentIntentGroup string) (moduleLib.IntentSpecData, error) {
	if im.Err != nil {
		return moduleLib.IntentSpecData{}, im.Err
	}

	for _, item := range im.Items {
		if item.MetaData.Name == intent {
			return item.Spec, nil
		}
	}

	return moduleLib.IntentSpecData{}, pkgerrors.New("Intent not found")
}

func (im *mockIntentManager) GetAllIntents(ctx context.Context, project, compositeApp, version, deploymentIntentGroup string) (moduleLib.ListOfIntents, error) {
	if im.Err != nil {
		return moduleLib.ListOfIntents{}, im.Err
	}

	if len(im.ListOfIntents.ListOfIntents) > 0 {
		return im.ListOfIntents, nil
	}

	return moduleLib.ListOfIntents{}, nil
}

func (im *mockIntentManager) AddIntent(ctx context.Context, intent moduleLib.Intent, project, compositeApp, version, deploymentIntentGroup string, failIfExists bool) (moduleLib.Intent, bool, error) {
	iExists := false
	index := 0

	if im.Err != nil {
		return moduleLib.Intent{}, iExists, im.Err
	}

	for i, item := range im.Items {
		if item.MetaData.Name == intent.MetaData.Name {
			iExists = true
			index = i
			break
		}
	}

	if iExists && failIfExists { // resource already exists
		return moduleLib.Intent{}, iExists, pkgerrors.New("Intent already exists")
	}

	if iExists && !failIfExists { // resource already exists. update the resource
		im.Items[index] = intent
		return im.Items[index], iExists, nil
	}

	im.Items = append(im.Items, intent) // create the resource

	return im.Items[len(im.Items)-1], iExists, nil

}

func (im *mockIntentManager) DeleteIntent(ctx context.Context, intent, project, compositeApp, version, deploymentIntentGroup string) error {
	if im.Err != nil {
		return im.Err
	}

	for k, item := range im.Items {
		if item.MetaData.Name == intent { // resource exist
			im.Items[k] = im.Items[len(im.Items)-1]
			im.Items[len(im.Items)-1] = moduleLib.Intent{}
			im.Items = im.Items[:len(im.Items)-1]
			return nil
		}
	}

	return pkgerrors.New("db Remove resource not found") // resource does not exist
}

func init() {
	addIntentJSONFile = "../json-schemas/deployment-intent.json"
}

func TestGetIntentHandler(t *testing.T) {
	testCases := []struct {
		err, label, name string
		client           *mockIntentManager
		code             int
		result           moduleLib.Intent
	}{
		{
			label: "Get Intent",
			code:  http.StatusOK,
			result: moduleLib.Intent{
				MetaData: moduleLib.IntentMetaData{
					Name:        "testIntent",
					Description: "Test Intent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
				},
			},
			name: "testIntent",
			client: &mockIntentManager{
				Items: []moduleLib.Intent{
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent_1",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
				},
			},
		},
		{
			label: "Intent Not Found",
			code:  http.StatusNotFound,
			err:   "Intent not found",
			name:  "nonExistingIntent",
			client: &mockIntentManager{
				Items: []moduleLib.Intent{
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents/"+test.name, nil)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("getIntentHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if resp.Code == http.StatusOK {
				i := moduleLib.Intent{}
				json.NewDecoder(resp.Body).Decode(&i)
				if reflect.DeepEqual(test.result, i) == false {
					t.Fatalf("getIntentHandler returned an unexpected body. Expected %v; Got: %v", test.result, i)
				}
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("getIntentHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}

		})
	}
}

func TestGetIntentByNameHandler(t *testing.T) {
	testCases := []struct {
		err, label, name string
		client           *mockIntentManager
		code             int
		result           moduleLib.Intent
	}{
		{
			label: "Get Intent",
			code:  http.StatusOK,
			result: moduleLib.Intent{
				MetaData: moduleLib.IntentMetaData{
					Name:        "testIntent",
					Description: "Test Intent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
				},
			},
			name: "testIntent",
			client: &mockIntentManager{
				Items: []moduleLib.Intent{
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
				},
			},
		},
		{
			label: "Intent Not Found",
			code:  http.StatusNotFound,
			err:   "Intent not found",
			name:  "nonExistingIntent",
			client: &mockIntentManager{
				Items: []moduleLib.Intent{
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents/"+test.name, nil)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("getIntentByNameHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if resp.Code == http.StatusOK {
				i := moduleLib.Intent{}
				json.NewDecoder(resp.Body).Decode(&i)
				if reflect.DeepEqual(test.result, i) == false {
					t.Fatalf("getIntentByNameHandler returned an unexpected body. Expected %v; Got: %v", test.result, i)
				}
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("getIntentByNameHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}

		})
	}
}

func TestGetAllIntentsHandler(t *testing.T) {
	testCases := []struct {
		label  string
		client *mockIntentManager
		code   int
		result moduleLib.ListOfIntents
	}{
		{
			label: "Get All Intents",
			code:  http.StatusOK,
			result: moduleLib.ListOfIntents{
				ListOfIntents: []map[string]string{
					{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
					{
						"ovnaction": "testGenericPlacementIntent",
					},
				},
			},
			client: &mockIntentManager{
				Items: []moduleLib.Intent{
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent_1",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"ovnaction": "testGenericPlacementIntent",
							},
						},
					},
				},
				ListOfIntents: moduleLib.ListOfIntents{
					ListOfIntents: []map[string]string{
						{
							"genericPlacementIntent": "testGenericPlacementIntent",
						},
						{
							"ovnaction": "testGenericPlacementIntent",
						},
					},
				},
			},
		},
		{
			label:  "Get All Intents Not Exists",
			code:   http.StatusOK,
			result: moduleLib.ListOfIntents{},
			client: &mockIntentManager{
				Items: []moduleLib.Intent{
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent_1",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"ovnaction": "testGenericPlacementIntent",
							},
						},
					},
				},
				ListOfIntents: moduleLib.ListOfIntents{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil))
			if resp.StatusCode != test.code {
				t.Fatalf("getAllIntentsHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.StatusCode)
			}

			if resp.StatusCode == http.StatusOK {
				iList := moduleLib.ListOfIntents{}
				json.NewDecoder(resp.Body).Decode(&iList)
				if reflect.DeepEqual(test.result, iList) == false {
					t.Fatalf("getAllIntentsHandler returned an unexpected body. Expected %v; Got: %v", test.result, iList)
				}
			}
		})
	}
}

func TestAddIntentHandler(t *testing.T) {
	testCases := []struct {
		err, label string
		client     *mockIntentManager
		code       int
		result     moduleLib.Intent
		reader     io.Reader
	}{
		{
			label:  "Empty Request Body",
			code:   http.StatusBadRequest,
			client: &mockIntentManager{},
			err:    "Empty body",
		},
		{
			label: "Invalid Input. Missing Intent Name",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "Test Intent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"intent": {
						"genericPlacementIntent": "testGenericPlacementIntent"
					}
				}
			}`)),
			code:   http.StatusBadRequest,
			client: &mockIntentManager{},
			err:    "Invalid Input",
		},
		{
			label: "Invalid Input. Missing Intent Spec Data",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testIntent",
					"description": "Test Intent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
				}
			}`)),
			code:   http.StatusBadRequest,
			client: &mockIntentManager{},
			err:    "Invalid Input",
		},
		{
			label: "Create Intent",
			code:  http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testIntent",
					"description": "Test Intent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"intent": {
						"genericPlacementIntent": "testGenericPlacementIntent"
					}
				}
			}`)),
			result: moduleLib.Intent{
				MetaData: moduleLib.IntentMetaData{
					Name:        "testIntent",
					Description: "Test Intent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
				},
			},
			client: &mockIntentManager{
				Items: []moduleLib.Intent{
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent_1",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
				},
			},
		},
		{
			label: "Intent Already Exists",
			code:  http.StatusConflict,
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testIntent",
					"description": "Test Intent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"intent": {
						"genericPlacementIntent": "testGenericPlacementIntent"
					}
				}
			}`)),
			result: moduleLib.Intent{},
			err:    "Intent already exists",
			client: &mockIntentManager{
				Items: []moduleLib.Intent{
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent_1",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents", test.reader)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("addIntentHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if resp.Code == http.StatusCreated {
				i := moduleLib.Intent{}
				json.NewDecoder(resp.Body).Decode(&i)
				if reflect.DeepEqual(test.result, i) == false {
					t.Fatalf("addIntentHandler returned an unexpected body. Expected %v; Got: %v", test.result, i)
				}
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("addIntentHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}

func TestUpdateIntentHandler(t *testing.T) {
	testCases := []struct {
		err, label, name string
		client           *mockIntentManager
		code             int
		result           moduleLib.Intent
		reader           io.Reader
	}{
		{
			label:  "Empty Request Body",
			name:   "testIntent",
			code:   http.StatusBadRequest,
			client: &mockIntentManager{},
			err:    "Empty body",
		},
		{
			label: "Invalid Input. Missing Intent Name",
			name:  "testIntent",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "Test Intent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"intent": {
						"genericPlacementIntent": "testGenericPlacementIntent"
					}
				}
			}`)),
			code:   http.StatusBadRequest,
			client: &mockIntentManager{},
			err:    "Invalid Input",
		},
		{
			label: "Invalid Input. Missing Intent Spec Data",
			name:  "testIntent",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testIntent",
					"description": "Test Intent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
				}
			}`)),
			code:   http.StatusBadRequest,
			client: &mockIntentManager{},
			err:    "Invalid Input",
		},
		{
			label: "Update Existing Intent",
			name:  "testIntent",
			code:  http.StatusOK,
			reader: bytes.NewBuffer([]byte(`{
					"metadata" : {
					"name": "testIntent",
    				"description": "Test Intent updated for unit testing",
    				"userData1": "data1_new",
    				"userData2": "data2_new"
				},
				"spec":{
					"intent":{
						"ovnaction":"testOvnaction"
					 }
				}
			}`)),
			result: moduleLib.Intent{
				MetaData: moduleLib.IntentMetaData{
					Name:        "testIntent",
					Description: "Test Intent updated for unit testing",
					UserData1:   "data1_new",
					UserData2:   "data2_new",
				},
				Spec: moduleLib.IntentSpecData{
					Intent: map[string]string{
						"ovnaction": "testOvnaction",
					},
				},
			},
			client: &mockIntentManager{
				Items: []moduleLib.Intent{
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent_1",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
				},
			},
		},
		{
			label: "Update Non Existing Intent",
			name:  "nonExistingIntent",
			code:  http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "nonExistingIntent",
    				"description": "Test NonExistingIntent used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec":{
					"intent":{
						"ovnaction":"testOvnaction"
					 }
				}
			}`)),
			result: moduleLib.Intent{
				MetaData: moduleLib.IntentMetaData{
					Name:        "nonExistingIntent",
					Description: "Test NonExistingIntent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.IntentSpecData{
					Intent: map[string]string{
						"ovnaction": "testOvnaction",
					},
				},
			},
			client: &mockIntentManager{
				Items: []moduleLib.Intent{
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("PUT", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents/"+test.name, test.reader)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("putIntentHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if resp.Code == http.StatusCreated || resp.Code == http.StatusOK {
				i := moduleLib.Intent{}
				json.NewDecoder(resp.Body).Decode(&i)
				if reflect.DeepEqual(test.result, i) == false {
					t.Fatalf("putIntentHandler returned an unexpected body. Expected %v; Got: %v", test.result, i)
				}
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("putIntentHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}

func TestDeleteIntentHandler(t *testing.T) {
	testCases := []struct {
		err, label, name string
		client           *mockIntentManager
		code             int
	}{
		{
			label: "Delete Intent",
			code:  http.StatusNoContent,
			name:  "testIntent",
			client: &mockIntentManager{
				Items: []moduleLib.Intent{
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent_1",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
				},
			},
		},
		{
			label: "Delete Non Existing Intent",
			code:  http.StatusNotFound,
			err:   "The requested resource not found",
			name:  "nonExistingIntent",
			client: &mockIntentManager{
				Items: []moduleLib.Intent{
					{
						MetaData: moduleLib.IntentMetaData{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.IntentSpecData{
							Intent: map[string]string{
								"genericPlacementIntent": "testGenericPlacementIntent",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents/"+test.name, nil)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("deleteIntentHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("deleteIntentHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}
