// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"

	hpaModel "gitlab.com/project-emco/core/emco-base/src/hpa-plc/pkg/model"
)

func executeRequest(request *http.Request, router *mux.Router) *http.Response {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	resp := recorder.Result()
	return resp
}

type mockIntentManager struct {
	Items             []hpaModel.DeploymentHpaIntent
	ItemsSpec         []hpaModel.DeploymentHpaIntentSpec
	ConsumerItems     []hpaModel.HpaResourceConsumer
	ConsumerItemsSpec []hpaModel.HpaResourceConsumerSpec
	ResourceItems     []hpaModel.HpaResourceRequirement
	ResourceItemsSpec []hpaModel.HpaResourceRequirementSpec
	Err               error
}

func (m *mockIntentManager) AddIntent(ctx context.Context, a hpaModel.DeploymentHpaIntent, p string, ca string, v string, di string, exists bool) (hpaModel.DeploymentHpaIntent, error) {
	if m.Err != nil {
		return hpaModel.DeploymentHpaIntent{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockIntentManager) GetIntent(ctx context.Context, name string, p string, ca string, v string, di string) (hpaModel.DeploymentHpaIntent, bool, error) {
	if m.Err != nil {
		return hpaModel.DeploymentHpaIntent{}, false, m.Err
	}

	return m.Items[0], true, nil
}

func (m *mockIntentManager) GetAllIntents(ctx context.Context, p string, ca string, v string, di string) ([]hpaModel.DeploymentHpaIntent, error) {

	if m.Err != nil {
		return []hpaModel.DeploymentHpaIntent{}, m.Err
	}

	return m.Items, nil
}

func (m *mockIntentManager) GetAllIntentsByApp(ctx context.Context, app string, p string, ca string, v string, di string) ([]hpaModel.DeploymentHpaIntent, error) {

	if m.Err != nil {
		return []hpaModel.DeploymentHpaIntent{}, m.Err
	}

	return m.Items, nil
}

func (m *mockIntentManager) GetIntentByName(ctx context.Context, i string, p string, ca string, v string, di string) (hpaModel.DeploymentHpaIntent, error) {

	if m.Err != nil {
		return hpaModel.DeploymentHpaIntent{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockIntentManager) DeleteIntent(ctx context.Context, i string, p string, ca string, v string, di string) error {
	return m.Err
}

// consumers
func (m *mockIntentManager) AddConsumer(ctx context.Context, a hpaModel.HpaResourceConsumer, p string, ca string, v string, di string, i string, exists bool) (hpaModel.HpaResourceConsumer, error) {
	if m.Err != nil {
		return hpaModel.HpaResourceConsumer{}, m.Err
	}

	return m.ConsumerItems[0], nil
}

func (m *mockIntentManager) GetConsumer(ctx context.Context, cn string, p string, ca string, v string, di string, i string) (hpaModel.HpaResourceConsumer, bool, error) {
	if m.Err != nil {
		return hpaModel.HpaResourceConsumer{}, false, m.Err
	}
	return m.ConsumerItems[0], false, nil
}

func (m *mockIntentManager) GetAllConsumers(ctx context.Context, p, ca, v, di, i string) ([]hpaModel.HpaResourceConsumer, error) {
	if m.Err != nil {
		return []hpaModel.HpaResourceConsumer{}, m.Err
	}
	return m.ConsumerItems, nil
}

func (m *mockIntentManager) GetConsumerByName(ctx context.Context, cn, p, ca, v, di, i string) (hpaModel.HpaResourceConsumer, error) {
	if m.Err != nil {
		return hpaModel.HpaResourceConsumer{}, m.Err
	}
	return m.ConsumerItems[0], nil
}

func (m *mockIntentManager) DeleteConsumer(ctx context.Context, cn, p string, ca string, v string, di string, i string) error {
	return nil
}

// resources
func (m *mockIntentManager) AddResource(ctx context.Context, a hpaModel.HpaResourceRequirement, p string, ca string, v string, di string, i string, cn string, exists bool) (hpaModel.HpaResourceRequirement, error) {
	if m.Err != nil {
		return hpaModel.HpaResourceRequirement{}, m.Err
	}

	return m.ResourceItems[0], nil
}

func (m *mockIntentManager) GetResource(ctx context.Context, rn string, p string, ca string, v string, di string, i string, cn string) (hpaModel.HpaResourceRequirement, bool, error) {
	if m.Err != nil {
		return hpaModel.HpaResourceRequirement{}, false, m.Err
	}

	return m.ResourceItems[0], false, nil
}

func (m *mockIntentManager) GetAllResources(ctx context.Context, p, ca, v, di, i, cn string) ([]hpaModel.HpaResourceRequirement, error) {
	if m.Err != nil {
		return []hpaModel.HpaResourceRequirement{}, m.Err
	}

	return m.ResourceItems, nil

}

func (m *mockIntentManager) GetResourceByName(ctx context.Context, rn, p, ca, v, di, i, cn string) (hpaModel.HpaResourceRequirement, error) {
	if m.Err != nil {
		return hpaModel.HpaResourceRequirement{}, m.Err
	}

	return m.ResourceItems[0], nil
}

func (m *mockIntentManager) DeleteResource(ctx context.Context, rn string, p string, ca string, v string, di string, i string, cn string) error {
	return nil
}
