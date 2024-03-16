// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"reflect"

	"github.com/gorilla/mux"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	moduleLib "gitlab.com/project-emco/core/emco-base/src/migration/pkg/module"
)

var moduleClient *moduleLib.MigrationClient
var migrationIntentJSONFile string = "json-schemas/migration-intent.json"
var migrationAppIntentJSONFile string = "json-schemas/migration-app-intent.json"
var migrationPolicyJSONFile string = "json-schemas/migration-policy.json"

// MigrationPlacementIntentHandler .. Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type MigrationPlacementIntentHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client moduleLib.MigrationPlacementManager
}

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *moduleLib.MigrationClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.MigrationPlacementManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.MigrationPlacementManager)
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
	moduleClient = moduleLib.NewMigrationClient()

	router := mux.NewRouter()
	v2Router := router.PathPrefix("/v2").Subrouter()

	MigrationIntentHandler := MigrationPlacementIntentHandler{
		client: setClient(moduleClient, testClient).(moduleLib.MigrationPlacementManager),
	}

	const emcoMigrationIntentsURL = "/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/migration-intents"
	const emcoMigrationIntentsGetURL = emcoMigrationIntentsURL + "/{intent-name}"
	const emcoMigrationAppIntentsURL = emcoMigrationIntentsGetURL + "/app-intents"
	const emcoMigrationAppIntentsGetURL = emcoMigrationAppIntentsURL + "/{app-intent-name}"

	const emcoMigrationPoliciesURL = "/projects/{project-name}/migration-policies"
	const emcoMigrationPoliciesGetURL = emcoMigrationPoliciesURL + "/{policy-name}"

	// migration-intent => /projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/migration-intents
	v2Router.HandleFunc(emcoMigrationIntentsURL, MigrationIntentHandler.addMigrationIntentHandler).Methods("POST")
	v2Router.HandleFunc(emcoMigrationIntentsGetURL, MigrationIntentHandler.getMigrationIntentHandler).Methods("GET")
	v2Router.HandleFunc(emcoMigrationIntentsURL, MigrationIntentHandler.getMigrationIntentHandler).Methods("GET")
	v2Router.HandleFunc(emcoMigrationIntentsURL, MigrationIntentHandler.getMigrationIntentByNameHandler).Queries("intent", "{intent-name}")
	v2Router.HandleFunc(emcoMigrationIntentsGetURL, MigrationIntentHandler.putMigrationIntentHandler).Methods("PUT")
	v2Router.HandleFunc(emcoMigrationIntentsGetURL, MigrationIntentHandler.deleteMigrationIntentHandler).Methods("DELETE")
	v2Router.HandleFunc(emcoMigrationIntentsURL, MigrationIntentHandler.deleteAllMigrationIntentsHandler).Methods("DELETE")

	// migration-app-intent => /projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/migration-intents/{intent-name}/app-intents
	v2Router.HandleFunc(emcoMigrationAppIntentsURL, MigrationIntentHandler.addMigrationAppIntentHandler).Methods("POST")
	v2Router.HandleFunc(emcoMigrationAppIntentsGetURL, MigrationIntentHandler.getMigrationAppIntentHandler).Methods("GET")
	v2Router.HandleFunc(emcoMigrationAppIntentsURL, MigrationIntentHandler.getMigrationAppIntentHandler).Methods("GET")
	v2Router.HandleFunc(emcoMigrationAppIntentsURL, MigrationIntentHandler.getMigrationAppIntentByNameHandler).Queries("consumer", "{app-intent-name}")
	v2Router.HandleFunc(emcoMigrationAppIntentsGetURL, MigrationIntentHandler.putMigrationAppIntentHandler).Methods("PUT")
	v2Router.HandleFunc(emcoMigrationAppIntentsGetURL, MigrationIntentHandler.deleteMigrationAppIntentHandler).Methods("DELETE")
	v2Router.HandleFunc(emcoMigrationAppIntentsURL, MigrationIntentHandler.deleteAllMigrationAppIntentsHandler).Methods("DELETE")

	// migration-policy => /projects/{project-name}/migration-policies
	v2Router.HandleFunc(emcoMigrationPoliciesURL, MigrationIntentHandler.addMigrationPolicyHandler).Methods("POST")
	v2Router.HandleFunc(emcoMigrationPoliciesGetURL, MigrationIntentHandler.getMigrationPolicyHandler).Methods("GET")
	v2Router.HandleFunc(emcoMigrationPoliciesURL, MigrationIntentHandler.getMigrationPolicyHandler).Methods("GET")
	v2Router.HandleFunc(emcoMigrationPoliciesURL, MigrationIntentHandler.getMigrationPolicyByNameHandler).Queries("policy", "{policy-name}")
	v2Router.HandleFunc(emcoMigrationPoliciesGetURL, MigrationIntentHandler.putMigrationPolicyHandler).Methods("PUT")
	v2Router.HandleFunc(emcoMigrationPoliciesGetURL, MigrationIntentHandler.deleteMigrationPolicyHandler).Methods("DELETE")
	v2Router.HandleFunc(emcoMigrationPoliciesURL, MigrationIntentHandler.deleteAllMigrationPoliciesHandler).Methods("DELETE")

	return router
}
