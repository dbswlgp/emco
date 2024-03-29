// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "edgemetricscollector/pkg/generated/clientset/versioned/typed/edgemetricscollector/v1alpha1"

	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeEdgemetricscollectorV1alpha1 struct {
	*testing.Fake
}

func (c *FakeEdgemetricscollectorV1alpha1) MetricsCollectors(namespace string) v1alpha1.MetricsCollectorInterface {
	return &FakeMetricsCollectors{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeEdgemetricscollectorV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
