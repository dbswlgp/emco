// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

func TestCreateIntent(t *testing.T) {
	testCases := []struct {
		compositeApp          string
		deploymentIntentGroup string
		err                   string
		label                 string
		project               string
		version               string
		db                    *db.MockDB
		exists                bool
		i                     Intent
		result                Intent
	}{
		{
			label:  "Create Intent",
			exists: false,
			i: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "A sample Intent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
				},
			},
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			result: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "A sample Intent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
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
			label:  "Intent Already Exists",
			exists: true,
			i: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "A sample Intent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
				},
			},
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			err:                   "Intent already exists",
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
						IntentKey{
							Name:                  "testIntent",
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"genericPlacementIntent\": \"testGenericPlacementIntent\"" +
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
			cli := NewIntentClient()
			intent, iExists, err := cli.AddIntent(ctx, test.i, test.project, test.compositeApp, test.version, test.deploymentIntentGroup, true)
			if err != nil {
				if test.err == "" {
					t.Fatalf("AddIntent returned an unexpected error %s, ", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("AddIntent returned an unexpected error %s", err.Error())
				}
			}

			if reflect.DeepEqual(test.result, intent) == false {
				t.Errorf("AddIntent returned an unexpected body: got %v; "+" result %v", intent, test.result)
			}

			if iExists != test.exists {
				t.Errorf("AddIntent returned an unexpected status: got %v; "+" result %v", iExists, test.exists)
			}

		})
	}
}

func TestUpdateIntent(t *testing.T) {
	testCases := []struct {
		compositeApp          string
		deploymentIntentGroup string
		err                   string
		label                 string
		project               string
		version               string
		db                    *db.MockDB
		exists                bool
		i                     Intent
		result                Intent
	}{
		{
			label:  "Update Non Existing Intent",
			exists: false,
			i: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "A sample Intent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
				},
			},
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			result: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "A sample Intent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
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
						IntentKey{
							Name:                  "testIntent_1",
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent_1\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"genericPlacementIntent\": \"testGenericPlacementIntent\"" +
									"}}}"),
						},
					},
				},
			},
		},
		{
			label:  "Update Existing Intent",
			exists: true,
			i: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "This is the new description for testAppIntent",
					UserData1:   "This is the new data for userData1",
					UserData2:   "This is the new data for userData2",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
				},
			},
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			result: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "This is the new description for testAppIntent",
					UserData1:   "This is the new data for userData1",
					UserData2:   "This is the new data for userData2",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
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
						IntentKey{
							Name:                  "testIntent",
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"genericPlacementIntent\": \"testGenericPlacementIntent\"" +
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
			cli := NewIntentClient()
			intent, iExists, err := cli.AddIntent(ctx, test.i, test.project, test.compositeApp, test.version, test.deploymentIntentGroup, false)
			if err != nil {
				if test.err == "" {
					t.Fatalf("AddIntent returned an unexpected error %s, ", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("AddIntent returned an unexpected error %s", err.Error())
				}
			}

			if reflect.DeepEqual(test.result, intent) == false {
				t.Errorf("AddIntent returned an unexpected body: got %v; "+" result %v", intent, test.result)
			}

			if iExists != test.exists {
				t.Errorf("AddIntent returned an unexpected status: got %v; "+" result %v", iExists, test.exists)
			}
		})
	}
}

func TestGetIntent(t *testing.T) {
	testCases := []struct {
		compositeApp          string
		deploymentIntentGroup string
		err                   string
		intent                string
		label                 string
		project               string
		version               string
		db                    *db.MockDB
		result                Intent
	}{
		{
			label:                 "Get Intent",
			intent:                "testIntent",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			result: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "A sample Intent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
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
						IntentKey{
							Name:                  "testIntent",
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"genericPlacementIntent\": \"testGenericPlacementIntent\"" +
									"}}}"),
						},
						IntentKey{
							Name:                  "testIntent_1",
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent_1\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"genericPlacementIntent\": \"testGenericPlacementIntent\"" +
									"}}}"),
						},
					},
				},
			},
		},

		{
			label:                 "Intent Not Found",
			intent:                "nonExistingIntent",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			err:                   "Intent not found",
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
						IntentKey{
							Name:                  "testIntent",
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"genericPlacementIntent\": \"testGenericPlacementIntent\"" +
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
			cli := NewIntentClient()
			intent, err := cli.GetIntent(ctx, test.intent, test.project, test.compositeApp, test.version, test.deploymentIntentGroup)
			if err != nil {
				if test.err == "" {
					t.Fatalf("GetIntent returned an unexpected error %s, ", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("GetIntent returned an unexpected error %s", err.Error())
				}
			}

			if reflect.DeepEqual(test.result, intent) == false {
				t.Errorf("GetIntent returned an unexpected body: got %v; "+" result %v", intent, test.result)
			}
		})
	}
}

