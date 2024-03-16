// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"context"
	"fmt"
	"reflect"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
)

// CaCertManager exposes all the caCert functionalities
type CaCertManager interface {
	CreateCert(ctx context.Context, cert CaCert, failIfExists bool) (CaCert, bool, error)
	DeleteCert(ctx context.Context) error
	GetAllCert(ctx context.Context) ([]CaCert, error)
	GetCert(ctx context.Context) (CaCert, error)
}

// CaCertClient holds the client properties
type CaCertClient struct {
	dbInfo db.DbInfo
	dbKey  interface{}
}

// NewCertClients returns an instance of the CaCertClient which implements the Manager
func NewCaCertClient(dbKey interface{}) *CaCertClient {
	return &CaCertClient{
		dbInfo: db.DbInfo{
			StoreName: "resources",
			TagMeta:   "data",
			TagState:  "stateInfo"},
		dbKey: dbKey}
}

// CreateCert creates a caCert
func (c *CaCertClient) CreateCert(ctx context.Context, cert CaCert, failIfExists bool) (CaCert, bool, error) {
	certExists := false

	if cer, err := c.GetCert(ctx); err == nil &&
		!reflect.DeepEqual(cer, CaCert{}) {
		certExists = true
	}

	if certExists &&
		failIfExists {
		return CaCert{}, certExists, emcoerror.NewEmcoError(
			CaCertAlreadyExists,
			emcoerror.Conflict,
		)
	}

	if err := db.DBconn.Insert(ctx, c.dbInfo.StoreName, c.dbKey, nil, c.dbInfo.TagMeta, cert); err != nil {
		return CaCert{}, certExists, err
	}

	return cert, certExists, nil
}

// DeleteCert deletes a caCert
func (c *CaCertClient) DeleteCert(ctx context.Context) error {
	return db.DBconn.Remove(ctx, c.dbInfo.StoreName, c.dbKey)
}

// GetAllCert
func (c *CaCertClient) GetAllCert(ctx context.Context) ([]CaCert, error) {
	values, err := db.DBconn.Find(ctx, c.dbInfo.StoreName, c.dbKey, c.dbInfo.TagMeta)
	if err != nil {
		return []CaCert{}, err
	}

	var certs []CaCert
	for _, value := range values {
		cert := CaCert{}
		if err = db.DBconn.Unmarshal(value, &cert); err != nil {
			return []CaCert{}, err
		}
		certs = append(certs, cert)
	}

	return certs, nil
}

// GetCert returns the caCert
func (c *CaCertClient) GetCert(ctx context.Context) (CaCert, error) {
	value, err := db.DBconn.Find(ctx, c.dbInfo.StoreName, c.dbKey, c.dbInfo.TagMeta)
	if err != nil {
		return CaCert{}, err
	}

	if len(value) == 0 {
		return CaCert{}, emcoerror.NewEmcoError(
			CaCertNotFound,
			emcoerror.NotFound,
		)
	}

	if value != nil {
		cert := CaCert{}
		if err = db.DBconn.Unmarshal(value[0], &cert); err != nil {
			return CaCert{}, err
		}
		return cert, nil
	}

	return CaCert{}, emcoerror.NewEmcoError(
		emcoerror.UnknownErrorMessage,
		emcoerror.Unknown,
	)
}

// UpdateCert update the caCert
func (c *CaCertClient) UpdateCert(ctx context.Context, cert CaCert) error {
	return db.DBconn.Insert(ctx, c.dbInfo.StoreName, c.dbKey, nil, c.dbInfo.TagMeta, cert)
}

// VerifyStateBeforeDelete verifies a caCert can be deleted or not
func (c *CaCertClient) VerifyStateBeforeDelete(ctx context.Context, cert, lifecycle string) error {
	sc := NewStateClient(c.dbKey)
	stateInfo, err := sc.Get(ctx)
	if err != nil {
		return err
	}

	cState, err := state.GetCurrentStateFromStateInfo(stateInfo)
	if err != nil {
		return err
	}

	if cState == state.StateEnum.Instantiated ||
		cState == state.StateEnum.InstantiateStopped {
		err := emcoerror.NewEmcoError(
			fmt.Sprintf("%s must be terminated for CaCert %s before it can be deleted", lifecycle, cert),
			emcoerror.Conflict,
		)
		logutils.Error("",
			logutils.Fields{
				"Error": err.Error()})
		return err
	}

	if cState == state.StateEnum.Terminated ||
		cState == state.StateEnum.TerminateStopped {
		// verify that the appcontext has completed terminating
		ctxID := state.GetLastContextIdFromStateInfo(stateInfo)
		acStatus, err := state.GetAppContextStatus(ctx, ctxID)
		if err == nil &&
			!(acStatus.Status == appcontext.AppContextStatusEnum.Terminated ||
				acStatus.Status == appcontext.AppContextStatusEnum.TerminateFailed) {
			err := emcoerror.NewEmcoError(
				fmt.Sprintf("%s termination has not completed for CaCert %s", lifecycle, cert),
				emcoerror.Conflict,
			)
			logutils.Error("",
				logutils.Fields{
					"Error": err.Error()})
			return err
		}

		for _, id := range state.GetContextIdsFromStateInfo(stateInfo) {
			appCtx, err := state.GetAppContextFromId(ctx, id)
			if err != nil {
				logutils.Error("Failed to get appContext from id",
					logutils.Fields{
						"ID":    id,
						"Error": err.Error()})
				return err
			}
			err = appCtx.DeleteCompositeApp(ctx)
			if err != nil {
				logutils.Error("Failed to delete the appContext",
					logutils.Fields{
						"Error": err.Error()})
				return err
			}
		}
	}

	return nil
}

// VerifyStateBeforeUpdate verifies a caCert can be updated or not
func (c *CaCertClient) VerifyStateBeforeUpdate(ctx context.Context, cert, lifecycle string) error {
	sc := NewStateClient(c.dbKey)
	stateInfo, err := sc.Get(ctx)
	if err != nil {
		return err
	}

	cState, err := state.GetCurrentStateFromStateInfo(stateInfo)
	if err != nil {
		return err
	}

	if cState != state.StateEnum.Created {
		err := emcoerror.NewEmcoError(
			fmt.Sprintf("Failed to update the CaCert. %s for the CaCert %s is in %s state", lifecycle, cert, cState),
			emcoerror.Conflict,
		)
		logutils.Error("",
			logutils.Fields{
				"Error": err.Error()})
		return err
	}

	return nil
}
