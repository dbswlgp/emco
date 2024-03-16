// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package distribution

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	tcsv1 "github.com/intel/trusted-certificate-issuer/api/v1alpha1"
	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/certmanagerissuer"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/notifyclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"

	"context"

	v1 "k8s.io/api/core/v1"
)

const (
	AppName string = "cert-distribution"
)

// Instantiate the caCert distribution
func (distCtx *DistributionContext) Instantiate(ctx context.Context) error {
	// create resources for the edge clsuters based on the issuer
	switch distCtx.CaCert.Spec.IssuerRef.Group {
	case "cert-manager.io":
		return distCtx.createCertManagerIssuerResources(ctx)

	default:
		err := errors.New("unsupported Issuer")
		logutils.Error("",
			logutils.Fields{
				"Issuer": distCtx.CaCert.Spec.IssuerRef.Group,
				"Error":  err.Error()})
		return err
	}

}

// Update the caCert distribution appContext
func (distCtx *DistributionContext) Update(ctx context.Context, prevContextID string) error {
	if err := state.UpdateAppContextStatusContextID(ctx, distCtx.ContextID, prevContextID); err != nil {
		logutils.Error("Failed to update appContext status",
			logutils.Fields{
				"ContextID": distCtx.ContextID,
				"AppName":   AppName,
				"Error":     err.Error()})
		return err
	}

	if err := notifyclient.CallRsyncUpdate(ctx, prevContextID, distCtx.ContextID); err != nil {
		logutils.Error("Rsync update failed",
			logutils.Fields{
				"ContextID": distCtx.ContextID,
				"AppName":   AppName,
				"Error":     err.Error()})
		return err
	}

	// subscribe to alerts
	stream, _, err := notifyclient.InvokeReadyNotify(ctx, distCtx.ContextID, distCtx.ClientName)
	if err != nil {
		logutils.Error("Failed to subscribe to alerts",
			logutils.Fields{
				"ContextID":  distCtx.ContextID,
				"ClientName": distCtx.ClientName,
				"AppName":    AppName,
				"Error":      err.Error()})
		return err
	}

	if err := stream.CloseSend(); err != nil {
		logutils.Error("Failed to close the send stream",
			logutils.Fields{
				"ContextID":  distCtx.ContextID,
				"ClientName": distCtx.ClientName,
				"AppName":    AppName,
				"Error":      err.Error()})
		return err
	}

	return nil
}

// Terminate the caCert distribution
func Terminate(ctx context.Context, dbKey interface{}) error {
	sc := module.NewStateClient(dbKey)
	// check the current state of the Instantiation, if any
	contextID, err := sc.VerifyState(ctx, common.Terminate)
	if err != nil {
		return err
	}

	// call resource synchronizer to delete the resources under this appContext
	distCtx := module.CaCertAppContext{
		ContextID: contextID}
	if err := distCtx.CallRsyncUninstall(ctx); err != nil {
		return err
	}

	// update the state object for the caCert distribution resource
	if err := sc.Update(ctx, state.StateEnum.Terminated, contextID, false); err != nil {
		return err
	}

	return nil
}

// createCertManagerIssuerResources creates cert-manager specific resources
// in this case, secret, clusterIssuer

