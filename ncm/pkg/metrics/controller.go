// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks

package metrics

import (
	"context"
	"os"
	"strconv"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	netintents "gitlab.com/project-emco/core/emco-base/src/ncm/pkg/networkintents"
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
		tracer := otel.Tracer("ncm")
		for {
			ctx, span := tracer.Start(context.Background(), "get-metrics")
			do(ctx)
			span.End()
			time.Sleep(time.Duration(time.Duration(pollingInterval) * time.Second))
		}
	}()
}

func do(ctx context.Context) {
	clusterClient := cluster.NewClusterClient()
	netClient := netintents.NewNetworkClient()
	providerNetClient := netintents.NewProviderNetClient()
	fields := log.Fields{"service": "ncm"}
	clps, err := clusterClient.GetClusterProviders(ctx)
	if err != nil {
		log.Error(err.Error(), fields)
		return
	}

	for _, clp := range clps {
		fields := fields
		fields["cluster_provider"] = clp.Metadata.Name
		clusters, err := clusterClient.GetClusters(ctx, clp.Metadata.Name)
		if err != nil {
			log.Error(err.Error(), fields)
			continue
		}
		for _, cl := range clusters {
			fields := fields
			fields["cluster"] = cl.Metadata.Name
			networks, err := netClient.GetNetworks(ctx, clp.Metadata.Name, cl.Metadata.Name)
			if err != nil {
				log.Error(err.Error(), fields)
				continue
			}
			for _, network := range networks {
				NetworkGauge.WithLabelValues(clp.Metadata.Name, cl.Metadata.Name, network.Metadata.Name, network.Spec.CniType).Set(1)
			}

			providerNets, err := providerNetClient.GetProviderNets(ctx, clp.Metadata.Name, cl.Metadata.Name)
			if err != nil {
				log.Error(err.Error(), fields)
				continue
			}
			for _, network := range providerNets {
				ProviderNetworkGauge.WithLabelValues(
					clp.Metadata.Name,
					cl.Metadata.Name,
					network.Metadata.Name,
					network.Spec.CniType,
					network.Spec.ProviderNetType,
					network.Spec.Vlan.VlanId,
					network.Spec.Vlan.ProviderInterfaceName,
					network.Spec.Vlan.LogicalInterfaceName,
					network.Spec.Vlan.VlanNodeSelector,
				).Set(1)
			}
		}
	}
}
