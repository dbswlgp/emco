// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package networkpolicy_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	"gitlab.com/project-emco/core/emco-base/src/dtc/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/nps/internal/networkpolicy"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

type contextForCompositeApp struct {
	context            appcontext.AppContext
	ctxval             interface{}
	compositeAppHandle interface{}
}

func makeAppContextForCompositeApp(p, ca, v, rName, dig string, namespace string, level string) (contextForCompositeApp, error) {
	appCtx := appcontext.AppContext{}
	ctxval, err := appCtx.InitAppContext()
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}
	compositeHandle, err := appCtx.CreateCompositeApp(context.Background())
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error creating CompositeApp handle")
	}
	compMetadata := appcontext.CompositeAppMeta{Project: p, CompositeApp: ca, Version: v, Release: rName, DeploymentIntentGroup: dig, Namespace: namespace, Level: level}
	err = appCtx.AddCompositeAppMeta(context.Background(), compMetadata)
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error Adding CompositeAppMeta")
	}

	cca := contextForCompositeApp{context: appCtx, ctxval: ctxval, compositeAppHandle: compositeHandle}

	return cca, nil

}

func CreateClusterLabel(provider, clu string) error {

	client := cluster.NewClusterClient()

	cp := cluster.ClusterProvider{
		Metadata: mtypes.Metadata{
			Name:        provider,
			Description: "add provider",
			UserData1:   "user data1",
			UserData2:   "user data2",
		},
	}

	_, _ = client.CreateClusterProvider(context.Background(), cp, false)
	c := cluster.Cluster{
		Metadata: mtypes.Metadata{
			Name:        clu,
			Description: "add cluster",
			UserData1:   "user data1",
			UserData2:   "user data2",
		},
	}
	cc := cluster.ClusterContent{
		Kubeconfig: "dummydata",
	}
	_, _ = client.CreateCluster(context.Background(), provider, c, cc)

	cl := cluster.ClusterLabel{
		LabelName: "networkpolicy-supported",
	}

	_, _ = client.CreateClusterLabel(context.Background(), provider, clu, cl, false)

	return nil

}

