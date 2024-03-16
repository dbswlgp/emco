// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
)

// Client for using the services in the ncm
type Client struct {
	TrafficGroupIntent         *TrafficGroupIntentDbClient
	ServerInboundIntent        *InboundServerIntentDbClient
	ClientsInboundIntent       *InboundClientsIntentDbClient
	ClientsAccessInboundIntent *InboundClientsAccessIntentDbClient
	Controller                 *controller.ControllerClient
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.TrafficGroupIntent = NewTrafficGroupIntentClient()
	c.ServerInboundIntent = NewServerInboundIntentClient()
	c.ClientsInboundIntent = NewClientsInboundIntentClient()
	c.ClientsAccessInboundIntent = NewClientsAccessInboundIntentClient()
	c.Controller = controller.NewControllerClient("resources", "data", "dtc")
	return c
}
