// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"context"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
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
				key := clusterprovider.ClusterGroupKey{
					Cert:            "cert1",
					ClusterGroup:    "new-clusterGroup-1",
					ClusterProvider: "provider1"}
				client := module.NewClusterGroupClient(key)
				cg, cExists, err := client.CreateClusterGroup(ctx, mClusterGroup, true)
				validateError(err, "")
				Expect(cExists).To(Equal(false))
				Expect(cg).To(Equal(mClusterGroup))
				Expect(len(mockdb.Items)).To(Equal(l + 1))
			})
		})
		Context("create a clusterGroup for a clusterProvider that already exists", func() {
			It("returns an error, no clusterGroup and, the exists flag is true", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mClusterGroup := mockClusterGroup("test-clusterGroup-1")
				key := clusterprovider.ClusterGroupKey{
					Cert:            "cert1",
					ClusterGroup:    "test-clusterGroup-1",
					ClusterProvider: "provider1"}
				client := module.NewClusterGroupClient(key)
				cg, cExists, err := client.CreateClusterGroup(ctx, mClusterGroup, true)
				validateError(err, module.CaCertClusterGroupAlreadyExists)
				Expect(cg).To(Equal(module.ClusterGroup{}))
				Expect(cExists).To(Equal(true))
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
		Context("create a clusterGroup for a logicalCloud", func() {
			It("returns the clusterGroup, no error and, the exists flag is false", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mClusterGroup := mockClusterGroup("new-clusterGroup-1")
				key := logicalcloud.ClusterGroupKey{
					Cert:         "cert1",
					ClusterGroup: "new-clusterGroup-1",
					Project:      "proj1"}
				client := module.NewClusterGroupClient(key)
				cg, cExists, err := client.CreateClusterGroup(ctx, mClusterGroup, true)
				validateError(err, "")
				Expect(cExists).To(Equal(false))
				Expect(cg).To(Equal(mClusterGroup))
				Expect(len(mockdb.Items)).To(Equal(l + 1))
			})
		})
		Context("create a clusterGroup for a logicalCloud that already exists", func() {
			It("returns an error, no clusterGroup and, the exists flag is true", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mClusterGroup := mockClusterGroup("test-clusterGroup-4")
				key := logicalcloud.ClusterGroupKey{
					Cert:               "cert1",
					CaCertLogicalCloud: "lc1",
					ClusterGroup:       "test-clusterGroup-4",
					Project:            "proj1"}
				client := module.NewClusterGroupClient(key)
				cg, cExists, err := client.CreateClusterGroup(ctx, mClusterGroup, true)
				validateError(err, module.CaCertClusterGroupAlreadyExists)
				Expect(cg).To(Equal(module.ClusterGroup{}))
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
		Context("delete an existing clusterGroup, clusterProvider", func() {
			It("returns no error and delete the entry from the db", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				key := clusterprovider.ClusterGroupKey{
					Cert:            "cert1",
					ClusterGroup:    "test-clusterGroup-1",
					ClusterProvider: "provider1"}
				client := module.NewClusterGroupClient(key)
				err := client.DeleteClusterGroup(ctx)
				validateError(err, "")
				Expect(len(mockdb.Items)).To(Equal(l - 1))
			})
		})
		Context("delete a nonexisting clusterGroup, clusterProvider", func() {
			It("returns an error and no change in the db", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mockdb.Err = errors.New("db Remove resource not found")
				key := clusterprovider.ClusterGroupKey{
					Cert:            "cert1",
					ClusterGroup:    "non-existing-clusterGroup",
					ClusterProvider: "provider1"}
				client := module.NewClusterGroupClient(key)
				err := client.DeleteClusterGroup(ctx)
				validateError(err, "db Remove resource not found")
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
		Context("delete an existing clusterGroup, logicalCloud", func() {
			It("returns no error and delete the entry from the db", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				key := logicalcloud.ClusterGroupKey{
					Cert:               "cert1",
					CaCertLogicalCloud: "lc1",
					ClusterGroup:       "test-clusterGroup-4",
					Project:            "proj1"}
				client := module.NewClusterGroupClient(key)
				err := client.DeleteClusterGroup(ctx)
				validateError(err, "")
				Expect(len(mockdb.Items)).To(Equal(l - 1))
			})
		})
		Context("delete a nonexisting clusterGroup, logicalCloud", func() {
			It("returns an error and no change in the db", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mockdb.Err = errors.New("db Remove resource not found")
				key := logicalcloud.ClusterGroupKey{
					Cert:               "cert1",
					CaCertLogicalCloud: "lc1",
					ClusterGroup:       "non-existing-clusterGroup",
					Project:            "proj1"}
				client := module.NewClusterGroupClient(key)
				err := client.DeleteClusterGroup(ctx)
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
		Context("get all the clusterGroups, clusterProvider", func() {
			It("returns all the clusterGroups, no error", func() {
				ctx := context.Background()
				key := clusterprovider.ClusterGroupKey{
					Cert:            "cert1",
					ClusterProvider: "provider1"}
				client := module.NewClusterGroupClient(key)
				clusters, err := client.GetAllClusterGroups(ctx)
				validateError(err, "")
				Expect(len(clusters)).To(Equal(3))
			})
		})
		Context("get all the clusterGroups without creating any, clusterProvider", func() {
			It("returns an empty array, no error", func() {
				ctx := context.Background()
				mockdb.Items = []map[string]map[string][]byte{}
				key := clusterprovider.ClusterGroupKey{
					Cert:            "cert1",
					ClusterProvider: "provider1"}
				client := module.NewClusterGroupClient(key)
				clusters, err := client.GetAllClusterGroups(ctx)
				validateError(err, "")
				Expect(len(clusters)).To(Equal(0))
			})
		})

		Context("get all the clusterGroups, logicalCloud", func() {
			It("returns all the clusterGroups, no error", func() {
				ctx := context.Background()
				key := logicalcloud.ClusterGroupKey{
					Cert:               "cert1",
					CaCertLogicalCloud: "lc1",
					Project:            "proj1"}
				client := module.NewClusterGroupClient(key)
				clusters, err := client.GetAllClusterGroups(ctx)
				validateError(err, "")
				Expect(len(clusters)).To(Equal(3))
			})
		})
		Context("get all the clusterGroups without creating any, logicalCloud", func() {
			It("returns an empty array, no error", func() {
				ctx := context.Background()
				mockdb.Items = []map[string]map[string][]byte{}
				key := logicalcloud.ClusterGroupKey{
					Cert:               "cert1",
					CaCertLogicalCloud: "lc1",
					Project:            "proj1"}
				client := module.NewClusterGroupClient(key)
				clusters, err := client.GetAllClusterGroups(ctx)
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
		Context("get an existing clusterGroups", func() {
			It("returns the clusterGroups, no error", func() {
				ctx := context.Background()
				key := clusterprovider.ClusterGroupKey{
					Cert:            "cert1",
					ClusterGroup:    "test-clusterGroup-1",
					ClusterProvider: "provider1"}
				client := module.NewClusterGroupClient(key)
				cluster, err := client.GetClusterGroup(ctx)
				validateError(err, "")
				validateClusterGroup(cluster, mockClusterGroup("test-clusterGroup-1"))
			})
		})
		Context("get a nonexisting clusterGroups", func() {
			It("returns an error, no clusterGroups", func() {
				ctx := context.Background()
				key := clusterprovider.ClusterGroupKey{
					Cert:            "cert1",
					ClusterGroup:    "non-existing-clusterGroups",
					ClusterProvider: "provider1"}
				client := module.NewClusterGroupClient(key)
				cluster, err := client.GetClusterGroup(ctx)
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
			Description: "test clusterGroups",
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

	// clusterGroups 1
	cluster := mockClusterGroup("test-clusterGroup-1")
	cpKey := clusterprovider.ClusterGroupKey{
		Cert:            "cert1",
		ClusterGroup:    cluster.MetaData.Name,
		ClusterProvider: "provider1"}
	_ = mockdb.Insert(ctx, "resources", cpKey, nil, "data", cluster)

	// clusterGroups 2
	cluster = mockClusterGroup("test-clusterGroup-2")
	cpKey = clusterprovider.ClusterGroupKey{
		Cert:            "cert1",
		ClusterGroup:    cluster.MetaData.Name,
		ClusterProvider: "provider1"}
	_ = mockdb.Insert(ctx, "resources", cpKey, nil, "data", cluster)

	// clusterGroups 3
	cluster = mockClusterGroup("test-clusterGroup-3")
	cpKey = clusterprovider.ClusterGroupKey{
		Cert:            "cert1",
		ClusterGroup:    cluster.MetaData.Name,
		ClusterProvider: "provider1"}
	_ = mockdb.Insert(ctx, "resources", cpKey, nil, "data", cluster)

	// clusterGroups 4
	cluster = mockClusterGroup("test-clusterGroup-4")
	lcKey := logicalcloud.ClusterGroupKey{
		Cert:               "cert1",
		CaCertLogicalCloud: "lc1",
		ClusterGroup:       cluster.MetaData.Name,
		Project:            "proj1"}
	_ = mockdb.Insert(ctx, "resources", lcKey, nil, "data", cluster)

	// clusterGroups 5
	cluster = mockClusterGroup("test-clusterGroup-5")
	lcKey = logicalcloud.ClusterGroupKey{
		Cert:               "cert1",
		CaCertLogicalCloud: "lc1",
		ClusterGroup:       cluster.MetaData.Name,
		Project:            "proj1"}
	_ = mockdb.Insert(ctx, "resources", lcKey, nil, "data", cluster)

	// clusterGroups 6
	cluster = mockClusterGroup("test-clusterGroup-6")
	lcKey = logicalcloud.ClusterGroupKey{
		Cert:               "cert1",
		CaCertLogicalCloud: "lc1",
		ClusterGroup:       cluster.MetaData.Name,
		Project:            "proj1"}
	_ = mockdb.Insert(ctx, "resources", lcKey, nil, "data", cluster)

}
