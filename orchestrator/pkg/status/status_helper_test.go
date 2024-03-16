package status_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/resourcestatus"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var file1, _ = ioutil.ReadFile("test/packetgen.yaml")
var packetgenContent = string(file1)

var file2, _ = ioutil.ReadFile("test/firewall.yaml")
var firewallContent = string(file2)

var file3, _ = ioutil.ReadFile("test/sink.yaml")
var sinkContent = string(file3)

var file4, _ = ioutil.ReadFile("test/packetgen-resourcebundlestate.json")
var packetgenStatusContent = string(file4)

type appContext struct {
	ac     appcontext.AppContext
	ctxVal interface{}
}

// populate an example deployment intent group app context
func createSampleVfwAppContext() (appContext, error) {
	ctx := context.Background()
	ac := appcontext.AppContext{}
	ctxVal, err := ac.InitAppContext()
	if err != nil {
		return appContext{}, pkgerrors.Wrap(err, "Error making appcontext")
	}

	// composite app create
	cah, err := ac.CreateCompositeApp(ctx)
	if err != nil {
		return appContext{}, pkgerrors.Wrap(err, "Error making composite app handle")
	}

	// add composite app meta structure
	caMeta := appcontext.CompositeAppMeta{Project: "testvfw", CompositeApp: "compositevfw", Version: "v1", Release: "fw0", DeploymentIntentGroup: "vfw_deployment_intent_group", Namespace: "default", Level: "0"}
	err = ac.AddCompositeAppMeta(ctx, caMeta)
	if err != nil {
		return appContext{}, pkgerrors.Wrap(err, "Error making ca meta data")
	}

	// add composite app status
	_, err = ac.AddLevelValue(ctx, cah, "status", appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Instantiated})

	// Add app instructions
	appDep, _ := json.Marshal(map[string]map[string]string{"appdependency": map[string]string{"packetgen": "go", "firewall": "go", "sink": "go"}})
	_, err = ac.AddInstruction(ctx, cah, "app", "dependency", string(appDep))
	if err != nil {
		return appContext{}, pkgerrors.Wrap(err, "Error making app dependency instruction")
	}
	appOrder, _ := json.Marshal(map[string][]string{"apporder": []string{"packetgen", "firewall", "sink"}})
	_, err = ac.AddInstruction(ctx, cah, "app", "order", string(appOrder))
	if err != nil {
		return appContext{}, pkgerrors.Wrap(err, "Error making app order instruction")
	}

	// add apps
	for _, app := range []string{"firewall", "packetgen", "sink"} {
		apph, err := ac.AddApp(ctx, cah, app)
		if err != nil {
			return appContext{}, pkgerrors.Wrapf(err, "Error making app: %v", app)
		}

		for _, cluster := range []string{"edge01", "edge02"} {
			clh, err := ac.AddCluster(ctx, apph, "vfw-cluster-provider+"+cluster)
			if err != nil {
				return appContext{}, pkgerrors.Wrapf(err, "Error making cluster: %v", cluster)
			}

			// Note - the same resourcebundlestate sample for just the 'packetgen' app is saved
			// for both clusters.
			if app == "packetgen" {
				_, err = ac.AddLevelValue(ctx, clh, "status", packetgenStatusContent)
				if err != nil {
					return appContext{}, pkgerrors.Wrapf(err, "Error making clustesr status: %v", cluster)
				}
			}

			// Add resource instructions
			resDep, _ := json.Marshal(map[string]map[string]string{"resdependency": map[string]string{"fw0-" + app + "+Deployment": "go"}})
			_, err = ac.AddInstruction(ctx, clh, "resource", "dependency", string(resDep))
			if err != nil {
				return appContext{}, pkgerrors.Wrap(err, "Error making resource dependency instruction")
			}
			resOrder, _ := json.Marshal(map[string][]string{"resorder": []string{"fw0-" + app + "+Deployment"}})
			_, err = ac.AddInstruction(ctx, clh, "resource", "order", string(resOrder))
			if err != nil {
				return appContext{}, pkgerrors.Wrap(err, "Error making resource order instruction")
			}
			// Add a reference
			_, err = ac.AddLevelValue(ctx, clh, "reference", ctxVal)
			// Add resources
			var rh interface{}
			switch app {
			case "firewall":
				rh, err = ac.AddResource(ctx, clh, "fw0-"+app+"+Deployment", firewallContent)
			case "packetgen":
				rh, err = ac.AddResource(ctx, clh, "fw0-"+app+"+Deployment", packetgenContent)
			case "sink":
				rh, err = ac.AddResource(ctx, clh, "fw0-"+app+"+Deployment", sinkContent)
			}
			if err != nil {
				return appContext{}, pkgerrors.Wrap(err, "Error making resource order instruction")
			}
			// Add a reference
			_, err = ac.AddLevelValue(ctx, rh, "reference", ctxVal)
			// Add resource status
			_, err = ac.AddLevelValue(ctx, rh, "status", resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Applied})
		}
	}
	return appContext{ac: ac, ctxVal: ctxVal}, nil
}

