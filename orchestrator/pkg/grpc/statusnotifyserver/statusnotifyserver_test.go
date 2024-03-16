// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2022 Intel Corporation

package statusnotifyserver

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/statusnotify"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
)

type mockHelpers struct{}

func (m mockHelpers) GetAppContextId(ctx context.Context, reg *pb.StatusRegistration) (string, error) {
	return "", nil
}

func (m mockHelpers) StatusQuery(ctx context.Context, reg *pb.StatusRegistration, qInstance, qType, qOutput string, qApps, qClusters, qResources []string) status.StatusResult {
	return status.StatusResult{}
}

func (m mockHelpers) PrepareStatusNotification(reg *pb.StatusRegistration, statusResult status.StatusResult) *pb.StatusNotification {
	n := new(pb.StatusNotification)

	return n
}

// client1 does:
//   status = ready
//   output = summary
//   no filters
func makeClient1Registration() *pb.StatusRegistration {
	var reg pb.StatusRegistration

	reg.StatusType = pb.StatusValue_READY
	reg.Output = pb.OutputType_SUMMARY

	return &reg
}

// client2 does:
//   status = deployed
//   output = summary
//   no filters
func makeClient2Registration() *pb.StatusRegistration {
	var reg pb.StatusRegistration

	reg.StatusType = pb.StatusValue_DEPLOYED
	reg.Output = pb.OutputType_SUMMARY

	return &reg
}

func addSummaryClients(ns *StatusNotifyServer) {
	// make appcontext1 data
	acInfo1 := appContextInfo{
		statusClientIDs: make(map[string]struct{}),
		queryFilters:    make(map[string]filters),
	}
	rFilter := filters{
		qOutputSummary: true,
		apps:           make(map[string]struct{}),
		clusters:       make(map[string]struct{}),
		resources:      make(map[string]struct{}),
	}
	dFilter := filters{
		qOutputSummary: true,
		apps:           make(map[string]struct{}),
		clusters:       make(map[string]struct{}),
		resources:      make(map[string]struct{}),
	}
	acInfo1.queryFilters["ready"] = rFilter
	acInfo1.queryFilters["deployed"] = dFilter
	acInfo1.statusClientIDs["client1"] = struct{}{}
	acInfo1.statusClientIDs["client2"] = struct{}{}
	ns.appContexts["appcontext1"] = acInfo1

	streamInfo1 := streamInfo{
		appContextID: "appcontext1",
		reg:          makeClient1Registration(),
	}
	streamInfo2 := streamInfo{
		appContextID: "appcontext1",
		reg:          makeClient2Registration(),
	}

	ns.statusClients["client1"] = streamInfo1
	ns.statusClients["client2"] = streamInfo2
}

var _ = Describe("StatusNotifyServer", func() {
	var (
		ns *StatusNotifyServer
	)

	BeforeEach(func() {
		ns = NewStatusNotifyServer("mockStatus", mockHelpers{})
		addSummaryClients(ns)
	})

	It("get apps list of instantiated vfw", func() {
		updateQueryFilters("client1", false)
		Expect(len(ns.appContexts["appcontext1"].queryFilters["ready"].apps)).Should(Equal(0))
		Expect(len(ns.appContexts["appcontext1"].queryFilters["ready"].clusters)).Should(Equal(0))
		Expect(len(ns.appContexts["appcontext1"].queryFilters["ready"].resources)).Should(Equal(0))
		updateQueryFilters("client2", false)
		Expect(len(ns.appContexts["appcontext1"].queryFilters["ready"].apps)).Should(Equal(0))
		Expect(len(ns.appContexts["appcontext1"].queryFilters["ready"].clusters)).Should(Equal(0))
		Expect(len(ns.appContexts["appcontext1"].queryFilters["ready"].resources)).Should(Equal(0))

		apps := make(map[string]struct{})
		apps["app1"] = struct{}{}
		clusters := make(map[string]struct{})
		clusters["cluster1"] = struct{}{}
		queryNeeded, qOutput, qApps, qClusters, qResources := queryNeeded("ready", apps, clusters, ns.appContexts["appcontext1"])
		Expect(len(qApps)).Should(Equal(0))
		Expect(len(qClusters)).Should(Equal(0))
		Expect(len(qResources)).Should(Equal(0))
		Expect(queryNeeded).Should(Equal(true))
		Expect(qOutput).Should(Equal("summary"))
	})

	/*
		Expect(err).To(BeNil())
	*/
})