func TestGetIntentByName(t *testing.T) {
	testCases := []struct {
		compositeApp          string
		deploymentIntentGroup string
		err                   string
		intent                string
		label                 string
		project               string
		version               string
		db                    *db.MockDB
		result                IntentSpecData
	}{
		{
			label:                 "Get Intent By Name",
			intent:                "testIntent",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			result: IntentSpecData{
				Intent: map[string]string{
					"genericPlacementIntent": "testGenericPlacementIntent",
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
						IntentKey{
							Name:                  "testIntent",
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"genericPlacementIntent\": \"testGenericPlacementIntent\"" +
									"}}}"),
						},
						IntentKey{
							Name:                  "testIntent_1",
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent_1\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"genericPlacementIntent\": \"testGenericPlacementIntent\"" +
									"}}}"),
						},
					},
				},
			},
		},
		{
			label:                 "Intent Not Found",
			intent:                "nonExistingIntent",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			err:                   "Intent not found",
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
						IntentKey{
							Name:                  "testIntent",
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"genericPlacementIntent\": \"testGenericPlacementIntent\"" +
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
			cli := NewIntentClient()
			intentSpecData, err := cli.GetIntentByName(ctx, test.intent, test.project, test.compositeApp, test.version, test.deploymentIntentGroup)
			if err != nil {
				if test.err == "" {
					t.Fatalf("GetIntentByName returned an unexpected error %s, ", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("GetIntentByName returned an unexpected error %s", err.Error())
				}
			}

			if reflect.DeepEqual(test.result, intentSpecData) == false {
				t.Errorf("GetIntentByName returned an unexpected body: got %v; "+" result %v", intentSpecData, test.result)
			}
		})
	}
}

func TestGetAllIntents(t *testing.T) {
	testCases := []struct {
		compositeApp          string
		deploymentIntentGroup string
		err                   string
		label                 string
		project               string
		version               string
		db                    *db.MockDB
		result                ListOfIntents
	}{
		{
			label:                 "Get All Intents",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			result: ListOfIntents{
				ListOfIntents: []map[string]string{
					{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
					{
						"ovnaction": "ovnaction",
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
						IntentKey{
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"genericPlacementIntent\": \"testGenericPlacementIntent\"" +
									"}}}"),
						},
					},
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
						IntentKey{
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent_1\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"ovnaction\": \"ovnaction\"" +
									"}}}"),
						},
					},
				},
			},
		},
		{
			label:                 "Get All Intents Not Exists",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = test.db
			cli := NewIntentClient()
			listOfIntents, err := cli.GetAllIntents(ctx, test.project, test.compositeApp, test.version, test.deploymentIntentGroup)
			if err != nil {
				if test.err == "" {
					t.Fatalf("GetAllIntents returned an unexpected error %s, ", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("GetAllIntents returned an unexpected error %s", err.Error())
				}
			}

			if reflect.DeepEqual(test.result, listOfIntents) == false {
				t.Errorf("GetAllIntents returned an unexpected body: got %v; "+" result %v", listOfIntents, test.result)
			}
		})
	}
}

func TestDeleteIntent(t *testing.T) {
	testCases := []struct {
		compositeApp          string
		deploymentIntentGroup string
		err                   string
		intent                string
		label                 string
		project               string
		version               string
		db                    *db.MockDB
	}{
		{
			label:                 "Delete Intent",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			intent:                "testIntent",
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
						IntentKey{
							Name:                  "testIntent",
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"genericPlacementIntent\": \"testGenericPlacementIntent\"" +
									"}}}"),
						},
					},
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
						IntentKey{
							Name:                  "testIntent_1",
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent_1\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"ovnaction\": \"ovnaction\"" +
									"}}}"),
						},
					},
				},
			},
		},
		{
			label:                 "Delete Non Existing Intent",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			intent:                "nonExistingIntent",
			err:                   "db Remove resource not found",
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
						IntentKey{
							Name:                  "testIntent",
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"genericPlacementIntent\": \"testGenericPlacementIntent\"" +
									"}}}"),
						},
					},
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
						IntentKey{
							Name:                  "testIntent_1",
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent_1\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"ovnaction\": \"ovnaction\"" +
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
			cli := NewIntentClient()
			err := cli.DeleteIntent(ctx, test.intent, test.project, test.compositeApp, test.version, test.deploymentIntentGroup)
			if err != nil {
				if test.err == "" {
					t.Fatalf("DeleteIntent returned an unexpected error %s, ", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("DeleteIntent returned an unexpected error %s", err.Error())
				}
			}
		})
	}
}
