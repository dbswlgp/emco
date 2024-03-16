package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"gitlab.com/project-emco/core/emco-base/src/dcm/api/mocks"
	orch_mocks "gitlab.com/project-emco/core/emco-base/src/orchestrator/api/mocks"
	module "gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
)

func init() {
	logicalCloudJSONValidation = "../json-schemas/logical-cloud.json"
}

var _ = Describe("LogicalCloudHandler", func() {
	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     module.LogicalCloud
		mockError    error
		mockVal      module.LogicalCloud
		mockVals     []module.LogicalCloud
		expectedCode int
		lcClient     *mocks.LogicalCloudManager
		clClient     *mocks.ClusterManager
		upClient     *mocks.UserPermissionManager
		quotaClient  *mocks.QuotaManager
		kvClient     *mocks.KeyValueManager
		prClient     *orch_mocks.ProjectManager
		lcStatus     status.LogicalCloudStatus
		// stateInfo    state.StateInfo
	}

	DescribeTable("Create LogicalCloud tests",
		func(t testCase) {
			// mockedProject := orch.Project{
			// 	MetaData: orch.ProjectMetaData{
			// 		Name:        "test-project",
			// 		Description: "description",
			// 		UserData1:   "some user data 1",
			// 		UserData2:   "some user data 2",
			// 	},
			// }
			// t.prClient.On("GetProject", mock.Anything, "test-project").Return(mockedProject, nil)
			// set up client mock responses
			t.lcClient.On("Create", mock.Anything, "test-project", t.inStruct).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/logical-clouds", t.inputReader)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.LogicalCloud{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		// TODO uncomment after mocking GetProject()
		// Entry("successful create", testCase{
		// 	expectedCode: http.StatusCreated,
		// 	inputReader: bytes.NewBuffer([]byte(`{
		// 	"metadata": {
		// 		"name": "testlogicalcloud",
		// 		"description": "description",
		// 		"userData1": "some user data 1",
		// 		"userData2": "some user data 2"
		// 	}
		// }`)),
		// 	inStruct: module.LogicalCloud{
		// 		MetaData: types.Metadata{
		// 			Name: "testlogicalcloud",
		// 			Description:      "description",
		// 			UserData1:        "some user data 1",
		// 			UserData2:        "some user data 2",
		// 		},
		// 	},
		// 	mockError: nil,
		// 	mockVal: module.LogicalCloud{
		// 		MetaData: types.Metadata{
		// 			Name: "testlogicalcloud",
		// 			Description:      "description",
		// 			UserData1:        "some user data 1",
		// 			UserData2:        "some user data 2",
		// 		},
		// 	},
		// 	lcClient: &mocks.LogicalCloudManager{},
		// }),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.LogicalCloud{},
			mockError:    nil,
			mockVal:      module.LogicalCloud{},
			lcClient:     &mocks.LogicalCloudManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "description",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
		}`)),
			inStruct:  module.LogicalCloud{},
			mockError: nil,
			lcClient:  &mocks.LogicalCloudManager{},
		}),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to other json validation error", testCase{
		// 	// name field has an '=' character
		// 	expectedCode: http.StatusBadRequest,
		// 	inputReader: bytes.NewBuffer([]byte(`{
		// 	"metadata": {
		// 		"name": "logical-cloud",
		// 		"description": "description",
		// 		"userData1": "some user data 1",
		// 		"userData2": "some user data 2"
		// 	}
		// }`)),
		// 	inStruct:    module.LogicalCloud{},
		// 	mockError:   nil,
		// 	lcClient:&mocks.LogicalCloudManager{},
		// }),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			expectedCode: http.StatusUnprocessableEntity,
			inputReader: bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": "testlogicalcloud",
				"description": "description",
				"userData1": "some user data 1",
				"userData2": "some user data 2",
			}
		}`)),
			inStruct:  module.LogicalCloud{},
			mockError: nil,
			lcClient:  &mocks.LogicalCloudManager{},
		}),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to entry already exists", testCase{
		// 	expectedCode: http.StatusConflict,
		// 	inputReader: bytes.NewBuffer([]byte(`{
		// 	"metadata": {
		// 		"name": "testlogicalcloud",
		// 		"description": "description",
		// 		"userData1": "some user data 1",
		// 		"userData2": "some user data 2"
		// 	}
		// }`)),
		// 	inStruct: module.LogicalCloud{
		// 		MetaData: types.Metadata{
		// 			Name:   "testlogicalcloud",
		// 			Description: "description",
		// 			UserData1:   "some user data 1",
		// 			UserData2:   "some user data 2",
		// 		},
		// 	},
		// 	mockVal:     module.LogicalCloud{},
		// 	mockError:   pkgerrors.New("LogicalCloud already exists"),
		// 	lcClient:&mocks.LogicalCloudManager{},
		// }),

		// TODO uncomment after mocking GetProject()
		// Entry("fails due to db error", testCase{
		// 	expectedCode: http.StatusInternalServerError,
		// 	inputReader: bytes.NewBuffer([]byte(`{
		// 	"metadata": {
		// 		"name": "testlogicalcloud",
		// 		"description": "description",
		// 		"userData1": "some user data 1",
		// 		"userData2": "some user data 2"
		// 	}
		// }`)),
		// 	inStruct: module.LogicalCloud{
		// 		MetaData: types.Metadata{
		// 			Name: "testlogicalcloud",
		// 			Description:      "description",
		// 			UserData1:        "some user data 1",
		// 			UserData2:        "some user data 2",
		// 		},
		// 	},
		// 	mockVal:   module.LogicalCloud{},
		// 	mockError: pkgerrors.New("Creating DB Entry"),
		// 	lcClient:  &mocks.LogicalCloudManager{},
		// }),
	)

	// DCM PUT API currently disabled, so all tests commented out
	// DescribeTable("Put LogicalCloud tests",
	// 	func(t testCase) {
	// 		// set up client mock responses
	// 		t.lcClient.On("Update", mock.Anything, "test-project", t.inputName, t.inStruct).Return(t.mockVal, t.mockError)

	// 		// make HTTP request
	// 		request := httptest.NewRequest("PUT", "/v2/projects/test-project/logical-clouds/"+t.inputName, t.inputReader)
	// 		resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

	// 		//Check returned code
	// 		Expect(resp.StatusCode).To(Equal(t.expectedCode))

	// 		//Check returned body
	// 		got := module.LogicalCloud{}
	// 		json.NewDecoder(resp.Body).Decode(&got)
	// 		Expect(got).To(Equal(t.mockVal))
	// 	},

	// 	Entry("successful put", testCase{
	// 		expectedCode: http.StatusOK, // TODO: change to StatusCreated?
	// 		inputName:    "logicalcloud",
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"name": "logicalcloud",
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2"
	// 		},
	// 		"spec" : {
	// 			"limits.cpu": "500",
	// 			"limits.memory": "2000Gi"
	// 		}
	// 	}`)),
	// 		inStruct: module.LogicalCloud{
	// 			MetaData: types.Metadata{
	// 				Name:   "logicalcloud",
	// 				Description: "description",
	// 				UserData1:   "some user data 1",
	// 				UserData2:   "some user data 2",
	// 			},
	// 			Specification: map[string]string{
	// 				"limits.cpu":    "500",
	// 				"limits.memory": "2000Gi",
	// 			},
	// 		},
	// 		mockError: nil,
	// 		mockVal: module.LogicalCloud{
	// 			MetaData: types.Metadata{
	// 				Name:   "logicalcloud",
	// 				Description: "description",
	// 				UserData1:   "some user data 1",
	// 				UserData2:   "some user data 2",
	// 			},
	// 			Specification: map[string]string{
	// 				"limits.cpu":    "400",
	// 				"limits.memory": "1000Gi",
	// 			},
	// 		},
	// 		lcClient:&mocks.LogicalCloudManager{},
	// 	}),

	// 	Entry("fails due to empty body", testCase{
	// 		inputName:    "logicalcloud",
	// 		expectedCode: http.StatusBadRequest,
	// 		inStruct:     module.LogicalCloud{},
	// 		mockError:    nil,
	// 		mockVal:      module.LogicalCloud{},
	// 		lcClient: &mocks.LogicalCloudManager{},
	// 	}),

	// 	Entry("fails due missing name", testCase{
	// 		inputName:    "logicalcloud",
	// 		expectedCode: http.StatusBadRequest,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2"
	// 		}
	// 	}`)),
	// 		inStruct:    module.LogicalCloud{},
	// 		mockError:   nil,
	// 		lcClient:&mocks.LogicalCloudManager{},
	// 	}),

	// 	// TODO: implement logic and then enable this test:
	// 	// Entry("fails due to other json validation error", testCase{
	// 	// 	// name field in body has an '=' character
	// 	// 	inputName:    "logicalcloud",
	// 	// 	expectedCode: http.StatusBadRequest,
	// 	// 	inputReader: bytes.NewBuffer([]byte(`{
	// 	// 	"metadata": {
	// 	// 		"name": "test=logicalcloud",
	// 	// 		"description": "description",
	// 	// 		"userData1": "some user data 1",
	// 	// 		"userData2": "some user data 2"
	// 	// 	}
	// 	// }`)),
	// 	// 	inStruct:    module.LogicalCloud{},
	// 	// 	mockError:   nil,
	// 	// 	lcClient:&mocks.LogicalCloudManager{},
	// 	// }),

	// 	Entry("fails due to json body decoding error", testCase{
	// 		// extra comma at the end of the userData2 line
	// 		inputName:    "logicalcloud",
	// 		expectedCode: http.StatusUnprocessableEntity,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"name": "logicalcloud",
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2",
	// 		}
	// 	}`)),
	// 		inStruct:    module.LogicalCloud{},
	// 		mockError:   nil,
	// 		lcClient:&mocks.LogicalCloudManager{},
	// 	}),

	// 	// TODO: implement logic and then enable this test:
	// 	// Entry("fails due to mismatched name", testCase{
	// 	// 	inputName:    "quotaXYZ",
	// 	// 	expectedCode: http.StatusBadRequest,
	// 	// 	inputReader: bytes.NewBuffer([]byte(`{
	// 	// 	"metadata": {
	// 	// 		"name": "logicalcloud",
	// 	// 		"description": "description",
	// 	// 		"userData1": "some user data 1",
	// 	// 		"userData2": "some user data 2"
	// 	// 	}
	// 	// }`)),
	// 	// 	inStruct: module.LogicalCloud{
	// 	// 		MetaData: types.Metadata{
	// 	// 			Name:   "logicalcloud",
	// 	// 			Description: "description",
	// 	// 			UserData1:   "some user data 1",
	// 	// 			UserData2:   "some user data 2",
	// 	// 		},
	// 	// 	},
	// 	// 	mockVal:     module.LogicalCloud{},
	// 	// 	mockError:   pkgerrors.New("Creating DB Entry"),
	// 	// 	lcClient:&mocks.LogicalCloudManager{},
	// 	// }),

	// 	Entry("fails due to db error", testCase{
	// 		inputName:    "logicalcloud",
	// 		expectedCode: http.StatusInternalServerError,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"name": "logicalcloud",
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2"
	// 		}
	// 	}`)),
	// 		inStruct: module.LogicalCloud{
	// 			MetaData: types.Metadata{
	// 				Name:   "logicalcloud",
	// 				Description: "description",
	// 				UserData1:   "some user data 1",
	// 				UserData2:   "some user data 2",
	// 			},
	// 		},
	// 		mockVal:     module.LogicalCloud{},
	// 		mockError:   pkgerrors.New("Creating DB Entry"),
	// 		lcClient:&mocks.LogicalCloudManager{},
	// 	}),
	// )

	DescribeTable("List LogicalCloud tests",
		func(t testCase) {
			// set up client mock responses
			t.lcClient.On("GetAll", mock.Anything, "test-project").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/logical-clouds", nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := []module.LogicalCloud{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},

		Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVals: []module.LogicalCloud{
				{
					MetaData: types.Metadata{
						Name:        "testlogicalcloud1",
						Description: "description",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
				},
				{
					MetaData: types.Metadata{
						Name:        "testlogicalcloud2",
						Description: "description",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
				},
			},
			lcClient: &mocks.LogicalCloudManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVals:     []module.LogicalCloud{},
			lcClient:     &mocks.LogicalCloudManager{},
		}),
	)

	DescribeTable("Get LogicalCloud tests",
		func(t testCase) {
			// set up client mock responses
			t.lcClient.On("Get", mock.Anything, "test-project", t.inputName).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/logical-clouds/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.LogicalCloud{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful get", testCase{
			inputName:    "testlogicalcloud",
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVal: module.LogicalCloud{
				MetaData: types.Metadata{
					Name:        "testlogicalcloud",
					Description: "description",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			lcClient: &mocks.LogicalCloudManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testlogicalcloud",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("Logical Cloud not found"),
			mockVal:      module.LogicalCloud{},
			lcClient:     &mocks.LogicalCloudManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			inputName:    "testlogicalcloud",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVal:      module.LogicalCloud{},
			lcClient:     &mocks.LogicalCloudManager{},
		}),
	)

	DescribeTable("Delete LogicalCloud tests",
		func(t testCase) {
			// set up client mock responses
			t.lcClient.On("Delete", mock.Anything, "test-project", t.inputName).Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/logical-clouds/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.LogicalCloud{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful delete", testCase{
			inputName:    "testlogicalcloud",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			lcClient:     &mocks.LogicalCloudManager{},
		}),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to not found", testCase{
		// 	inputName:    "testlogicalcloud",
		// 	expectedCode: http.StatusNotFound,
		// 	mockError:    pkgerrors.New("db Remove error - not found"),
		// 	lcClient: &mocks.LogicalCloudManager{},
		// }),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to a conflict", testCase{
		// 	inputName:    "testlogicalcloud",
		// 	expectedCode: http.StatusConflict,
		// 	mockError:    pkgerrors.New("db Remove error - conflict"),
		// 	lcClient:      &mocks.LogicalCloudManager{},
		// }),

		Entry("fails due to other backend error", testCase{
			inputName:    "testlogicalcloud",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error"),
			lcClient:     &mocks.LogicalCloudManager{},
		}),
	)

	// TODO add testing for instantiate and terminate
	// TODO add additional mocking for cluster client:
	// DescribeTable("Instantiate Logical Cloud (L1)",
	// 	func(t testCase) {
	// 		mockedClusters := []module.Cluster{}
	// 		mockedClusters = append(mockedClusters, module.Cluster{})
	// 		mockedClusters[0].MetaData = module.ClusterMeta{
	// 			ClusterReference: "cluster1",
	// 			Description:      "description",
	// 			UserData1:        "some user data 1",
	// 			UserData2:        "some user data 2",
	// 		}
	// 		mockedQuotas := []module.Quota{}
	// 		mockedQuotas = append(mockedQuotas, module.Quota{})
	// 		mockedQuotas[0].MetaData = module.QMetaDataList{
	// 			QuotaName:   "cluster1",
	// 			Description: "description",
	// 			UserData1:   "some user data 1",
	// 			UserData2:   "some user data 2",
	// 		}
	// 		// set up client mock responses
	// 		t.lcClient.On("Get", mock.Anything, "test-project", "testlogicalcloud").Return(t.mockVal, t.mockError)
	// 		t.clClient.On("GetAllClusters", mock.Anything, "test-project", "testlogicalcloud").Return(mockedClusters, nil)
	// 		t.quotaClient.On("GetAllQuotas", mock.Anything, "test-project", "testlogicalcloud").Return(mockedQuotas, nil)
	// 		t.lcClient.On("Instantiate", mock.Anything, "test-project", "testlogicalcloud", mockedClusters, mockedQuotas).Return(t.mockVal, t.mockError)

	// 		// make HTTP request
	// 		request := httptest.NewRequest("POST", "/v2/projects/test-project/logical-clouds/"+t.inputName+"/instantiate", nil)
	// 		resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

	// 		// Check returned code
	// 		Expect(resp.StatusCode).To(Equal(t.expectedCode))

	// 		// Check returned body
	// 		got := module.LogicalCloud{}
	// 		json.NewDecoder(resp.Body).Decode(&got)
	// 		Expect(got).To(Equal(t.mockVal))
	// 	},

	// 	Entry("instantiate", testCase{
	// 		expectedCode: http.StatusCreated,
	// 		inputName:    "testlogicalcloud",
	// 		inputReader:  bytes.NewBuffer([]byte(``)),
	// 		inStruct:     module.LogicalCloud{},
	// 		mockError:    nil,
	// 		lcClient:     &mocks.LogicalCloudManager{},
	// 		mockVal: module.LogicalCloud{
	// 			MetaData: types.Metadata{
	// 				Name: "testlogicalcloud",
	// 				Description:      "description",
	// 				UserData1:        "some user data 1",
	// 				UserData2:        "some user data 2",
	// 			},
	// 		},
	// 	}),
	// )

	DescribeTable("Get State of LogicalCloud tests",
		func(t testCase) {
			// set up client mock responses
			t.lcClient.On("Get", mock.Anything, "test-project", "testlogicalcloud").Return(t.mockVal, t.mockError)
			t.lcClient.On("Status", mock.Anything, "test-project", "testlogicalcloud", "", "ready", "all", []string{}, []string{}).Return(t.lcStatus, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/logical-clouds/"+t.inputName+"/status", nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := status.LogicalCloudStatus{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.lcStatus))
		},

		Entry("successful get", testCase{
			inputName:    "testlogicalcloud",
			expectedCode: http.StatusOK,
			mockError:    nil,
			lcStatus: status.LogicalCloudStatus{
				Project:      "test-project",
				LogicalCloud: "testlogicalcloud",
				StatusResult: status.StatusResult{"logical-cloud", state.StateInfo{}, "", "", "", nil, nil, nil, nil, nil, nil},
				// StatusContextId: "",
				// Actions:         nil,
			},
			lcClient: &mocks.LogicalCloudManager{},
		}),
	)
})
