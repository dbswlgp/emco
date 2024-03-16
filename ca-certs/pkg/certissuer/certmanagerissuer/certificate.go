// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certmanagerissuer

import (
	"errors"
	"fmt"
	"time"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newCertificate returns an instance of the Certificate
func newCertificate() *cmv1.Certificate {
	return &cmv1.Certificate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "Certificate",
		},
	}
}

// CertificateName retun the Certificate name
func CertificateName(contextID, cert, clusterProvider, cluster string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", contextID, cert, clusterProvider, cluster, "cert")
}

// CreateCertificate retun the cert-manager Certificate object
func CreateCertificate(caCert module.CaCert, name, secret string, secretTemplate cmv1.CertificateSecretTemplate) (*cmv1.Certificate, error) {
	// parse certificate duration
	duration, err := time.ParseDuration(caCert.Spec.Duration)
	if err != nil {
		logutils.Error("Failed to parse the certificate duration",
			logutils.Fields{
				"Duration": caCert.Spec.Duration,
				"Error":    err.Error()})
		return nil, err
	}

	c := newCertificate()
	c.ObjectMeta = metav1.ObjectMeta{
		Name: name,
		// Namespace: "cert-manager",
	}
	c.Spec = cmv1.CertificateSpec{
		CommonName: caCert.Spec.Certificate.CommonName,
		Duration: &metav1.Duration{
			Duration: duration,
		},
		IsCA: caCert.Spec.IsCA,
		IssuerRef: cmmetav1.ObjectReference{
			Name:  caCert.Spec.IssuerRef.Name,
			Kind:  caCert.Spec.IssuerRef.Kind,
			Group: caCert.Spec.IssuerRef.Group,
		},
		SecretName:     secret,
		SecretTemplate: &secretTemplate,
	}

	return c, nil
}

// ValidateCertificate validate the certificate status
func ValidateCertificate(cert cmv1.Certificate) error {
	approved := false
	for _, state := range cert.Status.Conditions {
		if state.Status == cmmetav1.ConditionTrue &&
			state.Type == cmv1.CertificateConditionReady {
			approved = true
			break
		}
	}

	if !approved {
		err := errors.New("the certificate is not yet approved by the root CA")
		logutils.Error("",
			logutils.Fields{
				"Certificate": cert.ObjectMeta.Name,
				"Error":       err.Error()})
		return err
	}
	return nil
}
