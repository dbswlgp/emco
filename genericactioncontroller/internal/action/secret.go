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

// Secret holds secret data of a certain type
type Secret struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	MetaData   MetaData          `yaml:"metadata"`
	Type       string            `yaml:"type"`
	Data       map[string]string `yaml:"data"`
}

// createSecret create the Secret based on the JSON  patch,
// content in the template file, and the customization file, if any
func (o *UpdateOptions) createSecret(ctx context.Context) error {
	// create a new Secret object based on the template file
	secret, err := newSecret(o.resourceContent.Content, o.Resource.Spec.ResourceGVK.Name)
	if err != nil {
		return err
	}

	if len(o.CustomizationContent.Content) > 0 {
		// apply the customization data to the Secret
		if err := handleSecretCustomization(secret, o.CustomizationContent.Content,
			o.Customization.Spec.SecretOptions.DataKeyOptions); err != nil {
			return err
		}
	}

	value, err := yamlV2.Marshal(secret)
	if err != nil {
		logutils.Error("Failed to serialize the secret object into a YAML document",
			logutils.Fields{
				"Secret": secret,
				"Error ": err.Error()})
		return err
	}

	// apply the patch, if any
	if len(o.Customization.Spec.PatchType) > 0 {
		value, err = o.MergePatch(value)
		if err != nil {
			return err
		}
	}

	// create the Secret
	if err = o.create(ctx, value); err != nil {
		return err
	}

	return nil
}

// newSecret creates a new Secret object based on the template file
func newSecret(template, name string) (*Secret, error) {
	if len(template) > 0 {
		// set the base struct from the associated template file
		value, err := base64.StdEncoding.DecodeString(template)
		if err != nil {
			logutils.Error("Failed to decode the secret template content",
				logutils.Fields{
					"Error": err.Error()})
			return &Secret{}, err
		}

		if len(value) > 0 {
			secret := Secret{}
			secret.Data = map[string]string{} // initialize to avoid nil value
			if err = yamlV2.Unmarshal(value, &secret); err != nil {
				logutils.Error("Failed to unmarshal the secret template content",
					logutils.Fields{
						"Error": err.Error()})
				return &Secret{}, err
			}

			if len(secret.Type) == 0 {
				secret.Type = "Opaque"
			}

			if err = validateSecret(secret); err != nil {
				return &secret, err
			}

			return &secret, nil
		}
	}

	// construct the Secret base struct since there is no template associated with the Secret
	return &Secret{
		APIVersion: "v1",
		Kind:       "Secret",
		Type:       "Opaque",
		MetaData: MetaData{
			Name: name,
		},
		Data: map[string]string{},
	}, nil
}

// handleSecretCustomization adds the specified customization data to the Secret
func handleSecretCustomization(s *Secret, customizations []module.Content, dataKeyOptions []module.KeyOptions) error {
	// the number of customization file contents and filenames should be equal and in the same order
	for _, c := range customizations {
		mergePatch := false
		// exclude any merge patch content
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
		if err := validateSecretDataKey(s, c.KeyName); err != nil {
			return err
		}

		s.Data[c.KeyName] = string([]byte(c.Content))
	}

	return nil
}

// validateSecretDataKey checks whether the data key name is valid
func validateSecretDataKey(s *Secret, key string) error {
	//  check whether the key is a valid key for the Secret
	if errs := validation.IsConfigMapKey(key); len(errs) > 0 {
		logutils.Error("Invalid key",
			logutils.Fields{
				"Secret": s.MetaData.Name,
				"Key":    key,
				"Error":  strings.Join(errs, "\n")})
		return fmt.Errorf("%s is not a valid key name for a Secret", key)
	}

	// check for duplicate key
	if _, exists := s.Data[key]; exists {
		logutils.Error("Duplicate key",
			logutils.Fields{
				"Secret": s.MetaData.Name,
				"Key":    key,
				"Error":  "A key with the name already exists in Data"})
		return fmt.Errorf("a key with the name %s already exists in Data for Secret %s", key, s.MetaData.Name)
	}

	return nil
}

// validateSecret checks whether the Secret has basic configurations
func validateSecret(s Secret) error {
	var err []string

	if len(s.APIVersion) == 0 {
		err = append(err, "apiVersion not set for secret")
	}

	if len(s.Kind) == 0 ||
		strings.ToLower(s.Kind) != "secret" {
		err = append(err, "kind not set for secret")
	}

	if len(s.MetaData.Name) == 0 {
		err = append(err, "secret name may not be empty")
	}

	if len(err) > 0 {
		return errors.New(strings.Join(err, "\n"))
	}

	return nil
}
