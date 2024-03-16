// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package action

import (
	"encoding/json"
	"strings"

	"context"

	jyaml "github.com/ghodss/yaml"
	pkgerrors "github.com/pkg/errors"
	dcm "gitlab.com/project-emco/core/emco-base/src/dcm/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	orch "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	catypes "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
	cacontext "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/utils"
	sfc "gitlab.com/project-emco/core/emco-base/src/sfc/pkg/module"
	sfcclient "gitlab.com/project-emco/core/emco-base/src/sfcclient/pkg/module"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// getChainApps will return the list of applications that are present in
// the provided string which follows the format of the NetworkChain field.
// "net=virutal-net1,app=slb,dync-net1,app=ngfw,dync-net2,app=sdewan,net=virutal-net2"
func getChainApps(networkChain string) ([]string, error) {
	netsAndApps := strings.Split(networkChain, ",")
	apps := make([]string, 0)
	for _, netOrApp := range netsAndApps {
		elem := strings.Split(netOrApp, "=")
		if len(elem) != 2 {
			return []string{}, pkgerrors.Errorf("Invalid network chain format: %v", networkChain)
		}
		if elem[0] == "app" {
			apps = append(apps, elem[1])
		}
	}
	return apps, nil
}

// chainClusters returns the list of clusters to which the Network Chain needs to be
// deployed.  To qualify, a cluster must be present for each app in the apps list.
func chainClusters(apps []string, ac catypes.CompositeApp) map[string]struct{} {
	clusters := make(map[string]struct{}, 0)
	for i, a := range apps {
		// an app in the chain is not in the AppContext, so the clusters list is empty
		if _, ok := ac.Apps[a]; !ok {
			return make(map[string]struct{}, 0)
		}

		// for first app, the list of that apps clusters in the AppContext is the starting cluster list
		if i == 0 {
			for k, _ := range ac.Apps[a].Clusters {
				clusters[k] = struct{}{}
			}
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

func getDigLabels(ctx context.Context, pr, ca, caver, dig string) (map[string]string, error) {
	labels := make(map[string]string)

	d, err := orch.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(ctx, dig, pr, ca, caver)
	if err != nil {
		log.Error("Error find DeploymentIntentGroup ", log.Fields{"DeploymentIntentGroup: ": dig})
		return labels, err
	}

	lc, err := dcm.NewLogicalCloudClient().Get(ctx, pr, d.Spec.LogicalCloud)
	if err != nil {
		log.Error("Error find Logical Cloud for DeploymentIntentGroup ", log.Fields{"DeploymentIntentGroup: ": dig, "Logicalcloud": d.Spec.LogicalCloud})
		return labels, err
	}

	for k, v := range lc.Specification.Labels {
		labels[k] = v
	}
	return labels, nil
}

func matchesDigLabels(intentLabels, digLabels map[string]string) bool {
	if len(intentLabels) == 0 {
		return false
	}

	for k, v := range intentLabels {
		if dv, ok := digLabels[k]; !ok || v != dv {
			return false
		}
	}
	return true
}

// UpdateAppContext applies the supplied intent against the given AppContext ID
// The SFC Client controller will handle all SFC Client intents that are found for the
// Deployment Intent Group of the appContext
func UpdateAppContext(ctx context.Context, intentName, appContextId string) error {

	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(ctx, appContextId)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error loading AppContext with Id: %v", appContextId)
	}
	//cahandle, err := ac.GetCompositeAppHandle(ctx)
	_, err = ac.GetCompositeAppHandle(ctx)
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
	//ns := appContext.CompMetadata.Namespace

	digLabels, err := getDigLabels(ctx, pr, ca, caver, dig)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting logical namespace labels for SFC Client Deployment Intent Group: %v", dig)
	}

	// Look up all SFC Client Intents
	sfcClientIntents, err := sfcclient.NewSfcClient().GetAllSfcClientIntents(ctx, pr, ca, caver, dig)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting SFC Client Intents for Deployment Intent Group: %v", dig)
	}

	if len(sfcClientIntents) == 0 {
		return pkgerrors.Errorf("No SFC Client Intents are defined for the Deployment Intent Group: %v", dig)
	}

	// For each SFC Client Intent ...
	for _, sfcClientInt := range sfcClientIntents {

		// query all SFC Client Selector Intents that match the chainEnd of this SFC Client Intent
		sfcClientSelectorIntents, err := sfc.NewSfcClientSelectorIntentClient().GetSfcClientSelectorIntentsByEnd(ctx,
			pr,
			sfcClientInt.Spec.ChainCompositeApp,
			sfcClientInt.Spec.ChainCompositeAppVersion,
			sfcClientInt.Spec.ChainDeploymentIntentGroup,
			sfcClientInt.Spec.ChainName,
			sfcClientInt.Spec.ChainEnd)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error finding SFC Client Selector Intents for %v by end: %v", sfcClientInt.Spec.ChainName, sfcClientInt.Spec.ChainEnd)
		}
		if len(sfcClientSelectorIntents) == 0 {
			return pkgerrors.Errorf("No SFC Client Selector Intents found for %v by end: %v", sfcClientInt.Spec.ChainName, sfcClientInt.Spec.ChainEnd)
		}

		// Identify the SFC Client Selector intent that matches this Dig - by matching the namespace labels
		// on this Dig's Logical Cloud with the namespace labels in the SFC Client Selector Intents.
		// Copy Pod Selector labels in the SFC Client Selector intent for application to the
		// identified resource Pod template further below.
		labels := make(map[string]string)
		for _, sfcClientSelectorIntent := range sfcClientSelectorIntents {
			if matchesDigLabels(sfcClientSelectorIntent.Spec.NamespaceSelector.MatchLabels, digLabels) {
				for k, v := range sfcClientSelectorIntent.Spec.PodSelector.MatchLabels {
					labels[k] = v
				}
				// break out, only need to get labels for the first identified SFC Client Selector intent
				break
			}
		}
		if len(labels) == 0 {
			return pkgerrors.Errorf("No matching SFC Client Selector Intents found for %v by end: %v", sfcClientInt.Spec.ChainName, sfcClientInt.Spec.ChainEnd)
		}

		// Get all clusters for the current App from the AppContext
		clusters, err := ac.GetClusterNames(ctx, sfcClientInt.Spec.AppName)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error getting clusters for app: %v", sfcClientInt.Spec.AppName)
		}
		for _, c := range clusters {
			rh, err := ac.GetResourceHandle(ctx, sfcClientInt.Spec.AppName, c,
				strings.Join([]string{sfcClientInt.Spec.WorkloadResource,
					sfcClientInt.Spec.ResourceType}, "+"))
			if err != nil {
				log.Error("App Context resource handle not found", log.Fields{
					"project":                 pr,
					"composite app":           ca,
					"composite app version":   caver,
					"deployment intent group": dig,
					"sfc client":              sfcClientInt.Metadata.Name,
					"app":                     sfcClientInt.Spec.AppName,
					"resource":                sfcClientInt.Spec.WorkloadResource,
					"resource type":           sfcClientInt.Spec.ResourceType,
				})
				return pkgerrors.Wrapf(err, "Error getting resource handle [%v] for SFC client [%v] from cluster [%v]",
					strings.Join([]string{sfcClientInt.Spec.WorkloadResource,
						sfcClientInt.Spec.ResourceType}, "+"),
					sfcClientInt.Metadata.Name, c)
			}
			r, err := ac.GetValue(ctx, rh)
			if err != nil {
				log.Error("Error retrieving resource from App Context", log.Fields{
					"error":           err,
					"resource handle": rh,
				})
				return pkgerrors.Wrapf(err, "Error getting resource value [%v] for SFC client [%v] from cluster [%v]",
					strings.Join([]string{sfcClientInt.Spec.WorkloadResource,
						sfcClientInt.Spec.ResourceType}, "+"),
					sfcClientInt.Metadata.Name, c)
			}

			// Unmarshal resource to K8S object
			robj, err := runtime.Decode(scheme.Codecs.UniversalDeserializer(), []byte(r.(string)))
			if err != nil {
				return pkgerrors.Wrapf(err, "Error decoding resource: %v", sfcClientInt.Spec.WorkloadResource)
			}

			// add labels to resource
			AddLabelsToPodTemplates(robj, labels)

			// Marshal object back to yaml format (via json - seems to eliminate most clutter)
			j, err := json.Marshal(robj)
			if err != nil {
				log.Error("Error marshalling resource to JSON", log.Fields{
					"error": err,
				})
				return pkgerrors.Wrapf(err,
					"Error marshalling to JSON resource value [%v] for SFC client [%v] from cluster [%v]",
					strings.Join([]string{sfcClientInt.Spec.WorkloadResource,
						sfcClientInt.Spec.ResourceType}, "+"),
					sfcClientInt.Metadata.Name,
					c)
			}
			y, err := jyaml.JSONToYAML(j)
			if err != nil {
				log.Error("Error marshalling resource to YAML", log.Fields{
					"error": err,
				})
				return pkgerrors.Wrapf(err,
					"Error marshalling to YAML resource value [%v] for SFC client [%v] from cluster [%v]",
					strings.Join([]string{sfcClientInt.Spec.WorkloadResource,
						sfcClientInt.Spec.ResourceType}, "+"),
					sfcClientInt.Metadata.Name, c)
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
