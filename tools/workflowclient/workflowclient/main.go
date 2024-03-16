// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"go.temporal.io/sdk/client"

	eta "gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/emcotemporalapi"
)

// environmental variable that will have the temporal servers IP,
// and the temporal servers port
const (
	temporal_env_var  = "TEMPORAL_SERVER"
	temporal_env_port = "TEMPORAL_PORT"
)

func main() {
	// the name of the file that contains the temporal workflow parameters
	// and the golang struct that it will be unmarshalled into
	var argFileName string
	var spec *eta.WfTemporalSpec

	// Get the Temporal Server's IP from the env
	temporal_server := os.Getenv(temporal_env_var)
	temporal_port := os.Getenv(temporal_env_port)
	if temporal_server == "" {
		fmt.Fprintf(os.Stderr, "Error: Need to define $TEMPORAL_SERVER\n")
		os.Exit(1)
	}

	// Create the full url that points to the location of the running temporal server
	hostPort := temporal_server + ":" + temporal_port
	fmt.Printf("Temporal server endpoint: (%s)\n", hostPort)

	// Get the JSON file that contains the temporal workflow parameters.
	// Will exit and signal error if can't be found.
	flag.StringVar(&argFileName, "a", "", "Workflow params as JSON file")
	flag.Parse()
	if argFileName != "" {
		fmt.Printf("Will read parameters from file: %s\n", argFileName)
	} else {
		fmt.Printf("Error finding the JSON file.\n")
		os.Exit(1)
	}

	// Get the json file with temporal parameters and unmarshall it into
	// the golang struct for it.
	argFile, _ := ioutil.ReadFile(argFileName)
	err := json.Unmarshal([]byte(argFile), &spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Issues unmarshalling the json file into golang struct")
		os.Exit(1)
	}

	// Create a client object that will connect to the temporal server,
	// and we will use it to schedule a workflow for execution.
	// Use NewLazyClient in case the server is busy for a moment.
	clientOptions := client.Options{HostPort: hostPort}
	c, err := client.NewLazyClient(clientOptions)
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
		c.Close()
		os.Exit(1)
	}
	defer c.Close()

	// Put the user specified parameters into a temporal data structure,
	// then schedule the workflow to be executed by temporal
	options := client.StartWorkflowOptions(spec.WfStartOpts)
	we, err := c.ExecuteWorkflow(context.Background(), options,
		spec.WfStartOpts.TaskQueue, &spec.WfParams)
	if err != nil {
		log.Fatalln("error starting Migration Workflow", err)
	}

	// successfully ran the workflow.
	log.Printf("\nFinished workflow. WorkflowID: %s RunID: %s\n", we.GetID(), we.GetRunID())
}
