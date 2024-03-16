// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"

	"context"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

var _ = Describe("Create Cert",
	func() {
		BeforeEach(func() {
			populateCertTestData()
		})
		Context("create a caCert for a clusterProvider", func() {
			It("returns the caCert, no error and, the exists flag is false", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mCert := mockCert("new-cert-1")
				key := clusterprovider.CaCertKey{
					Cert:            mCert.MetaData.Name,
					ClusterProvider: "provider1"}
				client := module.NewCaCertClient(key)
				c, cExists, err := client.CreateCert(ctx, mCert, true)
				validateError(err, "")
				Expect(c).To(Equal(mCert))
				Expect(cExists).To(Equal(false))
				Expect(len(mockdb.Items)).To(Equal(l + 1))
			})
		})
		Context("create a caCert for a clusterProvider that already exists", func() {
			It("returns an error, no caCert and, the exists flag is true", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mCert := mockCert("test-cert-1")
				key := clusterprovider.CaCertKey{
					Cert:            mCert.MetaData.Name,
					ClusterProvider: "provider1"}
				client := module.NewCaCertClient(key)
				c, cExists, err := client.CreateCert(ctx, mCert, true)
				validateError(err, module.CaCertAlreadyExists)
				Expect(cExists).To(Equal(true))
				Expect(c).To(Equal(module.CaCert{}))
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
		Context("create a caCert for a logicalCloud", func() {
			It("returns the caCert, no error and, the exists flag is false", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mCert := mockCert("new-cert-1")
				key := logicalcloud.CaCertKey{
					Cert:    mCert.MetaData.Name,
					Project: "proj1"}
				client := module.NewCaCertClient(key)
				c, cExists, err := client.CreateCert(ctx, mCert, true)
				validateError(err, "")
				Expect(cExists).To(Equal(false))
				Expect(c).To(Equal(mCert))
				Expect(len(mockdb.Items)).To(Equal(l + 1))
			})
		})
		Context("create a caCert for a logicalCloud that already exists", func() {
			It("returns an error, no caCert and, the exists flag is true", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mCert := mockCert("test-cert-4")
				key := logicalcloud.CaCertKey{
					Cert:    mCert.MetaData.Name,
					Project: "proj1"}
				client := module.NewCaCertClient(key)
				c, cExists, err := client.CreateCert(ctx, mCert, true)
				validateError(err, module.CaCertAlreadyExists)
				Expect(c).To(Equal(module.CaCert{}))
				Expect(cExists).To(Equal(true))
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
	},
)

var _ = Describe("Delete Cert",
	func() {
		BeforeEach(func() {
			populateCertTestData()
		})
		Context("delete an existing caCert, clusterProvider", func() {
			It("returns no error and delete the entry from the db", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				key := clusterprovider.CaCertKey{
					Cert:            "test-cert-1",
					ClusterProvider: "provider1"}
				client := module.NewCaCertClient(key)
				err := client.DeleteCert(ctx)
				validateError(err, "")
				Expect(len(mockdb.Items)).To(Equal(l - 1))
			})
		})
		Context("delete a nonexisting caCert, clusterProvider", func() {
			It("returns an error and no change in the db", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mockdb.Err = errors.New("db Remove resource not found")
				key := clusterprovider.CaCertKey{
					Cert:            "non-existing-cert",
					ClusterProvider: "provider1"}
				client := module.NewCaCertClient(key)
				err := client.DeleteCert(ctx)
				validateError(err, "db Remove resource not found")
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
		Context("delete an existing caCert, logicalCloud", func() {
			It("returns no error and delete the entry from the db", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				key := logicalcloud.CaCertKey{
					Cert:    "test-cert-4",
					Project: "proj1"}
				client := module.NewCaCertClient(key)
				err := client.DeleteCert(ctx)
				validateError(err, "")
				Expect(len(mockdb.Items)).To(Equal(l - 1))
			})
		})
		Context("delete a nonexisting caCert, logicalCloud", func() {
			It("returns an error and no change in the db", func() {
				ctx := context.Background()
				l := len(mockdb.Items)
				mockdb.Err = errors.New("db Remove resource not found")
				key := logicalcloud.CaCertKey{
					Cert:    "non-existing-cert",
					Project: "proj1"}
				client := module.NewCaCertClient(key)
				err := client.DeleteCert(ctx)
				validateError(err, "db Remove resource not found")
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
	},
)

var _ = Describe("Get All CaCerts",
	func() {
		BeforeEach(func() {
			populateCertTestData()
		})
		Context("get all the caCerts, clusterProvider", func() {
			It("returns all the caCerts, no error", func() {
				ctx := context.Background()
				key := clusterprovider.CaCertKey{
					ClusterProvider: "provider1"}
				client := module.NewCaCertClient(key)
				certs, err := client.GetAllCert(ctx)
				validateError(err, "")
				Expect(len(certs)).To(Equal(3))
			})
		})
		Context("get all the caCerts without creating any, clusterProvider", func() {
			It("returns an empty array, no error", func() {
				ctx := context.Background()
				mockdb.Items = []map[string]map[string][]byte{}
				key := clusterprovider.CaCertKey{
					ClusterProvider: "provider1"}
				client := module.NewCaCertClient(key)
				certs, err := client.GetAllCert(ctx)
				validateError(err, "")
				Expect(len(certs)).To(Equal(0))
			})
		})

		Context("get all the caCerts, logicalCloud", func() {
			It("returns all the caCerts, no error", func() {
				ctx := context.Background()
				key := logicalcloud.CaCertKey{
					Project: "proj1"}
				client := module.NewCaCertClient(key)
				certs, err := client.GetAllCert(ctx)
				validateError(err, "")
				Expect(len(certs)).To(Equal(3))
			})
		})
		Context("get all the caCerts without creating any, logicalCloud", func() {
			It("returns an empty array, no error", func() {
				ctx := context.Background()
				mockdb.Items = []map[string]map[string][]byte{}
				key := logicalcloud.CaCertKey{
					Project: "proj1"}
				client := module.NewCaCertClient(key)
				certs, err := client.GetAllCert(ctx)
				validateError(err, "")
				Expect(len(certs)).To(Equal(0))
			})
		})
	},
)

var _ = Describe("Get Cert",
	func() {
		BeforeEach(func() {
			populateCertTestData()
		})
		Context("get an existing caCert, clusterProvider", func() {
			It("returns the caCert, no error", func() {
				ctx := context.Background()
				key := clusterprovider.CaCertKey{
					Cert:            "test-cert-1",
					ClusterProvider: "provider1"}
				client := module.NewCaCertClient(key)
				cert, err := client.GetCert(ctx)
				validateError(err, "")
				validateCert(cert, mockCert("test-cert-1"))
			})
		})
		Context("get a nonexisting caCert, clusterProvider", func() {
			It("returns an error, no caCert", func() {
				ctx := context.Background()
				key := clusterprovider.CaCertKey{
					Cert:            "non-existing-cert",
					ClusterProvider: "provider1"}
				client := module.NewCaCertClient(key)
				cert, err := client.GetCert(ctx)
				validateError(err, module.CaCertNotFound)
				validateCert(cert, module.CaCert{})
			})
		})
	},
)

// validateCert
func validateCert(in, out module.CaCert) {
	Expect(in).To(Equal(out))
}

// mockCert
func mockCert(name string) module.CaCert {
	return module.CaCert{
		MetaData: types.Metadata{
			Name:        name,
			Description: "test cert",
			UserData1:   "some user data 1",
			UserData2:   "some user data 2",
		},
	}
}

// populateCertTestData
func populateCertTestData() {
	ctx := context.Background()
	mockdb.Err = nil
	mockdb.Items = []map[string]map[string][]byte{}
	mockdb.MarshalErr = nil

	// cert 1
	cert := mockCert("test-cert-1")
	cpKey := clusterprovider.CaCertKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: "provider1"}
	_ = mockdb.Insert(ctx, "resources", cpKey, nil, "data", cert)

	// cert 2
	cert = mockCert("test-cert-2")
	cpKey = clusterprovider.CaCertKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: "provider1"}
	_ = mockdb.Insert(ctx, "resources", cpKey, nil, "data", cert)

	// cert 3
	cert = mockCert("test-cert-3")
	cpKey = clusterprovider.CaCertKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: "provider1"}
	_ = mockdb.Insert(ctx, "resources", cpKey, nil, "data", cert)

	// cert 4
	cert = mockCert("test-cert-4")
	lcKey := logicalcloud.CaCertKey{
		Cert:    cert.MetaData.Name,
		Project: "proj1"}
	_ = mockdb.Insert(ctx, "resources", lcKey, nil, "data", cert)

	// cert 5
	cert = mockCert("test-cert-5")
	lcKey = logicalcloud.CaCertKey{
		Cert:    cert.MetaData.Name,
		Project: "proj1"}
	_ = mockdb.Insert(ctx, "resources", lcKey, nil, "data", cert)

	// cert 6
	cert = mockCert("test-cert-6")
	lcKey = logicalcloud.CaCertKey{
		Cert:    cert.MetaData.Name,
		Project: "proj1"}
	_ = mockdb.Insert(ctx, "resources", lcKey, nil, "data", cert)

}

func validateError(err error, message string) {
	if len(message) == 0 {
		Expect(err).NotTo(HaveOccurred())
		Expect(err).To(BeNil())
		return
	}
	Expect(err.Error()).To(ContainSubstring(message))
}
