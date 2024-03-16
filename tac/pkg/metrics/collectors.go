// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks

package metrics

import "github.com/prometheus/client_golang/prometheus"

var TACIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_tac_intent",
	Help: "Count of Temporal Action Controller Intents",
}, []string{
	"name", "project", "composite_app", "composite_app_version", "deployment_intent_group",
	"hoot_type", "client_endpoint_name", "client_endpoint_port", "workflow_client_name",
})
