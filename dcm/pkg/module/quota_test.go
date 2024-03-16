package module_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"
	dcm "gitlab.com/project-emco/core/emco-base/src/dcm/pkg/module"
	common "gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	orch "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	types "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

var _ = Describe("Quota", func() {

	var (
		mdb    *db.NewMockDB
		client *dcm.QuotaClient
	)

	BeforeEach(func() {
		client = dcm.NewQuotaClient()
		mdb = new(db.NewMockDB)
		mdb.Err = nil
		mdb.Items = []map[string]map[string][]byte{}
		db.DBconn = mdb
	})
	Describe("Quota operations", func() {
		Context("from an empty database", func() {
			BeforeEach(func() {
				// create project in mocked db
				okey := orch.ProjectKey{
					ProjectName: "project",
				}
				p := orch.Project{}
				p.MetaData = orch.ProjectMetaData{
					Name:        "project",
					Description: "",
				}
				mdb.Insert(context.Background(), "resources", okey, nil, "data", p)
				// create logical cloud in mocked db
				lkey := common.LogicalCloudKey{
					Project:          "project",
					LogicalCloudName: "logicalcloud",
				}
				lc := common.LogicalCloud{}
				lc.MetaData = types.Metadata{
					Name:        "logicalcloud",
					Description: "",
				}
				lc.Specification = common.Spec{
					NameSpace: "anything",
					Level:     "1",
				}
				mdb.Insert(context.Background(), "resources", lkey, nil, "data", lc)
			})
			It("creation should succeed and return the resource created", func() {
				ctx := context.Background()
				quota := _createTestQuota("testquota")
				quota, err := client.CreateQuota(ctx, "project", "logicalcloud", quota)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(quota.MetaData.QuotaName).To(Equal("testquota"))
				Expect(quota.MetaData.Description).To(Equal(""))
			})
			It("get should fail and not return anything", func() {
				ctx := context.Background()
				quota, err := client.GetQuota(ctx, "project", "logicalcloud", "testquota")
				Expect(err).Should(HaveOccurred())
				Expect(quota).To(Equal(dcm.Quota{}))
			})
			It("create followed by get should return what was created", func() {
				ctx := context.Background()
				quota := _createTestQuota("testquota")
				_, _ = client.CreateQuota(ctx, "project", "logicalcloud", quota)
				quota, err := client.GetQuota(ctx, "project", "logicalcloud", "testquota")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(quota).To(Equal(quota))
			})
			It("create followed by get-all should return only what was created", func() {
				ctx := context.Background()
				quota := _createTestQuota("testquota")
				_, _ = client.CreateQuota(ctx, "project", "logicalcloud", quota)
				quotas, err := client.GetAllQuotas(ctx, "project", "logicalcloud")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(quotas)).To(Equal(1))
				Expect(quotas[0]).To(Equal(quota))
			})
			It("three creates followed by get-all should return all that was created", func() {
				ctx := context.Background()
				quota1 := _createTestQuota("testquota1")
				quota2 := _createTestQuota("testquota2")
				quota3 := _createTestQuota("testquota3")
				_, _ = client.CreateQuota(ctx, "project", "logicalcloud", quota1)
				_, _ = client.CreateQuota(ctx, "project", "logicalcloud", quota2)
				_, _ = client.CreateQuota(ctx, "project", "logicalcloud", quota3)
				quotas, err := client.GetAllQuotas(ctx, "project", "logicalcloud")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(quotas)).To(Equal(3))
				Expect(quotas[0]).To(Equal(quota1))
				Expect(quotas[1]).To(Equal(quota2))
				Expect(quotas[2]).To(Equal(quota3))
			})
			It("delete after creation should succeed and database remain empty", func() {
				ctx := context.Background()
				quota := _createTestQuota("testquota")
				_, _ = client.CreateQuota(ctx, "project", "logicalcloud", quota)
				err := client.DeleteQuota(ctx, "project", "logicalcloud", "testquota")
				Expect(err).ShouldNot(HaveOccurred())
				quotas, err := client.GetAllQuotas(ctx, "project", "logicalcloud")
				Expect(len(quotas)).To(Equal(0))
			})
			// will uncomment after general mockdb issues resolved
			// It("delete when nothing exists should fail", func() {
			// 	err := client.DeleteQuota("project", "logicalcloud", "testquota")
			// 	Expect(err).Should(HaveOccurred())
			// })
			It("update after creation should succeed and return updated resource", func() {
				ctx := context.Background()
				quota := _createTestQuota("testquota")
				_, _ = client.CreateQuota(ctx, "project", "logicalcloud", quota)
				quota.MetaData.Description = "new description"
				quota, err := client.UpdateQuota(ctx, "project", "logicalcloud", "testquota", quota)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(quota.MetaData.QuotaName).To(Equal("testquota"))
				Expect(quota.MetaData.Description).To(Equal("new description"))
			})
			It("create followed by updating the name is disallowed and should fail", func() {
				ctx := context.Background()
				quota := _createTestQuota("testquota")
				_, _ = client.CreateQuota(ctx, "project", "logicalcloud", quota)
				quota.MetaData.QuotaName = "updated"
				quota, err := client.UpdateQuota(ctx, "project", "logicalcloud", "testquota", quota)
				Expect(err).Should(HaveOccurred())
				Expect(quota).To(Equal(dcm.Quota{}))
			})
		})
	})
})

func _createTestQuota(name string) dcm.Quota {
	quota := dcm.Quota{}
	quota.MetaData = dcm.QMetaDataList{
		QuotaName:   name,
		Description: "",
	}
	quota.Specification = map[string]string{}
	quota.Specification["limits.cpu"] = "4"
	quota.Specification["limits.memory"] = "4096"
	return quota
}
