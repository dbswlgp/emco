package module_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/ovnaction/pkg/module"
)

var _ = Describe("Workloadifintent", func() {
	var (
		NCI    module.NetControlIntent
		NCIDBC *module.NetControlIntentClient

		WLI    module.WorkloadIntent
		WLIDBC *module.WorkloadIntentClient

		WLFI    module.WorkloadIfIntent
		WLFIDBC *module.WorkloadIfIntentClient

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

		WLFIDBC = module.NewWorkloadIfIntentClient()
		WLFI = module.WorkloadIfIntent{
			Metadata: module.Metadata{
				Name:        "theThirdName",
				Description: "work load if intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}

		mdb = new(db.MockDB)
		mdb.Err = nil
		db.DBconn = mdb
	})

	Describe("Create client intent", func() {
		It("with pre created net control intent should return nil", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(ctx, WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			_, err = (*WLFIDBC).CreateWorkloadIfIntent(ctx, WLFI, "test", "capp1", "v1", "dig", "theName", "theSecondName", false)
			Expect(err).To(BeNil())
		})
		It("followed by create again should return error", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(ctx, WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			_, err = (*WLFIDBC).CreateWorkloadIfIntent(ctx, WLFI, "test", "capp1", "v1", "dig", "theName", "theSecondName", false)
			Expect(err).To(BeNil())
			_, err = (*WLFIDBC).CreateWorkloadIfIntent(ctx, WLFI, "test", "capp1", "v1", "dig", "theName", "theSecondName", false)
			Expect(err).To(HaveOccurred())
		})
		It("followed by get should return nil", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(ctx, WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			_, err = (*WLFIDBC).CreateWorkloadIfIntent(ctx, WLFI, "test", "capp1", "v1", "dig", "theName", "theSecondName", false)
			Expect(err).To(BeNil())
			wlfi, err := (*WLFIDBC).GetWorkloadIfIntent(ctx, "theThirdName", "test", "capp1", "v1", "dig", "theName", "theSecondName")
			Expect(wlfi).Should(Equal(WLFI))
		})
		It("followed by delete should return nil", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(ctx, WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			_, err = (*WLFIDBC).CreateWorkloadIfIntent(ctx, WLFI, "test", "capp1", "v1", "dig", "theName", "theSecondName", false)
			Expect(err).To(BeNil())
			err = (*WLFIDBC).DeleteWorkloadIfIntent(ctx, "theThirdName", "test", "capp1", "v1", "dig", "theName", "theSecondName")
			Expect(err).To(BeNil())
		})
	})

	Describe("Create workload if intent", func() {
		It("followed by create,get,delete,get workload if intent should return an error", func() {
			ctx := context.Background()
			_, err := (*NCIDBC).CreateNetControlIntent(ctx, NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(ctx, WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			_, err = (*WLFIDBC).CreateWorkloadIfIntent(ctx, WLFI, "test", "capp1", "v1", "dig", "theName", "theSecondName", false)
			Expect(err).To(BeNil())
			wlfi, err := (*WLFIDBC).GetWorkloadIfIntent(ctx, "theThirdName", "test", "capp1", "v1", "dig", "theName", "theSecondName")
			Expect(wlfi).Should(Equal(WLFI))
			err = (*WLFIDBC).DeleteWorkloadIfIntent(ctx, "theThirdName", "test", "capp1", "v1", "dig", "theName", "theSecondName")
			Expect(err).To(BeNil())
			wlfi, err = (*WLFIDBC).GetWorkloadIfIntent(ctx, "theThirdName", "test", "capp1", "v1", "dig", "theName", "theSecondName")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Get workload if intents", func() {
		It("should return error for non-existing record", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("WorkloadIfIntent not found")
			_, err := (*WLFIDBC).GetWorkloadIfIntents(ctx, "test", "capp1", "v1", "dig", "theName", "theSecondName")
			Expect(err).To(HaveOccurred())
		})
	})

})
