// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
)

func TestCreateDeploymentIntentGroup(t *testing.T) {
	testCases := []struct {
		compositeApp string
		err          string
		label        string
		project      string
		version      string
		db           *db.MockDB
		dig          DeploymentIntentGroup
		result       DeploymentIntentGroup
	}{
		{
			label: "Create DeploymentIntentGroup",
			dig: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "DescriptionTestDeploymentIntentGroup",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
					},
					LogicalCloud: "cloud1",
				},
			},
			project:      "testProject",
			compositeApp: "testCompositeApp",
			version:      "testCompositeAppVersion",
			result: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "DescriptionTestDeploymentIntentGroup",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
					},
					LogicalCloud: "cloud1",
				},
			},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project-name\":\"testProject\"," +
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
									"\"compositeAppVersion\":\"version of the composite app\"}}"),
						},
					},
				},
			},
		},
		{
			label: "DeploymentIntentGroup Already Exists",
			dig: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "DescriptionTestDeploymentIntentGroup",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
					},
					LogicalCloud: "cloud1",
				},
			},
			project:      "testProject",
			compositeApp: "testCompositeApp",
			version:      "testCompositeAppVersion",
			result:       DeploymentIntentGroup{},
			err:          "Intent already exists",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project-name\":\"testProject\"," +
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
									"\"compositeAppVersion\":\"version of the composite app\"}}"),
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
									"\"spec\":{\"compositeProfile\": \"Testprofile\"," +
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
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = test.db
			cli := NewDeploymentIntentGroupClient()
			deploymentIntentGroup, _, err := cli.CreateDeploymentIntentGroup(ctx, test.dig, test.project, test.compositeApp, test.version, true)
			if err != nil {
				if test.err == "" {
					t.Fatalf("CreateDeploymentIntentGroup returned an unexpected error %s, ", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("CreateDeploymentIntentGroup returned an unexpected error %s", err.Error())
				}
			}

			if reflect.DeepEqual(test.result, deploymentIntentGroup) == false {
				t.Errorf("CreateDeploymentIntentGroup returned an unexpected body: got %v; "+" expected %v", deploymentIntentGroup, test.result)
			}

		})
	}
}

func TestGetDeploymentIntentGroup(t *testing.T) {
	testCases := []struct {
		compositeApp          string
		deploymentIntentGroup string
		err                   string
		label                 string
		project               string
		version               string
		db                    *db.MockDB
		result                DeploymentIntentGroup
	}{
		{
			label:                 "Get DeploymentIntentGroup",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			result: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "DescriptionTestDeploymentIntentGroup",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
					},
					LogicalCloud: "cloud1",
				},
			},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
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
									"\"spec\":{\"compositeProfile\": \"Testprofile\"," +
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
						DeploymentIntentGroupKey{
							Name:         "testDeploymentIntentGroup_1",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"testDeploymentIntentGroup_1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"compositeProfile\": \"Testprofile\"," +
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
			label:                 "DeploymentIntentGroup Not Found",
			deploymentIntentGroup: "nonExistingDeploymentIntentGroup",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			result:                DeploymentIntentGroup{},
			err:                   "DeploymentIntentGroup not found",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
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
									"\"spec\":{\"compositeProfile\": \"Testprofile\"," +
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
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = test.db
			cli := NewDeploymentIntentGroupClient()
			deploymentIntentGroup, err := cli.GetDeploymentIntentGroup(ctx, test.deploymentIntentGroup, test.project, test.compositeApp, test.version)
			if err != nil {
				if test.err == "" {
					t.Fatalf("GetDeploymentIntentGroup returned an unexpected error: %s", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("GetDeploymentIntentGroup returned an unexpected error: %s", err.Error())
				}
			}

			if reflect.DeepEqual(test.result, deploymentIntentGroup) == false {
				t.Errorf("GetDeploymentIntentGroup returned an unexpected body: got %v; expected %v", deploymentIntentGroup, test.result)
			}

		})
	}
}

