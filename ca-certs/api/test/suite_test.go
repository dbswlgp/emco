// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/api"
)

const (
	// clusterProvider
	clusterProviderCertURL             string = "/v2/cluster-providers/{clusterProvider}/ca-certs"
	clusterProviderCertDistributionURL string = clusterProviderCertURL + "/{caCert}/distribution"
	clusterProviderCertEnrollmentURL   string = clusterProviderCertURL + "/{caCert}/enrollment"
	// logicalCloud
	logicalCloudCertURL             string = "/v2/projects/{project}/ca-certs"
	logicalCloudCertDistributionURL string = logicalCloudCertURL + "/{caCert}/distribution"
	logicalCloudCertEnrollmentURL   string = logicalCloudCertURL + "/{caCert}/enrollment"
)

type test struct {
	entry      string
	name       string
	input      io.Reader
	result     interface{}
	err        error
	statusCode int
	client     interface{} // to implements the client manager interface
}

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Api Suite")
}

func executeRequest(method, target, url string, client interface{}, body io.Reader) *http.Response {
	recorder := httptest.NewRecorder()
	request := newRequest(method, target, url, body)

	router := api.NewRouter(client)
	router.ServeHTTP(recorder, request)

	return recorder.Result()
}

func newRequest(method, target, url string, body io.Reader) *http.Request {
	return httptest.NewRequest(method, url+target, body)
}
