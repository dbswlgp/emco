// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package enrollment

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"reflect"
	"strings"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/certmanagerissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/tcsissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
)

// createCertManagerCertificateRequest creates a certificaterequest to generate the certificate
func (enrollCtx *EnrollmentContext) createCertManagerCertificateRequest(ctx context.Context) error {
	name := certmanagerissuer.CertificateRequestName(enrollCtx.ContextID, enrollCtx.CaCert.MetaData.Name, enrollCtx.ClusterGroup.Spec.Provider, enrollCtx.Cluster)
	if _, exists := enrollCtx.Resources.CertificateRequest[name]; exists {
		// a certificaterequest already exists
		return nil
	}

	commonName := strings.Join([]string{enrollCtx.CaCert.MetaData.Name, enrollCtx.ClusterGroup.Spec.Provider, enrollCtx.Cluster, "ca"}, "-")
	if len(enrollCtx.CaCert.Spec.CertificateSigningInfo.Subject.Names.CommonNamePrefix) > 0 {
		commonName = strings.Join([]string{commonName, enrollCtx.CaCert.Spec.CertificateSigningInfo.Subject.Names.CommonNamePrefix}, "-")
	}

	// This needs to be a unique name for each cluster
	enrollCtx.CaCert.Spec.CertificateSigningInfo.Subject.Names.CommonName = commonName

	// check if a cluster specific commonName is available

	if val, err := clm.NewClusterClient().GetClusterKvPairsValue(ctx, enrollCtx.ClusterGroup.Spec.Provider, enrollCtx.Cluster, "csrData", "commonName"); err == nil {
		v, e := module.GetValue(val)
		if e != nil {
			return e
		}

		enrollCtx.CaCert.Spec.CertificateSigningInfo.Subject.Names.CommonName = v
	}

	// generate the private key for the csr
	pemBlock, err := certificate.GeneratePrivateKey(enrollCtx.CaCert.Spec.CertificateSigningInfo.KeySize)
	if err != nil {
		return err
	}

	// parse the RSA key in PKCS #1, ASN.1 DER form
	pk, err := certificate.ParsePrivateKey(pemBlock.Bytes)
	if err != nil {
		return err
	}

	// create a certificate signing request
	request, err := enrollCtx.createCertificateSigningRequest(pk)
	if err != nil {
		return err
	}

	// create the cert-manager CertificateRequest resource
	cr, err := certmanagerissuer.CreateCertificateRequest(enrollCtx.CaCert, name, request)
	if err != nil {
		return err
	}

	// add the CertificateRequest resource in to the appContext
	if err := module.AddResource(ctx, enrollCtx.AppContext, cr, enrollCtx.IssuerHandle, module.ResourceName(cr.ObjectMeta.Name, cr.TypeMeta.Kind)); err != nil {
		return err
	}

	enrollCtx.Resources.CertificateRequest[name] = cr

	enrollCtx.ResOrder = append(enrollCtx.ResOrder, module.ResourceName(cr.ObjectMeta.Name, cr.TypeMeta.Kind))

	// save the PK in mongo
	return enrollCtx.savePrivateKey(ctx, name, base64.StdEncoding.EncodeToString(pem.EncodeToMemory(pemBlock)))

}

// createCertificateSigningRequest creates a certificate signing request
func (enrollCtx *EnrollmentContext) createCertificateSigningRequest(pk *rsa.PrivateKey) ([]byte, error) {
	return certificate.CreateCertificateSigningRequest(x509.CertificateRequest{
		Version:            enrollCtx.CaCert.Spec.CertificateSigningInfo.Version,
		SignatureAlgorithm: certificate.SignatureAlgorithm(enrollCtx.CaCert.Spec.CertificateSigningInfo.Algorithm.SignatureAlgorithm),
		PublicKeyAlgorithm: certificate.PublicKeyAlgorithm(enrollCtx.CaCert.Spec.CertificateSigningInfo.Algorithm.PublicKeyAlgorithm),
		Subject: pkix.Name{
			Country:            enrollCtx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.Country,
			Locality:           enrollCtx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.Locality,
			PostalCode:         enrollCtx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.PostalCode,
			Province:           enrollCtx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.Province,
			StreetAddress:      enrollCtx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.StreetAddress,
			CommonName:         enrollCtx.CaCert.Spec.CertificateSigningInfo.Subject.Names.CommonName,
			Organization:       enrollCtx.CaCert.Spec.CertificateSigningInfo.Subject.Organization.Names,
			OrganizationalUnit: enrollCtx.CaCert.Spec.CertificateSigningInfo.Subject.Organization.Units},
		DNSNames:       enrollCtx.CaCert.Spec.CertificateSigningInfo.DNSNames,
		EmailAddresses: enrollCtx.CaCert.Spec.CertificateSigningInfo.EmailAddresses}, pk)
}

