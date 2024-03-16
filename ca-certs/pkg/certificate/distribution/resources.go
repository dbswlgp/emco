// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package distribution

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	tcsv1 "github.com/intel/trusted-certificate-issuer/api/v1alpha1"
	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/certmanagerissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/tcsissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/service/istioservice"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/service/knccservice"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	v1 "k8s.io/api/core/v1"
)

// createSecret creates a secret to store the certificate
func (distCtx *DistributionContext) createSecret(ctx context.Context, cr cmv1.CertificateRequest, name, namespace string) error {
	if s, exists := distCtx.Resources.Secret[name]; exists {
		// a secret already exists, update the context resource order
		distCtx.ResOrder = append(distCtx.ResOrder, module.ResourceName(s.ObjectMeta.Name, s.Kind))
		return nil
	}
	// retrieve the Private Key from mongo
	key, err := distCtx.retrievePrivateKey(ctx)
	if err != nil {
		return err
	}

	data := map[string][]byte{}
	data[v1.TLSCertKey] = bytes.Join([][]byte{cr.Status.Certificate, cr.Status.CA}, []byte{}) // create the caCert chain
	data[v1.TLSPrivateKeyKey] = key

	s := certmanagerissuer.CreateSecret(name, namespace, data)
	if err := module.AddResource(ctx, distCtx.AppContext, s, distCtx.ClusterHandle, module.ResourceName(s.ObjectMeta.Name, s.Kind)); err != nil {
		return err
	}

	distCtx.ResOrder = append(distCtx.ResOrder, module.ResourceName(s.ObjectMeta.Name, s.Kind))

	distCtx.Resources.Secret[name] = s // this is needed to create the kncc config

	return nil
}

// createClusterIssuer creates a clusterIssuer to issue the certificates
func (distCtx *DistributionContext) createClusterIssuer(ctx context.Context, secretName string) error {
	iName := certmanagerissuer.ClusterIssuerName(distCtx.ContextID, distCtx.CaCert.MetaData.Name, distCtx.ClusterGroup.Spec.Provider, distCtx.Cluster)
	if i, exists := distCtx.Resources.ClusterIssuer[iName]; exists {
		// a clusterIssuer already exists, update the context resource order
		distCtx.ResOrder = append(distCtx.ResOrder, module.ResourceName(i.ObjectMeta.Name, i.Kind))
		return nil
	}

	i := certmanagerissuer.CreateClusterIssuer(iName, secretName)
	if err := module.AddResource(ctx, distCtx.AppContext, i, distCtx.ClusterHandle, module.ResourceName(i.ObjectMeta.Name, i.Kind)); err != nil {
		return err
	}

	distCtx.ResOrder = append(distCtx.ResOrder, module.ResourceName(i.ObjectMeta.Name, i.Kind))
	distCtx.Resources.ClusterIssuer[iName] = i // this is needed to create the proxyconfig

	return nil
}

// createProxyConfig creates a proxyConfig to control the traffic between workloads

func (distCtx *DistributionContext) createProxyConfig(ctx context.Context, issuer string) error {
	pcName := istioservice.ProxyConfigName(distCtx.ContextID, distCtx.CaCert.MetaData.Name, distCtx.ClusterGroup.Spec.Provider, distCtx.Cluster, distCtx.Namespace)
	if pc, exists := distCtx.Resources.ProxyConfig[pcName]; exists {
		// a proxyConfig already exists, update the context resource order
		distCtx.ResOrder = append(distCtx.ResOrder, module.ResourceName(pc.MetaData.Name, pc.Kind))
		return nil
	}

	environmentVariables := map[string]string{}

	environmentVariables["ISTIO_META_CERT_SIGNER"] = issuer
	pc := istioservice.CreateProxyConfig(pcName, distCtx.Namespace, environmentVariables)

	if err := module.AddResource(ctx, distCtx.AppContext, pc, distCtx.ClusterHandle, module.ResourceName(pc.MetaData.Name, pc.Kind)); err != nil {
		return err
	}

	distCtx.Resources.ProxyConfig[pcName] = pc

	distCtx.ResOrder = append(distCtx.ResOrder, module.ResourceName(pc.MetaData.Name, pc.Kind))

	for _, p := range distCtx.Resources.ProxyConfig {
		// update the context resource order with any other proxyConfig created for the same appContext, caCert, clusterProvider and cluster
		if strings.Contains(p.MetaData.Name, fmt.Sprintf("%s-%s-%s-%s", distCtx.ContextID, distCtx.CaCert.MetaData.Name, distCtx.ClusterGroup.Spec.Provider, distCtx.Cluster)) {
			exists := false
			for _, o := range distCtx.ResOrder {
				if o == module.ResourceName(p.MetaData.Name, p.Kind) {
					exists = true
					break
				}
			}

			if !exists {
				distCtx.ResOrder = append(distCtx.ResOrder, module.ResourceName(p.MetaData.Name, p.Kind))
			}
		}
	}

	return nil
}

