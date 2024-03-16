// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/project-emco/core/emco-base/src/clm/api"
	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/metrics"
	contextDb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
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

	prometheus.MustRegister(metrics.CLPGauge)
	prometheus.MustRegister(metrics.ClusterGauge)

	server, err := controller.NewControllerServer("clm",
		api.NewRouter(nil),
		nil)
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
