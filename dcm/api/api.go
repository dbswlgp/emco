// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"gitlab.com/project-emco/core/emco-base/src/dcm/pkg/module"

	"github.com/gorilla/mux"
)

// NewRouter creates a router that registers the various urls that are
// supported
func NewRouter(
	logicalCloudClient module.LogicalCloudManager,
	clusterClient module.ClusterManager,
	userPermissionClient module.UserPermissionManager,
	quotaClient module.QuotaManager,
	keyValueClient module.KeyValueManager) *mux.Router {

	router := mux.NewRouter()

	// Set up Logical Cloud handler routes
	if logicalCloudClient == nil {
		logicalCloudClient = module.NewLogicalCloudClient()
	}

	if clusterClient == nil {
		clusterClient = module.NewClusterClient()
	}

	if quotaClient == nil {
		quotaClient = module.NewQuotaClient()
	}

	if userPermissionClient == nil {
		userPermissionClient = module.NewUserPermissionClient()
	}

	// Set up Logical Cloud API
	logicalCloudHandler := logicalCloudHandler{
		client:               logicalCloudClient,
		clusterClient:        clusterClient,
		quotaClient:          quotaClient,
		userPermissionClient: userPermissionClient,
	}
	lcRouter := router.PathPrefix("/v2/projects/{project}").Subrouter()
	lcRouter.HandleFunc(
		"/logical-clouds",
		logicalCloudHandler.createHandler).Methods("POST")
	lcRouter.HandleFunc(
		"/logical-clouds",
		logicalCloudHandler.getAllHandler).Methods("GET")
	lcRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}",
		logicalCloudHandler.getHandler).Methods("GET")
	lcRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}",
		logicalCloudHandler.deleteHandler).Methods("DELETE")
	lcRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}",
		logicalCloudHandler.putHandler).Methods("PUT")
	lcRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/instantiate",
		logicalCloudHandler.instantiateHandler).Methods("POST")
	lcRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/terminate",
		logicalCloudHandler.terminateHandler).Methods("POST")
	lcRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/update",
		logicalCloudHandler.updateHandler).Methods("POST")
	lcRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/stop",
		logicalCloudHandler.stopHandler).Methods("POST")
	lcRouter.HandleFunc("/logical-clouds/{logicalCloud}/status", logicalCloudHandler.statusHandler).Methods("GET")
	lcRouter.HandleFunc("/logical-clouds/{logicalCloud}/status",
		logicalCloudHandler.statusHandler).Queries("status", "{status}", "type", "{type}", "output", "{output}", "cluster", "{cluster}", "clusters", "{clusters}")

	// Set up Cluster API
	clusterHandler := clusterHandler{client: clusterClient}
	clusterRouter := router.PathPrefix("/v2/projects/{project}").Subrouter()
	clusterRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/cluster-references",
		clusterHandler.createHandler).Methods("POST")
	clusterRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/cluster-references",
		clusterHandler.getAllHandler).Methods("GET")
	clusterRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/cluster-references/{clusterReference}",
		clusterHandler.getHandler).Methods("GET")
	// clusterRouter.HandleFunc(
	// 	"/logical-clouds/{logicalCloud}/cluster-references/{clusterReference}",
	// 	clusterHandler.updateHandler).Methods("PUT")
	clusterRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/cluster-references/{clusterReference}",
		clusterHandler.deleteHandler).Methods("DELETE")
	clusterRouter.HandleFunc( // unsupported, developer-use only at the moment
		"/logical-clouds/{logicalCloud}/cluster-references/{clusterReference}/kubeconfig",
		clusterHandler.getConfigHandler).Methods("GET")

	userPermissionHandler := userPermissionHandler{client: userPermissionClient}
	upRouter := router.PathPrefix("/v2/projects/{project}").Subrouter()
	upRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/user-permissions",
		userPermissionHandler.createHandler).Methods("POST")
	upRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/user-permissions",
		userPermissionHandler.getAllHandler).Methods("GET")
	upRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/user-permissions/{userPermission}",
		userPermissionHandler.getHandler).Methods("GET")
	// upRouter.HandleFunc(
	// 	"/logical-clouds/{logicalCloud}/user-permissions/{userPermission}",
	// 	userPermissionHandler.updateHandler).Methods("PUT")
	upRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/user-permissions/{userPermission}",
		userPermissionHandler.deleteHandler).Methods("DELETE")

	// Set up Quota API
	quotaHandler := quotaHandler{client: quotaClient}
	quotaRouter := router.PathPrefix("/v2/projects/{project}").Subrouter()
	quotaRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/cluster-quotas",
		quotaHandler.createHandler).Methods("POST")
	quotaRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/cluster-quotas",
		quotaHandler.getAllHandler).Methods("GET")
	quotaRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/cluster-quotas/{clusterQuota}",
		quotaHandler.getHandler).Methods("GET")
	// quotaRouter.HandleFunc(
	// 	"/logical-clouds/{logicalCloud}/cluster-quotas/{clusterQuota}",
	// 	quotaHandler.updateHandler).Methods("PUT")
	quotaRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/cluster-quotas/{clusterQuota}",
		quotaHandler.deleteHandler).Methods("DELETE")

	// Set up Key Value API
	if keyValueClient == nil {
		keyValueClient = module.NewKeyValueClient()
	}
	keyValueHandler := keyValueHandler{client: keyValueClient}
	kvRouter := router.PathPrefix("/v2/projects/{project}").Subrouter()
	kvRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/kv-pairs",
		keyValueHandler.createHandler).Methods("POST")
	kvRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/kv-pairs",
		keyValueHandler.getAllHandler).Methods("GET")
	kvRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/kv-pairs/{logicalCloudKv}",
		keyValueHandler.getHandler).Methods("GET")
	// kvRouter.HandleFunc(
	// 	"/logical-clouds/{logicalCloud}/kv-pairs/{logicalCloudKv}",
	// 	keyValueHandler.updateHandler).Methods("PUT")
	kvRouter.HandleFunc(
		"/logical-clouds/{logicalCloud}/kv-pairs/{logicalCloudKv}",
		keyValueHandler.deleteHandler).Methods("DELETE")
	return router
}