func (distCtx *DistributionContext) createCertManagerIssuerResources(ctx context.Context) error {
	// retrieve enrolled cert-manager resource
	resources, err := certmanagerissuer.RetrieveCertManagerResources(ctx, distCtx.EnrollmentContextID)
	if err != nil {
		return err
	}

	distCtx.CertificateRequests = resources.CertificateRequests
	distCtx.Certificates = resources.Certificates
	distCtx.Secrets = resources.Secrets

	for _, distCtx.ClusterGroup = range distCtx.ClusterGroups {
		// get all the clusters in this clusterGroup
		clusters, err := module.GetClusters(ctx, distCtx.ClusterGroup, distCtx.Project, distCtx.LogicalCloud)
		if err != nil {
			return err
		}

		for _, distCtx.Cluster = range clusters {
			distCtx.ResOrder = []string{}
			distCtx.ClusterHandle, err = distCtx.AppContext.AddCluster(ctx, distCtx.AppHandle,
				strings.Join([]string{distCtx.ClusterGroup.Spec.Provider, distCtx.Cluster}, "+"))
			if err != nil {
				logutils.Error("Failed to add the cluster",
					logutils.Fields{
						"Error": err.Error()})

				if er := distCtx.AppContext.DeleteCompositeApp(ctx); er != nil {
					logutils.Error("Failed to delete the compositeApp",
						logutils.Fields{
							"ContextID": distCtx.ContextID,
							"Error":     er.Error()})
					return er
				}

				return err
			}
			var sgxEnabled bool
			// check whether the cluster is SGX enabled or not
			if val, err := clm.NewClusterClient().GetClusterKvPairsValue(ctx, distCtx.ClusterGroup.Spec.Provider, distCtx.Cluster, "sgx", "enabled"); err == nil {
				v, e := module.GetValue(val)
				if e != nil {
					return e
				}

				if v == "true" {
					sgxEnabled = true
				}
			}

			ready := false
			if sgxEnabled {
				certName := certmanagerissuer.CertificateName(distCtx.EnrollmentContextID, distCtx.CaCert.MetaData.Name, distCtx.ClusterGroup.Spec.Provider, distCtx.Cluster)
				for _, cert := range distCtx.Certificates {
					if cert.ObjectMeta.Name == certName { // to make sure we are creating the resource(s) in the same cluster
						if err := certmanagerissuer.ValidateCertificate(cert); err != nil {
							return err
						}

						// create a TCSIssuer to use the secret and the certificate
						if err := distCtx.createTCSIssuer(ctx, cert.Spec.SecretName, false); err != nil {
							return err
						}

						ready = true
						break
					}
				}
			} else {
				crName := certmanagerissuer.CertificateRequestName(distCtx.EnrollmentContextID, distCtx.CaCert.MetaData.Name, distCtx.ClusterGroup.Spec.Provider, distCtx.Cluster)
				for _, cr := range distCtx.CertificateRequests {
					if cr.ObjectMeta.Name == crName { // to make sure we are creating the resource(s) in the same cluster
						if err := certmanagerissuer.ValidateCertificateRequest(cr); err != nil {
							return err
						}

						// create a Secret to store the certificate
						sName := certmanagerissuer.SecretName(distCtx.ContextID, distCtx.CaCert.MetaData.Name, distCtx.ClusterGroup.Spec.Provider, distCtx.Cluster)
						if err := distCtx.createSecret(ctx, cr, sName, "cert-manager"); err != nil {
							return err
						}

						// create a ClusterIssuer to use the secret and the certificate
						if err := distCtx.createClusterIssuer(ctx, sName); err != nil {
							return err
						}

						ready = true
						break
					}
				}
			}

			if !ready {
				err := errors.New("cert-manager resource is not ready. Update the enrollment")
				logutils.Error("",
					logutils.Fields{
						"Cluster": distCtx.Cluster,
						"Error":   err.Error()})
				return err
			}

			// create service specific resources for this issuer
			if err := distCtx.createServiceResources(ctx); err != nil {
				return err
			}

			if err := module.AddInstruction(ctx, distCtx.AppContext, distCtx.ClusterHandle, distCtx.ResOrder); err != nil {
				return err
			}

			distCtx.Cluster = ""
			distCtx.ResOrder = []string{}
			distCtx.ClusterHandle = nil
		}

		distCtx.ClusterGroup = module.ClusterGroup{}
	}

	return nil
}

// createServiceResources creates the resources based on the service mesh type
func (distCtx *DistributionContext) createServiceResources(ctx context.Context) error {
	var serviceType string
	val, err := clm.NewClusterClient().GetClusterKvPairsValue(ctx, distCtx.ClusterGroup.Spec.Provider, distCtx.Cluster, "serviceMeshInfo", "serviceType")
	if err == nil {
		v, e := module.GetValue(val)
		if e != nil {
			return e
		}

		serviceType = v
	}

	if err != nil &&
		err.Error() != "Cluster key value pair not found" &&
		err.Error() != "Cluster KV pair key value not found" {
		return err
	}

	switch serviceType {
	case "istio":
		return distCtx.createIstioServiceResourcess(ctx)
	default:
		return distCtx.createIstioServiceResourcess(ctx)
	}

}

// createIstioServiceResources creates the istio resources on the edge cluster
// in this case, proxyConfig, kncc resource to update the istio configMap

