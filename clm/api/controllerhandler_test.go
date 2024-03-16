// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	pkgerrors "github.com/pkg/errors"
	clmModel "gitlab.com/project-emco/core/emco-base/src/clm/pkg/model"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

func init() {
	controllerJSONFile = "../json-schemas/controller.json"
}

func TestControllerCreateHandler(t *testing.T) {
	testCases := []struct {
		label            string
		reader           io.Reader
		expected         clmModel.Controller
		expectedCode     int
		ControllerClient *mockControllerManager
	}{
		{
			label:            "Missing Body Failure",
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label:        "Create Controller",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"host":"mycontroller",
					"port":9095,
					"priority":1
				}
			}`)),
			expected: clmModel.Controller{
				Metadata: mtypes.Metadata{
					Name:        "TestController",
					Description: "Test Controller used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: clmModel.ControllerSpec{
					Host:     "mycontroller",
					Port:     9095,
					Priority: 1,
				},
			},
			ControllerClient: &mockControllerManager{
				//Items that will be returned by the mocked Client
				Items: []clmModel.Controller{
					{
						Metadata: mtypes.Metadata{
							Name:        "TestController",
							Description: "Test Controller used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: clmModel.ControllerSpec{
							Host:     "mycontroller",
							Port:     9095,
							Priority: 1,
						},
					},
				},
			},
		},
		{
			label:        "Create Controller without specifying Priority",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"host":"mycontroller",
					"port":9095
				}
			}`)),
			expected: clmModel.Controller{
				Metadata: mtypes.Metadata{
					Name:        "TestController",
					Description: "Test Controller used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: clmModel.ControllerSpec{
					Host: "mycontroller",
					Port: 9095,
				},
			},
			ControllerClient: &mockControllerManager{
				//Items that will be returned by the mocked Client
				Items: []clmModel.Controller{
					{
						Metadata: mtypes.Metadata{
							Name:        "TestController",
							Description: "Test Controller used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: clmModel.ControllerSpec{
							Host: "mycontroller",
							Port: 9095,
						},
					},
				},
			},
		},
		{
			label: "Missing Controller Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
                "description":"test description"
                }`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Empty Controller Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"name": "",
                "description":"test description"
                }`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Missing  Host Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"port":9095
				}
			}`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Empty Host Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"host":""
				}
			}`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Missing  Port Number in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"host":"clm"
				}
			}`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Empty Port number in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"port":0
				}
			}`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Missing  Host Name & Port Number in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
				}
			}`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Empty Host Name & Port number in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"host":"",
					"port":0
				}
			}`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Invalid Host Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"host":"clm@"
				}
			}`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
	}

	fmt.Printf("\n================== TestControllerCreateHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestControllerCreateHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/clm-controllers", testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.ControllerClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := clmModel.Controller{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestControllerUpdateHandler(t *testing.T) {
	testCases := []struct {
		label            string
		name             string
		reader           io.Reader
		expected         clmModel.Controller
		expectedCode     int
		ControllerClient *mockControllerManager
	}{
		{
			label:            "Missing Body Failure",
			name:             "TestController",
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label:        "Update Controller",
			name:         "TestController",
			expectedCode: http.StatusOK,
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"host":"mycontroller",
					"port":9095,
					"priority":1
				}
			}`)),
			expected: clmModel.Controller{
				Metadata: mtypes.Metadata{
					Name:        "TestController",
					Description: "Test Controller used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: clmModel.ControllerSpec{
					Host:     "mycontroller",
					Port:     9095,
					Priority: 1,
				},
			},
			ControllerClient: &mockControllerManager{
				//Items that will be returned by the mocked Client
				Items: []clmModel.Controller{
					{
						Metadata: mtypes.Metadata{
							Name:        "TestController",
							Description: "Test Controller used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: clmModel.ControllerSpec{
							Host:     "mycontroller",
							Port:     9095,
							Priority: 1,
						},
					},
				},
			},
		},
		{
			label: "Missing Controller Name in Request Body",
			name:  "TestController",
			reader: bytes.NewBuffer([]byte(`{
                "description":"test description"
                }`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Empty Controller Name in Request Body",
			name:  "TestController",
			reader: bytes.NewBuffer([]byte(`{
				"name": "",
                "description":"test description"
                }`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Missing  Host Name in Request Body",
			name:  "TestController",
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"port":9095
				}
			}`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Empty Host Name in Request Body",
			name:  "TestController",
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"host":""
				}
			}`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Missing  Port Number in Request Body",
			name:  "TestController",
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"host":"clm"
				}
			}`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Empty Port number in Request Body",
			name:  "TestController",
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"port":0
				}
			}`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Missing  Host Name & Port Number in Request Body",
			name:  "TestController",
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
				}
			}`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Empty Host Name & Port number in Request Body",
			name:  "TestController",
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"host":"",
					"port":0
				}
			}`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
		{
			label: "Invalid Host Name in Request Body",
			name:  "TestController",
			reader: bytes.NewBuffer([]byte(`{
				"Metadata" : {
					"name": "TestController",
    				"description": "Test Controller used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"host":"clm@"
				}
			}`)),
			expectedCode:     http.StatusBadRequest,
			ControllerClient: &mockControllerManager{},
		},
	}

	fmt.Printf("\n================== TestControllerUpdateHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestControllerUpdateHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("PUT", "/v2/clm-controllers/"+testCase.name, testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.ControllerClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := clmModel.Controller{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("updateHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestControllerGetHandler(t *testing.T) {

	testCases := []struct {
		label            string
		expected         clmModel.Controller
		name, version    string
		expectedCode     int
		ControllerClient *mockControllerManager
	}{
		{
			label:        "Get Controller Metadata",
			expectedCode: http.StatusOK,
			expected: clmModel.Controller{
				Metadata: mtypes.Metadata{
					Name:        "TestController",
					Description: "Test Controller used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			name: "TestController",
			ControllerClient: &mockControllerManager{
				Items: []clmModel.Controller{
					{
						Metadata: mtypes.Metadata{
							Name:        "TestController",
							Description: "Test Controller used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label:        "Get Controller request",
			expectedCode: http.StatusOK,
			expected: clmModel.Controller{
				Metadata: mtypes.Metadata{
					Name:        "TestController",
					Description: "Test Controller used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: clmModel.ControllerSpec{Host: "app1", Port: 9016},
			},
			name: "TestController",
			ControllerClient: &mockControllerManager{
				Items: []clmModel.Controller{
					{
						Metadata: mtypes.Metadata{
							Name:        "TestController",
							Description: "Test Controller used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: clmModel.ControllerSpec{Host: "app1", Port: 9016},
					},
				},
			},
		},
		{
			label:        "Get Non-Exiting Controller",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingController",
			ControllerClient: &mockControllerManager{
				Items: []clmModel.Controller{},
				Err:   pkgerrors.New("ClmController not found"),
			},
		},
		{
			label:        "Get Controller db error",
			expectedCode: http.StatusInternalServerError,
			name:         "testGetControllerDBError",
			ControllerClient: &mockControllerManager{
				Items: []clmModel.Controller{},
				Err:   pkgerrors.New("db Find error"),
			},
		},
	}

	fmt.Printf("\n================== TestControllerGetHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestControllerGetHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/clm-controllers/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.ControllerClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := clmModel.Controller{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestControllerDeleteHandler(t *testing.T) {

	testCases := []struct {
		label            string
		name             string
		version          string
		expectedCode     int
		ControllerClient *mockControllerManager
	}{
		{
			label:        "Delete Controller",
			expectedCode: http.StatusNoContent,
			name:         "TestController",
			ControllerClient: &mockControllerManager{
				//Items that will be returned by the mocked Client
				Items: []clmModel.Controller{
					{
						Metadata: mtypes.Metadata{
							Name:        "TestController",
							Description: "Test Controller used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label:        "Delete Non-Exiting Controller",
			expectedCode: http.StatusNotFound,
			name:         "TestController",
			ControllerClient: &mockControllerManager{
				Err: pkgerrors.New("db Remove resource not found"),
			},
		},
	}

	fmt.Printf("\n================== TestControllerDeleteHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestControllerDeleteHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/clm-controllers/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.ControllerClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}

// Controller Mock
type mockControllerManager struct {
	Items []clmModel.Controller
	Err   error
}

func (m *mockControllerManager) CreateController(ctx context.Context, ms clmModel.Controller, mayExist bool) (clmModel.Controller, error) {
	if m.Err != nil {
		return clmModel.Controller{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockControllerManager) GetController(ctx context.Context, name string) (clmModel.Controller, error) {
	if m.Err != nil {
		return clmModel.Controller{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockControllerManager) GetControllers(ctx context.Context) ([]clmModel.Controller, error) {
	return []clmModel.Controller{}, m.Err
}

func (m *mockControllerManager) DeleteController(ctx context.Context, name string) error {
	return m.Err
}

func (m *mockControllerManager) InitControllers(ctx context.Context) {
}
