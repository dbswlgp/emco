// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

var ResourceSchemaJson string = "json-schemas/resource.json"

// resourceHandler implements the handler functions
type resourceHandler struct {
	client module.ResourceManager
}

type rVars struct {
	compositeApp,
	version,
	deploymentIntentGroup,
	intent,
	resource,
	project string
}

// handleResourceCreate handles the route for creating a new Resource
func (h resourceHandler) handleResourceCreate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateResource(w, r)
}

// handleResourceDelete handles the route for deleting Resource from the database
func (h resourceHandler) handleResourceDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := _rVars(mux.Vars(r))
	if err := h.client.DeleteResource(ctx, vars.resource, vars.project, vars.compositeApp, vars.version,
		vars.deploymentIntentGroup, vars.intent); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleResourceGet handles the route for retrieving a Resource from the database
func (h resourceHandler) handleResourceGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := _rVars(mux.Vars(r))
	if len(vars.resource) == 0 {
		resources, err := h.client.GetAllResources(ctx, vars.project, vars.compositeApp, vars.version,
			vars.deploymentIntentGroup, vars.intent)
		if err != nil {
			apiErr := emcoerror.HandleAPIError(err)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}
		sendResponse(w, resources, http.StatusOK)
		return
	}

	resource, err := h.client.GetResource(ctx, vars.resource, vars.project, vars.compositeApp, vars.version,
		vars.deploymentIntentGroup, vars.intent)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	content, err := h.client.GetResourceContent(ctx, vars.resource, vars.project, vars.compositeApp, vars.version,
		vars.deploymentIntentGroup, vars.intent)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	var files []file
	f := file{
		Name:    "resourceTemplate",
		Content: content.Content,
	}
	files = append(files, f)

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
		sendMultipartResponse(w, resource, files, "resource")
		return

	case "application/json":
		sendResponse(w, resource, http.StatusOK)
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

// handleResourceUpdate handles the route for updating the existing Resource
func (h resourceHandler) handleResourceUpdate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateResource(w, r)
}

// createOrUpdateResource create/update the Resource based on the request method
func (h resourceHandler) createOrUpdateResource(w http.ResponseWriter, r *http.Request) {
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

	var resource module.Resource
	// the multipart/form-data should contain the key `metadata` with resource payload as the value
	data := bytes.NewBuffer([]byte(r.FormValue("metadata")))
	// validate the request body before storing it in the database
	if err := validateRequestBody(data, &resource, ResourceSchemaJson); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	// the multipart/form-data may contain the key `file` with the resource template file
	file, _, err := r.FormFile("file")
	if err != nil &&
		err != http.ErrMissingFile {
		log.Error("Unable to process the file",
			log.Fields{
				"Error": err.Error()})
		http.Error(w, "unable to process the file", http.StatusUnprocessableEntity)
		return
	}

	newObject := strings.ToLower(resource.Spec.NewObject)
	objectKind := strings.ToLower(resource.Spec.ResourceGVK.Kind)

	// if the template file is missing and the object is not a ConfigMap/Secret, return an error
	// a template file is mandatory for objects other than ConfigMap/Secret
	if err == http.ErrMissingFile &&
		objectKind != "configmap" &&
		objectKind != "secret" &&
		newObject == "true" {
		log.Error("The provided file field name is not present in the request or not a file field",
			log.Fields{
				"Error": err.Error()})
		http.Error(w, "the provided file field name is not present in the request or not a file field",
			http.StatusUnprocessableEntity)
		return
	}

	var resourceContent module.ResourceContent
	// create the resource content from the uploaded resource template file
	content, err := createResourceContent(w, file)
	if err != nil {
		return
	}
	resourceContent.Content = content

	methodPost := false
	if r.Method == http.MethodPost {
		methodPost = true
	}

	vars := _rVars(mux.Vars(r))
	res, rExists, err := h.client.CreateResource(ctx, resource, resourceContent,
		vars.project, vars.compositeApp, vars.version, vars.deploymentIntentGroup, vars.intent,
		methodPost)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	code := http.StatusCreated
	if rExists {
		// resource does have a current representation and that representation is successfully modified
		code = http.StatusOK
	}

	sendResponse(w, res, code)
}

// createResourceContent create the Resource content from the uploaded Resource template file
func createResourceContent(w http.ResponseWriter, file multipart.File) (string, error) {
	if file != nil {
		defer file.Close()
		data, err := ioutil.ReadAll(file)
		if err != nil {
			log.Error("Failed to read the resource template file",
				log.Fields{
					"Error": err.Error()})
			http.Error(w, "failed to read the resource template file", http.StatusUnprocessableEntity)
			return "", err
		}
		if len(data) > 0 {
			// validate the resource object
			if err = validateContent(data); err != nil {
				log.Error("Failed to validate the resource template file",
					log.Fields{
						"Error": err.Error()})
				http.Error(w, fmt.Sprintf("failed to validate the resource template file\n. Error validating data\n %s",
					err.Error()),
					http.StatusBadRequest)
				return "", err
			}

			return base64.StdEncoding.EncodeToString(data), nil
		}
	}

	return "", nil
}

// validateResourceData validate the Resource payload for the required values
func validateResourceData(r module.Resource) error {
	var err []string

	if len(r.Metadata.Name) == 0 {
		log.Error("Resource name may not be empty",
			log.Fields{})
		err = append(err, "resource name may not be empty")
	}

	if len(r.Spec.AppName) == 0 {
		log.Error("App may not be empty",
			log.Fields{})
		err = append(err, "app may not be empty")
	}

	if len(r.Spec.NewObject) == 0 {
		log.Error("NewObject may not be empty",
			log.Fields{})
		err = append(err, "newObject may not be empty")
	}

	if len(r.Spec.ResourceGVK.APIVersion) == 0 {
		log.Error("apiVersion may not be empty",
			log.Fields{})
		err = append(err, "apiVersion may not be empty")
	}

	if len(r.Spec.ResourceGVK.Kind) == 0 {
		log.Error("kind may not be empty",
			log.Fields{})
		err = append(err, "kind may not be empty")
	}

	if len(r.Spec.ResourceGVK.Name) == 0 {
		log.Error("Name may not be empty",
			log.Fields{})
		err = append(err, "name may not be empty")
	}

	if len(err) > 0 {
		return emcoerror.NewEmcoError(
			strings.Join(err, "\n"),
			emcoerror.BadRequest,
		)
	}

	return nil
}

// _rVars returns the route variables for the current request
func _rVars(vars map[string]string) rVars {
	return rVars{
		compositeApp:          vars["compositeApp"],
		deploymentIntentGroup: vars["deploymentIntentGroup"],
		intent:                vars["genericK8sIntent"],
		project:               vars["project"],
		resource:              vars["genericResource"],
		version:               vars["compositeAppVersion"],
	}
}