func TestGetDeploymentIntentGroupState(t *testing.T) {
	timeStamp, _ := time.Parse(time.RFC3339Nano, "2021-10-15T19:26:06.865+00:00")
	testCases := []struct {
		compositeApp          string
		deploymentIntentGroup string
		err                   string
		label                 string
		project               string
		version               string
		db                    *db.MockDB
		result                state.StateInfo
	}{
		{
			label:                 "Get DeploymentIntentGroup StateInfo",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			result: state.StateInfo{
				StatusContextId: "",
				Actions: []state.ActionEntry{
					{
						State:     state.StateEnum.Approved,
						ContextId: "",
						TimeStamp: timeStamp,
						Revision:  0,
					},
				},
			},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
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
									"\"spec\":{\"compositeProfile\": \"Testprofile\"," +
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
							"stateInfo": []byte(
								"{ \"statusctxid\": \"\"," +
									"\"actions\": [{" +
									"\"state\":\"Approved\"," +
									"\"instance\":\"\"," +
									"\"time\":\"2021-10-15T19:26:06.865+00:00\", " +
									"\"revision\":0" +
									"}]" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:                 "DeploymentIntentGroup StateInfo Not Found",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			err:                   "DeploymentIntentGroup StateInfo not found",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
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
									"\"spec\":{\"compositeProfile\": \"Testprofile\"," +
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
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = test.db
			cli := NewDeploymentIntentGroupClient()
			stateInfo, err := cli.GetDeploymentIntentGroupState(ctx, test.deploymentIntentGroup, test.project, test.compositeApp, test.version)
			if err != nil {
				if test.err == "" {
					t.Fatalf("GetDeploymentIntentGroupState returned an unexpected error: %s", err)
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("GetDeploymentIntentGroupState returned an unexpected error: %s", err)
				}
			}

			if reflect.DeepEqual(test.result, stateInfo) == false {
				t.Errorf("GetDeploymentIntentGroupState returned an unexpected body: got %v; expected %v", stateInfo, test.result)
			}
		})
	}
}

func TestDeleteDeploymentIntentGroup(t *testing.T) {
	testCases := []struct {
		compositeApp          string
		deploymentIntentGroup string
		err                   string
		label                 string
		project               string
		version               string
		db                    *db.MockDB
		result                DeploymentIntentGroup
	}{
		{
			label:                 "Delete DeploymentIntentGroup",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			result:                DeploymentIntentGroup{},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
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
									"\"spec\":{\"compositeProfile\": \"Testprofile\"," +
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
							"stateInfo": []byte(
								"{ \"statusctxid\": \"\"," +
									"\"actions\": [{" +
									"\"state\":\"Created\"," +
									"\"instance\":\"\"," +
									"\"time\":\"2021-10-15T19:26:06.865+00:00\", " +
									"\"revision\":0" +
									"}]" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:                 "Delete Non Existing DeploymentIntentGroup",
			deploymentIntentGroup: "nonExistingDeploymentIntentGroup",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			err:                   "db Remove resource not found",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
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
									"\"spec\":{\"compositeProfile\": \"Testprofile\"," +
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
							"stateInfo": []byte(
								"{ \"statusctxid\": \"\"," +
									"\"actions\": [{" +
									"\"state\":\"Created\"," +
									"\"instance\":\"\"," +
									"\"time\":\"2021-10-15T19:26:06.865+00:00\", " +
									"\"revision\":0" +
									"}]" +
									"}"),
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
			cli := NewDeploymentIntentGroupClient()
			err := cli.DeleteDeploymentIntentGroup(ctx, test.deploymentIntentGroup, test.project, test.compositeApp, test.version)
			if err != nil {
				if test.err == "" {
					t.Fatalf("DeleteDeploymentIntentGroup returned an unexpected error: %s", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("DeleteDeploymentIntentGroup returned an unexpected error: %s", err.Error())
				}
			}
		})
	}
}

func TestGetAllDeploymentIntentGroups(t *testing.T) {
	testCases := []struct {
		compositeApp          string
		deploymentIntentGroup string
		err                   string
		label                 string
		project               string
		version               string
		db                    *db.MockDB
		result                []DeploymentIntentGroup
	}{
		{
			label:        "Get All DeploymentIntentGroups",
			project:      "testProject",
			compositeApp: "testCompositeApp",
			version:      "testCompositeAppVersion",
			result: []DeploymentIntentGroup{
				{
					MetaData: DepMetaData{
						Name:        "testDeploymentIntentGroup",
						Description: "DescriptionTestDeploymentIntentGroup",
						UserData1:   "userData1",
						UserData2:   "userData2",
					},
					Spec: DepSpecData{
						Profile: "Testprofile",
						Version: "version of deployment",
						OverrideValuesObj: []OverrideValues{
							{
								AppName: "TestAppName",
								ValuesObj: map[string]string{
									"imageRepository": "registry.hub.docker.com",
								},
							},
							{
								AppName: "TestAppName",
								ValuesObj: map[string]string{
									"imageRepository": "registry.hub.docker.com",
								},
							},
						},
						LogicalCloud: "cloud1",
					},
				},
			},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project-name\":\"testProject\"," +
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
									"\"compositeAppVersion\":\"version of the composite app\"}}"),
						},
						DeploymentIntentGroupKey{
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"testDeploymentIntentGroup\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"compositeProfile\": \"Testprofile\"," +
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
			label:        "Get All DeploymentIntentGroups Not Exists",
			project:      "testProject",
			compositeApp: "testCompositeApp",
			version:      "testCompositeAppVersion",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project-name\":\"testProject\"," +
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
									"\"compositeAppVersion\":\"version of the composite app\"}}"),
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
			cli := NewDeploymentIntentGroupClient()
			deploymentIntentGroups, err := cli.GetAllDeploymentIntentGroups(ctx, test.project, test.compositeApp, test.version)
			if err != nil {
				if test.err == "" {
					t.Fatalf("GetAllDeploymentIntentGroups returned an unexpected error: %s", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("GetAllDeploymentIntentGroups returned an unexpected error: %s", err.Error())
				}
			}

			if reflect.DeepEqual(test.result, deploymentIntentGroups) == false {
				t.Errorf("GetAllDeploymentIntentGroups returned an unexpected body: got %v; expected %v", deploymentIntentGroups, test.result)
			}
		})
	}
}

