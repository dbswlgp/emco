// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package action

import (
	"encoding/json"
	"fmt"
	"strings"

	"context"

	nodus "github.com/akraino-edge-stack/icn-nodus/pkg/apis/k8s/v1alpha1"
	"github.com/ghodss/yaml"
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	catypes "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
	cacontext "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/utils"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/model"
	sfc "gitlab.com/project-emco/core/emco-base/src/sfc/pkg/module"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

type sfcLinks []model.SfcLinkIntent

// getChainApps will return the list of apps that are required to support
// the SFC intent.  Return as a map so that duplicates are eliminated.
func getChainApps(links []model.SfcLinkIntent) map[string]struct{} {
	apps := make(map[string]struct{}, 0)
	for _, link := range links {
		apps[link.Spec.AppName] = struct{}{}
	}
	return apps
}

// chainClusters returns the list of clusters to which the Network Chain needs to be
// deployed.  To qualify, a cluster must be present for each app in the apps list.
func chainClusters(apps map[string]struct{}, ac catypes.CompositeApp) map[string]struct{} {
	clusters := make(map[string]struct{}, 0)
	first := true
	for a, _ := range apps {
		// an app in the chain is not in the AppContext, so the clusters list is empty
		if _, ok := ac.Apps[a]; !ok {
			return make(map[string]struct{}, 0)
		}

		// for first app, the list of that apps clusters in the AppContext is the starting cluster list
		if first {
			for k, _ := range ac.Apps[a].Clusters {
				clusters[k] = struct{}{}
			}
			first = false
		} else {
			// for the rest of the apps, whittle down the clusters list to find the
			// common intersection for all apps in the chain
			for k, _ := range clusters {
				if _, ok := ac.Apps[a].Clusters[k]; !ok {
					delete(clusters, k)
				}
			}
		}
	}
	return clusters
}

// handleSfcLinkIntents - queries all links associated with the sfcIntent and
//   returns the list of link intents
//   returns the network chain string defined by the set of links
//   returns the list of apps that are used in the chain (as a map - since the same app may be present in >1 link)
func handleSfcLinkIntents(ctx context.Context, pr, ca, caver, dig, sfcIntentName string) ([]model.SfcLinkIntent, string, map[string]struct{}, error) {
	apps := make(map[string]struct{}, 0) // returned as the list of apps in the chain

	// Lookup all SFC Link Intents
	sfcLinkIntents, err := sfc.NewSfcLinkIntentClient().GetAllSfcLinkIntents(ctx, pr, ca, caver, dig, sfcIntentName)
	if err != nil {
		return []model.SfcLinkIntent{}, "", apps, pkgerrors.Wrapf(err, "Error getting SFC Link intents for SFC Intent: %v", sfcIntentName)
	}

	leftMap := make(map[string]string)
	rightNets := make(map[string]struct{})
	rightMap := make(map[string]string)

	// initialize the maps
	for _, link := range sfcLinkIntents {
		apps[link.Spec.AppName] = struct{}{}

		if _, ok := leftMap[link.Spec.LeftNet]; ok {
			log.Error("Duplicate left networks in SFC Link Intents", log.Fields{
				"sfc intent":      sfcIntentName,
				"sfc link intent": link,
			})
			return []model.SfcLinkIntent{}, "", apps, pkgerrors.Errorf("Duplicate Left Network in SFC Link Intent: %v", link)
		}
		leftMap[link.Spec.LeftNet] = link.Spec.LinkLabel

		if _, ok := rightNets[link.Spec.RightNet]; ok {
			log.Error("Duplicate right networks in SFC Link Intents", log.Fields{
				"sfc intent":      sfcIntentName,
				"sfc link intent": link,
			})
			return []model.SfcLinkIntent{}, "", apps, pkgerrors.Errorf("Duplicate Right Network in SFC Link Intent: %v", link)
		}
		rightNets[link.Spec.RightNet] = struct{}{}

		if _, ok := rightMap[link.Spec.LinkLabel]; ok {
			log.Error("Duplicate link label in SFC Link Intents", log.Fields{
				"sfc intent":      sfcIntentName,
				"sfc link intent": link,
			})
			return []model.SfcLinkIntent{}, "", apps, pkgerrors.Errorf("Duplicate Link Labelin SFC Link Intent: %v", link)
		}
		rightMap[link.Spec.LinkLabel] = link.Spec.RightNet
	}

	// find the leftmost link
	var leftNet = ""
	for net, _ := range leftMap {
		if _, ok := rightNets[net]; !ok {
			if len(leftNet) > 0 {
				log.Error("Multiple leftmost networks in SFC Link Intents", log.Fields{"sfc intent": sfcIntentName})
				return []model.SfcLinkIntent{}, "", apps, pkgerrors.Errorf("Multiple leftmost Networks in SFC Link Intents")
			}
			leftNet = net
		}
	}

	if len(leftNet) == 0 {
		log.Error("No SFC Link Intents", log.Fields{"sfc intent": sfcIntentName})
		return []model.SfcLinkIntent{}, "", apps, pkgerrors.Errorf("No SFC Link Intents")
	}

	// construct the network chain
	chain := "net=" + leftNet
	cnt := 1
	for true {
		label, ok := leftMap[leftNet]
		if !ok {
			break
		}
		chain = chain + "," + label + ",net=" + rightMap[label]
		leftNet = rightMap[label]
		cnt += 2
	}
	if len(sfcLinkIntents)*2+1 != cnt {
		log.Error("Invalid set of SFC link intents", log.Fields{"sfc intent": sfcIntentName})
		return []model.SfcLinkIntent{}, "", apps, pkgerrors.Errorf("Invalid set of SFC link intents")
	}

	return sfcLinkIntents, chain, apps, nil
}

// Action applies the supplied intent against the given AppContext ID
func UpdateAppContext(ctx context.Context, intentName, appContextId string) error {

	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(ctx, appContextId)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error loading AppContext with Id: %v", appContextId)
	}
	cahandle, err := ac.GetCompositeAppHandle(ctx)
	if err != nil {
		return err
	}

	appContext, err := cacontext.ReadAppContext(ctx, appContextId)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error reading AppContext with Id: %v", appContextId)
	}

	pr := appContext.CompMetadata.Project
	ca := appContext.CompMetadata.CompositeApp
	caver := appContext.CompMetadata.Version
	dig := appContext.CompMetadata.DeploymentIntentGroup

	// Look up all SFC Intents
	sfcIntents, err := sfc.NewSfcIntentClient().GetAllSfcIntents(ctx, pr, ca, caver, dig)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting SFC Intents for Deployment Intent Group: %v", dig)
	}

	if len(sfcIntents) == 0 {
		return pkgerrors.Errorf("No SFC Intents are defined for the Deployment Intent Group: %v", dig)
	}

	// For each SFC Intent prepare a NetworkChaining resource and add to the AppContext
	for i, sfcInt := range sfcIntents {
		// Lookup all SFC Client Selector Intents
		sfcClientSelectorIntents, err := sfc.NewSfcClientSelectorIntentClient().GetAllSfcClientSelectorIntents(ctx, pr, ca, caver, dig, sfcInt.Metadata.Name)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error getting SFC Client Selector intents for SFC Intent: %v", sfcInt.Metadata.Name)
		}

		// Lookup all SFC Provider Network Intents
		sfcProviderNetworkIntents, err := sfc.NewSfcProviderNetworkIntentClient().GetAllSfcProviderNetworkIntents(ctx, pr, ca, caver, dig, sfcInt.Metadata.Name)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error getting SFC Provider Network intents for SFC Intent: %v", sfcInt.Metadata.Name)
		}

		// Prepare the networkchainings CR
		leftNetwork := make([]nodus.RoutingNetwork, 0)
		rightNetwork := make([]nodus.RoutingNetwork, 0)

		for _, sfcClientSelectorInt := range sfcClientSelectorIntents {
			var entry nodus.RoutingNetwork
			entry.PodSelector = sfcClientSelectorInt.Spec.PodSelector
			entry.NamespaceSelector = sfcClientSelectorInt.Spec.NamespaceSelector
			if sfcClientSelectorInt.Spec.ChainEnd == model.LeftChainEnd {
				leftNetwork = append(leftNetwork, entry)
			} else if sfcClientSelectorInt.Spec.ChainEnd == model.RightChainEnd {
				rightNetwork = append(rightNetwork, entry)
			}
		}

		leftIndex := 0
		rightIndex := 0
		for _, sfcProviderNetInt := range sfcProviderNetworkIntents {
			if sfcProviderNetInt.Spec.ChainEnd == model.LeftChainEnd {
				if leftIndex < len(leftNetwork) {
					leftNetwork[leftIndex].NetworkName = sfcProviderNetInt.Spec.NetworkName
					leftNetwork[leftIndex].GatewayIP = sfcProviderNetInt.Spec.GatewayIp
					leftNetwork[leftIndex].Subnet = sfcProviderNetInt.Spec.Subnet
					leftIndex++
				} else {
					var entry nodus.RoutingNetwork
					entry.NetworkName = sfcProviderNetInt.Spec.NetworkName
					entry.GatewayIP = sfcProviderNetInt.Spec.GatewayIp
					entry.Subnet = sfcProviderNetInt.Spec.Subnet
					leftNetwork = append(leftNetwork, entry)
				}
			} else if sfcProviderNetInt.Spec.ChainEnd == model.RightChainEnd {
				if rightIndex < len(rightNetwork) {
					rightNetwork[rightIndex].NetworkName = sfcProviderNetInt.Spec.NetworkName
					rightNetwork[rightIndex].GatewayIP = sfcProviderNetInt.Spec.GatewayIp
					rightNetwork[rightIndex].Subnet = sfcProviderNetInt.Spec.Subnet
					rightIndex++
				} else {
					var entry nodus.RoutingNetwork
					entry.NetworkName = sfcProviderNetInt.Spec.NetworkName
					entry.GatewayIP = sfcProviderNetInt.Spec.GatewayIp
					entry.Subnet = sfcProviderNetInt.Spec.Subnet
					rightNetwork = append(rightNetwork, entry)
				}
			}
		}

		if len(leftNetwork) == 0 && len(rightNetwork) == 0 {
			return pkgerrors.Errorf("provider network or client selector intents were not provided for SFC: %v", sfcInt.Metadata.Name)
		}
		if len(leftNetwork) == 0 {
			return pkgerrors.Errorf("provider network or client selector intents were not provided for left end of SFC: %v", sfcInt.Metadata.Name)
		}
		if len(rightNetwork) == 0 {
			return pkgerrors.Errorf("provider network or client selector intents were not provided for right end of SFC: %v", sfcInt.Metadata.Name)
		}

		sfcLinkIntents, networkChain, chainApps, err := handleSfcLinkIntents(ctx, pr, ca, caver, dig, sfcInt.Metadata.Name)
		if err != nil {
			return err
		}

		log.Info("NetworkChain", log.Fields{"networkChain": networkChain})
		fmt.Println("NetworkChain: " + networkChain)

		chain := nodus.NetworkChaining{
			TypeMeta: metav1.TypeMeta{
				APIVersion: model.ChainingAPIVersion,
				Kind:       model.ChainingKind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: sfcInt.Metadata.Name,
			},
			Spec: nodus.NetworkChainingSpec{
				ChainType: sfcInt.Spec.ChainType,
				RoutingSpec: nodus.RouteSpec{
					Namespace:    sfcInt.Spec.Namespace,
					NetworkChain: networkChain,
					LeftNetwork:  leftNetwork,
					RightNetwork: rightNetwork,
				},
			},
		}
		chainYaml, err := yaml.Marshal(&chain)
		if err != nil {
			return pkgerrors.Wrapf(err, "Failed to marshal NetworkChaining CR: %v", sfcInt.Metadata.Name)
		}

		// Get the clusters which should get the NetworkChaining resource
		clusters := chainClusters(chainApps, appContext)
		if len(clusters) == 0 {
			return pkgerrors.Errorf("There are no clusters with all the apps for the Network Chain: %v", networkChain)
		}

		// Add the network intents chaining app to the AppContext
		var apphandle interface{}
		if i == 0 {
			apphandle, err = ac.AddApp(ctx, cahandle, model.ChainingApp)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding ChainingApp to AppContext: %v", sfcInt.Metadata.Name)
			}

			// need to update the app order instruction
			apporder, err := ac.GetAppInstruction(ctx, appcontext.OrderInstruction)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error getting order instruction while adding ChainingApp to AppContext: %v", sfcInt.Metadata.Name)
			}
			aov := make(map[string][]string)
			json.Unmarshal([]byte(apporder.(string)), &aov)
			aov["apporder"] = append(aov["apporder"], model.ChainingApp)
			jappord, _ := json.Marshal(aov)

			_, err = ac.AddInstruction(ctx, cahandle, appcontext.AppLevel, appcontext.OrderInstruction, string(jappord))
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding ChainingApp to order instruction: %v", sfcInt.Metadata.Name)
			}
		} else {
			apphandle, err = ac.GetAppHandle(ctx, model.ChainingApp)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error getting ChainingApp handle from AppContext: %v", sfcInt.Metadata.Name)
			}
		}

		// Add each cluster to the chaining app and the chaining CR resource to each cluster
		for cluster, _ := range clusters {
			clusterhandle, err := ac.AddCluster(ctx, apphandle, cluster)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding cluster to ChainingApp: %v", cluster)
			}

			resName := sfcInt.Metadata.Name + appcontext.Separator + model.ChainingKind
			_, err = ac.AddResource(ctx, clusterhandle, resName, string(chainYaml))
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding Network Chain resource: %v", sfcInt.Metadata.Name)
			}

			// add (first time) or update the resource order instruction
			aov := make(map[string][]string)
			resorder, err := ac.GetResourceInstruction(ctx, model.ChainingApp, cluster, appcontext.OrderInstruction)
			if err != nil {
				// instruction not found - create it
				aov["resorder"] = []string{resName}
			} else {
				json.Unmarshal([]byte(resorder.(string)), &aov)
				aov["resorder"] = append(aov["resorder"], resName)
			}
			jresord, _ := json.Marshal(aov)

			_, err = ac.AddInstruction(ctx, clusterhandle, appcontext.ResourceLevel, appcontext.OrderInstruction, string(jresord))
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding Network Chain to resource order instruction: %v", sfcInt.Metadata.Name)
			}
		}

		for c, _ := range clusters {
			for _, link := range sfcLinkIntents {

				rh, err := ac.GetResourceHandle(ctx, link.Spec.AppName, c,
					strings.Join([]string{link.Spec.WorkloadResource,
						link.Spec.ResourceType}, "+"))
				if err != nil {
					log.Error("App Context resource handle not found", log.Fields{
						"project":                 pr,
						"composite app":           ca,
						"composite app version":   caver,
						"deployment intent group": dig,
						"sfc intent":              sfcInt.Metadata.Name,
						"sfc client":              link.Metadata.Name,
						"app":                     link.Spec.AppName,
						"resource":                link.Spec.WorkloadResource,
						"resource type":           link.Spec.ResourceType,
					})
					return pkgerrors.Wrapf(err, "Error getting resource handle [%v] for SFC client [%v] from cluster [%v]",
						strings.Join([]string{link.Spec.WorkloadResource,
							link.Spec.ResourceType}, "+"),
						link.Metadata.Name, c)
				}
				r, err := ac.GetValue(ctx, rh)
				if err != nil {
					log.Error("Error retrieving resource from App Context", log.Fields{
						"error":           err,
						"resource handle": rh,
					})
					return pkgerrors.Wrapf(err, "Error getting resource value [%v] for SFC client [%v] from cluster [%v]",
						strings.Join([]string{link.Spec.WorkloadResource,
							link.Spec.ResourceType}, "+"),
						link.Metadata.Name, c)
				}

				// Unmarshal resource to K8S object
				robj, err := runtime.Decode(scheme.Codecs.UniversalDeserializer(), []byte(r.(string)))
				if err != nil {
					return pkgerrors.Wrapf(err, "Error decoding resource: %v", link.Spec.WorkloadResource)
				}

				// add labels to resource
				addLabelToPodTemplates(robj, link.Spec.LinkLabel)

				// Marshal object back to yaml format (via json - seems to eliminate most clutter)
				j, err := json.Marshal(robj)
				if err != nil {
					log.Error("Error marshalling resource to JSON", log.Fields{
						"error": err,
					})
					return pkgerrors.Wrapf(err,
						"Error marshalling to JSON resource value [%v] for SFC link resource labelling [%v] from cluster [%v]",
						strings.Join([]string{link.Spec.WorkloadResource,
							link.Spec.ResourceType}, "+"),
						link.Metadata.Name,
						c)
				}
				y, err := yaml.JSONToYAML(j)
				if err != nil {
					log.Error("Error marshalling resource to YAML", log.Fields{
						"error": err,
					})
					return pkgerrors.Wrapf(err,
						"Error marshalling to YAML resource value [%v] for SFC client [%v] from cluster [%v]",
						strings.Join([]string{link.Spec.WorkloadResource,
							link.Spec.ResourceType}, "+"),
						link.Metadata.Name, c)
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
	}
	return nil
}
