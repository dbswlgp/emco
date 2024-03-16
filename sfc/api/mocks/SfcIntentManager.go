// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	"context"
	model "gitlab.com/project-emco/core/emco-base/src/sfc/pkg/model"
	mock "github.com/stretchr/testify/mock"
)

// SfcIntentManager is an autogenerated mock type for the SfcIntentManager type
type SfcIntentManager struct {
	mock.Mock
}

// CreateSfcIntent provides a mock function with given fields: sfc, pr, ca, caver, dig, exists
func (_m *SfcIntentManager) CreateSfcIntent(ctx context.Context, sfc model.SfcIntent, pr string, ca string, caver string, dig string, exists bool) (model.SfcIntent, error) {
	ret := _m.Called(ctx, sfc, pr, ca, caver, dig, exists)

	var r0 model.SfcIntent
	if rf, ok := ret.Get(0).(func(model.SfcIntent, string, string, string, string, bool) model.SfcIntent); ok {
		r0 = rf(sfc, pr, ca, caver, dig, exists)
	} else {
		r0 = ret.Get(0).(model.SfcIntent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.SfcIntent, string, string, string, string, bool) error); ok {
		r1 = rf(sfc, pr, ca, caver, dig, exists)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteSfcIntent provides a mock function with given fields: name, pr, ca, caver, dig
func (_m *SfcIntentManager) DeleteSfcIntent(ctx context.Context, name string, pr string, ca string, caver string, dig string) error {
	ret := _m.Called(ctx, name, pr, ca, caver, dig)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) error); ok {
		r0 = rf(name, pr, ca, caver, dig)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllSfcIntents provides a mock function with given fields: pr, ca, caver, dig
func (_m *SfcIntentManager) GetAllSfcIntents(ctx context.Context, pr string, ca string, caver string, dig string) ([]model.SfcIntent, error) {
	ret := _m.Called(ctx, pr, ca, caver, dig)

	var r0 []model.SfcIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string) []model.SfcIntent); ok {
		r0 = rf(pr, ca, caver, dig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.SfcIntent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string) error); ok {
		r1 = rf(pr, ca, caver, dig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSfcIntent provides a mock function with given fields: name, pr, ca, caver, dig
func (_m *SfcIntentManager) GetSfcIntent(ctx context.Context, name string, pr string, ca string, caver string, dig string) (model.SfcIntent, error) {
	ret := _m.Called(ctx, name, pr, ca, caver, dig)

	var r0 model.SfcIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) model.SfcIntent); ok {
		r0 = rf(name, pr, ca, caver, dig)
	} else {
		r0 = ret.Get(0).(model.SfcIntent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string) error); ok {
		r1 = rf(name, pr, ca, caver, dig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}