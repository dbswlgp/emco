// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package action

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	yamlV2 "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/validation"
)

// ConfigMap holds the configuration data for pods to consume
type ConfigMap struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	MetaData   MetaData          `yaml:"metadata"`
	Data       map[string]string `yaml:"data,omitempty"`
}

// createConfigMap create the ConfigMap based on the JSON  patch,
// content in the template file, and the customization file, if any
func (o *UpdateOptions) createConfigMap(ctx context.Context) error {
	// create a new ConfigMap object based on the template file
	configMap, err := newConfigMap(o.resourceContent.Content, o.Resource.Spec.ResourceGVK.Name)
	if err != nil {
		return err
	}

	if len(o.CustomizationContent.Content) > 0 {
		// apply the customization data to the ConfigMap
		if err := handleConfigMapCustomization(configMap, o.CustomizationContent.Content,
			o.Customization.Spec.ConfigMapOptions.DataKeyOptions); err != nil {
			return err
		}
	}

	value, err := yamlV2.Marshal(configMap)
	if err != nil {
		logutils.Error("Failed to serialize the configMap object into a yaml document",
			logutils.Fields{
				"ConfigMap": configMap,
				"Error ":    err.Error()})
		return err
	}

	// apply the patch, if any
	if len(o.Customization.Spec.PatchType) > 0 {
		value, err = o.MergePatch(value)
		if err != nil {
			return err
		}
	}

	// create the ConfigMap
	if err = o.create(ctx, value); err != nil {
		return err
	}

	return nil
}

// newConfigMap creates a new ConfigMap object based on the template file
func newConfigMap(template, name string) (*ConfigMap, error) {
	if len(template) > 0 {
		// set the base struct from the associated template file
		value, err := base64.StdEncoding.DecodeString(template)
		if err != nil {
			logutils.Error("Failed to decode the configMap template content",
				logutils.Fields{
					"Error": err.Error()})
			return &ConfigMap{}, err
		}

		if len(value) > 0 {
			configMap := ConfigMap{}
			configMap.Data = map[string]string{} // initialize to avoid nil value
			if err = yamlV2.Unmarshal(value, &configMap); err != nil {
				logutils.Error("Failed to unmarshal the configMap template content",
					logutils.Fields{
						"Error": err.Error()})
				return &ConfigMap{}, err
			}

			if err = validateConfigMap(configMap); err != nil {
				return &configMap, err
			}

			return &configMap, nil
		}
	}

	// construct the ConfigMap base struct since there is no template associated with the ConfigMap
	return &ConfigMap{
		APIVersion: "v1",
		Kind:       "ConfigMap",
		MetaData: MetaData{
			Name: name,
		},
		Data: map[string]string{},
	}, nil
}

// handleConfigMapCustomization adds the specified customization data to the ConfigMap
func handleConfigMapCustomization(cm *ConfigMap, customizations []module.Content, dataKeyOptions []module.KeyOptions) error {
	// the number of customization file contents and filenames should be equal and in the same order
	for _, c := range customizations {
		mergePatch := false
		// exclude any patch content
		if len(dataKeyOptions) > 0 {
			for _, k := range dataKeyOptions {
				if c.FileName == k.FileName && strings.ToLower(k.MergePatch) == "true" {
					// this is a merge patch, not a customization data
					mergePatch = true
					break
				}
			}
		}

		if mergePatch {
			continue // continue with the next customization
		}

		// check whether the key name is valid
		if err := validateConfigMapDataKey(cm, c.KeyName); err != nil {
			return err
		}

		content, err := decodeString(c.Content)
		if err != nil {
			return err
		}

		cm.Data[c.KeyName] = string(content)
	}

	return nil
}

// validateConfigMapDataKey checks whether the data key name is valid
func validateConfigMapDataKey(cm *ConfigMap, key string) error {
	//  check whether the key is a valid key for the ConfigMap
	if errs := validation.IsConfigMapKey(key); len(errs) > 0 {
		logutils.Error("Invalid key",
			logutils.Fields{
				"ConfigMap": cm.MetaData.Name,
				"Key":       key,
				"Error":     strings.Join(errs, "\n")})
		return fmt.Errorf("%s is not a valid key name for a ConfigMap", key)
	}

	// check for duplicate key
	if _, exists := cm.Data[key]; exists {
		logutils.Error("Duplicate key",
			logutils.Fields{
				"ConfigMap": cm.MetaData.Name,
				"Key":       key,
				"Error":     "A key with the name already exists in Data"})
		return fmt.Errorf("a key with the name %s already exists in Data for ConfigMap %s", key, cm.MetaData.Name)
	}

	return nil
}

// validateConfigMap checks whether the ConfigMap has basic configurations
func validateConfigMap(cm ConfigMap) error {
	var err []string

	if len(cm.APIVersion) == 0 {
		err = append(err, "apiVersion not set for configmap")
	}

	if len(cm.Kind) == 0 ||
		strings.ToLower(cm.Kind) != "configmap" {
		err = append(err, "kind not set for configmap")
	}

	if len(cm.MetaData.Name) == 0 {
		err = append(err, "configmap name may not be empty")
	}

	if len(err) > 0 {
		return errors.New(strings.Join(err, "\n"))
	}

	return nil
}
