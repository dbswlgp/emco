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

// mockGenericPlacementIntentManager allows us to mock the GenericPlacementIntentManager functionalities and the database connections
type mockGenericPlacementIntentManager struct {
	Err   error
	Items []moduleLib.GenericPlacementIntent
}

func (gpim *mockGenericPlacementIntentManager) GetGenericPlacementIntent(ctx context.Context, intentName string, projectName string, compositeAppName string, version string, digName string) (moduleLib.GenericPlacementIntent, error) {
	if gpim.Err != nil {
		return moduleLib.GenericPlacementIntent{}, gpim.Err
	}

	for _, item := range gpim.Items {
		if item.MetaData.Name == intentName {
			return item, nil
		}
	}

	return moduleLib.GenericPlacementIntent{}, pkgerrors.New("Intent not found")
}

func (gpim *mockGenericPlacementIntentManager) GetAllGenericPlacementIntents(ctx context.Context, p string, ca string, v string, digName string) ([]moduleLib.GenericPlacementIntent, error) {
	if gpim.Err != nil {
		return []moduleLib.GenericPlacementIntent{}, gpim.Err
	}

	if len(gpim.Items) > 0 {
		return gpim.Items, nil
	}

	return []moduleLib.GenericPlacementIntent{}, nil
}

func (gpim *mockGenericPlacementIntentManager) CreateGenericPlacementIntent(ctx context.Context, g moduleLib.GenericPlacementIntent, p string, ca string, v string, digName string, failIfExists bool) (moduleLib.GenericPlacementIntent, bool, error) {
	gpiExists := false
	index := 0

	if gpim.Err != nil {
		return moduleLib.GenericPlacementIntent{}, gpiExists, gpim.Err
	}

	for i, item := range gpim.Items {
		if item.MetaData.Name == g.MetaData.Name {
			gpiExists = true
			index = i
			break
		}
	}

	if gpiExists && failIfExists { // resource already exists
		return moduleLib.GenericPlacementIntent{}, gpiExists, pkgerrors.New("Intent already exists")
	}

	if gpiExists && !failIfExists { // resource already exists. update the resource
		gpim.Items[index] = g
		return gpim.Items[index], gpiExists, nil
	}

	gpim.Items = append(gpim.Items, g) // create the resource

	return gpim.Items[len(gpim.Items)-1], gpiExists, nil

}

func (gpim *mockGenericPlacementIntentManager) DeleteGenericPlacementIntent(ctx context.Context, intentName string, projectName string, compositeAppName string, version string, digName string) error {
	if gpim.Err != nil {
		return gpim.Err
	}

	for i, item := range gpim.Items {
		if item.MetaData.Name == intentName { // resource exist
			gpim.Items[i] = gpim.Items[len(gpim.Items)-1]
			gpim.Items[len(gpim.Items)-1] = moduleLib.GenericPlacementIntent{}
			gpim.Items = gpim.Items[:len(gpim.Items)-1]
			return nil
		}
	}

	return pkgerrors.New("db Remove resource not found") // resource does not exist
}

func init() {
	gpiJSONFile = "../json-schemas/generic-placement-intent.json"
}

