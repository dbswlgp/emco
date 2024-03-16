// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package controller

import (
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"net/http"
)

func (c Controller) Health(w http.ResponseWriter, _ *http.Request) {
	//c.DbTest()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Controller is UP")); err != nil {
		log.Warn("Couldn't write response for heath check", log.Fields{"err": err})
	}
}
