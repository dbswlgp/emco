// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks

package metrics

import "github.com/prometheus/client_golang/prometheus"

var SFCIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_sfc_intent",
	Help: "Count of Network Chain Intents",
}, []string{
	"name", "project", "composite_app", "composite_app_version", "deployment_intent_group",
	"chainType", "namespace",
})

var SFCIntentLinkGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_sfc_intent_link",
	Help: "Count of Network Chain Intent Links",
}, []string{
	"name", "project", "composite_app", "composite_app_version", "deployment_intent_group", "sfc",
	"left_net",
	"right_net",
	"link_label",
	"app_name",
	"workload_resource",
	"resource_type",
})
