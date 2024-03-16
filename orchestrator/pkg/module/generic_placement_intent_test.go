// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"reflect"
	"strings"
	"testing"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

func TestCreateGenericPlacementIntent(t *testing.T) {
	testCases := []struct {
		compositeApp          string
		deploymentIntentGroup string
		err                   string
		label                 string
		project               string
		version               string
		db                    *db.MockDB
		gpi                   GenericPlacementIntent
		result                GenericPlacementIntent
	}{
		{
			label: "Create GenericPlacementIntent",
			gpi: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testGenericPlacementIntent",
					Description: "A sample GenericPlacementIntent for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			result: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testGenericPlacementIntent",
					Description: "A sample GenericPlacementIntent for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
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
			label: "GenericPlacementIntent Already Exists",
			gpi: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testGenericPlacementIntent",
					Description: "A sample GenericPlacementIntent for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			result:                GenericPlacementIntent{},
			err:                   "Intent already exists",
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
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacementIntent\"," +
									"\"Description\":\"A sample GenericPlacementIntent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}" +
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
			cli := NewGenericPlacementIntentClient()
			genericPlacementIntent, _, err := cli.CreateGenericPlacementIntent(ctx, test.gpi, test.project, test.compositeApp, test.version, test.deploymentIntentGroup, true)
			if err != nil {
				if test.err == "" {
					t.Fatalf("CreateGenericPlacementIntent returned an unexpected error %s", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("CreateGenericPlacementIntent returned an unexpected error %s", err.Error())
				}
			}

			if reflect.DeepEqual(test.result, genericPlacementIntent) == false {
				t.Errorf("CreateGenericPlacementIntent returned an unexpected body: got %v; "+" expected %v", genericPlacementIntent, test.result)
			}
		})

	}
}

func TestGetGenericPlacementIntent(t *testing.T) {
	testCases := []struct {
		compositeApp           string
		deploymentIntentGroup  string
		err                    string
		genericPlacementIntent string
		label                  string
		project                string
		version                string
		db                     *db.MockDB
		result                 GenericPlacementIntent
	}{
		{
			label:                  "Get GenericPlacementIntent",
			genericPlacementIntent: "testGenericPlacementIntent",
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			result: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testGenericPlacementIntent",
					Description: "A sample GenericPlacementIntent for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacementIntent\"," +
									"\"Description\":\"A sample GenericPlacementIntent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}" +
									"}"),
						},
					},
					{
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent_1",
							Project:      "testProject_1",
							CompositeApp: "testCompositeApp_1",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup_1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacementIntent_1\"," +
									"\"Description\":\"A sample GenericPlacementIntent for testing\"," +
									"\"UserData1\": \"userData1_1\"," +
									"\"UserData2\": \"userData2_1\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:                  "GenericPlacementIntent Not Found",
			genericPlacementIntent: "nonExistingGenericPlacementIntent",
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			result:                 GenericPlacementIntent{},
			err:                    "Intent not found",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacementIntent\"," +
									"\"Description\":\"A sample GenericPlacementIntent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}" +
									"}"),
						},
					},
					{
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent_1",
							Project:      "testProject_1",
							CompositeApp: "testCompositeApp_1",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup_1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacementIntent_1\"," +
									"\"Description\":\"A sample GenericPlacementIntent for testing\"," +
									"\"UserData1\": \"userData1_1\"," +
									"\"UserData2\": \"userData2_1\"}" +
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
			cli := NewGenericPlacementIntentClient()
			genericPlacementIntent, err := cli.GetGenericPlacementIntent(ctx, test.genericPlacementIntent, test.project, test.compositeApp, test.version, test.deploymentIntentGroup)
			if err != nil {
				if test.err == "" {
					t.Fatalf("GetGenericPlacementIntent returned an unexpected error: %s", err)
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("GetGenericPlacementIntent returned an unexpected error: %s", err)
				}
			}

			if reflect.DeepEqual(test.result, genericPlacementIntent) == false {
				t.Errorf("GetGenericPlacementIntent returned an unexpected body: got %v; expected %v", genericPlacementIntent, test.result)
			}
		})
	}
}

