// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package action

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"errors"

	pkgerrors "github.com/pkg/errors"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	//orchUtils "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/utils"
	orchModuleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"

	clmcontrollerpb "gitlab.com/project-emco/core/emco-base/src/clm/pkg/grpc/controller-eventchannel"
	//migrationModel "gitlab.com/project-emco/core/emco-base/src/migration/pkg/model"
	migrationModuleLib "gitlab.com/project-emco/core/emco-base/src/migration/pkg/module"
	hpaModuleLib "gitlab.com/project-emco/core/emco-base/src/hpa-plc/pkg/module"
	// intentRs "gitlab.com/project-emco/core/emco-base/src/migration/pkg/resources"
	//migrationUtils "gitlab.com/project-emco/core/emco-base/src/migration/pkg/utils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"


	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	//clmModuleLib "gitlab.com/project-emco/core/emco-base/src/clm/pkg/module"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
        "encoding/base64"
        "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"
	//rcontext "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/context"
	hpaModel "gitlab.com/project-emco/core/emco-base/src/hpa-plc/pkg/model"
//	v1alpha1 "k8s.io/metrics/pkg/apis/metrics/v1alpha1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const StatusAppContextIDKey string = "statusappctxid"

type AppMigrationIntent struct {
	AppContextID string
	Project string
	CompositeApp string
	CompositeAppVersion string
	App string
	DeploymentIntentGroup string
	HpaIntent string
	MigrationIntent string
	MigrationAppIntent string
	Migration bool
	Priority int64
	CPU int64
	MEMORY int64
	CPUUtilization int64
	Dispersion float64
}