func TestUpdateDeploymentIntentGroup(t *testing.T) {
	testCases := []struct {
		compositeApp          string
		deploymentIntentGroup string
		err                   string
		label                 string
		project               string
		compositeAppVersion   string
		db                    *db.MockDB
		exists                bool
		dig                   DeploymentIntentGroup
		result                DeploymentIntentGroup
	}{
		{
			label:  "Update Non Existing DeploymentIntentGroup",
			exists: false,
			dig: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "newDeploymentIntentGroup",
					Description: "DescriptionTestDeploymentIntentGroup",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
					},
					LogicalCloud: "cloud1",
				},
			},
			project:             "testProject",
			compositeApp:        "testCompositeApp",
			compositeAppVersion: "testCompositeAppVersion",
			result: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "newDeploymentIntentGroup",
					Description: "DescriptionTestDeploymentIntentGroup",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
					},
					LogicalCloud: "cloud1",
				},
			},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project-name\":\"testProject\"," +
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
									"\"compositeAppVersion\":\"version of the composite app\"}}"),
						},
					},
				},
			},
		},
		{
			label:  "Update Existing DeploymentIntentGroup",
			exists: true,
			dig: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "This is a new DeploymentIntentGroup for testing",
					UserData1:   "This is a new User Data 1 for testing",
					UserData2:   "This is a new User Data 2 for testing",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
					},
					LogicalCloud: "cloud1",
				},
			},
			project:             "testProject",
			compositeApp:        "testCompositeApp",
			compositeAppVersion: "testCompositeAppVersion",
			result: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "This is a new DeploymentIntentGroup for testing",
					UserData1:   "This is a new User Data 1 for testing",
					UserData2:   "This is a new User Data 2 for testing",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
					},
					LogicalCloud: "cloud1",
				},
			},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"name\":\"testProject\"," +
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
									"\"compositeAppVersion\":\"version of the composite app\"}}"),
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
									"\"spec\":{\"compositeProfile\": \"Testprofile\"," +
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
							"stateInfo": []byte(
								"{ \"statusctxid\": \"\"," +
									"\"actions\": [{" +
									"\"state\":\"Created\"," +
									"\"instance\":\"\"," +
									"\"time\":\"2021-10-15T19:26:06.865+00:00\", " +
									"\"revision\":0" +
									"}]" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:  "Update Existing DeploymentIntentGroup with Approved State",
			exists: true,
			dig: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "This is a new DeploymentIntentGroup for testing",
					UserData1:   "This is a new User Data 1 for testing",
					UserData2:   "This is a new User Data 2 for testing",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
						{
							AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							},
						},
					},
					LogicalCloud: "cloud1",
				},
			},
			project:             "testProject",
			compositeApp:        "testCompositeApp",
			compositeAppVersion: "testCompositeAppVersion",
			err:                 "The DeploymentIntentGroup is not updated",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"name\":\"testProject\"," +
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
									"\"compositeAppVersion\":\"version of the composite app\"}}"),
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
									"\"spec\":{\"compositeProfile\": \"Testprofile\"," +
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
							"stateInfo": []byte(
								"{ \"statusctxid\": \"\"," +
									"\"actions\": [{" +
									"\"state\":\"Approved\"," +
									"\"instance\":\"\"," +
									"\"time\":\"2021-10-15T19:26:06.865+00:00\", " +
									"\"revision\":0" +
									"}]" +
									"}"),
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
			cli := NewDeploymentIntentGroupClient()
			deploymentIntentGroup, digExists, err := cli.CreateDeploymentIntentGroup(ctx, test.dig, test.project, test.compositeApp, test.compositeAppVersion, false)
			if err != nil {
				if test.err == "" {
					t.Fatalf("CreateDeploymentIntentGroup returned an unexpected error %s, ", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("CreateDeploymentIntentGroup returned an unexpected error %s", err.Error())
				}
			}

			if reflect.DeepEqual(test.result, deploymentIntentGroup) == false {
				t.Errorf("CreateDeploymentIntentGroup returned an unexpected body: got %v; "+" expected %v", deploymentIntentGroup, test.result)
			}

			if digExists != test.exists {
				t.Errorf("CreateDeploymentIntentGroup returned an unexpected status: got %v; "+" expected %v", digExists, test.exists)
			}

		})
	}
}
