// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// These test cases are to validate the module level functionalities.
package module_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	orchmodule "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
	"gitlab.com/project-emco/core/emco-base/src/tac/pkg/model"
	"gitlab.com/project-emco/core/emco-base/src/tac/pkg/module"
)

var _ = Describe("WorkerIntent",
	func() {
		var (
			project   orchmodule.Project
			pClient   *orchmodule.ProjectClient
			app       orchmodule.CompositeApp
			aClient   *orchmodule.CompositeAppClient
			diGroup   orchmodule.DeploymentIntentGroup
			digClient *orchmodule.DeploymentIntentGroupClient
			wfhIntent model.WorkflowHookIntent
			hiClient  *module.WorkflowIntentClient
			wInent    model.WorkerIntent
			iClient   *module.WorkerIntentClient
			mdb       *db.NewMockDB
		)

		BeforeEach(
			func() {
				pClient = orchmodule.NewProjectClient()
				project = orchmodule.Project{
					MetaData: orchmodule.ProjectMetaData{
						Name: "testProj",
					},
				}

				aClient = orchmodule.NewCompositeAppClient()
				app = orchmodule.CompositeApp{
					Metadata: orchmodule.CompositeAppMetaData{
						Name: "app",
					},
					Spec: orchmodule.CompositeAppSpec{
						Version: "v1",
					},
				}

				digClient = orchmodule.NewDeploymentIntentGroupClient()
				diGroup = orchmodule.DeploymentIntentGroup{
					MetaData: orchmodule.DepMetaData{
						Name: "diGroup",
					},
					Spec: orchmodule.DepSpecData{
						Profile:      "profilename",
						Version:      "testver",
						LogicalCloud: "logCloud",
					},
				}

				hiClient = module.NewWorkflowIntentClient()
				wfhIntent = model.WorkflowHookIntent{
					Metadata: mtypes.Metadata{
						Name:        "WorkflowIntentHookSampleName",
						Description: "Example Description",
					},
				}

				iClient = module.NewWorkerIntentClient()
				wInent = model.WorkerIntent{
					Metadata: mtypes.Metadata{
						Name:        "WorkerIntentName",
						Description: "Example Description",
					},
				}

				mdb = new(db.NewMockDB)
				mdb.Err = nil
				db.DBconn = mdb
			},
		)

		Describe("Create a worker intent",
			func() {
				It("Successful Create of worker intent",
					func() {
						// set up prerequisites
						_, err := (*pClient).CreateProject(context.Background(), project, false)
						Expect(err).To(BeNil())
						_, err = (*aClient).CreateCompositeApp(context.Background(), app, "testProj", false)
						Expect(err).To(BeNil())
						_, _, err = (*digClient).CreateDeploymentIntentGroup(context.Background(), diGroup, "testProj", "app", "v1", true)
						Expect(err).To(BeNil())

						// create a workflow hook
						resp, err := (*hiClient).CreateWorkflowHookIntent(context.Background(), wfhIntent, "testProj", "app", "v1", "diGroup", false)
						Expect(err).To(BeNil())
						Expect(resp.Metadata.Name).To(Equal(wfhIntent.Metadata.Name))

						// create a worker intent
						res, err := (*iClient).CreateOrUpdateWorkerIntent(wInent, "WorkerIntentName", "testProj", "app", "v1", "diGroup", false)
						Expect(err).To(BeNil())
						Expect(res.Metadata.Name).To(Equal(wInent.Metadata.Name))
					},
				)
			},
		)

		Describe("Get a worker intent",
			func() {
				It("Successful get a worker intent",
					func() {
						// set up prerequisites
						_, err := (*pClient).CreateProject(context.Background(), project, false)
						Expect(err).To(BeNil())
						_, err = (*aClient).CreateCompositeApp(context.Background(), app, "testProj", false)
						Expect(err).To(BeNil())
						_, _, err = (*digClient).CreateDeploymentIntentGroup(context.Background(), diGroup, "testProj", "app", "v1", true)
						Expect(err).To(BeNil())

						// create a workflow hook
						resp, err := (*hiClient).CreateWorkflowHookIntent(context.Background(), wfhIntent, "testProj", "app", "v1", "diGroup", true)
						Expect(err).To(BeNil())
						Expect(resp.Metadata.Name).To(Equal(wfhIntent.Metadata.Name))

						// create a worker intent
						res, err := (*iClient).CreateOrUpdateWorkerIntent(wInent, "WorkerIntentName", "testProj", "app", "v1", "diGroup", false)
						Expect(err).To(BeNil())
						Expect(res.Metadata.Name).To(Equal(wInent.Metadata.Name))

						// get a workflow
						res, err = (*iClient).GetWorkerIntent(wInent.Metadata.Name, "testProj", "app", "v1", "diGroup", "WorkerIntentName")
						Expect(err).To(BeNil())
						Expect(resp.Metadata.Name).To(Equal(wfhIntent.Metadata.Name))

					})
			})

		Describe("Delete Workflow hook Intent",
			func() {
				It("Successful delete workflow hook intents",
					func() {
						// set up prerequisites
						_, err := (*pClient).CreateProject(context.Background(), project, false)
						Expect(err).To(BeNil())
						_, err = (*aClient).CreateCompositeApp(context.Background(), app, "testProj", false)
						Expect(err).To(BeNil())
						_, _, err = (*digClient).CreateDeploymentIntentGroup(context.Background(), diGroup, "testProj", "app", "v1", true)
						Expect(err).To(BeNil())

						// create a workflow hook
						_, err = (*hiClient).CreateWorkflowHookIntent(context.Background(), wfhIntent, "testProj", "app", "v1", "diGroup", true)
						Expect(err).To(BeNil())

						// create a worker intent
						_, err = (*iClient).CreateOrUpdateWorkerIntent(wInent, "WorkerIntentName", "testProj", "app", "v1", "diGroup", false)
						Expect(err).To(BeNil())

						// delete the workflow.
						err = (*iClient).DeleteWorkerIntents("testProj", "app", "v1", "diGroup", wfhIntent.Metadata.Name, wInent.Metadata.Name)
						Expect(err).To(BeNil())

					})
			})
	},
)