// MigrationApps .. Filter clusters based on hpa-intents attached to the AppContext ID
func MigrationApps(ctx context.Context, appContextID string, failedapp string) error {
	var ac appcontext.AppContext
	log.Warn("MigrationApps .. start", log.Fields{"appContextID": appContextID})
	_, err := ac.LoadAppContext(ctx, appContextID)
	if err != nil {
		log.Error("MigrationApps .. Error getting AppContext", log.Fields{"appContextID": appContextID})
		return pkgerrors.Wrapf(err, "MigrationApps .. Error getting AppContext with Id: %v", appContextID)
	}

	caMeta, err := ac.GetCompositeAppMeta(ctx)

	if err != nil {
		log.Error("MigrationApps .. Error getting metadata for AppContext", log.Fields{"appContextID": appContextID})
		return pkgerrors.Wrapf(err, "MigrationApps .. Error getting metadata for AppContext with Id: %v", appContextID)
	}

	project := caMeta.Project
	compositeApp := caMeta.CompositeApp
	compositeAppVersion := caMeta.Version
	deploymentIntentGroup := caMeta.DeploymentIntentGroup

	log.Info("MigrationApps .. AppContext details", log.Fields{"project": project, "compositeApp": compositeApp, "deploymentGroup": deploymentIntentGroup})

	// Get all apps in this composite app
	apps, err := orchModuleLib.NewAppClient().GetApps(ctx, project, compositeApp, compositeAppVersion)
	if err != nil {
		log.Error("MigrationApps .. Not finding the compositeApp attached apps", log.Fields{"appContextID": appContextID, "compositeApp": compositeApp})
		return pkgerrors.Wrapf(err, "MigrationApps .. Not finding the compositeApp[%s] attached apps", compositeApp)
	}

	if len(apps) == 0 {
		return pkgerrors.Errorf(
			"Apps not found for composite app '%s' with version '%s'",
			compositeApp,
			compositeAppVersion,
		)
	}

	allAppNames := make([]string, 0)
	for _, a := range apps {
		allAppNames = append(allAppNames, a.Metadata.Name)
	}
	log.Info("MigrationApps .. Applications attached to compositeApp",
		log.Fields{"appContextID": appContextID, "project": project, "compositeApp": compositeApp, "deploymentGroup": deploymentIntentGroup, "app-names": allAppNames})


	// Dump group-clusters map
	for index, eachApp := range allAppNames {
		grpMap, _ := ac.GetClusterGroupMap(ctx, eachApp)
		log.Warn("MigrationApps .. ClusterGroupMap dump before invoking HPA Placement filtering",
			log.Fields{"index": index, "appContextID": appContextID, "appName": eachApp, "group-map_size": len(grpMap), "groupMap": grpMap})
	}


	k := "/context/"

	allkeys, _ := contextdb.Db.GetAllKeys(ctx, k)

	//compositeapp 내에서 현재 배포 불가능한 app에 대해 그 app을 배포해야 하는 클러스터를 불러오기 위한 코드
	getappcontextfromid, _ := state.GetAppContextFromId(ctx, appContextID)
	//fmt.Println("\n getappcontextfromid: ",getappcontextfromid)

	getcluster, _ := getappcontextfromid.GetClusterGroupMap(ctx, failedapp)
	cluster := getcluster["1"][0]
	fmt.Println("getcluster: ", getcluster)
	fmt.Println("cluster: ", cluster,"\n")

	fmt.Println("\n\n")
	allappwithmigrationintents := getmigappswithmigrationintents(ctx, allkeys, cluster)
	fmt.Println("allappwithmigrationintents: ",allappwithmigrationintents,"\n")

	allappwithmigrationintents = calculateapputilizations(ctx, cluster, allappwithmigrationintents)

	allappwithmigrationintents = calculateappdispersion(ctx, cluster, allappwithmigrationintents)
	fmt.Println("\n\n")
	fmt.Println("after dispersion: ",allappwithmigrationintents)
	fmt.Println("\n\n")

	migrationPolicies, _:= migrationModuleLib.NewMigrationClient().GetAllPolicies(ctx, project)
	for _, migrationPolicy := range migrationPolicies {
		fmt.Println("\n\n")
		fmt.Println("migrationPolicy: ",migrationPolicy)
		fmt.Println("\n\n")

		orders := make(map[string]int64)
		if migrationPolicy.Spec.Considerations.CPUUtilization.Active == true {
			orders["cpuUtilization"] = migrationPolicy.Spec.Considerations.CPUUtilization.Order
		}
		if migrationPolicy.Spec.Considerations.Dispersion.Active == true {
			orders["dispersion"] = migrationPolicy.Spec.Considerations.Dispersion.Order
		}

		//fmt.Println(orders)

		keys := make([]string, 0, len(orders))
		for key := range orders {
			keys = append(keys, key)
		}

		sort.Slice(keys, func(i, j int) bool {
			return orders[keys[i]] < orders[keys[j]]
		})

		fmt.Println(keys)

		if keys[0] == "cpuUtilization" {

			sort.SliceStable(allappwithmigrationintents, func(i, j int) bool {
				if allappwithmigrationintents[i].Priority != allappwithmigrationintents[j].Priority {
					return allappwithmigrationintents[i].Priority < allappwithmigrationintents[j].Priority
				}
				if allappwithmigrationintents[i].CPUUtilization != allappwithmigrationintents[j].CPUUtilization {
					return allappwithmigrationintents[i].CPUUtilization < allappwithmigrationintents[j].CPUUtilization
				}
				return allappwithmigrationintents[i].Dispersion > allappwithmigrationintents[j].Dispersion
			})
		} else {
			sort.SliceStable(allappwithmigrationintents, func(i, j int) bool {
				if allappwithmigrationintents[i].Priority != allappwithmigrationintents[j].Priority {
					return allappwithmigrationintents[i].Priority < allappwithmigrationintents[j].Priority
				}
				if allappwithmigrationintents[i].Dispersion != allappwithmigrationintents[j].Dispersion {
					return allappwithmigrationintents[i].Dispersion > allappwithmigrationintents[j].Dispersion
				}
				return allappwithmigrationintents[i].CPUUtilization < allappwithmigrationintents[j].CPUUtilization
			})
		}

		fmt.Println("after applying priority")
		fmt.Println(allappwithmigrationintents)


	}

	// 배포해야되는 app이 필요한 리소스를 만들 수 있을 만큼 이전할 app을 선택하는 코드
	var selectedApps []AppMigrationIntent
	var totalCPU, totalMEMORY int64

	compositeappmeta, _ := getappcontextfromid.GetCompositeAppMeta(ctx)

	project = compositeappmeta.Project
	compositeApp = compositeappmeta.CompositeApp
	compositeAppVersion = compositeappmeta.Version
	deploymentIntentGroup = compositeappmeta.DeploymentIntentGroup
	//fmt.Println("\n compositeappmeta: ",compositeapp)

	requiredcpu, requiredmemory := getrequiredresforapp(ctx, failedapp, project, compositeApp, compositeAppVersion, deploymentIntentGroup)

	for i, eachappwithmigrationintent := range allappwithmigrationintents {
		selectedApps = append(selectedApps, eachappwithmigrationintent)
		totalCPU += eachappwithmigrationintent.CPU
		totalMEMORY += eachappwithmigrationintent.MEMORY

		// CPU와 MEMORY의 합이 배포되어야 하는 app의 요구사항을 충족하면 선택 종료
		if totalCPU >= requiredcpu && totalMEMORY >= requiredmemory {
			break
		}

		// 모든 app를 선택해도 조건을 만족하지 않을 경우
		if i == len(allappwithmigrationintents)-1 {
			fmt.Println("Unable to deploy even after migrating all migratable apps")
		}
	}

	// 선택된 app  출력
	fmt.Println("\nSelected App: ",selectedApps,"\n")
	if len(selectedApps) == 0 {
		log.Error("MigrationApps .. Error selecting Apps .. None of apps can be migrated", log.Fields{"appContextID": appContextID})
		//return pkgerrors.Wrapf(err, "MigrationApps .. Error selecting Apps .. None of apps can be migrated for Id: %v", appContextID)
		return errors.New("MigrationApps .. Error selecting Apps .. None of apps can be migrated")
	}

	// selectedApps 정렬
	sort.Slice(selectedApps, func(i, j int) bool {
		// CPU Request에 따라 내림차순 정렬
		if selectedApps[i].CPU > selectedApps[j].CPU {
			return true
		} else if selectedApps[i].CPU == selectedApps[j].CPU {
		// CPU 값이 같을 경우 Memory에 따라 내림차순 정렬
			return selectedApps[i].MEMORY > selectedApps[j].MEMORY
		}
		return false
	})

	var selectedAppsCPU, selectedAppsMemory int64

	var index int
	// requiredcpu, requiredmemory를 충족할 때까지 선택된 Apps의 CPU와 MEMORY 누적
	for i, selectedApp := range selectedApps {
		selectedAppsCPU += selectedApp.CPU
		selectedAppsMemory += selectedApp.MEMORY

		// 충족 시 종료
		if selectedAppsCPU >= requiredcpu && selectedAppsMemory >= requiredmemory {
			index = i
			break
		}
	}

	selectedApps = selectedApps[:index+1]
	fmt.Println("\nSelected App after minimize: ",selectedApps,"\n")

	var selectedCA []string

	MigrationClient := migrationModuleLib.NewMigrationClient()
	for _, selectedApp := range selectedApps {

		caExists := false
		for _, v := range selectedCA {
			if v == selectedApp.CompositeApp {
				caExists = true
				break
			}
		}
		if !caExists {
			migrationIntent, _, _ := MigrationClient.GetIntent(ctx, selectedApp.MigrationIntent, selectedApp.Project, selectedApp.CompositeApp, selectedApp.CompositeAppVersion, selectedApp.DeploymentIntentGroup)
			migrationIntent.Status.Selected = true
			MigrationClient.AddIntent(ctx, migrationIntent, selectedApp.Project, selectedApp.CompositeApp, selectedApp.CompositeAppVersion, selectedApp.DeploymentIntentGroup, true)
			selectedCA = append(selectedCA, selectedApp.CompositeApp)
		}

		MigrationAppIntent, _, _ := MigrationClient.GetAppIntent(ctx, selectedApp.MigrationAppIntent, selectedApp.Project, selectedApp.CompositeApp, selectedApp.CompositeAppVersion, selectedApp.DeploymentIntentGroup, selectedApp.MigrationIntent)
		MigrationAppIntent.Status.SelectedApp = true
		MigrationClient.AddAppIntent(ctx, MigrationAppIntent, selectedApp.Project, selectedApp.CompositeApp, selectedApp.CompositeAppVersion, selectedApp.DeploymentIntentGroup, selectedApp.MigrationIntent, true)
	}


	return nil
}

