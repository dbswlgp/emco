// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package enrollment

import (
	"strings"

	"github.com/pkg/errors"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"

	"context"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/notifyclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
)

const AppName string = "cert-enrollment"

// Instantiate the caCert enrollment
func (enrollCtx *EnrollmentContext) Instantiate(ctx context.Context) error {
	for _, enrollCtx.ClusterGroup = range enrollCtx.ClusterGroups {
		// get all the clusters in this clusterGroup
		clusters, err := module.GetClusters(ctx, enrollCtx.ClusterGroup, enrollCtx.Project, enrollCtx.LogicalCloud)
		if err != nil {
			return err
		}

		for _, enrollCtx.Cluster = range clusters {
			// create resources for the edge clsuters based on the issuer
			switch enrollCtx.CaCert.Spec.IssuerRef.Group {
			case "cert-manager.io":

				// check whether the cluster is sgx enabled or not
				// if it's an sgx capable cluster, create a cert-manager certificate resource
				// otherwise, create a cert-manger certificaterequest resource
				if val, err := clm.NewClusterClient().GetClusterKvPairsValue(ctx, enrollCtx.ClusterGroup.Spec.Provider, enrollCtx.Cluster, "sgx", "enabled"); err == nil {
					v, e := module.GetValue(val)
					if e != nil {
						return e
					}

					if v == "true" { // sgx enabled
						if err := enrollCtx.createCertManagerCertificate(ctx); err != nil {
							return err
						}

						continue // with the next cluster
					}
				}

				// cluster is not SGX enabled, create a certificaterequest
				if err := enrollCtx.createCertManagerCertificateRequest(ctx); err != nil {
					return err
				}

			default:
				err := errors.New("unsupported Issuer")
				logutils.Error("",
					logutils.Fields{
						"Issuer": enrollCtx.CaCert.Spec.IssuerRef.Group,
						"Error":  err.Error()})
				return err
			}

			enrollCtx.Cluster = ""
		}

		enrollCtx.ClusterGroup = module.ClusterGroup{}
	}

	return nil
}

