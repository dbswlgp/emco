// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// http_server handles the incoming requests from various emco controllers
// and starting their respective workflows. Intended to be run inside the same
// container as workflow client.

const (
	name       = "workflow-listener"
	execDir    = "/opt/emco"
	invokerURL = "/invoke/{wfclient:[a-zA-Z0-9-_]+}" // URL to invoke the workflow client
	httpPort   = "9090"
)

// runWorkflowClient runs the workflow client named by the URL.
//  The URL is expected to be of the form /invoke/$workflow_client_name .
//  The executable binary for the workflow client must be in execDir.
func runWorkflowClient(w http.ResponseWriter, r *http.Request) {
	// Read the body of the post request from the user.
	// This should contain all the information Temporal needs
	// to run the workflow being requested.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		wrapErr := fmt.Errorf("POST body read err; %v", err)
		log.Println(wrapErr.Error())
		http.Error(w, wrapErr.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("POST body: %s", string(body))

	// This should be the name of the workflow the user wants to execute.
	// http://{our-ip}/invoke/{name-of-workflow-to-be-executed}
	params := mux.Vars(r)
	wfClientName := params["wfclient"]

	// Create a temp file, in /tmp by default.
	// This temp file will store all of the temporal information
	// needed to run a workflow for the invoking client.
	// NOTE: Go replaces "*" in the name with a random number.
	tmpfile, err := ioutil.TempFile("", wfClientName+".*.json")
	if err != nil {
		wrapErr := fmt.Errorf("failed to create temp file for %s", wfClientName)
		log.Println(wrapErr.Error())
		http.Error(w, wrapErr.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Created file: %s", tmpfile.Name())

	// Write POST body to the temp file.
	// this should be a filled out JSON object from:
	// emco-base/src/workflowmgr/pkg/emcotemporalapi/emco_temporal_api.go
	if err = ioutil.WriteFile(tmpfile.Name(), body, 0444); err != nil {
		wrapErr := fmt.Errorf("failed to write POST body to temp file %s"+"Error: %s", tmpfile.Name(), err)
		log.Println(wrapErr.Error())
		http.Error(w, wrapErr.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Wrote POST body to file: %s", tmpfile.Name())

	// Prepare the command about to be executed. This will run the workflow client binary.
	// Which will tell temporal which workflow to run, and give it all the information it needs to run it.
	// Hard code the workflow client because we only need the one.
	wfClient := path.Join(execDir, "workflowclient")
	log.Printf("Will execute: (%s -a %s)\n", wfClient, tmpfile.Name())

	// Execute the workflow client binary, and tell it where the JSON file
	// with all the workflow running parameters is via the command line
	cmd := exec.Command(wfClient, "-a", tmpfile.Name())
	cmdOutErr, err := cmd.CombinedOutput()
	if err != nil {
		wrapErr := fmt.Errorf("%s finished with error: %v", wfClient, err)
		log.Println(wrapErr.Error())
		http.Error(w, wrapErr.Error(), http.StatusInternalServerError)
		return
	}

	// output the result from running the temporal workflow
	log.Printf("\nOutput from %s :\n%s\n", wfClient, cmdOutErr)
	w.WriteHeader(http.StatusNoContent)
}

// NewRouter creates a router and registers the invoke workflow client route
func NewRouter() *mux.Router {

	router := mux.NewRouter()

	router.HandleFunc(invokerURL, runWorkflowClient).Methods("POST")

	return router
}

// main function which will spin up a server and handle incoming temporal workflow requests.
func main() {
	httpRouter := NewRouter()
	loggedRouter := handlers.LoggingHandler(os.Stdout, httpRouter)
	log.Println("Starting http server")

	httpServer := &http.Server{
		Handler: loggedRouter,
		Addr:    ":" + httpPort,
	}

	connectionsClose := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		httpServer.Shutdown(context.Background())
		close(connectionsClose)
	}()

	err := httpServer.ListenAndServe()
	log.Printf("httpServer returned: %s\n", err)
}
