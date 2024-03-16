// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"encoding/json"
	"reflect"

	"context"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

// Customization holds the customization data
type Customization struct {
	Metadata types.Metadata    `json:"metadata"`
	Spec     CustomizationSpec `json:"spec"`
}

// CustomizationSpec holds the cluster-specific customization data
type CustomizationSpec struct {
	ClusterSpecific  string                   `json:"clusterSpecific"`
	ClusterInfo      ClusterInfo              `json:"clusterInfo"`
	PatchType        string                   `json:"patchType,omitempty"`
	PatchJSON        []map[string]interface{} `json:"patchJson,omitempty"`
	ConfigMapOptions ConfigMapOptions         `json:"configMapOptions,omitempty"`
	SecretOptions    SecretOptions            `json:"secretOptions,omitempty"`
}

// ClusterInfo holds the cluster info
type ClusterInfo struct {
	Scope           string `json:"scope"`
	ClusterProvider string `json:"clusterProvider"`
	ClusterName     string `json:"cluster"`
	ClusterLabel    string `json:"clusterLabel"`
	Mode            string `json:"mode"`
}

// ConfigMapOptions holds properties for customizing ConfigMap
type ConfigMapOptions struct {
	DataKeyOptions []KeyOptions `json:"dataKeyOptions,omitempty"`
}

// SecretOptions holds properties for customizing Secret
type SecretOptions struct {
	DataKeyOptions []KeyOptions `json:"dataKeyOptions,omitempty"`
}

// KeyOptions holds properties for customizing ConfigMap/Secret configuration data keys
type KeyOptions struct {
	FileName   string `json:"fileName,omitempty"`
	KeyName    string `json:"keyName,omitempty"`
	MergePatch string `json:"mergePatch,omitempty"` // indicates whether the customization files contain merge patch data
}

// Content holds the configuration data for a ConfigMap/Secret
type Content struct {
	FileName string
	Content  string
	KeyName  string
}

// CustomizationContent is a list of configuration data for a ConfigMap/Secret
type CustomizationContent struct {
	Content []Content
}

// CustomizationKey represents the resources associated with a Customization
type CustomizationKey struct {
	Customization         string `json:"customization"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeApp"`
	CompositeAppVersion   string `json:"compositeAppVersion"`
	DeploymentIntentGroup string `json:"deploymentIntentGroup"`
	GenericK8sIntent      string `json:"genericK8sIntent"`
	Resource              string `json:"genericResource"`
}

// CustomizationClient holds the client properties
type CustomizationClient struct {
	db ClientDbInfo
}

// Convert the key to string to preserve the underlying structure
func (k CustomizationKey) String() string {
	out, err := json.Marshal(k)
	if err != nil {
		return ""
	}
	return string(out)
}

// NewCustomizationClient returns an instance of the CustomizationClient which implements the Manager
func NewCustomizationClient() *CustomizationClient {
	return &CustomizationClient{
		db: ClientDbInfo{
			storeName:  "resources",
			tagMeta:    "data",
			tagContent: "customizationcontent",
		},
	}
}

// CustomizationManager exposes all the functionalities related to Customization
type CustomizationManager interface {
	CreateCustomization(ctx context.Context, customization Customization, content CustomizationContent,
		project, compositeApp, version, deploymentIntentGroup, intent, resource string,
		failIfExists bool) (Customization, bool, error)
	DeleteCustomization(ctx context.Context, customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) error
	GetAllCustomization(ctx context.Context, project, compositeApp, version, deploymentIntentGroup, intent, resource string) ([]Customization, error)
	GetCustomization(ctx context.Context, customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) (Customization, error)
	GetCustomizationContent(ctx context.Context, customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) (CustomizationContent, error)
}

