// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks

package metrics

import (
	"context"
	"os"
	"strconv"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
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
		tracer := otel.Tracer("clm")
		for {
			ctx, span := tracer.Start(context.Background(), "get-metrics")
			do(ctx)
			span.End()
			time.Sleep(time.Duration(time.Duration(pollingInterval) * time.Second))
		}
	}()
}

func do(ctx context.Context) {
	client := cluster.NewClusterClient()
	fields := log.Fields{"service": "clm"}

	clps, err := client.GetClusterProviders(ctx)
	if err != nil {
		log.Error(err.Error(), fields)
		return
	}

	for _, clp := range clps {
		fields := fields
		fields["cluster_provider"] = clp.Metadata.Name
		CLPGauge.WithLabelValues(clp.Metadata.Name).Set(1)
		clusters, err := client.GetClusters(ctx, clp.Metadata.Name)
		if err != nil {
			log.Error(err.Error(), fields)
			continue
		}
		for _, cl := range clusters {
			ClusterGauge.WithLabelValues(cl.Metadata.Name, clp.Metadata.Name).Set(1)
		}
	}
}