type AppContextReference struct {
        acID string
        ac   appcontext.AppContext
}

func (a *AppContextReference) GetStatusAppContext(ctx context.Context, key string) (string, error) {
        h, err := a.ac.GetCompositeAppHandle(ctx)
        if err != nil {
                log.Error("Error GetAppContextFlag", log.Fields{"err": err})
                return "", err
        }
        sh, err := a.ac.GetLevelHandle(ctx, h, key)
        if sh != nil {
                if v, err := a.ac.GetValue(ctx, sh); err == nil {
                        return v.(string), nil
                }
        }
        return "", err
}


func NewAppContextReference(ctx context.Context, acID string) (AppContextReference, error) {
        ac := appcontext.AppContext{}
        if len(acID) == 0 {
                log.Error("Error loading AppContext - appContexID is nil", log.Fields{})
                return AppContextReference{}, pkgerrors.Errorf("appContexID is nil")
        }
        _, err := ac.LoadAppContext(ctx, acID)
        if err != nil {
                log.Error("Error loading AppContext", log.Fields{"err": err, "acID": acID})
                return AppContextReference{}, err
        }
        return AppContextReference{ac: ac, acID: acID}, nil
}

func (a *AppContextReference) GetAppContextHandle() appcontext.AppContext {
        return a.ac
}



// GetKubeConfig uses the connectivity client to get the kubeconfig based on the name
// of the clustername.
var GetKubeConfig = func(ctx context.Context, clustername string, level string, namespace string) ([]byte, error) {
        if !strings.Contains(clustername, "+") {
                return nil, pkgerrors.New("Not a valid cluster name")
        }
        strs := strings.Split(clustername, "+")
        if len(strs) != 2 {
                return nil, pkgerrors.New("Not a valid cluster name")
        }

        ccc := db.NewCloudConfigClient()

        log.Info("Querying CloudConfig", log.Fields{"strs": strs, "level": level, "namespace": namespace})
        cconfig, err := ccc.GetCloudConfig(ctx, strs[0], strs[1], level, namespace)
        if err != nil {
                return nil, pkgerrors.Wrap(err, "Get kubeconfig failed")
        }
        log.Info("Successfully looked up CloudConfig", log.Fields{".Provider": cconfig.Provider, ".Cluster": cconfig.Cluster, ".Level": cconfig.Level, ".Namespace": cconfig.Namespace})

        dec, err := base64.StdEncoding.DecodeString(cconfig.Config)
        if err != nil {
                return nil, err
        }
        return dec, nil
}



