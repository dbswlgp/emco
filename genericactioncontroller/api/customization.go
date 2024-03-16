// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"bytes"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

var CustomizationSchemaJson string = "json-schemas/customization.json"

// customizationHandler implements the handler functions
type customizationHandler struct {
	client module.CustomizationManager
}

type cVars struct {
	compositeApp,
	customization,
	deploymentIntentGroup,
	intent,
	project,
	resource,
	version string
}

// handleCustomizationCreate handles the route for creating a new Customization
func (h customizationHandler) handleCustomizationCreate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCustomization(w, r)
}

// handleCustomizationDelete handles the route for deleting Customization from the database
func (h customizationHandler) handleCustomizationDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := _cVars(mux.Vars(r))
	if err := h.client.DeleteCustomization(ctx, vars.customization, vars.project, vars.compositeApp,
		vars.version, vars.deploymentIntentGroup, vars.intent, vars.resource); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleCustomizationGet handles the route for retrieving a Customization from the database
func (h customizationHandler) handleCustomizationGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := _cVars(mux.Vars(r))
	if len(vars.customization) == 0 {
		customizations, err := h.client.GetAllCustomization(ctx, vars.project, vars.compositeApp,
			vars.version, vars.deploymentIntentGroup, vars.intent, vars.resource)
		if err != nil {
			apiErr := emcoerror.HandleAPIError(err)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}
		sendResponse(w, customizations, http.StatusOK)
		return
	}

	customization, err := h.client.GetCustomization(ctx, vars.customization, vars.project, vars.compositeApp,
		vars.version, vars.deploymentIntentGroup, vars.intent, vars.resource)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	content, err := h.client.GetCustomizationContent(ctx, vars.customization, vars.project, vars.compositeApp,
		vars.version, vars.deploymentIntentGroup, vars.intent, vars.resource)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	var files []file
	for _, p := range content.Content {
		f := file{
			Name:    p.FileName,
			Content: p.Content,
		}
		files = append(files, f)
	}

	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Accept"))
	if err != nil {
		log.Error("Failed to parse the media type",
			log.Fields{
				"Error": err.Error()})
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	switch mediaType {
	case "multipart/form-data":
		sendMultipartResponse(w, customization, files, "customization")
		return

	case "application/json":
		sendResponse(w, customization, http.StatusOK)
		return

	case "application/octet-stream":
		sendOctetStreamResponse(w, files)
		return

	default:
		log.Error("Set a media type. Set Accept header to  multipart/form-data, application/json or application/octet-stream",
			log.Fields{})
		http.Error(w, "set Accept header to multipart/form-data, application/json or application/octet-stream",
			http.StatusMultipleChoices)
		return
	}
}

// handleCustomizationUpdate handles the route for updating the existing Customization
func (h customizationHandler) handleCustomizationUpdate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCustomization(w, r)
}

// createOrUpdateCustomization create/update the Customization based on the request method
func (h customizationHandler) createOrUpdateCustomization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const maxMemory = 16777216 // set maxSize 16MB

	// parse the request body as multipart/form-data
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		log.Error("Failed to parse the multipart/form-data request body",
			log.Fields{
				"Error": err.Error()})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var customization module.Customization
	// the multipart/form-data should contain the key `metadata` with customization payload as the value
	data := bytes.NewBuffer([]byte(r.FormValue("metadata")))
	// validate the request body before storing it in the database
	if err := validateRequestBody(data, &customization, CustomizationSchemaJson); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	// validate customization specific prerequisites
	if err := validateCustomization(customization); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var customizationContent module.CustomizationContent
	// get the parsed multipart form, including file uploads
	form := r.MultipartForm
	// the multipart/form-data may contain the key `files` with the customization files
	fileHeaders := form.File["files"]
	if len(fileHeaders) > 0 {
		// parse each customization file attached in the request
		files, err := parseFile(fileHeaders)
		if err != nil {
			apiErr := emcoerror.HandleAPIError(err)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}

		// create customization content from the uploaded customization files
		customizationContent = createCustomizationContent(files)

		// handle configmap specific customizations
		handleConfigMapCustomization(customizationContent, customization.Spec.ConfigMapOptions)

		// handle secret specific customizations
		handleSecretCustomization(customizationContent, customization.Spec.SecretOptions)

	}

	vars := _cVars(mux.Vars(r))

	methodPost := false
	if r.Method == http.MethodPost {
		methodPost = true
	}

	c, cExists, err := h.client.CreateCustomization(ctx, customization, customizationContent,
		vars.project, vars.compositeApp, vars.version, vars.deploymentIntentGroup, vars.intent, vars.resource,
		methodPost)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	code := http.StatusCreated
	if cExists {
		// customization does have a current representation and that representation is successfully modified
		code = http.StatusOK
	}

	sendResponse(w, c, code)
}

