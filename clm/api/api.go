// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	controller "gitlab.com/project-emco/core/emco-base/src/clm/pkg/controller"
	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/module"
)

var moduleClient *module.Client
var moduleController *module.Client

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *cluster.ClusterClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*cluster.ClusterManager)(nil)).Elem()) {
			c, ok := testClient.(cluster.ClusterManager)
			if ok {
				return c
			}
		}
	case *controller.ControllerClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*controller.ControllerManager)(nil)).Elem()) {
			c, ok := testClient.(controller.ControllerManager)
			if ok {
				return c
			}
		}
	default:
		fmt.Printf("unknown type %T\n", cl)
	}
	return client
}

// NewRouter creates a router that registers the various urls that are supported
// testClient parameter allows unit testing for a given client
func NewRouter(testClient interface{}) *mux.Router {

	moduleClient = module.NewClient()
	moduleController = module.NewController()

	router := mux.NewRouter()

	v2Router := router.PathPrefix("/v2").Subrouter()

	clusterHandler := clusterHandler{
		client: setClient(moduleClient.Cluster, testClient).(cluster.ClusterManager),
	}
	v2Router.HandleFunc("/cluster-providers", clusterHandler.createClusterProviderHandler).Methods("POST")
	v2Router.HandleFunc("/cluster-providers", clusterHandler.getClusterProviderHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}", clusterHandler.putClusterProviderHandler).Methods("PUT")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}", clusterHandler.getClusterProviderHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}", clusterHandler.deleteClusterProviderHandler).Methods("DELETE")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters", clusterHandler.createClusterHandler).Methods("POST")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters", clusterHandler.getClusterHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters", clusterHandler.getClusterHandler).Queries("label", "{label}")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters", clusterHandler.getClusterHandler).Queries("withLabels", "{withLabels}")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}", clusterHandler.getClusterHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}", clusterHandler.deleteClusterHandler).Methods("DELETE")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/labels", clusterHandler.createClusterLabelHandler).Methods("POST")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/labels", clusterHandler.getClusterLabelHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/labels/{clusterLabel}", clusterHandler.putClusterLabelHandler).Methods("PUT")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/labels/{clusterLabel}", clusterHandler.getClusterLabelHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/labels/{clusterLabel}", clusterHandler.deleteClusterLabelHandler).Methods("DELETE")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/kv-pairs", clusterHandler.createClusterKvPairsHandler).Methods("POST")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/kv-pairs", clusterHandler.getClusterKvPairsHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/kv-pairs/{clusterKv}", clusterHandler.putClusterKvPairsHandler).Methods("PUT")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/kv-pairs/{clusterKv}", clusterHandler.getClusterKvPairsHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/kv-pairs/{clusterKv}", clusterHandler.getClusterKvPairsHandler).Queries("key", "{key}")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/kv-pairs/{clusterKv}", clusterHandler.deleteClusterKvPairsHandler).Methods("DELETE")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/cluster-sync-objects", clusterHandler.createClusterSyncObjectsHandler).Methods("POST")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/cluster-sync-objects", clusterHandler.getClusterSyncObjectsHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/cluster-sync-objects/{clusterSyncObject}", clusterHandler.getClusterSyncObjectsHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/cluster-sync-objects/{clusterSyncObject}", clusterHandler.getClusterSyncObjectsHandler).Queries("key", "{key}")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/cluster-sync-objects/{clusterSyncObject}", clusterHandler.deleteClusterSyncObjectsHandler).Methods("DELETE")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/cluster-sync-objects/{clusterSyncObject}", clusterHandler.putClusterSyncObjectsHandler).Methods("PUT")

	controlHandler := controllerHandler{
		client: setClient(moduleController.Controller, testClient).(controller.ControllerManager),
	}
	v2Router.HandleFunc("/clm-controllers", controlHandler.createHandler).Methods("POST")
	v2Router.HandleFunc("/clm-controllers", controlHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/clm-controllers/{controller-name}", controlHandler.putHandler).Methods("PUT")
	v2Router.HandleFunc("/clm-controllers/{controller-name}", controlHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/clm-controllers/{controller-name}", controlHandler.deleteHandler).Methods("DELETE")

	return router
}