func TestDeleteGenericPlacementIntent(t *testing.T) {
	testCases := []struct {
		compositeApp           string
		deploymentIntentGroup  string
		err                    string
		genericPlacementIntent string
		label                  string
		project                string
		version                string
		db                     *db.MockDB
		result                 GenericPlacementIntent
	}{
		{
			label:                  "Delete GenericPlacementIntent",
			genericPlacementIntent: "testGenericPlacementIntent",
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			result:                 GenericPlacementIntent{},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacementIntent\"," +
									"\"Description\":\"A sample GenericPlacementIntent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}" +
									"}"),
						},
					},
					{
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent_1",
							Project:      "testProject_1",
							CompositeApp: "testCompositeApp_1",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup_1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacementIntent_1\"," +
									"\"Description\":\"A sample GenericPlacementIntent for testing\"," +
									"\"UserData1\": \"userData1_1\"," +
									"\"UserData2\": \"userData2_1\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:                  "Delete Non Existing GenericPlacementIntent",
			genericPlacementIntent: "nonExistingGenericPlacementIntent",
			project:                "testProject",
			compositeApp:           "testCompositeApp",
			version:                "testCompositeAppVersion",
			deploymentIntentGroup:  "testDeploymentIntentGroup",
			result:                 GenericPlacementIntent{},
			err:                    "db Remove resource not found",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacementIntent\"," +
									"\"Description\":\"A sample GenericPlacementIntent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}" +
									"}"),
						},
					},
					{
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent_1",
							Project:      "testProject_1",
							CompositeApp: "testCompositeApp_1",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup_1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacementIntent_1\"," +
									"\"Description\":\"A sample GenericPlacementIntent for testing\"," +
									"\"UserData1\": \"userData1_1\"," +
									"\"UserData2\": \"userData2_1\"}" +
									"}"),
						},
					},
				},
				Err: pkgerrors.New("db Remove resource not found"),
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = test.db
			cli := NewGenericPlacementIntentClient()
			err := cli.DeleteGenericPlacementIntent(ctx, test.genericPlacementIntent, test.project, test.compositeApp, test.version, test.deploymentIntentGroup)
			if err != nil {
				if test.err == "" {
					t.Fatalf("DeleteGenericPlacementIntent returned an unexpected error: %s", err.Error())
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("DeleteGenericPlacementIntent returned an unexpected error: %s", err.Error())
				}
			}
		})
	}
}

