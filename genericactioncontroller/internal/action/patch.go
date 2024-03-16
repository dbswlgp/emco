// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package action

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// applyJSONPatch reconciles a modified configuration with an original configuration
func applyJSONPatch(patch []map[string]interface{}, original []byte) ([]byte, error) {
	patchData, err := json.MarshalIndent(patch, "", " ")
	if err != nil {
		logutils.Error("Failed to marshal the customization json patch",
			logutils.Fields{
				"Patch": patch,
				"Error": err.Error()})
		return nil, err
	}

	originalData, err := yamlToJson(original)
	if err != nil {
		return []byte{}, err
	}

	decodedPatch, err := jsonpatch.DecodePatch([]byte(patchData))
	if err != nil {
		logutils.Error("Failed to decode the customization json patch data",
			logutils.Fields{
				"Error": err.Error(),
			})
		return []byte{}, err
	}

	modifiedData, err := decodedPatch.Apply(originalData)
	if err != nil {
		logutils.Error("Failed to apply the customization json patch data",
			logutils.Fields{
				"Error": err.Error()})
		return []byte{}, err
	}

	modifiedPatch, err := jsonToYaml(modifiedData)
	if err != nil {
		return []byte{}, err
	}

	return modifiedPatch, nil
}

// validateJSONPatchValue looks for any HTTP URL in the JSON patch value
// and replace it with the URL response, if needed
func (o *UpdateOptions) validateJSONPatchValue() error {
	var (
		err          []string
		placeholders = []string{"{clusterProvider}", "{cluster}"} // supported placeholders in the URL
	)

	for _, p := range o.Customization.Spec.PatchJSON {
		switch value := p["value"].(type) {
		case string:
			if strings.HasPrefix(value, "$(http") &&
				strings.HasSuffix(value, ")$") {
				// replace the patch value with the URL response
				rawURL := strings.ReplaceAll(strings.ReplaceAll(value, "$(", ""), ")$", "")
				if strings.Contains(rawURL, "/{") {
					// look for placeholders in the URL and replace it, if needed
					for _, ph := range placeholders {
						if strings.Contains(rawURL, ph) {
							switch {
							case ph == "{clusterProvider}":
								rawURL = strings.Replace(rawURL, ph, o.Customization.Spec.ClusterInfo.ClusterProvider, -1) // -1-> replace all the instances
							case ph == "{cluster}":
								rawURL = strings.Replace(rawURL, ph, o.Customization.Spec.ClusterInfo.ClusterName, -1) // -1-> replace all the instances
							}
						}
					}
				}

				val, e := getJSONPatchValueFromExternalService(rawURL)
				if e != nil {
					err = append(err, e.Error())
					continue // verify the value for all the patches and capture errors if there are any
				}
				// update the patch value with the response
				p["value"] = val
			}
		}
	}

	if len(err) > 0 {
		return errors.New(strings.Join(err, "\n"))
	}

	return nil
}

// getJSONPatchValueFromExternalService invoke the URL and returns the value
func getJSONPatchValueFromExternalService(rawURL string) (interface{}, error) {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		logutils.Error("Failed to parse the raw URL into a URL structure",
			logutils.Fields{
				"URL":   rawURL,
				"Error": err.Error()})
		return nil, err
	}

	resp, err := http.Get(u.String())
	if err != nil {
		logutils.Error("Failed to get the URL response",
			logutils.Fields{
				"URL":   u.String(),
				"Error": err.Error()})
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logutils.Error("Unexpected status code when reading patch value from the URL",
			logutils.Fields{
				"URL":        u.String(),
				"Status":     resp.Status,
				"StatusCode": resp.StatusCode})
		return nil, fmt.Errorf("unexpected status code when reading patch value from %s. response: %v, code: %d", u.String(), resp.Status, resp.StatusCode)
	}

	var v map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	if _, exist := v["value"]; !exist {
		return nil, fmt.Errorf("unexpected patch value from %s. response: %v", u.String(), v)
	}

	return v["value"], nil
}

// MergePatch merge the original document with the patch content based on the patch type
func (o *UpdateOptions) MergePatch(original []byte) ([]byte, error) {
	switch strings.ToLower(o.Customization.Spec.PatchType) {
	case "json":
		// make sure we have a valid JSON patch to update the resource
		if len(o.Customization.Spec.PatchJSON) == 0 {
			o.logUpdateError(
				updateError{
					message: "invalid json patch"})
			return []byte{}, errors.New("invalid json patch")
		}

		// validate the JSON patch value before applying
		if err := o.validateJSONPatchValue(); err != nil {
			return []byte{}, err
		}

		return applyJSONPatch(o.Customization.Spec.PatchJSON, original)

	case "merge":
		// make sure we have the cutomization files
		if len(o.CustomizationContent.Content) == 0 {
			return []byte{}, errors.New("no patch file")
		}

		modifiedData, err := yamlToJson(original)
		if err != nil {
			return []byte{}, err
		}

		var (
			content string
			val     []byte
		)

		for _, c := range o.CustomizationContent.Content {
			// you can use the customization files to specify the merge patch content for Kubernetes resources,
			// including configmap/secret

			// check whether the customization content is a merge patch or not
			switch o.ObjectKind {
			case "configmap":
				for _, k := range o.Customization.Spec.ConfigMapOptions.DataKeyOptions {
					if c.FileName == k.FileName && strings.ToLower(k.MergePatch) == "true" {
						content = c.Content
						break
					}
				}
			case "secret":
				for _, k := range o.Customization.Spec.SecretOptions.DataKeyOptions {
					if c.FileName == k.FileName && strings.ToLower(k.MergePatch) == "true" {
						content = c.Content
						break
					}
				}

			default:
				content = c.Content // customization content for any other resources should be the merge patch
			}

			if len(content) > 0 {
				data, err := decodeString(content)
				if err != nil {
					return []byte{}, err
				}

				patch, err := yamlToJson(data)
				if err != nil {
					return []byte{}, err
				}

				ds, err := getResourceStructFromGVK(o.Resource.Spec.ResourceGVK.APIVersion, o.Resource.Spec.ResourceGVK.Kind)
				if err != nil {
					return []byte{}, err
				}

				val, err = strategicpatch.StrategicMergePatch(modifiedData, patch, ds)
				if err != nil {
					return []byte{}, err
				}

				modifiedData = val
			}
		}

		modifiedPatch, err := jsonToYaml(modifiedData)
		if err != nil {
			return []byte{}, err
		}

		return modifiedPatch, nil
	}

	return []byte{}, errors.New("patch type not supported")
}
