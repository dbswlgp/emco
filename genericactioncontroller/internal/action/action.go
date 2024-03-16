// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package action

import (
	"strings"

	"context"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

// SEPARATOR used while creating resourceNames to store in etcd
const SEPARATOR = "+"

// UpdateOptions
type UpdateOptions struct {
	appContext           appcontext.AppContext
	appMeta              appcontext.CompositeAppMeta
	Customization        module.Customization
	customizations       []module.Customization
	CustomizationContent module.CustomizationContent
	intent               string
	ObjectKind           string
	Resource             module.Resource
	resourceContent      module.ResourceContent
}

// UpdateAppContext creates/updates the k8s resources in the given appContext and intent
func UpdateAppContext(ctx context.Context, intent, appContextID string) error {
	log.Info("Begin app context update",
		log.Fields{
			"AppContext": appContextID,
			"Intent":     intent})

	var appContext appcontext.AppContext
	if _, err := appContext.LoadAppContext(ctx, appContextID); err != nil {
		logError("failed to load appContext", appContextID, intent, appcontext.CompositeAppMeta{}, err)
		return err
	}

	appMeta, err := appContext.GetCompositeAppMeta(ctx)
	if err != nil {
		logError("failed to get compositeApp meta", appContextID, intent, appcontext.CompositeAppMeta{}, err)
		return err
	}

	resources, err := module.NewResourceClient().GetAllResources(ctx,
		appMeta.Project, appMeta.CompositeApp, appMeta.Version, appMeta.DeploymentIntentGroup, intent)
	if err != nil {
		logError("failed to get resources", appContextID, intent, appMeta, err)
		return err
	}

	o := UpdateOptions{
		appContext: appContext,
		appMeta:    appMeta,
		intent:     intent,
	}

	for _, o.Resource = range resources {
		o.ObjectKind = strings.ToLower(o.Resource.Spec.ResourceGVK.Kind)
		if err := o.createOrUpdateResource(ctx); err != nil {
			return err
		}
	}

	return nil
}

// createOrUpdateResource creates a new k8s object or updates the existing one
func (o *UpdateOptions) createOrUpdateResource(ctx context.Context) error {
	if err := o.getResourceContent(ctx); err != nil {
		return err
	}

	if err := o.getAllCustomization(ctx); err != nil {
		return err
	}

	for _, o.Customization = range o.customizations {
		if o.ObjectKind == "configmap" ||
			o.ObjectKind == "secret" ||
			o.Customization.Spec.PatchType == "merge" {
			if err := o.getCustomizationContent(ctx); err != nil {
				return err
			}
		}

		if strings.ToLower(o.Resource.Spec.NewObject) == "true" {
			if err := o.createNewResource(ctx); err != nil {
				return err
			}
			continue
		}

		if err := o.updateExistingResource(ctx); err != nil {
			return err
		}
	}

	return nil
}

// createNewResource creates a new k8s object
func (o *UpdateOptions) createNewResource(ctx context.Context) error {
	switch o.ObjectKind {
	case "configmap":
		if err := o.createConfigMap(ctx); err != nil {
			return err
		}
	case "secret":
		if err := o.createSecret(ctx); err != nil {
			return err
		}
	default:
		if err := o.createK8sResource(ctx); err != nil {
			return err
		}
	}

	return nil
}

// createK8sResource creates a new k8s object
func (o *UpdateOptions) createK8sResource(ctx context.Context) error {
	if len(o.resourceContent.Content) == 0 {
		o.logUpdateError(
			updateError{
				message: "resources content is empty"})
		return errors.New("resources content is empty")
	}

	// decode the template value
	value, err := decodeString(o.resourceContent.Content)
	if err != nil {
		return err
	}

	// apply the patch, if any
	if len(o.Customization.Spec.PatchType) > 0 {
		value, err = o.MergePatch(value)
		if err != nil {
			return err
		}
	}

	if err = o.create(ctx, value); err != nil {
		return err
	}

	return nil
}

// create adds the resource under the app and cluster
// also add instruction under the given handle and instruction type
func (o *UpdateOptions) create(ctx context.Context, data []byte) error {
	clusters, err := o.getClusterNames(ctx)
	if err != nil {
		return err
	}

	var (
		clusterName     string = o.Customization.Spec.ClusterInfo.ClusterName
		clusterSpecific string = strings.ToLower(o.Customization.Spec.ClusterSpecific)
		label           string = o.Customization.Spec.ClusterInfo.ClusterLabel
		mode            string = strings.ToLower(o.Customization.Spec.ClusterInfo.Mode)
		provider        string = o.Customization.Spec.ClusterInfo.ClusterProvider
		scope           string = strings.ToLower(o.Customization.Spec.ClusterInfo.Scope)
	)

	for _, cluster := range clusters {
		if clusterSpecific == "true" && scope == "label" {
			isValid, err := isValidClusterToApplyByLabel(ctx, provider, cluster, label, mode)
			if err != nil {
				return err
			}
			if !isValid {
				continue
			}
		}

		if clusterSpecific == "true" && scope == "name" {
			isValid, err := isValidClusterToApplyByName(ctx, provider, cluster, clusterName, mode)
			if err != nil {
				return err
			}
			if !isValid {
				continue
			}
		}

		handle, err := o.getClusterHandle(ctx, cluster)
		if err != nil {
			return err
		}

		resource := o.Resource.Spec.ResourceGVK.Name + SEPARATOR + o.Resource.Spec.ResourceGVK.Kind

		if err = o.addResource(ctx, handle, resource, string(data)); err != nil {
			return err
		}

		resorder, err := o.getResourceInstruction(ctx, cluster)
		if err != nil {
			return err
		}

		if err = o.addInstruction(ctx, handle, resorder, cluster, resource); err != nil {
			return err
		}

	}

	return nil
}

// updateExistingResource update the existing k8s object
func (o *UpdateOptions) updateExistingResource(ctx context.Context) error {
	// make sure a patch type is specified
	if len(o.Customization.Spec.PatchType) == 0 {
		return errors.New("patch type not defined")
	}

	clusters, err := o.getClusterNames(ctx)
	if err != nil {
		return err
	}

	var (
		modifiedPatch   []byte
		clusterName     string = o.Customization.Spec.ClusterInfo.ClusterName
		clusterSpecific string = strings.ToLower(o.Customization.Spec.ClusterSpecific)
		label           string = o.Customization.Spec.ClusterInfo.ClusterLabel
		mode            string = strings.ToLower(o.Customization.Spec.ClusterInfo.Mode)
		provider        string = o.Customization.Spec.ClusterInfo.ClusterProvider
		scope           string = strings.ToLower(o.Customization.Spec.ClusterInfo.Scope)
	)

	for _, cluster := range clusters {
		if clusterSpecific == "true" && scope == "label" {
			isValid, err := isValidClusterToApplyByLabel(ctx, provider, cluster, label, mode)
			if err != nil {
				return err
			}
			if !isValid {
				continue
			}
		}

		if clusterSpecific == "true" && scope == "name" {
			isValid, err := isValidClusterToApplyByName(ctx, provider, cluster, clusterName, mode)
			if err != nil {
				return err
			}
			if !isValid {
				continue
			}
		}

		handle, err := o.getResourceHandle(ctx, cluster, strings.Join([]string{o.Resource.Spec.ResourceGVK.Name,
			o.Resource.Spec.ResourceGVK.Kind}, SEPARATOR))
		if err != nil {
			continue
		}

		val, err := o.getValue(ctx, handle)
		if err != nil {
			continue
		}

		// apply the patch, if any
		modifiedPatch, err = o.MergePatch([]byte(val.(string)))
		if err != nil {
			return err
		}

		if err = o.updateResourceValue(ctx, handle, string(modifiedPatch)); err != nil {
			return err
		}
	}

	return nil
}
