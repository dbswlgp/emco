// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"github.com/gorilla/mux"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	controller "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
)

var moduleClient *moduleLib.Client

// NewRouter creates a router that registers the various urls that are supported
func NewRouter(projectClient moduleLib.ProjectManager,
	compositeAppClient moduleLib.CompositeAppManager,
	appClient moduleLib.AppManager,
	ControllerClient controller.ControllerManager,
	genericPlacementIntentClient moduleLib.GenericPlacementIntentManager,
	appIntentClient moduleLib.AppIntentManager,
	deploymentIntentGrpClient moduleLib.DeploymentIntentGroupManager,
	intentClient moduleLib.IntentManager,
	compositeProfileClient moduleLib.CompositeProfileManager,
	appProfileClient moduleLib.AppProfileManager,
	instantiationClient moduleLib.InstantiationManager,
	appDependencyClient moduleLib.AppDependencyManager) *mux.Router {

	router := mux.NewRouter()

	v2Router := router.PathPrefix("/v2").Subrouter()

	moduleClient = moduleLib.NewClient()

	//setting routes for project
	if projectClient == nil {
		projectClient = moduleClient.Project

	}
	projHandler := projectHandler{
		client: projectClient,
	}
	if ControllerClient == nil {
		ControllerClient = moduleClient.Controller
	}
	controlHandler := controllerHandler{
		client: ControllerClient,
	}
	v2Router.HandleFunc("/projects", projHandler.createHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}", projHandler.updateHandler).Methods("PUT")
	v2Router.HandleFunc("/projects/{project}", projHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects", projHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}", projHandler.deleteHandler).Methods("DELETE")

	//setting routes for compositeApp
	if compositeAppClient == nil {
		compositeAppClient = moduleClient.CompositeApp
	}
	compAppHandler := compositeAppHandler{
		client: compositeAppClient,
	}
	v2Router.HandleFunc("/projects/{project}/composite-apps", compAppHandler.createHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}", compAppHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps", compAppHandler.getAllCompositeAppsHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}", compAppHandler.deleteHandler).Methods("DELETE")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}", compAppHandler.updateHandler).Methods("PUT")

	if appClient == nil {
		appClient = moduleClient.App
	}
	appHandler := appHandler{
		client: appClient,
	}

	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps", appHandler.createAppHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps/{app}", appHandler.getAppHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps/{app}", appHandler.updateAppHandler).Methods("PUT")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps", appHandler.getAppHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps/{app}", appHandler.deleteAppHandler).Methods("DELETE")

	if compositeProfileClient == nil {
		compositeProfileClient = moduleClient.CompositeProfile
	}
	compProfilepHandler := compositeProfileHandler{
		client: compositeProfileClient,
	}
	if appProfileClient == nil {
		appProfileClient = moduleClient.AppProfile
	}
	appProfileHandler := appProfileHandler{
		client: appProfileClient,
	}

	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles", compProfilepHandler.createHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles", compProfilepHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}", compProfilepHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}", compProfilepHandler.updateHandler).Methods("PUT")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}", compProfilepHandler.deleteHandler).Methods("DELETE")

	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles", appProfileHandler.createAppProfileHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles", appProfileHandler.getAppProfileHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles", appProfileHandler.getAppProfileHandler).Queries("app", "{app}")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles/{appProfile}", appProfileHandler.getAppProfileHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles/{appProfile}", appProfileHandler.updateAppProfileHandler).Methods("PUT")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles/{appProfile}", appProfileHandler.deleteAppProfileHandler).Methods("DELETE")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles", appProfileHandler.createAppProfileHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles", appProfileHandler.getAppProfileHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles", appProfileHandler.getAppProfileHandler).Queries("app", "{app}")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles/{appProfile}", appProfileHandler.getAppProfileHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/composite-profiles/{compositeProfile}/profiles/{appProfile}", appProfileHandler.deleteAppProfileHandler).Methods("DELETE")

	v2Router.HandleFunc("/controllers", controlHandler.createHandler).Methods("POST")
	v2Router.HandleFunc("/controllers", controlHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/controllers/{controller}", controlHandler.putHandler).Methods("PUT")
	v2Router.HandleFunc("/controllers/{controller}", controlHandler.getHandler).Methods("GET")
	v2Router.HandleFunc("/controllers/{controller}", controlHandler.deleteHandler).Methods("DELETE")

	//setting routes for genericPlacementIntent
	if genericPlacementIntentClient == nil {
		genericPlacementIntentClient = moduleClient.GenericPlacementIntent
	}

	genericPlacementIntentHandler := genericPlacementIntentHandler{
		client: genericPlacementIntentClient,
	}
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents", genericPlacementIntentHandler.createGenericPlacementIntentHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}", genericPlacementIntentHandler.getGenericPlacementHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents", genericPlacementIntentHandler.getAllGenericPlacementIntentsHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}", genericPlacementIntentHandler.deleteGenericPlacementHandler).Methods("DELETE")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}", genericPlacementIntentHandler.putGenericPlacementHandler).Methods("PUT")

	//setting routes for AppIntent
	if appIntentClient == nil {
		appIntentClient = moduleClient.AppIntent
	}

	appIntentHandler := appIntentHandler{
		client: appIntentClient,
	}

	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents", appIntentHandler.createAppIntentHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents/{genericAppPlacementIntent}", appIntentHandler.getAppIntentHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents", appIntentHandler.getAllAppIntentsHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents/", appIntentHandler.getAllIntentsByAppHandler).Queries("app", "{app}")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents/{genericAppPlacementIntent}", appIntentHandler.deleteAppIntentHandler).Methods("DELETE")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents/{genericAppPlacementIntent}", appIntentHandler.putAppIntentHandler).Methods("PUT")

	//setting routes for deploymentIntentGroup
	if deploymentIntentGrpClient == nil {
		deploymentIntentGrpClient = moduleClient.DeploymentIntentGroup
	}

	deploymentIntentGrpHandler := deploymentIntentGroupHandler{
		client: deploymentIntentGrpClient,
	}
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups", deploymentIntentGrpHandler.createDeploymentIntentGroupHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}", deploymentIntentGrpHandler.getDeploymentIntentGroupHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups", deploymentIntentGrpHandler.getAllDeploymentIntentGroupsHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}", deploymentIntentGrpHandler.deleteDeploymentIntentGroupHandler).Methods("DELETE")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}", deploymentIntentGrpHandler.putDeploymentIntentGroupHandler).Methods("PUT")

	// setting routes for AddingIntents
	if intentClient == nil {
		intentClient = moduleClient.Intent
	}

	intentHandler := intentHandler{
		client: intentClient,
	}

	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents", intentHandler.addIntentHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents/{groupIntent}", intentHandler.getIntentHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents", intentHandler.getAllIntentsHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents/", intentHandler.getIntentByNameHandler).Queries("intent", "{intent}")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents/{groupIntent}", intentHandler.deleteIntentHandler).Methods("DELETE")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/intents/{groupIntent}", intentHandler.putIntentHandler).Methods("PUT")

	// setting routes for Instantiation
	if instantiationClient == nil {
		instantiationClient = moduleClient.Instantiation
	}

	instantiationHandler := instantiationHandler{
		client: instantiationClient,
	}

	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/approve", instantiationHandler.approveHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/terminate", instantiationHandler.terminateHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/instantiate", instantiationHandler.instantiateHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/stop", instantiationHandler.stopHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/status", instantiationHandler.statusHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/status",
		instantiationHandler.statusHandler).Queries(
		"instance", "{instance}",
		"type", "{type}", // deprecated - to be replaced with "status" parameter
		"status", "{status}",
		"output", "{output}",
		"app", "{app}",
		"cluster", "{cluster}",
		"resource", "{resource}",
		"apps", "{apps}",
		"clusters", "{clusters}",
		"resources", "{resources}")

	// setting routes for Update
	updateHandler := updateHandler{
		client: instantiationClient,
	}

	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/migrate", updateHandler.migrateHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/update", updateHandler.updateHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/rollback", updateHandler.rollbackHandler).Methods("POST")

	if appDependencyClient == nil {
		appDependencyClient = moduleClient.AppDependency
	}
	appDependencyHandler := appDependencyHandler{
		client: appDependencyClient,
	}

	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps/{app}/dependency", appDependencyHandler.createAppDependencyHandler).Methods("POST")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps/{app}/dependency/{dependency}", appDependencyHandler.getAppDependencyHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps/{app}/dependency/{dependency}", appDependencyHandler.updateAppDependencyHandler).Methods("PUT")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps/{app}/dependency", appDependencyHandler.getAllAppDependencyHandler).Methods("GET")
	v2Router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/apps/{app}/dependency/{dependency}", appDependencyHandler.deleteappDependencyHandler).Methods("DELETE")

	return router
}
