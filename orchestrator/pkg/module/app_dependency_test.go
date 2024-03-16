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

func TestCreateAppDependency(t *testing.T) {
	testCases := []struct {
		label                    string
		appDependency            AppDependency
		inputProject             string
		inputCompositeApp        string
		inputCompositeAppVersion string
		app                      string
		expectedError            string
		mockdb                   db.MockDB
	}{
		{
			label: "Test Create AppDependency",
			appDependency: AppDependency{
				MetaData: AdMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "DescriptionTestDeploymentIntentGroup",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: AdSpecData{
					AppName:  "Testprofile",
					OpStatus: "version of deployment",
					Wait:     10,
				},
			},
			inputProject:             "testProject",
			inputCompositeApp:        "testCompositeApp",
			inputCompositeAppVersion: "testCompositeAppVersion",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = &testCase.mockdb
			depIntentCli := NewAppDependencyClient()
			got, err := depIntentCli.CreateAppDependency(ctx, testCase.appDependency, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.app, false)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateDeploymentIntentGroup returned an unexpected error %s, ", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateDeploymentIntentGroup returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.appDependency, got) == false {
					t.Errorf("CreateDeploymentIntentGroup returned unexpected body: got %v; "+" expected %v", got, testCase.appDependency)
				}
			}
		})
	}
}

func TestAppDependency(t *testing.T) {
	appDependency := make(map[string][]AppDependency)

	testCases := []struct {
		label string
		//appDependency			 map[string][]AppDependency
		inputProject             string
		inputCompositeApp        string
		inputCompositeAppVersion string
		app1                     string
		app2                     string
		app3                     string
		expectedError            string
		mockdb                   db.NewMockDB
	}{
		{
			label:                    "Test Cyclic AppDependency",
			inputProject:             "testProject",
			inputCompositeApp:        "testCompositeApp",
			inputCompositeAppVersion: "testCompositeAppVersion",
		},
	}
	appDependency["app1"] = []AppDependency{
		{MetaData: AdMetaData{Name: "test1"}, Spec: AdSpecData{AppName: "app2", OpStatus: "Ready", Wait: 10}},
		{MetaData: AdMetaData{Name: "test2"}, Spec: AdSpecData{AppName: "app3", OpStatus: "Ready", Wait: 10}},
	}
	appDependency["app2"] = []AppDependency{
		{MetaData: AdMetaData{Name: "test1"}, Spec: AdSpecData{AppName: "app1", OpStatus: "Ready", Wait: 10}},
		{MetaData: AdMetaData{Name: "test2"}, Spec: AdSpecData{AppName: "app4", OpStatus: "Ready", Wait: 10}},
	}
	appDependency["app3"] = []AppDependency{}
	appDependency["app4"] = []AppDependency{}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			ctx := context.Background()
			db.DBconn = &testCase.mockdb
			depIntentCli := NewAppDependencyClient()
			var allApps []App
			for app, dl := range appDependency {
				a := App{
					Metadata: AppMetaData{Name: app},
				}
				allApps = append(allApps, a)
				for _, dep := range dl {
					_, err := depIntentCli.CreateAppDependency(ctx, dep, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, app, false)
					if err != nil {
						if testCase.expectedError == "" {
							t.Fatalf("CreateDeploymentIntentGroup returned an unexpected error %s, ", err)
						}
					}
				}
			}
			b := checkDependency(ctx, allApps, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion)
			t.Log(b)
		})
	}
}
