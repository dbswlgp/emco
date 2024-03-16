// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/ovnaction/pkg/module"
)

var moduleClient *moduleLib.Client

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *moduleLib.NetControlIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.NetControlIntentManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.NetControlIntentManager)
			if ok {
				return c
			}
		}
	case *moduleLib.WorkloadIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.WorkloadIntentManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.WorkloadIntentManager)
			if ok {
				return c
			}
		}
	case *moduleLib.WorkloadIfIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.WorkloadIfIntentManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.WorkloadIfIntentManager)
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
	moduleClient = moduleLib.NewClient()

	router := mux.NewRouter()
	v2Router := router.PathPrefix("/v2").Subrouter()

	netcontrolintentHandler := netcontrolintentHandler{
		client: setClient(moduleClient.NetControlIntent, testClient).(moduleLib.NetControlIntentManager),
	}
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent", netcontrolintentHandler.createHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent", netcontrolintentHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent/{netControllerIntent}", netcontrolintentHandler.putHandler).Methods("PUT")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent/{netControllerIntent}", netcontrolintentHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent/{netControllerIntent}", netcontrolintentHandler.deleteHandler).Methods("DELETE")

	workloadintentHandler := workloadintentHandler{
		client: setClient(moduleClient.WorkloadIntent, testClient).(moduleLib.WorkloadIntentManager),
	}
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent/{netControllerIntent}/workload-intents", workloadintentHandler.createHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent/{netControllerIntent}/workload-intents", workloadintentHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent/{netControllerIntent}/workload-intents/{workloadIntent}", workloadintentHandler.putHandler).Methods("PUT")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent/{netControllerIntent}/workload-intents/{workloadIntent}", workloadintentHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent/{netControllerIntent}/workload-intents/{workloadIntent}", workloadintentHandler.deleteHandler).Methods("DELETE")

	workloadifintentHandler := workloadifintentHandler{
		client: setClient(moduleClient.WorkloadIfIntent, testClient).(moduleLib.WorkloadIfIntentManager),
	}
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent/{netControllerIntent}/workload-intents/{workloadIntent}/interfaces", workloadifintentHandler.createHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent/{netControllerIntent}/workload-intents/{workloadIntent}/interfaces", workloadifintentHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent/{netControllerIntent}/workload-intents/{workloadIntent}/interfaces/{interfaceIntent}", workloadifintentHandler.putHandler).Methods("PUT")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent/{netControllerIntent}/workload-intents/{workloadIntent}/interfaces/{interfaceIntent}", workloadifintentHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-controller-intent/{netControllerIntent}/workload-intents/{workloadIntent}/interfaces/{interfaceIntent}", workloadifintentHandler.deleteHandler).Methods("DELETE")

	return router
}
