package migrationcontroller_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"context"
	pkgerrors "github.com/pkg/errors"
	migrationcontrollerserver "gitlab.com/project-emco/core/emco-base/src/migration/pkg/grpc/migrationcontrollerserver"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	migrationcontrollerpb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/migrationcontroller"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	orchLog "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

func TestMigrationControllerServer(t *testing.T) {

	fmt.Printf("\n================== TestMigrationControllerServer .. start ==================\n")

	orchLog.SetLoglevel(logrus.InfoLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "MigrationControllerServer")

	fmt.Printf("\n================== TestMigrationControllerServer .. end ==================\n")
}

type contextForCompositeApp struct {
	context            appcontext.AppContext
	ctxval             interface{}
	compositeAppHandle interface{}
}

func makeAppContextForCompositeApp(p, ca, v, rName, dig string, namespace string, level string) (contextForCompositeApp, error) {
	ctx := context.Background()
	appCtx := appcontext.AppContext{}
	ctxval, err := appCtx.InitAppContext()
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}
	compositeHandle, err := appCtx.CreateCompositeApp(ctx)
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error creating CompositeApp handle")
	}
	compMetadata := appcontext.CompositeAppMeta{Project: p, CompositeApp: ca, Version: v, Release: rName, DeploymentIntentGroup: dig, Namespace: namespace, Level: level}
	err = appCtx.AddCompositeAppMeta(ctx, compMetadata)
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error Adding CompositeAppMeta")
	}

	cca := contextForCompositeApp{context: appCtx, ctxval: ctxval, compositeAppHandle: compositeHandle}

	return cca, nil
}

var _ = Describe("MigrationControllerServer", func() {

	var (
		edb *contextdb.MockConDb

		project    string = "p"
		compApp    string = "ca"
		version    string = "v1"
		dig        string = "dig"
		logicCloud string = "default"
		release    string = "r1"
		namespace  string = "n1"
	)

	BeforeEach(func() {

		// etcd mockdb
		edb = new(contextdb.MockConDb)
		edb.Err = nil
		contextdb.Db = edb

		// Initialize etcd with default values
		var err error
		_, err = makeAppContextForCompositeApp(project, compApp, version, release, dig, namespace, logicCloud)
		Expect(err).To(BeNil())

	})

	It("unsuccessful MigrationApps", func() {
		MigrationcontrollerServer := migrationcontrollerserver.NewMigrationControllerServer()
		resp, _ := MigrationcontrollerServer.FilterClusters(context.TODO(), nil)
		Expect(resp.Status).To(Equal(false))
	})

	It("unsuccessful MigrationApps request AppContext is empty", func() {
		var req migrationcontrollerpb.ResourceRequest
		req.AppContext = ""

		MigrationcontrollerServer := migrationcontrollerserver.NewMigrationControllerServer()
		resp, _ := MigrationcontrollerServer.FilterClusters(context.TODO(), &req)
		Expect(resp.Status).To(Equal(false))
	})

	It("unsuccessful MigrationApps request AppContext is invalid", func() {
		var req migrationcontrollerpb.ResourceRequest
		req.AppContext = "1234"

		MigrationcontrollerServer := migrationcontrollerserver.NewMigrationControllerServer()
		resp, _ := MigrationcontrollerServer.FilterClusters(context.TODO(), &req)
		Expect(resp.Status).To(Equal(false))
	})
})
