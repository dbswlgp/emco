// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package tcsissuer

import (
	"fmt"

	tcsv1 "github.com/intel/trusted-certificate-issuer/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newClusterIssuer returns an instance of the TCSIssuer
func newTCSIssuer() *tcsv1.TCSIssuer {
	// by default, generate a self-signed certificate for the TCSIssuer
	// set the value to false to generate the quote and use an external key server
	selfSign := true
	return &tcsv1.TCSIssuer{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tcs.intel.com/v1alpha1",
			Kind:       "TCSIssuer",
		},
		Spec: tcsv1.TCSIssuerSpec{
			SelfSignCertificate: &selfSign,
		},
		Status: tcsv1.TCSIssuerStatus{
			Conditions: []tcsv1.TCSIssuerCondition{},
		},
	}
}

// TCSIssuerName retun the TCSIssuer name
func TCSIssuerName(contextID, cert, clusterProvider, cluster string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", contextID, cert, clusterProvider, cluster, "tcsissuer")
}

// CreateTCSIssuer retun the TCSIssuer object
func CreateTCSIssuer(name, secret string, selfSign bool) *tcsv1.TCSIssuer {
	i := newTCSIssuer()
	i.ObjectMeta.Name = name
	i.Spec.SecretName = secret
	if !selfSign {
		i.Spec.SelfSignCertificate = &selfSign
	}

	return i
}
