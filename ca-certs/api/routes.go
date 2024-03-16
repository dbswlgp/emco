// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
)

// route holds the caCert api route details
type route struct {
	router *mux.Router
	client *client.Client
	mock   interface{}
}

const (
	// clusterProvider
	clusterProviderCertURL             string = "/cluster-providers/{clusterProvider}/ca-certs"
	clusterProviderCertDistributionURL string = clusterProviderCertURL + "/{caCert}/distribution"
	clusterProviderCertEnrollmentURL   string = clusterProviderCertURL + "/{caCert}/enrollment"
	// logicalCloud
	logicalCloudCertURL             string = "/projects/{project}/ca-certs"
	logicalCloudCertDistributionURL string = logicalCloudCertURL + "/{caCert}/distribution"
	logicalCloudCertEnrollmentURL   string = logicalCloudCertURL + "/{caCert}/enrollment"
)

// setClusterProviderRoutes set the handlers for clusterProvider caCert
func (r *route) setClusterProviderRoutes() {
	// routes to handle the clusterProvider caCert
	cpCert := cpCertHandler{
		manager: setClient(r.client.ClusterProviderCert, r.mock).(clusterprovider.CaCertManager)}
	r.router.HandleFunc(clusterProviderCertURL, cpCert.handleCertificateGet).Methods("GET")
	r.router.HandleFunc(clusterProviderCertURL, cpCert.handleCertificateCreate).Methods("POST")
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}", cpCert.handleCertificateGet).Methods("GET")
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}", cpCert.handleCertificateDelete).Methods("DELETE")
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}", cpCert.handleCertificateUpdate).Methods("PUT")

	// routes to handle the clusterProvider caCert clusterGroup
	cpCluster := cpClusterHandler{
		manager: setClient(r.client.ClusterProviderCluster, r.mock).(clusterprovider.ClusterGroupManager)}
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters", cpCluster.handleClusterGet).Methods("GET")
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters", cpCluster.handleClusterCreate).Methods("POST")
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters/{cluster}", cpCluster.handleClusterGet).Methods("GET")
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters/{cluster}", cpCluster.handleClusterDelete).Methods("DELETE")
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters/{cluster}", cpCluster.handleClusterUpdate).Methods("PUT")

	// routes to handle the clusterProvider caCert enrollment
	cpCertEnrollment := cpCertEnrollmentHandler{
		manager: setClient(r.client.ClusterProviderCertEnrollment, r.mock).(clusterprovider.CaCertEnrollmentManager)}
	r.router.HandleFunc(clusterProviderCertEnrollmentURL+"/status", cpCertEnrollment.handleStatus).Methods("GET")
	r.router.HandleFunc(clusterProviderCertEnrollmentURL+"/status",
		cpCertEnrollment.handleStatus).Queries(
		"instance", "{instance}",
		"status", "{status}",
		"output", "{output}",
		"cluster", "{cluster}",
		"resource", "{resource}",
		"clusters", "{clusters}",
		"resources", "{resources}")
	r.router.HandleFunc(clusterProviderCertEnrollmentURL+"/instantiate", cpCertEnrollment.handleInstantiate).Methods("POST")
	r.router.HandleFunc(clusterProviderCertEnrollmentURL+"/terminate", cpCertEnrollment.handleTerminate).Methods("POST")
	r.router.HandleFunc(clusterProviderCertEnrollmentURL+"/update", cpCertEnrollment.handleUpdate).Methods("POST")

	// routes to handle the clusterProvider caCert distribution
	cpCertDistribution := cpCertDistributionHandler{
		manager: setClient(r.client.ClusterProviderCertDistribution, r.mock).(clusterprovider.CaCertDistributionManager)}
	r.router.HandleFunc(clusterProviderCertDistributionURL+"/status", cpCertDistribution.handleStatus).Methods("GET")
	r.router.HandleFunc(clusterProviderCertDistributionURL+"/status",
		cpCertDistribution.handleStatus).Queries(
		"instance", "{instance}",
		"status", "{status}",
		"output", "{output}",
		"cluster", "{cluster}",
		"resource", "{resource}",
		"clusters", "{clusters}",
		"resources", "{resources}")
	r.router.HandleFunc(clusterProviderCertDistributionURL+"/instantiate", cpCertDistribution.handleInstantiate).Methods("POST")
	r.router.HandleFunc(clusterProviderCertDistributionURL+"/terminate", cpCertDistribution.handleTerminate).Methods("POST")
	r.router.HandleFunc(clusterProviderCertDistributionURL+"/update", cpCertDistribution.handleUpdate).Methods("POST")

}

