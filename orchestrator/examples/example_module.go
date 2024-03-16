// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package test

import (
	"context"
	"log"

	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

// ExampleClient_Project to test Project
func ExampleClient_Project() {
	// Get handle to the client
	c := moduleLib.NewClient()
	// Check if project is initialized
	if c.Project == nil {
		log.Println("Project is Uninitialized")
		return
	}
	// Perform operations on Project Module
	// POST request (exists == false)
	ctx := context.TODO()
	_, err := c.Project.CreateProject(ctx, moduleLib.Project{MetaData: moduleLib.ProjectMetaData{Name: "test", Description: "test", UserData1: "userData1", UserData2: "userData2"}}, false)
	if err != nil {
		log.Println(err)
		return
	}
	// PUT request (exists == true)
	_, err = c.Project.CreateProject(ctx, moduleLib.Project{MetaData: moduleLib.ProjectMetaData{Name: "test", Description: "test", UserData1: "userData1", UserData2: "userData2"}}, true)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = c.Project.GetProject(ctx, "test")
	if err != nil {
		log.Println(err)
		return
	}
	err = c.Project.DeleteProject(ctx, "test")
	if err != nil {
		log.Println(err)
	}
}
