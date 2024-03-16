// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks

package metrics

import "github.com/prometheus/client_golang/prometheus"

var NetworkControllerIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_network_controller_intent",
	Help: "Count of Network Controller Intents",
}, []string{"name", "project", "composite_app", "composite_app_version", "deployment_intent_group"})

var WorkloadIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_workload_intent",
	Help: "Count of Workload Intents",
}, []string{"name", "project", "composite_app", "composite_app_version", "deployment_intent_group", "network_controller_intent", "app_label", "workload_resource", "type"})

var WorkloadInterfaceIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_workload_interface_intent",
	Help: "Count of Workload Interface Intents",
}, []string{
	"name",
	"project",
	"composite_app",
	"composite_app_version",
	"deployment_intent_group",
	"network_controller_intent",
	"workload_intent",
	"interface",
	"network_name",
	"default_gateway",
	"ip_address",
	"mac_address",
})
