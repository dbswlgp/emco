// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
)

var CertificateSchemaJson string = "json-schemas/certificate.json"
var ClusterSchemaJson string = "json-schemas/cluster.json"
var LogicalCloudSchemaJson string = "json-schemas/logicalCloud.json"

// cpVars holds the clusterProvider caCert api route variables
type cpVars struct {
	cert,
	cluster,
	clusterProvider string
}

// lcVars holds the logicalCloud caCert api route variables
type lcVars struct {
	cert,
	cluster,
	logicalCloud,
	project string
}

// qParams holds the caCert status api query params
type qParams struct {
	qInstance,
	qType,
	qOutput string
	fApps,
	fClusters,
	fResources []string
	qApps,
	qClusters,
	qResources bool
}

// validateRequestBody validate the request body before storing it in the database
func validateRequestBody(r io.Reader, v interface{}, jsonSchema string) (int, error) {
	err := json.NewDecoder(r).Decode(&v)
	switch {
	case err == io.EOF:
		er := errors.New("empty request body")
		logutils.Error(er.Error(),
			logutils.Fields{
				"Error": err.Error()})
		return http.StatusBadRequest, er
	case err != nil:
		er := errors.New("error decoding the request body")
		logutils.Error(er.Error(),
			logutils.Fields{
				"Error": err.Error()})
		return http.StatusUnprocessableEntity, er
	}

	// validate the payload for the required values
	if err = validateData(v); err != nil {
		return http.StatusBadRequest, err
	}

	// ensure that the request body matches the schema defined in the JSON file
	err, code := validation.ValidateJsonSchemaData(jsonSchema, v)
	if err != nil {
		logutils.Error("Json schema validation failed",
			logutils.Fields{
				"JsonSchema": jsonSchema,
				"Error":      err.Error()})
		return code, err
	}

	return 0, nil
}

// validateData validate the payload for the required values
func validateData(i interface{}) error {
	switch p := i.(type) {
	case *module.CaCert:
		return validateCertData(*p)
	case *module.ClusterGroup:
		return validateClusterGroupData(*p)
	case *logicalcloud.CaCertLogicalCloud:
		return validateLogicalCloudData(*p)
	default:
		return nil
	}
}

