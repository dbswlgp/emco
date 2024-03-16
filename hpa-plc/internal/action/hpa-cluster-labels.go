// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package action

import (
	"context"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	connector "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/utils"
)

// PlacementClusterKey is the key structure that is used in the database
type PlacementClusterKey struct {
	ClusterProviderName   string `json:"hpaPlcClusterProvider"`
	ClusterName           string `json:"hpaPlcCluster"`
	ClusterControllerData string `json:"hpaPlcClusterControllerData"`
}

const (
	// name of the mongodb collection to use for client documents
	clmClusterCollectionName = "resources"
	// name of the cluster controller information key
	hpaClusterControllerData = "clusterHpaInfo"
	// attribute key name for the K8S Cluster Labels of a client document
	tagClusterLabels = "clusterLabels"
	// attribute key name for the K8S Cluster nodes of a client document
	tagClusterNodes = "clusterNodes"
)

// KubeClusterInfo .. structure for storing kubernetes cluster info
type KubeClusterInfo struct {
	NodeNames []string `json:"cluster-nodes"`
}

// SaveClusterLabelsDB - save Cluster Labels to DB
func SaveClusterLabelsDB(ctx context.Context, provider string, cluster string) error {
	log.Info("SaveClusterLabelsDB .. start", log.Fields{
		"clusterProvider": provider,
		"cluster":         cluster,
	})
	//Construct key and tag to select the entry
	key := PlacementClusterKey{
		ClusterProviderName:   provider,
		ClusterName:           cluster,
		ClusterControllerData: hpaClusterControllerData,
	}

	// Connect to Cluster
	con := connector.Connection{}
	con.Init("cluster")
	//Cleanup
	defer con.RemoveClient()

	clusterFQDN := (provider + "+" + cluster)
	// Get Kube Client handle
	c, err := con.GetClient(ctx, clusterFQDN, "0", "default")
	if err != nil {
		log.Error("SaveClusterLabelsDB .. Error in creating kubeconfig client", logutils.Fields{
			"error":       err,
			"clusterFQDN": clusterFQDN,
		})
		return err
	}

	// Fetch K8s Cluster Nodes & store in db
	nodeList, err := c.GetClusterNodes(ctx)
	if err != nil {
		log.Error("SaveClusterLabelsDB .. Error in fetching K8s Cluster Nodes", log.Fields{
			"clusterFQDN": clusterFQDN})
		return pkgerrors.Wrapf(err, "SaveClusterLabelsDB .. Error in fetching K8s Cluster Nodes of clusterFQDN[%s]", clusterFQDN)
	}
	clInfo := KubeClusterInfo{}
	for _, node := range nodeList.Items {
		clInfo.NodeNames = append(clInfo.NodeNames, node.Name)
	}
	err = db.DBconn.Insert(ctx, clmClusterCollectionName, key, nil, tagClusterNodes, clInfo)
	if err != nil {
		log.Error("SaveClusterLabelsDB .. Error in creating DB Entry for Cluster Nodes", log.Fields{"clusterFQDN": clusterFQDN})
		return pkgerrors.Wrapf(err, "SaveClusterLabelsDB .. Error in creating DB Entry for cluster[%s] to store Cluster Nodes", clusterFQDN)
	}
	log.Info("SaveClusterLabelsDB .. Added Cluster Nodes to db", log.Fields{
		"clusterFQDN": clusterFQDN,
		"nodes":       clInfo.NodeNames})

	// Fetch K8s Cluster Labels & store in db
	nodeLabels, err := c.GetNodeLabels(ctx)
	if err != nil {
		log.Error("SaveClusterLabelsDB .. Error in fetching K8s Cluster Labels", log.Fields{
			"clusterFQDN": clusterFQDN})
		return pkgerrors.Wrapf(err, "SaveClusterLabelsDB .. Error in fetching K8s Cluster Labels of clusterFQDN[%s]", clusterFQDN)
	}
	log.Info("SaveClusterLabelsDB .. Node Labels", log.Fields{"clusterFQDN": clusterFQDN, "kubeNodeLabelsMap": nodeLabels})

	err = db.DBconn.Insert(ctx, clmClusterCollectionName, key, nil, tagClusterLabels, nodeLabels)
	if err != nil {
		log.Error("SaveClusterLabelsDB .. Error in creating DB Entry for Cluster Labels", log.Fields{"clusterFQDN": clusterFQDN})
		return pkgerrors.Wrapf(err, "SaveClusterLabelsDB .. Error in creating DB Entry for clusterFQDN[%s] to store Cluster Labels", clusterFQDN)
	}
	log.Info("SaveClusterLabelsDB .. end. Added Cluster Labels to db", log.Fields{
		"db.storeName": clmClusterCollectionName,
		"db.key":       key,
		"clusterFQDN":  clusterFQDN,
		"nodes":        clInfo.NodeNames,
		"node-labels":  nodeLabels})

	return nil
}