// setLogicalCloudRoutes set the handlers for logicalCloud caCert
func (r *route) setLogicalCloudRoutes() {
	// routes to handle the logicalCloud caCert
	lcCert := lcCertHandler{
		manager: setClient(r.client.LogicalCloudCert, r.mock).(logicalcloud.CaCertManager)}
	r.router.HandleFunc(logicalCloudCertURL, lcCert.handleCertificateGet).Methods("GET")
	r.router.HandleFunc(logicalCloudCertURL, lcCert.handleCertificateCreate).Methods("POST")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}", lcCert.handleCertificateGet).Methods("GET")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}", lcCert.handleCertificateDelete).Methods("DELETE")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}", lcCert.handleCertificateUpdate).Methods("PUT")

	// routes to handle the logicalCloud caCert logicalCloud
	lc := lcHandler{
		manager: setClient(r.client.LogicalCloud, r.mock).(logicalcloud.CaCertLogicalCloudManager)}
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds", lc.handleLogicalCloudGet).Methods("GET")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds", lc.handleLogicalCloudCreate).Methods("POST")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}", lc.handleLogicalCloudGet).Methods("GET")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}", lc.handleLogicalCloudDelete).Methods("DELETE")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}", lc.handleLogicalCloudUpdate).Methods("PUT")

	// routes to handle the logicalCloud caCert clusterGroup
	lcCluster := lcClusterHandler{
		manager: setClient(r.client.LogicalCloudCluster, r.mock).(logicalcloud.ClusterGroupManager)}
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters", lcCluster.handleClusterGet).Methods("GET")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters", lcCluster.handleClusterCreate).Methods("POST")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters/{cluster}", lcCluster.handleClusterGet).Methods("GET")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters/{cluster}", lcCluster.handleClusterDelete).Methods("DELETE")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters/{cluster}", lcCluster.handleClusterUpdate).Methods("PUT")

	// routes to handle the logicalCloud caCert enrollment
	lcCertEnrollment := lcCertEnrollmentHandler{
		manager: setClient(r.client.LogicalCloudCertEnrollment, r.mock).(logicalcloud.CaCertEnrollmentManager)}
	r.router.HandleFunc(logicalCloudCertEnrollmentURL+"/status", lcCertEnrollment.handleStatus).Methods("GET")
	r.router.HandleFunc(logicalCloudCertEnrollmentURL+"/status",
		lcCertEnrollment.handleStatus).Queries(
		"instance", "{instance}",
		"status", "{status}",
		"output", "{output}",
		"cluster", "{cluster}",
		"resource", "{resource}",
		"clusters", "{clusters}",
		"resources", "{resources}")
	r.router.HandleFunc(logicalCloudCertEnrollmentURL+"/instantiate", lcCertEnrollment.handleInstantiate).Methods("POST")
	r.router.HandleFunc(logicalCloudCertEnrollmentURL+"/terminate", lcCertEnrollment.handleTerminate).Methods("POST")
	r.router.HandleFunc(logicalCloudCertEnrollmentURL+"/update", lcCertEnrollment.handleUpdate).Methods("POST")

	// routes to handle the logicalCloud caCert distribution
	lcCertDistribution := lcCertDistributionHandler{
		manager: setClient(r.client.LogicalCloudCertDistribution, r.mock).(logicalcloud.CaCertDistributionManager)}
	r.router.HandleFunc(logicalCloudCertDistributionURL+"/status", lcCertDistribution.handleStatus).Methods("GET")
	r.router.HandleFunc(logicalCloudCertDistributionURL+"/status",
		lcCertDistribution.handleStatus).Queries(
		"instance", "{instance}",
		"status", "{status}",
		"output", "{output}",
		"cluster", "{cluster}",
		"resource", "{resource}",
		"clusters", "{clusters}",
		"resources", "{resources}")
	r.router.HandleFunc(logicalCloudCertDistributionURL+"/instantiate", lcCertDistribution.handleInstantiate).Methods("POST")
	r.router.HandleFunc(logicalCloudCertDistributionURL+"/terminate", lcCertDistribution.handleTerminate).Methods("POST")
	r.router.HandleFunc(logicalCloudCertDistributionURL+"/update", lcCertDistribution.handleUpdate).Methods("POST")
}
