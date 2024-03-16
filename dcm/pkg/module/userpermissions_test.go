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

var _ = Describe("Userpermissions", func() {

	var (
		mdb    *db.NewMockDB
		client *dcm.UserPermissionClient
	)

	BeforeEach(func() {
		client = dcm.NewUserPermissionClient()
		mdb = new(db.NewMockDB)
		mdb.Err = nil
		mdb.Items = []map[string]map[string][]byte{}
		db.DBconn = mdb
	})
	Describe("User permission operations", func() {
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
					NameSpace: "testns",
					Level:     "1",
				}
				mdb.Insert(context.Background(), "resources", lkey, nil, "data", lc)
			})
			It("creation should succeed and return the resource created", func() {
				ctx := context.Background()
				up := _createTestUserPermission("testup", "testns")
				userPermission, err := client.CreateUserPerm(ctx, "project", "logicalcloud", up)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(userPermission.MetaData.UserPermissionName).To(Equal("testup"))
				Expect(userPermission.Specification.Namespace).To(Equal("testns"))
				Expect(userPermission.Specification.APIGroups).To(Equal([]string{"", "apps"}))
				Expect(userPermission.Specification.Resources).To(Equal([]string{"deployments", "pods"}))
				Expect(userPermission.Specification.Verbs).To(Equal([]string{"get", "list"}))
			})
			It("creation should succeed and return the resource created (cluster-wide)", func() {
				ctx := context.Background()
				up := _createTestUserPermission("testup", "")
				userPermission, err := client.CreateUserPerm(ctx, "project", "logicalcloud", up)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(userPermission.MetaData.UserPermissionName).To(Equal("testup"))
				Expect(userPermission.Specification.Namespace).To(Equal(""))
				Expect(userPermission.Specification.APIGroups).To(Equal([]string{"", "apps"}))
				Expect(userPermission.Specification.Resources).To(Equal([]string{"deployments", "pods"}))
				Expect(userPermission.Specification.Verbs).To(Equal([]string{"get", "list"}))
			})
			It("get should fail and not return anything", func() {
				ctx := context.Background()
				userPermission, err := client.GetUserPerm(ctx, "project", "logicalcloud", "testup")
				Expect(err).Should(HaveOccurred())
				Expect(userPermission).To(Equal(dcm.UserPermission{}))
			})
			It("create followed by get should return what was created", func() {
				ctx := context.Background()
				up := _createTestUserPermission("testup", "testns")
				_, _ = client.CreateUserPerm(ctx, "project", "logicalcloud", up)
				userPermission, err := client.GetUserPerm(ctx, "project", "logicalcloud", "testup")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(userPermission).To(Equal(up))
			})
			It("create followed by get-all should return only what was created", func() {
				ctx := context.Background()
				up := _createTestUserPermission("testup", "testns")
				_, _ = client.CreateUserPerm(ctx, "project", "logicalcloud", up)
				userPermissions, err := client.GetAllUserPerms(ctx, "project", "logicalcloud")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(userPermissions)).To(Equal(1))
				Expect(userPermissions[0]).To(Equal(up))
			})
			It("three creates followed by get-all should return all that was created", func() {
				ctx := context.Background()
				up1 := _createTestUserPermission("testup1", "testns")
				up2 := _createTestUserPermission("testup2", "testns")
				up3 := _createTestUserPermission("testup3", "testns")
				_, _ = client.CreateUserPerm(ctx, "project", "logicalcloud", up1)
				_, _ = client.CreateUserPerm(ctx, "project", "logicalcloud", up2)
				_, _ = client.CreateUserPerm(ctx, "project", "logicalcloud", up3)
				userPermissions, err := client.GetAllUserPerms(ctx, "project", "logicalcloud")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(userPermissions)).To(Equal(3))
				Expect(userPermissions[0]).To(Equal(up1))
				Expect(userPermissions[1]).To(Equal(up2))
				Expect(userPermissions[2]).To(Equal(up3))
			})
			It("delete after creation should succeed and database remain empty", func() {
				ctx := context.Background()
				up := _createTestUserPermission("testup", "testns")
				_, _ = client.CreateUserPerm(ctx, "project", "logicalcloud", up)
				err := client.DeleteUserPerm(ctx, "project", "logicalcloud", "testup")
				Expect(err).ShouldNot(HaveOccurred())
				userPermissions, err := client.GetAllUserPerms(ctx, "project", "logicalcloud")
				Expect(len(userPermissions)).To(Equal(0))
			})
			// will uncomment after general mockdb issues resolved
			// It("delete when nothing exists should fail", func() {
			// 	err := client.DeleteUserPerm("project", "logicalcloud", "testup")
			// 	Expect(err).Should(HaveOccurred())
			// })
			It("update after creation should succeed and return updated resource", func() {
				ctx := context.Background()
				up := _createTestUserPermission("testup", "testns")
				_, _ = client.CreateUserPerm(ctx, "project", "logicalcloud", up)
				up.Specification.APIGroups = []string{"", "apps", "k8splugin.io"}
				userPermission, err := client.UpdateUserPerm(ctx, "project", "logicalcloud", "testup", up)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(userPermission.MetaData.UserPermissionName).To(Equal("testup"))
				Expect(userPermission.Specification.APIGroups).To(Equal([]string{"", "apps", "k8splugin.io"}))
				Expect(userPermission.Specification.Resources).To(Equal([]string{"deployments", "pods"}))
				Expect(userPermission.Specification.Verbs).To(Equal([]string{"get", "list"}))
			})
			It("create followed by updating the name is disallowed and should fail", func() {
				ctx := context.Background()
				up := _createTestUserPermission("testup", "testns")
				_, _ = client.CreateUserPerm(ctx, "project", "logicalcloud", up)
				up.MetaData.UserPermissionName = "updated"
				userPermission, err := client.UpdateUserPerm(ctx, "project", "logicalcloud", "testup", up)
				Expect(err).Should(HaveOccurred())
				Expect(userPermission).To(Equal(dcm.UserPermission{}))
			})
		})
	})
})

// _createTestUserPermission is an helper function to reduce code duplication
func _createTestUserPermission(name string, namespace string) dcm.UserPermission {

	up := dcm.UserPermission{}

	up.MetaData = dcm.UPMetaDataList{
		UserPermissionName: name,
		Description:        "",
		UserData1:          "",
		UserData2:          "",
	}
	up.Specification = dcm.UPSpec{
		Namespace: namespace,
		APIGroups: []string{"", "apps"},
		Resources: []string{"deployments", "pods"},
		Verbs:     []string{"get", "list"},
	}

	return up
}