// GetKubeClusterLabels .. returns the Cluster Labels of K8s cluster
func GetKubeClusterLabels(ctx context.Context, provider, cluster string) (map[string](map[string]string), error) {
	log.Info("GetKubeClusterLabels .. start ", log.Fields{
		"clusterProvider": provider,
		"cluster":         cluster})

	//Construct key and tag to select the entry
	key := PlacementClusterKey{
		ClusterProviderName:   provider,
		ClusterName:           cluster,
		ClusterControllerData: hpaClusterControllerData,
	}

	values, err := db.DBconn.Find(ctx, clmClusterCollectionName, key, tagClusterLabels)
	if err != nil {
		log.Error("GetKubeClusterLabels .. Error in getting Kube Cluster Labels", log.Fields{
			"clusterProvider": provider,
			"cluster":         cluster,
			"nodeLabels":      values})
		return nil, pkgerrors.Wrap(err, "GetKubeClusterLabels .. Error in getting Kube Cluster Labels")
	}
	var resp map[string](map[string]string)
	resp = make(map[string](map[string]string))
	if values != nil && values[0] != nil {
		err = db.DBconn.Unmarshal(values[0], &resp)
		if err != nil {
			log.Error("GetKubeClusterLabels ..  Error in unmarshalling Cluster labels", log.Fields{
				"clusterProvider": provider,
				"cluster":         cluster,
				"nodeLabels":      values[0]})

			return nil, err
		}
	}

	log.Info("GetKubeClusterLabels .. end", log.Fields{
		"clusterProvider": provider,
		"cluster":         cluster,
		"db.storeName":    clmClusterCollectionName,
		"db.key":          key,
		"nodeLabels":      resp})
	return resp, nil
}

// DeleteKubeClusterLabelsDB .. delete cluster Labels from db
func DeleteKubeClusterLabelsDB(ctx context.Context, provider, cluster string) error {
	log.Info("DeleteKubeClusterLabelsDB .. start ", log.Fields{
		"clusterProvider": provider,
		"cluster":         cluster})

	//Construct key and tag to select the entry
	key := PlacementClusterKey{
		ClusterProviderName:   provider,
		ClusterName:           cluster,
		ClusterControllerData: hpaClusterControllerData,
	}

	values, err := db.DBconn.Find(ctx, clmClusterCollectionName, key, tagClusterLabels)
	if err != nil {
		log.Error("DeleteKubeClusterLabelsDB .. Error in getting Kube Cluster Labels", log.Fields{
			"clusterProvider": provider,
			"cluster":         cluster,
			"nodeLabels":      values})
		return pkgerrors.Wrap(err, "DeleteKubeClusterLabelsDB .. Error in getting Kube Cluster Labels")
	}

	log.Info("DeleteKubeClusterLabelsDB ... Delete Cluster labels entry", log.Fields{"StoreName": clmClusterCollectionName, "key": key})
	err = db.DBconn.Remove(ctx, clmClusterCollectionName, key)
	if err != nil {
		log.Error("DeleteKubeClusterLabelsDB ... DB Error .. Delete Cluster labels entry error", log.Fields{"err": err, "StoreName": clmClusterCollectionName, "key": key})
		return pkgerrors.Wrapf(err, "DeleteKubeClusterLabelsDB ... DB Error .. Delete Cluster labels for key[%s] DB Error", key)
	}

	log.Info("DeleteKubeClusterLabelsDB .. end", log.Fields{
		"clusterProvider": provider,
		"cluster":         cluster,
		"db.storeName":    clmClusterCollectionName,
		"db.key":          key})
	return nil
}
