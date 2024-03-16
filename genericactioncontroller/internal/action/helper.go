// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package action

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"context"

	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/yaml"
)

// MetaData holds the object Name, Namespace and Annotations
type MetaData struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type updateError struct {
	message, cluster string
	handle           interface{}
	err              error
}

// isValidClusterToApplyByLabel checks if a given cluster falls under the given label and provider
func isValidClusterToApplyByLabel(ctx context.Context, provider, clusterName, clusterLabel, mode string) (bool, error) {
	clusters, err := cluster.NewClusterClient().GetClustersWithLabel(ctx, provider, clusterLabel)
	if err != nil {
		log.Error("Failed to get clusters by the provider and label",
			log.Fields{
				"Provider":                provider,
				"AutheticatingForCluster": clusterName,
				"ClusterLabel":            clusterLabel,
				"Mode":                    mode})
		return false, err
	}

	clusterName = strings.Split(clusterName, SEPARATOR)[1]
	for _, c := range clusters {
		if c == clusterName && mode == "allow" {
			return true, nil
		}
	}

	return false, nil
}

// isValidClusterToApplyByName checks if a given cluster under a provider matches with the cluster which is authenticated for
func isValidClusterToApplyByName(ctx context.Context, provider, authenticatedCluster, givenCluster, mode string) (bool, error) {
	clusters, err := cluster.NewClusterClient().GetClusters(ctx, provider)
	if err != nil {
		log.Error("Failed to get clusters by the provider",
			log.Fields{
				"Provider":                provider,
				"GivenCluster":            givenCluster,
				"AutheticatingForCluster": authenticatedCluster,
				"Mode":                    mode,
				"Error":                   err.Error()})
		return false, err
	}

	authenticatedCluster = strings.Split(authenticatedCluster, SEPARATOR)[1]
	for _, c := range clusters {
		if c.Metadata.Name == authenticatedCluster && authenticatedCluster == givenCluster && mode == "allow" {
			return true, nil
		}
	}

	return false, nil
}

// decodeString returns the bytes represented by the base64 string s
func decodeString(s string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Error("Failed to decode the base64 string",
			log.Fields{
				"Error": err.Error()})
		return []byte{}, err
	}

	return data, nil
}

// getResourceContent retrieves the content of the Resource template from the db
func (o *UpdateOptions) getResourceContent(ctx context.Context) error {
	resourceContent, err := module.NewResourceClient().GetResourceContent(ctx, o.Resource.Metadata.Name, o.appMeta.Project,
		o.appMeta.CompositeApp, o.appMeta.Version, o.appMeta.DeploymentIntentGroup, o.intent)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get resource content",
				err:     err})
		return err
	}

	o.resourceContent = resourceContent

	return nil
}

// getAllCustomization returns all the Customizations for the given Intent and Resource
func (o *UpdateOptions) getAllCustomization(ctx context.Context) error {
	customizations, err := module.NewCustomizationClient().GetAllCustomization(ctx, o.appMeta.Project,
		o.appMeta.CompositeApp, o.appMeta.Version, o.appMeta.DeploymentIntentGroup, o.intent, o.Resource.Metadata.Name)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get customizations",
				err:     err})
		return err
	}

	if len(customizations) == 0 {
		log.Warn("No customization is available for the resource",
			log.Fields{
				"Resource": o.Resource.Metadata.Name})
	}

	o.customizations = customizations

	return nil
}

// getCustomizationContent retrieves the content of the Customization files from the db
func (o *UpdateOptions) getCustomizationContent(ctx context.Context) error {
	customizationContent, err := module.NewCustomizationClient().GetCustomizationContent(ctx, o.Customization.Metadata.Name, o.appMeta.Project,
		o.appMeta.CompositeApp, o.appMeta.Version, o.appMeta.DeploymentIntentGroup, o.intent, o.Resource.Metadata.Name)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get customization content",
				err:     err})
		return err
	}

	o.CustomizationContent = customizationContent

	return nil
}

// getClusterNames returns a list of all clusters for a given app
func (o *UpdateOptions) getClusterNames(ctx context.Context) ([]string, error) {
	clusters, err := o.appContext.GetClusterNames(ctx, o.Resource.Spec.AppName)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get cluster names",
				err:     err})
		return []string{}, err
	}

	return clusters, nil
}

// getClusterHandle returns the handle for a given app and cluster
func (o *UpdateOptions) getClusterHandle(ctx context.Context, cluster string) (interface{}, error) {
	handle, err := o.appContext.GetClusterHandle(ctx, o.Resource.Spec.AppName, cluster)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get cluster handle",
				cluster: cluster,
				err:     err})
		return nil, err
	}

	return handle, nil
}

// addResource add the resource under the app and cluster
func (o *UpdateOptions) addResource(ctx context.Context, handle interface{}, resource, value string) error {
	if _, err := o.appContext.AddResource(ctx, handle, resource, value); err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to add the resource",
				handle:  handle,
				err:     err})
		return err
	}

	return nil
}

