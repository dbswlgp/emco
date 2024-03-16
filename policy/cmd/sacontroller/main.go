// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package main

import (
	"context"
	"emcopolicy/internal/controller"
	"emcopolicy/pkg/http"
	"github.com/sirupsen/logrus"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"sync"
)

func main() {
	// Currently, we are reading workflow manager IP from environment
	// We are supporting only workflow manager now, and hence this parameter is must
	// But when we support more plugins, this need not be mandatory and should read from config file
	//workflowManagerUrl := flag.String("workflowmgr", "", "EMCO workflow manager endpoint")
	//flag.Parse()
	log.SetLoglevel(logrus.InfoLevel)
	log.Info("Starting Policy Controller", log.Fields{})
	wg := new(sync.WaitGroup)
	// Create Controller context and start scheduler.
	// Scheduler should start before the api & event server
	c, err := controller.Init()
	if err != nil {
		log.Fatal("Policy controller init failed", log.Fields{"Err": err})
	}
	if c == nil {
		log.Fatal("Policy controller init failed. Controller is nil", log.Fields{})
	}
	if err = c.StartScheduler(context.Background()); err != nil {
		log.Fatal("Scheduler failed to start", log.Fields{"err": err})
	}
	// HTTP Server Initialization
	wg.Add(1)
	go http.StartHTTPServer(c, wg)
	wg.Wait()
}
