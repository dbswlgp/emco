// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"reflect"
	"strings"
	"testing"

	gpic "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/gpic"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

func TestCreateAppIntent(t *testing.T) {
	testCases := []struct {
		compositeApp           string
		deploymentIntentGroup  string
		err                    string
		genericPlacementIntent string
		label                  string
		project                string
		version                string
		db                     *db.MockDB
		ai                     AppIntent
		result                 AppIntent
	}{
		{
			label: "Create AppIntent",
			ai: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: "A sample AppIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: SpecData{
					AppName: "testApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us2",
									},
								},
							},
						},

						AnyOfArray: []gpic.AnyOf{},
					},
				},
			},
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			genericPlacementIntent: "testGenericPlacementIntent",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			result: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: "A sample AppIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: SpecData{
					AppName: "testApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us2",
									},
								},
							},
						},
						AnyOfArray: []gpic.AnyOf{},
					},
				},
			},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project\":\"testProject\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						CompositeAppKey{CompositeAppName: "testCompositeApp",
							Version: "testCompositeAppVersion", Project: "testProject"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"testCompositeApp\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testIntent\"," +
									"\"Description\":\"A sample intent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"Logical-Cloud\": \"logicalCloud1\"}}"),
						},
						DeploymentIntentGroupKey{
							Name:         "testDeploymentIntentGroup",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"testDeploymentIntentGroup\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"overrideValues\":[" +
									"{" +
									"\"app\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logicalCloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label: "AppIntent Already Exists",
			ai: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: "A sample AppIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: SpecData{
					AppName: "testApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us2",
									},
								},
							},
						},
						AnyOfArray: []gpic.AnyOf{},
					},
				},
			},
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			genericPlacementIntent: "testGenericPlacementIntent",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			err:                    "Intent already exists",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project\":\"testProject\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						CompositeAppKey{CompositeAppName: "testCompositeApp",
							Version: "testCompositeAppVersion", Project: "testProject"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"testCompositeApp\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testIntent\"," +
									"\"Description\":\"A sample intent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"Logical-Cloud\": \"logicalCloud1\"}}"),
						},
						DeploymentIntentGroupKey{
							Name:         "testDeploymentIntentGroup",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"testDeploymentIntentGroup\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"overrideValues\":[" +
									"{" +
									"\"app\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logicalCloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						AppIntentKey{
							Name:                      "testAppIntent",
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							Version:                   "testCompositeAppVersion",
							Intent:                    "testGenericPlacementIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent\"," +
									"\"Description\":\"A sample AppIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"testApp\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = test.db
			cli := NewAppIntentClient()
			appIntent, _, err := cli.CreateAppIntent(ctx, test.ai, test.project, test.compositeApp, test.version, test.genericPlacementIntent, test.deploymentIntentGroup, true)
			if err != nil {
				if test.err == "" {
					t.Fatalf("CreateAppIntent returned an unexpected error %s, ", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("CreateAppIntent returned an unexpected error %s", err.Error())
				}
			}

			if reflect.DeepEqual(test.result, appIntent) == false {
				t.Errorf("CreateAppIntent returned an unexpected body: got %v; "+" result %v", appIntent, test.result)
			}
		})
	}
}