var _ = Describe("StatusHelper", func() {
	var (
		cdb *contextdb.MockConDb

		stateInfoInstantiated state.StateInfo

		actionCreated      state.ActionEntry
		actionApproved     state.ActionEntry
		actionInstantiated state.ActionEntry

		expectedAppsList              status.AppsListResult
		expectedClustersByApp         status.ClustersByAppResult
		expectedRsyncResourcesByApp   status.ResourcesByAppResult
		expectedClusterResourcesByApp status.ResourcesByAppResult
		expectedClusterStatusResult   status.StatusResult
		expectedRsyncStatusResult     status.StatusResult
	)

	BeforeEach(func() {
		cdb = new(contextdb.MockConDb)
		contextdb.Db = cdb
		appContext, err := createSampleVfwAppContext()
		if err != nil {
			fmt.Printf("make app context  error: %v\n", err)
		}

		actionCreated = state.ActionEntry{
			State: state.StateEnum.Created,
		}
		actionApproved = state.ActionEntry{
			State: state.StateEnum.Approved,
		}
		actionInstantiated = state.ActionEntry{
			State:     state.StateEnum.Instantiated,
			ContextId: fmt.Sprintf("%v", appContext.ctxVal),
		}
		stateInfoInstantiated.StatusContextId = fmt.Sprintf("%v", appContext.ctxVal)
		stateInfoInstantiated.Actions = make([]state.ActionEntry, 0)
		stateInfoInstantiated.Actions = append(stateInfoInstantiated.Actions, actionCreated)
		stateInfoInstantiated.Actions = append(stateInfoInstantiated.Actions, actionApproved)
		stateInfoInstantiated.Actions = append(stateInfoInstantiated.Actions, actionInstantiated)

		expectedAppsList = status.AppsListResult{
			Apps: []string{"firewall", "packetgen", "sink"},
		}
		expectedClustersByApp = status.ClustersByAppResult{
			ClustersByApp: []status.ClustersByAppEntry{
				{
					App: "sink",
					Clusters: []status.ClusterEntry{
						{ClusterProvider: "vfw-cluster-provider", Cluster: "edge01"},
						{ClusterProvider: "vfw-cluster-provider", Cluster: "edge02"},
					},
				},
			},
		}
		expectedRsyncResourcesByApp = status.ResourcesByAppResult{
			ResourcesByApp: []status.ResourcesByAppEntry{
				{
					App:             "firewall",
					ClusterProvider: "vfw-cluster-provider",
					Cluster:         "edge01",
					Resources: []status.ResourceEntry{
						{
							Name: "fw0-firewall",
							Gvk:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
						},
					},
				},
				{
					App:             "firewall",
					ClusterProvider: "vfw-cluster-provider",
					Cluster:         "edge02",
					Resources: []status.ResourceEntry{
						{
							Name: "fw0-firewall",
							Gvk:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
						},
					},
				},
				{
					App:             "packetgen",
					ClusterProvider: "vfw-cluster-provider",
					Cluster:         "edge01",
					Resources: []status.ResourceEntry{
						{
							Name: "fw0-packetgen",
							Gvk:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
						},
					},
				},
				{
					App:             "packetgen",
					ClusterProvider: "vfw-cluster-provider",
					Cluster:         "edge02",
					Resources: []status.ResourceEntry{
						{
							Name: "fw0-packetgen",
							Gvk:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
						},
					},
				},
				{
					App:             "sink",
					ClusterProvider: "vfw-cluster-provider",
					Cluster:         "edge01",
					Resources: []status.ResourceEntry{
						{
							Name: "fw0-sink",
							Gvk:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
						},
					},
				},
				{
					App:             "sink",
					ClusterProvider: "vfw-cluster-provider",
					Cluster:         "edge02",
					Resources: []status.ResourceEntry{
						{
							Name: "fw0-sink",
							Gvk:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
						},
					},
				},
			},
		}

		expectedClusterResourcesByApp = status.ResourcesByAppResult{
			ResourcesByApp: []status.ResourcesByAppEntry{
				{
					App:             "packetgen",
					ClusterProvider: "vfw-cluster-provider",
					Cluster:         "edge01",
					Resources: []status.ResourceEntry{
						{
							Name: "fw0-packetgen",
							Gvk:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
						},
						{
							Name: "fw0-packetgen-67d8fb7b68-8g824",
							Gvk:  schema.GroupVersionKind{Version: "v1", Kind: "Pod"},
						},
						{
							Name: "packetgen-service",
							Gvk:  schema.GroupVersionKind{Version: "v1", Kind: "Service"},
						},
					},
				},
				{
					App:             "packetgen",
					ClusterProvider: "vfw-cluster-provider",
					Cluster:         "edge02",
					Resources: []status.ResourceEntry{
						{
							Name: "fw0-packetgen",
							Gvk:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
						},
						{
							Name: "fw0-packetgen-67d8fb7b68-8g824",
							Gvk:  schema.GroupVersionKind{Version: "v1", Kind: "Pod"},
						},
						{
							Name: "packetgen-service",
							Gvk:  schema.GroupVersionKind{Version: "v1", Kind: "Service"},
						},
					},
				},
			},
		}

		expectedClusterStatusResult = status.StatusResult{
			State:         stateInfoInstantiated,
			Status:        appcontext.AppContextStatusEnum.Instantiated,
			ReadyStatus:   "NotReady",
			ClusterStatus: map[string]int{"Ready": 4, "NotReady": 2},
			RsyncStatus:   map[string]int{},
			Apps: []status.AppStatus{
				{
					Name: "packetgen",
					Clusters: []status.ClusterStatus{
						{
							ClusterProvider: "vfw-cluster-provider",
							Cluster:         "edge01",
							ReadyStatus:     "Unknown",
							Resources: []status.ResourceStatus{
								{
									Name:          "fw0-packetgen",
									Gvk:           schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
									ClusterStatus: "Ready",
								},
								{
									Name:          "fw0-packetgen-67d8fb7b68-8g824",
									Gvk:           schema.GroupVersionKind{Version: "v1", Kind: "Pod"},
									ClusterStatus: "Ready",
								},
								{
									Name:          "packetgen-service",
									Gvk:           schema.GroupVersionKind{Version: "v1", Kind: "Service"},
									ClusterStatus: "NotReady",
								},
							},
						},
						{
							ClusterProvider: "vfw-cluster-provider",
							Cluster:         "edge02",
							ReadyStatus:     "Unknown",
							Resources: []status.ResourceStatus{
								{
									Name:          "fw0-packetgen",
									Gvk:           schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
									ClusterStatus: "Ready",
								},
								{
									Name:          "fw0-packetgen-67d8fb7b68-8g824",
									Gvk:           schema.GroupVersionKind{Version: "v1", Kind: "Pod"},
									ClusterStatus: "Ready",
								},
								{
									Name:          "packetgen-service",
									Gvk:           schema.GroupVersionKind{Version: "v1", Kind: "Service"},
									ClusterStatus: "NotReady",
								},
							},
						},
					},
				},
			},
		}

		expectedRsyncStatusResult = status.StatusResult{
			State:         stateInfoInstantiated,
			Status:        appcontext.AppContextStatusEnum.Instantiated,
			ClusterStatus: map[string]int{},
			RsyncStatus:   map[string]int{"Applied": 2},
			Apps: []status.AppStatus{
				{
					Name: "packetgen",
					Clusters: []status.ClusterStatus{
						{
							ClusterProvider: "vfw-cluster-provider",
							Cluster:         "edge01",
							ReadyStatus:     "Unknown",
							Resources: []status.ResourceStatus{
								{
									Name:        "fw0-packetgen",
									Gvk:         schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
									RsyncStatus: "Applied",
								},
							},
						},
						{
							ClusterProvider: "vfw-cluster-provider",
							Cluster:         "edge02",
							ReadyStatus:     "Unknown",
							Resources: []status.ResourceStatus{
								{
									Name:        "fw0-packetgen",
									Gvk:         schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
									RsyncStatus: "Applied",
								},
							},
						},
					},
				},
			},
		}
	})

	It("get apps list of instantiated vfw", func() {
		result, err := status.PrepareAppsListStatusResult(context.Background(), stateInfoInstantiated, "")
		Expect(err).To(BeNil())
		Expect(result).Should(Equal(expectedAppsList))
	})

	It("get clusters for sink of instantiated vfw", func() {
		result, err := status.PrepareClustersByAppStatusResult(context.Background(), stateInfoInstantiated, "", []string{"sink"})
		Expect(err).To(BeNil())
		Expect(result).Should(Equal(expectedClustersByApp))
	})

	It("get rsync resources for packetgen of instantiated vfw", func() {
		result, err := status.PrepareResourcesByAppStatusResult(context.Background(), stateInfoInstantiated, "", "rsync", []string{}, []string{})
		Expect(err).To(BeNil())
		Expect(result).Should(Equal(expectedRsyncResourcesByApp))
	})

	It("get cluster resources for packetgen of instantiated vfw", func() {
		result, err := status.PrepareResourcesByAppStatusResult(context.Background(), stateInfoInstantiated, "", "cluster", []string{"packetgen"}, []string{})
		Expect(err).To(BeNil())
		Expect(result).Should(Equal(expectedClusterResourcesByApp))
	})

	It("get cluster status for packetgen app of instantiated vfw", func() {
		result, err := status.PrepareStatusResult(context.Background(), stateInfoInstantiated, "", "cluster", "all", []string{"packetgen"}, []string{}, []string{})
		Expect(err).To(BeNil())
		Expect(result).Should(Equal(expectedClusterStatusResult))
	})

	It("get rsync status for packetgen app of instantiated vfw", func() {
		result, err := status.PrepareStatusResult(context.Background(), stateInfoInstantiated, "", "rsync", "all", []string{"packetgen"}, []string{}, []string{})
		Expect(err).To(BeNil())
		Expect(result).Should(Equal(expectedRsyncStatusResult))
	})
})
