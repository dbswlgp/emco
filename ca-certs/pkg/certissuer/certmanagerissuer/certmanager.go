// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certmanagerissuer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	v1 "k8s.io/api/core/v1"
)

type Resources struct {
	Certificates        []cmv1.Certificate
	CertificateRequests []cmv1.CertificateRequest
	Secrets             []v1.Secret
}

// RetrieveCertManagerResources retrieves the cert-manager resources created by the caCert enrollment
func RetrieveCertManagerResources(ctx context.Context, contextID string) (Resources, error) {
	var (
		appContext                 appcontext.AppContext
		clusters                   []string
		certificateRequestStatuses []cmv1.CertificateRequest
		certificateStatuses        []cmv1.Certificate
		secretStatus               []v1.Secret
		certmanagerResources       Resources
	)

	// load the appContext
	_, err := appContext.LoadAppContext(ctx, contextID)
	if err != nil {
		logutils.Error("Failed to load the appContext",
			logutils.Fields{
				"ContextID": contextID,
				"Error":     err.Error()})
		return certmanagerResources, err
	}

	// get the app instruction for 'order'
	appsOrder, err := appContext.GetAppInstruction(ctx, "order")
	if err != nil {
		logutils.Error("Failed to get the app instruction for the 'order' instruction type",
			logutils.Fields{
				"ContextID": contextID,
				"Error":     err.Error()})
		return certmanagerResources, err
	}

	var appList map[string][]string
	if err := json.Unmarshal([]byte(appsOrder.(string)), &appList); err != nil {
		logutils.Error("Failed to unmarshal app order",
			logutils.Fields{
				"ContextID": contextID,
				"Error":     err.Error()})
		return certmanagerResources, err
	}

	for _, app := range appList["apporder"] {
		//  get all the clusters associated with the app
		clusters, err = appContext.GetClusterNames(ctx, app)
		if err != nil {
			logutils.Error("Failed to get cluster names",
				logutils.Fields{
					"App":       app,
					"ContextID": contextID,
					"Error":     err.Error()})
			return certmanagerResources, err
		}

		for _, cluster := range clusters {
			// get the resources
			resources, err := appContext.GetResourceNames(ctx, app, cluster)
			if err != nil {
				logutils.Error("Failed to get the resource names",
					logutils.Fields{
						"App":       app,
						"Cluster":   cluster,
						"ContextID": contextID,
						"Error":     err.Error()})
				return certmanagerResources, err
			}

			// get the cluster handle
			cHandle, err := appContext.GetClusterHandle(ctx, app, cluster)
			if err != nil {
				logutils.Error("Failed to get the cluster handle",
					logutils.Fields{
						"App":       app,
						"Cluster":   cluster,
						"ContextID": contextID,
						"Error":     err.Error()})
				return certmanagerResources, err
			}

			// get the cluster status handle
			sHandle, err := appContext.GetLevelHandle(ctx, cHandle, "status")
			if err != nil {
				logutils.Error("Failed to get the handle of level 'status'",
					logutils.Fields{
						"App":       app,
						"Cluster":   cluster,
						"ContextID": contextID,
						"Error":     err.Error()})
				return certmanagerResources, err
			}

			// wait for the resources to be created and available in the monitor resource bundle state
			// retry if the resources satatus are not available
			statusReady := false
			for !statusReady {
				// get the value of 'status' handle
				val, err := appContext.GetValue(ctx, sHandle)
				if err != nil {
					logutils.Error("Failed to get the value of 'status' handle",
						logutils.Fields{
							"App":       app,
							"Cluster":   cluster,
							"ContextID": contextID,
							"Error":     err.Error()})
					continue
				}

				s := certissuer.ResourceBundleStateStatus{}
				if err := json.Unmarshal([]byte(val.(string)), &s); err != nil {
					logutils.Error("Failed to unmarshal 'resourcebundlestatestatus' status",
						logutils.Fields{
							"App":       app,
							"Cluster":   cluster,
							"ContextID": contextID,
							"Error":     err.Error()})
					return certmanagerResources, err
				}

				if len(s.ResourceStatuses) == 0 {
					continue
				}

				certificateRequestStatuses = []cmv1.CertificateRequest{}
				// for each resource make sure the status is available
				for _, resource := range resources {
				l:
					for _, rStatus := range s.ResourceStatuses {
						if module.ResourceName(rStatus.Name, rStatus.Kind) == resource {
							if len(rStatus.Res) == 0 {
								logutils.Warn(fmt.Sprintf("resource status does not contain any details of %s", rStatus.Name),
									logutils.Fields{})
								break
							}

							data, err := base64.StdEncoding.DecodeString(rStatus.Res)
							if err != nil {
								logutils.Error("Failed to decode the resource status response",
									logutils.Fields{
										"App":       app,
										"Cluster":   cluster,
										"ContextID": contextID,
										"Error":     err.Error()})
								return certmanagerResources, err
							}

							switch rStatus.Kind {
							case "CertificateRequest":
								status := cmv1.CertificateRequest{}
								if err := json.Unmarshal(data, &status); err != nil {
									logutils.Error("Failed to unmarshal the CertificateRequest resource",
										logutils.Fields{
											"App":       app,
											"Cluster":   cluster,
											"ContextID": contextID,
											"Error":     err.Error()})
									return certmanagerResources, err
								}
								certificateRequestStatuses = append(certificateRequestStatuses, status)
								break l
							case "Certificate": // cluster is sgx enabled
								cert := cmv1.Certificate{}
								if err := json.Unmarshal(data, &cert); err != nil {
									logutils.Error("Failed to unmarshal the Certificate resource",
										logutils.Fields{
											"App":       app,
											"Cluster":   cluster,
											"ContextID": contextID,
											"Error":     err.Error()})
									return certmanagerResources, err
								}
								certificateStatuses = append(certificateStatuses, cert)
								// get the secret associated with this certificate
								for _, rs := range s.ResourceStatuses {
									if rs.Name == cert.Spec.SecretName {
										data, err := base64.StdEncoding.DecodeString(rs.Res)
										if err != nil {
											logutils.Error("Failed to decode the resource status response",
												logutils.Fields{
													"App":       app,
													"Cluster":   cluster,
													"ContextID": contextID,
													"Error":     err.Error()})
											return certmanagerResources, err
										}

										s := v1.Secret{}
										if err := json.Unmarshal(data, &s); err != nil {
											logutils.Error("Failed to unmarshal the Secret",
												logutils.Fields{
													"App":       app,
													"Cluster":   cluster,
													"ContextID": contextID,
													"Error":     err.Error()})
											return certmanagerResources, err
										}

										secretStatus = append(secretStatus, s)
										break
									}
								}
								break l
							}
						}
					}
				}

				if len(resources) == len(certificateRequestStatuses)+len(certificateStatuses) {
					logutils.Info(fmt.Sprintf("resourcebundlestatestatus contains the resources for App: %s, ClusterGroup: %s and ContextID: %s", app, cluster, contextID),
						logutils.Fields{})

					certmanagerResources.CertificateRequests = append(certmanagerResources.CertificateRequests, certificateRequestStatuses...)
					certmanagerResources.Certificates = append(certmanagerResources.Certificates, certificateStatuses...)
					certmanagerResources.Secrets = append(certmanagerResources.Secrets, secretStatus...)

					// At this point we assume we have the resources created and the status is available
					statusReady = true
				}
			}
		}
	}

	return certmanagerResources, nil
}
