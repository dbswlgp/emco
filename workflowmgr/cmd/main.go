// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
	"gitlab.com/project-emco/core/emco-base/src/workflowmgr/api"
)

func main() {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())

	err := db.InitializeDatabaseConnection(ctx, "emco")
	if err != nil {
		log.Println("Unable to initialize mongo database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}
	// workflowmgr does not update appcontext

	server, err := controller.NewControllerServer("orchestrator",
		api.NewRouter(nil),
		nil)
	if err != nil {
		log.Println("Unable to create server...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}

	connectionsClose := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		server.Shutdown(ctx)
		close(connectionsClose)
	}()

	err = server.ListenAndServe()
	log.Printf("HTTP server returned error: %s", err)
}