func TestUpdateAppIntent(t *testing.T) {
	testCases := []struct {
		compositeApp           string
		deploymentIntentGroup  string
		err                    string
		genericPlacementIntent string
		label                  string
		project                string
		version                string
		db                     *db.MockDB
		exists                 bool
		ai                     AppIntent
		result                 AppIntent
	}{
		{
			label:  "Update Non Existing AppIntent",
			exists: false,
			ai: AppIntent{
				MetaData: MetaData{
					Name:        "nonExistingAppIntent",
					Description: "A sample AppIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: SpecData{
					AppName: "testApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us2",
									},
								},
							},
						},

						AnyOfArray: []gpic.AnyOf{},
					},
				},
			},
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			genericPlacementIntent: "testGenericPlacementIntent",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			result: AppIntent{
				MetaData: MetaData{
					Name:        "nonExistingAppIntent",
					Description: "A sample AppIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: SpecData{
					AppName: "testApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us2",
									},
								},
							},
						},

						AnyOfArray: []gpic.AnyOf{},
					},
				},
			},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project\":\"testProject\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						CompositeAppKey{CompositeAppName: "testCompositeApp",
							Version: "testCompositeAppVersion", Project: "testProject"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"testCompositeApp\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testIntent\"," +
									"\"Description\":\"A sample intent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"Logical-Cloud\": \"logicalCloud1\"}}"),
						},
						DeploymentIntentGroupKey{
							Name:         "testDeploymentIntentGroup",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"testDeploymentIntentGroup\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"overrideValues\":[" +
									"{" +
									"\"app\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logicalCloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						AppIntentKey{
							Name:                      "testAppIntent",
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							Version:                   "testCompositeAppVersion",
							Intent:                    "testGenericPlacementIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent\"," +
									"\"Description\":\"A sample AppIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"testApp\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
					},
				},
			},
		},
		{
			label:  "Update Existing AppIntent",
			exists: true,
			ai: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: "This is the new description for testAppIntent",
					UserData1:   "This is the new data for userData1",
					UserData2:   "This is the new data for userData2",
				},
				Spec: SpecData{
					AppName: "testApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us2",
									},
								},
							},
						},

						AnyOfArray: []gpic.AnyOf{},
					},
				},
			},

			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			genericPlacementIntent: "testGenericPlacementIntent",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			result: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: "This is the new description for testAppIntent",
					UserData1:   "This is the new data for userData1",
					UserData2:   "This is the new data for userData2",
				},
				Spec: SpecData{
					AppName: "testApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us2",
									},
								},
							},
						},
						AnyOfArray: []gpic.AnyOf{},
					},
				},
			},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project\":\"testProject\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						CompositeAppKey{CompositeAppName: "testCompositeApp",
							Version: "testCompositeAppVersion", Project: "testProject"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"testCompositeApp\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testIntent\"," +
									"\"Description\":\"A sample intent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"Logical-Cloud\": \"logicalCloud1\"}}"),
						},
						DeploymentIntentGroupKey{
							Name:         "testDeploymentIntentGroup",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"testDeploymentIntentGroup\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"overrideValues\":[" +
									"{" +
									"\"app\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logicalCloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						AppIntentKey{
							Name:                      "testAppIntent",
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							Version:                   "testCompositeAppVersion",
							Intent:                    "testGenericPlacementIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent\"," +
									"\"Description\":\"testAppIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"testApp\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
						AppIntentKey{
							Name:                      "testAppIntent_1",
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							Version:                   "testCompositeAppVersion",
							Intent:                    "testGenericPlacementIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent_1\"," +
									"\"Description\":\"testAppIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"testApp\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = test.db
			cli := NewAppIntentClient()
			appIntent, aiExists, err := cli.CreateAppIntent(ctx, test.ai, test.project, test.compositeApp, test.version, test.genericPlacementIntent, test.deploymentIntentGroup, false)
			if err != nil {
				if test.err == "" {
					t.Fatalf("CreateAppIntent returned an unexpected error %s, ", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("CreateAppIntent returned an unexpected error %s", err.Error())
				}
			}

			if reflect.DeepEqual(test.result, appIntent) == false {
				t.Errorf("CreateAppIntent returned an unexpected body: got %v; "+" result %v", appIntent, test.result)
			}

			if aiExists != test.exists {
				t.Errorf("CreateAppIntent returned an unexpected status: got %v; "+" result %v", aiExists, test.exists)
			}
		})
	}
}

func TestGetAppIntent(t *testing.T) {
	testCases := []struct {
		appIntent              string
		compositeApp           string
		deploymentIntentGroup  string
		err                    string
		genericPlacementIntent string
		label                  string
		project                string
		version                string
		db                     *db.MockDB
		result                 AppIntent
	}{
		{
			label:                  "Get Intent",
			appIntent:              "testAppIntent",
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			genericPlacementIntent: "testGenericPlacementIntent",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			result: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: "A sample AppIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: SpecData{
					AppName: "testApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us2",
									},
								},
							},
						},
					},
				},
			},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						AppIntentKey{
							Name:                      "testAppIntent",
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							Version:                   "testCompositeAppVersion",
							Intent:                    "testGenericPlacementIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent\"," +
									"\"Description\":\"A sample AppIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"testApp\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
					},
				},
			},
		},
		{
			label:                  "Intent Not Found",
			appIntent:              "nonExistingAppIntent",
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			genericPlacementIntent: "testGenericPlacementIntent",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			err:                    "Intent not found",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						AppIntentKey{
							Name:                      "testAppIntent",
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							Version:                   "testCompositeAppVersion",
							Intent:                    "testGenericPlacementIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent\"," +
									"\"Description\":\"A sample AppIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"testApp\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = test.db
			cli := NewAppIntentClient()
			appIntent, err := cli.GetAppIntent(ctx, test.appIntent, test.project, test.compositeApp, test.version, test.genericPlacementIntent, test.deploymentIntentGroup)
			if err != nil {
				if test.err == "" {
					t.Fatalf("GetAppIntent returned an unexpected error: %s", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("GetAppIntent returned an unexpected error: %s", err.Error())
				}
			}

			if reflect.DeepEqual(test.result, appIntent) == false {
				t.Errorf("GetAppIntent returned an unexpected body: got %v; result %v", appIntent, test.result)
			}
		})
	}
}

