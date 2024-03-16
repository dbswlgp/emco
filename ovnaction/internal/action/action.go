// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package action

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	jyaml "github.com/ghodss/yaml"
	nettypes "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	clusterPkg "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/ovnaction/pkg/module"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	pkgerrors "github.com/pkg/errors"
)

var CNI_Networking_Nodus_CNI_For_all_interfaces string = "CNI-Networking-Nodus-CNI-For-all-interfaces"
var CNI_Networking_Multi_CNI_Wrapper string = "CNI-Networking-Multi-CNI-Wrapper"
var MultusCNINetworking string = "multus"

// Action applies the supplied intent against the given AppContext ID
func UpdateAppContext(ctx context.Context, intentName, appContextId string) error {
	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(ctx, appContextId)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting AppContext with Id: %v", appContextId)
	}
	caMeta, err := ac.GetCompositeAppMeta(ctx)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting metadata for AppContext with Id: %v", appContextId)
	}

	project := caMeta.Project
	compositeapp := caMeta.CompositeApp
	compositeappversion := caMeta.Version
	deployIntentGroup := caMeta.DeploymentIntentGroup

	// Handle all Workload Intents for the Network Control Intent
	wis, err := module.NewWorkloadIntentClient().GetWorkloadIntents(ctx, project, compositeapp, compositeappversion, deployIntentGroup, intentName)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting Workload Intents for Network Control Intent %v for %v/%v%v/%v not found", intentName, project, compositeapp, deployIntentGroup, compositeappversion)
	}

	// Handle all intents (currently just Workload Interface intents) for each Workload Intent
	for _, wi := range wis {
		// The app/resource identified in the workload intent needs to be updated with two annotations.
		// 1 - The "k8s.v1.cni.cncf.io/networks" annotation will have {"name": "ovn-networkobj", "namespace": "default"} added
		//     to it (preserving any existing values for this annotation.
		// 2 - The "k8s.plugin.opnfv.org/nfn-network" annotation will add any network interfaces that are provided by the
		//     workload/interfaces intents.

		// Prepare the list of interfaces from the workload intent
		wifs, err := module.NewWorkloadIfIntentClient().GetWorkloadIfIntents(ctx, project,
			compositeapp,
			compositeappversion,
			deployIntentGroup,
			intentName,
			wi.Metadata.Name)
		if err != nil {
			return pkgerrors.Wrapf(err,
				"Error getting Workload Interface Intents for Workload Intent %v under Network Control Intent %v for %v/%v%v/%v not found",
				wi.Metadata.Name, intentName, project, compositeapp, compositeappversion, deployIntentGroup)
		}
		if len(wifs) == 0 {
			log.Warn("No interface intents provided for workload intent", log.Fields{
				"project":                 project,
				"composite app":           compositeapp,
				"composite app version":   compositeappversion,
				"deployment intent group": deployIntentGroup,
				"network control intent":  intentName,
				"workload intent":         wi.Metadata.Name,
			})
			continue
		}

		// Get all clusters for the current App from the AppContext
		clusters, err := ac.GetClusterNames(ctx, wi.Spec.AppName)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error getting clusters for app: %v", wi.Spec.AppName)
		}
		for _, c := range clusters {
			rh, err := ac.GetResourceHandle(ctx, wi.Spec.AppName, c,
				strings.Join([]string{wi.Spec.WorkloadResource, wi.Spec.Type}, "+"))
			if err != nil {
				log.Error("App Context resource handle not found", log.Fields{
					"project":                 project,
					"composite app":           compositeapp,
					"composite app version":   compositeappversion,
					"deployment intent group": deployIntentGroup,
					"network control intent":  intentName,
					"workload name":           wi.Metadata.Name,
					"app":                     wi.Spec.AppName,
					"resource":                wi.Spec.WorkloadResource,
					"resource type":           wi.Spec.Type,
				})
				continue
			}
			r, err := ac.GetValue(ctx, rh)
			if err != nil {
				log.Error("Error retrieving resource from App Context", log.Fields{
					"error":           err,
					"resource handle": rh,
				})
				continue
			}
			pc := strings.Split(c, "+")
			// Read the cluster kv pairs for cluster capabilities
			ckv, err := clusterPkg.NewClusterClient().GetAllClusterKvPairs(ctx, pc[0], pc[1])
			var allIntf, cniWrapper string
			// Deafults
			allIntf = "false"
			cniWrapper = MultusCNINetworking
			// Go with defaults if err
			if err == nil {
				for _, kvp := range ckv {
					for _, mkey := range kvp.Spec.Kv {
						if v, ok := mkey[CNI_Networking_Nodus_CNI_For_all_interfaces]; ok {
							allIntf = fmt.Sprintf("%v", v)
							break
						}
						if v, ok := mkey[CNI_Networking_Multi_CNI_Wrapper]; ok {
							cniWrapper = fmt.Sprintf("%v", v)
							break
						}
					}
				}
			}
			var multus bool
			// If all interfaces is nodus, Multus not required
			b, _ := strconv.ParseBool(allIntf)
			if b {
				multus = false
			} else {
				// Compare case insenstive
				// Currently only Multus CNI Wrapper
				multus = strings.EqualFold(MultusCNINetworking, cniWrapper)
			}
			// Add network annotation to object
			netAnnot := nettypes.NetworkSelectionElement{
				Name:      "ovn-networkobj",
				Namespace: "default",
			}
			// Add nfn interface annotations to object
			var newNfnIfs []module.WorkloadIfIntentSpec
			for _, i := range wifs {
				newNfnIfs = append(newNfnIfs, i.Spec)
			}
			var j []byte
			// Unmarshal resource to K8S object
			robj, err := runtime.Decode(scheme.Codecs.UniversalDeserializer(), []byte(r.(string)))
			if err != nil {
				// Not a standard K8s Resource
				//Check if it follows the K8s API Conventions
				j, err = module.AddTemplateAnnotation(r, netAnnot, newNfnIfs, multus)
				if err != nil {
					log.Error("Error AddTemplateAnnotation", log.Fields{
						"error": err,
					})
					continue
				}
			} else {
				if multus {
					module.AddNetworkAnnotation(robj, netAnnot)
				}
				module.AddNfnAnnotation(robj, newNfnIfs)
				// Marshal object back to yaml format (via json - seems to eliminate most clutter)
				j, err = json.Marshal(robj)
				if err != nil {
					log.Error("Error marshalling resource to JSON", log.Fields{
						"error": err,
					})
					continue
				}
			}
			y, err := jyaml.JSONToYAML(j)
			if err != nil {
				log.Error("Error marshalling resource to YAML", log.Fields{
					"error": err,
				})
				continue
			}

			// Update resource in AppContext
			err = ac.UpdateResourceValue(ctx, rh, string(y))
			if err != nil {
				log.Error("Network updating app context resource handle", log.Fields{
					"error":           err,
					"resource handle": rh,
				})
			}
		}
	}

	return nil
}
