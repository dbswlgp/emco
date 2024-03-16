// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
	hpaModel "gitlab.com/project-emco/core/emco-base/src/hpa-plc/pkg/model"
	orchMod "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

func TestCreateIntent(t *testing.T) {
	testCases := []struct {
		label         string
		inp           hpaModel.DeploymentHpaIntent
		expectedError string
		mockdb        *db.MockDB
		expected      hpaModel.DeploymentHpaIntent
	}{
		{
			label: "Create Intent",
			inp: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "A sample Intent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			expected: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "A sample Intent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"data": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
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
			label: "Create Existing Intent",
			inp: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "A sample Intent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			expectedError: "Intent already exists",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"data": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
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
						HpaIntentKey{IntentName: "testIntent",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"data": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"testIntent\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label: "Failed Create Intent due to DB error",
			inp: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "A sample Intent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			expected: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "A sample Intent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			expectedError: "Error",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"data": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
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
				Err: pkgerrors.New("Error"),
			},
		},
	}

	fmt.Printf("\n================== TestCreateIntent .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			ctx := context.Background()
			fmt.Printf("\n================== TestCreateIntent .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, err := impl.AddIntent(ctx, testCase.inp, "project1", "compositeapp1", "version1", "dgroup1", false)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected-error[%s] expected[%s]", err, testCase.expectedError)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Create returned an unexpected-error[%s] expected[%s]", err, testCase.expectedError)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Create returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestUpdateIntent(t *testing.T) {
	testCases := []struct {
		label         string
		inp           hpaModel.DeploymentHpaIntent
		expectedError string
		mockdb        *db.MockDB
		expected      hpaModel.DeploymentHpaIntent
	}{
		{
			label: "Update Intent",
			inp: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "Test Intent for unit testing",
					UserData1:   "update userData1",
					UserData2:   "update userData2",
				},
			},
			expected: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "Test Intent for unit testing",
					UserData1:   "update userData1",
					UserData2:   "update userData2",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"data": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
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
						HpaIntentKey{IntentName: "testIntent",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"data": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"testIntent\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
	}

	fmt.Printf("\n================== TestUpdateIntent .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			ctx := context.Background()
			fmt.Printf("\n================== TestUpdateIntent .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, err := impl.AddIntent(ctx, testCase.inp, "project1", "compositeapp1", "version1", "dgroup1", true)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Update returned an unexpected error [%s]", err)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Update returned an unexpected-error[%s] expected[%s]", err, testCase.expectedError)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Update returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetAllIntents(t *testing.T) {

	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      []hpaModel.DeploymentHpaIntent
	}{
		{
			label: "GetAll Intents",
			expected: []hpaModel.DeploymentHpaIntent{
				{
					MetaData: mtypes.Metadata{
						Name:        "testIntent",
						Description: "Test Intent for unit testing",
						UserData1:   "userData1",
						UserData2:   "userData2",
					},
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"data": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
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
						HpaIntentKey{IntentName: "",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"data": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"testIntent\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:         "GetAll Intents with no-intents",
			expected:      []hpaModel.DeploymentHpaIntent{{}},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"data": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
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
						HpaIntentKey{IntentName: "",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"data": []byte(
								"{}"),
						},
					},
				},
			},
		},
		{
			label:         "GetAll Error",
			expectedError: "Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	fmt.Printf("\n================== TestGetAllIntents .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestGetAllIntents .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, err := impl.GetAllIntents(ctx, "project1", "compositeapp1", "version1", "dgroup1")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected error: %s", err)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Get returned an unexpected error: [%s] expectedError[%s]", err, testCase.expectedError)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Get returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetAllIntentsByApp(t *testing.T) {

	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      []hpaModel.DeploymentHpaIntent
	}{
		{
			label: "GetAllByApp Intents",
			expected: []hpaModel.DeploymentHpaIntent{
				{
					MetaData: mtypes.Metadata{
						Name:        "testIntent",
						Description: "Test Intent for unit testing",
						UserData1:   "userData1",
						UserData2:   "userData2",
					},
					Spec: hpaModel.DeploymentHpaIntentSpec{
						AppName: "testApp1",
					},
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"data": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
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
						HpaIntentKey{IntentName: "testIntent",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"data": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"testIntent\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
						HpaIntentKey{IntentName: "",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"data": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"testIntent\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}," +
									"\"spec\" : {" +
									"\"app\":\"testApp1\"}" +
									"}",
							),
						},
					},
				},
			},
		},
	}

	fmt.Printf("\n================== TestGetAllIntentsByApp .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestGetAllIntentsByApp .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, err := impl.GetAllIntentsByApp(ctx, "testApp1", "project1", "compositeapp1", "version1", "dgroup1")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected error: %s", err)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Get returned an unexpected error: [%s] expectedError[%s]", err, testCase.expectedError)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Get returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}
func TestGetIntent(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      hpaModel.DeploymentHpaIntent
	}{
		{
			label: "Get Intent",
			name:  "testIntent",
			expected: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "Test Intent for unit testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"data": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
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
						HpaIntentKey{IntentName: "testIntent",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"data": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"testIntent\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:         "Get Error",
			expectedError: "Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	fmt.Printf("\n================== TestGetIntent .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestGetIntent .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, _, err := impl.GetIntent(ctx, testCase.name, "project1", "compositeapp1", "version1", "dgroup1")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected error: %s", err)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Get returned an unexpected error: [%s] expectedError[%s]", err, testCase.expectedError)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Get returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetIntentByName(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      hpaModel.DeploymentHpaIntent
	}{
		{
			label: "GetIntentByName Intent",
			name:  "testIntent",
			expected: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "Test Intent for unit testing",
					UserData1:   "user data",
					UserData2:   "user data",
				},
				Spec: hpaModel.DeploymentHpaIntentSpec{AppName: "app1"},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"data": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
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
						HpaIntentKey{IntentName: "testIntent",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"testIntent\"," +
									"\"description\":\"Test Intent for unit testing\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"app\":\"app1\"}}"),
						},
					},
				},
			},
		},
		{
			label:         "GetIntentByName Error",
			expectedError: "Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	fmt.Printf("\n================== GetIntentByName .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== GetIntentByName .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, err := impl.GetIntentByName(ctx, testCase.name, "project1", "compositeapp1", "version1", "dgroup1")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected error: %s", err)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Get returned an unexpected error: [%s] expectedError[%s]", err, testCase.expectedError)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Get returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}
func TestDeleteIntent(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedCode  int
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label:         "Delete Intent",
			expectedError: "",
			name:          "testIntent",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"data": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
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
						HpaIntentKey{IntentName: "testIntent",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"data": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"testIntent\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:         "Delete Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
		{
			label:         "Delete Non-Existing Project",
			expectedError: "DB Error",
			name:          "testIntent",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"data": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
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
						HpaIntentKey{IntentName: "testIntent",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"data": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"testIntent\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
	}

	fmt.Printf("\n================== TestDeleteIntent .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestDeleteIntent .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			err := impl.DeleteIntent(ctx, testCase.name, "project1", "compositeapp1", "version1", "dgroup1")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Delete returned an unexpected-error[%s] expected[%s]", err, testCase.expectedError)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Delete returned an unexpected-error[%s] expected[%s]", err, testCase.expectedError)
				}
			}
		})
	}
}