func TestGetAllIntentsByApp(t *testing.T) {
	testCases := []struct {
		appName                string
		compositeApp           string
		deploymentIntentGroup  string
		err                    string
		genericPlacementIntent string
		label                  string
		project                string
		version                string
		db                     *db.MockDB
		result                 SpecData
	}{
		{
			label:                  "Get All Intents By App",
			appName:                "testApp",
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			genericPlacementIntent: "testGenericPlacementIntent",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			result: SpecData{
				AppName: "testApp",
				Intent: gpic.IntentStruc{
					AllOfArray: []gpic.AllOf{
						{
							ProviderName: "aws",
							ClusterName:  "edge1",
						},
						{
							ProviderName: "aws",
							ClusterName:  "edge2",
						},
						{
							AnyOfArray: []gpic.AnyOf{
								{
									ProviderName:     "aws",
									ClusterLabelName: "east-us1",
								},
								{
									ProviderName:     "aws",
									ClusterLabelName: "east-us2",
								},
							},
						},
					},
				},
			},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						AppIntentFindByAppKey{
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							CompositeAppVersion:       "testCompositeAppVersion",
							Intent:                    "testGenericPlacementIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
							AppName:                   "testApp",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent\"," +
									"\"Description\":\"A sample AppIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"testApp\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
					},
					{
						AppIntentFindByAppKey{
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							CompositeAppVersion:       "testCompositeAppVersion",
							Intent:                    "testGenericPlacementIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
							AppName:                   "testApp_1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent\"," +
									"\"Description\":\"A sample AppIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"testApp\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
					},
				},
			},
		},

		{
			label:                  "Get All Intents By Non Existing App",
			appName:                "nonExistingApp",
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			genericPlacementIntent: "testGenericPlacementIntent",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						AppIntentFindByAppKey{
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							CompositeAppVersion:       "testCompositeAppVersion",
							Intent:                    "testGenericPlacementIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
							AppName:                   "testApp",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent\"," +
									"\"Description\":\"A sample AppIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"testApp\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
					},
					{
						AppIntentFindByAppKey{
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							CompositeAppVersion:       "testCompositeAppVersion",
							Intent:                    "testGenericPlacementIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
							AppName:                   "testApp_1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent\"," +
									"\"Description\":\"A sample AppIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"testApp\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = test.db
			cli := NewAppIntentClient()
			specData, err := cli.GetAllIntentsByApp(ctx, test.appName, test.project, test.compositeApp, test.version, test.genericPlacementIntent, test.deploymentIntentGroup)
			if err != nil {
				if test.err == "" {
					t.Fatalf("GetAllIntentsByApp returned an unexpected error: %s", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("GetAllIntentsByApp returned an unexpected error: %s", err.Error())
				}
			}

			if reflect.DeepEqual(test.result, specData) == false {
				t.Errorf("GetAllIntentsByApp returned an unexpected body: got %v;  result %v", specData, test.result)
			}
		})
	}
}

func TestGetAllAppIntents(t *testing.T) {
	testCases := []struct {
		compositeApp           string
		deploymentIntentGroup  string
		err                    string
		genericPlacementIntent string
		label                  string
		project                string
		version                string
		db                     *db.MockDB
		result                 []AppIntent
	}{
		{
			label:                  "Get All AppIntents",
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			genericPlacementIntent: "testGenericPlacementIntent",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			result: []AppIntent{
				{
					MetaData: MetaData{
						Name:        "testAppIntent_1",
						Description: "Description of testAppIntent_1",
						UserData1:   "userData1",
						UserData2:   "userData2",
					},
					Spec: SpecData{
						AppName: "testApp_1",
						Intent: gpic.IntentStruc{
							AllOfArray: []gpic.AllOf{
								{
									ProviderName: "aws",
									ClusterName:  "edge1",
								},
								{
									ProviderName: "aws",
									ClusterName:  "edge2",
								},
								{
									AnyOfArray: []gpic.AnyOf{
										{
											ProviderName:     "aws",
											ClusterLabelName: "east-us1",
										},
										{
											ProviderName:     "aws",
											ClusterLabelName: "east-us2",
										},
									},
								},
							},
						},
					},
				},
				{
					MetaData: MetaData{
						Name:        "testAppIntent_2",
						Description: "Description of testAppIntent_2",
						UserData1:   "userData1",
						UserData2:   "userData2",
					},
					Spec: SpecData{
						AppName: "testApp_2",
						Intent: gpic.IntentStruc{
							AllOfArray: []gpic.AllOf{
								{
									ProviderName: "aws",
									ClusterName:  "edge1",
								},
								{
									ProviderName: "aws",
									ClusterName:  "edge2",
								},
								{
									AnyOfArray: []gpic.AnyOf{
										{
											ProviderName:     "aws",
											ClusterLabelName: "east-us1",
										},
										{
											ProviderName:     "aws",
											ClusterLabelName: "east-us2",
										},
									},
								},
							},
						},
					},
				},
			},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						AppIntentKey{
							Name:                      "",
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							Version:                   "testCompositeAppVersion",
							Intent:                    "testGenericPlacementIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent_1\"," +
									"\"Description\":\"Description of testAppIntent_1\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"testApp_1\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
					},
					{
						AppIntentKey{
							Name:                      "",
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							Version:                   "testCompositeAppVersion",
							Intent:                    "testGenericPlacementIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent_2\"," +
									"\"Description\":\"Description of testAppIntent_2\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"testApp_2\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
					},
				},
			},
		},
		{
			label:                  "Get All AppIntents Not Exists",
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			genericPlacementIntent: "testGenericPlacementIntent",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = test.db
			cli := NewAppIntentClient()
			appIntents, err := cli.GetAllAppIntents(ctx, test.project, test.compositeApp, test.version, test.genericPlacementIntent, test.deploymentIntentGroup)
			if err != nil {
				if test.err == "" {
					t.Fatalf("GetAllAppIntents returned an unexpected error: %s", err)
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("GetAllAppIntents returned an unexpected error: %s", err)
				}
			}

			if reflect.DeepEqual(test.result, appIntents) == false {
				t.Errorf("GetAllAppIntents returned an unexpected body: got %v; result %v", appIntents, test.result)
			}
		})
	}
}

