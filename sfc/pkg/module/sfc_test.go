// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pkgerrors "github.com/pkg/errors"

	"context"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	orch "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/model"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/module"
)

var _ = Describe("SFCIntent", func() {
	var (
		proj       orch.Project
		projClient *orch.ProjectClient

		ca       orch.CompositeApp
		caClient *orch.CompositeAppClient

		dig       orch.DeploymentIntentGroup
		digClient *orch.DeploymentIntentGroupClient

		sfcIntent model.SfcIntent
		sfcClient *module.SfcIntentClient

		mdb *db.NewMockDB
	)

	BeforeEach(func() {
		projClient = orch.NewProjectClient()
		proj = orch.Project{
			MetaData: orch.ProjectMetaData{
				Name: "testproject",
			},
		}

		caClient = orch.NewCompositeAppClient()
		ca = orch.CompositeApp{
			Metadata: orch.CompositeAppMetaData{
				Name: "ca",
			},
			Spec: orch.CompositeAppSpec{
				Version: "v1",
			},
		}

		digClient = orch.NewDeploymentIntentGroupClient()
		dig = orch.DeploymentIntentGroup{
			MetaData: orch.DepMetaData{
				Name: "dig",
			},
			Spec: orch.DepSpecData{
				Profile:      "profilename",
				Version:      "testver",
				LogicalCloud: "logCloud",
			},
		}

		sfcClient = module.NewSfcIntentClient()
		sfcIntent = model.SfcIntent{
			Metadata: model.Metadata{
				Name: "sfcIntentName",
			},
		}

		mdb = new(db.NewMockDB)
		mdb.Err = nil
		db.DBconn = mdb
	})

	Describe("Create SFC intent", func() {
		It("successful creation of sfc intent", func() {
			ctx := context.Background()
			// set up prerequisites
			_, err := (*projClient).CreateProject(ctx, proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ctx, ca, "testproject", false)
			Expect(err).To(BeNil())
			_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, dig, "testproject", "ca", "v1", true)
			Expect(err).To(BeNil())

			// test SFC intent creation
			_, err = (*sfcClient).CreateSfcIntent(ctx, sfcIntent, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
		})
		It("followed by create again should return error", func() {
			ctx := context.Background()
			// set up prerequisites
			_, err := (*projClient).CreateProject(ctx, proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ctx, ca, "testproject", false)
			Expect(err).To(BeNil())
			_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, dig, "testproject", "ca", "v1", true)
			Expect(err).To(BeNil())

			// test SFC intent creation
			_, err = (*sfcClient).CreateSfcIntent(ctx, sfcIntent, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			// test SFC intent creation
			_, err = (*sfcClient).CreateSfcIntent(ctx, sfcIntent, "testproject", "ca", "v1", "dig", false)
			Expect(strings.Contains(err.Error(), "SFC Intent already exists")).To(Equal(true))
		})
		It("successful creation of sfc intent with update version of call", func() {
			ctx := context.Background()
			// set up prerequisites
			_, err := (*projClient).CreateProject(ctx, proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ctx, ca, "testproject", false)
			Expect(err).To(BeNil())
			_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, dig, "testproject", "ca", "v1", true)
			Expect(err).To(BeNil())

			// test SFC intent creation, with update form of call (exists bool == true)
			_, err = (*sfcClient).CreateSfcIntent(ctx, sfcIntent, "testproject", "ca", "v1", "dig", true)
			Expect(err).To(BeNil())
		})
		It("successful creation of sfc intent with update version of call", func() {
			ctx := context.Background()
			// set up prerequisites
			_, err := (*projClient).CreateProject(ctx, proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ctx, ca, "testproject", false)
			Expect(err).To(BeNil())
			_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, dig, "testproject", "ca", "v1", true)
			Expect(err).To(BeNil())

			// test SFC intent creation
			_, err = (*sfcClient).CreateSfcIntent(ctx, sfcIntent, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			// test SFC intent update (exists bool == true)
			_, err = (*sfcClient).CreateSfcIntent(ctx, sfcIntent, "testproject", "ca", "v1", "dig", true)
			Expect(err).To(BeNil())
		})
	})

	Describe("Get all sfc intents", func() {
		It("Parent Deployment Intent Group does not exist - return not found error", func() {
			ctx := context.Background()
			_, err := (*sfcClient).GetAllSfcIntents(ctx, "testproject", "ca", "v1", "dig")
			Expect(strings.Contains(err.Error(), "not found")).To(Equal(true))
		})
		It("Parent Deployment Intent Group does exist - No SFC Intents - should return empty list", func() {
			ctx := context.Background()
			// set up prerequisites
			_, err := (*projClient).CreateProject(ctx, proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ctx, ca, "testproject", false)
			Expect(err).To(BeNil())
			_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, dig, "testproject", "ca", "v1", true)
			Expect(err).To(BeNil())

			list, err := (*sfcClient).GetAllSfcIntents(ctx, "testproject", "ca", "v1", "dig")
			Expect(len(list)).To(Equal(0))
		})
		It("Parent Deployment Intent Group does exist - 2 SFC Intents created - should return list of len 2", func() {
			ctx := context.Background()
			// set up prerequisites
			_, err := (*projClient).CreateProject(ctx, proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ctx, ca, "testproject", false)
			Expect(err).To(BeNil())
			_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, dig, "testproject", "ca", "v1", true)
			Expect(err).To(BeNil())

			// test SFC intent creation - make 2 of them
			_, err = (*sfcClient).CreateSfcIntent(ctx, sfcIntent, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			sfcIntent.Metadata.Name = "2nd_name"
			_, err = (*sfcClient).CreateSfcIntent(ctx, sfcIntent, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())

			list, err := (*sfcClient).GetAllSfcIntents(ctx, "testproject", "ca", "v1", "dig")
			Expect(len(list)).To(Equal(2))

		})
		It("should return error for general db error", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("db Find error")
			_, err := (*sfcClient).GetAllSfcIntents(ctx, "testproject", "ca", "v1", "dig")
			Expect(strings.Contains(err.Error(), "db Find error")).To(Equal(true))
		})
		It("should return error for unmarshalling db error", func() {
			ctx := context.Background()
			// set up prerequisites
			_, err := (*projClient).CreateProject(ctx, proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ctx, ca, "testproject", false)
			Expect(err).To(BeNil())
			_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, dig, "testproject", "ca", "v1", true)
			Expect(err).To(BeNil())

			mdb.MarshalErr = pkgerrors.New("Unmarshalling bson")
			_, err = (*sfcClient).GetAllSfcIntents(ctx, "testproject", "ca", "v1", "dig")
			Expect(strings.Contains(err.Error(), "Unmarshalling bson")).To(Equal(true))
		})
	})

	Describe("Get sfc intent", func() {
		It("Parent Deployment Intent Group does not exist - return not found error", func() {
			ctx := context.Background()
			_, err := (*sfcClient).GetSfcIntent(ctx, "sfcIntentName", "testproject", "ca", "v1", "dig")
			Expect(strings.Contains(err.Error(), "not found")).To(Equal(true))
		})
		It("Successful get of sfcIntent", func() {
			ctx := context.Background()
			// set up prerequisites
			_, err := (*projClient).CreateProject(ctx, proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ctx, ca, "testproject", false)
			Expect(err).To(BeNil())
			_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, dig, "testproject", "ca", "v1", true)
			Expect(err).To(BeNil())

			// test SFC intent creation
			_, err = (*sfcClient).CreateSfcIntent(ctx, sfcIntent, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())

			_, err = (*sfcClient).GetSfcIntent(ctx, "sfcIntentName", "testproject", "ca", "v1", "dig")
			Expect(err).To(BeNil())
		})
		It("should return error for general db error", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("db Find error")
			_, err := (*sfcClient).GetSfcIntent(ctx, "sfcIntentName", "testproject", "ca", "v1", "dig")
			Expect(strings.Contains(err.Error(), "db Find error")).To(Equal(true))
		})
		It("should return error for unmarshalling db error", func() {
			ctx := context.Background()
			// set up prerequisites
			_, err := (*projClient).CreateProject(ctx, proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ctx, ca, "testproject", false)
			Expect(err).To(BeNil())
			_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, dig, "testproject", "ca", "v1", true)
			Expect(err).To(BeNil())

			_, err = (*sfcClient).CreateSfcIntent(ctx, sfcIntent, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			mdb.MarshalErr = pkgerrors.New("Unmarshalling bson")
			_, err = (*sfcClient).GetSfcIntent(ctx, "sfcIntentName", "testproject", "ca", "v1", "dig")
			Expect(strings.Contains(err.Error(), "Unmarshalling bson")).To(Equal(true))
		})
	})

	Describe("Delete SFC intent", func() {
		It("successful delete", func() {
			ctx := context.Background()
			// set up prerequisites
			_, err := (*projClient).CreateProject(ctx, proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ctx, ca, "testproject", false)
			Expect(err).To(BeNil())
			_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, dig, "testproject", "ca", "v1", true)
			Expect(err).To(BeNil())

			// test SFC intent creation
			_, err = (*sfcClient).CreateSfcIntent(ctx, sfcIntent, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())

			err = (*sfcClient).DeleteSfcIntent(ctx, "sfcIntentName", "testproject", "ca", "v1", "dig")
			Expect(err).To(BeNil())
		})
		It("should return not found error for non-existing record", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("db Remove resource not found")
			err := (*sfcClient).DeleteSfcIntent(ctx, "sfcIntentName", "testproject", "ca", "v1", "dig")
			Expect(strings.Contains(err.Error(), "db Remove resource not found")).To(Equal(true))
		})
		It("should return error for deleting parent without deleting child", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("db Remove parent child constraint")
			err := (*sfcClient).DeleteSfcIntent(ctx, "sfcIntentName", "testproject", "ca", "v1", "dig")
			Expect(strings.Contains(err.Error(), "db Remove parent child constraint")).To(Equal(true))
		})
		It("should return error for general db error", func() {
			ctx := context.Background()
			mdb.Err = pkgerrors.New("db Remove error")
			err := (*sfcClient).DeleteSfcIntent(ctx, "sfcIntentName", "testproject", "ca", "v1", "dig")
			Expect(strings.Contains(err.Error(), "db Remove error")).To(Equal(true))
		})
	})
})