func (distCtx *DistributionContext) createIstioServiceResourcess(ctx context.Context) error {
	switch distCtx.CaCert.Spec.IssuerRef.Group {
	case "cert-manager.io":
		var sgxEnabled bool
		var issuerName string
		var secret v1.Secret
		// check whether the cluster is SGX enabled or not
		if val, err := clm.NewClusterClient().GetClusterKvPairsValue(ctx, distCtx.ClusterGroup.Spec.Provider, distCtx.Cluster, "sgx", "enabled"); err == nil {
			v, e := module.GetValue(val)
			if e != nil {
				return e
			}

			if v == "true" {
				sgxEnabled = true
			}
		}

		if sgxEnabled {
			// retrieve the tcsissuer
			if issuer := distCtx.retrieveTCSIssuer(distCtx.Cluster); !reflect.DeepEqual(issuer, tcsv1.TCSIssuer{}) {
				issuerName = issuer.ObjectMeta.Name
			}
			// retrieve the tcsissuer secret
			sName := certmanagerissuer.SecretName(distCtx.EnrollmentContextID, distCtx.CaCert.MetaData.Name, distCtx.ClusterGroup.Spec.Provider, distCtx.Cluster)
			for _, s := range distCtx.Secrets {
				if s.ObjectMeta.Name == sName {
					secret = s
					break
				}
			}
		} else {
			// retrieve the clusterissuer
			if issuer := distCtx.retrieveClusterIssuer(distCtx.Cluster); !reflect.DeepEqual(issuer, cmv1.ClusterIssuer{}) {
				issuerName = issuer.ObjectMeta.Name
			}
			// retrieve the clusterissuer secret
			secret = *distCtx.retrieveSecret(distCtx.Cluster)

		}

		if len(issuerName) == 0 {
			// a clusterIssuer is not available, return error
			err := errors.New("A clusterIssuer is not available for the cluster")
			if sgxEnabled {
				err = errors.New("A tcsIssuer is not available for the cluster")
			}
			logutils.Error("",
				logutils.Fields{
					"Cluster": distCtx.Cluster,
					"Error":   err.Error()})
			return err
		}

		if reflect.DeepEqual(secret, v1.Secret{}) {
			// a secret is not available, return error
			err := errors.New("A secret is not available for the cluster")
			logutils.Error("",
				logutils.Fields{
					"Cluster": distCtx.Cluster,
					"Error":   err.Error()})
			return err
		}

		var (
			key,
			value string
		)

		if len(distCtx.Namespace) > 0 {
			// create a proxyconfig for this namespace
			if err := distCtx.createProxyConfig(ctx, issuerName); err != nil {
				return err
			}
		}

		// create a kncc patch config using the secret created fot this issuer
		key = "mesh:\n  caCertificates\n"
		if sgxEnabled {
			pem := strings.Replace(string(bytes.Join([][]byte{secret.Data[v1.TLSCertKey], secret.Data[v1.ServiceAccountRootCAKey]},
				[]byte{})), "\n", "\n    ", -1) // create the caCert chain
			value = fmt.Sprintf("- certSigners:\n  - tcsissuer.tcs.intel.com/default.%s\n  pem: |\n    %s",
				issuerName, pem)
			if len(distCtx.Namespace) > 0 {
				// set the issuer name with the namespace
				value = fmt.Sprintf("- certSigners:\n  - tcsissuer.tcs.intel.com/%s.%s\n  pem: |\n    %s",
					distCtx.Namespace, issuerName, pem)
			}

		} else {
			pem := strings.Replace(string(secret.Data[v1.TLSCertKey]), "\n", "\n    ", -1)
			value = fmt.Sprintf("- certSigners:\n  - clusterissuers.cert-manager.io/%s\n  pem: |\n    %s",
				issuerName, pem)
		}

		if err := distCtx.createKnccConfig(ctx, "istio-system", "istio", "istio-system",
			[]map[string]string{{key: value}}); err != nil {
			return err
		}

		if len(distCtx.Namespace) == 0 {
			// use this issuer as the dfault issuer for the cluster
			key = "mesh:\n  defaultConfig:\n    proxyMetadata:\n      ISTIO_META_CERT_SIGNER\n"
			value = issuerName
			if sgxEnabled {
				value = fmt.Sprintf("default.%s", issuerName)
			}
			if err := distCtx.createKnccConfig(ctx, "istio-system", "istio", "istio-system",
				[]map[string]string{{key: value}}); err != nil {
				return err
			}
		}

		return nil
	}

	err := errors.New("Unsupported Issuer")
	logutils.Error("",
		logutils.Fields{
			"Cluster": distCtx.Cluster,
			"Error":   err.Error()})
	return err
}