// Status provides the caCert enrollment status
func Status(ctx context.Context, stateInfo state.StateInfo, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (module.CaCertStatus, error) {
	statusResult, err := status.PrepareCaCertStatusResult(ctx, stateInfo, qInstance, qType, qOutput, fApps, fClusters, fResources)
	if err != nil {
		logutils.Error("Failed to get the enrollemnt status",
			logutils.Fields{
				"Error": err.Error()})
		return module.CaCertStatus{}, err
	}

	caCertStatus := module.CaCertStatus{}
	caCertStatus.Name = statusResult.Name
	caCertStatus.State = statusResult.State
	caCertStatus.DeployedStatus = statusResult.DeployedStatus
	caCertStatus.ReadyStatus = statusResult.ReadyStatus
	caCertStatus.DeployedCounts = statusResult.DeployedCounts
	caCertStatus.ReadyCounts = statusResult.ReadyCounts
	caCertStatus.Clusters = statusResult.Clusters

	return caCertStatus, nil
}

// Terminate the caCert enrollment
func (enrollCtx *EnrollmentContext) Terminate(ctx context.Context) error {
	for _, enrollCtx.ClusterGroup = range enrollCtx.ClusterGroups {
		// get all the clusters in this clusterGoup
		clusters, err := module.GetClusters(ctx, enrollCtx.ClusterGroup, enrollCtx.Project, enrollCtx.LogicalCloud)
		if err != nil {
			return err
		}
		// delete all the resources associated with enrollment
		for _, enrollCtx.Cluster = range clusters {
			// delete the primary key, if it exists
			if enrollCtx.privateKeyExists(ctx) {
				if err := enrollCtx.deletePrivateKey(ctx); err != nil {
					return err
				}
			}

			enrollCtx.Cluster = ""
		}

		enrollCtx.ClusterGroup = module.ClusterGroup{}
	}

	return nil
}

// Update the caCert enrollment
func (enrollCtx *EnrollmentContext) Update(ctx context.Context, contextID string) error {
	// initialize the Instantiation
	if err := enrollCtx.Instantiate(ctx); err != nil {
		return err
	}

	// add instruction under the given handle and type
	if err := module.AddInstruction(ctx, enrollCtx.AppContext, enrollCtx.IssuerHandle, enrollCtx.ResOrder); err != nil {
		return err
	}

	if err := state.UpdateAppContextStatusContextID(ctx, enrollCtx.ContextID, contextID); err != nil {
		logutils.Error("Failed to update appContext status",
			logutils.Fields{
				"ContextID": enrollCtx.ContextID,
				"AppName":   AppName,
				"Error":     err.Error()})
		return err
	}

	if err := notifyclient.CallRsyncUpdate(ctx, contextID, enrollCtx.ContextID); err != nil {
		logutils.Error("Rsync update failed",
			logutils.Fields{
				"ContextID": enrollCtx.ContextID,
				"AppName":   AppName,
				"Error":     err.Error()})
		return err
	}

	// subscribe to alerts
	stream, _, err := notifyclient.InvokeReadyNotify(ctx, enrollCtx.ContextID, enrollCtx.ClientName)
	if err != nil {
		logutils.Error("Failed to subscribe to alerts",
			logutils.Fields{
				"ContextID":  enrollCtx.ContextID,
				"ClientName": enrollCtx.ClientName,
				"AppName":    AppName,
				"Error":      err.Error()})
		return err
	}

	if err := stream.CloseSend(); err != nil {
		logutils.Error("Failed to close the send stream",
			logutils.Fields{
				"ContextID":  enrollCtx.ContextID,
				"ClientName": enrollCtx.ClientName,
				"AppName":    AppName,
				"Error":      err.Error()})
		return err
	}

	return nil
}

// IssuingClusterHandle returns the handle of certificate issuing cluster
func (enrollCtx *EnrollmentContext) IssuingClusterHandle(ctx context.Context) (interface{}, error) {
	var (
		handle interface{}
		err    error
	)

	// add handle for the issuing cluster
	handle, err = enrollCtx.AppContext.AddCluster(ctx, enrollCtx.AppHandle,
		strings.Join([]string{enrollCtx.CaCert.Spec.IssuingCluster.ClusterProvider, enrollCtx.CaCert.Spec.IssuingCluster.Cluster}, "+"))
	if err != nil {
		logutils.Error("Failed to add the issuing cluster",
			logutils.Fields{
				"Error": err.Error()})

		if er := enrollCtx.AppContext.DeleteCompositeApp(ctx); er != nil {
			logutils.Error("Failed to delete the compositeApp",
				logutils.Fields{
					"ContextID": enrollCtx.ContextID,
					"Error":     er.Error()})
			return handle, er
		}

		return handle, err

	}

	return handle, err
}

// VerifyEnrollmentState verify the caCert enrollment state
func VerifyEnrollmentState(ctx context.Context, stateInfo state.StateInfo) (enrollmentContextID string, err error) {
	// get the caCert enrollemnt instantiation state
	enrollmentContextID = state.GetLastContextIdFromStateInfo(stateInfo)
	if len(enrollmentContextID) == 0 {
		err := errors.New("enrollment is not completed")
		logutils.Error("",
			logutils.Fields{
				"Error": err.Error()})
		return "", err
	}

	status, err := state.GetAppContextStatus(ctx, enrollmentContextID)
	if err != nil {
		logutils.Error("Failed to get the appContext status",
			logutils.Fields{
				"ContextID": enrollmentContextID,
				"Error":     err.Error()})
		return "", err
	}

	if status.Status != appcontext.AppContextStatusEnum.Instantiated &&
		status.Status != appcontext.AppContextStatusEnum.Updated {
		err := errors.New("enrollment is not completed")
		logutils.Error("",
			logutils.Fields{
				"Status": status.Status,
				"Error":  err.Error()})
		return "", err
	}

	return enrollmentContextID, err
}

// ValidateEnrollmentStatus verify the caCert enrollment status
func ValidateEnrollmentStatus(ctx context.Context, stateInfo state.StateInfo) (readyCount int, err error) {
	//  verify the status of the enrollemnt
	certEnrollmentStatus, err := Status(ctx, stateInfo, "", "ready", "all", make([]string, 0), make([]string, 0), make([]string, 0))
	if err != nil {
		return readyCount, err
	}

	if strings.ToLower(string(certEnrollmentStatus.DeployedStatus)) != "instantiated" {
		err := errors.New("enrollment is not ready")
		logutils.Error("",
			logutils.Fields{
				"DeployedStatus": certEnrollmentStatus.DeployedStatus,
				"Error":          err.Error()})
		return readyCount, err
	}
	if strings.ToLower(certEnrollmentStatus.ReadyStatus) != "ready" {
		err := errors.New("enrollment is not ready")
		logutils.Error("",
			logutils.Fields{
				"ReadyStatus": certEnrollmentStatus.ReadyStatus,
				"Error":       err.Error()})
		return readyCount, err
	}

	return certEnrollmentStatus.ReadyCounts["Ready"], nil
}
