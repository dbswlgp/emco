// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package common

// EmcoEvent is the supported life-cycle events
type EmcoEvent string

const (
	Approve     EmcoEvent = "Approve"
	Instantiate EmcoEvent = "Instantiate"
	Migrate     EmcoEvent = "Migrate"
	Rollback    EmcoEvent = "Rollback"
	Stop        EmcoEvent = "Stop"
	Status      EmcoEvent = "Status"
	Terminate   EmcoEvent = "Terminate"
	Update      EmcoEvent = "Update"
)
