// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

var (
	mockdb *db.NewMockDB
)

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Module Suite")
}

func init() {
	mockdb = &db.NewMockDB{}
	db.DBconn = mockdb
}
