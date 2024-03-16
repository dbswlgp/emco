// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"fmt"

	rsyncclient "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/updateappclient"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

func callRsyncUpdate(ctx context.Context, FromContextid, ToContextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo(ctx)
	log.Info("Calling the Rsync ", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		return err
	}

	fromAppContextID := fmt.Sprintf("%v", FromContextid)
	toAppContextID := fmt.Sprintf("%v", ToContextid)
	err = rsyncclient.InvokeUpdateApp(ctx, fromAppContextID, toAppContextID)
	if err != nil {
		return err
	}
	return nil
}
