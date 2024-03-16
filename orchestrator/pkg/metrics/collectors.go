// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks

package metrics

import "github.com/prometheus/client_golang/prometheus"

var ControllerGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_controller",
	Help: "Count of Controllers",
}, []string{"name"})

var ProjectGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_project",
	Help: "Count of Projects",
}, []string{"name"})

var ComAppGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_composite_app",
	Help: "Count of Composite Apps",
}, []string{"project", "name", "version"})

var AppGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_app",
	Help: "Count of Apps",
}, []string{"name", "project", "composite_app", "composite_app_version"})

var DependencyGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_dependency",
	Help: "Count of Dependencies",
}, []string{"name", "project", "app", "composite_app", "composite_app_version"})

var DIGGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_deployment_intent_group",
	Help: "Count of Deployment Intent Groups",
}, []string{"name", "project", "composite_app", "composite_app_version"})

var GenericPlacementIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_generic_placement_intent",
	Help: "Count of Generic Placement Intents",
}, []string{"name", "deployment_intent_group", "project", "composite_app", "composite_app_version"})

var GenericAppPlacementIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_generic_app_placement_intent",
	Help: "Count of Generic App Placement Intents",
}, []string{"name", "generic_placement_intent", "deployment_intent_group", "project", "composite_app", "composite_app_version"})

var CompositeProfileGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_composite_profile",
	Help: "Count of Composite Profiles",
}, []string{"name", "project", "composite_app", "composite_app_version"})

var AppProfileGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_app_profile",
	Help: "Count of App Profiles",
}, []string{"name", "composite_profile", "project", "composite_app", "composite_app_version"})