// createKnccConfig creates a kncc config patch to update the configMap
func (distCtx *DistributionContext) createKnccConfig(ctx context.Context, namespace, resourceName, resourceNamespace string,
	patch []map[string]string) error {
	cName := knccservice.KnccConfigName(distCtx.ContextID, distCtx.CaCert.MetaData.Name, distCtx.ClusterGroup.Spec.Provider, distCtx.Cluster)
	if c, exists := distCtx.Resources.KnccConfig[cName]; exists {
		// a KnccConfig already exists, update the config
		return distCtx.updateKnccConfig(ctx, patch, c)
	}

	c := knccservice.CreateKnccConfig(cName, namespace, resourceName, resourceNamespace, patch)
	if err := module.AddResource(ctx, distCtx.AppContext, c, distCtx.ClusterHandle, module.ResourceName(c.ObjectMeta.Name, c.TypeMeta.Kind)); err != nil {
		return err
	}

	distCtx.Resources.KnccConfig[cName] = c

	distCtx.ResOrder = append(distCtx.ResOrder, module.ResourceName(c.ObjectMeta.Name, c.TypeMeta.Kind))

	return nil
}

// retrievePrivateKey retrieves the rsa private key from mongo
func (distCtx *DistributionContext) retrievePrivateKey(ctx context.Context) ([]byte, error) {
	dbKey := module.DBKey{
		Cert:            distCtx.CaCert.MetaData.Name,
		Cluster:         distCtx.Cluster,
		ClusterProvider: distCtx.ClusterGroup.Spec.Provider,
		ContextID:       distCtx.EnrollmentContextID}

	key, err := module.NewKeyClient(dbKey).Get(ctx)
	if err != nil {
		return []byte{}, err
	}

	if key.Name != certmanagerissuer.CertificateRequestName(distCtx.EnrollmentContextID, distCtx.CaCert.MetaData.Name, distCtx.ClusterGroup.Spec.Provider, distCtx.Cluster) {
		err := errors.New("PrivateKey not found")
		logutils.Error("",
			logutils.Fields{
				"CaCert": distCtx.CaCert.MetaData.Name,
				"Error":  err.Error()})
		return []byte{}, err
	}

	return base64.StdEncoding.DecodeString(key.Val)
}

// retrieveClusterIssuer returns the specific issuer from the distribution resources list
func (distCtx *DistributionContext) retrieveClusterIssuer(cluster string) *cmv1.ClusterIssuer {
	var iName string
	for _, issuer := range distCtx.Resources.ClusterIssuer {
		iName = certmanagerissuer.ClusterIssuerName(distCtx.ContextID, distCtx.CaCert.MetaData.Name, distCtx.ClusterGroup.Spec.Provider, cluster)
		if issuer.ObjectMeta.Name == iName {
			return issuer
		}
	}

	return &cmv1.ClusterIssuer{}
}

// retrieveSecret returns the specific secret from the distribution resources list
func (distCtx *DistributionContext) retrieveSecret(cluster string) *v1.Secret {
	var sName string
	for _, s := range distCtx.Resources.Secret {
		sName = certmanagerissuer.SecretName(distCtx.ContextID, distCtx.CaCert.MetaData.Name, distCtx.ClusterGroup.Spec.Provider, cluster)
		if s.ObjectMeta.Name == sName {
			return s
		}
	}

	return &v1.Secret{}
}

// updateKnccConfig updates the kncc resource patch
func (distCtx *DistributionContext) updateKnccConfig(ctx context.Context, patch []map[string]string, c *knccservice.Config) error {
	// update kncc config
	knccservice.UpdateKnccConfig(patch, c)
	// update resource appContext
	if err := module.AddResource(ctx, distCtx.AppContext, c, distCtx.ClusterHandle, module.ResourceName(c.ObjectMeta.Name, c.TypeMeta.Kind)); err != nil {
		return err
	}
	// store the new config
	distCtx.Resources.KnccConfig[c.ObjectMeta.Name] = c
	// update the resource order
	distCtx.ResOrder = append(distCtx.ResOrder, module.ResourceName(c.ObjectMeta.Name, c.TypeMeta.Kind))

	return nil
}

// createTCSIssuer creates a TCSIssuer to issue the certificates
func (distCtx *DistributionContext) createTCSIssuer(ctx context.Context, secret string, selfSign bool) error {
	iName := tcsissuer.TCSIssuerName(distCtx.EnrollmentContextID, distCtx.CaCert.MetaData.Name, distCtx.ClusterGroup.Spec.Provider, distCtx.Cluster)
	if i, exists := distCtx.Resources.TCSIssuer[iName]; exists {
		// a tcsIssuer already exists, update the context resource order
		distCtx.ResOrder = append(distCtx.ResOrder, module.ResourceName(i.ObjectMeta.Name, i.Kind))
		return nil
	}

	i := tcsissuer.CreateTCSIssuer(iName, secret, selfSign)
	if err := module.AddResource(ctx, distCtx.AppContext, i, distCtx.ClusterHandle, module.ResourceName(i.ObjectMeta.Name, i.Kind)); err != nil {
		return err
	}

	distCtx.ResOrder = append(distCtx.ResOrder, module.ResourceName(i.ObjectMeta.Name, i.Kind))
	distCtx.Resources.TCSIssuer[iName] = i // this is needed to create the proxyconfig

	return nil
}

// retrieveTCSIssuer returns the specific tcsissuer from the distribution resources list
func (distCtx *DistributionContext) retrieveTCSIssuer(cluster string) *tcsv1.TCSIssuer {
	var iName string
	for _, issuer := range distCtx.Resources.TCSIssuer {
		iName = tcsissuer.TCSIssuerName(distCtx.EnrollmentContextID, distCtx.CaCert.MetaData.Name, distCtx.ClusterGroup.Spec.Provider, cluster)
		if issuer.ObjectMeta.Name == iName {
			return issuer
		}
	}

	return &tcsv1.TCSIssuer{}
}
