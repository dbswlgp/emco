// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"bytes"
	"encoding/json"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

func init() {
	appDepJSONFile = "../json-schemas/app-dependency.json"
	db.DBconn = &db.NewMockDB{}
}

func Test_appdepencency_createHandler(t *testing.T) {
	testCases := []struct {
		label        string
		reader       io.Reader
		expected     moduleLib.AppDependency
		expectedCode int
	}{
		{
			label:        "Missing Body Failure",
			expectedCode: http.StatusBadRequest,
		},
		{
			label:        "Create App Dependency",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testAppDependency",
					"description": "Test AppDependency used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"opStatus":"Ready",
					"wait": 0,
					"app": "test"
				}
			}`)),
			expected: moduleLib.AppDependency{
				MetaData: moduleLib.AdMetaData{
					Name:        "testAppDependency",
					Description: "Test AppDependency used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.AdSpecData{
					OpStatus: "Ready",
					Wait:     0,
					AppName:  "test",
				},
			},
		},
		{
			label:        "Create App Dependency Bad Json",
			expectedCode: http.StatusUnprocessableEntity,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testAppDependency",
					"description": "Test AppDependency used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec: {
					"opStatus":Ready,
					"wait": 0,
					"app": test }
			}`)),
		},
		{
			label:        "Create App Dependency Bad OpStatus",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testAppDependencyError",
					"description": "Test AppDependency used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"opStatus":"deploye",
					"wait": 0,
					"app": "test"
				}
			}`)),
		},
		{
			label: "Missing App Dependency Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"description":"test description"
				}`)),
			expectedCode: http.StatusBadRequest,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/projects/{project}/composite-apps/{compositeApp}/{version}/apps/{AppName}/dependency", testCase.reader)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, moduleLib.NewAppDependencyClient()))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := moduleLib.AppDependency{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}

}

