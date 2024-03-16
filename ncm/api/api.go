// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ncm/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/ncm/pkg/networkintents"
	"gitlab.com/project-emco/core/emco-base/src/ncm/pkg/scheduler"
)

var moduleClient *module.Client

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *networkintents.NetworkClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*networkintents.NetworkManager)(nil)).Elem()) {
			c, ok := testClient.(networkintents.NetworkManager)
			if ok {
				return c
			}
		}
	case *networkintents.ProviderNetClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*networkintents.ProviderNetManager)(nil)).Elem()) {
			c, ok := testClient.(networkintents.ProviderNetManager)
			if ok {
				return c
			}
		}
	case *scheduler.SchedulerClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*scheduler.SchedulerManager)(nil)).Elem()) {
			c, ok := testClient.(scheduler.SchedulerManager)
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

	router := mux.NewRouter()
	v2Router := router.PathPrefix("/v2").Subrouter()

	networkHandler := networkHandler{
		client: setClient(moduleClient.Network, testClient).(networkintents.NetworkManager),
	}
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/networks", networkHandler.createNetworkHandler).Methods("POST")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/networks", networkHandler.getNetworkHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/networks/{network}", networkHandler.putNetworkHandler).Methods("PUT")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/networks/{network}", networkHandler.getNetworkHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/networks/{network}", networkHandler.deleteNetworkHandler).Methods("DELETE")

	providernetHandler := providernetHandler{
		client: setClient(moduleClient.ProviderNet, testClient).(networkintents.ProviderNetManager),
	}
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/provider-networks", providernetHandler.createProviderNetHandler).Methods("POST")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/provider-networks", providernetHandler.getProviderNetHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/provider-networks/{providerNetwork}", providernetHandler.putProviderNetHandler).Methods("PUT")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/provider-networks/{providerNetwork}", providernetHandler.getProviderNetHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/provider-networks/{providerNetwork}", providernetHandler.deleteProviderNetHandler).Methods("DELETE")

	schedulerHandler := schedulerHandler{
		client: setClient(moduleClient.Scheduler, testClient).(scheduler.SchedulerManager),
	}
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/apply", schedulerHandler.applySchedulerHandler).Methods("POST")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/terminate", schedulerHandler.terminateSchedulerHandler).Methods("POST")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/stop", schedulerHandler.stopSchedulerHandler).Methods("POST")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/status", schedulerHandler.statusSchedulerHandler).Methods("GET")
	v2Router.HandleFunc("/cluster-providers/{clusterProvider}/clusters/{cluster}/status",
		schedulerHandler.statusSchedulerHandler).Queries("instance", "{instance}", "status", "{status}", "type", "{type}", "output", "{output}", "app", "{app}", "cluster", "{cluster}", "resource", "{resource}")

	return router
}
