// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certmanagerissuer

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newSecret returns an instance of the Secret
func newSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{},
		Data:       map[string][]byte{}}
}

// SecretName retun the Secret name
func SecretName(contextID, cert, clusterProvider, cluster string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", contextID, cert, clusterProvider, cluster, "ca")
}

// CreateSecret retun the Secret object
func CreateSecret(name, namespace string, data map[string][]byte) *v1.Secret {
	s := newSecret()
	s.ObjectMeta.Name = name
	s.ObjectMeta.Namespace = namespace

	for key, val := range data {
		s.Data[key] = val
	}

	return s
}
