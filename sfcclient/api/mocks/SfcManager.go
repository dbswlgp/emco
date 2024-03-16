// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	"context"
	mock "github.com/stretchr/testify/mock"
	model "gitlab.com/project-emco/core/emco-base/src/sfcclient/pkg/model"
)

// SfcManager is an autogenerated mock type for the SfcManager type
type SfcManager struct {
	mock.Mock
}

// CreateSfcClientIntent provides a mock function with given fields: sfc, pr, ca, caver, dig, exists
func (_m *SfcManager) CreateSfcClientIntent(ctx context.Context, sfc model.SfcClientIntent, pr string, ca string, caver string, dig string, exists bool) (model.SfcClientIntent, error) {
	ret := _m.Called(ctx, sfc, pr, ca, caver, dig, exists)

	var r0 model.SfcClientIntent
	if rf, ok := ret.Get(0).(func(model.SfcClientIntent, string, string, string, string, bool) model.SfcClientIntent); ok {
		r0 = rf(sfc, pr, ca, caver, dig, exists)
	} else {
		r0 = ret.Get(0).(model.SfcClientIntent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.SfcClientIntent, string, string, string, string, bool) error); ok {
		r1 = rf(sfc, pr, ca, caver, dig, exists)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteSfcClientIntent provides a mock function with given fields: name, pr, ca, caver, dig
func (_m *SfcManager) DeleteSfcClientIntent(ctx context.Context, name string, pr string, ca string, caver string, dig string) error {
	ret := _m.Called(ctx, name, pr, ca, caver, dig)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) error); ok {
		r0 = rf(name, pr, ca, caver, dig)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllSfcClientIntents provides a mock function with given fields: pr, ca, caver, dig
func (_m *SfcManager) GetAllSfcClientIntents(ctx context.Context, pr string, ca string, caver string, dig string) ([]model.SfcClientIntent, error) {
	ret := _m.Called(ctx, pr, ca, caver, dig)

	var r0 []model.SfcClientIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string) []model.SfcClientIntent); ok {
		r0 = rf(pr, ca, caver, dig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.SfcClientIntent)
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

// GetSfcClientIntent provides a mock function with given fields: name, pr, ca, caver, dig
func (_m *SfcManager) GetSfcClientIntent(ctx context.Context, name string, pr string, ca string, caver string, dig string) (model.SfcClientIntent, error) {
	ret := _m.Called(ctx, name, pr, ca, caver, dig)

	var r0 model.SfcClientIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) model.SfcClientIntent); ok {
		r0 = rf(name, pr, ca, caver, dig)
	} else {
		r0 = ret.Get(0).(model.SfcClientIntent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string) error); ok {
		r1 = rf(name, pr, ca, caver, dig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}