// savePrivateKey saves the rsa private key in mongo
func (enrollCtx *EnrollmentContext) savePrivateKey(ctx context.Context, name, val string) error {
	dbKey := module.DBKey{
		Cert:            enrollCtx.CaCert.MetaData.Name,
		Cluster:         enrollCtx.Cluster,
		ClusterProvider: enrollCtx.ClusterGroup.Spec.Provider,
		ContextID:       enrollCtx.ContextID}
	key := module.Key{
		Name: name,
		Val:  val}

	return module.NewKeyClient(dbKey).Save(ctx, key)
}

// deletePrivateKey delete the rsa private key from mongo
func (enrollCtx *EnrollmentContext) deletePrivateKey(ctx context.Context) error {
	dbKey := module.DBKey{
		Cert:            enrollCtx.CaCert.MetaData.Name,
		Cluster:         enrollCtx.Cluster,
		ClusterProvider: enrollCtx.ClusterGroup.Spec.Provider,
		ContextID:       enrollCtx.ContextID}

	return module.NewKeyClient(dbKey).Delete(ctx)
}

// privateKeyExists verifies the rsa private key exists in mongo
func (enrollCtx *EnrollmentContext) privateKeyExists(ctx context.Context) bool {
	dbKey := module.DBKey{
		Cert:            enrollCtx.CaCert.MetaData.Name,
		Cluster:         enrollCtx.Cluster,
		ClusterProvider: enrollCtx.ClusterGroup.Spec.Provider,
		ContextID:       enrollCtx.ContextID}

	if k, err := module.NewKeyClient(dbKey).Get(ctx); err == nil &&
		!reflect.DeepEqual(k, module.Key{}) {
		return true
	}

	return false
}

// createCertManagerCertificate creates a cert-manager certificate resource to generate the key and the cert
func (enrollCtx *EnrollmentContext) createCertManagerCertificate(ctx context.Context) error {
	name := certmanagerissuer.CertificateName(enrollCtx.ContextID, enrollCtx.CaCert.MetaData.Name, enrollCtx.ClusterGroup.Spec.Provider, enrollCtx.Cluster)
	if _, exists := enrollCtx.Resources.Certificate[name]; exists {
		// a certificate already exists
		return nil
	}

	// create a Secret to store the certificate
	sName := certmanagerissuer.SecretName(enrollCtx.ContextID, enrollCtx.CaCert.MetaData.Name, enrollCtx.ClusterGroup.Spec.Provider, enrollCtx.Cluster)
	// specify the issuer which uses this secret
	iName := tcsissuer.TCSIssuerName(enrollCtx.ContextID, enrollCtx.CaCert.MetaData.Name, enrollCtx.ClusterGroup.Spec.Provider, enrollCtx.Cluster)

	ns := "default"
	if len(enrollCtx.Namespace) > 0 {
		ns = enrollCtx.Namespace
	}

	secretTemplate := cmv1.CertificateSecretTemplate{
		Annotations: map[string]string{
			"issuer-name":      iName,
			"issuer-namespace": ns,
		},
		Labels: map[string]string{
			"emco/deployment-id": fmt.Sprintf("%s-%s", enrollCtx.ContextID, AppName),
		},
	}

	cert, err := certmanagerissuer.CreateCertificate(enrollCtx.CaCert, name, sName, secretTemplate)
	if err != nil {
		return err
	}

	// add the Certificate resource in to the appContext
	if err := module.AddResource(ctx, enrollCtx.AppContext, cert, enrollCtx.IssuerHandle, module.ResourceName(cert.ObjectMeta.Name, cert.TypeMeta.Kind)); err != nil {
		return err
	}

	enrollCtx.Resources.Certificate[name] = cert

	enrollCtx.ResOrder = append(enrollCtx.ResOrder, module.ResourceName(cert.ObjectMeta.Name, cert.TypeMeta.Kind))

	return nil
}
