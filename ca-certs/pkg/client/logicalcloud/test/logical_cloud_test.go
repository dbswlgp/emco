// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"context"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

var (
	lcClient = logicalcloud.NewCaCertLogicalCloudClient()
)

var _ = Describe("Create CaCertLogicalCloud",
	func() {
		BeforeEach(func() {
			populateLogicalCloudTestData()
		})
		Context("create a caCertLogicalCloud for a logicalCloud", func() {
			It("returns the caCertLogicalCloud, no error and, the exists flag is false", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mLogicalCloud := mockLogicalCloud("new-caCertLogicalCloud-1")
				c, cExists, err := lcClient.CreateLogicalCloud(ctx, mLogicalCloud, "cert1", "proj1", true)
				validateError(err, "")
				validateLogicalCloud(c, mLogicalCloud)
				Expect(cExists).To(Equal(false))
				Expect(len(mockdb.Items)).To(Equal(l + 1))
			})
		})
		Context("create a caCertLogicalCloud for a logicalCloud that already exists", func() {
			It("returns an error, no caCertLogicalCloud and, the exists flag is true", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mLogicalCloud := mockLogicalCloud("test-caCertLogicalCloud-1")
				c, cExists, err := lcClient.CreateLogicalCloud(ctx, mLogicalCloud, "cert1", "proj1", true)
				validateError(err, module.CaCertLogicalCloudAlreadyExists)
				validateLogicalCloud(c, logicalcloud.CaCertLogicalCloud{})
				Expect(cExists).To(Equal(true))
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
	},
)

var _ = Describe("Delete CaCertLogicalCloud",
	func() {
		BeforeEach(func() {
			populateLogicalCloudTestData()
		})
		Context("delete an existing caCertLogicalCloud", func() {
			It("returns no error and delete the entry from the db", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				err := lcClient.DeleteLogicalCloud(ctx, "test-caCertLogicalCloud-1", "cert1", "proj1")
				validateError(err, "")
				Expect(len(mockdb.Items)).To(Equal(l - 1))
			})
		})
		Context("delete a nonexisting caCertLogicalCloud", func() {
			It("returns an error and no change in the db", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mockdb.Err = errors.New("db Remove resource not found")
				err := lcClient.DeleteLogicalCloud(ctx, "non-existing-caCertLogicalCloud", "cert1", "proj1")
				validateError(err, "db Remove resource not found")
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
	},
)

var _ = Describe("Get All CaCertLogicalCloud",
	func() {
		BeforeEach(func() {
			populateLogicalCloudTestData()
		})
		Context("get all the caCertLogicalClouds", func() {
			It("returns all the caCertLogicalClouds, no error", func() {
				ctx := context.Background()
				clusters, err := lcClient.GetAllLogicalClouds(ctx, "cert1", "proj1")
				validateError(err, "")
				Expect(len(clusters)).To(Equal(len(mockdb.Items)))
			})
		})
		Context("get all the caCertLogicalClouds without creating any", func() {
			It("returns an empty array, no error", func() {
				ctx := context.Background()
				mockdb.Items = []map[string]map[string][]byte{}
				clusters, err := lcClient.GetAllLogicalClouds(ctx, "cert1", "proj1")
				validateError(err, "")
				Expect(len(clusters)).To(Equal(0))
			})
		})
	},
)

var _ = Describe("Get CaCertLogicalCloud",
	func() {
		BeforeEach(func() {
			populateLogicalCloudTestData()
		})
		Context("get an existing caCertLogicalCloud", func() {
			It("returns the caCertLogicalCloud, no error", func() {
				ctx := context.Background()
				cluster, err := lcClient.GetLogicalCloud(ctx, "test-caCertLogicalCloud-1", "cert1", "proj1")
				validateError(err, "")
				validateLogicalCloud(cluster, mockLogicalCloud("test-caCertLogicalCloud-1"))
			})
		})
		Context("get a nonexisting caCertLogicalCloud", func() {
			It("returns an error, no caCertLogicalCloud", func() {
				ctx := context.Background()
				cluster, err := lcClient.GetLogicalCloud(ctx, "non-existing-caCertLogicalCloud", "cert1", "proj1")
				validateError(err, module.CaCertLogicalCloudNotFound)
				validateLogicalCloud(cluster, logicalcloud.CaCertLogicalCloud{})
			})
		})
	},
)

// validateLogicalCloud
func validateLogicalCloud(in, out logicalcloud.CaCertLogicalCloud) {
	Expect(in).To(Equal(out))
}

// mockLogicalCloud
func mockLogicalCloud(name string) logicalcloud.CaCertLogicalCloud {
	return logicalcloud.CaCertLogicalCloud{
		MetaData: types.Metadata{
			Name:        name,
			Description: "test cluster",
			UserData1:   "some user data 1",
			UserData2:   "some user data 2",
		},
		Spec: logicalcloud.CaCertLogicalCloudSpec{
			LogicalCloud: "lc1",
		},
	}
}

// populateLogicalCloudTestData
func populateLogicalCloudTestData() {
	ctx := context.Background()
	mockdb.Err = nil
	mockdb.Items = []map[string]map[string][]byte{}
	mockdb.MarshalErr = nil

	// caCertLogicalCloud 1
	lc := mockLogicalCloud("test-caCertLogicalCloud-1")
	cpKey := logicalcloud.CaCertLogicalCloudKey{
		Cert:               "cert1",
		CaCertLogicalCloud: lc.MetaData.Name,
		Project:            "proj1"}
	_ = mockdb.Insert(ctx, "resources", cpKey, nil, "data", lc)

	// caCertLogicalCloud 2
	lc = mockLogicalCloud("test-caCertLogicalCloud-2")
	cpKey = logicalcloud.CaCertLogicalCloudKey{
		Cert:               "cert1",
		CaCertLogicalCloud: lc.MetaData.Name,
		Project:            "proj1"}
	_ = mockdb.Insert(ctx, "resources", cpKey, nil, "data", lc)

	// caCertLogicalCloud 3
	lc = mockLogicalCloud("test-caCertLogicalCloud-3")
	cpKey = logicalcloud.CaCertLogicalCloudKey{
		Cert:               "cert1",
		CaCertLogicalCloud: lc.MetaData.Name,
		Project:            "proj1"}
	_ = mockdb.Insert(ctx, "resources", cpKey, nil, "data", lc)
}

func validateError(err error, message string) {
	if len(message) == 0 {
		Expect(err).NotTo(HaveOccurred())
		Expect(err).To(BeNil())
		return
	}
	Expect(err.Error()).To(ContainSubstring(message))
}
