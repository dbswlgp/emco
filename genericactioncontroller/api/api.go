// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"

	"fmt"
	"reflect"
)

// NewRouter returns the mux router after plugging in all the handlers
func NewRouter(mockClient interface{}) *mux.Router {
	const baseURL string = "/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-k8s-intents"

	client := module.NewClient()
	router := mux.NewRouter()
	v2Router := router.PathPrefix("/v2").Subrouter()

	genericK8sIntentHandler := genericK8sIntentHandler{
		client: setClient(client.GenericK8sIntent, mockClient).(module.GenericK8sIntentManager),
	}
	v2Router.HandleFunc(baseURL, genericK8sIntentHandler.handleGenericK8sIntentCreate).Methods("POST")
	v2Router.HandleFunc(baseURL, genericK8sIntentHandler.handleGenericK8sIntentGet).Methods("GET")
	v2Router.HandleFunc(baseURL+"/{genericK8sIntent}", genericK8sIntentHandler.handleGenericK8sIntentGet).Methods("GET")
	v2Router.HandleFunc(baseURL+"/{genericK8sIntent}", genericK8sIntentHandler.handleGenericK8sIntentUpdate).Methods("PUT")
	v2Router.HandleFunc(baseURL+"/{genericK8sIntent}", genericK8sIntentHandler.handleGenericK8sIntentDelete).Methods("DELETE")

	resourceHandler := resourceHandler{
		client: setClient(client.Resource, mockClient).(module.ResourceManager),
	}
	v2Router.HandleFunc(baseURL+"/{genericK8sIntent}/resources", resourceHandler.handleResourceCreate).Methods("POST")
	v2Router.HandleFunc(baseURL+"/{genericK8sIntent}/resources", resourceHandler.handleResourceGet).Methods("GET")
	v2Router.HandleFunc(baseURL+"/{genericK8sIntent}/resources/{genericResource}", resourceHandler.handleResourceGet).Methods("GET")
	v2Router.HandleFunc(baseURL+"/{genericK8sIntent}/resources/{genericResource}", resourceHandler.handleResourceUpdate).Methods("PUT")
	v2Router.HandleFunc(baseURL+"/{genericK8sIntent}/resources/{genericResource}", resourceHandler.handleResourceDelete).Methods("DELETE")

	customizationHandler := customizationHandler{
		client: setClient(client.Customization, mockClient).(module.CustomizationManager),
	}
	v2Router.HandleFunc(baseURL+"/{genericK8sIntent}/resources/{genericResource}/customizations", customizationHandler.handleCustomizationCreate).Methods("POST")
	v2Router.HandleFunc(baseURL+"/{genericK8sIntent}/resources/{genericResource}/customizations", customizationHandler.handleCustomizationGet).Methods("GET")
	v2Router.HandleFunc(baseURL+"/{genericK8sIntent}/resources/{genericResource}/customizations/{customization}", customizationHandler.handleCustomizationGet).Methods("GET")
	v2Router.HandleFunc(baseURL+"/{genericK8sIntent}/resources/{genericResource}/customizations/{customization}", customizationHandler.handleCustomizationUpdate).Methods("PUT")
	v2Router.HandleFunc(baseURL+"/{genericK8sIntent}/resources/{genericResource}/customizations/{customization}", customizationHandler.handleCustomizationDelete).Methods("DELETE")

	return router
}

// setClient set the client and its corresponding manager interface
// If the mockClient parameter is not nil and implements the manager interface
// corresponding to the client, return the mockClient. Otherwise, return the client
func setClient(client, mockClient interface{}) interface{} {
	switch cl := client.(type) {
	case *module.GenericK8sIntentClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*module.GenericK8sIntentManager)(nil)).Elem()) {
			c, ok := mockClient.(module.GenericK8sIntentManager)
			if ok {
				return c
			}
		}
	case *module.ResourceClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*module.ResourceManager)(nil)).Elem()) {
			c, ok := mockClient.(module.ResourceManager)
			if ok {
				return c
			}
		}

	case *module.CustomizationClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*module.CustomizationManager)(nil)).Elem()) {
			c, ok := mockClient.(module.CustomizationManager)
			if ok {
				return c
			}
		}
	default:
		fmt.Printf("unknown type %T\n", cl)
	}

	return client
}
