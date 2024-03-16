// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package http

import (
	"context"
	"emcopolicy/api"
	"emcopolicy/internal/controller"
	"github.com/gorilla/handlers"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

func StartHTTPServer(ctrl *controller.Controller, wg *sync.WaitGroup) {
	defer wg.Done()
	httpServer := &http.Server{
		Handler: handlers.LoggingHandler(os.Stdout, api.NewRouter(ctrl)),
		Addr:    ":" + config.GetConfiguration().ServicePort,
	}
	go func() {
		log.Info("Starting HTTP Server", log.Fields{"port": httpServer.Addr})
		if err := httpServer.ListenAndServe(); err != nil {
			log.Warn("http server exit status", log.Fields{"Error": err})
		}
	}()

	// Graceful shutdown of Mux Server
	// https://github.com/gorilla/mux#graceful-shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	if err := httpServer.Shutdown(context.Background()); err != nil {
		log.Warn("Shutting down httpServer failed.", log.Fields{"err:": err})
	}
}
