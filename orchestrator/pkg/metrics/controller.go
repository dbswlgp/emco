// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks

package metrics

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
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
		tracer := otel.Tracer("orchestrator")
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
	fields := log.Fields{"service": "orchestrator"}
	if err := handleControllers(ctx, ControllerGauge, client); err != nil {
		log.Error(err.Error(), fields)
		return
	}

	projects, err := module.NewProjectClient().GetAllProjects(ctx)
	if err != nil {
		log.Error(err.Error(), fields)
		return
	}

	for _, proj := range projects {
		fields := fields
		fields["project"] = proj.MetaData.Name

		ProjectGauge.WithLabelValues(proj.MetaData.Name).Set(1)
		apps, err := client.CompositeApp.GetAllCompositeApps(ctx, proj.MetaData.Name)
		if err != nil {

			log.Error(err.Error(), fields)
			continue
		}
		for _, app := range apps {
			fields := fields
			fields["composite_app"] = app.Metadata.Name

			ComAppGauge.WithLabelValues(app.Spec.Version, app.Metadata.Name, proj.MetaData.Name).Set(1)

			applications, err := client.App.GetApps(ctx, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version)
			if err != nil {
				log.Error(err.Error(), fields)
				continue
			}
			for _, application := range applications {
				fields := fields
				fields["app"] = application.Metadata.Name

				AppGauge.WithLabelValues(application.Metadata.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)

				dependencies, err := client.AppDependency.GetAllAppDependency(ctx, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, application.Metadata.Name)
				if err != nil {
					log.Error(err.Error(), fields)
					continue
				}

				for _, dependency := range dependencies {
					DependencyGauge.WithLabelValues(dependency.MetaData.Name, proj.MetaData.Name, application.Metadata.Name, app.Metadata.Name, app.Spec.Version).Set(1)
				}
			}

			digs, err := client.DeploymentIntentGroup.GetAllDeploymentIntentGroups(ctx, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version)
			if err != nil {
				log.Error(err.Error(), fields)
				continue
			}
			for _, dig := range digs {
				fields := fields
				fields["dig"] = dig.MetaData.Name

				DIGGauge.WithLabelValues(dig.MetaData.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)
				gpIntents, err := client.GenericPlacementIntent.GetAllGenericPlacementIntents(ctx, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, dig.MetaData.Name)
				if err != nil {
					log.Error(err.Error(), fields)
					continue
				}
				for _, gpi := range gpIntents {
					fields := fields
					fields["gpi"] = dig.MetaData.Name

					GenericPlacementIntentGauge.WithLabelValues(gpi.MetaData.Name, dig.MetaData.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)
					appIntents, err := client.AppIntent.GetAllAppIntents(ctx, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, gpi.MetaData.Name, dig.MetaData.Name)
					if err != nil {
						log.Error(err.Error(), fields)
						continue
					}
					for _, appIntent := range appIntents {
						GenericAppPlacementIntentGauge.WithLabelValues(appIntent.MetaData.Name, gpi.MetaData.Name, dig.MetaData.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)
					}
				}
			}

			comProfiles, err := client.CompositeProfile.GetCompositeProfiles(ctx, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version)
			if err != nil {
				log.Error(err.Error(), fields)
				continue
			}
			for _, comProfile := range comProfiles {
				fields := fields
				fields["composite_profile"] = comProfile.Metadata.Name

				CompositeProfileGauge.WithLabelValues(comProfile.Metadata.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)
				appProfiles, err := client.AppProfile.GetAppProfiles(ctx, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version, comProfile.Metadata.Name)
				if err != nil {
					log.Error(err.Error(), fields)
					continue
				}
				for _, appProfile := range appProfiles {
					AppProfileGauge.WithLabelValues(appProfile.Metadata.Name, comProfile.Metadata.Name, proj.MetaData.Name, app.Metadata.Name, app.Spec.Version).Set(1)
				}
			}
		}
	}
}

func handleControllers(ctx context.Context, p *prometheus.GaugeVec, client *module.Client) error {
	controllers, err := client.Controller.GetControllers(ctx)
	if err != nil {
		return err
	}
	p.Reset()
	for _, c := range controllers {
		p.WithLabelValues(c.Metadata.Name).Set(1)
	}
	return nil
}
