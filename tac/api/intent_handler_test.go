// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation
package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
	. "gitlab.com/project-emco/core/emco-base/src/tac/api"
	"gitlab.com/project-emco/core/emco-base/src/tac/api/mocks"
	"gitlab.com/project-emco/core/emco-base/src/tac/pkg/model"
	tmpl "gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/emcotemporalapi"
	wfMod "gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/module"
	history "go.temporal.io/api/history/v1"
	wfsvc "go.temporal.io/api/workflowservice/v1"
	cl "go.temporal.io/sdk/client"
)

type testCase struct {
	/* Universal */
	inputReader  io.Reader
	expectedCode int
	mockError    error
	client       *mocks.WorkflowIntentManager

	/* Workflow Intent Hook */
	inStructHook model.WorkflowHookIntent
	mockVals     []model.WorkflowHookIntent
	mockVal      model.WorkflowHookIntent

	/* Cancel and get status */
	inStructCancel model.WfhTemporalCancelRequest
	inStructStatus wfMod.WfTemporalStatusQuery
	mockStatus     wfMod.WfTemporalStatusResponse
}

func init() {
	CrJSONFile = "json-schemas/cancel_request.json"
	TacIntentJSONFile = "../json-schemas/tac_intent.json"
	SqJSONFile = "../json-schemas/temporal_status_query.json"

}

