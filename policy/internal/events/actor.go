// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package event

import (
	"github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

func DoAction(actorName string, evaluationResult []byte, intentSpec []byte, agentSpec []byte, actors map[string]Actor) error {
	var (
		actor Actor
	)
	switch actorName {
	case "temporal":
		actor = actors["temporal"]
	default:
		log.Error("DoAction: No Actor Plugin matched", log.Fields{"actor": actorName})
		return errors.Errorf("DoAction: No Actor Plugin matched with provided actor name: %s", actorName)
	}
	return actor.Execute(evaluationResult, intentSpec, agentSpec)
}
