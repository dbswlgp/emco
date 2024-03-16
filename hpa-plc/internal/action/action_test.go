package action_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	clmcontrollerpb "gitlab.com/project-emco/core/emco-base/src/clm/pkg/grpc/controller-eventchannel"
	"gitlab.com/project-emco/core/emco-base/src/hpa-plc/internal/action"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
	hpaMod "gitlab.com/project-emco/core/emco-base/src/hpa-plc/pkg/module"
	orchLog "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	orchMod "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	connector "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/utils"
)

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

func TestAction(t *testing.T) {

	fmt.Printf("\n================== hpa-placementcontroller TestAction .. start ==================\n")

	orchLog.SetLoglevel(logrus.InfoLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "HPA-PLACEMENT-CONTROLLER")

	fmt.Printf("\n================== hpa-placementcontroller TestAction .. end ==================\n")
}

var _ = Describe("HPA-PLACEMENT-CONTROLLER", func() {

	var (
		mdb *db.MockDB
		edb *contextdb.MockConDb

		project    string = "p"
		compApp    string = "ca"
		version    string = "v1"
		dig        string = "dig"
		app1       string = "client"
		app2       string = "server"
		logicCloud string = "default"
		release    string = "r1"
		namespace  string = "n1"

		hpaIntentName1        string = "hpa-intent-1"
		hpaConsumerName1      string = "hpa-consumer-1"
		replicaCount          int64  = 1
		replicaCountMax       int64  = 20
		deploymentName1       string = "r1-http-client"
		containerName1        string = "http-client-1"
		containerName2        string = "http-client-2"
		hpaIntentName2        string = "hpa-intent-2"
		hpaIntentName3        string = "hpa-intent-3"
		hpaConsumerName2      string = "hpa-consumer-2"
		hpaAllocResourceName1 string = "hpa-alloc-resource-1"
	)

	var cfca contextForCompositeApp

	BeforeEach(func() {
		//logrus.SetOutput(ioutil.Discard)
		fmt.Printf("\n\n================== GINKGO ACTION TESTCASE START .. [%v] ==================\n\n", CurrentGinkgoTestDescription().TestText)

		// mongo mockdb
		mdb = &db.MockDB{
			Items: []map[string]map[string][]byte{
				{
					orchMod.ProjectKey{ProjectName: project}.String(): {
						"data": []byte(
							"{\"project-name\":\"" + project + "\"," +
								"\"description\":\"Test project for unit testing\"}"),
					},
					orchMod.CompositeAppKey{CompositeAppName: compApp,
						Version: version, Project: project}.String(): {
						"data": []byte(
							"{\"metadata\":{" +
								"\"name\":\"" + compApp + "\"," +
								"\"description\":\"description\"," +
								"\"userData1\":\"user data\"," +
								"\"userData2\":\"user data\"" +
								"}," +
								"\"spec\":{" +
								"\"version\":\"version of the composite app\"}}"),
					},
					orchMod.AppKey{App: "", Project: project, CompositeApp: compApp, CompositeAppVersion: version}.String(): {
						"data": []byte(
							"{" +
								"\"metadata\": {" +
								"\"name\":\"" + app1 + "\"," +
								"\"description\": \"Test App for unit testing\"," +
								"\"userData1\": \"userData1\"," +
								"\"userData2\": \"userData2\"}" +
								"}"),
						"appcontent": []byte(
							"{" +
								"\"FileContent\": \"sample file content\"" +
								"}"),
					},
					orchMod.DeploymentIntentGroupKey{
						Name:         dig,
						Project:      project,
						CompositeApp: compApp,
						Version:      version,
					}.String(): {
						"data": []byte(
							"{" +
								"\"metadata\":{" +
								"\"name\":\"" + dig + "\"," +
								"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
								"\"userData1\": \"userData1\"," +
								"\"userData2\": \"userData2\"}" +
								"}"),
					},
					hpaMod.HpaIntentKey{IntentName: "",
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"data": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaIntentName1 + "\"," +
								"\"Description\":\"Test Intent for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"app\":\"" + app1 + "\"}" +
								"}"),
					},
					hpaMod.HpaIntentKey{IntentName: hpaIntentName1,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"data": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaIntentName1 + "\"," +
								"\"Description\":\"Test Intent for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"app\":\"" + app1 + "\"}" +
								"}"),
					},
					hpaMod.HpaConsumerKey{ConsumerName: "", IntentName: hpaIntentName1,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"data": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaConsumerName1 + "\"," +
								"\"Description\":\"Test Consumer for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"replicas\":" + fmt.Sprint(replicaCount) + "," +
								"\"name\":\"" + deploymentName1 + "\"," +
								"\"container\":\"" + containerName1 + "\"}" +
								"}"),
					},
					hpaMod.HpaConsumerKey{ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"data": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaConsumerName1 + "\"," +
								"\"Description\":\"Test Consumer for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"replicas\":" + fmt.Sprint(replicaCount) + "," +
								"\"name\":\"" + deploymentName1 + "\"," +
								"\"container\":\"" + containerName1 + "\"}" +
								"}"),
					},
					hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"data": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
								"\"Description\":\"Test Resource for unit testing\"," +
								"\"UserData1\":\"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"allocatable\":true," +
								"\"resource\":{\"name\":\"cpu\", \"requests\":15, \"limits\":15}" +
								"}" +
								"}"),
					},
				},
				{
					orchMod.ProjectKey{ProjectName: project}.String(): {
						"data": []byte(
							"{\"project-name\":\"" + project + "\"," +
								"\"description\":\"Test project for unit testing\"}"),
					},
					orchMod.CompositeAppKey{CompositeAppName: compApp,
						Version: version, Project: project}.String(): {
						"data": []byte(
							"{\"metadata\":{" +
								"\"name\":\"" + compApp + "\"," +
								"\"description\":\"description\"," +
								"\"userData1\":\"user data\"," +
								"\"userData2\":\"user data\"" +
								"}," +
								"\"spec\":{" +
								"\"version\":\"version of the composite app\"}}"),
					},
					orchMod.AppKey{App: "", Project: project, CompositeApp: compApp, CompositeAppVersion: version}.String(): {
						"data": []byte(
							"{" +
								"\"metadata\": {" +
								"\"name\":\"" + app2 + "\"," +
								"\"description\": \"Test App for unit testing\"," +
								"\"userData1\": \"userData1\"," +
								"\"userData2\": \"userData2\"}" +
								"}"),
						"appcontent": []byte(
							"{" +
								"\"FileContent\": \"sample file content\"" +
								"}"),
					},
					orchMod.DeploymentIntentGroupKey{
						Name:         dig,
						Project:      project,
						CompositeApp: compApp,
						Version:      version,
					}.String(): {
						"data": []byte(
							"{" +
								"\"metadata\":{" +
								"\"name\":\"" + dig + "\"," +
								"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
								"\"userData1\": \"userData1\"," +
								"\"userData2\": \"userData2\"}" +
								"}"),
					},
					hpaMod.HpaIntentKey{IntentName: "",
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"data": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaIntentName2 + "\"," +
								"\"Description\":\"Test Intent for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"app\":\"" + app2 + "\"}" +
								"}"),
					},
					hpaMod.HpaIntentKey{IntentName: hpaIntentName2,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"data": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaIntentName2 + "\"," +
								"\"Description\":\"Test Intent for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"app\":\"" + app2 + "\"}" +
								"}"),
					},
					hpaMod.HpaConsumerKey{ConsumerName: "", IntentName: hpaIntentName2,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"data": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaConsumerName2 + "\"," +
								"\"Description\":\"Test Consumer for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"replicas\":" + fmt.Sprint(replicaCount) + "," +
								"\"name\":\"" + deploymentName1 + "\"," +
								"\"container\":\"" + containerName2 + "\"}" +
								"}"),
					},
					hpaMod.HpaConsumerKey{ConsumerName: hpaConsumerName2, IntentName: hpaIntentName2,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"data": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaConsumerName2 + "\"," +
								"\"Description\":\"Test Consumer for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"replicas\":" + fmt.Sprint(replicaCount) + "," +
								"\"name\":\"" + deploymentName1 + "\"," +
								"\"container\":\"" + containerName2 + "\"}" +
								"}"),
					},
					hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName2, IntentName: hpaIntentName2,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"data": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
								"\"Description\":\"Test Resource for unit testing\"," +
								"\"UserData1\":\"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"allocatable\":true," +
								"\"resource\":{\"name\":\"cpu\", \"requests\":15, \"limits\":15}" +
								"}" +
								"}"),
					},
				},
			},
		}
		mdb.Err = nil
		db.DBconn = mdb

		// etcd mockdb
		edb = new(contextdb.MockConDb)
		edb.Err = nil
		contextdb.Db = edb

		// Initialize etcd with default values
		var err error
		ctx := context.Background()
		cfca, err = makeAppContextForCompositeApp(project, compApp, version, release, dig, namespace, logicCloud)
		Expect(err).To(BeNil())

		cap, err := cfca.context.AddApp(ctx, cfca.compositeAppHandle, app1)
		Expect(err).To(BeNil())

		ch, err := cfca.context.AddCluster(ctx, cap, "provider1-cluster1")
		Expect(err).To(BeNil())

		sap, err := cfca.context.AddApp(ctx, cfca.compositeAppHandle, app2)
		Expect(err).To(BeNil())

		sh, err := cfca.context.AddCluster(ctx, sap, "provider1-cluster2")
		Expect(err).To(BeNil())
		err = cfca.context.AddClusterMetaGrp(ctx, ch, "1")
		Expect(err).To(BeNil())
		err = cfca.context.AddClusterMetaGrp(ctx, sh, "1")
		Expect(err).To(BeNil())

		// Use Kube Fake client for unit-testing
		connector.IsTestKubeClient = true
	})

	AfterEach(func() {
		mdb = nil
		db.DBconn = mdb

		// etcd mockdb
		edb = nil
		contextdb.Db = edb
	})

	Describe("Filter Clusters", func() {

		It("*** GINKGO ACTION TESTCASE: successful allocatable-resource filter-clusters", func() {
			ctx := context.Background()
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: successful non-allocatable-resource filter-cluster", func() {
			ctx := context.Background()
			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"data": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
						"\"Description\":\"Test Resource for unit testing\"," +
						"\"UserData1\":\"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"allocatable\":false," +
						"\"resource\":{\"key\":\"feature.node.kubernetes.io/intel_qat\", \"value\":\"true\"}" +
						"}" +
						"}"),
			}

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful action to due to invalid app-context", func() {
			ctx := context.Background()
			err := action.FilterClusters(ctx, "1234")
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful allocatable-resource filter-clusters due to invalid request count", func() {
			ctx := context.Background()
			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"data": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
						"\"Description\":\"Test Resource for unit testing\"," +
						"\"UserData1\":\"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"allocatable\":true," +
						"\"resource\":{\"name\":\"cpu\", \"requests\":20, \"limits\":20}" +
						"}" +
						"}"),
			}

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful allocatable-resource filter-clusters due to non-exsiting k8s resource name", func() {
			ctx := context.Background()
			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"data": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
						"\"Description\":\"Test Resource for unit testing\"," +
						"\"UserData1\":\"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"allocatable\":true," +
						"\"resource\":{\"name\":\"cpu1\", \"requests\":1, \"limits\":1}" +
						"}" +
						"}"),
			}

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful non-allocatable-resource filter-cluster due to non-existing label", func() {
			ctx := context.Background()
			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"data": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
						"\"Description\":\"Test Resource for unit testing\"," +
						"\"UserData1\":\"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"allocatable\":false," +
						"\"resource\":{\"key\":\"feature.node.kubernetes1.io/intel_qa\", \"value\":\"true\"}" +
						"}" +
						"}"),
			}

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful filter-cluster when there are no apps in composite-app apps", func() {
			ctx := context.Background()
			(mdb.Items[0])[orchMod.AppKey{App: "", Project: project, CompositeApp: compApp, CompositeAppVersion: version}.String()] = nil
			(mdb.Items[1])[orchMod.AppKey{App: "", Project: project, CompositeApp: compApp, CompositeAppVersion: version}.String()] = nil
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: successful filter-cluster when hpa-intent app is not one of composite-app apps", func() {
			ctx := context.Background()
			(mdb.Items[0])[orchMod.AppKey{App: "", Project: project, CompositeApp: compApp, CompositeAppVersion: version}.String()] = map[string][]byte{
				"data": []byte(
					"{" +
						"\"metadata\": {" +
						"\"name\":\"" + app2 + "\"," +
						"\"description\": \"Test App for unit testing\"," +
						"\"userData1\": \"userData1\"," +
						"\"userData2\": \"userData2\"}" +
						"}"),
				"appcontent": []byte(
					"{" +
						"\"FileContent\": \"sample file content\"" +
						"}"),
			}
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(BeNil())
		})

		It("failed filter-cluster with NO hpa-intents", func() {
			ctx := context.Background()
			(mdb.Items[0])[hpaMod.HpaIntentKey{IntentName: "",
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = nil
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(BeNil())

		})

		It("*** GINKGO ACTION TESTCASE: Successful hpa-intent filter-cluster for non-existing hpa-consumers", func() {
			ctx := context.Background()
			(mdb.Items[0])[hpaMod.HpaIntentKey{IntentName: "",
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"data": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaIntentName3 + "\"," +
						"\"Description\":\"Test Intent for unit testing\"," +
						"\"UserData1\": \"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"app\":\"" + app1 + "\"}" +
						"}"),
			}
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(BeNil())

		})

		It("*** GINKGO ACTION TESTCASE: Successful filter-cluster for non-existing hpa-consumer", func() {
			ctx := context.Background()
			(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: "", IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"data": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaConsumerName2 + "\"," +
						"\"Description\":\"Test Consumer for unit testing\"," +
						"\"UserData1\": \"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"name\":\"" + deploymentName1 + "\"," +
						"\"container\":\"" + containerName1 + "\"}" +
						"}"),
			}
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: successful filter-cluster with NO hpa-resources", func() {
			ctx := context.Background()
			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = nil

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: successful non-allocatable resource filter-cluster with maximum replicaCount", func() {
			ctx := context.Background()
			(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: "", IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"data": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaConsumerName1 + "\"," +
						"\"Description\":\"Test Consumer for unit testing\"," +
						"\"UserData1\": \"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"replicas\":" + fmt.Sprint(replicaCountMax) + "," +
						"\"name\":\"" + deploymentName1 + "\"," +
						"\"container\":\"" + containerName1 + "\"}" +
						"}"),
			}

			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"data": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
						"\"Description\":\"Test Resource for unit testing\"," +
						"\"UserData1\":\"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"allocatable\":false," +
						"\"resource\":{\"key\":\"feature.node.kubernetes.io/intel_qat\", \"value\":\"true\"}" +
						"}" +
						"}"),
			}

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful filter-clusters with original kube client", func() {
			ctx := context.Background()
			// Use Kube Fake client for unit-testing
			connector.IsTestKubeClient = false

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful filter-cluster with maximum replicaCount", func() {
			ctx := context.Background()
			(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: "", IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"data": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaConsumerName1 + "\"," +
						"\"Description\":\"Test Consumer for unit testing\"," +
						"\"UserData1\": \"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"replicas\":" + fmt.Sprint(replicaCountMax) + "," +
						"\"name\":\"" + deploymentName1 + "\"," +
						"\"container\":\"" + containerName1 + "\"}" +
						"}"),
			}

			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"data": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
						"\"Description\":\"Test Resource for unit testing\"," +
						"\"UserData1\":\"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"allocatable\":true," +
						"\"resource\":{\"name\":\"cpu\", \"requests\":5, \"limits\":5}" +
						"}" +
						"}"),
			}
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: successful filter-cluster with NO hpa-consumers", func() {
			ctx := context.Background()
			(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: "", IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = nil
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(ctx, contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: successful publish received cluster-create event", func() {
			var req clmcontrollerpb.ClmControllerEventRequest
			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED

			err := action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: successful publish received cluster-modify event", func() {
			var req clmcontrollerpb.ClmControllerEventRequest
			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_UPDATED

			err := action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful publish received unknown event", func() {
			var req clmcontrollerpb.ClmControllerEventRequest
			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = 5

			err := action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: successful publish received cluster-delete event", func() {
			var req clmcontrollerpb.ClmControllerEventRequest

			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED
			err := action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())

			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_DELETED
			err = action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: successful get labels", func() {
			ctx := context.Background()
			var req clmcontrollerpb.ClmControllerEventRequest
			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED

			err := action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())

			_, err = action.GetKubeClusterLabels(ctx, "provider1", "cluster1")
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful publish cluster-create", func() {
			mdb.Err = pkgerrors.New("Error")
			var req clmcontrollerpb.ClmControllerEventRequest
			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED

			err := action.Publish(context.TODO(), &req)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful publish cluster-delete", func() {
			var req clmcontrollerpb.ClmControllerEventRequest

			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED
			err := action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())

			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			mdb.Err = pkgerrors.New("Error")
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_DELETED
			err = action.Publish(context.TODO(), &req)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful get labels cluster-create", func() {
			ctx := context.Background()
			var req clmcontrollerpb.ClmControllerEventRequest
			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED

			err := action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())

			mdb.Err = pkgerrors.New("Error")
			_, err = action.GetKubeClusterLabels(ctx, "provider1", "cluster1")
			Expect(err).To(HaveOccurred())
		})

	})
})
