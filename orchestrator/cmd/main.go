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
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/api"
	register "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc"
	contextDb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/rpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/metrics"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/statusnotify"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()

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

	grpcServer, err := register.NewGrpcServer("orchestrator", "ORCHESTRATOR_NAME", 9016,
		register.RegisterStatusNotifyService, statusnotify.StartStatusNotifyServer())
	if err != nil {
		log.Error("Unable to create gRPC server", log.Fields{"Error": err})
		os.Exit(1)
	}

	prometheus.MustRegister(metrics.ComAppGauge)
	prometheus.MustRegister(metrics.ProjectGauge)
	prometheus.MustRegister(metrics.ControllerGauge)
	prometheus.MustRegister(metrics.DIGGauge)
	prometheus.MustRegister(metrics.GenericPlacementIntentGauge)
	prometheus.MustRegister(metrics.CompositeProfileGauge)
	prometheus.MustRegister(metrics.AppProfileGauge)
	prometheus.MustRegister(metrics.GenericAppPlacementIntentGauge)
	prometheus.MustRegister(metrics.AppGauge)
	prometheus.MustRegister(metrics.DependencyGauge)

	server, err := controller.NewControllerServer("orchestrator",
		api.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil),
		grpcServer)
	if err != nil {
		log.Error("Unable to create server", log.Fields{"Error": err})
		os.Exit(1)
	}

	controller.NewControllerClient("resources", "data", "orchestrator").InitControllers(ctx)

	connectionsClose := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		server.Shutdown(ctx)
		rpc.CloseAllRpcConn()
		close(connectionsClose)
	}()

	metrics.Start()
	err = server.ListenAndServe()
	if err != nil {
		log.Error("Server failed", log.Fields{"Error": err})
	}
}
