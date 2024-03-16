// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"

	hpaModel "gitlab.com/project-emco/core/emco-base/src/hpa-plc/pkg/model"
)

/*
addHpaConsumerHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers
*/
// Add Hpa Intent Consumer
func (h HpaPlacementIntentHandler) addHpaConsumerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var hpa hpaModel.HpaResourceConsumer
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: addHpaConsumerHandler .. start ::", log.Fields{"req": string(reqDump)})

	err := json.NewDecoder(r.Body).Decode(&hpa)
	switch {
	case err == io.EOF:
		log.Error(":: addHpaConsumerHandler ..Empty POST body ::", log.Fields{"Error": err})
		http.Error(w, "addHpaConsumerHandler .. Empty POST body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: addHpaConsumerHandler .. Error decoding POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(hpaConsumerJSONFile, hpa)
	if err != nil {
		log.Error(":: addHpaConsumerHandler .. JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]
	i := vars["intent-name"]

	log.Info(":: AddConsumer .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "consumer-name": hpa.MetaData.Name})
	consumer, err := h.client.AddConsumer(ctx, hpa, p, ca, v, di, i, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, hpa, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(consumer)
	if err != nil {
		log.Error(":: addHpaConsumerHandler .. Encoder error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: addHpaConsumerHandler .. end ::", log.Fields{"consumer": consumer})
}

/*
getHpaConsumerHandlerByName handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers?consumer=<consumer>
*/
// Query Hpa Intent Consumer
func (h HpaPlacementIntentHandler) getHpaConsumerHandlerByName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: getHpaConsumerHandlerByName .. start ::", log.Fields{"req": string(reqDump)})

	p, ca, v, di, i, _, err := parseHpaConsumerReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cN := r.URL.Query().Get("consumer")
	if cN == "" {
		log.Error(":: getHpaConsumerHandlerByName .. Missing intent-name in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(w, "getHpaConsumerHandlerByName .. Missing intent-name in request", http.StatusBadRequest)
		return
	}

	consumer, err := h.client.GetConsumerByName(ctx, cN, p, ca, v, di, i)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(consumer)
	if err != nil {
		log.Error(":: getHpaConsumerHandlerByName .. Encoder error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: getHpaConsumerHandlerByName .. end ::", log.Fields{"consumer": consumer})
}

/*
getHpaConsumerHandler/getHpaConsumerHandlers handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers/{consumer-name}
*/
// Get Hpa Intent Consumer
func (h HpaPlacementIntentHandler) getHpaConsumerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: getHpaConsumerHandler .. start ::", log.Fields{"req": string(reqDump)})
	p, ca, v, di, i, name, err := parseHpaConsumerReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: getHpaConsumerHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": name})

	var consumers interface{}
	if len(name) == 0 {
		consumers, err = h.client.GetAllConsumers(ctx, p, ca, v, di, i)
	} else {
		consumers, _, err = h.client.GetConsumer(ctx, name, p, ca, v, di, i)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(consumers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: getHpaConsumerHandler .. end ::", log.Fields{"consumers": consumers})
}

/*
putHpaConsumerHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers/{consumer-name}
*/
// Update Hpa Intent Consumer
func (h HpaPlacementIntentHandler) putHpaConsumerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var hpa hpaModel.HpaResourceConsumer
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: putHpaConsumerHandler .. start ::", log.Fields{"req": string(reqDump)})

	err := json.NewDecoder(r.Body).Decode(&hpa)
	switch {
	case err == io.EOF:
		log.Error(":: putHpaConsumerHandler .. Empty  body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: putHpaConsumerHandler .. Error decoding resource body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(hpaConsumerJSONFile, hpa)
	if err != nil {
		log.Error(":: putHpaConsumerHandler .. JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Parse request parameters
	p, ca, v, di, i, name, err := parseHpaConsumerReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if hpa.MetaData.Name != name {
		log.Error(":: putHpaConsumerHandler ..Mismatched name in PUT request ::", log.Fields{"bodyname": hpa.MetaData.Name, "name": name})
		http.Error(w, "putHpaConsumerHandler ..Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	log.Info(":: putHpaConsumerHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": name})
	consumer, err := h.client.AddConsumer(ctx, hpa, p, ca, v, di, i, true)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, hpa, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(consumer)
	if err != nil {
		log.Error(":: putHpaConsumerHandler .. encoding failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: putHpaConsumerHandler .. end ::", log.Fields{"consumer": consumer})
}

/*
deleteHpaConsumerHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers/{consumer-name}
*/
// Delete Hpa Intent Consumer
func (h HpaPlacementIntentHandler) deleteHpaConsumerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: deleteHpaConsumerHandler .. start ::", log.Fields{"req": string(reqDump)})

	// Parse request parameters
	p, ca, v, di, i, name, err := parseHpaConsumerReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: deleteHpaConsumerHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "consumer-name": name})

	_, _, err = h.client.GetConsumer(ctx, name, p, ca, v, di, i)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	err = h.client.DeleteConsumer(ctx, name, p, ca, v, di, i)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Info(":: deleteHpaConsumerHandler .. end ::", log.Fields{"req": string(reqDump)})
}

/*
deleteAllHpaConsumersHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers
*/
// Delete all Hpa Intent Consumers
func (h HpaPlacementIntentHandler) deleteAllHpaConsumersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: deleteAllHpaConsumersHandler .. start ::", log.Fields{"req": string(reqDump)})

	// Parse request parameters
	p, ca, v, di, i, name, err := parseHpaConsumerReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: deleteAllHpaConsumersHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "consumer-name": name})

	hpaConsumers, err := h.client.GetAllConsumers(ctx, p, ca, v, di, i)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	for _, hpaConsumer := range hpaConsumers {
		err = h.client.DeleteConsumer(ctx, hpaConsumer.MetaData.Name, p, ca, v, di, i)
		if err != nil {
			apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
	log.Info(":: deleteAllHpaConsumersHandler .. end ::", log.Fields{"consumers": hpaConsumers})
}

/* Parse Http request Parameters */
func parseHpaConsumerReqParameters(w *http.ResponseWriter, r *http.Request) (string, string, string, string, string, string, error) {
	vars := mux.Vars(r)

	cn := vars["consumer-name"]

	i := vars["intent-name"]
	if i == "" {
		log.Error(":: parseHpaIntentReqParameters ..  Missing intentName in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseHpaIntentReqParameters .. Missing name of intentName in request", http.StatusBadRequest)
		return "", "", "", "", "", "", pkgerrors.New("Missing intent-name")
	}

	p, ca, v, di, i, err := parseHpaIntentReqParameters(w, r)
	if err != nil {
		log.Error(":: parseHpaConsumerReqParameters .. Failed intent validation ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseHpaConsumerReqParameters .. Failed intent validation", http.StatusBadRequest)
		return "", "", "", "", "", "", err
	}

	return p, ca, v, di, i, cn, nil
}