// sendResponse sends an application/json response to the client
func sendResponse(w http.ResponseWriter, v interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logutils.Error("Failed to encode the response",
			logutils.Fields{
				"Error":    err,
				"Response": v})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// validateCertData validate the caCert payload for the required values
func validateCertData(cert module.CaCert) error {
	var err []string
	if len(cert.MetaData.Name) == 0 {
		er := "caCert name may not be empty"
		logutils.Error(er,
			logutils.Fields{})
		err = append(err, er)
	}

	if len(cert.Spec.IssuerRef.Name) == 0 {
		er := "issuer name may not be empty"
		logutils.Error(er,
			logutils.Fields{})
		err = append(err, er)
	}

	if len(cert.Spec.IssuerRef.Kind) == 0 {
		er := "issuer kind may not be empty"
		logutils.Error(er,
			logutils.Fields{})
		err = append(err, er)
	}

	if len(cert.Spec.IssuerRef.Group) == 0 &&
		len(cert.Spec.IssuerRef.Version) == 0 {
		er := "issuer group/version may not be empty"
		logutils.Error(er,
			logutils.Fields{})
		err = append(err, er)
	}

	if len(cert.Spec.IssuingCluster.Cluster) == 0 {
		er := "issuing cluster may not be empty"
		logutils.Error(er,
			logutils.Fields{})
		err = append(err, er)
	}

	if len(cert.Spec.IssuingCluster.ClusterProvider) == 0 {
		er := "issuing clusterProvider may not be empty"
		logutils.Error(er,
			logutils.Fields{})
		err = append(err, er)
	}

	if len(err) > 0 {
		return errors.New(strings.Join(err, "\n"))
	}

	return nil
}

// validateClusterGroupData validate the clusterGroup payload for the required values
func validateClusterGroupData(group module.ClusterGroup) error {
	var err []string
	if len(group.MetaData.Name) == 0 {
		er := "clusterGroup name may not be empty"
		logutils.Error(er,
			logutils.Fields{})
		err = append(err, er)
	}

	if len(group.Spec.Scope) == 0 {
		er := "scope may not be empty"
		logutils.Error(er,
			logutils.Fields{})
		err = append(err, er)
	}

	if len(group.Spec.Provider) == 0 {
		er := "clusterProvider may not be empty"
		logutils.Error(er,
			logutils.Fields{})
		err = append(err, er)
	}

	if group.Spec.Scope == "name" &&
		len(group.Spec.Cluster) == 0 {
		er := "cluster name may not be empty when the scope is the name"
		logutils.Error(er,
			logutils.Fields{})
		err = append(err, er)
	}

	if group.Spec.Scope == "label" &&
		len(group.Spec.Label) == 0 {
		er := "cluster label may not be empty when the scope is the label"
		logutils.Error(er,
			logutils.Fields{})
		err = append(err, er)
	}

	if len(err) > 0 {
		return errors.New(strings.Join(err, "\n"))
	}

	return nil
}

// _cpVars returns the clusterProvider caCert route variables for the current request
func _cpVars(vars map[string]string) cpVars {
	return cpVars{
		cert:            vars["caCert"],
		cluster:         vars["cluster"],
		clusterProvider: vars["clusterProvider"]}
}

// _lcVars returns the logicalCloud caCert route variables for the current request
func _lcVars(vars map[string]string) lcVars {
	return lcVars{
		cert:         vars["caCert"],
		cluster:      vars["cluster"],
		logicalCloud: vars["logicalCloud"],
		project:      vars["project"]}
}

// _statusQueryParams returns the caCert status query params for the current request
func _statusQueryParams(r *http.Request) (qParams, error) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return qParams{}, err
	}

	// initialize qParams with defaults
	qp := qParams{
		qInstance:  "",
		qType:      "ready",
		qOutput:    "all",
		fApps:      make([]string, 0),
		fClusters:  make([]string, 0),
		fResources: make([]string, 0),
		qApps:      false,
		qClusters:  false,
		qResources: false,
	}

	if o, found := params["instance"]; found {
		if o[0] == "" {
			err := errors.New("invalid query instance")
			logutils.Error(err.Error(),
				logutils.Fields{
					"Error": err.Error()})
			return qParams{}, err
		}
		qp.qInstance = o[0]
	}

	if t, found := params["status"]; found {
		if t[0] != "ready" && t[0] != "deployed" {
			err := errors.New("invalid query status")
			logutils.Error(err.Error(),
				logutils.Fields{
					"Error": err.Error()})
			return qParams{}, err
		}
		qp.qType = t[0]
	}

	if o, found := params["output"]; found {
		if o[0] != "summary" && o[0] != "all" && o[0] != "detail" {
			err := errors.New("invalid query output")
			logutils.Error(err.Error(),
				logutils.Fields{
					"Error": err.Error()})
			return qParams{}, err
		}
		qp.qOutput = o[0]
	}

	if _, found := params["clusters"]; found {
		qp.qClusters = true
	}

	if _, found := params["resources"]; found {
		qp.qResources = true
	}

	if c, found := params["cluster"]; found {
		for _, cl := range c {
			parts := strings.Split(cl, "+")
			if len(parts) != 2 {
				err := errors.New("invalid cluster query")
				logutils.Error(err.Error(),
					logutils.Fields{
						"Error": err.Error()})
				return qParams{}, err
			}

			for _, p := range parts {
				errs := validation.IsValidName(p)
				if len(errs) > 0 {
					err := errors.New("invalid cluster query")
					logutils.Error(err.Error(),
						logutils.Fields{
							"Error": err.Error()})
					return qParams{}, err
				}
			}
		}
		qp.fClusters = c
	}

	if r, found := params["resource"]; found {
		for _, res := range r {
			errs := validation.IsValidName(res)
			if len(errs) > 0 {
				err := errors.New("invalid resources query")
				logutils.Error(err.Error(),
					logutils.Fields{
						"Error": err.Error()})
				return qParams{}, err
			}
		}
		qp.fResources = r
	}

	return qp, nil
}