func getmigappswithmigrationintents(ctx context.Context, allkeys []string, cluster string)(allappwithmigrationintents []AppMigrationIntent) {

	clusterappcontext := GetClusterApp(ctx, allkeys, cluster)

	fmt.Println("\n\n")
	fmt.Println("clusterappcontext: ",clusterappcontext)
	fmt.Println("\n\n")

	for _, appcontext := range clusterappcontext{

		getappcontextfromid, _ := state.GetAppContextFromId(ctx, appcontext)
		compositeappmeta, _ := getappcontextfromid.GetCompositeAppMeta(ctx)

		fmt.Println("internal/action/action.go:getmigappswithmigrationintents()")
		fmt.Println("getappcontextfromid",getappcontextfromid)
		fmt.Println("compositeappmeta",compositeappmeta)

		project := compositeappmeta.Project
		compositeApp := compositeappmeta.CompositeApp
		compositeAppVersion := compositeappmeta.Version
		deploymentIntentGroup := compositeappmeta.DeploymentIntentGroup


		apps, err := orchModuleLib.NewAppClient().GetApps(ctx, project, compositeApp, compositeAppVersion)
		if err != nil {
			log.Error(" getmigappswithmigrationintents.. Not finding the compositeApp attached apps", log.Fields{"compositeApp": compositeApp})
		}

		allAppNames := make([]string, 0)

		for _, a := range apps {
			allAppNames = append(allAppNames, a.Metadata.Name)
		}


		migrationIntents, err := migrationModuleLib.NewMigrationClient().GetAllIntents(ctx, project, compositeApp, compositeAppVersion, deploymentIntentGroup)
		fmt.Println("migrationIntents: ",migrationIntents)



		for index, migrationIntent := range migrationIntents {
			log.Info("Migrations .. migrationIntents details => ", log.Fields{
			"intent-index":            index,
			"migration-intent":        migrationIntent,
			"project":                 project,
			"composite-app":           compositeApp,
			"composite-app-version":   compositeAppVersion,
			"deployment-intent-group": deploymentIntentGroup,
			"migration-intent-name":   migrationIntent.MetaData.Name,
			})

			migrationAppIntents, _ := migrationModuleLib.NewMigrationClient().GetAllAppIntents(ctx, project, compositeApp, compositeAppVersion, deploymentIntentGroup, migrationIntent.MetaData.Name)

			for _, migrationAppIntent := range migrationAppIntents {

				//fmt.Println("cluster: ",cluster)
				gmap, err := getappcontextfromid.GetClusterGroupMap(ctx, migrationAppIntent.Spec.App)
				fmt.Println("gmap: ", gmap)
				var getcluster string
				for _, cl := range gmap {
					//fmt.Println("cl: ",cl)
					if len(cl) != 0 {
						getcluster = cl[0]
					}
				}

				migrationAppIntent.Status.DeployedCluster = getcluster
				migrationModuleLib.NewMigrationClient().AddAppIntent(ctx, migrationAppIntent, project, compositeApp, compositeAppVersion, deploymentIntentGroup, migrationIntent.MetaData.Name, true)


				if migrationAppIntent.Spec.Migration == true && getcluster == cluster {

					fmt.Println("\n\n")
					fmt.Println("migrationAppIntent.Spec.App: ",migrationAppIntent.Spec.App)
					fmt.Println("cluster checked")
					fmt.Println("\n\n")

					AllGenericPlacementIntent, _ := orchModuleLib.NewGenericPlacementIntentClient().GetAllGenericPlacementIntents(ctx, project, compositeApp, compositeAppVersion, deploymentIntentGroup)
					GenericPlacementIntentName := AllGenericPlacementIntent[0].MetaData.Name
					fmt.Println("GenericPlacementIntentName: ",GenericPlacementIntentName,"\n")

					genericAppIntent, _ := orchModuleLib.NewAppIntentClient().GetAllIntentsByApp(ctx, migrationAppIntent.Spec.App, project, compositeApp, compositeAppVersion, GenericPlacementIntentName, deploymentIntentGroup)
					clusterlist := genericAppIntent.Intent.AnyOfArray
					fmt.Println("clusterlist: ",clusterlist, "len(clusterlist): ", len(clusterlist), "\n")

					var list []string


					if len(clusterlist) == 1 {
						list, err = clm.NewClusterClient().GetClustersWithLabel(ctx, clusterlist[0].ProviderName, clusterlist[0].ClusterLabelName)
						if err != nil {
							fmt.Println("GetClusterWithLabel() Error")
						}
						//clusterlistwithlabel = len(list)
						//fmt.Println("\n\n\n")
						//fmt.Println("list: ",list)
						//fmt.Println("\n\n\n")
					}


					fmt.Println("len(list): ",len(list))

					if len(clusterlist) >= 2 || len(list) >= 2 {

						var appwithmigrationintent AppMigrationIntent

						hpaIntents, _ := hpaModuleLib.NewHpaPlacementClient().GetAllIntentsByApp(ctx, migrationAppIntent.Spec.App, project, compositeApp, compositeAppVersion, deploymentIntentGroup)

						var allcpulimits int64
						var allmemorylimits int64

						var hpaIntentName string

						for _ , hpaIntent := range hpaIntents {
							//fmt.Println("\n\n",migrationAppIntent.Spec.App,":hpaIntent:",hpaIntent,"\n\n")

							//appwithmigrationintent 타입의 필드를 채우기 위함
							hpaIntentName = hpaIntent.MetaData.Name

							hpaConsumers, _ := hpaModuleLib.NewHpaPlacementClient().GetAllConsumers(ctx, project, compositeApp, compositeAppVersion, deploymentIntentGroup, hpaIntent.MetaData.Name)

							for _ , hpaConsumer := range hpaConsumers {
								hpaResources, _ := hpaModuleLib.NewHpaPlacementClient().GetAllResources(ctx, project, compositeApp, compositeAppVersion, deploymentIntentGroup, hpaIntent.MetaData.Name, hpaConsumer.MetaData.Name)
								//fmt.Println("hpaResources: ",hpaResources)

								for index, hpaResource := range hpaResources {
									fmt.Println(index,": ",hpaResource.Spec.Resource)
									if hpaResource.Spec.Resource.Name == "cpu" {
										allcpulimits = allcpulimits + (hpaResource.Spec.Resource.Limits * hpaConsumer.Spec.Replicas)
									}
									if hpaResource.Spec.Resource.Name == "memory" {
										allmemorylimits = allmemorylimits + (hpaResource.Spec.Resource.Limits * hpaConsumer.Spec.Replicas)
									}
								}
							}
						}

						key := "/context/" + appcontext + "/app/" + migrationAppIntent.Spec.App + "/cluster/" + cluster + "/"
						var value1 interface{}
						contextdb.Db.Get(ctx, key, &value1)
						if value1 == cluster {
							//fmt.Println("\nallcpulimits: ",allcpulimits, " allmemorylimits: ",allmemorylimits)
							appwithmigrationintent.AppContextID = appcontext
							appwithmigrationintent.Project = project
							appwithmigrationintent.CompositeApp = compositeApp
							appwithmigrationintent.CompositeAppVersion = compositeAppVersion
							appwithmigrationintent.App = migrationAppIntent.Spec.App
							appwithmigrationintent.DeploymentIntentGroup = deploymentIntentGroup
							appwithmigrationintent.HpaIntent = hpaIntentName
							appwithmigrationintent.MigrationIntent = migrationIntent.MetaData.Name
							appwithmigrationintent.MigrationAppIntent = migrationAppIntent.MetaData.Name
							appwithmigrationintent.Migration = migrationAppIntent.Spec.Migration
							appwithmigrationintent.Priority = migrationAppIntent.Spec.Priority
							appwithmigrationintent.CPU = allcpulimits
							appwithmigrationintent.MEMORY = allmemorylimits
							allappwithmigrationintents = append(allappwithmigrationintents, appwithmigrationintent)
						}
					}
				}
			}

			fmt.Println("\ninternal/action/action.go")
			fmt.Println("migrationAppIntents: ",migrationAppIntents,"\n")

		}

	}

	return
}