// createCustomizationContent create the Customization content from the uploaded Customization files
func createCustomizationContent(files []file) module.CustomizationContent {
	var customizationContent module.CustomizationContent
	for _, f := range files {
		c := module.Content{
			Content:  f.Content,
			FileName: f.Name,
			KeyName:  f.Name, // default to filename
		}
		customizationContent.Content = append(customizationContent.Content, c)
	}
	return customizationContent
}

// handleConfigMapCustomization handles the ConfigMap specific customizations
func handleConfigMapCustomization(customizationContent module.CustomizationContent, configMapOptions module.ConfigMapOptions) error {
	if len(configMapOptions.DataKeyOptions) > 0 {
		if err := customizeDataKey(customizationContent, configMapOptions.DataKeyOptions); err != nil {
			return err
		}
	}

	return nil
}

// handleSecretCustomization handles the Secret specific customizations
func handleSecretCustomization(customizationContent module.CustomizationContent, secretOptions module.SecretOptions) error {
	if len(secretOptions.DataKeyOptions) > 0 {
		if err := customizeDataKey(customizationContent, secretOptions.DataKeyOptions); err != nil {
			return err
		}
	}

	return nil
}

// validateCustomization validate the Customization specific prerequisites
func validateCustomization(customization module.Customization) error {
	var err []string

	clusterSpecific := strings.ToLower(customization.Spec.ClusterSpecific)
	scope := strings.ToLower(customization.Spec.ClusterInfo.Scope)

	if clusterSpecific == "true" &&
		(module.ClusterInfo{}) == customization.Spec.ClusterInfo {
		log.Error("ClusterInfo is missing when ClusterSpecific is true",
			log.Fields{
				"CustomizationSpec": customization.Spec})
		err = append(err, "clusterInfo is missing")
	}

	if clusterSpecific == "true" &&
		scope == "label" &&
		len(customization.Spec.ClusterInfo.ClusterLabel) == 0 {
		log.Error("ClusterLabel is missing when ClusterSpecific is true and ClusterScope is label",
			log.Fields{
				"CustomizationSpec": customization.Spec})
		err = append(err, "clusterLabel is missing")
	}

	if clusterSpecific == "true" &&
		scope == "name" &&
		len(customization.Spec.ClusterInfo.ClusterName) == 0 {
		log.Error("ClusterName is missing when ClusterSpecific is true and ClusterScope is name",
			log.Fields{
				"CustomizationSpec": customization.Spec})
		err = append(err, "clusterName is missing")
	}

	if len(err) > 0 {
		return emcoerror.NewEmcoError(
			strings.Join(err, "\n"),
			emcoerror.BadRequest,
		)
	}

	return nil
}

// validateCustomizationData validate the Customization payload for the required values
func validateCustomizationData(customization module.Customization) error {
	var err []string

	if len(customization.Metadata.Name) == 0 {
		log.Error("Customization name may not be empty",
			log.Fields{})
		err = append(err, "customization name may not be empty")
	}

	if len(customization.Spec.ClusterSpecific) == 0 {
		log.Error("ClusterSpecific may not be empty",
			log.Fields{})
		err = append(err, "cluster specific may not be empty")
	}

	if len(customization.Spec.ClusterInfo.Scope) == 0 {
		log.Error("Scope may not be empty",
			log.Fields{})
		err = append(err, "scope may not be empty")
	}

	if len(customization.Spec.ClusterInfo.ClusterProvider) == 0 {
		log.Error("ClusterProvider may not be empty",
			log.Fields{})
		err = append(err, "cluster provider may not be empty")
	}

	if len(customization.Spec.ClusterInfo.Mode) == 0 {
		log.Error("Mode may not be empty",
			log.Fields{})
		err = append(err, "mode may not be empty")
	}

	for _, p := range customization.Spec.PatchJSON {
		switch value := p["value"].(type) {
		case string:
			if strings.HasPrefix(value, "$(http") &&
				strings.HasSuffix(value, ")$") {
				rawURL := strings.ReplaceAll(strings.ReplaceAll(value, "$(", ""), ")$", "")
				if _, e := url.ParseRequestURI(rawURL); e != nil {
					log.Error("Failed to parse the raw URL into a URL structure",
						log.Fields{
							"URL":   rawURL,
							"Error": e.Error()})
					err = append(err, e.Error())
				}
			}
		}
	}

	if len(err) > 0 {
		return emcoerror.NewEmcoError(
			strings.Join(err, "\n"),
			emcoerror.BadRequest,
		)
	}

	return nil
}

// _cVars returns the route variables for the current request
func _cVars(vars map[string]string) cVars {
	return cVars{
		compositeApp:          vars["compositeApp"],
		customization:         vars["customization"],
		deploymentIntentGroup: vars["deploymentIntentGroup"],
		intent:                vars["genericK8sIntent"],
		project:               vars["project"],
		resource:              vars["genericResource"],
		version:               vars["compositeAppVersion"],
	}
}
