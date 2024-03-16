// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"bytes"
	//"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"gitlab.com/project-emco/core/emco-base/src/workflowmgr/api/mocks"
	tmpl "gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/emcotemporalapi"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/module"
	cl "go.temporal.io/sdk/client"
)

var _ = Describe("workflowIntenthandler", func() {

	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     moduleLib.WorkflowIntent
		mockError    error
		mockVal      moduleLib.WorkflowIntent
		mockVals     []moduleLib.WorkflowIntent
		tf           moduleLib.WfTemporalStatusQuery
		tfResponse   moduleLib.WfTemporalStatusResponse
		expectedCode int
		client       *mocks.WorkflowIntentManager
	}

	DescribeTable("Create workflowIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateWorkflowIntent", mock.Anything, t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", false).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/temporal-workflow-intents", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			//got := moduleLib.WorkflowIntent{}
			//json.NewDecoder(resp.Body).Decode(&got)
			//Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testworkflowintent",
					"description": "test workflow intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"workflowClient": {
					"clientEndpointName": "endpoint1",
					"clientEndpointPort": 9091
					},
					"temporal": {
						"workflowClientName": "client1",
						"workflowStartOptions": {
							"id":  "xyz",
							"taskQueue": "my-task-queue"
						}
					}
				}
			}`)),
			inStruct: moduleLib.WorkflowIntent{
				Metadata: moduleLib.Metadata{
					Name:        "testworkflowintent",
					Description: "test workflow intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: moduleLib.WorkflowIntentSpec{
					WfClientSpec: moduleLib.WfClientSpec{
						WfClientEndpointName: "endpoint1",
						WfClientEndpointPort: 9091,
					},
					WfTemporalSpec: tmpl.WfTemporalSpec{
						WfClientName: "client1",
						WfStartOpts: cl.StartWorkflowOptions{
							ID:        "xyz",
							TaskQueue: "my-task-queue",
						},
					},
				},
			},
			mockError: nil,
			mockVal: moduleLib.WorkflowIntent{
				Metadata: moduleLib.Metadata{
					Name:        "testworkflowintent",
					Description: "test workflow intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: moduleLib.WorkflowIntentSpec{
					WfClientSpec: moduleLib.WfClientSpec{
						WfClientEndpointName: "endpoint1",
						WfClientEndpointPort: 9091,
					},
					WfTemporalSpec: tmpl.WfTemporalSpec{
						WfClientName: "client1",
						WfStartOpts: cl.StartWorkflowOptions{
							ID:        "xyz",
							TaskQueue: "my-task-queue",
						},
					},
				},
			},
			client: &mocks.WorkflowIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     moduleLib.WorkflowIntent{},
			mockError:    nil,
			mockVal:      moduleLib.WorkflowIntent{},
			client:       &mocks.WorkflowIntentManager{},
		}),
	)

	DescribeTable("Workflow intent get all handlers",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetWorkflowIntents", mock.Anything, "test-project", "test-compositeapp", "v1", "test-dig").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/temporal-workflow-intents", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

		},

		Entry("Get handlers", testCase{
			expectedCode: http.StatusOK,
			inStruct:     moduleLib.WorkflowIntent{},
			mockError:    nil,
			mockVal:      moduleLib.WorkflowIntent{},
			mockVals:     []moduleLib.WorkflowIntent{},
			client:       &mocks.WorkflowIntentManager{},
		}),
	)

	DescribeTable("Workflow intent get handler specific entry ",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetWorkflowIntent", mock.Anything, "test-wfi", "test-project", "test-compositeapp", "v1", "test-dig").Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/temporal-workflow-intents/test-wfi", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

		},

		Entry("Get handlers", testCase{
			expectedCode: http.StatusOK,
			inStruct:     moduleLib.WorkflowIntent{},
			mockError:    nil,
			mockVal:      moduleLib.WorkflowIntent{},
			mockVals:     []moduleLib.WorkflowIntent{},
			client:       &mocks.WorkflowIntentManager{},
		}),
	)

	DescribeTable("Delete a workflow intent",
		func(t testCase) {
			// set up client mock responses
			t.client.On("DeleteWorkflowIntent", mock.Anything, "test-wfi", "test-project", "test-compositeapp", "v1", "test-dig").Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/temporal-workflow-intents/test-wfi", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

		},

		Entry("Get handlers", testCase{
			expectedCode: http.StatusNoContent,
			inStruct:     moduleLib.WorkflowIntent{},
			mockError:    nil,
			mockVal:      moduleLib.WorkflowIntent{},
			mockVals:     []moduleLib.WorkflowIntent{},
			client:       &mocks.WorkflowIntentManager{},
		}),
	)

	DescribeTable("Start workflow Intent",
		func(t testCase) {
			// set up client mock responses
			t.client.On("StartWorkflowIntent", mock.Anything, "test-wfi", "test-project", "test-compositeapp", "v1", "test-dig").Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/temporal-workflow-intents/test-wfi/start", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

		},

		Entry("Get handlers", testCase{
			expectedCode: http.StatusNoContent,
			inStruct:     moduleLib.WorkflowIntent{},
			mockError:    nil,
			mockVal:      moduleLib.WorkflowIntent{},
			mockVals:     []moduleLib.WorkflowIntent{},
			client:       &mocks.WorkflowIntentManager{},
		}),
	)

	DescribeTable("Get status of a workflow",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetStatusWorkflowIntent", mock.Anything, "test-wfi", "test-project", "test-compositeapp", "v1", "test-dig", &t.tf).Return(&t.tfResponse, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/temporal-workflow-intents/test-wfi/status", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

		},

		Entry("Get handlers", testCase{
			expectedCode: http.StatusOK,
			inStruct:     moduleLib.WorkflowIntent{},
			mockError:    nil,
			mockVal:      moduleLib.WorkflowIntent{},
			mockVals:     []moduleLib.WorkflowIntent{},
			client:       &mocks.WorkflowIntentManager{},
			tf:           moduleLib.WfTemporalStatusQuery{TemporalServer: "", WfID: "", RunID: "", WaitForResult: false, RunDescribeWfExec: false, GetWfHistory: false, QueryType: "", QueryParams: []interface{}(nil)},
			tfResponse:   moduleLib.WfTemporalStatusResponse{},
		}),
	)

})
