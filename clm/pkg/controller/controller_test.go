// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package controller

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"

	pkgerrors "github.com/pkg/errors"
	clmModel "gitlab.com/project-emco/core/emco-base/src/clm/pkg/model"
)

func TestCreateController(t *testing.T) {
	testCases := []struct {
		label         string
		inp           clmModel.Controller
		expectedError string
		mockdb        *db.MockDB
		expected      clmModel.Controller
	}{
		{
			label: "Create Controller",
			inp: clmModel.Controller{
				Metadata: types.Metadata{
					Name: "testController",
				},
				Spec: clmModel.ControllerSpec{
					Host: "132.156.0.10",
					Port: 8080,
				},
			},
			expected: clmModel.Controller{
				Metadata: types.Metadata{
					Name: "testController",
				},
				Spec: clmModel.ControllerSpec{
					Host: "132.156.0.10",
					Port: 8080,
				},
			},
			expectedError: "",
			mockdb:        &db.MockDB{},
		},
		{
			label:         "Failed Create Controller",
			expectedError: "Error Creating Controller",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("Error Creating Controller"),
			},
		},
	}

	fmt.Printf("\n================== TestCreateController .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			fmt.Printf("\n================== TestCreateController .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
			db.DBconn = testCase.mockdb
			impl := NewControllerClient()
			ctx := context.Background()
			got, err := impl.CreateController(ctx, testCase.inp, false)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
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

func TestGetController(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      clmModel.Controller
	}{
		{
			label:         "Get Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
		{
			label: "Get Controller",
			name:  "testController",
			expected: clmModel.Controller{
				Metadata: types.Metadata{
					Name: "testController",
				},
				Spec: clmModel.ControllerSpec{
					Host: "132.156.0.10",
					Port: 8080,
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						clmModel.ControllerKey{ControllerGroup: "cluster", ControllerName: "testController"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"testController\"" +
									"}," +
									"\"spec\":{" +
									"\"host\":\"132.156.0.10\"," +
									"\"port\": 8080 }}"),
						},
					},
				},
			},
		},
	}

	fmt.Printf("\n================== TestGetController .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestGetController .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewControllerClient()
			ctx := context.Background()
			got, err := impl.GetController(ctx, testCase.name)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected-error[%s] expected[%s]", err, testCase.expectedError)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Get returned an unexpected-error[%s] expected[%s]", err, testCase.expectedError)
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

func TestDeleteController(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label: "Delete Controller",
			name:  "testController",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						clmModel.ControllerKey{ControllerGroup: "cluster", ControllerName: "testController"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"testController\"" +
									"}," +
									"\"spec\":{" +
									"\"host\":\"132.156.0.10\"," +
									"\"port\": 8080 }}"),
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
	}

	fmt.Printf("\n================== TestDeleteController .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestDeleteController .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewControllerClient()
			ctx := context.Background()
			err := impl.DeleteController(ctx, testCase.name)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Delete returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Delete returned an unexpected error %s", err)
				}
			}
		})
	}
}
