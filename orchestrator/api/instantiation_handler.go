// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"fmt"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type instantiationHandler struct {
	client moduleLib.InstantiationManager
}

func (h instantiationHandler) approveHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	vars := mux.Vars(r)
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	di := vars["deploymentIntentGroup"]

	iErr := h.client.Approve(ctx, p, ca, v, di)
	if iErr != nil {
		log.Error(iErr.Error(), log.Fields{})
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)

}

func (h instantiationHandler) instantiateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	fmt.Println(vars)
	fmt.Println("")
	fmt.Println("")

	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	di := vars["deploymentIntentGroup"]

	iErr := h.client.Instantiate(ctx, p, ca, v, di)
	if iErr != nil {
		log.Error(":: Error instantiate handler ::", log.Fields{"Error": iErr.Error(), "project": p, "compositeApp": ca, "compositeAppVer": v, "depGroup": di})
		apiErr := apierror.HandleLogicalCloudErrors(vars, iErr, lcErrors)
		if (apiErr == apierror.APIError{}) {
			// There are no logical cloud error(s). Check for api specific error(s)
			apiErr = apierror.HandleErrors(vars, iErr, nil, apiErrors)
		}
		if apiErr.Status == http.StatusInternalServerError {
			http.Error(w, pkgerrors.Cause(iErr).Error(), apiErr.Status)
		} else {
			http.Error(w, apiErr.Message, apiErr.Status)
		}
		return
	}
	log.Info("instantiateHandler ... end ", log.Fields{"project": p, "compositeApp": ca, "compositeAppVer": v, "depGroup": di, "returnValue": iErr})
	w.WriteHeader(http.StatusAccepted)
}

func (h instantiationHandler) terminateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	di := vars["deploymentIntentGroup"]

	iErr := h.client.Terminate(ctx, p, ca, v, di)
	if iErr != nil {
		log.Error(iErr.Error(), log.Fields{})
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)

}

func (h instantiationHandler) stopHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	di := vars["deploymentIntentGroup"]

	iErr := h.client.Stop(ctx, p, ca, v, di)
	if iErr != nil {
		log.Error(iErr.Error(), log.Fields{})
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)

}

func (h instantiationHandler) statusHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	vars := mux.Vars(r)
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	di := vars["deploymentIntentGroup"]

	qParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var queryInstance string
	if o, found := qParams["instance"]; found {
		queryInstance = o[0]
		if queryInstance == "" {
			log.Error("Invalid query instance", log.Fields{})
			http.Error(w, "Invalid query instance", http.StatusBadRequest)
			return
		}
	} else {
		queryInstance = "" // default instance value
	}

	var queryType string
	// the type= parameter is deprecated - and will be replaced by the status flag
	if t, found := qParams["type"]; found {
		queryType = t[0]
		if queryType != "cluster" && queryType != "rsync" {
			log.Error("Invalid query type", log.Fields{})
			http.Error(w, "Invalid query type", http.StatusBadRequest)
			return
		}
	} else {
		queryType = "rsync" // default type (deprecated)
	}
	if t, found := qParams["status"]; found {
		queryType = t[0]
		if queryType != "ready" && queryType != "deployed" {
			log.Error("Invalid query status", log.Fields{})
			http.Error(w, "Invalid query status", http.StatusBadRequest)
			return
		}
	}

	var queryOutput string
	if o, found := qParams["output"]; found {
		queryOutput = o[0]
		if queryOutput != "summary" && queryOutput != "all" && queryOutput != "detail" {
			log.Error("Invalid query output", log.Fields{})
			http.Error(w, "Invalid query output", http.StatusBadRequest)
			return
		}
	} else {
		queryOutput = "all" // default output format
	}

	var queryApps bool
	if _, found := qParams["apps"]; found {
		queryApps = true
	} else {
		queryApps = false
	}

	var queryClusters bool
	if _, found := qParams["clusters"]; found {
		queryClusters = true
	} else {
		queryClusters = false
	}

	var queryResources bool
	if _, found := qParams["resources"]; found {
		queryResources = true
	} else {
		queryResources = false
	}

	var filterApps []string
	if a, found := qParams["app"]; found {
		filterApps = a
		for _, app := range filterApps {
			errs := validation.IsValidName(app)
			if len(errs) > 0 {
				log.Error(errs[len(errs)-1], log.Fields{}) // log the most recently appended msg
				http.Error(w, "Invalid app query", http.StatusBadRequest)
				return
			}
		}
	} else {
		filterApps = make([]string, 0)
	}

	var filterClusters []string
	if c, found := qParams["cluster"]; found {
		filterClusters = c
		for _, cl := range filterClusters {
			parts := strings.Split(cl, "+")
			if len(parts) != 2 {
				log.Error("Invalid cluster query", log.Fields{})
				http.Error(w, "Invalid cluster query", http.StatusBadRequest)
				return
			}
			for _, p := range parts {
				errs := validation.IsValidName(p)
				if len(errs) > 0 {
					log.Error(errs[len(errs)-1], log.Fields{}) // log the most recently appended msg
					http.Error(w, "Invalid cluster query", http.StatusBadRequest)
					return
				}
			}
		}
	} else {
		filterClusters = make([]string, 0)
	}

	var filterResources []string
	if r, found := qParams["resource"]; found {
		filterResources = r
		for _, res := range filterResources {
			errs := validation.IsValidName(res)
			if len(errs) > 0 {
				log.Error(errs[len(errs)-1], log.Fields{}) // log the most recently appended msg
				http.Error(w, "Invalid resources query", http.StatusBadRequest)
				return
			}
		}
	} else {
		filterResources = make([]string, 0)
	}

	// Different backend status functions are invoked based on which query parameters have been provided.
	// The query parameters will be handled with the following precedence to determine which status query is
	// invoked:
	//   if queryApps ("apps") == true then
	//		call StatusAppsList()
	//   else if queryClusters ("clusters") == true then
	//		call StatusClustersByApp()
	//   else if queryResources ("resources") == true then
	//		call StatusResourcesByApp()
	//   else
	//      call Status()
	//
	// Supplied query parameters which are not appropriate for the select function call are simply ignored.
	var iErr error
	var status interface{}

	if queryApps {
		status, iErr = h.client.StatusAppsList(ctx, p, ca, v, di, queryInstance)
	} else if queryClusters {
		status, iErr = h.client.StatusClustersByApp(ctx, p, ca, v, di, queryInstance, filterApps)
	} else if queryResources {
		status, iErr = h.client.StatusResourcesByApp(ctx, p, ca, v, di, queryInstance, queryType, filterApps, filterClusters)
	} else {
		status, iErr = h.client.Status(ctx, p, ca, v, di, queryInstance, queryType, queryOutput, filterApps, filterClusters, filterResources)
	}
	if iErr != nil {
		log.Error(iErr.Error(), log.Fields{})
		// Most errors are due to client side errors.
		// TODO Handle 5xx errors by propagating error type from source of error.
		http.Error(w, iErr.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	iErr = json.NewEncoder(w).Encode(status)
	if iErr != nil {
		log.Error(iErr.Error(), log.Fields{})
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}
}
