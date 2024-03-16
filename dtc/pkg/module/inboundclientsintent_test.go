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

var _ = Describe("Inboundclientsintent", func() {

	var (
		TGI    module.TrafficGroupIntent
		TGIDBC *module.TrafficGroupIntentDbClient

		ISI    module.InboundServerIntent
		ISIDBC *module.InboundServerIntentDbClient

		ICI    module.InboundClientsIntent
		ICIDBC *module.InboundClientsIntentDbClient

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

		ICIDBC = module.NewClientsInboundIntentClient()
		ICI = module.InboundClientsIntent{
			Metadata: module.Metadata{
				Name:        "testici",
				Description: "inbound client intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}
		mdb = new(db.MockDB)
		mdb.Err = nil
		db.DBconn = mdb
	})

	Describe("Create client intent", func() {
		It("with pre created traffic and server intent should return nil", func() {
			ctx := context.Background()
			_, err := (*TGIDBC).CreateTrafficGroupIntent(ctx, TGI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*ISIDBC).CreateServerInboundIntent(ctx, ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(BeNil())
			_, err = (*ICIDBC).CreateClientsInboundIntent(ctx, ICI, "test", "capp1", "v1", "dig", "testtgi", "testisi", false)
			Expect(err).To(BeNil())
		})
		/* The DTC code no longer checks for parent resource so test is not valid
		It("should return error", func() {
			_, err := (*ICIDBC).CreateClientsInboundIntent(ICI, "test", "capp1", "v1", "dig", "test tgi", "test ici", false)
			Expect(err).To(HaveOccurred())
		})
		*/

		It("create again should return error", func() {
			ctx := context.Background()
			_, err := (*TGIDBC).CreateTrafficGroupIntent(ctx, TGI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*ISIDBC).CreateServerInboundIntent(ctx, ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			_, err = (*ICIDBC).CreateClientsInboundIntent(ctx, ICI, "test", "capp1", "v1", "dig", "testtgi", "testisi", false)
			Expect(err).To(BeNil())
			_, err = (*ICIDBC).CreateClientsInboundIntent(ctx, ICI, "test", "capp1", "v1", "dig", "testtgi", "testisi", false)
			Expect(err).To(HaveOccurred())
		})
		It("followed by get clients intent should return nil", func() {
			ctx := context.Background()
			_, err := (*TGIDBC).CreateTrafficGroupIntent(ctx, TGI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*ISIDBC).CreateServerInboundIntent(ctx, ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(BeNil())
			_, err = (*ICIDBC).CreateClientsInboundIntent(ctx, ICI, "test", "capp1", "v1", "dig", "testtgi", "testisi", false)
			Expect(err).To(BeNil())
			ici, err := (*ICIDBC).GetClientsInboundIntent(ctx, "testici", "test", "capp1", "v1", "dig", "testtgi", "testisi")
			Expect(err).To(BeNil())
			Expect(ici).Should(Equal(ICI))
		})
		It("followed by delete clients intent should return nil", func() {
			ctx := context.Background()
			_, err := (*TGIDBC).CreateTrafficGroupIntent(ctx, TGI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*ISIDBC).CreateServerInboundIntent(ctx, ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(BeNil())
			_, err = (*ICIDBC).CreateClientsInboundIntent(ctx, ICI, "test", "capp1", "v1", "dig", "testtgi", "testisi", false)
			Expect(err).To(BeNil())
			err = (*ICIDBC).DeleteClientsInboundIntent(ctx, "testici", "test", "capp1", "v1", "dig", "testtgi", "testisi")
			Expect(err).To(BeNil())
		})

	})

	Describe("Get client intent", func() {
		It("should return error for non-existing record", func() {
			ctx := context.Background()
			_, err := (*ICIDBC).GetClientsInboundIntent(ctx, "testici", "test", "capp1", "v1", "dig", "testtgi", "testisi")
			Expect(err).To(HaveOccurred())
		})

	})
	Describe("Get clients intents", func() {
		It("should return error for non-existing record", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("Inbound clients intent not found")
			_, err := (*ICIDBC).GetClientsInboundIntents(ctx, "test", "capp1", "v1", "dig", "testtgi", "testisi")
			Expect(err).To(HaveOccurred())
		})

	})
	Describe("Delete client intent", func() {
		It("should return error for non-existing record", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("db Remove resource not found")
			err := (*ICIDBC).DeleteClientsInboundIntent(ctx, "testici", "test", "capp1", "v1", "dig", "testtgi", "testisi")
			Expect(err).To(HaveOccurred())
		})
		It("should return error for deleting parent without deleting child", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("Cannot delete parent without deleting child references first")
			err := (*ICIDBC).DeleteClientsInboundIntent(ctx, "testici", "test", "capp1", "v1", "dig", "testtgi", "testisi")
			Expect(err).To(HaveOccurred())
		})
		It("should return error for general db error", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("db Remove error")
			err := (*ICIDBC).DeleteClientsInboundIntent(ctx, "testici", "test", "capp1", "v1", "dig", "testtgi", "testisi")
			Expect(err).To(HaveOccurred())
		})

	})

})