// CreateCustomization creates a Customization
func (cc *CustomizationClient) CreateCustomization(ctx context.Context, customization Customization, customizationContent CustomizationContent,
	project, compositeApp, version, deploymentIntentGroup, intent, resource string,
	failIfExists bool) (Customization, bool, error) {

	cExists := false
	key := CustomizationKey{
		Customization:         customization.Metadata.Name,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   version,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      intent,
		Resource:              resource,
	}

	c, err := cc.GetCustomization(ctx,
		customization.Metadata.Name, project, compositeApp, version, deploymentIntentGroup, intent, resource)
	if err == nil &&
		!reflect.DeepEqual(c, Customization{}) {
		cExists = true
	}

	if cExists &&
		failIfExists {
		return Customization{}, cExists, emcoerror.NewEmcoError(
			CustomizationAlreadyExists,
			emcoerror.Conflict,
		)
	}

	if err = db.DBconn.Insert(ctx, cc.db.storeName, key, nil, cc.db.tagMeta, customization); err != nil {
		return Customization{}, cExists, err
	}

	if len(customizationContent.Content) > 0 {
		if err = db.DBconn.Insert(ctx, cc.db.storeName, key, nil, cc.db.tagContent, customizationContent); err != nil {
			return Customization{}, cExists, err
		}
	}

	return customization, cExists, nil
}

// GetCustomization returns a Customization
func (cc *CustomizationClient) GetCustomization(ctx context.Context,
	customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) (Customization, error) {

	key := CustomizationKey{
		Customization:         customization,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   version,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      intent,
		Resource:              resource,
	}

	value, err := db.DBconn.Find(ctx, cc.db.storeName, key, cc.db.tagMeta)
	if err != nil {
		return Customization{}, err
	}

	if len(value) == 0 {
		return Customization{}, emcoerror.NewEmcoError(
			CustomizationNotFound,
			emcoerror.NotFound,
		)
	}

	if value != nil {
		c := Customization{}
		if err = db.DBconn.Unmarshal(value[0], &c); err != nil {
			return Customization{}, err
		}
		return c, nil
	}

	return Customization{}, emcoerror.NewEmcoError(
		emcoerror.UnknownErrorMessage,
		emcoerror.Unknown,
	)
}

// GetAllCustomization returns all the Customizations for an Intent and Resource
func (cc *CustomizationClient) GetAllCustomization(ctx context.Context,
	project, compositeApp, version, deploymentIntentGroup, intent, resource string) ([]Customization, error) {

	key := CustomizationKey{
		Customization:         "",
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   version,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      intent,
		Resource:              resource,
	}

	values, err := db.DBconn.Find(ctx, cc.db.storeName, key, cc.db.tagMeta)
	if err != nil {
		return []Customization{}, err
	}

	var customizations []Customization
	for _, value := range values {
		c := Customization{}
		if err = db.DBconn.Unmarshal(value, &c); err != nil {
			return []Customization{}, err
		}
		customizations = append(customizations, c)
	}

	return customizations, nil
}

// GetCustomizationContent returns the content of the Customization files
func (cc *CustomizationClient) GetCustomizationContent(ctx context.Context,
	customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) (CustomizationContent, error) {

	key := CustomizationKey{
		Customization:         customization,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   version,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      intent,
		Resource:              resource,
	}

	value, err := db.DBconn.Find(ctx, cc.db.storeName, key, cc.db.tagContent)
	if err != nil {
		return CustomizationContent{}, err
	}

	if len(value) > 0 &&
		value[0] != nil {
		c := CustomizationContent{}
		if err = db.DBconn.Unmarshal(value[0], &c); err != nil {
			return CustomizationContent{}, err
		}
		return c, nil
	}

	return CustomizationContent{}, nil
}

// DeleteCustomization deletes a given Customization
func (cc *CustomizationClient) DeleteCustomization(ctx context.Context,
	customization, project, compositeApp, version, deploymentIntentGroup, intent, resource string) error {

	key := CustomizationKey{
		Customization:         customization,
		Project:               project,
		CompositeApp:          compositeApp,
		CompositeAppVersion:   version,
		DeploymentIntentGroup: deploymentIntentGroup,
		GenericK8sIntent:      intent,
		Resource:              resource,
	}

	return db.DBconn.Remove(ctx, cc.db.storeName, key)
}
