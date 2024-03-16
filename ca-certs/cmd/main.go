// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/api"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdate"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()

	err := db.InitializeDatabaseConnection(ctx, "emco")
	if err != nil {
		logutils.Error("Failed to initialize mongo database connection.", logutils.Fields{"Error": err.Error()})
		os.Exit(1)
	}
	err = contextdb.InitializeContextDatabase()
	if err != nil {
		logutils.Error("Failed to initialize etcd database connection.", logutils.Fields{"Error": err.Error()})
		os.Exit(1)
	}

	type contextupdateServer struct {
		contextupdate.UnimplementedContextupdateServer
	}
	grpcServer, err := grpc.NewGrpcServer("ca-certs", "CACERT_NAME", 9035,
		grpc.RegisterContextUpdateService, &contextupdateServer{})
	if err != nil {
		logutils.Error("Unable to create gRPC server", logutils.Fields{"Error": err})
		os.Exit(1)
	}

	server, err := controller.NewControllerServer("ca-certs",
		api.NewRouter(nil),
		grpcServer)
	if err != nil {
		logutils.Error("Unable to create server", logutils.Fields{"Error": err})
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
		logutils.Error("Server failed", logutils.Fields{"Error": err})
	}
}
