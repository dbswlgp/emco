// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks

package metrics

import (
	"context"
	"os"
	"strconv"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/dcm/pkg/module"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	orchModule "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	"go.opentelemetry.io/otel"
)

func Start() {
	pollingInterval, _ := strconv.Atoi(os.Getenv("EMCO_METRICS_POLLING_INTERVAL_SECS"))
	if pollingInterval == -1 {
		return
	}
	if pollingInterval == 0 {
		pollingInterval = 15
	}
	go func() {
		tracer := otel.Tracer("dcm")
		for {
			ctx, span := tracer.Start(context.Background(), "get-metrics")
			do(ctx)
			span.End()
			time.Sleep(time.Duration(time.Duration(pollingInterval) * time.Second))
		}
	}()
}

func do(ctx context.Context) {
	client := module.NewClient()
	fields := log.Fields{"service": "dcm"}
	projects, err := orchModule.NewProjectClient().GetAllProjects(ctx)
	if err != nil {
		log.Error(err.Error(), fields)
		return
	}

	for _, proj := range projects {
		fields := fields
		fields["project"] = proj.MetaData.Name

		lcs, err := client.LogicalCloud.GetAll(ctx, proj.MetaData.Name)
		if err != nil {
			log.Error(err.Error(), fields)
			continue
		}
		for _, lc := range lcs {
			fields := fields
			fields["logical_cloud"] = lc.MetaData.Name
			status, err := client.LogicalCloud.Status(ctx, proj.MetaData.Name, lc.MetaData.Name, "", "ready", "all", make([]string, 0), make([]string, 0))
			if err != nil {
				log.Error(err.Error(), fields)
			}
			LCGauge.WithLabelValues(proj.MetaData.Name, lc.MetaData.Name, string(status.DeployedStatus), status.ReadyStatus).Set(1)
		}

	}
}
