// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

var _ = Describe("Create ClusterGroup",
	func() {
		BeforeEach(func() {
			populateClusterGroupTestData()
		})
		Context("create a clusterGroup for a clusterProvider", func() {
			It("returns the clusterGroup, no error and, the exists flag is false", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mClusterGroup := mockClusterGroup("new-clusterGroup-1")
				cg, cExists, err := client.CreateClusterGroup(ctx, mClusterGroup, "cert1", "provider1", true)
				validateError(err, "")
				Expect(mClusterGroup).To(Equal(cg))
				Expect(cExists).To(Equal(false))
				Expect(len(mockdb.Items)).To(Equal(l + 1))
			})
		})
		Context("create a clusterGroup for a clusterProvider that already exists", func() {
			It("returns an error, no clusterGroup and, the exists flag is true", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mClusterGroup := mockClusterGroup("test-clusterGroup-1")
				cg, cExists, err := client.CreateClusterGroup(ctx, mClusterGroup, "cert1", "provider1", true)
				validateError(err, module.CaCertClusterGroupAlreadyExists)
				Expect(module.ClusterGroup{}).To(Equal(cg))
				Expect(cExists).To(Equal(true))
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
	},
)

var _ = Describe("Delete ClusterGroup",
	func() {
		BeforeEach(func() {
			populateClusterGroupTestData()
		})
		Context("delete an existing clusterGroup", func() {
			It("returns no error and delete the entry from the db", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				err := client.DeleteClusterGroup(ctx, "cert1", "test-clusterGroup-1", "provider1")
				validateError(err, "")
				Expect(len(mockdb.Items)).To(Equal(l - 1))
			})
		})
		Context("delete a nonexisting clusterGroup", func() {
			It("returns an error and no change in the db", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mockdb.Err = errors.New("db Remove resource not found")
				err := client.DeleteClusterGroup(ctx, "cert1", "non-existing-cluster", "provider1")
				validateError(err, "db Remove resource not found")
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
	},
)

var _ = Describe("Get All ClusterGroups",
	func() {
		BeforeEach(func() {
			populateClusterGroupTestData()
		})
		Context("get all the clusterGroups", func() {
			It("returns all the clusterGroups, no error", func() {
				ctx := context.Background()
				clusters, err := client.GetAllClusterGroups(ctx, "cert1", "provider1")
				validateError(err, "")
				Expect(len(clusters)).To(Equal(len(mockdb.Items)))
			})
		})
		Context("get all the clusterGroups without creating any", func() {
			It("returns an empty array, no error", func() {
				ctx := context.Background()
				mockdb.Items = []map[string]map[string][]byte{}
				clusters, err := client.GetAllClusterGroups(ctx, "cert1", "provider1")
				validateError(err, "")
				Expect(len(clusters)).To(Equal(0))
			})
		})
	},
)

var _ = Describe("Get ClusterGroup",
	func() {
		BeforeEach(func() {
			populateClusterGroupTestData()
		})
		Context("get an existing clusterGroup", func() {
			It("returns the clusterGroup, no error", func() {
				ctx := context.Background()
				cluster, err := client.GetClusterGroup(ctx, "cert1", "test-clusterGroup-1", "provider1")
				validateError(err, "")
				validateClusterGroup(cluster, mockClusterGroup("test-clusterGroup-1"))
			})
		})
		Context("get a nonexisting clusterGroup", func() {
			It("returns an error, no clusterGroup", func() {
				ctx := context.Background()
				cluster, err := client.GetClusterGroup(ctx, "cert1", "non-existing-cluster", "provider1")
				validateError(err, module.CaCertClusterGroupNotFound)
				validateClusterGroup(cluster, module.ClusterGroup{})
			})
		})
	},
)

// validateClusterGroup
func validateClusterGroup(in, out module.ClusterGroup) {
	Expect(in).To(Equal(out))
}

// mockClusterGroup
func mockClusterGroup(name string) module.ClusterGroup {
	return module.ClusterGroup{
		MetaData: types.Metadata{
			Name:        name,
			Description: "test cluster",
			UserData1:   "some user data 1",
			UserData2:   "some user data 2",
		},
	}
}

// populateClusterGroupTestData
func populateClusterGroupTestData() {
	ctx := context.Background()
	mockdb.Err = nil
	mockdb.Items = []map[string]map[string][]byte{}
	mockdb.MarshalErr = nil

	// cluster 1
	cluster := mockClusterGroup("test-clusterGroup-1")
	cpKey := clusterprovider.ClusterGroupKey{
		Cert:            "cert1",
		ClusterGroup:    "test-clusterGroup-1",
		ClusterProvider: "provider1"}
	_ = mockdb.Insert(ctx, "resources", cpKey, nil, "data", cluster)

	// cluster 2
	cluster = mockClusterGroup("test-clusterGroup-2")
	cpKey = clusterprovider.ClusterGroupKey{
		Cert:            "cert1",
		ClusterGroup:    "test-clusterGroup-2",
		ClusterProvider: "provider1"}
	_ = mockdb.Insert(ctx, "resources", cpKey, nil, "data", cluster)

	// cluster 3
	cluster = mockClusterGroup("test-clusterGroup-3")
	cpKey = clusterprovider.ClusterGroupKey{
		Cert:            "cert1",
		ClusterGroup:    "test-clusterGroup-3",
		ClusterProvider: "provider1"}
	_ = mockdb.Insert(ctx, "resources", cpKey, nil, "data", cluster)
}

func validateError(err error, message string) {
	if len(message) == 0 {
		Expect(err).NotTo(HaveOccurred())
		Expect(err).To(BeNil())
		return
	}
	Expect(err.Error()).To(ContainSubstring(message))
}