func TestGetAllGenericPlacementIntents(t *testing.T) {
	testCases := []struct {
		compositeApp          string
		deploymentIntentGroup string
		err                   string
		label                 string
		project               string
		version               string
		db                    *db.MockDB
		result                []GenericPlacementIntent
	}{

		{
			label:                 "Get All GenericPlacementIntents Not Exist",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent",
							Project:      "testProject_2",
							CompositeApp: "testCompositeApp_2",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup_2",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacementIntent\"," +
									"\"Description\":\"A sample GenericPlacementIntent for testing\"," +
									"\"UserData1\": \"userData1_2\"," +
									"\"UserData2\": \"userData2_2\"}" +
									"}"),
						},
					},
					{
						GenericPlacementIntentKey{
							Name:         "testGenericPlacementIntent_1",
							Project:      "testProject_1",
							CompositeApp: "testCompositeApp_1",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup_1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacementIntent_1\"," +
									"\"Description\":\"A sample GenericPlacementIntent for testing\"," +
									"\"UserData1\": \"userData1_1\"," +
									"\"UserData2\": \"userData2_1\"}" +
									"}"),
						},
					},
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project-name\":\"testProject\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
					},
					{
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
					},
				},
			},
		},
		{
			label:                 "Get All GenericPlacementIntents",
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			result: []GenericPlacementIntent{
				{
					MetaData: GenIntentMetaData{
						Name:        "testGenericPlacementIntent",
						Description: "A sample GenericPlacementIntent for testing",
						UserData1:   "userData1",
						UserData2:   "userData2",
					},
				},
			},
			db: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						GenericPlacementIntentKey{
							Name:         "",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacementIntent\"," +
									"\"Description\":\"A sample GenericPlacementIntent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}" +
									"}"),
						},
					},
					{
						GenericPlacementIntentKey{
							Name:         "",
							Project:      "testProject_1",
							CompositeApp: "testCompositeApp_1",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup_1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacementIntent_1\"," +
									"\"Description\":\"A sample GenericPlacementIntent for testing\"," +
									"\"UserData1\": \"userData1_1\"," +
									"\"UserData2\": \"userData2_1\"}" +
									"}"),
						},
					},
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project-name\":\"testProject\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
					},
					{
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
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = test.db
			cli := NewGenericPlacementIntentClient()
			genericPlacementIntents, err := cli.GetAllGenericPlacementIntents(ctx, test.project, test.compositeApp, test.version, test.deploymentIntentGroup)
			if err != nil {
				if test.err == "" {
					t.Fatalf("GetAllGenericPlacementIntents returned an unexpected error: %s", err)
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("GetAllGenericPlacementIntents returned an unexpected error: %s", err)
				}

				t.Fatalf("GetAllGenericPlacementIntents failed with an unexpected error %s", err.Error())
			}

			if reflect.DeepEqual(test.result, genericPlacementIntents) == false {
				t.Errorf("GetAllGenericPlacementIntents returned an unexpected body: got %v; expected %v", genericPlacementIntents, test.result)
			}
		})
	}

}

func TestUpdateGenericPlacementIntent(t *testing.T) {
	testCases := []struct {
		compositeApp          string
		deploymentIntentGroup string
		err                   string
		label                 string
		project               string
		version               string
		db                    *db.MockDB
		exists                bool
		gpi                   GenericPlacementIntent
		result                GenericPlacementIntent
	}{
		{
			label: "Update Non Existing GenericPlacementIntent",
			gpi: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "newGenericPlacement",
					Description: "A sample GenericPlacementIntent for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			exists:                false,
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			result: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "newGenericPlacement",
					Description: "A sample GenericPlacementIntent for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
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
			label: "Update Existing GenericPlacementIntent",
			gpi: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testGenericPlacement",
					Description: "This is an existing GenericPlacementIntent for testing",
					UserData1:   "userData1_1",
					UserData2:   "userData2_1",
				},
			},
			exists:                true,
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			version:               "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			result: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testGenericPlacement",
					Description: "This is an existing GenericPlacementIntent for testing",
					UserData1:   "userData1_1",
					UserData2:   "userData2_1",
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
						GenericPlacementIntentKey{
							Name:         "testGenericPlacement",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacement\"," +
									"\"Description\":\"A sample GenericPlacementIntent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}" +
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
			cli := NewGenericPlacementIntentClient()
			genericPlacementIntent, gpiExists, err := cli.CreateGenericPlacementIntent(ctx, test.gpi, test.project, test.compositeApp, test.version, test.deploymentIntentGroup, false)
			if err != nil {
				if test.err == "" {
					t.Fatalf("CreateGenericPlacementIntent returned an unexpected error %s", err)
				}

				if strings.Contains(err.Error(), test.err) == false {
					t.Fatalf("CreateGenericPlacementIntent returned an unexpected error %s", err)
				}

				t.Fatalf("CreateGenericPlacementIntent failed with an unexpected error %s", err.Error())
			}

			if reflect.DeepEqual(test.result, genericPlacementIntent) == false {
				t.Errorf("CreateGenericPlacementIntent returned an unexpected body: got %v; "+" expected %v", genericPlacementIntent, test.result)
			}

			if test.exists != gpiExists {
				t.Errorf("CreateGenericPlacementIntent returned an unexpected status: got %v; "+" expected %v", gpiExists, test.exists)
			}

		})
	}
}