func calculateappdispersion(ctx context.Context, cluster string, allappwithmigrationintents []AppMigrationIntent)([]AppMigrationIntent) {

	allCA := make(map[string]float64)
	MigrationClient := migrationModuleLib.NewMigrationClient()
	for index, appwithmigrationintent := range allappwithmigrationintents {

		caExists := false
		for key, _ := range allCA {
			if key == appwithmigrationintent.CompositeApp {
				caExists = true
				break
			}
		}
		if !caExists {
			getappcontextfromid, _ := state.GetAppContextFromId(ctx, appwithmigrationintent.AppContextID)
			compositeappmeta, _ := getappcontextfromid.GetCompositeAppMeta(ctx)

			fmt.Println("internal/action/action.go:calculateappdispersion()")
			fmt.Println("getappcontextfromid",getappcontextfromid)
			fmt.Println("compositeappmeta",compositeappmeta)

			project := compositeappmeta.Project
			compositeApp := compositeappmeta.CompositeApp
			compositeAppVersion := compositeappmeta.Version
			deploymentIntentGroup := compositeappmeta.DeploymentIntentGroup


			apps, err := orchModuleLib.NewAppClient().GetApps(ctx, project, compositeApp, compositeAppVersion)
			if err != nil {
				log.Error(" getmigappswithmigrationintents.. Not finding the compositeApp attached apps", log.Fields{"compositeApp": compositeApp})
			}

			migrationIntents, err := MigrationClient.GetAllIntents(ctx, project, compositeApp, compositeAppVersion, deploymentIntentGroup)
			fmt.Println("migrationIntents: ",migrationIntents)


			var allDC []string

			for index, migrationIntent := range migrationIntents {
				log.Info("Migrations .. migrationIntents details => ", log.Fields{
				"intent-index":            index,
				"migration-intent":        migrationIntent,
				"project":                 project,
				"composite-app":           compositeApp,
				"composite-app-version":   compositeAppVersion,
				"deployment-intent-group": deploymentIntentGroup,
				"migration-intent-name":   migrationIntent.MetaData.Name,
				})

				migrationAppIntents, _ := MigrationClient.GetAllAppIntents(ctx, project, compositeApp, compositeAppVersion, deploymentIntentGroup, migrationIntent.MetaData.Name)

				for _, migrationAppIntent := range migrationAppIntents {

					dpExists := false
					for _, DC := range allDC {
						if DC == migrationAppIntent.Status.DeployedCluster {
							dpExists = true
							break
						}
					}
					if !dpExists {
						allDC = append(allDC, migrationAppIntent.Status.DeployedCluster)
					}
				}

				fmt.Println(allDC)
			}

			allCA[appwithmigrationintent.CompositeApp] = float64(len(allDC))/float64(len(apps))
			//fmt.Println(allCA)
		}
		allappwithmigrationintents[index].Dispersion = allCA[appwithmigrationintent.CompositeApp]
	}

	return allappwithmigrationintents

}