var _ = Describe("HookIntentHandlers", func() {

	/* Workflow Hook Intent */

	DescribeTable("Workflow Intent Create",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateWorkflowHookIntent", mock.Anything, t.inStructHook, "test-project", "test-compositeapp", "v1", "test-dig", false).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/temporal-action-controller", t.inputReader)
			res := executeRequest(request, NewRouter(t.client))

			// check the code
			Expect(res.StatusCode).To(Equal(t.expectedCode))

			if http.StatusCreated == res.StatusCode {
				got := model.WorkflowHookIntent{}
				json.NewDecoder(res.Body).Decode(&got)
				Expect(got).To(Equal(t.mockVal))
			}

		},

		Entry("successfully create new workflow", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testworkflowintent"
				},
				"spec": {
					"hookType": "pre-install",
					"hookBlocking": true,
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
			inStructHook: model.WorkflowHookIntent{
				Metadata: mtypes.Metadata{
					Name: "testworkflowintent",
				},
				Spec: model.WorkflowHookSpec{
					HookType: "pre-install",
					WfClientSpec: wfMod.WfClientSpec{
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
			mockVal: model.WorkflowHookIntent{
				Metadata: mtypes.Metadata{
					Name: "testworkflowintent",
				},
				Spec: model.WorkflowHookSpec{
					HookType: "pre-install",
					WfClientSpec: wfMod.WfClientSpec{
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

		Entry("Missing required fields", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": ""
				},
				"spec": {
					"hookType": "pre-install",
					"hookBlocking": true,
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
			inStructHook: model.WorkflowHookIntent{
				Metadata: mtypes.Metadata{
					Name: "",
				},
				Spec: model.WorkflowHookSpec{
					HookType: "pre-install",
					WfClientSpec: wfMod.WfClientSpec{
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
			mockVal: model.WorkflowHookIntent{
				Metadata: mtypes.Metadata{
					Name: "",
				},
				Spec: model.WorkflowHookSpec{
					HookType: "pre-install",
					WfClientSpec: wfMod.WfClientSpec{
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
	)

	DescribeTable("Get All WorkflowIntentHooks",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetWorkflowHookIntents", mock.Anything, "test-project", "test-compositeapp", "v1", "test-dig").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/temporal-action-controller", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

		},

		Entry("Get handlers", testCase{
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVals:     []model.WorkflowHookIntent{},
			mockVal:      model.WorkflowHookIntent{},
			client:       &mocks.WorkflowIntentManager{},
		}),

		Entry("Fail to get handlers", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError:    errors.New("some error"),
			mockVals:     []model.WorkflowHookIntent{},
			mockVal:      model.WorkflowHookIntent{},
			client:       &mocks.WorkflowIntentManager{},
		}),
	)

	DescribeTable("Get One WorkflowIntentHooks",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetWorkflowHookIntents", mock.Anything, "test-project", "test-compositeapp", "v1", "test-dig").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/temporal-action-controller", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

		},

		Entry("Get handlers", testCase{
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVals:     []model.WorkflowHookIntent{},
			mockVal:      model.WorkflowHookIntent{},
			client:       &mocks.WorkflowIntentManager{},
		}),
	)

	DescribeTable("Delete action intent hook",
		func(t testCase) {
			// set up client mock responses
			t.client.On("DeleteWorkflowHookIntent", mock.Anything, "test-hook", "test-project", "test-compositeapp", "v1", "test-dig").Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/temporal-action-controller/test-hook", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

		},

		Entry("Delete hook success", testCase{
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			mockVals:     []model.WorkflowHookIntent{},
			mockVal:      model.WorkflowHookIntent{},
			client:       &mocks.WorkflowIntentManager{},
		}),
	)

	DescribeTable("Workflow Intent Update",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateWorkflowHookIntent", mock.Anything, t.inStructHook, "test-project", "test-compositeapp", "v1", "test-dig", true).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("PUT", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/temporal-action-controller/test-hook", t.inputReader)
			res := executeRequest(request, NewRouter(t.client))

			// check the code
			Expect(res.StatusCode).To(Equal(t.expectedCode))

			if http.StatusCreated == res.StatusCode {
				got := model.WorkflowHookIntent{}
				json.NewDecoder(res.Body).Decode(&got)
				Expect(got).To(Equal(t.mockVal))
			}

		},

		Entry("Succsefully update workflow", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testworkflowintent"
				},
				"spec": {
					"hookType": "pre-install",
					"hookBlocking": true,
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
			inStructHook: model.WorkflowHookIntent{
				Metadata: mtypes.Metadata{
					Name: "testworkflowintent",
				},
				Spec: model.WorkflowHookSpec{
					HookType: "pre-install",
					WfClientSpec: wfMod.WfClientSpec{
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
			mockVal: model.WorkflowHookIntent{
				Metadata: mtypes.Metadata{
					Name: "testworkflowintent",
				},
				Spec: model.WorkflowHookSpec{
					HookType: "pre-install",
					WfClientSpec: wfMod.WfClientSpec{
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

		Entry("Missing required fields", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": ""
				},
				"spec": {
					"hookType": "pre-install",
					"hookBlocking": true,
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
			inStructHook: model.WorkflowHookIntent{
				Metadata: mtypes.Metadata{
					Name: "",
				},
				Spec: model.WorkflowHookSpec{
					HookType: "pre-install",
					WfClientSpec: wfMod.WfClientSpec{
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
			mockVal: model.WorkflowHookIntent{
				Metadata: mtypes.Metadata{
					Name: "",
				},
				Spec: model.WorkflowHookSpec{
					HookType: "pre-install",
					WfClientSpec: wfMod.WfClientSpec{
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
	)

	DescribeTable("Cancel a workflow", func(t testCase) {
		// set up client mock responses
		t.client.On("CancelWorkflowIntent", mock.Anything, "test-hook", "test-project", "test-compositeapp", "v1", "test-dig", &t.inStructCancel).Return(t.mockError)

		// make HTTP request
		request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/temporal-action-controller/test-hook/cancel", t.inputReader)
		resp := executeRequest(request, NewRouter(t.client))

		//Check returned code
		Expect(resp.StatusCode).To(Equal(t.expectedCode))

	},

		Entry("Cancel workflow bad request", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata":{
					"name":"testname"
				},
				"spec":{
					"temporalServer":"testservers",
					"workflowID":"81YaroA4U5PDEBe01b3r1-0M-RGYc94Wvb1A4jKeAp0omaOg8hdg8Bc",
					"runID":"",
					"terminate":true,
					"reason":"",
					"details": []
				}
			}`)),
			mockError: nil,
			inStructCancel: model.WfhTemporalCancelRequest{
				Metadata: mtypes.Metadata{
					Name: "testname",
				},
				Spec: model.WfhTemporalCancelRequestSpec{
					TemporalServer: "testservers",
					WfID:           "81YaroA4U5PDEBe01b3r1-0M-RGYc94Wvb1A4jKeAp0omaOg8hdg8Bc",
					RunID:          "",
					Terminate:      true,
					Details:        []interface{}{},
				}},
			client: &mocks.WorkflowIntentManager{},
		}),
		Entry("Empty Body Fail", testCase{
			expectedCode:   http.StatusBadRequest,
			mockError:      nil,
			inStructCancel: model.WfhTemporalCancelRequest{},
			client:         &mocks.WorkflowIntentManager{},
		}),
	)

	//GetStatusWorkflowIntent(name string, project string, cApp string, cAppVer string, dig string, query *pkgmodule.WfTemporalStatusQuery) (*pkgmodule.WfTemporalStatusResponse, error)
	DescribeTable("Status of a workflow", func(t testCase) {
		// set up client mock responses
		t.client.On("GetStatusWorkflowIntent", mock.Anything, "test-hook", "test-project", "test-compositeapp", "v1", "test-dig", &t.inStructStatus).Return(&t.mockStatus, t.mockError)

		// make HTTP request
		request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/temporal-action-controller/test-hook/status", t.inputReader)
		resp := executeRequest(request, NewRouter(t.client))

		//Check returned code
		Expect(resp.StatusCode).To(Equal(t.expectedCode))

	},

		Entry("Succesful Get Status", testCase{
			/* Universal */
			expectedCode: http.StatusOK,
			inputReader: bytes.NewBuffer([]byte(`{
				"temporalServer":    "testserver",
				"workflowID":              "123abc",
				"runID":             "123abc",
				"waitForResult":     true,
				"runDescribeWfExec": false,
				"getWfHistory":      false,
				"queryType":         "",
				"queryParams":       []
			}`)),
			mockError: nil,
			client:    &mocks.WorkflowIntentManager{},
			/* Route Specific */
			inStructStatus: wfMod.WfTemporalStatusQuery{
				TemporalServer:    "testserver",
				WfID:              "123abc",
				RunID:             "123abc",
				WaitForResult:     true,
				RunDescribeWfExec: false,
				GetWfHistory:      false,
				QueryType:         "",
				QueryParams:       []interface{}{},
			},
			mockStatus: wfMod.WfTemporalStatusResponse{
				WfID:          "123abc",
				RunID:         "123abc",
				WfExecDesc:    wfsvc.DescribeWorkflowExecutionResponse{},
				WfHistory:     []history.HistoryEvent{},
				WfResult:      []interface{}{},
				WfQueryResult: []interface{}{},
			},
		}),

		Entry("Empty Body Fail", testCase{
			/* Universal */
			expectedCode: http.StatusBadRequest,
			mockError:    nil,
			client:       &mocks.WorkflowIntentManager{},
			/* Route Specific */
			inStructStatus: wfMod.WfTemporalStatusQuery{

				WfID:              "123abc",
				RunID:             "123abc",
				WaitForResult:     true,
				RunDescribeWfExec: false,
				GetWfHistory:      false,
				QueryType:         "",
				QueryParams:       []interface{}{},
			},
			mockStatus: wfMod.WfTemporalStatusResponse{
				WfID:          "123abc",
				RunID:         "123abc",
				WfExecDesc:    wfsvc.DescribeWorkflowExecutionResponse{},
				WfHistory:     []history.HistoryEvent{},
				WfResult:      []interface{}{},
				WfQueryResult: []interface{}{},
			},
		}),
		Entry("Failed to get Status", testCase{
			/* Universal */
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
				"temporalServer":    "testserver",
				"workflowID":              "123abc",
				"runID":             "123abc",
				"waitForResult":     true,
				"runDescribeWfExec": false,
				"getWfHistory":      false,
				"queryType":         "",
				"queryParams":       []
			}`)),
			mockError: errors.New("Could not get status"),
			client:    &mocks.WorkflowIntentManager{},
			/* Route Specific */
			inStructStatus: wfMod.WfTemporalStatusQuery{
				TemporalServer:    "testserver",
				WfID:              "123abc",
				RunID:             "123abc",
				WaitForResult:     true,
				RunDescribeWfExec: false,
				GetWfHistory:      false,
				QueryType:         "",
				QueryParams:       []interface{}{},
			},
			mockStatus: wfMod.WfTemporalStatusResponse{
				WfID:          "123abc",
				RunID:         "123abc",
				WfExecDesc:    wfsvc.DescribeWorkflowExecutionResponse{},
				WfHistory:     []history.HistoryEvent{},
				WfResult:      []interface{}{},
				WfQueryResult: []interface{}{},
			},
		}),
	)

})