var _ = Describe("Action", func() {

	var (
		mdb *db.MockDB
		edb *contextdb.MockConDb

		TGI    module.TrafficGroupIntent
		TGIDBC *module.TrafficGroupIntentDbClient

		ISI    module.InboundServerIntent
		ISIDBC *module.InboundServerIntentDbClient

		ICI         module.InboundClientsIntent
		ICIDBC      *module.InboundClientsIntentDbClient
		expectedOut string = `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: testtgi-testisi
spec:
  podSelector:
    matchLabels:
      app: server
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: client
    - ipBlock:
        cidr: 0.0.0.0/0
    ports:
    - protocol: TCP
    - port: 4443
`
	)

	BeforeEach(func() {
		mdb = new(db.MockDB)
		mdb.Err = nil
		db.DBconn = mdb
		edb = new(contextdb.MockConDb)
		edb.Err = nil
		contextdb.Db = edb

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
			Spec: module.InbondServerIntentSpec{
				AppName:         "server",
				AppLabel:        "app=server",
				ServiceName:     "server-svc",
				ExternalName:    "",
				Port:            4443,
				Protocol:        "TCP",
				ExternalSupport: false,
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
			Spec: module.InboundClientsIntentSpec{
				AppName:     "client",
				AppLabel:    "app=client",
				ServiceName: "client-svc",
				Namespaces:  []string{},
				IpRange:     []string{},
			},
		}
	})

	Describe("App context", func() {
		It("successful update", func() {
			ctx := context.Background()
			cfca, err := makeAppContextForCompositeApp("project1", "ca", "v2", "r1", "dig", "n1", "app")
			Expect(err).To(BeNil())
			sap, err := cfca.context.AddApp(context.Background(), cfca.compositeAppHandle, "server")
			Expect(err).To(BeNil())
			sapc1, err := cfca.context.AddCluster(context.Background(), sap, "provider1+cluster1")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(context.Background(), sapc1, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err := json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(context.Background(), sapc1, "resource", "order", string(resOrder))

			cap, err := cfca.context.AddApp(context.Background(), cfca.compositeAppHandle, "client")
			Expect(err).To(BeNil())
			capc2, err := cfca.context.AddCluster(context.Background(), cap, "provider2+cluster2")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(context.Background(), capc2, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err = json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(context.Background(), capc2, "resource", "order", string(resOrder))

			tgi, err := (*TGIDBC).CreateTrafficGroupIntent(ctx, TGI, "project1", "ca", "v2", "dig", false)
			Expect(tgi).To(Equal(TGI))
			Expect(err).To(BeNil())
			isi, err := (*ISIDBC).CreateServerInboundIntent(ctx, ISI, "project1", "ca", "v2", "dig", "testtgi", false)
			Expect(isi).To(Equal(ISI))
			Expect(err).To(BeNil())
			ici, err := (*ICIDBC).CreateClientsInboundIntent(ctx, ICI, "project1", "ca", "v2", "dig", "testtgi", "testisi", false)
			Expect(ici).To(Equal(ICI))
			Expect(err).To(BeNil())

			_ = CreateClusterLabel("provider1", "cluster1")
			_ = CreateClusterLabel("provider2", "cluster2")

			contextID := fmt.Sprintf("%v", cfca.ctxval)

			err = networkpolicy.UpdateAppContext(ctx, "testtgi", contextID)
			Expect(err).To(BeNil())
			rh, err := cfca.context.GetResourceHandle(context.Background(), "server", "provider1+cluster1", "testtgi-testisi")
			Expect(err).To(BeNil())

			v, err := cfca.context.GetValue(context.Background(), rh)
			Expect(v).To(Equal(expectedOut))

			resorder, err := cfca.context.GetResourceInstruction(context.Background(), "server", "provider1+cluster1", "order")
			Expect(err).To(BeNil())
			resOrder, err = json.Marshal(map[string][]string{"resorder": []string{"r1", "testtgi-testisi"}})
			Expect(err).To(BeNil())
			Expect(resorder).To(Equal(string(resOrder)))

		})
		It("cover invalid appCtx error", func() {
			ctx := context.Background()
			edb.Err = pkgerrors.New("Error invalid appCtx ID:")
			err := networkpolicy.UpdateAppContext(ctx, "testtgi", "dummycontextid")
			Expect(err).To(HaveOccurred())
		})
		It("cover invalid meta data error", func() {
			ctx := context.Background()
			appCtx := appcontext.AppContext{}
			ctxval, err := appCtx.InitAppContext()
			Expect(err).To(BeNil())
			contextID := fmt.Sprintf("%v", ctxval)
			err = networkpolicy.UpdateAppContext(ctx, "testtgi", contextID)
			Expect(err).To(HaveOccurred())
		})
		It("cover error getting server inbound intents", func() {
			ctx := context.Background()
			cfca, err := makeAppContextForCompositeApp("project1", "ca", "v2", "r1", "dig", "n1", "app")
			Expect(err).To(BeNil())
			sap, err := cfca.context.AddApp(context.Background(), cfca.compositeAppHandle, "server")
			Expect(err).To(BeNil())
			sapc1, err := cfca.context.AddCluster(context.Background(), sap, "provider1+cluster1")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(context.Background(), sapc1, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err := json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(context.Background(), sapc1, "resource", "order", string(resOrder))

			cap, err := cfca.context.AddApp(context.Background(), cfca.compositeAppHandle, "client")
			Expect(err).To(BeNil())
			capc2, err := cfca.context.AddCluster(context.Background(), cap, "provider2+cluster2")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(context.Background(), capc2, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err = json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(context.Background(), capc2, "resource", "order", string(resOrder))
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = networkpolicy.UpdateAppContext(ctx, "testtgi", contextID)
			Expect(err).To(HaveOccurred())
		})
		It("cover getting empty client inbound intents", func() {
			ctx := context.Background()
			cfca, err := makeAppContextForCompositeApp("project1", "ca", "v2", "r1", "dig", "n1", "app")
			Expect(err).To(BeNil())
			sap, err := cfca.context.AddApp(context.Background(), cfca.compositeAppHandle, "server")
			Expect(err).To(BeNil())
			sapc1, err := cfca.context.AddCluster(context.Background(), sap, "provider1+cluster1")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(context.Background(), sapc1, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err := json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(context.Background(), sapc1, "resource", "order", string(resOrder))

			cap, err := cfca.context.AddApp(context.Background(), cfca.compositeAppHandle, "client")
			Expect(err).To(BeNil())
			capc2, err := cfca.context.AddCluster(context.Background(), cap, "provider2+cluster2")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(context.Background(), capc2, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err = json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(context.Background(), capc2, "resource", "order", string(resOrder))
			tgi, err := (*TGIDBC).CreateTrafficGroupIntent(ctx, TGI, "project1", "ca", "v2", "dig", false)
			Expect(tgi).To(Equal(TGI))
			Expect(err).To(BeNil())
			isi, err := (*ISIDBC).CreateServerInboundIntent(ctx, ISI, "project1", "ca", "v2", "dig", "testtgi", false)
			Expect(isi).To(Equal(ISI))
			Expect(err).To(BeNil())
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = networkpolicy.UpdateAppContext(ctx, "testtgi", contextID)
			Expect(err).To(BeNil())
		})
		It("cover invalid cluster name", func() {
			ctx := context.Background()
			cfca, err := makeAppContextForCompositeApp("project1", "ca", "v2", "r1", "dig", "n1", "app")
			Expect(err).To(BeNil())
			sap, err := cfca.context.AddApp(context.Background(), cfca.compositeAppHandle, "server")
			Expect(err).To(BeNil())
			sapc1, err := cfca.context.AddCluster(context.Background(), sap, "provider1-cluster1")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(context.Background(), sapc1, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err := json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(context.Background(), sapc1, "resource", "order", string(resOrder))

			cap, err := cfca.context.AddApp(context.Background(), cfca.compositeAppHandle, "client")
			Expect(err).To(BeNil())
			capc2, err := cfca.context.AddCluster(context.Background(), cap, "provider2-cluster2")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(context.Background(), capc2, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err = json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(context.Background(), capc2, "resource", "order", string(resOrder))
			tgi, err := (*TGIDBC).CreateTrafficGroupIntent(ctx, TGI, "project1", "ca", "v2", "dig", false)
			Expect(tgi).To(Equal(TGI))
			Expect(err).To(BeNil())
			isi, err := (*ISIDBC).CreateServerInboundIntent(ctx, ISI, "project1", "ca", "v2", "dig", "testtgi", false)
			Expect(isi).To(Equal(ISI))
			Expect(err).To(BeNil())
			ici, err := (*ICIDBC).CreateClientsInboundIntent(ctx, ICI, "project1", "ca", "v2", "dig", "testtgi", "testisi", false)
			Expect(ici).To(Equal(ICI))
			Expect(err).To(BeNil())
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = networkpolicy.UpdateAppContext(ctx, "testtgi", contextID)
			Expect(err).To(HaveOccurred())
		})
	})
})