func TestDeleteAppIntent(t *testing.T) {
	testCases := []struct {
		appIntent              string
		compositeApp           string
		deploymentIntentGroup  string
		err                    string
		genericPlacementIntent string
		label                  string
		project                string
		version                string
		db                     *db.MockDB
	}{
		{
			label:                  "Delete AppIntent",
			appIntent:              "testAppIntent",
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			genericPlacementIntent: "testGenericPlacementIntent",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						AppIntentKey{
							Name:                      "testAppIntent",
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							Version:                   "testCompositeAppVersion",
							Intent:                    "testGenericPlacementIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent\"," +
									"\"Description\":\"testAppIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"testApp\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
					},
				},
			},
		},
		{
			label:                  "Delete Non Existing AppIntent",
			appIntent:              "nonExistingAppIntent",
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			genericPlacementIntent: "testGenericPlacementIntent",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			err:                    "db Remove resource not found",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						AppIntentKey{
							Name:                      "testAppIntent",
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							Version:                   "testCompositeAppVersion",
							Intent:                    "testGenericPlacementIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent\"," +
									"\"Description\":\"testAppIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"testApp\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = test.db
			cli := NewAppIntentClient()
			err := cli.DeleteAppIntent(ctx, test.appIntent, test.project, test.compositeApp, test.version, test.genericPlacementIntent, test.deploymentIntentGroup)
			if err != nil {
				if test.err == "" {
					t.Fatalf("GetAppIntent returned an unexpected error: %s", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("GetAppIntent returned an unexpected error: %s", err.Error())
				}
			}
		})
	}
}
