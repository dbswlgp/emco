// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks

package metrics

import "github.com/prometheus/client_golang/prometheus"

var NetworkGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_cluster_network",
	Help: "Count of Cluster Networks",
}, []string{"clusterprovider", "cluster", "name", "cnitype"})

var ProviderNetworkGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_cluster_provider_network",
	Help: "Count of Cluster Provider Networks",
}, []string{"clusterprovider", "cluster", "name", "cnitype", "nettype", "vlanid", "providerinterfacename", "logicalinterfacename", "vlannodeselector"})
