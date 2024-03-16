// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package action_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Action Suite")
}
