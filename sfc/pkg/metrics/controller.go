// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks

package metrics

import (
	"context"
	"os"
	"strconv"
	"time"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	orchModule "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/module"
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
		tracer := otel.Tracer("sfc")
		for {
			ctx, span := tracer.Start(context.Background(), "get-metrics")
			do(ctx)
			span.End()
			time.Sleep(time.Duration(time.Duration(pollingInterval) * time.Second))
		}
	}()
}

func do(ctx context.Context) {
	orchClient := orchModule.NewClient()
	client := module.NewClient()
	fields := log.Fields{"service": "sfc"}
	projects, err := orchModule.NewProjectClient().GetAllProjects(ctx)
	if err != nil {
		log.Error(err.Error(), fields)
		return
	}

	for _, proj := range projects {
		fields := fields
		fields["project"] = proj.MetaData.Name

		apps, err := orchClient.CompositeApp.GetAllCompositeApps(ctx, proj.MetaData.Name)
		if err != nil {
			log.Error(err.Error(), fields)
			continue
		}
		for _, app := range apps {
			fields := fields
			fields["composite_app"] = app.Metadata.Name
			digs, err := orchClient.DeploymentIntentGroup.GetAllDeploymentIntentGroups(ctx, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version)
			if err != nil {
				log.Error(err.Error(), fields)
				continue
			}
			for _, dig := range digs {
				fields := fields
				fields["dig"] = dig.MetaData.Name

				sfcs, err := client.SfcIntent.GetAllSfcIntents(ctx, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name)
				if err != nil {
					log.Error(err.Error(), fields)
					continue
				}
				for _, sfc := range sfcs {
					fields := fields
					fields["sfc"] = sfc.Metadata.Name
					SFCIntentGauge.WithLabelValues(
						sfc.Metadata.Name,
						proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name,
						sfc.Spec.ChainType,
						sfc.Spec.Namespace,
					).Set(1)
					links, err := client.SfcLinkIntent.GetAllSfcLinkIntents(ctx, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, sfc.Metadata.Name)
					if err != nil {
						log.Error(err.Error(), fields)
						continue
					}
					for _, link := range links {
						SFCIntentLinkGauge.WithLabelValues(
							link.Metadata.Name,
							proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name, sfc.Metadata.Name,
							link.Spec.LeftNet,
							link.Spec.RightNet,
							link.Spec.LinkLabel,
							link.Spec.AppName,
							link.Spec.WorkloadResource,
							link.Spec.ResourceType,
						).Set(1)
					}
				}

			}
		}
	}
}
