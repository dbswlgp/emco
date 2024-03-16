// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks

package metrics

import "github.com/prometheus/client_golang/prometheus"

var CLPGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_cluster_provider",
	Help: "Count of Cluster Providers",
}, []string{"name"})

var ClusterGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_cluster",
	Help: "Count of Clusters",
}, []string{"name", "cluster_provider"})
