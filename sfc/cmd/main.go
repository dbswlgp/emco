// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	register "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	contextDb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
	"gitlab.com/project-emco/core/emco-base/src/sfc/api"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/grpc/contextupdateserver"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/metrics"
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

	grpcServer, err := register.NewGrpcServer("sfc", "SFC_NAME", 9056,
		register.RegisterContextUpdateService, contextupdateserver.NewContextupdateServer())
	if err != nil {
		log.Error("Unable to create gRPC server", log.Fields{"Error": err})
		os.Exit(1)
	}

	prometheus.MustRegister(metrics.SFCIntentGauge)
	prometheus.MustRegister(metrics.SFCIntentLinkGauge)

	log.Info("Starting SFC Action Controller", log.Fields{"Port": config.GetConfiguration().ServicePort})

	server, err := controller.NewControllerServer("sfc",
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

	metrics.Start()
	err = server.ListenAndServe()
	if err != nil {
		log.Error("Server failed", log.Fields{"Error": err})
	}
}
