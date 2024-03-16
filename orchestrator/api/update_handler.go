package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

var migrateJSONFile string = "json-schemas/migrate.json"
var rollbackJSONFile string = "json-schemas/rollback.json"

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type updateHandler struct {
	client moduleLib.InstantiationManager
}

func (h updateHandler) migrateHandler(w http.ResponseWriter, r *http.Request) {
	var migrate moduleLib.MigrateJson

	ctx := r.Context()
	vars := mux.Vars(r)
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	di := vars["deploymentIntentGroup"]

	err := json.NewDecoder(r.Body).Decode(&migrate)
	log.Info("migrateJson:", log.Fields{"json:": migrate})
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Empty body", http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		return
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(migrateJSONFile, migrate)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		w.WriteHeader(httpError)
		return
	}

	tCav := migrate.Spec.TargetCompositeAppVersion
	tDig := migrate.Spec.TargetDigName

	log.Info("targetDeploymentName and targetCompositeAppVersion", log.Fields{"targetDeploymentName": tDig, "targetCompositeAppVersion": tCav})
	iErr := h.client.Migrate(ctx, p, ca, v, tCav, di, tDig)
	if iErr != nil {
		log.Error(":: Error migrate handler ::", log.Fields{"Error": iErr.Error(), "project": p, "compositeApp": ca, "compositeAppVer": v,
			"targetCompositeAppVersion": tCav, "depGroup": di, "targetDigName": tDig})
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
	log.Info("migrateHandler ... end ", log.Fields{"project": p, "compositeApp": ca, "compositeAppVer": v,
		"targetCompositeAppVersion": tCav, "depGroup": di, "targetDigName": tDig, "returnValue": iErr})
	w.WriteHeader(http.StatusAccepted)
}

func (h updateHandler) updateHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	vars := mux.Vars(r)
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	di := vars["deploymentIntentGroup"]

	revisionID, iErr := h.client.Update(ctx, p, ca, v, di)
	if iErr != nil {
		log.Error(":: Error update handler ::", log.Fields{"Error": iErr.Error(), "project": p, "compositeApp": ca, "compositeAppVer": v,
			"depGroup": di})
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
	log.Info("updateHandler ... end ", log.Fields{"project": p, "compositeApp": ca, "compositeAppVer": v,
		"depGroup": di, "returnValue": iErr})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err := json.NewEncoder(w).Encode(revisionID)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h updateHandler) rollbackHandler(w http.ResponseWriter, r *http.Request) {
	var rollback moduleLib.RollbackJson

	ctx := r.Context()
	vars := mux.Vars(r)
	p := vars["project"]
	ca := vars["compositeApp"]
	v := vars["compositeAppVersion"]
	di := vars["deploymentIntentGroup"]

	err := json.NewDecoder(r.Body).Decode(&rollback)
	log.Info("rollbackJson:", log.Fields{"json:": rollback})
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(rollbackJSONFile, rollback)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	rbRev := rollback.Spec.Revison

	iErr := h.client.Rollback(ctx, p, ca, v, di, rbRev)
	if iErr != nil {
		log.Error(":: Error rollback handler ::", log.Fields{"Error": iErr.Error(), "project": p, "compositeApp": ca, "compositeAppVer": v,
			"depGroup": di, "revision": rbRev})
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
	log.Info("rollbackHandler ... end ", log.Fields{"project": p, "compositeApp": ca, "compositeAppVer": v,
		"depGroup": di, "revision": rbRev, "returnValue": iErr})
	w.WriteHeader(http.StatusAccepted)

}
