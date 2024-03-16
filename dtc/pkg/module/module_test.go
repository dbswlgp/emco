// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/project-emco/core/emco-base/src/dtc/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
)

var _ = Describe("Trafficgroupintent", func() {

	var (
		client *module.Client
	)

	BeforeEach(func() {
		client = &module.Client{}
		client.TrafficGroupIntent = module.NewTrafficGroupIntentClient()
		client.ServerInboundIntent = module.NewServerInboundIntentClient()
		client.ClientsInboundIntent = module.NewClientsInboundIntentClient()
		client.ClientsAccessInboundIntent = module.NewClientsAccessInboundIntentClient()
		client.Controller = controller.NewControllerClient("resources", "data", "dtc")
	})

	Describe("Create new client", func() {
		It("should return client", func() {
			c := module.NewClient()
			Expect(c).Should(Equal(client))
		})
	})
})
