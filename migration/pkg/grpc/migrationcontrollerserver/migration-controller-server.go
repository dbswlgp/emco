// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package migrationcontroller

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/project-emco/core/emco-base/src/migration/internal/action"
	migrationcontrollerpb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/migrationcontroller"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

// MigrationcontrollerServer ...
type MigrationcontrollerServer struct {
}

// MigrationApps ...
func (cs *MigrationcontrollerServer) MigrationApps(ctx context.Context, req *migrationcontrollerpb.MigrationRequest) (*migrationcontrollerpb.MigrationResponse, error) {
	log.Info("Received MigrationApps request .. start", log.Fields{"ctx": ctx, "req": req})

	if (req != nil) && (len(req.AppContext) > 0) {
		err := action.MigrationApps(ctx, req.AppContext, req.App)
		if err != nil {
			log.Error("Received MigrationApps request .. internal error.", log.Fields{"req": req, "err": err})
			return &migrationcontrollerpb.MigrationResponse{AppContext: req.AppContext, Status: false, Message: err.Error()}, nil
		}
	} else {
		log.Error("Received MigrationApps request .. invalid request error.", log.Fields{"req": req})
		return &migrationcontrollerpb.MigrationResponse{Status: false, Message: errors.New("invalid request error").Error()}, nil
	}

	log.Info("Received MigrationApps request .. end", log.Fields{"req": req})
	return &migrationcontrollerpb.MigrationResponse{AppContext: req.AppContext, Status: true, Message: fmt.Sprintf("Successful Selection of Apps" )}, nil
}

// NewMigrationControllerServer ...
func NewMigrationControllerServer() *MigrationcontrollerServer {
	s := &MigrationcontrollerServer{}
	return s
}
