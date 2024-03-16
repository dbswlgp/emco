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

var _ = Describe("Keyvalue", func() {

	var (
		mdb    *db.NewMockDB
		client *dcm.KeyValueClient
	)

	BeforeEach(func() {
		client = dcm.NewKeyValueClient()
		mdb = new(db.NewMockDB)
		mdb.Err = nil
		mdb.Items = []map[string]map[string][]byte{}
		db.DBconn = mdb
	})
	Describe("Key value operations", func() {
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
					UserData1:   "",
					UserData2:   "",
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
					UserData1:   "",
					UserData2:   "",
				}
				lc.Specification = common.Spec{
					NameSpace: "anything",
					Level:     "1",
				}
				mdb.Insert(context.Background(), "resources", lkey, nil, "data", lc)
			})
			It("creation should succeed and return the resource created", func() {
				ctx := context.Background()
				kv := _createTestKeyValue("testkv")
				keyValue, err := client.CreateKVPair(ctx, "project", "logicalcloud", kv)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(keyValue.MetaData.KeyValueName).To(Equal("testkv"))
				Expect(keyValue.MetaData.Description).To(Equal(""))
				Expect(keyValue.MetaData.UserData1).To(Equal(""))
				Expect(keyValue.MetaData.UserData2).To(Equal(""))
			})
			It("get should fail and not return anything", func() {
				ctx := context.Background()
				keyValue, err := client.GetKVPair(ctx, "project", "logicalcloud", "testkv")
				Expect(err).Should(HaveOccurred())
				Expect(keyValue).To(Equal(dcm.KeyValue{}))
			})
			It("create followed by get should return what was created", func() {
				ctx := context.Background()
				kv := _createTestKeyValue("testkv")
				_, _ = client.CreateKVPair(ctx, "project", "logicalcloud", kv)
				keyValue, err := client.GetKVPair(ctx, "project", "logicalcloud", "testkv")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(keyValue).To(Equal(kv))
			})
			It("create followed by get-all should return only what was created", func() {
				ctx := context.Background()
				kv := _createTestKeyValue("testkv")
				_, _ = client.CreateKVPair(ctx, "project", "logicalcloud", kv)
				keyValues, err := client.GetAllKVPairs(ctx, "project", "logicalcloud")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(keyValues)).To(Equal(1))
				Expect(keyValues[0]).To(Equal(kv))
			})
			It("three creates followed by get-all should return all that was created", func() {
				ctx := context.Background()
				kv1 := _createTestKeyValue("testkv1")
				kv2 := _createTestKeyValue("testkv2")
				kv3 := _createTestKeyValue("testkv3")
				_, _ = client.CreateKVPair(ctx, "project", "logicalcloud", kv1)
				_, _ = client.CreateKVPair(ctx, "project", "logicalcloud", kv2)
				_, _ = client.CreateKVPair(ctx, "project", "logicalcloud", kv3)
				keyValues, err := client.GetAllKVPairs(ctx, "project", "logicalcloud")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(keyValues)).To(Equal(3))
				Expect(keyValues[0]).To(Equal(kv1))
				Expect(keyValues[1]).To(Equal(kv2))
				Expect(keyValues[2]).To(Equal(kv3))
			})
			It("delete after creation should succeed and database remain empty", func() {
				ctx := context.Background()
				kv := _createTestKeyValue("testkv")
				_, _ = client.CreateKVPair(ctx, "project", "logicalcloud", kv)
				err := client.DeleteKVPair(ctx, "project", "logicalcloud", "testkv")
				Expect(err).ShouldNot(HaveOccurred())
				keyValues, err := client.GetAllKVPairs(ctx, "project", "logicalcloud")
				Expect(len(keyValues)).To(Equal(0))
			})
			// will uncomment after general mockdb issues resolved
			// It("delete when nothing exists should fail", func() {
			// 	err := client.DeleteKVPair("project", "logicalcloud", "testkv")
			// 	Expect(err).Should(HaveOccurred())
			// })
			It("update after creation should succeed and return updated resource", func() {
				ctx := context.Background()
				kv := _createTestKeyValue("testkv")
				_, _ = client.CreateKVPair(ctx, "project", "logicalcloud", kv)
				kv.MetaData.UserData1 = "new user data"
				keyValue, err := client.UpdateKVPair(ctx, "project", "logicalcloud", "testkv", kv)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(keyValue.MetaData.KeyValueName).To(Equal("testkv"))
				Expect(keyValue.MetaData.Description).To(Equal(""))
				Expect(keyValue.MetaData.UserData1).To(Equal("new user data"))
				Expect(keyValue.MetaData.UserData2).To(Equal(""))
			})
			It("create followed by updating the name is disallowed and should fail", func() {
				ctx := context.Background()
				kv := _createTestKeyValue("testkv")
				_, _ = client.CreateKVPair(ctx, "project", "logicalcloud", kv)
				kv.MetaData.KeyValueName = "updated"
				keyValue, err := client.UpdateKVPair(ctx, "project", "logicalcloud", "testkv", kv)
				Expect(err).Should(HaveOccurred())
				Expect(keyValue).To(Equal(dcm.KeyValue{}))
			})
		})
	})
})

// _createTestKeyValue is an helper function to reduce code duplication
func _createTestKeyValue(name string) dcm.KeyValue {
	return dcm.KeyValue{
		MetaData: dcm.KVMetaDataList{
			KeyValueName: name,
			Description:  "",
			UserData1:    "",
			UserData2:    "",
		},
	}
}
