// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module

import (
	"context"
	"fmt"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	gpic "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/gpic"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/utils/helm"
	"strings"
	"testing"
)

func init() {
	edb := new(contextdb.MockConDb)
	edb.Err = nil
	contextdb.Db = edb
}

func TestHookInstructionAppContext(t *testing.T) {
	ctx := context.Background()

	context := appcontext.AppContext{}
	ctxval, err := context.InitAppContext()
	if err != nil {
		t.Fatalf("Got unexpected error message %s", err)
	}
	compositeHandle, err := context.CreateCompositeApp(ctx)
	if err != nil {
		t.Fatalf("Got unexpected error message %s", err)
	}
	cca := contextForCompositeApp{context: context, ctxval: ctxval, compositeAppHandle: compositeHandle}
	// SetUp clusters for the test
	mc := []gpic.ClusterGroup{{
		GroupNumber: "1",
		Clusters: []gpic.ClusterWithName{
			{ProviderName: "provider1", ClusterName: "cluster1"},
		},
	}}
	opt := []gpic.ClusterGroup{{
		GroupNumber: "1",
		Clusters: []gpic.ClusterWithName{
			{ProviderName: "provider1", ClusterName: "cluster2"},
		},
	}}
	listOfClusters := gpic.ClusterList{MandatoryClusters: mc, OptionalClusters: opt}

	testCases := []struct {
		chartDir         string
		label            string
		expectedError    string
		expectedValue    string
		mandatoryCluster string
		optionalCluster  string
		hookString       string
	}{
		{
			label:            "Test Instantiation AppContext Creation without Hooks Mandatory cluster",
			chartDir:         "../../mock_files/mock_charts/testchart1",
			mandatoryCluster: "provider1+cluster1",
			optionalCluster:  "",
			expectedValue:    "",
			expectedError:    "Key doesn't exist",
		},
		{
			label:            "Test Instantiation AppContext Creation without Hooks Optional cluster",
			chartDir:         "../../mock_files/mock_charts/testchart1",
			mandatoryCluster: "",
			optionalCluster:  "provider1+cluster2",
			expectedValue:    "",
			expectedError:    "Key doesn't exist",
		},
		{
			label:            "Test Instantiation AppContext Creation with Hooks Mandatory cluster",
			chartDir:         "../../mock_files/mock_charts/testchart2",
			mandatoryCluster: "provider1+cluster1",
			optionalCluster:  "",
			expectedError:    "",
			expectedValue:    "{\"resdependency\":{\"post-install\":[\"pod2+Pod\"],\"pre-install\":[\"pod1+Pod\"]}}",
		},
		{
			label:            "Test Instantiation AppContext Creation with Hooks Optional cluster",
			chartDir:         "../../mock_files/mock_charts/testchart2",
			mandatoryCluster: "",
			optionalCluster:  "provider1+cluster2",
			expectedError:    "",
			expectedValue:    "{\"resdependency\":{\"post-install\":[\"pod2+Pod\"],\"pre-install\":[\"pod1+Pod\"]}}",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			tc := helm.NewTemplateClient("1.12.3", "testnamespace", "testreleasename", "manifest.yaml")
			out, hooks, err := tc.GenerateKubernetesArtifacts(testCase.chartDir, []string{}, []string{})
			if err != nil {
				t.Fatalf("Got unexpected error message %s", err)
			}
			ah := AppHandler{appName: "testApp", clusters: listOfClusters, namespace: "testNamespace", ht: out, hk: hooks}
			// Test function
			err = ah.addAppToAppContext(ctx, cca)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Got unexpected error message %s", err)
				}
			}
			var cluster string
			if testCase.mandatoryCluster != "" {
				cluster = testCase.mandatoryCluster
			} else {
				cluster = testCase.optionalCluster
			}
			// Check for hook in the mandatory or optional cluster
			v, err := context.GetResourceInstruction(ctx, ah.appName, cluster, "dependency")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Got an error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Got unexpected error message %s", err)
				}
			} else {
				str := fmt.Sprintf("%v", v)
				if strings.Contains(str, testCase.expectedValue) == false {
					t.Fatalf("Unexpected helm hooks found Expectd: %s Got: %s ", testCase.expectedValue, str)
				}
			}
		})
	}
}