func calculateapputilizations(ctx context.Context, cluster string, allappwithmigrationintents []AppMigrationIntent)([]AppMigrationIntent) {

	kubeconfig, err := GetKubeConfig(ctx, cluster, "0", "default")
/*
	fmt.Println("\n\n")
	fmt.Println("kubeconfig: ",string(kubeconfig))
	fmt.Println("\n\n")
*/

	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %v\n", err)
	}



	metricsClientset, err := metricsv.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error building metrics client: %v\n", err)
	}
	for appindex, appwithmigrationintent := range allappwithmigrationintents {

		acRef2, _ := NewAppContextReference(ctx, appwithmigrationintent.AppContextID)
		statusAcID2, _ := acRef2.GetStatusAppContext(ctx, StatusAppContextIDKey)
		var labelSelector metav1.LabelSelector

		if statusAcID2 == "" {
			// Define label selector for pods
			labelSelector.MatchLabels = map[string]string{
				"emco/deployment-id": appwithmigrationintent.AppContextID + "-" + appwithmigrationintent.App,
			}
		} else {
			// Define label selector for pods
			labelSelector.MatchLabels = map[string]string{
				"emco/deployment-id": statusAcID2 + "-" + appwithmigrationintent.App,
			}
		}

		//fmt.Println(labelSelector)

		fmt.Println("statusAcID2: ",statusAcID2)
		fmt.Println("\n\n")


		listOptions := metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(&labelSelector),
		}

		podMetrics, err := metricsClientset.MetricsV1beta1().PodMetricses("").List(ctx, listOptions)
		if err != nil {
			fmt.Printf("Error getting pod metrics: %v\n", err)
		}

