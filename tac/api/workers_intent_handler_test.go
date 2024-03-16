// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation
package api_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	. "github.com/onsi/gomega"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
	. "gitlab.com/project-emco/core/emco-base/src/tac/api"
	"gitlab.com/project-emco/core/emco-base/src/tac/api/mocks"
	"gitlab.com/project-emco/core/emco-base/src/tac/pkg/model"
)

type workerTestCase struct {
	inStructWorker model.WorkerIntent
	client         *mocks.WorkerIntentManager
	inputReader    io.Reader
	expectedCode   int
	mockVal        model.WorkerIntent
	mockVals       []model.WorkerIntent
	mockError      error
}

func init() {
	WorkerIntentJSONFile = "../json-schemas/deploy_worker.json"
}

var _ = Describe("WorkerIntentHandlerTests", func() {
	DescribeTable("WorkerIntentCreate",
		func(t workerTestCase) {
			// set up the client mock response
			t.client.On("CreateOrUpdateWorkerIntent", t.inStructWorker, "test-tac", "test-project", "test-cApp", "v1", "test-dig", false).Return(t.mockVal, t.mockError)

			//make the http request.
			request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-cApp/v1/deployment-intent-groups/test-dig/temporal-action-controller/test-tac/workers", t.inputReader)
			res := executeRequest(request, NewRouter(t.client))

			// check the response code
			Expect(res.StatusCode).To(Equal(t.expectedCode))
		},
		Entry("successfully create new worker", workerTestCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": "worker-1"
			},
			"spec": {
				"startToCloseTimeout": 1000000,
				"deploymentIntentGroup": "test-dig",
				"compositeApp": "test-cApp",
				"compositeAppVersion": "v1"
			}
			}`)),
			client: &mocks.WorkerIntentManager{},
			inStructWorker: model.WorkerIntent{
				Metadata: mtypes.Metadata{
					Name: "worker-1",
				},
				Spec: model.WorkerSpec{
					StartToCloseTimeout: 1000000,
					DIG:                 "test-dig",
					CApp:                "test-cApp",
					CAppVersion:         "v1",
				},
			},
			mockVal: model.WorkerIntent{
				Metadata: mtypes.Metadata{
					Name: "worker-1",
				},
				Spec: model.WorkerSpec{
					StartToCloseTimeout: 1000000,
					DIG:                 "test-dig",
					CApp:                "test-cApp",
					CAppVersion:         "v1",
				},
			},
			mockError: nil,
		}),
		Entry("successfully create new worker", workerTestCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
			"metadata": {
				
			},
			"spec": {
				"startToCloseTimeout": 1000000,
				"deploymentIntentGroup": "test-dig",
				"compositeApp": "test-cApp",
				"compositeAppVersion": "v1"
			}
			}`)),
			client: &mocks.WorkerIntentManager{},
			inStructWorker: model.WorkerIntent{
				Metadata: mtypes.Metadata{},
				Spec: model.WorkerSpec{
					StartToCloseTimeout: 1000000,
					DIG:                 "test-dig",
					CApp:                "test-cApp",
					CAppVersion:         "v1",
				},
			},
			mockVal: model.WorkerIntent{
				Metadata: mtypes.Metadata{},
				Spec: model.WorkerSpec{
					StartToCloseTimeout: 1000000,
					DIG:                 "test-dig",
					CApp:                "test-cApp",
					CAppVersion:         "v1",
				},
			},
			mockError: nil,
		}),
	)
	DescribeTable("Get Specific Worker", func(t workerTestCase) {
		// set up the client mock response.
		t.client.On("GetWorkerIntent", "test-worker", "test-project", "test-cApp", "v1", "test-dig", "test-tac").Return(t.mockVal, t.mockError)

		// make the mock request.
		request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-cApp/v1/deployment-intent-groups/test-dig/temporal-action-controller/test-tac/workers/test-worker", t.inputReader)
		res := executeRequest(request, NewRouter(t.client))

		// Check the response code
		Expect(res.StatusCode).To(Equal(t.expectedCode))

	},
		Entry("Successfully get worker", workerTestCase{
			inStructWorker: model.WorkerIntent{},
			client:         &mocks.WorkerIntentManager{},
			inputReader:    nil,
			expectedCode:   http.StatusOK,
			mockVal:        model.WorkerIntent{},
			mockError:      nil,
		}))
	DescribeTable("Get Multiple Workers", func(t workerTestCase) {
		// set up the client mock response.
		//(project string, cApp string, cAppVer string, dig string, tac string) ([]model.WorkerIntent, error)
		t.client.On("GetWorkerIntents", "test-project", "test-cApp", "v1", "test-dig", "test-tac").Return(t.mockVals, t.mockError)

		// make the mock request.
		request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-cApp/v1/deployment-intent-groups/test-dig/temporal-action-controller/test-tac/workers", t.inputReader)
		res := executeRequest(request, NewRouter(t.client))

		// Check the response code
		Expect(res.StatusCode).To(Equal(t.expectedCode))
	},
		Entry("Successfully get many workers", workerTestCase{
			inStructWorker: model.WorkerIntent{},
			client:         &mocks.WorkerIntentManager{},
			inputReader:    nil,
			expectedCode:   http.StatusOK,
			mockVals:       []model.WorkerIntent{},
			mockError:      nil,
		}))
	DescribeTable("Update Worker", func(t workerTestCase) {
		// set up the client mock response
		t.client.On("CreateOrUpdateWorkerIntent", t.inStructWorker, "test-tac", "test-project", "test-cApp", "v1", "test-dig", true).Return(t.mockVal, t.mockError)

		//make the http request.
		request := httptest.NewRequest("PUT", "/v2/projects/test-project/composite-apps/test-cApp/v1/deployment-intent-groups/test-dig/temporal-action-controller/test-tac/workers/worker-1", t.inputReader)
		res := executeRequest(request, NewRouter(t.client))

		// check the response code
		Expect(res.StatusCode).To(Equal(t.expectedCode))
	},
		Entry("successfully create new worker", workerTestCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
	"metadata": {
		"name": "worker-1"
	},
	"spec": {
		"startToCloseTimeout": 1000000,
		"deploymentIntentGroup": "test-dig",
		"compositeApp": "test-cApp",
		"compositeAppVersion": "v1"
	}
	}`)),
			client: &mocks.WorkerIntentManager{},
			inStructWorker: model.WorkerIntent{
				Metadata: mtypes.Metadata{
					Name: "worker-1",
				},
				Spec: model.WorkerSpec{
					StartToCloseTimeout: 1000000,
					DIG:                 "test-dig",
					CApp:                "test-cApp",
					CAppVersion:         "v1",
				},
			},
			mockVal: model.WorkerIntent{
				Metadata: mtypes.Metadata{
					Name: "worker-1",
				},
				Spec: model.WorkerSpec{
					StartToCloseTimeout: 1000000,
					DIG:                 "test-dig",
					CApp:                "test-cApp",
					CAppVersion:         "v1",
				},
			},
			mockError: nil,
		}))
	DescribeTable("Delete Worker", func(t workerTestCase) {
		// set up the client mock response
		//project string, cApp string, cAppVer string, dig string, tac string, workerName string) error
		t.client.On("DeleteWorkerIntents", "test-project", "test-cApp", "v1", "test-dig", "test-tac", "worker-1").Return(t.mockError)

		//make the http request.
		request := httptest.NewRequest("DELETE", "/v2/projects/test-project/composite-apps/test-cApp/v1/deployment-intent-groups/test-dig/temporal-action-controller/test-tac/workers/worker-1", t.inputReader)
		res := executeRequest(request, NewRouter(t.client))

		// check the response code
		Expect(res.StatusCode).To(Equal(t.expectedCode))
	},
		Entry("Successfully Delete a worker", workerTestCase{
			inStructWorker: model.WorkerIntent{},
			client:         &mocks.WorkerIntentManager{},
			inputReader:    nil,
			expectedCode:   http.StatusNoContent,
			mockVals:       []model.WorkerIntent{},
			mockError:      nil,
		}))
})