func Test_appdepencency_updateHandler(t *testing.T) {
	testCases := []struct {
		label           string
		reader, reader1 io.Reader
		expected        moduleLib.AppDependency
		expectedCode    int
		expectedName    string
	}{
		{
			label:        "Update App Dependency",
			expectedCode: http.StatusCreated,
			expectedName: "testAppDependency",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testAppDependency",
					"description": "Test AppDependency used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"opStatus":"Ready",
					"wait": 0,
					"app": "test" }
			}`)),
			reader1: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testAppDependency",
					"description": "Test AppDependency used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"opStatus":"Deployed",
					"wait": 0,
					"app": "test" }
			}`)),
			expected: moduleLib.AppDependency{
				MetaData: moduleLib.AdMetaData{
					Name:        "testAppDependency",
					Description: "Test AppDependency used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.AdSpecData{
					OpStatus: "Deployed",
					Wait:     0,
					AppName:  "test",
				},
			},
		},
		{
			label:        "Update wrong App Dependency",
			expectedCode: http.StatusBadRequest,
			expectedName: "wrongAppDependency",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testAppDependency",
					"description": "Test AppDependency used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"opStatus":"Ready",
					"wait": 0,
					"app": "test" }
			}`)),
			reader1: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testAppDependency",
					"description": "Test AppDependency used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"opStatus":"Deployed",
					"wait": 0,
					"app": "test" }
			}`)),
			expected: moduleLib.AppDependency{
				MetaData: moduleLib.AdMetaData{
					Name:        "testAppDependency",
					Description: "Test AppDependency used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.AdSpecData{
					OpStatus: "Deployed",
					Wait:     0,
					AppName:  "test",
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			// Create first
			request := httptest.NewRequest("POST", "/v2/projects/{project}/composite-apps/{compositeApp}/{version}/apps/{app}/dependency", testCase.reader)
			_ = executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, moduleLib.NewAppDependencyClient()))

			request = httptest.NewRequest("PUT", "/v2/projects/{project}/composite-apps/{compositeApp}/{version}/apps/{app}/dependency/"+testCase.expectedName, testCase.reader1)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, moduleLib.NewAppDependencyClient()))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := moduleLib.AppDependency{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("updateHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}

}

func Test_appdepencency_getHandler(t *testing.T) {
	testCases := []struct {
		label        string
		reader       io.Reader
		expected     moduleLib.AppDependency
		expectedCode int
		expectedName string
	}{
		{
			label:        "Get App Dependency",
			expectedCode: http.StatusOK,
			expectedName: "testAppDependency",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testAppDependency",
					"description": "Test AppDependency used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"opStatus":"Ready",
					"wait": 0,
					"app": "test" }
			}`)),
			expected: moduleLib.AppDependency{
				MetaData: moduleLib.AdMetaData{
					Name:        "testAppDependency",
					Description: "Test AppDependency used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.AdSpecData{
					OpStatus: "Ready",
					Wait:     0,
					AppName:  "test",
				},
			},
		},
		{
			label:        "Get Wrong App Dependency",
			expectedCode: http.StatusNotFound,
			expectedName: "wrongAppDependency",
		},
	}
	// Create first
	request := httptest.NewRequest("POST", "/v2/projects/{project}/composite-apps/{compositeApp}/{version}/apps/{app}/dependency", testCases[0].reader)
	_ = executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, moduleLib.NewAppDependencyClient()))

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {

			request = httptest.NewRequest("GET", "/v2/projects/{project}/composite-apps/{compositeApp}/{version}/apps/{app}/dependency/"+testCase.expectedName, nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, moduleLib.NewAppDependencyClient()))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusOK {
				got := moduleLib.AppDependency{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("getHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func Test_appdepencency_getAllHandler(t *testing.T) {
	testCases := []struct {
		label        string
		reader       io.Reader
		expected     []moduleLib.AppDependency
		expectedCode int
		expectedName string
	}{
		{
			label:        "GetAll App Dependency",
			expectedCode: http.StatusOK,
			expectedName: "testAppDependency",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testAppDependency",
					"description": "Test AppDependency used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"opStatus":"Ready",
					"wait": 0,
					"app": "test" }
			}`)),
			expected: []moduleLib.AppDependency{{
				MetaData: moduleLib.AdMetaData{
					Name:        "testAppDependency",
					Description: "Test AppDependency used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.AdSpecData{
					OpStatus: "Ready",
					Wait:     0,
					AppName:  "test",
				},
			},
			},
		},
	}
	db.DBconn = &db.NewMockDB{}
	// Create first
	request := httptest.NewRequest("POST", "/v2/projects/{project}/composite-apps/{compositeApp}/{version}/apps/{app}/dependency", testCases[0].reader)
	_ = executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, moduleLib.NewAppDependencyClient()))

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {

			request = httptest.NewRequest("GET", "/v2/projects/{project}/composite-apps/{compositeApp}/{version}/apps/{app}/dependency", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, moduleLib.NewAppDependencyClient()))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusOK {
				got := []moduleLib.AppDependency{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetAllHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func Test_appdepencency_deleteHandler(t *testing.T) {
	testCases := []struct {
		label        string
		reader       io.Reader
		expected     moduleLib.AppDependency
		expectedCode int
		expectedName string
	}{
		{
			label:        "Delete App Dependency",
			expectedCode: http.StatusNoContent,
			expectedName: "testAppDependency",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testAppDependency",
					"description": "Test AppDependency used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"opStatus":"Ready",
					"wait": 0,
					"app": "test" }
			}`)),
			expected: moduleLib.AppDependency{
				MetaData: moduleLib.AdMetaData{
					Name:        "testAppDependency",
					Description: "Test AppDependency used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.AdSpecData{
					OpStatus: "Ready",
					Wait:     0,
					AppName:  "test",
				},
			},
		},
		{
			label:        "Delete Wrong App Dependency",
			expectedCode: http.StatusNotFound,
			expectedName: "wrongAppDependency",
		},
	}
	// Create first
	request := httptest.NewRequest("POST", "/v2/projects/{project}/composite-apps/{compositeApp}/{version}/apps/{app}/dependency", testCases[0].reader)
	_ = executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, moduleLib.NewAppDependencyClient()))

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {

			request = httptest.NewRequest("DELETE", "/v2/projects/{project}/composite-apps/{compositeApp}/{version}/apps/{app}/dependency/"+testCase.expectedName, nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, moduleLib.NewAppDependencyClient()))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

		})
	}
}
