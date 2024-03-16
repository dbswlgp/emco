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

var _ = Describe("WorkflowIntentHook",
	func() {
		var (
			project   orchmodule.Project
			pClient   *orchmodule.ProjectClient
			app       orchmodule.CompositeApp
			aClient   *orchmodule.CompositeAppClient
			diGroup   orchmodule.DeploymentIntentGroup
			digClient *orchmodule.DeploymentIntentGroupClient
			wfhIntent model.WorkflowHookIntent
			iClient   *module.WorkflowIntentClient
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

				iClient = module.NewWorkflowIntentClient()
				wfhIntent = model.WorkflowHookIntent{
					Metadata: mtypes.Metadata{
						Name:        "WorkflowIntentHookSampleName",
						Description: "Example Description",
					},
				}

				mdb = new(db.NewMockDB)
				mdb.Err = nil
				db.DBconn = mdb
			},
		)

		/* Workflow Hook Intent Routes */

		Describe("Create a workflow hook intent",
			func() {
				It("Successful Create a workflow hook intent",
					func() {
						ctx := context.Background()
						// set up prerequisites
						_, err := (*pClient).CreateProject(ctx, project, false)
						Expect(err).To(BeNil())
						_, err = (*aClient).CreateCompositeApp(ctx, app, "testProj", false)
						Expect(err).To(BeNil())
						_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, diGroup, "testProj", "app", "v1", true)
						Expect(err).To(BeNil())

						// create a workflow hook
						resp, err := (*iClient).CreateWorkflowHookIntent(ctx, wfhIntent, "testProj", "app", "v1", "diGroup", false)
						Expect(err).To(BeNil())
						Expect(resp.Metadata.Name).To(Equal(wfhIntent.Metadata.Name))
					})
			})

		Describe("Get a workflow hook intent",
			func() {
				It("Successful get a workflow hook intent",
					func() {
						ctx := context.Background()
						// set up prerequisites
						_, err := (*pClient).CreateProject(ctx, project, false)
						Expect(err).To(BeNil())
						_, err = (*aClient).CreateCompositeApp(ctx, app, "testProj", false)
						Expect(err).To(BeNil())
						_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, diGroup, "testProj", "app", "v1", true)
						Expect(err).To(BeNil())

						// create a workflow hook
						resp, err := (*iClient).CreateWorkflowHookIntent(ctx, wfhIntent, "testProj", "app", "v1", "diGroup", true)
						Expect(err).To(BeNil())
						Expect(resp.Metadata.Name).To(Equal(wfhIntent.Metadata.Name))

						// get a workflow
						resp, err = (*iClient).GetWorkflowHookIntent(ctx, wfhIntent.Metadata.Name, "testProj", "app", "v1", "diGroup")
						Expect(err).To(BeNil())
						Expect(resp.Metadata.Name).To(Equal(wfhIntent.Metadata.Name))

					})
			})

		Describe("Get workflow hook intents",
			func() {
				It("Successful get workflow hook intents",
					func() {
						ctx := context.Background()
						// set up prerequisites
						_, err := (*pClient).CreateProject(ctx, project, false)
						Expect(err).To(BeNil())
						_, err = (*aClient).CreateCompositeApp(ctx, app, "testProj", false)
						Expect(err).To(BeNil())
						_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, diGroup, "testProj", "app", "v1", true)
						Expect(err).To(BeNil())
						wfhIntentTwo := model.WorkflowHookIntent{
							Metadata: mtypes.Metadata{
								Name: "WorkflowHookTwo",
							},
						}

						// create a workflow hook
						_, err = (*iClient).CreateWorkflowHookIntent(ctx, wfhIntent, "testProj", "app", "v1", "diGroup", true)
						Expect(err).To(BeNil())

						// create a workflow hook
						_, err = (*iClient).CreateWorkflowHookIntent(ctx, wfhIntentTwo, "testProj", "app", "v1", "diGroup", true)
						Expect(err).To(BeNil())

						// get a workflow
						resp, err := (*iClient).GetWorkflowHookIntents(ctx, "testProj", "app", "v1", "diGroup")
						Expect(err).To(BeNil())

						// make sure workflow hooks are inside.
						Expect(resp[0].Metadata.Name).To(Equal(wfhIntent.Metadata.Name))
						Expect(resp[1].Metadata.Name).To(Equal(wfhIntentTwo.Metadata.Name))

						// make sure only workflow hooks are returned, and not intents
						Expect(len(resp)).To(Equal(2))

					})
			})

		Describe("Delete Workflow hook Intent",
			func() {
				It("Successful delete workflow hook intents",
					func() {
						ctx := context.Background()
						// set up prerequisites
						_, err := (*pClient).CreateProject(ctx, project, false)
						Expect(err).To(BeNil())
						_, err = (*aClient).CreateCompositeApp(ctx, app, "testProj", false)
						Expect(err).To(BeNil())
						_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, diGroup, "testProj", "app", "v1", true)
						Expect(err).To(BeNil())

						// create a workflow hook
						_, err = (*iClient).CreateWorkflowHookIntent(ctx, wfhIntent, "testProj", "app", "v1", "diGroup", true)
						Expect(err).To(BeNil())

						// delete the workflow.
						err = (*iClient).DeleteWorkflowHookIntent(ctx, wfhIntent.Metadata.Name, "testProj", "app", "v1", "diGroup")
						Expect(err).To(BeNil())

					})
			})

		It("Error Deleting DNE",
			func() {
				ctx := context.Background()
				// set up prerequisites
				_, err := (*pClient).CreateProject(ctx, project, false)
				Expect(err).To(BeNil())
				_, err = (*aClient).CreateCompositeApp(ctx, app, "testProj", false)
				Expect(err).To(BeNil())
				_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, diGroup, "testProj", "app", "v1", true)
				Expect(err).To(BeNil())

				// delete the workflow.
				err = (*iClient).DeleteWorkflowHookIntent(ctx, wfhIntent.Metadata.Name, "testProj", "app", "v1", "diGroup")
				Expect(err).To(BeNil())

			})

		Describe("Update Workflow Intent",
			func() {
				It("Successful update workflow hook intent",
					func() {
						ctx := context.Background()
						// set up prerequisites
						_, err := (*pClient).CreateProject(ctx, project, false)
						Expect(err).To(BeNil())
						_, err = (*aClient).CreateCompositeApp(ctx, app, "testProj", false)
						Expect(err).To(BeNil())
						_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, diGroup, "testProj", "app", "v1", true)
						Expect(err).To(BeNil())

						// create a workflow hook
						_, err = (*iClient).CreateWorkflowHookIntent(ctx, wfhIntent, "testProj", "app", "v1", "diGroup", false)
						Expect(err).To(BeNil())

						wfhIntent.Metadata.Description = "new description."

						// update the workflow
						resp, err := (*iClient).CreateWorkflowHookIntent(ctx, wfhIntent, "testProj", "app", "v1", "diGroup", true)
						Expect(err).To(BeNil())
						Expect(resp.Metadata.Description).To(Equal(wfhIntent.Metadata.Description))

					})

			})

	},
)
