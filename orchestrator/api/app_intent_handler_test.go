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
	gpic "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/gpic"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockAppIntentManager struct {
	// Items and err will be used to customize each test
	// via a localized instantiation of mockAppIntentManager
	Items    []moduleLib.AppIntent
	Err      error
	SpecData moduleLib.SpecData
}

func (aim *mockAppIntentManager) CreateAppIntent(ctx context.Context, ai moduleLib.AppIntent, project, compositeApp, version, genericPlacementIntent, deploymentIntentGroup string, failIfExists bool) (moduleLib.AppIntent, bool, error) {
	iExists := false
	index := 0

	if aim.Err != nil {
		return moduleLib.AppIntent{}, iExists, aim.Err
	}

	for i, item := range aim.Items {
		if item.MetaData.Name == ai.MetaData.Name {
			iExists = true
			index = i
			break
		}
	}

	if iExists && failIfExists { // resource already exists
		return moduleLib.AppIntent{}, iExists, pkgerrors.New("Intent already exists")
	}

	if iExists && !failIfExists { // resource already exists. update the resource
		aim.Items[index] = ai
		return aim.Items[index], iExists, nil
	}

	aim.Items = append(aim.Items, ai) // create the resource

	return aim.Items[len(aim.Items)-1], iExists, nil
}

func (aim *mockAppIntentManager) GetAppIntent(ctx context.Context, appIntent, project, compositeApp, version, genericPlacementIntent, deploymentIntentGroup string) (moduleLib.AppIntent, error) {
	if aim.Err != nil {
		return moduleLib.AppIntent{}, aim.Err
	}

	for _, item := range aim.Items {
		if item.MetaData.Name == appIntent {
			return item, nil
		}
	}

	return moduleLib.AppIntent{}, pkgerrors.New("Intent not found")
}

func (aim *mockAppIntentManager) GetAllIntentsByApp(ctx context.Context, app, project, compositeApp, version, genericPlacementIntent, deploymentIntentGroup string) (moduleLib.SpecData, error) {
	if aim.Err != nil {
		return moduleLib.SpecData{}, aim.Err
	}

	for _, item := range aim.Items {
		if item.Spec.AppName == app {
			return item.Spec, nil
		}
	}

	return moduleLib.SpecData{}, nil
}

func (aim *mockAppIntentManager) DeleteAppIntent(ctx context.Context, appIntent, project, compositeApp, version, genericPlacementIntent, deploymentIntentGroup string) error {
	if aim.Err != nil {
		return aim.Err
	}

	for k, item := range aim.Items {
		if item.MetaData.Name == appIntent { // resource exist. delete it
			aim.Items[k] = aim.Items[len(aim.Items)-1]
			aim.Items[len(aim.Items)-1] = moduleLib.AppIntent{}
			aim.Items = aim.Items[:len(aim.Items)-1]
			return nil
		}
	}

	return pkgerrors.New("db Remove resource not found") // resource does not exist
}

func (aim *mockAppIntentManager) GetAllAppIntents(ctx context.Context, project, compositeApp, version, genericPlacementIntent, deploymentIntentGroup string) ([]moduleLib.AppIntent, error) {
	if aim.Err != nil {
		return []moduleLib.AppIntent{}, aim.Err
	}

	if len(aim.Items) > 0 {
		return aim.Items, nil
	}

	return []moduleLib.AppIntent{}, nil
}

func init() {
	appIntentJSONFile = "../json-schemas/generic-placement-intent-app.json"
}

