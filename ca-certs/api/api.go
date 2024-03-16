// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
)

// NewRouter returns the mux router after plugging in all the handlers
func NewRouter(mockClient interface{}) *mux.Router {
	router := mux.NewRouter()
	r := route{
		router: router.PathPrefix("/v2").Subrouter(),
		client: client.NewClient(),
		mock:   mockClient}

	// set routes for adding caCert intent and clusterGroup(s) for clusterProvider scenario
	r.setClusterProviderRoutes()
	// set routes for adding caCert intent, logicalCloud(s) and clusterGroup(s) for logicalCloud scenario
	r.setLogicalCloudRoutes()

	return router
}

// setClient set the client and its corresponding manager interface
// If the mockClient parameter is not nil and implements the manager interface
// corresponding to the client, return the mockClient. Otherwise, return the client
func setClient(client, mockClient interface{}) interface{} {
	if mockClient == nil {
		return client
	}

	switch cl := client.(type) {
	case *clusterprovider.CaCertClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*clusterprovider.CaCertManager)(nil)).Elem()) {
			c, ok := mockClient.(clusterprovider.CaCertManager)
			if ok {
				return c
			}
		}

	case *clusterprovider.ClusterGroupClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*clusterprovider.ClusterGroupManager)(nil)).Elem()) {
			c, ok := mockClient.(clusterprovider.ClusterGroupManager)
			if ok {
				return c
			}
		}

	case *logicalcloud.CaCertClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*logicalcloud.CaCertManager)(nil)).Elem()) {
			c, ok := mockClient.(logicalcloud.CaCertManager)
			if ok {
				return c
			}
		}

	case *logicalcloud.CaCertLogicalCloudClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*logicalcloud.CaCertLogicalCloudManager)(nil)).Elem()) {
			c, ok := mockClient.(logicalcloud.CaCertLogicalCloudManager)
			if ok {
				return c
			}
		}
	case *logicalcloud.ClusterGroupClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*logicalcloud.ClusterGroupManager)(nil)).Elem()) {
			c, ok := mockClient.(logicalcloud.ClusterGroupManager)
			if ok {
				return c
			}
		}
	default:
		fmt.Printf("unknown type %T\n", cl)
	}

	return client
}
