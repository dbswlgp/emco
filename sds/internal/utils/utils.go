// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package utils

import (
	"context"
	"encoding/json"

	pkgerrors "github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	rb "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

const rsyncName = "rsync"

// CheckDeploymentStatus will check the deployment resource status of the app
func CheckDeploymentStatus(ctx context.Context, appContextID string, serverApp string) (bool, error) {

	deploymentState := false

	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(ctx, appContextID)
	if err != nil {
		log.Error("Error loading AppContext", log.Fields{
			"error": err,
		})
		return deploymentState, pkgerrors.Wrapf(err, "Error getting AppContext with Id: %v", appContextID)
	}

	// Get the clusters in the appcontext for this app
	clusters, err := ac.GetClusterNames(ctx, serverApp)
	if err != nil {
		log.Error("Unable to get the cluster names",
			log.Fields{"AppName": serverApp, "Error": err})
		return deploymentState, pkgerrors.Wrapf(err, "Unable to get the cluster names for the app - %v", serverApp)
	}

	for _, cluster := range clusters {

		rbValue, err := GetClusterResources(ctx, appContextID, serverApp, cluster)
		if err != nil {
			log.Error("Unable to get the cluster resources",
				log.Fields{"Cluster": cluster, "AppName": serverApp, "Error": err})
			deploymentState = false
			continue
		}

		// Get the parent composite app meta
		m, err := ac.GetCompositeAppMeta(ctx)
		if err != nil {
			log.Error("Error getting CompositeAppMeta",
				log.Fields{"Cluster": cluster, "AppName": serverApp, "Error": err})
			deploymentState = false
			continue
		}

		// Append the deployment intent group release version
		deploymentName := m.Release + "-" + serverApp

		deploymentState, err = getClusterDeploymentStatus(rbValue, deploymentName)
		if err != nil {
			log.Error("Error gathering cluster deployment status",
				log.Fields{"Cluster": cluster, "AppName": serverApp, "Error": err})
			deploymentState = false
			continue
		}
	}

	return deploymentState, nil
}

// GetClusterResources will retrieve the cluster resources
func GetClusterResources(ctx context.Context, appContextID string, app string, cluster string) (*rb.ResourceBundleStateStatus, error) {

	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(ctx, appContextID)
	if err != nil {
		log.Error("Error loading AppContext", log.Fields{
			"error": err,
		})
		return nil, err
	}

	csh, err := ac.GetClusterStatusHandle(ctx, app, cluster)
	if err != nil {
		log.Error("No cluster status handle for cluster, app",
			log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
		return nil, err
	}
	clusterRbValue, err := ac.GetValue(ctx, csh)
	if err != nil {
		log.Error("No cluster status value for cluster, app",
			log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
		return nil, err
	}
	var rbValue rb.ResourceBundleStateStatus
	err = json.Unmarshal([]byte(clusterRbValue.(string)), &rbValue)
	if err != nil {
		log.Error("Error unmarshalling cluster status value for cluster, app",
			log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
		return nil, err
	}

	return &rbValue, nil
}

// getClusterDeploymentStatus takes in a ResourceBundleStateStatus CR and returns the status of the deployment resource
func getClusterDeploymentStatus(rbData *rb.ResourceBundleStateStatus, deploymentName string) (bool, error) {

	deploymentStatus := false
	for _, d := range rbData.DeploymentStatuses {
		if !CompareResource(d.Name, deploymentName) {
			continue
		}

		for _, condition := range d.Status.Conditions {
			if (condition.Type == appsv1.DeploymentAvailable || condition.Type == appsv1.DeploymentProgressing) &&
				condition.Status == corev1.ConditionTrue {
				// If the deployment is in active/ready state then continue to check other deployments
				deploymentStatus = true
				break
			} else {
				// deployment is not in active/ready state
				deploymentStatus = false
				return deploymentStatus, nil
			}
		}

	}

	return deploymentStatus, nil
}

// CompareResource compares the input resource name with the monitor resource name
func CompareResource(r string, qResource string) bool {

	if r == qResource {
		return true
	}
	return false
}

// CleanupCompositeApp will delete the app context
func CleanupCompositeApp(ctx context.Context, appCtx appcontext.AppContext, err error, reason string, details []string) error {
	if err == nil {
		// create an error object to avoid wrap failures
		err = pkgerrors.New("Composite App cleanup.")
	}

	cleanuperr := appCtx.DeleteCompositeApp(ctx)
	newerr := pkgerrors.Wrap(err, reason)
	if cleanuperr != nil {
		log.Warn("Error cleaning AppContext, ", log.Fields{
			"Related details": details,
		})
		return pkgerrors.Wrap(err, "After previous error, cleaning the AppContext also failed.")
	}
	return newerr
}

// RemoveChildCtx removes the child appCtx ID in the parent's meta
func RemoveChildCtx(childContexts []string, childContextID string) {
	for i := range childContexts {
		if childContexts[i] == childContextID {
			childContexts[i] = childContexts[len(childContexts)-1]
			childContexts[len(childContexts)-1] = ""
			childContexts = childContexts[:len(childContexts)-1]
		}
	}
}
