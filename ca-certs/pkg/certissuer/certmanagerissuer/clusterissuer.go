// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certmanagerissuer

import (
	"fmt"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newClusterIssuer returns an instance of the ClusterIssuer
func newClusterIssuer() *cmv1.ClusterIssuer {
	return &cmv1.ClusterIssuer{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "ClusterIssuer",
		},
		Spec: cmv1.IssuerSpec{
			IssuerConfig: cmv1.IssuerConfig{
				CA: &cmv1.CAIssuer{},
			},
		},
	}
}

// ClusterIssuerName retun the ClusterIssuer name
func ClusterIssuerName(contextID, cert, clusterProvider, cluster string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", contextID, cert, clusterProvider, cluster, "issuer")
}

// CreateClusterIssuer retun the cert-manager ClusterIssuer object
func CreateClusterIssuer(name, secret string) *cmv1.ClusterIssuer {
	i := newClusterIssuer()
	i.ObjectMeta.Name = name
	i.Spec.IssuerConfig.CA.SecretName = secret

	return i
}
