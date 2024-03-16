// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

// Client used to manage exposed client interfaces.
type Client struct {
	WorkflowIntentClient *WorkflowIntentClient
	WorkerIntentClient   *WorkerIntentClient
}

// NewClient returns a new client instance
func NewClient() *Client {
	c := &Client{}
	c.WorkflowIntentClient = NewWorkflowIntentClient()
	c.WorkerIntentClient = NewWorkerIntentClient()
	return c
}
