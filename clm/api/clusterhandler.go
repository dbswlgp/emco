// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	//"fmt"

	"github.com/gorilla/mux"
	clusterPkg "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

// clusterproviderpriority
//var cppJSONFile string = "json-schemas/cluster-provider-priority.json"

var cpJSONFile string = "json-schemas/metadata.json"
var ckvJSONFile string = "json-schemas/cluster-kv.json"
var clJSONFile string = "json-schemas/cluster-label.json"
var copsJSONFile string = "json-schemas/cluster-gitops.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type clusterHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client clusterPkg.ClusterManager
}

// Create handles creation of the ClusterProvider entry in the database
func (h clusterHandler) createClusterProviderHandler(w http.ResponseWriter, r *http.Request) {
	var p clusterPkg.ClusterProvider

	ctx := r.Context()

	err := json.NewDecoder(r.Body).Decode(&p)

	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster provider POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding cluster provider POST body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(cpJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster provider POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing name in cluster provider POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	// If priority is not set, use default value (10)
	if p.Spec.Priority == 0 {
		p.Spec.Priority = 10
	}

	ret, err := h.client.CreateClusterProvider(ctx, p, false)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, p, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create cluster provider response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// putClusterProviderHandler handles updating of the ClusterProvider entry in the database
func (h clusterHandler) putClusterProviderHandler(w http.ResponseWriter, r *http.Request) {
	var p clusterPkg.ClusterProvider
	var err error
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["clusterProvider"]

	err = json.NewDecoder(r.Body).Decode(&p)

	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster provider PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding cluster provider PUT body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(cpJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster provider POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing name in cluster provider POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	// Name in URL shoudl match name in body
	if p.Metadata.Name != name {
		log.Error(":: Mismatched name in cluster provider PUT request ::", log.Fields{"Error": err})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClusterProvider(ctx, p, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, p, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding update cluster provider response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular ClusterProvider Name
// Returns a ClusterProvider
func (h clusterHandler) getClusterProviderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["clusterProvider"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetClusterProviders(ctx)
	} else {
		ret, err = h.client.GetClusterProvider(ctx, name)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding get cluster provider response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular ClusterProvider  Name
func (h clusterHandler) deleteClusterProviderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	name := vars["clusterProvider"]

	err := h.client.DeleteClusterProvider(ctx, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}








/*
// Create handles creation of the ClusterProviderPriority entry in the database
func (h clusterHandler) createClusterProviderPriorityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	//cluster := vars["cluster"]
	var p clusterPkg.ClusterProviderPriority

	err := json.NewDecoder(r.Body).Decode(&p)
	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster provider priority POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding cluster provider priority POST body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(cppJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster provider priority POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Priority is required.
	if p.Priority == 0 {
		log.Error(":: Missing cluster provider priority in POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing priority in POST request", http.StatusBadRequest)
		return
	}

	//ret, err := h.client.CreateClusterLabel(ctx, provider, cluster, p, false)
	ret, err := h.client.CreateClusterProviderPriority(ctx, provider, p, false)

	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, p, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding cluster label response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// putClusterProviderPriorityHandler handles updating of a ClusterProviderPriority entry in the database
func (h clusterHandler) putClusterProviderPriorityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	//cluster := vars["cluster"]
	//label := vars["clusterLabel"]
	priority := vars["priority"]
	var p clusterPkg.ClusterProviderPriority

	err := json.NewDecoder(r.Body).Decode(&p)
	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster provider priority PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding cluster provider priority PUT body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(cppJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster provider priority PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Priority is required.
	if p.Priority == 0 {
		log.Error(":: Missing cluster provider priority in PUT request ::", log.Fields{"Error": err})
		http.Error(w, "Missing priority in PUT request", http.StatusBadRequest)
		return
	}

	// LabelName should match in URL and body
	if p.Priority != priority {
		log.Error(":: Mismatched cluster provider priority in PUT request ::", log.Fields{"Error": err})
		http.Error(w, "Mismatched priority in PUT request", http.StatusBadRequest)
		return
	}


	ret, err := h.client.CreateClusterProviderPriority(ctx, provider, p, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, p, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding cluster provider priority response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular Cluster Label
// Returns a ClusterLabel
func (h clusterHandler) getClusterProviderPriorityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	//cluster := vars["cluster"]
	//label := vars["clusterLabel"]
	//priority := vars["priority"]

	var ret interface{}
	var err error


	if len(label) == 0 {
		ret, err = h.client.GetClusterLabels(ctx, provider, cluster)
	} else {
		ret, err = h.client.GetClusterLabel(ctx, provider, cluster, label)
	}


	ret, err = h.client.GetClusterProviderPriority(ctx, provider)
	//ret, err = h.client.GetClusterProviderPriority(ctx, provider, priority)




	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding cluster provider priority response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}


// Delete handles DELETE operations on a particular ClusterLabel Name
func (h clusterHandler) deleteClusterLabelHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	//cluster := vars["cluster"]
	//label := vars["clusterLabel"]

	err := h.client.DeleteClusterLabel(ctx, provider)
	//err := h.client.DeleteClusterLabel(ctx, provider, cluster, label)

	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
*/












// Create handles creation of the Cluster entry in the database
func (h clusterHandler) createClusterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	var p clusterPkg.Cluster
	var q clusterPkg.ClusterContent

	// Implemenation using multipart form
	// Review and enable/remove at a later date
	// Set Max size to 16mb here
	err := r.ParseMultipartForm(16777216)
	if err != nil {
		log.Error(":: Error parsing cluster multipart form ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	jsn := bytes.NewBuffer([]byte(r.FormValue("metadata")))
	err = json.NewDecoder(jsn).Decode(&p)

	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding cluster POST body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(copsJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing name in cluster POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	// check for spec section
	if p.Spec.Props.GitOpsType == "" {
		//Read the file section and ignore the header
		file, _, err := r.FormFile("file")
		if err != nil {
			log.Error(":: Error getting file section ::", log.Fields{"Error": err})
			http.Error(w, "Unable to process file", http.StatusUnprocessableEntity)
			return
		}

		defer file.Close()

		//Convert the file content to base64 for storage
		content, err := ioutil.ReadAll(file)
		if err != nil {
			log.Error(":: Error reading file section ::", log.Fields{"Error": err})
			http.Error(w, "Unable to read file", http.StatusUnprocessableEntity)
			return
		}

		q.Kubeconfig = base64.StdEncoding.EncodeToString(content)
		ret, err := h.client.CreateCluster(ctx, provider, p, q)
		if err != nil {
			apiErr := apierror.HandleErrors(vars, err, p, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}

		//	w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(ret)
		if err != nil {
			log.Error(":: Error encoding create cluster response ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		q.Kubeconfig = ""
		ret, err := h.client.CreateCluster(ctx, provider, p, q)
		if err != nil {
			apiErr := apierror.HandleErrors(vars, err, p, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}

		//	w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(ret)
		if err != nil {
			log.Error(":: Error encoding create cluster response ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Get handles GET operations on a particular Cluster Name
// Returns a Cluster
func (h clusterHandler) getClusterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	name := vars["cluster"]

	withLabels := r.URL.Query().Get("withLabels")
	log.Warn("with Labels ", log.Fields{"val": withLabels})
	if strings.ToLower(withLabels) == "true" && len(name) == 0 {
		ret, err := h.client.GetAllClustersAndLabels(ctx, provider)
		if err != nil {
			apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(ret)
		if err != nil {
			log.Error(":: Error encoding get clusters by label response ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	label := r.URL.Query().Get("label")
	log.Info(":: get clusters by label parameters ::", log.Fields{"label": label, "provider": provider, "cluster": name})
	if len(label) != 0 && len(name) == 0 {
		ret, err := h.client.GetClustersWithLabel(ctx, provider, label)
		if err != nil {
			apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(ret)
		if err != nil {
			log.Error(":: Error encoding get clusters by label response ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// handle the get all clusters case - return a list of only the json parts
	if len(name) == 0 {
		ret, err := h.client.GetClusters(ctx, provider)
		if err != nil {
			apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(ret)
		if err != nil {
			log.Error(":: Error encoding get clusters ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	accepted, _, err := mime.ParseMediaType(r.Header.Get("Accept"))
	if err != nil {
		log.Error(":: Missing Accept header in get cluster request ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	retCluster, err := h.client.GetCluster(ctx, provider, name)

	if err != nil {
		log.Error(":: Error getting cluster ::", log.Fields{"Error": err})
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	// check for spec section in the response
	if retCluster.Spec.Props.GitOpsType == "" {
		retKubeconfig, err := h.client.GetClusterContent(ctx, provider, name)
		if err != nil {
			apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}

		switch accepted {
		case "multipart/form-data":
			mpw := multipart.NewWriter(w)
			w.Header().Set("Content-Type", mpw.FormDataContentType())
			w.WriteHeader(http.StatusOK)
			pw, err := mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/json"}, "Content-Disposition": {"form-data; name=metadata"}})
			if err != nil {
				log.Error(":: Error creating metadata part of cluster response ::", log.Fields{"Error": err})
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := json.NewEncoder(pw).Encode(retCluster); err != nil {
				log.Error(":: Error encoding cluster response ::", log.Fields{"Error": err})
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			pw, err = mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/octet-stream"}, "Content-Disposition": {"form-data; name=file; filename=kubeconfig"}})
			if err != nil {
				log.Error(":: Error creating file part of cluster response ::", log.Fields{"Error": err})
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			kcBytes, err := base64.StdEncoding.DecodeString(retKubeconfig.Kubeconfig)
			if err != nil {
				log.Error(":: Error encoding file part of cluster response ::", log.Fields{"Error": err})
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, err = pw.Write(kcBytes)
			if err != nil {
				log.Error(":: Error writing multipart cluster response ::", log.Fields{"Error": err})
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "application/json":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(retCluster)
			if err != nil {
				log.Error(":: Error encoding cluster response ::", log.Fields{"Error": err})
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "application/octet-stream":
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			kcBytes, err := base64.StdEncoding.DecodeString(retKubeconfig.Kubeconfig)
			if err != nil {
				log.Error(":: Error encoding file part of cluster response ::", log.Fields{"Error": err})
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, err = w.Write(kcBytes)
			if err != nil {
				log.Error(":: Error writing cluster response ::", log.Fields{"Error": err})
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		default:
			log.Error(":: Missing Accept header for get cluster ::", log.Fields{"Error": err})
			http.Error(w, "set Accept: multipart/form-data, application/json or application/octet-stream", http.StatusMultipleChoices)
			return
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(retCluster)
		if err != nil {
			log.Error(":: Error encoding cluster response ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

// Delete handles DELETE operations on a particular Cluster Name
func (h clusterHandler) deleteClusterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	name := vars["cluster"]

	err := h.client.DeleteCluster(ctx, provider, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Create handles creation of the ClusterLabel entry in the database
func (h clusterHandler) createClusterLabelHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	cluster := vars["cluster"]
	var p clusterPkg.ClusterLabel

	err := json.NewDecoder(r.Body).Decode(&p)
	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster label POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding cluster label POST body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(clJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster label POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// LabelName is required.
	if p.LabelName == "" {
		log.Error(":: Missing cluster label name in POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing label name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClusterLabel(ctx, provider, cluster, p, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, p, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding cluster label response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// putClusterLabelHanderl handles updating of a ClusterLabel entry in the database
func (h clusterHandler) putClusterLabelHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	cluster := vars["cluster"]
	label := vars["clusterLabel"]
	var p clusterPkg.ClusterLabel

	err := json.NewDecoder(r.Body).Decode(&p)
	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster label PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding cluster label PUT body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(clJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster label PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// LabelName is required.
	if p.LabelName == "" {
		log.Error(":: Missing cluster label name in PUT request ::", log.Fields{"Error": err})
		http.Error(w, "Missing label name in PUT request", http.StatusBadRequest)
		return
	}

	// LabelName should match in URL and body
	if p.LabelName != label {
		log.Error(":: Mismatched cluster label name in PUT request ::", log.Fields{"Error": err})
		http.Error(w, "Mismatched label name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClusterLabel(ctx, provider, cluster, p, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, p, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding cluster label response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular Cluster Label
// Returns a ClusterLabel
func (h clusterHandler) getClusterLabelHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	cluster := vars["cluster"]
	label := vars["clusterLabel"]

	var ret interface{}
	var err error

	if len(label) == 0 {
		ret, err = h.client.GetClusterLabels(ctx, provider, cluster)
	} else {
		ret, err = h.client.GetClusterLabel(ctx, provider, cluster, label)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding cluster label response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular ClusterLabel Name
func (h clusterHandler) deleteClusterLabelHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	cluster := vars["cluster"]
	label := vars["clusterLabel"]

	err := h.client.DeleteClusterLabel(ctx, provider, cluster, label)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Create handles creation of the ClusterKvPairs entry in the database
func (h clusterHandler) createClusterKvPairsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	cluster := vars["cluster"]
	var p clusterPkg.ClusterKvPairs

	err := json.NewDecoder(r.Body).Decode(&p)
	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster kv pair POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding cluster kv pair POST body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(ckvJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster kv pair POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// KvPairsName is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing cluster kv pair name in POST body ::", log.Fields{"Error": err})
		http.Error(w, "Missing Key Value pair name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClusterKvPairs(ctx, provider, cluster, p, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, p, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding cluster kv pair ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// putClusterKvPairsHandler  handles update of a ClusterKvPairs entry in the database
func (h clusterHandler) putClusterKvPairsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	cluster := vars["cluster"]
	kvpair := vars["clusterKv"]
	var p clusterPkg.ClusterKvPairs

	err := json.NewDecoder(r.Body).Decode(&p)
	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster kv pair PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding cluster kv pair PUT body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(ckvJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster kv pair POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// KvPairsName is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing cluster kv pair name in POST body ::", log.Fields{"Error": err})
		http.Error(w, "Missing Key Value pair name in POST request", http.StatusBadRequest)
		return
	}

	// KvPairsName should match in URL and body
	if p.Metadata.Name != kvpair {
		log.Error(":: Mismatched cluster kv pair name in PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Mismatched Key Value pair name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClusterKvPairs(ctx, provider, cluster, p, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, p, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding cluster kv pair ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular Cluster Key Value Pair
// Returns a ClusterKvPairs
func (h clusterHandler) getClusterKvPairsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	cluster := vars["cluster"]
	kvpair := vars["clusterKv"]
	kvkey := r.URL.Query().Get("key")

	var ret interface{}
	var err error

	if len(kvpair) == 0 {
		ret, err = h.client.GetAllClusterKvPairs(ctx, provider, cluster)
	} else if len(kvkey) != 0 {
		ret, err = h.client.GetClusterKvPairsValue(ctx, provider, cluster, kvpair, kvkey)
	} else {
		ret, err = h.client.GetClusterKvPairs(ctx, provider, cluster, kvpair)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding cluster kv pair response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular Cluster Name
func (h clusterHandler) deleteClusterKvPairsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	cluster := vars["cluster"]
	kvpair := vars["clusterKv"]

	err := h.client.DeleteClusterKvPairs(ctx, provider, cluster, kvpair)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Create handles creation of the Cluster Sync Objects entry in the database
func (h clusterHandler) createClusterSyncObjectsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]

	var p mtypes.ClusterSyncObjects

	err := json.NewDecoder(r.Body).Decode(&p)
	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster sync object POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding cluster sync object POST body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(ckvJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster sync object POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// SyncObjectsName is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing cluster sync object name in POST body ::", log.Fields{"Error": err})
		http.Error(w, "Missing cluster sync object name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClusterSyncObjects(ctx, provider, p, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, p, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding cluster sync object ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// putClusterSyncObjectsHandler  handles update of a ClusterSyncObjects entry in the database
func (h clusterHandler) putClusterSyncObjectsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	syncobject := vars["clusterSyncObject"]
	var p mtypes.ClusterSyncObjects

	err := json.NewDecoder(r.Body).Decode(&p)
	switch {
	case err == io.EOF:
		log.Error(":: Empty cluster sync object PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding sync object PUT body ::", log.Fields{"Error": err, "Body": p})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(ckvJSONFile, p)
	if err != nil {
		log.Error(":: Invalid cluster sync object POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// SyncObjectsName is required.
	if p.Metadata.Name == "" {
		log.Error(":: Missing cluster sync object name in POST body ::", log.Fields{"Error": err})
		http.Error(w, "Missing cluster sync object name in POST request", http.StatusBadRequest)
		return
	}

	// SyncObjectsName should match in URL and body
	if p.Metadata.Name != syncobject {
		log.Error(":: Mismatched cluster sync object name in PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Mismatched cluster sync object name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClusterSyncObjects(ctx, provider, p, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, p, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding cluster sync object ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular Cluster Sync Object
// Returns a ClusterSyncObjects
func (h clusterHandler) getClusterSyncObjectsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	syncobject := vars["clusterSyncObject"]
	syncobjectkey := r.URL.Query().Get("key")

	var ret interface{}
	var err error

	if len(syncobject) == 0 {
		ret, err = h.client.GetAllClusterSyncObjects(ctx, provider)
	} else if len(syncobjectkey) != 0 {
		ret, err = h.client.GetClusterSyncObjectsValue(ctx, provider, syncobject, syncobjectkey)
	} else {
		ret, err = h.client.GetClusterSyncObjects(ctx, provider, syncobject)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding cluster sync object response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular Cluster Provider
func (h clusterHandler) deleteClusterSyncObjectsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	provider := vars["clusterProvider"]
	syncobject := vars["clusterSyncObject"]

	err := h.client.DeleteClusterSyncObjects(ctx, provider, syncobject)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
