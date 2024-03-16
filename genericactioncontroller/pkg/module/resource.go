// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"encoding/json"
	"reflect"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

// Resource holds the resource data
type Resource struct {
	Metadata types.Metadata `json:"metadata"`
	Spec     ResourceSpec   `json:"spec"`
}

// ResourceSpec holds the Kubernetes object details and the app using the object
type ResourceSpec struct {
	AppName     string      `json:"app"`
	NewObject   string      `json:"newObject"`
	ResourceGVK ResourceGVK `json:"resourceGVK,omitempty"`
}

// ResourceGVK holds the Kubernetes object details
type ResourceGVK struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

// ResourceContent holds configuration data for the Kubernetes object
type ResourceContent struct {
	Content string `json:"filecontent"`
}

// ResourceKey represents the resources associated with a Resource
type ResourceKey struct {
	Resource              string `json:"genericResource"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeApp"`
	CompositeAppVersion   string `json:"compositeAppVersion"`
	DeploymentIntentGroup string `json:"deploymentIntentGroup"`
	GenericK8sIntent      string `json:"genericK8sIntent"`
}

// ResourceClient holds the client properties
type ResourceClient struct {
	db ClientDbInfo
}

// Convert the key to string to preserve the underlying structure
func (k ResourceKey) String() string {
	out, err := json.Marshal(k)
	if err != nil {
		return ""
	}
	return string(out)
}

// NewResourceClient returns an instance of the ResourceClient which implements the Manager
func NewResourceClient() *ResourceClient {
	return &ResourceClient{
		db: ClientDbInfo{
			storeName:  "resources",
			tagMeta:    "data",
			tagContent: "resourcecontent",
		},
	}
}

// ResourceManager exposes all the functionalities related to Resource
type ResourceManager interface {
	CreateResource(ctx context.Context, res Resource, resContent ResourceContent,
		project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string,
		failIfExists bool) (Resource, bool, error)
	DeleteResource(ctx context.Context, resource, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string) error
	GetAllResources(ctx context.Context, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string) ([]Resource, error)
	GetResource(ctx context.Context, resource, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string) (Resource, error)
	GetResourceContent(ctx context.Context, resource, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string) (ResourceContent, error)
}

// CreateResource creates a Resource
func (rc *ResourceClient) CreateResource(ctx context.Context, res Resource, resContent ResourceContent,
	project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string,
	failIfExists bool) (Resource, bool, error) {

	rExists := false
	key := ResourceKey{
		Resource:              res.Metadata.Name,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   compositeAppVersion,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      genericK8sIntent,
	}

	r, err := rc.GetResource(ctx, res.Metadata.Name, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent)
	if err == nil &&
		!reflect.DeepEqual(r, Resource{}) {
		rExists = true
	}

	if rExists &&
		failIfExists {
		return Resource{}, rExists, emcoerror.NewEmcoError(
			ResourceAlreadyExists,
			emcoerror.Conflict,
		)
	}

	if err = db.DBconn.Insert(ctx, rc.db.storeName, key, nil, rc.db.tagMeta, res); err != nil {
		return Resource{}, rExists, err
	}

	if len(resContent.Content) > 0 {
		if err = db.DBconn.Insert(ctx, rc.db.storeName, key, nil, rc.db.tagContent, resContent); err != nil {
			return Resource{}, rExists, err
		}
	}

	return res, rExists, nil
}

// GetResource returns a Resource
func (rc *ResourceClient) GetResource(ctx context.Context, resource, project, compositeApp, compositeAppVersion, deploymentIntentGroup, genericK8sIntent string) (Resource, error) {

	key := ResourceKey{
		Resource:              resource,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   compositeAppVersion,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      genericK8sIntent,
	}

	value, err := db.DBconn.Find(ctx, rc.db.storeName, key, rc.db.tagMeta)
	if err != nil {
		return Resource{}, err
	}

	if len(value) == 0 {
		return Resource{}, emcoerror.NewEmcoError(
			ResourceNotFound,
			emcoerror.NotFound,
		)
	}

	if value != nil {
		r := Resource{}
		if err = db.DBconn.Unmarshal(value[0], &r); err != nil {
			return Resource{}, err
		}
		return r, nil
	}

	return Resource{}, emcoerror.NewEmcoError(
		emcoerror.UnknownErrorMessage,
		emcoerror.Unknown,
	)
}

// GetAllResources returns all the Resources for an Intent
func (rc *ResourceClient) GetAllResources(ctx context.Context, project, compositeApp, compositeAppVersion, deploymentIntentGroup,
	genericK8sIntent string) ([]Resource, error) {

	key := ResourceKey{
		Resource:              "",
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   compositeAppVersion,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      genericK8sIntent,
	}

	values, err := db.DBconn.Find(ctx, rc.db.storeName, key, rc.db.tagMeta)
	if err != nil {
		return []Resource{}, err
	}

	var resources []Resource
	for _, value := range values {
		r := Resource{}
		if err = db.DBconn.Unmarshal(value, &r); err != nil {
			return []Resource{}, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

// GetResourceContent returns the content of the Resource template
func (rc *ResourceClient) GetResourceContent(ctx context.Context, resource, project, compositeApp, compositeAppVersion,
	deploymentIntentGroup, genericK8sIntent string) (ResourceContent, error) {

	key := ResourceKey{
		Resource:              resource,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   compositeAppVersion,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      genericK8sIntent,
	}

	value, err := db.DBconn.Find(ctx, rc.db.storeName, key, rc.db.tagContent)
	if err != nil {
		return ResourceContent{}, err
	}

	if len(value) > 0 &&
		value[0] != nil {
		c := ResourceContent{}
		if err = db.DBconn.Unmarshal(value[0], &c); err != nil {
			return ResourceContent{}, err
		}
		return c, nil
	}

	return ResourceContent{}, nil
}

// DeleteResource deletes a given Resource
func (rc *ResourceClient) DeleteResource(ctx context.Context, resource, project, compositeApp, compositeAppVersion,
	deploymentIntentGroup, genericK8sIntent string) error {

	key := ResourceKey{
		Resource:              resource,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   compositeAppVersion,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      genericK8sIntent,
	}

	return db.DBconn.Remove(ctx, rc.db.storeName, key)
}
