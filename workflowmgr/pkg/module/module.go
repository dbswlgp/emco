// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

// Client for using the services in the ncm
type Client struct {
	WorkflowIntentClient *WorkflowIntentClient
	// Add Clients for API's here
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.WorkflowIntentClient = NewWorkflowIntentClient()
	// Add Client API handlers here
	return c
}
