// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks

package metrics

import "github.com/prometheus/client_golang/prometheus"

var TrafficGroupIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_deployment_intent_group_traffic_group_intent",
	Help: "Count of Traffic Group Intents",
}, []string{"name", "project", "composite_app", "composite_app_version", "deployment_intent_group"})

var InboundIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_deployment_intent_group_inbound_intent",
	Help: "Count of Inbound Intents",
}, []string{
	"name",
	"project",
	"composite_app",
	"composite_app_version",
	"deployment_intent_group",
	"traffic_group_intent",
	"spec_app",
	"app_label",
	"serviceName",
	"externalName",
	"port",
	"protocol",
	"externalSupport",
	"serviceMesh",
	"sidecarProxy",
	"tlsType",
})

var InboundIntentClientGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_deployment_intent_group_inbound_intent_client",
	Help: "Count of Inbound Intent Clients",
}, []string{
	"name",
	"project",
	"composite_app",
	"composite_app_version",
	"deployment_intent_group",
	"traffic_group_intent",
	"inbound_intent",
	"spec_app",
	"app_label",
	"serviceName",
})

var InboundIntentClientAPGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_deployment_intent_group_inbound_intent_client_access_point",
	Help: "Count of Inbound Intent Client Access Points",
}, []string{
	"name",
	"project",
	"composite_app",
	"composite_app_version",
	"deployment_intent_group",
	"traffic_group_intent",
	"inbound_intent",
	"client_name",
	"action",
})
