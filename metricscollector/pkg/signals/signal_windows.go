// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package signals

import (
	"os"
)

var shutdownSignals = []os.Signal{os.Interrupt}
