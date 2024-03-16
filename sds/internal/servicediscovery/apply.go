// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package servicediscovery

import (
	"encoding/json"
	"fmt"

	"context"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	rsyncclient "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/notifyclient"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/sds/internal/utils"
)

// DeployServiceEntry deploys service entry related resources on clusters
func DeployServiceEntry(ctx context.Context, ac appcontext.AppContext, appContextID string, serverName string, clientName string, serviceName string) error {

	// Check if the service is NodePort or LoadBalancer
	var errors error
	errors = nil

	// Get the clusters in the appcontext for this app
	clusters, errors := ac.GetClusterNames(ctx, serverName)
	if errors != nil {
		log.Error("Unable to get the cluster names",
			log.Fields{"AppName": serverName, "Error": errors})
		return pkgerrors.Wrap(errors, "Unable to get the cluster names")
	}

	for _, cluster := range clusters {

		rbValue, err := utils.GetClusterResources(ctx, appContextID, serverName, cluster)
		if err != nil {
			log.Error("Unable to get the cluster resources",
				log.Fields{"Cluster": cluster, "AppName": serverName, "Error": err})
			errors = err
			continue
		}

		se, err := getClusterServiceSpecs(ctx, ac, appContextID, rbValue, serviceName, serverName, cluster)
		if err != nil {
			log.Error("Unable to get the service specs",
				log.Fields{"Cluster": cluster, "AppName": serverName, "Error": err})
			errors = err
			continue
		}

		serviceData, err := createService(se)
		if err != nil {
			log.Error("Error Creating service YAML for service discovery",
				log.Fields{"AppName": serverName, "Error": err})
			return pkgerrors.Wrap(err, "Error Creating service YAML for service discovery")
		}

		endpointData, err := createEndpoint(se)
		if err != nil {
			log.Error("Error Creating endpoint YAML for service discovery",
				log.Fields{"AppName": serverName, "Error": err})
			return pkgerrors.Wrap(err, "Error Creating endpoint YAML for service discovery")
		}

		// Create child app context

		// Get the appcontext status value
		acStatus, err := state.GetAppContextStatus(ctx, appContextID)
		if err != nil {
			log.Error("Unable to get the status of the app context",
				log.Fields{"appContextID": appContextID, "Error": err})
			return pkgerrors.Wrap(err, "Unable to get the status of the app context")
		}

		if acStatus.Status == appcontext.AppContextStatusEnum.Instantiated {

			// Get the clusters in the appcontext for this app
			cs, err := ac.GetClusterNames(ctx, clientName)
			if err != nil {
				log.Error("Unable to get the cluster names",
					log.Fields{"AppName": clientName, "Error": err})
				return pkgerrors.Wrap(err, "Unable to get the cluster names")
			}

			// From this point on, we are dealing with a new app context (not "ac" from above)
			childContext := appcontext.AppContext{}
			childCtxVal, err := childContext.InitAppContext()
			if err != nil {
				log.Error("Error creating Child AppContext",
					log.Fields{"childContext": childContext, "Error": err})
				return pkgerrors.Wrap(err, "Error creating Child AppContext")
			}

			handle, err := childContext.CreateCompositeApp(ctx)
			if err != nil {
				log.Error("Error creating Child AppContext CompositeApp",
					log.Fields{"childContext": childContext, "Error": err})
				return pkgerrors.Wrap(err, "Error creating child AppContext CompositeApp")
			}

			appHandle, err := childContext.AddApp(ctx, handle, compositeApp)
			if err != nil {
				return utils.CleanupCompositeApp(ctx, childContext, err, "Error adding App to AppContext", []string{serviceName, childCtxVal.(string)})
			}

			// Iterate through cluster list and add all the clusters
			for _, c := range cs {
				if cluster == c {
					utils.CleanupCompositeApp(ctx, childContext, err,
						"Both server and client are deployed on same clusters, no need to deploy proxy service",
						[]string{serviceName, childCtxVal.(string)})
					continue
				}

				clusterHandle, err := childContext.AddCluster(ctx, appHandle, c)
				// pre-build array to pass to utils.CleanupCompositeApp() [for performance]
				details := []string{serviceName, c, childCtxVal.(string)}
				if err != nil {
					return utils.CleanupCompositeApp(ctx, childContext, err, "Error adding Cluster to AppContext", details)
				}

				// Add service k8s resource to each cluster
				appendedServiceName := serviceName + "+Service"
				_, err = childContext.AddResource(ctx, clusterHandle, appendedServiceName, serviceData)
				if err != nil {
					return utils.CleanupCompositeApp(ctx, childContext, err, "Error adding Namespace Resource to AppContext", details)
				}

				// Add endpoint k8s resource to each cluster
				appendedEndpointName := serviceName + "+Endpoint"
				_, err = childContext.AddResource(ctx, clusterHandle, appendedEndpointName, endpointData)
				if err != nil {
					return utils.CleanupCompositeApp(ctx, childContext, err, "Error adding Namespace Resource to AppContext", details)
				}

				// Add Resource Order and Resource Dependency
				resOrder, err := json.Marshal(map[string][]string{"resorder": {appendedEndpointName, appendedServiceName}})
				if err != nil {
					return utils.CleanupCompositeApp(ctx, childContext, err, "Error creating resource order JSON", details)
				}
				resDependency, err := json.Marshal(map[string]map[string]string{"resdependency": {appendedServiceName: "go", appendedEndpointName: "go"}})
				if err != nil {
					return utils.CleanupCompositeApp(ctx, childContext, err, "Error creating resource dependency JSON", details)
				}

				// Add App Order and App Dependency
				appOrder, err := json.Marshal(map[string][]string{"apporder": {compositeApp}})
				if err != nil {
					return utils.CleanupCompositeApp(ctx, childContext, err, "Error creating app order JSON", details)
				}
				appDependency, err := json.Marshal(map[string]map[string]string{"appdependency": {compositeApp: "go"}})
				if err != nil {
					return utils.CleanupCompositeApp(ctx, childContext, err, "Error creating app dependency JSON", details)
				}

				// Add Resource-level Order and Dependency
				_, err = childContext.AddInstruction(ctx, clusterHandle, "resource", "order", string(resOrder))
				if err != nil {
					return utils.CleanupCompositeApp(ctx, childContext, err, "Error adding instruction order to AppContext", details)
				}

				_, err = childContext.AddInstruction(ctx, clusterHandle, "resource", "dependency", string(resDependency))
				if err != nil {
					return utils.CleanupCompositeApp(ctx, childContext, err, "Error adding instruction dependency to AppContext", details)
				}

				// Add App-level Order and Dependency
				_, err = childContext.AddInstruction(ctx, handle, "app", "order", string(appOrder))
				if err != nil {
					return utils.CleanupCompositeApp(ctx, childContext, err, "Error adding app-level order to AppContext", details)
				}
				_, err = childContext.AddInstruction(ctx, handle, "app", "dependency", string(appDependency))
				if err != nil {
					return utils.CleanupCompositeApp(ctx, childContext, err, "Error adding app-level dependency to AppContext", details)
				}

			}

			// Get the parent composite app meta
			m, err := ac.GetCompositeAppMeta(ctx)
			if err != nil {
				return utils.CleanupCompositeApp(ctx, childContext, err, "Error getting CompositeAppMeta", []string{serviceName, childCtxVal.(string)})
			}

			err = childContext.AddCompositeAppMeta(ctx, appcontext.CompositeAppMeta{Project: m.Project, CompositeApp: compositeApp, Version: m.Version, Release: m.Release,
				DeploymentIntentGroup: m.DeploymentIntentGroup, Namespace: m.Namespace, Level: m.Level})
			if err != nil {
				return utils.CleanupCompositeApp(ctx, childContext, err, "Error adding CompositeAppMeta for child", []string{serviceName, childCtxVal.(string)})
			}

			childContextID := fmt.Sprintf("%v", childCtxVal)
			m.ChildContextIDs = append(m.ChildContextIDs, childContextID)
			// Add this child app context to Parent meta data
			err = ac.AddCompositeAppMeta(ctx, appcontext.CompositeAppMeta{
				Project:               m.Project,
				CompositeApp:          m.CompositeApp,
				Version:               m.Version,
				Release:               m.Release,
				DeploymentIntentGroup: m.DeploymentIntentGroup,
				Namespace:             m.Namespace,
				Level:                 m.Level,
				ChildContextIDs:       m.ChildContextIDs})
			if err != nil {
				return utils.CleanupCompositeApp(ctx, childContext, err, "Error adding CompositeAppMeta", []string{serviceName, childCtxVal.(string)})
			}

			// Check for parent app context status
			// Get the appcontext status value
			acStatus, err := state.GetAppContextStatus(ctx, appContextID)
			if err != nil {
				// Remove the child entry from the parent's meta
				utils.RemoveChildCtx(m.ChildContextIDs, childContextID)
				// Delete the child app context
				return utils.CleanupCompositeApp(ctx, childContext, err, "Unable to get the status of the app context", []string{serviceName, childCtxVal.(string)})
			}

			if acStatus.Status == appcontext.AppContextStatusEnum.Instantiated {

				// To-Do: race condition can be observed here. Will address this issue by mmodifying the rsync code to handle race condition issues.
				// Deploy the child app context
				err = rsyncclient.CallRsyncInstall(ctx, childCtxVal)
				if err != nil {
					// Remove the child entry from the parent's meta
					utils.RemoveChildCtx(m.ChildContextIDs, childContextID)
					// Delete the child app context
					return utils.CleanupCompositeApp(ctx, childContext, err, "Error calling rsync", []string{serviceName, childCtxVal.(string)})
				}
			} else {
				// Remove the child entry from the parent's meta
				utils.RemoveChildCtx(m.ChildContextIDs, childContextID)
				// Delete the child app context
				utils.CleanupCompositeApp(ctx, childContext, err, "Parent's app is not in instantiated state", []string{serviceName, childCtxVal.(string)})

			}

		}

	}

	return errors
}
