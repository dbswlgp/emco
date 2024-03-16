// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package action_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	orch "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	catypes "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
	cacontext "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/utils"
	sfcmodel "gitlab.com/project-emco/core/emco-base/src/sfc/pkg/model"
	sfcmodule "gitlab.com/project-emco/core/emco-base/src/sfc/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/sfcclient/internal/action"
	"gitlab.com/project-emco/core/emco-base/src/sfcclient/pkg/model"
	"gitlab.com/project-emco/core/emco-base/src/sfcclient/pkg/module"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// For testing need:

var TestCA1 catypes.CompositeApp = catypes.CompositeApp{
	CompMetadata: appcontext.CompositeAppMeta{
		Project:               "testp",
		CompositeApp:          "clientCA",
		Version:               "v1",
		Release:               "r1",
		DeploymentIntentGroup: "dig1",
		Namespace:             "default",
		Level:                 "0",
	},
	AppOrder: []string{"leftApp", "rightApp", "a3"},
	Apps: map[string]*catypes.App{
		"leftApp": &catypes.App{
			Name: "leftApp",
			Clusters: map[string]*catypes.Cluster{
				"provider1+cluster1": &catypes.Cluster{
					Name: "provider1+cluster1",
					Resources: map[string]*catypes.AppResource{
						"r1": &catypes.AppResource{Name: "r1", Data: "a1c1r1"},
						"r2": &catypes.AppResource{Name: "r2", Data: "a1c1r2"},
					},
					ResOrder: []string{"r1", "r2"}},
				"provider1+cluster2": &catypes.Cluster{
					Name: "provider1+cluster2",
					Resources: map[string]*catypes.AppResource{
						"r3": &catypes.AppResource{Name: "r3", Data: "a1c2r3"},
						"r4": &catypes.AppResource{Name: "r4", Data: "a1c2r4"},
					},
					ResOrder: []string{"r3", "r4"}}},
		},
		"rightApp": &catypes.App{
			Name: "rightApp",
			Clusters: map[string]*catypes.Cluster{
				"provider1+cluster1": &catypes.Cluster{
					Name: "provider1+cluster1",
					Resources: map[string]*catypes.AppResource{
						"r1": &catypes.AppResource{Name: "r3", Data: "a2c1r1"},
						"r2": &catypes.AppResource{Name: "r4", Data: "a2c1r2"},
					},
					ResOrder: []string{"r3", "r4"}},
				"provider1+cluster2": &catypes.Cluster{
					Name: "provider1+cluster2",
					Resources: map[string]*catypes.AppResource{
						"r3": &catypes.AppResource{Name: "r3", Data: "a2c2r3"},
						"r4": &catypes.AppResource{Name: "r4", Data: "a2c2r4"},
						"r5": &catypes.AppResource{Name: "r4", Data: "a2c2r5"},
					},
					ResOrder: []string{"r3", "r4", "r5"}}},
		},
		"a3": &catypes.App{
			Name: "a3",
			Clusters: map[string]*catypes.Cluster{
				"provider1+cluster2": &catypes.Cluster{
					Name: "provider1+cluster2",
					Resources: map[string]*catypes.AppResource{
						"r6": &catypes.AppResource{Name: "r3", Data: "a3c2r6"},
						"r7": &catypes.AppResource{Name: "r4", Data: "a3c2r7"},
					},
					ResOrder: []string{"r3", "r4"}}},
		},
	},
}