func TestGetGenericPlacementHandler(t *testing.T) {
	testCases := []struct {
		err, label, name string
		client           *mockGenericPlacementIntentManager
		code             int
		result           moduleLib.GenericPlacementIntent
	}{
		{
			label: "Get GenericPlacementIntent",
			code:  http.StatusOK,
			result: moduleLib.GenericPlacementIntent{
				MetaData: moduleLib.GenIntentMetaData{
					Name:        "testGenericPlacementIntent",
					Description: "Test GenericPlacementIntent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			name: "testGenericPlacementIntent",
			client: &mockGenericPlacementIntentManager{
				Items: []moduleLib.GenericPlacementIntent{
					{
						MetaData: moduleLib.GenIntentMetaData{
							Name:        "testGenericPlacementIntent_1",
							Description: "Test GenericPlacementIntent_1 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
					{
						MetaData: moduleLib.GenIntentMetaData{
							Name:        "testGenericPlacementIntent",
							Description: "Test GenericPlacementIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label: "GenericPlacementIntent Not Found",
			code:  http.StatusNotFound,
			err:   "Intent not found",
			name:  "nonExistingGenericPlacementIntent",
			client: &mockGenericPlacementIntentManager{
				Items: []moduleLib.GenericPlacementIntent{
					{
						MetaData: moduleLib.GenIntentMetaData{
							Name:        "testGenericPlacementIntent_1",
							Description: "Test GenericPlacementIntent_1 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
					{
						MetaData: moduleLib.GenIntentMetaData{
							Name:        "testGenericPlacementIntent",
							Description: "Test GenericPlacementIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/"+test.name, nil)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("getGenericPlacementHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if resp.Code == http.StatusOK {
				gpi := moduleLib.GenericPlacementIntent{}
				json.NewDecoder(resp.Body).Decode(&gpi)
				if reflect.DeepEqual(test.result, gpi) == false {
					t.Fatalf("getGenericPlacementHandler returned an unexpected body. Expected %v; Got: %v", test.result, gpi)
				}
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("getGenericPlacementHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}

func TestGetAllGenericPlacementIntentsHandler(t *testing.T) {
	testCases := []struct {
		err, label string
		client     *mockGenericPlacementIntentManager
		code       int
		result     []moduleLib.GenericPlacementIntent
	}{
		{
			label: "Get All GenericPlacementIntents",
			code:  http.StatusOK,
			result: []moduleLib.GenericPlacementIntent{
				{
					MetaData: moduleLib.GenIntentMetaData{
						Name:        "testGenericPlacementIntent_1",
						Description: "Test GenericPlacementIntent_1 used for unit testing",
						UserData1:   "data1",
						UserData2:   "data2",
					},
				},
				{
					MetaData: moduleLib.GenIntentMetaData{
						Name:        "testGenericPlacementIntent_2",
						Description: "Test GenericPlacementIntent_2 used for unit testing",
						UserData1:   "data1",
						UserData2:   "data2",
					},
				},
			},
			client: &mockGenericPlacementIntentManager{
				Items: []moduleLib.GenericPlacementIntent{
					{
						MetaData: moduleLib.GenIntentMetaData{
							Name:        "testGenericPlacementIntent_1",
							Description: "Test GenericPlacementIntent_1 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
					{
						MetaData: moduleLib.GenIntentMetaData{
							Name:        "testGenericPlacementIntent_2",
							Description: "Test GenericPlacementIntent_2 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label:  "Get All GenericPlacementIntents Not Exists",
			code:   http.StatusOK,
			result: []moduleLib.GenericPlacementIntent{},
			client: &mockGenericPlacementIntentManager{
				Items: []moduleLib.GenericPlacementIntent{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil, nil, nil))
			if resp.StatusCode != test.code {
				t.Fatalf("getAllGenericPlacementIntentsHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.StatusCode)
			}

			if resp.StatusCode == http.StatusOK {
				gpi := []moduleLib.GenericPlacementIntent{}
				json.NewDecoder(resp.Body).Decode(&gpi)
				if reflect.DeepEqual(test.result, gpi) == false {
					t.Fatalf("getAllGenericPlacementIntentsHandler returned an unexpected body. Expected %v; Got: %v", test.result, gpi)
				}
			}
		})
	}
}

func TestCreateGenericPlacementIntentHandler(t *testing.T) {
	testCases := []struct {
		err, label string
		client     *mockGenericPlacementIntentManager
		code       int
		result     moduleLib.GenericPlacementIntent
		reader     io.Reader
	}{
		{
			label:  "Empty Request Body",
			code:   http.StatusBadRequest,
			client: &mockGenericPlacementIntentManager{},
			err:    "Empty body",
		},
		{
			label: "Invalid Input. Missing GenericPlacementIntent Name",
			reader: bytes.NewBuffer([]byte(`{
				"description":"test description"
				}`)),
			code:   http.StatusBadRequest,
			client: &mockGenericPlacementIntentManager{},
			err:    "Invalid Input",
		},
		{
			label: "Create GenericPlacementIntent",
			code:  http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testGenericPlacementIntent",
    				"description": "Test GenericPlacementIntent used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				}
			}`)),
			result: moduleLib.GenericPlacementIntent{
				MetaData: moduleLib.GenIntentMetaData{
					Name:        "testGenericPlacementIntent",
					Description: "Test GenericPlacementIntent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			client: &mockGenericPlacementIntentManager{
				Items: []moduleLib.GenericPlacementIntent{},
			},
		},
		{
			label: "GenericPlacementIntent Already Exists",
			code:  http.StatusConflict,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testGenericPlacementIntent",
    				"description": "Test GenericPlacementIntent used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				}
			}`)),
			result: moduleLib.GenericPlacementIntent{},
			err:    "Intent already exists",
			client: &mockGenericPlacementIntentManager{
				Items: []moduleLib.GenericPlacementIntent{
					{
						MetaData: moduleLib.GenIntentMetaData{
							Name:        "testGenericPlacementIntent",
							Description: "Test GenericPlacementIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents", test.reader)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("createGenericPlacementIntentHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if resp.Code == http.StatusCreated {
				gpi := moduleLib.GenericPlacementIntent{}
				json.NewDecoder(resp.Body).Decode(&gpi)
				if reflect.DeepEqual(test.result, gpi) == false {
					t.Fatalf("createGenericPlacementIntentHandler returned an unexpected body. Expected %v; Got: %v", test.result, gpi)
				}
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("createGenericPlacementIntentHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}

func TestUpdateGenericPlacementHandler(t *testing.T) {
	testCases := []struct {
		err, label, name string
		client           *mockGenericPlacementIntentManager
		code             int
		result           moduleLib.GenericPlacementIntent
		reader           io.Reader
	}{
		{
			label:  "Empty Request Body",
			name:   "testGenericPlacementIntent",
			code:   http.StatusBadRequest,
			client: &mockGenericPlacementIntentManager{},
			err:    "Empty body",
		},
		{
			label: "Invalid Input. Missing GenericPlacementIntent Name",
			name:  "testGenericPlacementIntent",
			reader: bytes.NewBuffer([]byte(`{
				"description":"test description"
				}`)),
			code:   http.StatusBadRequest,
			client: &mockGenericPlacementIntentManager{},
			err:    "Invalid Input",
		},
		{
			label: "Update Existing GenericPlacementIntent",
			name:  "testGenericPlacementIntent",
			code:  http.StatusOK,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testGenericPlacementIntent",
    				"description": "Test GenericPlacementIntent updated for unit testing",
    				"userData1": "data1_new",
    				"userData2": "data2_new"
				}
			}`)),
			result: moduleLib.GenericPlacementIntent{
				MetaData: moduleLib.GenIntentMetaData{
					Name:        "testGenericPlacementIntent",
					Description: "Test GenericPlacementIntent updated for unit testing",
					UserData1:   "data1_new",
					UserData2:   "data2_new",
				},
			},
			client: &mockGenericPlacementIntentManager{
				Items: []moduleLib.GenericPlacementIntent{
					{
						MetaData: moduleLib.GenIntentMetaData{
							Name:        "testGenericPlacementIntent_1",
							Description: "Test GenericPlacementIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
					{
						MetaData: moduleLib.GenIntentMetaData{
							Name:        "testGenericPlacementIntent",
							Description: "Test GenericPlacementIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label: "Update Non Existing GenericPlacementIntent",
			name:  "nonExistingGenericPlacementIntent",
			code:  http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "nonExistingGenericPlacementIntent",
    				"description": "Test GenericPlacementIntent used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				}
			}`)),
			result: moduleLib.GenericPlacementIntent{
				MetaData: moduleLib.GenIntentMetaData{
					Name:        "nonExistingGenericPlacementIntent",
					Description: "Test GenericPlacementIntent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			client: &mockGenericPlacementIntentManager{
				Items: []moduleLib.GenericPlacementIntent{
					{
						MetaData: moduleLib.GenIntentMetaData{
							Name:        "testGenericPlacementIntent",
							Description: "Test GenericPlacementIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
					{
						MetaData: moduleLib.GenIntentMetaData{
							Name:        "testGenericPlacementIntent_1",
							Description: "Test GenericPlacementIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("PUT", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/"+test.name, test.reader)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("putGenericPlacementHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if resp.Code == http.StatusCreated || resp.Code == http.StatusOK {
				gpi := moduleLib.GenericPlacementIntent{}
				json.NewDecoder(resp.Body).Decode(&gpi)
				if reflect.DeepEqual(test.result, gpi) == false {
					t.Fatalf("putGenericPlacementHandler returned an unexpected body. Expected %v; Got: %v", test.result, gpi)
				}
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("putGenericPlacementHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}

func TestDeleteGenericPlacementHandler(t *testing.T) {
	testCases := []struct {
		client           *mockGenericPlacementIntentManager
		err, label, name string
		code             int
		result           moduleLib.GenericPlacementIntent
	}{
		{
			label: "Delete GenericPlacementIntent",
			code:  http.StatusNoContent,
			name:  "testGenericPlacementIntent",
			client: &mockGenericPlacementIntentManager{
				Items: []moduleLib.GenericPlacementIntent{
					{
						MetaData: moduleLib.GenIntentMetaData{
							Name:        "testGenericPlacementIntent_1",
							Description: "Test testGenericPlacementIntent_1 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
					{
						MetaData: moduleLib.GenIntentMetaData{
							Name:        "testGenericPlacementIntent",
							Description: "Test GenericPlacementIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label: "Delete Non Existing GenericPlacementIntent",
			code:  http.StatusNotFound,
			err:   "The requested resource not found",
			name:  "nonExistingGenericPlacementIntent",
			client: &mockGenericPlacementIntentManager{
				Items: []moduleLib.GenericPlacementIntent{
					{
						MetaData: moduleLib.GenIntentMetaData{
							Name:        "testGenericPlacementIntent",
							Description: "Test GenericPlacementIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/"+test.name, nil)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("deleteGenericPlacementHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("deleteGenericPlacementHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}
