// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/api"
)

const (
	baseURL string = "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/generic-k8s-intents"
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

func executeRequest(method, target string, header map[string]string, client interface{}, body io.Reader) *http.Response {
	recorder := httptest.NewRecorder()
	request := newRequest(method, target, body)

	for k, v := range header {
		request.Header.Set(k, v)
	}

	router := api.NewRouter(client)
	router.ServeHTTP(recorder, request)

	return recorder.Result()
}

func newRequest(method, target string, body io.Reader) *http.Request {
	if len(target) == 0 {
		return httptest.NewRequest(method, baseURL, body)
	}
	return httptest.NewRequest(method, baseURL+target, body)
}

func createMultiPartFormData(input io.Reader, body *bytes.Buffer) (string, error) {
	var (
		fw  io.Writer
		err error
	)

	w := multipart.NewWriter(body)

	defer w.Close()

	if fw, err = w.CreateFormField("metadata"); err != nil {
		return "", err
	}

	if _, err = io.Copy(fw, input); err != nil {
		return "", err
	}

	return w.FormDataContentType(), nil
}
