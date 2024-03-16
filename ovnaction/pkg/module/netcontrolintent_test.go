package module_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/ovnaction/pkg/module"
)

var _ = Describe("Netcontrolintent", func() {
	var (
		NCI       module.NetControlIntent
		OTHER_NCI module.NetControlIntent
		NCIDBC    *module.NetControlIntentClient
		mdb       *db.NewMockDB
	)

	BeforeEach(func() {
		NCIDBC = module.NewNetControlIntentClient()
		NCI = module.NetControlIntent{
			Metadata: module.Metadata{
				Name:        "theName",
				Description: "net control intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}
		OTHER_NCI = module.NetControlIntent{
			Metadata: module.Metadata{
				Name:        "Name",
				Description: "net control intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}
		mdb = new(db.NewMockDB)
		mdb.Err = nil
		db.DBconn = mdb
	})

	Describe("Create net intent", func() {
		It("should return nil", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
		})
		It("followed by create again should return error", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(HaveOccurred())
		})
		It("followed by get should return nil", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			nci, err := (*NCIDBC).GetNetControlIntent(ctx, "theName", "test", "capp1", "v1", "dig")
			Expect(nci).Should(Equal(NCI))
		})
		It("followed by delete should return nil", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			err = (*NCIDBC).DeleteNetControlIntent(ctx, "testnci", "test", "capp1", "v1", "dig")
			Expect(err).To(BeNil())
		})
	})

	Describe("Create net intent", func() {
		It("followed by create,get,delete,get net intent should return an error", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*NCIDBC).GetNetControlIntent(ctx, "theName", "test", "capp1", "v1", "dig")
			Expect(err).To(BeNil())
			err = (*NCIDBC).DeleteNetControlIntent(ctx, "theName", "test", "capp1", "v1", "dig")
			Expect(err).To(BeNil())
			_, err = (*NCIDBC).GetNetControlIntent(ctx, "theName", "test", "capp1", "v1", "dig")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Get net control intents", func() {
		It("should return error for non-existing record", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("Net Control Intent not found")
			_, err := (*NCIDBC).GetNetControlIntents(ctx, "test", "capp1", "v1", "dig")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Delete net control intent", func() {
		It("should return error for non-existing record", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("db Remove resource not found")
			err := (*NCIDBC).DeleteNetControlIntent(ctx, "testtgi", "test", "capp1", "v1", "dig")
			Expect(err).To(HaveOccurred())
		})
		It("should return error for deleting parent without deleting child", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("Cannot delete parent without deleting child references first")
			err := (*NCIDBC).DeleteNetControlIntent(ctx, "testtgi", "test", "capp1", "v1", "dig")
			Expect(err).To(HaveOccurred())
		})
		It("should return error for general db error", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("db Remove error")
			err := (*NCIDBC).DeleteNetControlIntent(ctx, "testtgi", "test", "capp1", "v1", "dig")
			Expect(err).To(HaveOccurred())
		})
	})
	Describe("Create 2 net control intents", func() {
		It("should get all the net control intents for the project", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*NCIDBC).CreateNetControlIntent(ctx, OTHER_NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			rval, err := (*NCIDBC).GetNetControlIntents(ctx, "test", "capp1", "v1", "dig")
			Expect(err).To(BeNil())
			Expect(len(rval)).To(Equal(2))
		})
	})
})
