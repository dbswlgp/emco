// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation
package module_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/dtc/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

var _ = Describe("Inboundserverintent", func() {

	var (
		TGI    module.TrafficGroupIntent
		TGIDBC *module.TrafficGroupIntentDbClient

		ISI    module.InboundServerIntent
		ISIDBC *module.InboundServerIntentDbClient

		mdb *db.MockDB
	)

	BeforeEach(func() {

		TGIDBC = module.NewTrafficGroupIntentClient()
		TGI = module.TrafficGroupIntent{
			Metadata: module.Metadata{
				Name:        "testtgi",
				Description: "traffic group intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}

		ISIDBC = module.NewServerInboundIntentClient()
		ISI = module.InboundServerIntent{
			Metadata: module.Metadata{
				Name:        "testisi",
				Description: "inbound server intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}
		mdb = new(db.MockDB)
		mdb.Err = nil
		db.DBconn = mdb

	})

	Describe("Create server intent", func() {
		It("with pre created traffic intent should return nil", func() {
			ctx := context.Background()
			_, err := (*TGIDBC).CreateTrafficGroupIntent(ctx, TGI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*ISIDBC).CreateServerInboundIntent(ctx, ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(BeNil())
		})
		/* DTC code does not check for the parent resource - so this test is no longer valid
		It("should return error", func() {
			_, err := (*ISIDBC).CreateServerInboundIntent(ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(HaveOccurred())
		})
		*/
		It("create again should return error", func() {
			ctx := context.Background()
			_, err := (*TGIDBC).CreateTrafficGroupIntent(ctx, TGI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*ISIDBC).CreateServerInboundIntent(ctx, ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(BeNil())
			_, err = (*ISIDBC).CreateServerInboundIntent(ctx, ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(HaveOccurred())
		})
		It("followed by get server intent should return nil", func() {
			ctx := context.Background()
			_, err := (*TGIDBC).CreateTrafficGroupIntent(ctx, TGI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*ISIDBC).CreateServerInboundIntent(ctx, ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(BeNil())
			isi, err := (*ISIDBC).GetServerInboundIntent(ctx, "testisi", "test", "capp1", "v1", "dig", "testtgi")
			Expect(err).To(BeNil())
			Expect(isi).Should(Equal(ISI))
		})
		It("followed by delete server intent should return nil", func() {
			ctx := context.Background()
			_, err := (*TGIDBC).CreateTrafficGroupIntent(ctx, TGI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*ISIDBC).CreateServerInboundIntent(ctx, ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(BeNil())
			err = (*ISIDBC).DeleteServerInboundIntent(ctx, "testisi", "test", "capp1", "v1", "dig", "testtgi")
			Expect(err).To(BeNil())
		})

	})

	Describe("Get server intent", func() {
		It("should return error for non-existing record", func() {
			ctx := context.Background()
			_, err := (*ISIDBC).GetServerInboundIntent(ctx, "testisi", "test", "capp1", "v1", "dig", "testtgi")
			Expect(err).To(HaveOccurred())
		})

	})
	Describe("Get server intents", func() {
		It("should return error for non-existing record", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("Inbound server intent not found")
			_, err := (*ISIDBC).GetServerInboundIntents(ctx, "test", "capp1", "v1", "dig", "testtgi")
			Expect(err).To(HaveOccurred())
		})

	})
	Describe("Delete server intent", func() {
		It("should return error for non-existing record", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("db Remove resource not found")
			err := (*ISIDBC).DeleteServerInboundIntent(ctx, "testisi", "test", "capp1", "v1", "dig", "testtgi")
			Expect(err).To(HaveOccurred())
		})
		It("should return error for deleting parent without deleting child", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("Cannot delete parent without deleting child references first")
			err := (*ISIDBC).DeleteServerInboundIntent(ctx, "testisi", "test", "capp1", "v1", "dig", "testtgi")
			Expect(err).To(HaveOccurred())
		})
		It("should return error for general db error", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("db Remove error")
			err := (*ISIDBC).DeleteServerInboundIntent(ctx, "testisi", "test", "capp1", "v1", "dig", "testtgi")
			Expect(err).To(HaveOccurred())
		})

	})
})