/*
		fmt.Println("\n\n")
		fmt.Println("podMetrics: ",podMetrics)
		fmt.Println("\n\n")
*/

		groupedPodMetrics := make(map[string][]v1beta1.PodMetrics)

		for _, podMetric := range podMetrics.Items {
//			fmt.Printf("Pod: %s\n", podMetric.Name)

			parts := strings.Split(podMetric.Name, "-")

			if len(parts) > 0 {
				var podName string

				if len(parts) >= 3 {
					podName = strings.Join(parts[0:len(parts)-2], "-")
				} else {
					podName = podMetric.Name
				}

				if group, exists := groupedPodMetrics[podName]; exists {
					groupedPodMetrics[podName] = append(group, podMetric)
				} else {
					groupedPodMetrics[podName] = []v1beta1.PodMetrics{podMetric}
				}
			}

/*			for _, container := range podMetric.Containers {
				fmt.Printf("Container: %s\n", container.Name)
				fmt.Printf("CPU Usage: %s\n", container.Usage["cpu"])
				fmt.Printf("Memory Usage: %s\n", container.Usage["memory"])
			}
*/		}

		mapforallcpuusageperpod := make(map[string]int64)
		mapforcountperpod := make(map[string]int64)

		for podName, group := range groupedPodMetrics {

			var allcpuusageperpod int64
			//var allmemoryusageperpod int64
			fmt.Printf("Group for pod %s:\n", podName)
			for _, podMetric := range group {
				fmt.Printf("  Pod: %s\n", podMetric.Name)
				for _, container := range podMetric.Containers {
					fmt.Printf("    Container: %s\n", container.Name)
					fmt.Printf("      CPU Usage: %s\n", container.Usage["cpu"])
					fmt.Printf("      Memory Usage: %s\n", container.Usage["memory"])

					//var cpuUsage *Quantity
					cpuUsage := container.Usage["cpu"]
					allcpuusageperpod = allcpuusageperpod + cpuUsage.MilliValue()
				}
			}
			mapforallcpuusageperpod[podName] = allcpuusageperpod
			mapforcountperpod[podName] = int64(len(group))
		}


		groupedHPAConsumers := make(map[string][]hpaModel.HpaResourceConsumer)

		hpaConsumers, _ := hpaModuleLib.NewHpaPlacementClient().GetAllConsumers(ctx, appwithmigrationintent.Project, appwithmigrationintent.CompositeApp, appwithmigrationintent.CompositeAppVersion, appwithmigrationintent.DeploymentIntentGroup, appwithmigrationintent.HpaIntent)


		for _ , hpaConsumer := range hpaConsumers {

			name := hpaConsumer.Spec.Name

			if group, exists := groupedHPAConsumers[name]; exists {
				groupedHPAConsumers[name] = append(group, hpaConsumer)
			} else {
				groupedHPAConsumers[name] = []hpaModel.HpaResourceConsumer{hpaConsumer}
			}
		}

		mapforallcpurequestsperpod := make(map[string]int64)

		for name, group := range groupedHPAConsumers {
			fmt.Printf("Group for Spec.Name %s:\n", name)

			var allcpurequestsperpod int64
			//var allmemoryrequestsperpod int64

			for _, hpaConsumer := range group {

				hpaResources, _ := hpaModuleLib.NewHpaPlacementClient().GetAllResources(ctx, appwithmigrationintent.Project, appwithmigrationintent.CompositeApp, appwithmigrationintent.CompositeAppVersion, appwithmigrationintent.DeploymentIntentGroup, appwithmigrationintent.HpaIntent, hpaConsumer.MetaData.Name)
				for index, hpaResource := range hpaResources {
					fmt.Println(index,": ",hpaResource.Spec.Resource)
					if hpaResource.Spec.Resource.Name == "cpu" {
						allcpurequestsperpod = allcpurequestsperpod + (hpaResource.Spec.Resource.Requests * 1000)
					}
				}
					fmt.Printf("  %v\n", hpaConsumer)
			}
			if mapforcountperpod[name] > 0 {
				mapforallcpurequestsperpod[name] = allcpurequestsperpod * mapforcountperpod[name]
			} else {
				mapforallcpurequestsperpod[name] = allcpurequestsperpod
			}
		}

		var sumofutilization float64
		for key, allcpuusageofpod := range mapforallcpuusageperpod{
			sumofutilization += (float64(allcpuusageofpod) / float64(mapforallcpurequestsperpod[key]))
		}


		allappwithmigrationintents[appindex].CPUUtilization = int64((sumofutilization / float64(len(mapforallcpuusageperpod)))*float64(100))
		fmt.Println("allappwithmigrationintents[appindex].CPUUtilization: ",allappwithmigrationintents[appindex].CPUUtilization)


	}

	return allappwithmigrationintents
}

