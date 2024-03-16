// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/sfcclient/pkg/module"
)

var moduleClient *module.SfcClient

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *module.SfcClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.SfcManager)(nil)).Elem()) {
			c, ok := testClient.(module.SfcManager)
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

	moduleClient = module.NewSfcClient()

	router := mux.NewRouter()
	v2Router := router.PathPrefix("/v2").Subrouter()

	const sfcClientIntentsURL = "/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/sfc-clients"
	const sfcClientIntentsGetURL = sfcClientIntentsURL + "/{sfcClientIntent}"

	sfcHandler := sfcHandler{
		client: setClient(moduleClient, testClient).(module.SfcManager),
	}
	v2Router.HandleFunc(sfcClientIntentsURL, sfcHandler.createHandler).Methods("POST")
	v2Router.HandleFunc(sfcClientIntentsURL, sfcHandler.getHandler).Methods("GET")
	v2Router.HandleFunc(sfcClientIntentsGetURL, sfcHandler.putHandler).Methods("PUT")
	v2Router.HandleFunc(sfcClientIntentsGetURL, sfcHandler.getHandler).Methods("GET")
	v2Router.HandleFunc(sfcClientIntentsGetURL, sfcHandler.deleteHandler).Methods("DELETE")

	return router
}
