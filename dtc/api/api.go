// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/dtc/pkg/module"
	controller "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
)

var (
	moduleClient     *module.Client
	moduleController *controller.ControllerClient
)

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *module.TrafficGroupIntentDbClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.TrafficGroupIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.TrafficGroupIntentManager)
			if ok {
				return c
			}
		}
	case *module.InboundServerIntentDbClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.InboundServerIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.InboundServerIntentManager)
			if ok {
				return c
			}
		}
	case *module.InboundClientsIntentDbClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.InboundClientsIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.InboundClientsIntentManager)
			if ok {
				return c
			}
		}
	case *module.InboundClientsAccessIntentDbClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.InboundClientsAccessIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.InboundClientsAccessIntentManager)
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
func NewRouter(testClient interface{}) *mux.Router {
	moduleClient = module.NewClient()

	router := mux.NewRouter()
	v2Router := router.PathPrefix("/v2").Subrouter()
	trafficgroupintentHandler := trafficgroupintentHandler{
		client: setClient(moduleClient.TrafficGroupIntent, testClient).(module.TrafficGroupIntentManager),
	}
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents", trafficgroupintentHandler.createHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents", trafficgroupintentHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}", trafficgroupintentHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}", trafficgroupintentHandler.putHandler).Methods("PUT")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}", trafficgroupintentHandler.deleteHandler).Methods("DELETE")

	inboundserverintentHandler := inboundserverintentHandler{
		client: setClient(moduleClient.ServerInboundIntent, testClient).(module.InboundServerIntentManager),
	}
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents", inboundserverintentHandler.createHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}", inboundserverintentHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents", inboundserverintentHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}", inboundserverintentHandler.putHandler).Methods("PUT")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}", inboundserverintentHandler.deleteHandler).Methods("DELETE")

	inboundclientsintentHandler := inboundclientsintentHandler{
		client: setClient(moduleClient.ClientsInboundIntent, testClient).(module.InboundClientsIntentManager),
	}
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients", inboundclientsintentHandler.createHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients", inboundclientsintentHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}", inboundclientsintentHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}", inboundclientsintentHandler.putHandler).Methods("PUT")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}", inboundclientsintentHandler.deleteHandler).Methods("DELETE")

	inboundclientsaccessintentHandler := inboundclientsaccessintentHandler{
		client: setClient(moduleClient.ClientsAccessInboundIntent, testClient).(module.InboundClientsAccessIntentManager),
	}
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}/access-points", inboundclientsaccessintentHandler.createHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}/access-points", inboundclientsaccessintentHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}/access-points/{inboundClientsAccessIntent}", inboundclientsaccessintentHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}/access-points/{inboundClientsAccessIntent}", inboundclientsaccessintentHandler.putHandler).Methods("PUT")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}/access-points/{inboundClientsAccessIntent}", inboundclientsaccessintentHandler.deleteHandler).Methods("DELETE")

	controlHandler := controllerHandler{
		client: setClient(moduleClient.Controller, testClient).(controller.ControllerManager),
	}
	v2Router.HandleFunc("/dtc-controllers", controlHandler.createHandler).Methods("POST")
	v2Router.HandleFunc("/dtc-controllers", controlHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/dtc-controllers/{dtcController}", controlHandler.putHandler).Methods("PUT")
	v2Router.HandleFunc("/dtc-controllers/{dtcController}", controlHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/dtc-controllers/{dtcController}", controlHandler.deleteHandler).Methods("DELETE")

	return router
}