// compositeapp 내에서 현재 배포 불가능한 app에 대해 그 app에서 필요로 하는 모든 리소스의 합을 구하는 코드
func getrequiredresforapp(ctx context.Context, failedapp string, project string, compositeApp string, compositeAppVersion string, deploymentIntentGroup string) (requiredcpu int64, requiredmemory int64){
	//var requiredcpu int64
	//var requiredmemory int64

	//hpaIntents, err := hpaModuleLib.NewHpaPlacementClient().GetAllIntents(ctx, project, compositeApp, compositeAppVersion, deploymentIntentGroup)
	hpaIntents, err := hpaModuleLib.NewHpaPlacementClient().GetAllIntentsByApp(ctx, failedapp, project, compositeApp, compositeAppVersion, deploymentIntentGroup)
	fmt.Println("\n\nhpaIntents: ",hpaIntents,"\n\n")
	if err != nil {
		log.Error("getrequiredresforapp .. Error getting hpa Intents", log.Fields{"project": project, "compositeApp": compositeApp, "deploymentGroup": deploymentIntentGroup})
		return
	}

	for index, hpaIntent := range hpaIntents {
		log.Info("Migrations .. migrationIntents details => ", log.Fields{
		"intent-index":            index,
		"hpa-intent":              hpaIntent,
		"project":                 project,
		"composite-app":           compositeApp,
		"composite-app-version":   compositeAppVersion,
		"deployment-intent-group": deploymentIntentGroup,
		"hpa-intent-name":         hpaIntent.MetaData.Name,
		"app-name":                hpaIntent.Spec.AppName,
		})

		hpaConsumers, _ := hpaModuleLib.NewHpaPlacementClient().GetAllConsumers(ctx, project, compositeApp, compositeAppVersion, deploymentIntentGroup, hpaIntent.MetaData.Name)

		for index, hpaConsumer := range hpaConsumers {
			log.Info("MigrationApps.. HpaConsumers .. start", log.Fields{"index": index, "app": hpaIntent.Spec.AppName, "hpa-intent": hpaIntent,
			"hpa-consumer": hpaConsumer})

			hpaResources, _ := hpaModuleLib.NewHpaPlacementClient().GetAllResources(ctx, project, compositeApp, compositeAppVersion, deploymentIntentGroup, hpaIntent.MetaData.Name, hpaConsumer.MetaData.Name)
			fmt.Println("hpaResources: ",hpaResources)

			for index, hpaResource := range hpaResources {
				fmt.Println(index,": ",hpaResource.Spec.Resource)
				if hpaResource.Spec.Resource.Name == "cpu" {
					requiredcpu = requiredcpu + hpaResource.Spec.Resource.Requests
				}
				if hpaResource.Spec.Resource.Name == "memory" {
					requiredmemory = requiredmemory + hpaResource.Spec.Resource.Requests
				}
			}
		}
	}
	fmt.Println("\nrequiredcpu: ",requiredcpu, " requiredmemory: ",requiredmemory)
	return
}

func GetClusterApp(ctx context.Context, keys []string, cluster string) (clusterappcontext []string) {
	var appcontexts []string
	for _, key := range keys {
		s := strings.Split(key, "/")
		if len(s) >= 7 && s[4] != "logical-cloud" && s[6] == cluster {
			appcontextId := s[2]
			contains := false
			for _, k := range appcontexts {
				if k == appcontextId {
					contains = true
					break
				}
			}
			if contains == false {
				appcontexts = append(appcontexts, appcontextId)
			}
		}
	}

	fmt.Println("\n\n\n")
	fmt.Println(appcontexts)
	fmt.Println("\n\n\n")

	for _, appcontextId := range appcontexts {
		var value1 map[string]string
		//contextdb.Db.Get(ctx, "/context/" + appcontextId + "/rsync/state/DesiredState/", &value1)
		contextdb.Db.Get(ctx, "/context/" + appcontextId + "/rsync/state/CurrentState/", &value1)
//			var value2 map[string]string
//			contextdb.Db.Get(ctx, "/context/" + appcontextId + /app/nginx-deployment/cluster/OpenStack+cluster1/resourcesready/
		if value1["Status"] == "Instantiated"{
			contains := false
			for _, ac := range clusterappcontext {
				if ac == appcontextId {
					contains = true
					break
				}
			}
			if contains == false {
				clusterappcontext = append(clusterappcontext, appcontextId)
			}
		}
        }
        return
}

func GetClusterAppContext(ctx context.Context, keys []string, cluster string) (clusterappcontext []string) {
	for _, key := range keys {
		s := strings.Split(key, "/")
	        if s[3] == "app"  && s[5] == "cluster" && s[6] == cluster  {
			contains := false
			key = "/context/" + s[2] + "/" /*+ "app/" + s[4] + "/" + "cluster/" + s[6] + "/"*/
			for _, k := range clusterappcontext {
				if k == key {
					contains = true
					break
				}
			}
			if contains == false {
				clusterappcontext = append(clusterappcontext, key)
			}
		}
        }
        return
}


// Publish ... Publish event
func Publish(ctx context.Context, req *clmcontrollerpb.ClmControllerEventRequest) error {

	log.Info("Publish .. start", log.Fields{"req": req, "event": req.Event.String()})

	var err error = nil
	switch req.Event {
	case clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED, clmcontrollerpb.ClmControllerEventType_CLUSTER_UPDATED:
		err = SaveClusterLabelsDB(ctx, req.ProviderName, req.ClusterName)
	case clmcontrollerpb.ClmControllerEventType_CLUSTER_DELETED:
		err = DeleteKubeClusterLabelsDB(ctx, req.ProviderName, req.ClusterName)
	default:
		log.Warn("Publish .. Received Unknown event", log.Fields{"req": req, "event": req.Event.String()})
	}
	if err != nil {
		return pkgerrors.Wrapf(err, "Error while saving Cluster labels[%v]", *req)
	}

	log.Info("Publish .. end", log.Fields{"req": req, "event": req.Event.String()})

	return nil
}

