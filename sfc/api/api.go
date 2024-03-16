// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/module"
)

var (
	moduleSfcIntentClient                *module.SfcIntentClient
	moduleSfcClientSelectorIntentClient  *module.SfcClientSelectorIntentClient
	moduleSfcProviderNetworkIntentClient *module.SfcProviderNetworkIntentClient
)

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type sfcIntentHandler struct {
	// Interface that implements SFC intent operations
	// We will set this variable with a mock interface for testing
	client module.SfcIntentManager
}
type sfcLinkIntentHandler struct {
	// Interface that implements SFC intent operations
	// We will set this variable with a mock interface for testing
	client module.SfcLinkIntentManager
}
type sfcClientSelectorIntentHandler struct {
	// Interface that implements SFC intent operations
	// We will set this variable with a mock interface for testing
	client module.SfcClientSelectorIntentManager
}
type sfcProviderNetworkIntentHandler struct {
	// Interface that implements SFC intent operations
	// We will set this variable with a mock interface for testing
	client module.SfcProviderNetworkIntentManager
}

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *module.SfcIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.SfcIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.SfcIntentManager)
			if ok {
				return c
			}
		}
	case *module.SfcClientSelectorIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.SfcClientSelectorIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.SfcClientSelectorIntentManager)
			if ok {
				return c
			}
		}
	case *module.SfcProviderNetworkIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.SfcProviderNetworkIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.SfcProviderNetworkIntentManager)
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
	moduleClient := module.NewClient()

	router := mux.NewRouter()
	v2Router := router.PathPrefix("/v2").Subrouter()

	const sfcIntentsURL = "/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-chains"
	const sfcIntentsGetURL = sfcIntentsURL + "/{sfcIntent}"
	const sfcLinkIntentsURL = sfcIntentsGetURL + "/links"
	const sfcLinkIntentsGetURL = sfcLinkIntentsURL + "/{sfcLink}"
	const sfcClientSelectorIntentsURL = sfcIntentsGetURL + "/client-selectors"
	const sfcClientSelectorIntentsGetURL = sfcClientSelectorIntentsURL + "/{sfcClientSelector}"
	const sfcProviderNetworkIntentsURL = sfcIntentsGetURL + "/provider-networks"
	const sfcProviderNetworkIntentsGetURL = sfcProviderNetworkIntentsURL + "/{sfcProviderNetwork}"

	sfcHandler := sfcIntentHandler{
		client: setClient(moduleClient.SfcIntent, testClient).(module.SfcIntentManager),
	}
	v2Router.HandleFunc(sfcIntentsURL, sfcHandler.createSfcHandler).Methods("POST")
	v2Router.HandleFunc(sfcIntentsURL, sfcHandler.getSfcHandler).Methods("GET")
	v2Router.HandleFunc(sfcIntentsGetURL, sfcHandler.putSfcHandler).Methods("PUT")
	v2Router.HandleFunc(sfcIntentsGetURL, sfcHandler.getSfcHandler).Methods("GET")
	v2Router.HandleFunc(sfcIntentsGetURL, sfcHandler.deleteSfcHandler).Methods("DELETE")

	sfcLinkHandler := sfcLinkIntentHandler{
		client: setClient(moduleClient.SfcLinkIntent, testClient).(module.SfcLinkIntentManager),
	}
	v2Router.HandleFunc(sfcLinkIntentsURL, sfcLinkHandler.createLinkHandler).Methods("POST")
	v2Router.HandleFunc(sfcLinkIntentsURL, sfcLinkHandler.getLinkHandler).Methods("GET")
	v2Router.HandleFunc(sfcLinkIntentsGetURL, sfcLinkHandler.putLinkHandler).Methods("PUT")
	v2Router.HandleFunc(sfcLinkIntentsGetURL, sfcLinkHandler.getLinkHandler).Methods("GET")
	v2Router.HandleFunc(sfcLinkIntentsGetURL, sfcLinkHandler.deleteLinkHandler).Methods("DELETE")

	sfcClientSelectorHandler := sfcClientSelectorIntentHandler{
		client: setClient(moduleClient.SfcClientSelectorIntent, testClient).(module.SfcClientSelectorIntentManager),
	}
	v2Router.HandleFunc(sfcClientSelectorIntentsURL, sfcClientSelectorHandler.createClientSelectorHandler).Methods("POST")
	v2Router.HandleFunc(sfcClientSelectorIntentsURL, sfcClientSelectorHandler.getClientSelectorHandler).Methods("GET")
	v2Router.HandleFunc(sfcClientSelectorIntentsGetURL, sfcClientSelectorHandler.putClientSelectorHandler).Methods("PUT")
	v2Router.HandleFunc(sfcClientSelectorIntentsGetURL, sfcClientSelectorHandler.getClientSelectorHandler).Methods("GET")
	v2Router.HandleFunc(sfcClientSelectorIntentsGetURL, sfcClientSelectorHandler.deleteClientSelectorHandler).Methods("DELETE")

	sfcProviderNetworkHandler := sfcProviderNetworkIntentHandler{
		client: setClient(moduleClient.SfcProviderNetworkIntent, testClient).(module.SfcProviderNetworkIntentManager),
	}
	v2Router.HandleFunc(sfcProviderNetworkIntentsURL, sfcProviderNetworkHandler.createProviderNetworkHandler).Methods("POST")
	v2Router.HandleFunc(sfcProviderNetworkIntentsURL, sfcProviderNetworkHandler.getProviderNetworkHandler).Methods("GET")
	v2Router.HandleFunc(sfcProviderNetworkIntentsGetURL, sfcProviderNetworkHandler.putProviderNetworkHandler).Methods("PUT")
	v2Router.HandleFunc(sfcProviderNetworkIntentsGetURL, sfcProviderNetworkHandler.getProviderNetworkHandler).Methods("GET")
	v2Router.HandleFunc(sfcProviderNetworkIntentsGetURL, sfcProviderNetworkHandler.deleteProviderNetworkHandler).Methods("DELETE")

	return router
}
