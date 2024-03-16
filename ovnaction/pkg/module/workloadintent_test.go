package module_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/ovnaction/pkg/module"
)

var _ = Describe("Workloadintent", func() {
	var (
		NCI    module.NetControlIntent
		NCIDBC *module.NetControlIntentClient

		WLI    module.WorkloadIntent
		WLIDBC *module.WorkloadIntentClient

		mdb *db.MockDB
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

		WLIDBC = module.NewWorkloadIntentClient()
		WLI = module.WorkloadIntent{
			Metadata: module.Metadata{
				Name:        "theSecondName",
				Description: "work load intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}

		mdb = new(db.MockDB)
		mdb.Err = nil
		db.DBconn = mdb
	})

	Describe("Create client intent", func() {
		It("with pre created netcontrolintent should return nil", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(ctx, WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
		})
		It("followed by create again should return error", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(ctx, WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			mdb.Err = pkgerrors.New("Already exists:")
			_, err = (*WLIDBC).CreateWorkloadIntent(ctx, WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(HaveOccurred())
		})
		It("followed by get should return nil", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(ctx, WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			wli, err := (*WLIDBC).GetWorkloadIntent(ctx, "theSecondName", "test", "capp1", "v1", "dig", "theName")
			Expect(wli).Should(Equal(WLI))
		})
		It("followed by delete should return nil", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(ctx, WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			err = (*WLIDBC).DeleteWorkloadIntent(ctx, "theSecondName", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(BeNil())
		})
	})

	Describe("Create workload intent", func() {
		It("followed by create,get,delete,get should return an error", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(ctx, WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			wli, err := (*WLIDBC).GetWorkloadIntent(ctx, "theSecondName", "test", "capp1", "v1", "dig", "theName")
			Expect(wli).Should(Equal(WLI))
			err = (*WLIDBC).DeleteWorkloadIntent(ctx, "theSecondName", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(BeNil())
			wli, err = (*WLIDBC).GetWorkloadIntent(ctx, "theSecondName", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Get workload intents", func() {
		It("should return error for non-existing record", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("WorkloadIntent not found")
			_, err := (*WLIDBC).GetWorkloadIntents(ctx, "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Delete workload intent", func() {
		It("should return error for non-existing record", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("db Remove resource not found")
			err := (*WLIDBC).DeleteWorkloadIntent(ctx, "testtgi", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(HaveOccurred())
		})
		It("should return error for deleting parent without deleting child", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("Cannot delete parent without deleting child references first")
			err := (*WLIDBC).DeleteWorkloadIntent(ctx, "testtgi", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(HaveOccurred())
		})
		It("should return error for general db error", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("db Remove error")
			err := (*WLIDBC).DeleteWorkloadIntent(ctx, "testtgi", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(HaveOccurred())
		})
	})
})
