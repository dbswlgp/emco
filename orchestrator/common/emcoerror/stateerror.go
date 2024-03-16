// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package emcoerror

import (
	"fmt"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
)

// StateError defines the life cycle event failures
type StateError struct {
	Resource string           // Resource Type e.g: LogicalCloud, DeploymentIntentGroup, CaCert etc.
	Event    common.EmcoEvent // Life Cycle Event e.g: Instantiate, Terminate etc.
	Status   appcontext.StatusValue
}

// Error implements the error interface
func (e *StateError) Error() string {
	switch e.Status {
	case appcontext.AppContextStatusEnum.Terminating:
		return fmt.Sprintf("Failed to %s. The %s is being terminated", e.Event, e.Resource)
	case appcontext.AppContextStatusEnum.Instantiating:
		return fmt.Sprintf("Failed to %s. The %s is in instantiating status", e.Event, e.Resource)
	case appcontext.AppContextStatusEnum.TerminateFailed:
		return fmt.Sprintf("Failed to %s. The %s has failed terminating, please delete the %s", e.Event, e.Resource, e.Resource)
	case appcontext.AppContextStatusEnum.Terminated:
		return fmt.Sprintf("Failed to %s. The %s is already terminated", e.Event, e.Resource)
	case appcontext.AppContextStatusEnum.Instantiated:
		return fmt.Sprintf("Failed to %s. The %s is already instantiated", e.Event, e.Resource)
	case appcontext.AppContextStatusEnum.InstantiateFailed:
		return fmt.Sprintf("Failed to %s. The %s has failed instantiating before, please terminate and try again", e.Event, e.Resource)
	}

	return fmt.Sprintf("The %s isn't in an expected status so not taking any action", e.Resource)
}
