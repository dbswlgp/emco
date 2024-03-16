// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"reflect"

	"github.com/gorilla/mux"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	moduleLib "gitlab.com/project-emco/core/emco-base/src/hpa-plc/pkg/module"
)

var moduleClient *moduleLib.HpaPlacementClient
var hpaIntentJSONFile string = "json-schemas/placement-hpa-intent.json"
var hpaConsumerJSONFile string = "json-schemas/placement-hpa-consumer.json"
var hpaResourceJSONFile string = "json-schemas/placement-hpa-resource.json"

// HpaPlacementIntentHandler .. Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type HpaPlacementIntentHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client moduleLib.HpaPlacementManager
}

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *moduleLib.HpaPlacementClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.HpaPlacementManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.HpaPlacementManager)
			if ok {
				return c
			}
		}
	default:
		log.Error(":: setClient .. unknown type ::", log.Fields{"client-type": cl})
	}
	return client
}

// NewRouter creates a router that registers the various urls that are supported
// testClient parameter allows unit testing for a given client
func NewRouter(testClient interface{}) *mux.Router {
	moduleClient = moduleLib.NewHpaPlacementClient()

	router := mux.NewRouter()
	v2Router := router.PathPrefix("/v2").Subrouter()

	hpaPlacementIntentHandler := HpaPlacementIntentHandler{
		client: setClient(moduleClient, testClient).(moduleLib.HpaPlacementManager),
	}

	const emcoHpaIntentsURL = "/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/hpa-intents"
	const emcoHpaIntentsGetURL = emcoHpaIntentsURL + "/{intent-name}"
	const emcoHpaConsumersURL = emcoHpaIntentsGetURL + "/hpa-resource-consumers"
	const emcoHpaConsumersGetURL = emcoHpaConsumersURL + "/{consumer-name}"
	const emcoHpaResourcesURL = emcoHpaConsumersGetURL + "/resource-requirements"
	const emcoHpaResourcesGetURL = emcoHpaResourcesURL + "/{resource-name}"

	// hpa-intent => /projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/hpa-intents
	v2Router.HandleFunc(emcoHpaIntentsURL, hpaPlacementIntentHandler.addHpaIntentHandler).Methods("POST")
	v2Router.HandleFunc(emcoHpaIntentsGetURL, hpaPlacementIntentHandler.getHpaIntentHandler).Methods("GET")
	v2Router.HandleFunc(emcoHpaIntentsURL, hpaPlacementIntentHandler.getHpaIntentHandler).Methods("GET")
	v2Router.HandleFunc(emcoHpaIntentsURL, hpaPlacementIntentHandler.getHpaIntentByNameHandler).Queries("intent", "{intent-name}")
	v2Router.HandleFunc(emcoHpaIntentsGetURL, hpaPlacementIntentHandler.putHpaIntentHandler).Methods("PUT")
	v2Router.HandleFunc(emcoHpaIntentsGetURL, hpaPlacementIntentHandler.deleteHpaIntentHandler).Methods("DELETE")
	v2Router.HandleFunc(emcoHpaIntentsURL, hpaPlacementIntentHandler.deleteAllHpaIntentsHandler).Methods("DELETE")

	// hpa-consumer => /projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers
	v2Router.HandleFunc(emcoHpaConsumersURL, hpaPlacementIntentHandler.addHpaConsumerHandler).Methods("POST")
	v2Router.HandleFunc(emcoHpaConsumersGetURL, hpaPlacementIntentHandler.getHpaConsumerHandler).Methods("GET")
	v2Router.HandleFunc(emcoHpaConsumersURL, hpaPlacementIntentHandler.getHpaConsumerHandler).Methods("GET")
	v2Router.HandleFunc(emcoHpaConsumersURL, hpaPlacementIntentHandler.getHpaConsumerHandlerByName).Queries("consumer", "{consumer-name}")
	v2Router.HandleFunc(emcoHpaConsumersGetURL, hpaPlacementIntentHandler.putHpaConsumerHandler).Methods("PUT")
	v2Router.HandleFunc(emcoHpaConsumersGetURL, hpaPlacementIntentHandler.deleteHpaConsumerHandler).Methods("DELETE")
	v2Router.HandleFunc(emcoHpaConsumersURL, hpaPlacementIntentHandler.deleteAllHpaConsumersHandler).Methods("DELETE")

	// hpa-resource => /projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers/{consumer-name}/resource-requirements
	v2Router.HandleFunc(emcoHpaResourcesURL, hpaPlacementIntentHandler.addHpaResourceHandler).Methods("POST")
	v2Router.HandleFunc(emcoHpaResourcesGetURL, hpaPlacementIntentHandler.getHpaResourceHandler).Methods("GET")
	v2Router.HandleFunc(emcoHpaResourcesURL, hpaPlacementIntentHandler.getHpaResourceHandler).Methods("GET")
	v2Router.HandleFunc(emcoHpaResourcesURL, hpaPlacementIntentHandler.getHpaResourceHandlerByName).Queries("resource", "{resource-name}")
	v2Router.HandleFunc(emcoHpaResourcesGetURL, hpaPlacementIntentHandler.putHpaResourceHandler).Methods("PUT")
	v2Router.HandleFunc(emcoHpaResourcesGetURL, hpaPlacementIntentHandler.deleteHpaResourceHandler).Methods("DELETE")
	v2Router.HandleFunc(emcoHpaResourcesURL, hpaPlacementIntentHandler.deleteAllHpaResourcesHandler).Methods("DELETE")

	return router
}
