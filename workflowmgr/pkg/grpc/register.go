// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package grpc

import (
	"os"
	"strconv"
	"strings"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

const default_host = "localhost"
const default_port = 9097
const default_workflowmgr_name = "workflowmgr"
const ENV_WORKFLOWMGR_NAME = "WORKFLOWMGR_NAME"

func GetServerHostPort() (string, int) {

	// expect name of this ncm program to be in env variable "WORKFLOWMGR_NAME"
	serviceName := os.Getenv(ENV_WORKFLOWMGR_NAME)
	if serviceName == "" {
		serviceName = default_workflowmgr_name
		log.Info("Using default name for WORKFLOWMGR service name", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service name to be in env variable - e.g. WORKFLOWMGR_SERVICE_HOST
	host := os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_HOST")
	if host == "" {
		host = default_host
		log.Info("Using default host for workflowmgr gRPC controller", log.Fields{
			"Host": host,
		})
	}

	// expect service port to be in env variable - e.g. WORKFLOWMGR_SERVICE_PORT
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = default_port
		log.Info("Using default port for workflowmgr gRPC controller", log.Fields{
			"Port": port,
		})
	}
	return host, port
}