var _ = Describe("SFC Client Action", func() {
	var (
		// Mock AppContext variables
		cdb          *contextdb.MockConDb
		contextIdCA1 string

		// Mock DB variables
		proj       orch.Project
		projClient *orch.ProjectClient

		// for test setup, there are two CAs (and corresponding child resources)
		// chainCa - is the CA to define the Network Chain (SFC)
		// ca - is the CA to define apps to connect to the SFC as a client
		chainCa  orch.CompositeApp
		ca       orch.CompositeApp
		caClient *orch.CompositeAppClient

		chainDig  orch.DeploymentIntentGroup
		dig       orch.DeploymentIntentGroup
		digClient *orch.DeploymentIntentGroupClient

		sfcIntent                      sfcmodel.SfcIntent
		sfcLeftClientSelectorIntent    sfcmodel.SfcClientSelectorIntent
		sfcRightClientSelectorIntent   sfcmodel.SfcClientSelectorIntent
		sfcLeftProviderNetworkIntent   sfcmodel.SfcProviderNetworkIntent
		sfcRightProviderNetworkIntent  sfcmodel.SfcProviderNetworkIntent
		sfcIntentClient                *sfcmodule.SfcIntentClient
		sfcClientSelectorIntentClient  *sfcmodule.SfcClientSelectorIntentClient
		sfcProviderNetworkIntentClient *sfcmodule.SfcProviderNetworkIntentClient

		sfcLeftClientIntent  model.SfcClientIntent
		sfcRightClientIntent model.SfcClientIntent
		sfcClient            *module.SfcClient

		//resultingCA catypes.CompositeApp

		mdb *db.NewMockDB
	)

	BeforeEach(func() {
		ctx := context.Background()
		cdb = new(contextdb.MockConDb)
		cdb.Err = nil
		contextdb.Db = cdb

		// make an AppContext
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
				Name: "clientCA",
			},
			Spec: orch.CompositeAppSpec{
				Version: "v1",
			},
		}
		chainCa = orch.CompositeApp{
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
		chainDig = orch.DeploymentIntentGroup{
			MetaData: orch.DepMetaData{
				Name: "chainDig1",
			},
			Spec: orch.DepSpecData{
				Profile:      "profilename",
				Version:      "r1",
				LogicalCloud: "logCloud",
			},
		}

		sfcIntentClient = sfcmodule.NewSfcIntentClient()
		sfcIntent = sfcmodel.SfcIntent{
			Metadata: sfcmodel.Metadata{
				Name: "sfcIntentName",
			},
			Spec: sfcmodel.SfcIntentSpec{
				ChainType: sfcmodel.RoutingChainType,
				Namespace: "chainspace",
			},
		}

		sfcClientSelectorIntentClient = sfcmodule.NewSfcClientSelectorIntentClient()
		sfcLeftClientSelectorIntent = sfcmodel.SfcClientSelectorIntent{
			Metadata: sfcmodel.Metadata{
				Name: "sfcLeftClientSelectorIntentName",
			},
			Spec: sfcmodel.SfcClientSelectorIntentSpec{
				ChainEnd: "left",
				PodSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":   "leftapp",
						"label": "leftvalue",
					},
				},
			},
		}

		sfcRightClientSelectorIntent = sfcmodel.SfcClientSelectorIntent{
			Metadata: sfcmodel.Metadata{
				Name: "sfcRightClientSelectorIntentName",
			},
			Spec: sfcmodel.SfcClientSelectorIntentSpec{
				ChainEnd: "right",
				PodSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "rightapp",
					},
				},
			},
		}

		sfcProviderNetworkIntentClient = sfcmodule.NewSfcProviderNetworkIntentClient()
		sfcLeftProviderNetworkIntent = sfcmodel.SfcProviderNetworkIntent{
			Metadata: sfcmodel.Metadata{
				Name: "sfcLeftProviderNetworkIntentName",
			},
			Spec: sfcmodel.SfcProviderNetworkIntentSpec{
				ChainEnd:    "left",
				NetworkName: "leftPNet",
				GatewayIp:   "10.10.10.1",
				Subnet:      "10.10.10.0/24",
			},
		}

		sfcRightProviderNetworkIntent = sfcmodel.SfcProviderNetworkIntent{
			Metadata: sfcmodel.Metadata{
				Name: "sfcRightProviderNetworkIntentName",
			},
			Spec: sfcmodel.SfcProviderNetworkIntentSpec{
				ChainEnd:    "right",
				NetworkName: "rightPNet",
				GatewayIp:   "11.11.11.1",
				Subnet:      "11.11.11.0/24",
			},
		}

		sfcClient = module.NewSfcClient()
		sfcLeftClientIntent = model.SfcClientIntent{
			Metadata: model.Metadata{
				Name: "sfcLeftClientIntentName",
			},
			Spec: model.SfcClientIntentSpec{
				ChainEnd:                   "left",
				ChainName:                  "sfcIntentName",
				ChainCompositeApp:          "chainCA",
				ChainCompositeAppVersion:   "v1",
				ChainDeploymentIntentGroup: "chainDig1",
				AppName:                    "chainNetctl",
				WorkloadResource:           "chainNetctl",
				ResourceType:               "Deployment",
			},
		}

		sfcRightClientIntent = model.SfcClientIntent{
			Metadata: model.Metadata{
				Name: "sfcRightClientIntentName",
			},
			Spec: model.SfcClientIntentSpec{
				ChainEnd:                   "right",
				ChainName:                  "sfcIntentName",
				ChainCompositeApp:          "chainCA",
				ChainCompositeAppVersion:   "v1",
				ChainDeploymentIntentGroup: "chainDig1",
				AppName:                    "chainNetctl",
				WorkloadResource:           "chainNetctl",
				ResourceType:               "Deployment",
			},
		}

		mdb = new(db.NewMockDB)
		mdb.Err = nil
		db.DBconn = mdb

		// set up prerequisites - client CA
		_, err := (*projClient).CreateProject(ctx, proj, false)
		Expect(err).To(BeNil())
		_, err = (*caClient).CreateCompositeApp(ctx, ca, "testp", false)
		Expect(err).To(BeNil())
		_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, dig, "testp", "clientCA", "v1", true)
		Expect(err).To(BeNil())
		_, err = (*sfcClient).CreateSfcClientIntent(ctx, sfcLeftClientIntent, "testp", "clientCA", "v1", "dig1", false)
		Expect(err).To(BeNil())
		_, err = (*sfcClient).CreateSfcClientIntent(ctx, sfcRightClientIntent, "testp", "clientCA", "v1", "dig1", false)
		Expect(err).To(BeNil())

		// set up prerequisites - chain CA
		_, err = (*caClient).CreateCompositeApp(ctx, chainCa, "testp", false)
		Expect(err).To(BeNil())
		_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, chainDig, "testp", "chainCA", "v1", true)
		Expect(err).To(BeNil())
		_, err = (*sfcIntentClient).CreateSfcIntent(ctx, sfcIntent, "testp", "chainCA", "v1", "chainDig1", false)
		Expect(err).To(BeNil())
		_, err = (*sfcClientSelectorIntentClient).CreateSfcClientSelectorIntent(ctx, sfcLeftClientSelectorIntent, "testp", "chainCA", "v1", "chainDig1", "sfcIntentName", false)
		Expect(err).To(BeNil())
		_, err = (*sfcClientSelectorIntentClient).CreateSfcClientSelectorIntent(ctx, sfcRightClientSelectorIntent, "testp", "chainCA", "v1", "chainDig1", "sfcIntentName", false)
		Expect(err).To(BeNil())
		_, err = (*sfcProviderNetworkIntentClient).CreateSfcProviderNetworkIntent(ctx, sfcLeftProviderNetworkIntent, "testp", "chainCA", "v1", "chainDig1", "sfcIntentName", false)
		Expect(err).To(BeNil())
		_, err = (*sfcProviderNetworkIntentClient).CreateSfcProviderNetworkIntent(ctx, sfcRightProviderNetworkIntent, "testp", "chainCA", "v1", "chainDig1", "sfcIntentName", false)
		Expect(err).To(BeNil())
	})

	It("Successful Apply SFC to an App Context", func() {
		ctx := context.Background()
		// TODO - unit test code needs to be completed (setup of test appcontexts, etc. need work)
		err := action.UpdateAppContext(ctx, "netctl", contextIdCA1)
		Expect(err).To(HaveOccurred())
	})

})
