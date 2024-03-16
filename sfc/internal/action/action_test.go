// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package action_test

import (
	"strings"

	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	orch "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	catypes "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
	cacontext "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/utils"
	"gitlab.com/project-emco/core/emco-base/src/sfc/internal/action"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/model"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/module"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// For testing need:
// 1. A Mock AppContext
// 2. with a set of apps which have been place in a set of clusters
//    a) variation 1:  all clusters have all apps that are part of the chain
//    b) variation 2:  no cluster has all apps that are listed in the chain (is this ok?)
//    c) variation 3:  1 cluster does not have all app, another does (treat as ok)
// 3. A Mock Deployment Intent Group
//    a) variation 1:  No network control that matches the input network control intent
//    b) variation 2:  Network control intent matches the input network control intent
// 4. A Mock SFC Intent
//    a) variation 1:  Zero SFC Intents
//    b) variation 2:  One SFC Intent
//    c) variation 3:  Two SFC Intents
func init() {
	var edb *contextdb.MockConDb
	edb = new(contextdb.MockConDb)
	edb.Err = nil
	contextdb.Db = edb
}

const deployment1 = "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: td\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      t: abc\n  template:\n    metadata:\n      labels:\n        t: abc\n    spec:\n      containers:\n      - name: nginx\n        image: t:1.2\n"

var TestCA1 catypes.CompositeApp = catypes.CompositeApp{
	CompMetadata: appcontext.CompositeAppMeta{
		Project:               "testp",
		CompositeApp:          "chainCA",
		Version:               "v1",
		Release:               "r1",
		DeploymentIntentGroup: "dig1",
		Namespace:             "default",
		Level:                 "0",
	},
	AppOrder: []string{"a1", "a2", "a3"},
	Apps: map[string]*catypes.App{
		"a1": &catypes.App{
			Name: "a1",
			Clusters: map[string]*catypes.Cluster{
				"provider1+cluster1": &catypes.Cluster{
					Name: "provider1+cluster1",
					Resources: map[string]*catypes.AppResource{
						"r1": &catypes.AppResource{Name: "r1+Deployment", Data: deployment1},
						"r2": &catypes.AppResource{Name: "r2+Deployment", Data: deployment1},
					},
					ResOrder: []string{"r1", "r2"}},
				"provider1+cluster2": &catypes.Cluster{
					Name: "provider1+cluster2",
					Resources: map[string]*catypes.AppResource{
						"r1": &catypes.AppResource{Name: "r1+Deployment", Data: deployment1},
						"r2": &catypes.AppResource{Name: "r2+Deployment", Data: deployment1},
					},
					ResOrder: []string{"r1", "r2"}}},
		},
		"a2": &catypes.App{
			Name: "a2",
			Clusters: map[string]*catypes.Cluster{
				"provider1+cluster1": &catypes.Cluster{
					Name: "provider1+cluster1",
					Resources: map[string]*catypes.AppResource{
						"r3": &catypes.AppResource{Name: "r3+Deployment", Data: deployment1},
						"r4": &catypes.AppResource{Name: "r4+Deployment", Data: deployment1},
						"r5": &catypes.AppResource{Name: "r5+Deployment", Data: deployment1},
					},
					ResOrder: []string{"r3", "r4", "r5"}},
				"provider1+cluster2": &catypes.Cluster{
					Name: "provider1+cluster2",
					Resources: map[string]*catypes.AppResource{
						"r3": &catypes.AppResource{Name: "r3+Deployment", Data: deployment1},
						"r4": &catypes.AppResource{Name: "r4+Deployment", Data: deployment1},
						"r5": &catypes.AppResource{Name: "r5+Deployment", Data: deployment1},
					},
					ResOrder: []string{"r3", "r4", "r5"}}},
		},
		"a3": &catypes.App{
			Name: "a3",
			Clusters: map[string]*catypes.Cluster{
				"provider1+cluster2": &catypes.Cluster{
					Name: "provider1+cluster2",
					Resources: map[string]*catypes.AppResource{
						"r6": &catypes.AppResource{Name: "r6+Deployment", Data: deployment1},
						"r7": &catypes.AppResource{Name: "r7+Deployment", Data: deployment1},
					},
					ResOrder: []string{"r6", "r7"}}},
		},
	},
}

var _ = Describe("SFCAction", func() {
	var (
		// Mock AppContext variables
		cdb          *contextdb.MockConDb
		contextIdCA1 string

		// Mock DB variables
		proj       orch.Project
		projClient *orch.ProjectClient

		ca       orch.CompositeApp
		caClient *orch.CompositeAppClient

		dig       orch.DeploymentIntentGroup
		digClient *orch.DeploymentIntentGroupClient

		sfcIntent                     model.SfcIntent
		sfcIntent2                    model.SfcIntent
		sfcLinkIntent1                model.SfcLinkIntent
		sfcLinkIntent2                model.SfcLinkIntent
		sfcLinkIntent3                model.SfcLinkIntent
		sfcLeftClientSelectorIntent   model.SfcClientSelectorIntent
		sfcRightClientSelectorIntent  model.SfcClientSelectorIntent
		sfcLeftProviderNetworkIntent  model.SfcProviderNetworkIntent
		sfcRightProviderNetworkIntent model.SfcProviderNetworkIntent
		sfcClient                     *module.SfcIntentClient
		sfcClientSelectorClient       *module.SfcClientSelectorIntentClient
		sfcLinkIntentClient           *module.SfcLinkIntentClient
		sfcProviderNetworkClient      *module.SfcProviderNetworkIntentClient

		resultingCA catypes.CompositeApp

		mdb *db.NewMockDB
	)

	BeforeEach(func() {
		ctx := context.Background()
		cdb = new(contextdb.MockConDb)
		cdb.Err = nil
		contextdb.Db = cdb
		cid, _ := cacontext.CreateCompApp(ctx, TestCA1)
		contextIdCA1 = cid

		// setup the mock DB resources
		// (needs to match the mock AppContext)
		projClient = orch.NewProjectClient()
		proj = orch.Project{
			MetaData: orch.ProjectMetaData{
				Name: "testp",
			},
		}

		caClient = orch.NewCompositeAppClient()
		ca = orch.CompositeApp{
			Metadata: orch.CompositeAppMetaData{
				Name: "chainCA",
			},
			Spec: orch.CompositeAppSpec{
				Version: "v1",
			},
		}

		digClient = orch.NewDeploymentIntentGroupClient()
		dig = orch.DeploymentIntentGroup{
			MetaData: orch.DepMetaData{
				Name: "dig1",
			},
			Spec: orch.DepSpecData{
				Profile:      "profilename",
				Version:      "r1",
				LogicalCloud: "logCloud",
			},
		}

		sfcClient = module.NewSfcIntentClient()
		sfcIntent = model.SfcIntent{
			Metadata: model.Metadata{
				Name: "sfcIntentName",
			},
			Spec: model.SfcIntentSpec{
				ChainType: model.RoutingChainType,
				Namespace: "chainspace",
			},
		}
		sfcLinkIntentClient = module.NewSfcLinkIntentClient()
		sfcLinkIntent1 = model.SfcLinkIntent{
			Metadata: model.Metadata{
				Name: "sfcLinkIntent1",
			},
			Spec: model.SfcLinkIntentSpec{
				LeftNet:          "left-virtual",
				RightNet:         "dyn1",
				LinkLabel:        "app=a1",
				AppName:          "a1",
				WorkloadResource: "r1",
				ResourceType:     "Deployment",
			},
		}
		sfcLinkIntent2 = model.SfcLinkIntent{
			Metadata: model.Metadata{
				Name: "sfcLinkIntent2",
			},
			Spec: model.SfcLinkIntentSpec{
				LeftNet:          "dyn1",
				RightNet:         "right-virtual",
				LinkLabel:        "app=a2",
				AppName:          "a2",
				WorkloadResource: "r3",
				ResourceType:     "Deployment",
			},
		}
		sfcIntent2 = model.SfcIntent{
			Metadata: model.Metadata{
				Name: "sfcIntentName2",
			},
			Spec: model.SfcIntentSpec{
				ChainType: model.RoutingChainType,
				Namespace: "chainspace",
			},
		}
		sfcLinkIntent3 = model.SfcLinkIntent{
			Metadata: model.Metadata{
				Name: "sfcLinkIntent3",
			},
			Spec: model.SfcLinkIntentSpec{
				LeftNet:          "left-virtual",
				RightNet:         "right-virtual",
				LinkLabel:        "app=a3",
				AppName:          "a3",
				WorkloadResource: "r6",
				ResourceType:     "Deployment",
			},
		}

		sfcClientSelectorClient = module.NewSfcClientSelectorIntentClient()
		sfcLeftClientSelectorIntent = model.SfcClientSelectorIntent{
			Metadata: model.Metadata{
				Name: "sfcLeftClientSelectorIntentName",
			},
			Spec: model.SfcClientSelectorIntentSpec{
				ChainEnd: "left",
				PodSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "leftapp",
					},
				},
			},
		}

		sfcRightClientSelectorIntent = model.SfcClientSelectorIntent{
			Metadata: model.Metadata{
				Name: "sfcRightClientSelectorIntentName",
			},
			Spec: model.SfcClientSelectorIntentSpec{
				ChainEnd: "right",
				PodSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "rightapp",
					},
				},
			},
		}

		sfcProviderNetworkClient = module.NewSfcProviderNetworkIntentClient()
		sfcLeftProviderNetworkIntent = model.SfcProviderNetworkIntent{
			Metadata: model.Metadata{
				Name: "sfcLeftProviderNetworkIntentName",
			},
			Spec: model.SfcProviderNetworkIntentSpec{
				ChainEnd:    "left",
				NetworkName: "leftPNet",
				GatewayIp:   "10.10.10.1",
				Subnet:      "10.10.10.0/24",
			},
		}

		sfcRightProviderNetworkIntent = model.SfcProviderNetworkIntent{
			Metadata: model.Metadata{
				Name: "sfcRightProviderNetworkIntentName",
			},
			Spec: model.SfcProviderNetworkIntentSpec{
				ChainEnd:    "right",
				NetworkName: "rightPNet",
				GatewayIp:   "11.11.11.1",
				Subnet:      "11.11.11.0/24",
			},
		}

		mdb = new(db.NewMockDB)
		mdb.Err = nil
		db.DBconn = mdb

		// set up prerequisites
		_, err := (*projClient).CreateProject(ctx, proj, false)
		Expect(err).To(BeNil())
		_, err = (*caClient).CreateCompositeApp(ctx, ca, "testp", false)
		Expect(err).To(BeNil())
		_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, dig, "testp", "chainCA", "v1", true)
		Expect(err).To(BeNil())
		_, err = (*sfcClient).CreateSfcIntent(ctx, sfcIntent, "testp", "chainCA", "v1", "dig1", false)
		Expect(err).To(BeNil())
		_, err = (*sfcLinkIntentClient).CreateSfcLinkIntent(ctx, sfcLinkIntent1, "testp", "chainCA", "v1", "dig1", "sfcIntentName", false)
		Expect(err).To(BeNil())
		_, err = (*sfcLinkIntentClient).CreateSfcLinkIntent(ctx, sfcLinkIntent2, "testp", "chainCA", "v1", "dig1", "sfcIntentName", false)
		Expect(err).To(BeNil())
		_, err = (*sfcClientSelectorClient).CreateSfcClientSelectorIntent(ctx, sfcLeftClientSelectorIntent, "testp", "chainCA", "v1", "dig1", "sfcIntentName", false)
		Expect(err).To(BeNil())
		_, err = (*sfcClientSelectorClient).CreateSfcClientSelectorIntent(ctx, sfcRightClientSelectorIntent, "testp", "chainCA", "v1", "dig1", "sfcIntentName", false)
		Expect(err).To(BeNil())
		_, err = (*sfcProviderNetworkClient).CreateSfcProviderNetworkIntent(ctx, sfcLeftProviderNetworkIntent, "testp", "chainCA", "v1", "dig1", "sfcIntentName", false)
		Expect(err).To(BeNil())
		_, err = (*sfcProviderNetworkClient).CreateSfcProviderNetworkIntent(ctx, sfcRightProviderNetworkIntent, "testp", "chainCA", "v1", "dig1", "sfcIntentName", false)
		Expect(err).To(BeNil())
	})

	It("No Client Selector intents", func() {
		ctx := context.Background()
		err := (*sfcClientSelectorClient).DeleteSfcClientSelectorIntent(ctx, "sfcLeftClientSelectorIntentName", "testp", "chainCA", "v1", "dig1", "sfcIntentName")
		Expect(err).To(BeNil())
		err = (*sfcClientSelectorClient).DeleteSfcClientSelectorIntent(ctx, "sfcRightClientSelectorIntentName", "testp", "chainCA", "v1", "dig1", "sfcIntentName")
		Expect(err).To(BeNil())

		err = action.UpdateAppContext(ctx, "dig1", contextIdCA1)
		Expect(err).To(BeNil())
	})

	It("No Left Client Selector intent", func() {
		ctx := context.Background()
		err := (*sfcClientSelectorClient).DeleteSfcClientSelectorIntent(ctx, "sfcLeftClientSelectorIntentName", "testp", "chainCA", "v1", "dig1", "sfcIntentName")
		Expect(err).To(BeNil())

		err = action.UpdateAppContext(ctx, "dig1", contextIdCA1)
		Expect(err).To(BeNil())
	})

	It("No Right Client Selector intent", func() {
		ctx := context.Background()
		err := (*sfcClientSelectorClient).DeleteSfcClientSelectorIntent(ctx, "sfcRightClientSelectorIntentName", "testp", "chainCA", "v1", "dig1", "sfcIntentName")
		Expect(err).To(BeNil())

		err = action.UpdateAppContext(ctx, "dig1", contextIdCA1)
		Expect(err).To(BeNil())
	})

	It("No Both Provider Network intents", func() {
		ctx := context.Background()
		err := (*sfcProviderNetworkClient).DeleteSfcProviderNetworkIntent(ctx, "sfcLeftProviderNetworkIntentName", "testp", "chainCA", "v1", "dig1", "sfcIntentName")
		Expect(err).To(BeNil())
		err = (*sfcProviderNetworkClient).DeleteSfcProviderNetworkIntent(ctx, "sfcRightProviderNetworkIntentName", "testp", "chainCA", "v1", "dig1", "sfcIntentName")
		Expect(err).To(BeNil())

		err = action.UpdateAppContext(ctx, "dig1", contextIdCA1)
		Expect(err).To(BeNil())
	})

	It("No Left Provider Network intent", func() {
		ctx := context.Background()
		err := (*sfcProviderNetworkClient).DeleteSfcProviderNetworkIntent(ctx, "sfcLeftProviderNetworkIntentName", "testp", "chainCA", "v1", "dig1", "sfcIntentName")
		Expect(err).To(BeNil())

		err = action.UpdateAppContext(ctx, "dig1", contextIdCA1)
		Expect(err).To(BeNil())
	})

	It("No Right Provider Network intent", func() {
		ctx := context.Background()
		err := (*sfcProviderNetworkClient).DeleteSfcProviderNetworkIntent(ctx, "sfcRightProviderNetworkIntentName", "testp", "chainCA", "v1", "dig1", "sfcIntentName")
		Expect(err).To(BeNil())

		err = action.UpdateAppContext(ctx, "dig1", contextIdCA1)
		Expect(err).To(BeNil())
	})

	It("Successful Apply SFC to an App Context", func() {
		ctx := context.Background()
		err := action.UpdateAppContext(ctx, "dig1", contextIdCA1)
		Expect(err).To(BeNil())
	})

	It("No SFC intents", func() {
		ctx := context.Background()
		// delete all the SFC intents
		err := (*sfcProviderNetworkClient).DeleteSfcProviderNetworkIntent(ctx, "sfcLeftProviderNetworkIntentName", "testp", "chainCA", "v1", "dig1", "sfcIntentName")
		Expect(err).To(BeNil())
		err = (*sfcProviderNetworkClient).DeleteSfcProviderNetworkIntent(ctx, "sfcRightProviderNetworkIntentName", "testp", "chainCA", "v1", "dig1", "sfcIntentName")
		Expect(err).To(BeNil())
		err = (*sfcClientSelectorClient).DeleteSfcClientSelectorIntent(ctx, "sfcLeftClientSelectorIntentName", "testp", "chainCA", "v1", "dig1", "sfcIntentName")
		Expect(err).To(BeNil())
		err = (*sfcClientSelectorClient).DeleteSfcClientSelectorIntent(ctx, "sfcRightClientSelectorIntentName", "testp", "chainCA", "v1", "dig1", "sfcIntentName")
		Expect(err).To(BeNil())
		err = (*sfcClient).DeleteSfcIntent(ctx, "sfcIntentName", "testp", "chainCA", "v1", "dig1")
		Expect(err).To(BeNil())

		resultingCA, err = cacontext.ReadAppContext(ctx, contextIdCA1)
		cacontext.PrintCompositeApp(resultingCA)

		err = action.UpdateAppContext(ctx, "dig1", contextIdCA1)
		Expect(strings.Contains(err.Error(), "No SFC Intents are defined for the Deployment Intent Group")).To(Equal(true))
	})

	It("Successful Apply two SFCs to an App Context", func() {
		ctx := context.Background()
		// set up second SFC
		_, err := (*sfcClient).CreateSfcIntent(ctx, sfcIntent2, "testp", "chainCA", "v1", "dig1", false)
		Expect(err).To(BeNil())
		_, err = (*sfcLinkIntentClient).CreateSfcLinkIntent(ctx, sfcLinkIntent3, "testp", "chainCA", "v1", "dig1", "sfcIntentName2", false)
		Expect(err).To(BeNil())
		_, err = (*sfcClientSelectorClient).CreateSfcClientSelectorIntent(ctx, sfcLeftClientSelectorIntent, "testp", "chainCA", "v1", "dig1", "sfcIntentName2", false)
		Expect(err).To(BeNil())
		_, err = (*sfcClientSelectorClient).CreateSfcClientSelectorIntent(ctx, sfcRightClientSelectorIntent, "testp", "chainCA", "v1", "dig1", "sfcIntentName2", false)
		Expect(err).To(BeNil())
		_, err = (*sfcProviderNetworkClient).CreateSfcProviderNetworkIntent(ctx, sfcLeftProviderNetworkIntent, "testp", "chainCA", "v1", "dig1", "sfcIntentName2", false)
		Expect(err).To(BeNil())
		_, err = (*sfcProviderNetworkClient).CreateSfcProviderNetworkIntent(ctx, sfcRightProviderNetworkIntent, "testp", "chainCA", "v1", "dig1", "sfcIntentName2", false)
		Expect(err).To(BeNil())

		resultingCA, err = cacontext.ReadAppContext(ctx, contextIdCA1)
		cacontext.PrintCompositeApp(resultingCA)

		err = action.UpdateAppContext(ctx, "dig1", contextIdCA1)
		Expect(err).To(BeNil())
	})
})
