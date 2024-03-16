// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/api"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/grpc/contextupdateserver"
	register "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc"
	contextDb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
)

func main() {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	err := db.InitializeDatabaseConnection(ctx, "emco")
	if err != nil {
		log.Error("Unable to initialize mongo database connection", log.Fields{"Error": err})
		os.Exit(1)
	}
	err = contextDb.InitializeContextDatabase()
	if err != nil {
		log.Error("Unable to initialize etcd database connection", log.Fields{"Error": err})
		os.Exit(1)
	}

	grpcServer, err := register.NewGrpcServer("gac", "GAC_NAME", 9033,
		register.RegisterContextUpdateService, contextupdateserver.NewContextupdateServer())
	if err != nil {
		log.Error("GRPC server failed to start", log.Fields{"Error": err})
		os.Exit(1)
	}

	server, err := controller.NewControllerServer("gac",
		api.NewRouter(nil),
		grpcServer)
	if err != nil {
		log.Error("Unable to create server", log.Fields{"Error": err})
		os.Exit(1)
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
	if err != nil {
		log.Error("HTTP server failed", log.Fields{"Error": err})
	}
}
