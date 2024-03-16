// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks

package metrics

import "github.com/prometheus/client_golang/prometheus"

var LCGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_logical_cloud",
	Help: "Count of Logical Clouds",
}, []string{"project", "name", "deployed_status", "ready_status"})