func TestGetAppIntentHandler(t *testing.T) {
	testCases := []struct {
		err, label, name string
		client           *mockAppIntentManager
		code             int
		result           moduleLib.AppIntent
	}{
		{
			label: "Get AppIntent",
			name:  "testAppIntent",
			code:  http.StatusOK,
			result: moduleLib.AppIntent{
				MetaData: moduleLib.MetaData{
					Name:        "testAppIntent",
					Description: "Test AppIntent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.SpecData{
					AppName: "testApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName:     "aws",
								ClusterLabelName: "west-us1",
							},
						},
					},
				},
			},
			client: &mockAppIntentManager{
				Items: []moduleLib.AppIntent{
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent_1",
							Description: "Test AppIntent_1 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp_1",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent",
							Description: "Test AppIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			label:  "AppIntent Not Found",
			name:   "nonExistingAppIntent",
			code:   http.StatusNotFound,
			err:    "Intent not found",
			result: moduleLib.AppIntent{},
			client: &mockAppIntentManager{
				Items: []moduleLib.AppIntent{
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent_1",
							Description: "Test AppIntent_1 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp_1",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent_2",
							Description: "Test AppIntent_2 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp_2",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
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
			request := httptest.NewRequest("GET", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents/"+test.name, nil)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("getAppIntentHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if resp.Code == http.StatusOK {
				ai := moduleLib.AppIntent{}
				json.NewDecoder(resp.Body).Decode(&ai)
				if reflect.DeepEqual(test.result, ai) == false {
					t.Fatalf("getAppIntentHandler returned an unexpected body. Expected %v; Got: %v", test.result, ai)
				}
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("getAppIntentHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}

func TestGetAllAppIntentsHandler(t *testing.T) {
	testCases := []struct {
		err, label, name string
		client           *mockAppIntentManager
		code             int
		result           []moduleLib.AppIntent
	}{
		{
			label: "Get All AppIntents",
			code:  http.StatusOK,
			result: []moduleLib.AppIntent{
				{
					MetaData: moduleLib.MetaData{
						Name:        "testAppIntent_1",
						Description: "Test AppIntent_1 used for unit testing",
						UserData1:   "data1",
						UserData2:   "data2",
					},
					Spec: moduleLib.SpecData{
						AppName: "testApp_1",
						Intent: gpic.IntentStruc{
							AllOfArray: []gpic.AllOf{
								{
									ProviderName: "aws",
									ClusterName:  "edge1",
								},
								{
									ProviderName:     "aws",
									ClusterLabelName: "west-us1",
								},
							},
						},
					},
				},
				{
					MetaData: moduleLib.MetaData{
						Name:        "testAppIntent",
						Description: "Test AppIntent used for unit testing",
						UserData1:   "data1",
						UserData2:   "data2",
					},
					Spec: moduleLib.SpecData{
						AppName: "testApp",
						Intent: gpic.IntentStruc{
							AllOfArray: []gpic.AllOf{
								{
									ProviderName: "aws",
									ClusterName:  "edge1",
								},
								{
									ProviderName:     "aws",
									ClusterLabelName: "west-us1",
								},
							},
						},
					},
				},
			},
			client: &mockAppIntentManager{
				Items: []moduleLib.AppIntent{
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent_1",
							Description: "Test AppIntent_1 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp_1",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent",
							Description: "Test AppIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			label:  "Get All AppIntents Not Exists",
			code:   http.StatusOK,
			result: []moduleLib.AppIntent{},
			client: &mockAppIntentManager{
				Items: []moduleLib.AppIntent{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents", nil)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("getAllAppIntentsHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if resp.Code == http.StatusOK {
				ai := []moduleLib.AppIntent{}
				json.NewDecoder(resp.Body).Decode(&ai)
				if reflect.DeepEqual(test.result, ai) == false {
					t.Fatalf("getAllAppIntentsHandler returned an unexpected body. Expected %v; Got: %v", test.result, ai)
				}
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("getAllAppIntentsHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}

func TestCreateAppIntentHandler(t *testing.T) {
	testCases := []struct {
		err, label string
		client     *mockAppIntentManager
		code       int
		result     moduleLib.AppIntent
		reader     io.Reader
	}{
		{
			label:  "Empty Request Body",
			code:   http.StatusBadRequest,
			client: &mockAppIntentManager{},
			err:    "Empty body",
		},
		{
			label: "Missing AppIntent Name",
			code:  http.StatusBadRequest,
			err:   "Missing name for the intent",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"app": "testApp",
					"intent": {
						"allOf": [
							{
								"clusterProvider": "testClusterProvider_1",
								"clusterLabel": "testClusterLabel_1"
							},
							{
								"clusterProvider": "testClusterProvider_2",
								"clusterLabel": "testClusterLabel_2"
							}
						]
					}
				}
		  }`)),
		},
		{
			label: "Missing App Name",
			code:  http.StatusBadRequest,
			err:   "Missing app for the intent",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"intent": {
						"allOf": [
							{
								"clusterProvider": "testClustrProvider_1",
								"cluster": "testCluster_1"
							},
							{
								"clusterProvider": "testClusterProvider_1",
								"clusterLabel": "testClusterLabel_1"
							}
						]
					}
				}
		  }`)),
		},
		{
			label: "Missing ClusterProvider Name",
			code:  http.StatusBadRequest,
			err:   "Missing clusterProvider in an intent",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				 "spec": {
					 "app": "testApp",
					 "intent": {
						 "anyOf": [
							  {
								  "cluster": "testCluster",
								  "clusterLabel": "testClusterLabel"
							  }
							]
						}
			  		}
			 	}
		  }`)),
		},
		{
			label: "ClusterLabel or Cluster is missing",
			code:  http.StatusBadRequest,
			err:   "Missing cluster or clusterLabel",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"app": "testApp",
					"intent": {
						"anyOf": [
							{
								"clusterProvider": "testClusterProvider"
							}
						]
					}
				}
		  }`)),
		},
		{
			label: "AnyOf Only Clusterlabel or Cluster Allowed",
			code:  http.StatusBadRequest,
			err:   "Only one of cluster name or cluster label allowed",
			reader: bytes.NewBuffer([]byte(`{ 
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				 "spec": { 
					 "app": "testApp",
					 "intent": { 
						 "anyOf": [ 
							{ 
								 "clusterProvider": "testClusterProvider",
								 "clusterLabel": "testClusterLabel",
								 "cluster": "testCluster"
							}
						]
					}
			  	}
		  }`)),
		},
		{
			label: "AllOf Missing ClusterProvider Name",
			code:  http.StatusBadRequest,
			err:   "Missing clusterProvider in an intent",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": { 
					"app": "testApp", 
					"intent": {
						"allOf": [
							{
								"name": "testClusterProvider_1",
								"clusterLabel": "testClusterLabel_1"
							},
							{
								"anyOf": [
									{
										"clusterProvider": "testClusterProvider_2",
										"clusterLabel": "testClusterLabel_2"
									}
								]
							}
						]
					}
				}
			}`)),
		},
		{
			label: "AllOf AnyOf Missing ClusterProvider Name",
			code:  http.StatusBadRequest,
			err:   "Missing clusterProvider in an intent",
			reader: bytes.NewBuffer([]byte(`{ 
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": { 
					"app": "testApp",  
					"intent": {
						"allOf": [
							{
								"clusterProvider": "testClusterProvider_1",
								"clusterLabel": "testClusterLabel_1"
							},
							{
								"anyOf": [
									{
										"name": "testClusterProvider_2",
										"clusterLabel": "testClusterLabel_2"
									}
								]
							}
						]
					}
				}
			}`)),
		},
		{
			label: "AllOf Only Clusterlabel or Cluster Allowed",
			code:  http.StatusBadRequest,
			err:   "Only one of cluster name or cluster label allowed",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": { 
					"app": "testApp",
					"intent": {
						"allOf": [
							{
								"clusterProvider": "testClusterProvider_1",
								"clusterLabel": "testClusterLabel_1",
								"cluster": "e"
							},
							{
								"anyOf": [
									{
										"clusterProvider": "testClusterProvider_2",
										"clusterLabel": "testClusterLabel_2"
									}
								]
							}
						]
					}
				}
			}`)),
		},
		{
			label: "AllOf AnyOf Only Clusterlabel or Cluster Allowed",
			code:  http.StatusBadRequest,
			err:   "Only one of cluster name or cluster label allowed",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": { 
					"app": "testApp",
					"intent": {
						"allOf": [
							{
								"clusterProvider": "testClusterProvider_1",
								"clusterLabel": "testClusterLabel_1"
							},
							{
								"anyOf": [
									{
										"clusterProvider": "testClusterProvider_2",
										"clusterLabel": "testClusterLabel_2",
										"cluster": "testCluster"
									}
								]
							}
						]
					}
				}
			}`)),
		},
		{
			label: "Create AppIntent",
			code:  http.StatusCreated,
			err:   "",
			reader: bytes.NewBuffer([]byte(`{   
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": { 
					"app": "testApp",
					"intent": {
						"allOf": [
							{
								"clusterProvider": "aws",
								"cluster": "edge1"
							},
							{
								"clusterProvider": "aws",
								"clusterLabel": "west-us1"
							}
						]
					}
				}
			}`)),
			result: moduleLib.AppIntent{
				MetaData: moduleLib.MetaData{
					Name:        "testAppIntent",
					Description: "Test AppIntent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.SpecData{
					AppName: "testApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName:     "aws",
								ClusterLabelName: "west-us1",
							},
						},
					},
				},
			},
			client: &mockAppIntentManager{
				Items: []moduleLib.AppIntent{
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent_1",
							Description: "Test AppIntent_1 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp_1",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			label: "AppIntent Already Exists",
			code:  http.StatusConflict,
			err:   "Intent already exists",
			reader: bytes.NewBuffer([]byte(`{   
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": { 
					"app": "testApp",
					"intent": {
						"allOf": [
							{
								"clusterProvider": "aws",
								"cluster": "edge1"
							},
							{
								"clusterProvider": "aws",
								"clusterLabel": "west-us1"
							}
						]
					}
				}
			}`)),
			result: moduleLib.AppIntent{},
			client: &mockAppIntentManager{
				Items: []moduleLib.AppIntent{
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent_1",
							Description: "Test AppIntent_1 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp_1",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent",
							Description: "Test AppIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
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
			request := httptest.NewRequest("POST", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents", test.reader)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("createAppIntentHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if resp.Code == http.StatusOK {
				ai := moduleLib.AppIntent{}
				json.NewDecoder(resp.Body).Decode(&ai)
				if reflect.DeepEqual(test.result, ai) == false {
					t.Fatalf("createAppIntentHandler returned an unexpected body. Expected %v; Got: %v", test.result, ai)
				}
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("createAppIntentHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}

		})
	}
}

func TestUpdateAppIntentHandler(t *testing.T) {
	testCases := []struct {
		err, label, name string
		client           *mockAppIntentManager
		code             int
		result           moduleLib.AppIntent
		reader           io.Reader
	}{

		{
			label:  "Empty Request Body",
			name:   "testAppIntent",
			code:   http.StatusBadRequest,
			client: &mockAppIntentManager{},
			err:    "Empty body",
		},
		{
			label: "Missing AppIntent Name",
			name:  "testAppIntent",
			code:  http.StatusBadRequest,
			err:   "Missing name for the intent",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"app": "testApp",
					"intent": {
						"allOf": [
							{
								"clusterProvider": "testClusterProvider_1",
								"clusterLabel": "testClusterLabel_1"
							},
							{
								"clusterProvider": "testClusterProvider_2",
								"clusterLabel": "testClusterLabel_2"
							}
						]
					}
				}
		  }`)),
		},
		{
			label: "Missing App Name",
			name:  "testAppIntent",
			code:  http.StatusBadRequest,
			err:   "Missing app for the intent",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"intent": {
						"allOf": [
							{
								"clusterProvider": "testClustrProvider_1",
								"cluster": "testCluster_1"
							},
							{
								"clusterProvider": "testClusterProvider_1",
								"clusterLabel": "testClusterLabel_1"
							}
						]
					}
				}
		  }`)),
		},
		{
			label: "Missing ClusterProvider Name",
			name:  "testAppIntent",
			code:  http.StatusBadRequest,
			err:   "Missing clusterProvider in an intent",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				 "spec": {
					 "app": "testApp",
					 "intent": {
						 "anyOf": [
							  {
								  "cluster": "testCluster",
								  "clusterLabel": "testClusterLabel"
							  }
							]
						}
			  		}
			 	}
		  }`)),
		},
		{
			label: "ClusterLabel or Cluster is missing",
			name:  "testAppIntent",
			code:  http.StatusBadRequest,
			err:   "Missing cluster or clusterLabel",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": {
					"app": "testApp",
					"intent": {
						"anyOf": [
							{
								"clusterProvider": "testClusterProvider"
							}
						]
					}
				}
		  }`)),
		},
		{
			label: "AnyOf Only Clusterlabel or Cluster Allowed",
			name:  "testAppIntent",
			code:  http.StatusBadRequest,
			err:   "Only one of cluster name or cluster label allowed",
			reader: bytes.NewBuffer([]byte(`{ 
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				 "spec": { 
					 "app": "testApp",
					 "intent": { 
						 "anyOf": [ 
							{ 
								 "clusterProvider": "testClusterProvider",
								 "clusterLabel": "testClusterLabel",
								 "cluster": "testCluster"
							}
						]
					}
			  	}
		  }`)),
		},
		{
			label: "AllOf Missing ClusterProvider Name",
			name:  "testAppIntent",
			code:  http.StatusBadRequest,
			err:   "Missing clusterProvider in an intent",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": { 
					"app": "testApp", 
					"intent": {
						"allOf": [
							{
								"name": "testClusterProvider_1",
								"clusterLabel": "testClusterLabel_1"
							},
							{
								"anyOf": [
									{
										"clusterProvider": "testClusterProvider_2",
										"clusterLabel": "testClusterLabel_2"
									}
								]
							}
						]
					}
				}
			}`)),
		},
		{
			label: "AllOf AnyOf Missing ClusterProvider Name",
			name:  "testAppIntent",
			code:  http.StatusBadRequest,
			err:   "Missing clusterProvider in an intent",
			reader: bytes.NewBuffer([]byte(`{ 
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": { 
					"app": "testApp",  
					"intent": {
						"allOf": [
							{
								"clusterProvider": "testClusterProvider_1",
								"clusterLabel": "testClusterLabel_1"
							},
							{
								"anyOf": [
									{
										"name": "testClusterProvider_2",
										"clusterLabel": "testClusterLabel_2"
									}
								]
							}
						]
					}
				}
			}`)),
		},
		{
			label: "AllOf Only Clusterlabel or Cluster Allowed",
			name:  "testAppIntent",
			code:  http.StatusBadRequest,
			err:   "Only one of cluster name or cluster label allowed",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": { 
					"app": "testApp",
					"intent": {
						"allOf": [
							{
								"clusterProvider": "testClusterProvider_1",
								"clusterLabel": "testClusterLabel_1",
								"cluster": "e"
							},
							{
								"anyOf": [
									{
										"clusterProvider": "testClusterProvider_2",
										"clusterLabel": "testClusterLabel_2"
									}
								]
							}
						]
					}
				}
			}`)),
		},
		{
			label: "AllOf AnyOf Only Clusterlabel or Cluster Allowed",
			name:  "testAppIntent",
			code:  http.StatusBadRequest,
			err:   "Only one of cluster name or cluster label allowed",
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": { 
					"app": "testApp",
					"intent": {
						"allOf": [
							{
								"clusterProvider": "testClusterProvider_1",
								"clusterLabel": "testClusterLabel_1"
							},
							{
								"anyOf": [
									{
										"clusterProvider": "testClusterProvider_2",
										"clusterLabel": "testClusterLabel_2",
										"cluster": "testCluster"
									}
								]
							}
						]
					}
				}
			}`)),
		},
		{
			label: "Update Non Existing AppIntent",
			code:  http.StatusCreated,
			name:  "nonExistingAppIntent",
			reader: bytes.NewBuffer([]byte(`{   
				"metadata": {
					"name": "nonExistingAppIntent",
					"description": "Test AppIntent used for unit testing",
					"userData1": "data1",
					"userData2": "data2"
				},
				"spec": { 
					"app": "testApp",
					"intent": {
						"allOf": [
							{
								"clusterProvider": "aws",
								"cluster": "edge1"
							},
							{
								"clusterProvider": "aws",
								"clusterLabel": "west-us1"
							}
						]
					}
				}
			}`)),
			result: moduleLib.AppIntent{
				MetaData: moduleLib.MetaData{
					Name:        "nonExistingAppIntent",
					Description: "Test AppIntent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: moduleLib.SpecData{
					AppName: "testApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName:     "aws",
								ClusterLabelName: "west-us1",
							},
						},
					},
				},
			},
			client: &mockAppIntentManager{
				Items: []moduleLib.AppIntent{
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent_1",
							Description: "Test AppIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp_1",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent",
							Description: "Test AppIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			label: "Update Existing AppIntent",
			code:  http.StatusOK,
			name:  "testAppIntent",
			reader: bytes.NewBuffer([]byte(`{   
				"metadata": {
					"name": "testAppIntent",
					"description": "Test AppIntent updated for unit testing",
					"userData1": "data1_new",
					"userData2": "data2_new"
				},
				"spec": { 
					"app": "testUpdatedApp",
					"intent": {
						"allOf": [
							{
								"clusterProvider": "aws",
								"cluster": "edge1"
							},
							{
								"clusterProvider": "aws",
								"clusterLabel": "west-us1"
							}
						]
					}
				}
			}`)),
			result: moduleLib.AppIntent{
				MetaData: moduleLib.MetaData{
					Name:        "testAppIntent",
					Description: "Test AppIntent updated for unit testing",
					UserData1:   "data1_new",
					UserData2:   "data2_new",
				},
				Spec: moduleLib.SpecData{
					AppName: "testUpdatedApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName:     "aws",
								ClusterLabelName: "west-us1",
							},
						},
					},
				},
			},
			client: &mockAppIntentManager{
				Items: []moduleLib.AppIntent{
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent_1",
							Description: "Test AppIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp_1",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent",
							Description: "Test AppIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
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
			request := httptest.NewRequest("PUT", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents/"+test.name, test.reader)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("putAppIntentHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if resp.Code == http.StatusCreated || resp.Code == http.StatusOK {
				ai := moduleLib.AppIntent{}
				json.NewDecoder(resp.Body).Decode(&ai)
				if reflect.DeepEqual(test.result, ai) == false {
					t.Fatalf("putAppIntentHandler returned an unexpected body. Expected %v; Got: %v", test.result, ai)
				}
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("putAppIntentHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}

func TestDeleteAppIntentHandler(t *testing.T) {
	testCases := []struct {
		err, label, name string
		client           *mockAppIntentManager
		code             int
		result           moduleLib.AppIntent
	}{
		{
			label: "Delete AppIntent",
			code:  http.StatusNoContent,
			name:  "testAppIntent",
			client: &mockAppIntentManager{
				Items: []moduleLib.AppIntent{
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent_1",
							Description: "Test AppIntent_1 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp_1",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent",
							Description: "Test AppIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			label: "Delete Non Existing AppIntent",
			code:  http.StatusNotFound,
			err:   "The requested resource not found",
			name:  "nonExistingAppIntent",
			client: &mockAppIntentManager{
				Items: []moduleLib.AppIntent{
					{
						MetaData: moduleLib.MetaData{
							Name:        "testAppIntent",
							Description: "Test AppIntent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: moduleLib.SpecData{
							AppName: "testApp",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
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
			request := httptest.NewRequest("DELETE", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents/"+test.name, nil)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, test.client, nil, nil, nil, nil, nil, nil))
			if resp.Code != test.code {
				t.Fatalf("deleteAppIntentHandler returned an unexpected status. Expected %d; Got: %d", test.code, resp.Code)
			}

			if strings.Contains(resp.Body.String(), test.err) == false {
				t.Fatalf("deleteAppIntentHandler returned an unexpected error. Expected %s; Got: %s", test.err, resp.Body.String())
			}
		})
	}
}