// getResourceInstruction returns the resource instruction for a given instruction type
func (o *UpdateOptions) getResourceInstruction(ctx context.Context, cluster string) (interface{}, error) {
	resorder, err := o.appContext.GetResourceInstruction(ctx, o.Resource.Spec.AppName, cluster, "order")
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get resource instruction",
				cluster: cluster,
				err:     err})
		return nil, err
	}

	return resorder, nil
}

// addInstruction add instruction under the given handle and instruction type
func (o *UpdateOptions) addInstruction(ctx context.Context, handle, resorder interface{}, cluster, resource string) error {
	v := make(map[string][]string)
	json.Unmarshal([]byte(resorder.(string)), &v)
	v["resorder"] = append(v["resorder"], resource)
	data, _ := json.Marshal(v)
	if _, err := o.appContext.AddInstruction(ctx, handle, "resource", "order", string(data)); err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to add instruction",
				cluster: cluster,
				handle:  handle,
				err:     err})
		return err
	}

	return nil
}

// getResourceHandle returns the handle for the given app, cluster, and resource
func (o *UpdateOptions) getResourceHandle(ctx context.Context, cluster, resource string) (interface{}, error) {
	handle, err := o.appContext.GetResourceHandle(ctx, o.Resource.Spec.AppName, cluster, resource)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get resource handle",
				cluster: cluster,
				err:     err})
		return nil, err
	}

	return handle, nil
}

// getValue returns the value for a given handle
func (o *UpdateOptions) getValue(ctx context.Context, handle interface{}) (interface{}, error) {
	val, err := o.appContext.GetValue(ctx, handle)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get handle value",
				handle:  handle,
				err:     err})
		return nil, err
	}

	log.Info("Manifest file for the resource",
		log.Fields{
			"Resource":      o.Resource.Spec.ResourceGVK.Name,
			"Manifest-File": val.(string)})

	return val, nil
}

// updateResourceValue updates the resource value using the given handle
func (o *UpdateOptions) updateResourceValue(ctx context.Context, handle interface{}, value string) error {
	if err := o.appContext.UpdateResourceValue(ctx, handle, value); err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to update resource value",
				handle:  handle,
				err:     err})
		return err
	}

	log.Info("Resource updated in appContext",
		log.Fields{
			"AppName":      o.Resource.Spec.AppName,
			"AppMeta":      o.appMeta,
			"Intent":       o.intent,
			"resourceName": o.Resource.Metadata.Name})

	return nil
}

// logError writes the error details to the log
func logError(message, appContext, intent string, appMeta appcontext.CompositeAppMeta, err error) {
	log.Error(message,
		log.Fields{
			"AppContext": appContext,
			"AppMeta":    appMeta,
			"Intent":     intent,
			"Error":      err.Error()})
}

// logUpdateError writes the update errors to the log
func (o *UpdateOptions) logUpdateError(uError updateError) {
	fields := make(log.Fields)
	fields["AppMeta"] = o.appMeta
	if len(o.Resource.Spec.AppName) > 0 {
		fields["AppName"] = o.Resource.Spec.AppName
	}
	if len(uError.cluster) > 0 {
		fields["Clsuter"] = uError.cluster
	}
	if len(o.Customization.Metadata.Name) > 0 {
		fields["Customization"] = o.Customization.Metadata.Name
	}
	if uError.err != nil {
		fields["Error"] = uError.err.Error()
	}
	if uError.handle != nil {
		fields["Handle"] = uError.handle
	}
	fields["Intent"] = o.intent
	if len(o.Resource.Spec.ResourceGVK.Kind) > 0 {
		fields["Kind"] = o.Resource.Spec.ResourceGVK.Kind
	}
	if len(o.Resource.Metadata.Name) > 0 {
		fields["Resource"] = o.Resource.Metadata.Name
	}

	log.Error(uError.message, fields)

}

// getResourceStructFromGVK returns the resource object struct
func getResourceStructFromGVK(apiVersion, kind string) (runtime.Object, error) {
	resourceGVK := schema.GroupVersionKind{Kind: kind}
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		log.Error("Failed to get the group version struct",
			log.Fields{
				"APIVersion": apiVersion,
				"Kind":       kind,
				"Error":      err.Error()})
	}

	resourceGVK = schema.GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: kind}
	obj, err := scheme.Scheme.New(resourceGVK)
	if err != nil {
		log.Error("Failed to get the resource object struct using the resource GVK",
			log.Fields{
				"APIVersion": apiVersion,
				"Kind":       kind,
				"Error":      err.Error()})

	}

	return obj, err
}

// yamlToJson convert yaml document to json
func yamlToJson(y []byte) ([]byte, error) {
	data, err := yaml.YAMLToJSON(y)
	if err != nil {
		logutils.Error("Failed to convert yaml to json document",
			logutils.Fields{
				"Error": err.Error(),
			})
		return []byte{}, err
	}

	return data, nil
}

// jsonToYaml convert json document to yaml
func jsonToYaml(j []byte) ([]byte, error) {
	data, err := yaml.JSONToYAML(j)
	if err != nil {
		logutils.Error("Failed to convert json to yaml document",
			logutils.Fields{
				"Error": err.Error(),
			})
		return []byte{}, err
	}

	return data, nil
}
