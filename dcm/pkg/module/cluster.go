// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2022 Intel Corporation

package module

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"context"

	pkgerrors "github.com/pkg/errors"
	rb "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	rsync "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"
	"gopkg.in/yaml.v2"
)

// ClusterManager is an interface that exposes the connection
// functionality
type ClusterManager interface {
	CreateCluster(ctx context.Context, project, logicalCloud string, c common.Cluster) (common.Cluster, error)
	GetCluster(ctx context.Context, project, logicalCloud, name string) (common.Cluster, error)
	GetAllClusters(ctx context.Context, project, logicalCloud string) ([]common.Cluster, error)
	DeleteCluster(ctx context.Context, project, logicalCloud, name string) error
	UpdateCluster(ctx context.Context, project, logicalCloud, name string, c common.Cluster) (common.Cluster, error)
	GetClusterConfig(ctx context.Context, project, logicalcloud, name string) (string, error)
}

// ClusterClient implements the ClusterManager
// It will also be used to maintain some localized state
type ClusterClient struct {
	storeName string
	tagMeta   string
}

// ClusterClient returns an instance of the ClusterClient
// which implements the ClusterManager
func NewClusterClient() *ClusterClient {
	return &ClusterClient{
		storeName: "resources",
		tagMeta:   "data",
	}
}

// Create entry for the cluster reference resource in the database
func (v *ClusterClient) CreateCluster(ctx context.Context, project, logicalCloud string, clusterReference common.Cluster) (common.Cluster, error) {

	//Construct key consisting of name
	key := common.ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: clusterReference.MetaData.Name,
	}
	lcClient := NewLogicalCloudClient()

	s, err := lcClient.GetState(ctx, project, logicalCloud)
	if err != nil {
		return common.Cluster{}, err
	}
	cid := state.GetLastContextIdFromStateInfo(s)

	if cid != "" {
		// Since there's a appCtx associated, if the logical cloud isn't fully
		// Terminated or fully Instantiated, then prevent clusters from being added
		acStatus, err := state.GetAppContextStatus(ctx, cid) // new from state
		if err != nil {
			return common.Cluster{}, err
		}
		if acStatus.Status != appcontext.AppContextStatusEnum.Terminated && acStatus.Status != appcontext.AppContextStatusEnum.Instantiated {
			return common.Cluster{}, pkgerrors.New("Cluster References cannot be added/removed unless the Logical Cloud is fully Instantiated or Terminated")
		}
	}

	//Check if this Cluster Reference already exists
	_, err = v.GetCluster(ctx, project, logicalCloud, clusterReference.MetaData.Name)
	if err == nil {
		return common.Cluster{}, pkgerrors.New("Cluster reference already exists")
	}

	log.Info("Adding cluster-reference. Changes won't take effect until the respective Logical Cloud is updated.", log.Fields{"clusterreference": clusterReference, "logicalcloud": logicalCloud})
	err = db.DBconn.Insert(ctx, v.storeName, key, nil, v.tagMeta, clusterReference)
	if err != nil {
		return common.Cluster{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return clusterReference, nil
}

// Get returns Cluster for corresponding cluster reference
func (v *ClusterClient) GetCluster(ctx context.Context, project, logicalCloud, clusterReference string) (common.Cluster, error) {

	//Construct the composite key to select the entry
	key := common.ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: clusterReference,
	}

	value, err := db.DBconn.Find(ctx, v.storeName, key, v.tagMeta)
	if err != nil {
		return common.Cluster{}, err
	}

	if len(value) == 0 {
		return common.Cluster{}, pkgerrors.New("Cluster reference not found")
	}

	//value is a byte array
	if value != nil {
		cl := common.Cluster{}
		err = db.DBconn.Unmarshal(value[0], &cl)
		if err != nil {
			return common.Cluster{}, err
		}
		return cl, nil
	}

	return common.Cluster{}, pkgerrors.New("Unknown Error")
}

// GetAll returns all cluster references in the logical cloud
func (v *ClusterClient) GetAllClusters(ctx context.Context, project, logicalCloud string) ([]common.Cluster, error) {
	//Construct the composite key to select clusters
	key := common.ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: "",
	}
	var resp []common.Cluster
	values, err := db.DBconn.Find(ctx, v.storeName, key, v.tagMeta)
	if err != nil {
		return []common.Cluster{}, err
	}
	if len(values) == 0 {
		return []common.Cluster{}, pkgerrors.New("No Cluster References associated")
	}

	for _, value := range values {
		cl := common.Cluster{}
		err = db.DBconn.Unmarshal(value, &cl)
		if err != nil {
			return []common.Cluster{}, err
		}
		resp = append(resp, cl)
	}

	return resp, nil
}

// Delete the Cluster reference entry from database
func (v *ClusterClient) DeleteCluster(ctx context.Context, project, logicalCloud, clusterReference string) error {
	//Construct the composite key to select the entry
	key := common.ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: clusterReference,
	}

	lcClient := NewLogicalCloudClient()

	s, err := lcClient.GetState(ctx, project, logicalCloud)
	if err != nil {
		return err
	}
	cid := state.GetLastContextIdFromStateInfo(s)

	if cid == "" {
		// Just go ahead and delete the reference if there is no logical cloud appCtx yet
		err := db.DBconn.Remove(ctx, v.storeName, key)
		if err != nil {
			return pkgerrors.Wrap(err, "Failed deleting Cluster Reference")
		}
		return nil
	}

	// Since there's a appCtx associated, if the logical cloud isn't fully
	// Terminated or fully Instantiated, then prevent clusters from being added
	acStatus, err := state.GetAppContextStatus(ctx, cid) // new from state
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting app context status")
	}
	switch acStatus.Status {
	case appcontext.AppContextStatusEnum.Terminating:
		log.Error("Can't remove Cluster Reference yet: the Logical Cloud is being terminated.", log.Fields{"clusterreference": clusterReference, "logicalcloud": logicalCloud})
		return pkgerrors.New("Can't remove Cluster Reference: the Logical Cloud is being terminated.")
	case appcontext.AppContextStatusEnum.Instantiating:
		log.Error("Can't remove Cluster Reference: the Logical Cloud is instantiating, please wait and then terminate.", log.Fields{"clusterreference": clusterReference, "logicalcloud": logicalCloud})
		return pkgerrors.New("Can't remove Cluster Reference: the Logical Cloud is instantiating, please wait and then terminate.")
	case appcontext.AppContextStatusEnum.InstantiateFailed:
		log.Error("Can't remove Cluster Reference: the Logical Cloud has failed instantiating, for safety please terminate and try again.", log.Fields{"clusterreference": clusterReference, "logicalcloud": logicalCloud})
		return pkgerrors.New("Can't remove Cluster Reference: the Logical Cloud has failed instantiating, for safety please terminate and try again.")
	case appcontext.AppContextStatusEnum.TerminateFailed:
		log.Info("The Logical Cloud has failed terminating, proceeding with the delete operation.", log.Fields{"clusterreference": clusterReference, "logicalcloud": logicalCloud})
		// try to delete anyway since termination failed
		fallthrough
	case appcontext.AppContextStatusEnum.Instantiated:
		log.Error("Removing cluster-reference. Changes won't take effect until the respective Logical Cloud is updated.", log.Fields{"clusterreference": clusterReference, "logicalcloud": logicalCloud})
		fallthrough
	case appcontext.AppContextStatusEnum.Terminated:
		err := db.DBconn.Remove(ctx, v.storeName, key)
		if err != nil {
			return pkgerrors.Wrap(err, "Error deleting Cluster Reference")
		}

		log.Info("Removed cluster reference from Logical Cloud.", log.Fields{"logicalcloud": logicalCloud})
		return nil
	default:
		log.Error("Failure removing Cluster Reference: the Logical Cloud isn't in an expected status so not taking any action.", log.Fields{"clusterreference": clusterReference, "logicalcloud": logicalCloud, "status": acStatus.Status})
		return pkgerrors.New("Failure removing Cluster Reference: the Logical Cloud isn't in an expected status so not taking any action.")
	}
}

// Update an entry for the Cluster reference in the database
func (v *ClusterClient) UpdateCluster(ctx context.Context, project, logicalCloud, clusterReference string, c common.Cluster) (common.Cluster, error) {

	key := common.ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: clusterReference,
	}

	//Check for name mismatch in cluster reference
	if c.MetaData.Name != clusterReference {
		return common.Cluster{}, pkgerrors.New("Cluster Reference mismatch")
	}
	//Check if this Cluster reference exists
	_, err := v.GetCluster(ctx, project, logicalCloud, clusterReference)
	if err != nil {
		return common.Cluster{}, err
	}
	err = db.DBconn.Insert(ctx, v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return common.Cluster{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}

// Get returns Cluster's kubeconfig for corresponding cluster reference
func (v *ClusterClient) GetClusterConfig(ctx context.Context, project, logicalCloud, clusterReference string) (string, error) {
	lcClient := NewLogicalCloudClient()
	lckey := common.LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
	}
	s, err := lcClient.GetState(ctx, project, logicalCloud)
	if err != nil {
		return "", err
	}
	cid := state.GetStatusContextIdFromStateInfo(s)

	if cid == "" {
		return "", pkgerrors.New("Logical Cloud hasn't been instantiated yet")
	}

	ac, err := state.GetAppContextFromId(ctx, cid)
	if err != nil {
		return "", err
	}

	// get logical cloud resource
	lc, err := lcClient.Get(ctx, project, logicalCloud)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error getting logical cloud")
	}
	// get cluster from dcm (need provider/name)
	cluster, err := v.GetCluster(ctx, project, logicalCloud, clusterReference)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error getting cluster")
	}
	ccc := rsync.NewCloudConfigClient()
	gitOps, err := IsGitOpsCluster(ctx, strings.Join([]string{cluster.Specification.ClusterProvider, "+", cluster.Specification.ClusterName}, ""))
	if err != nil {
		return "", err
	}
	if gitOps {

		// Create a record in Rsync for logical cloud
		_, err = ccc.CreateCloudConfig(
			ctx,
			cluster.Specification.ClusterProvider,
			cluster.Specification.ClusterName,
			lc.Specification.Level,
			lc.Specification.NameSpace,
			"")
		if err != nil {
			if err.Error() != "CloudConfig already exists" {
				return "", pkgerrors.Wrap(err, "Failed creating a new gitOps logical cloud in rsync's CloudConfig")
			}
		}
		return "", nil
	}
	// get user's private key
	pkDataArray, err := db.DBconn.Find(ctx, v.storeName, lckey, "privatekey")
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error getting private key from logical cloud")
	}
	if len(pkDataArray) == 0 {
		return "", pkgerrors.Errorf("Private key from logical cloud not found")
	}
	privateKey := common.PrivateKey{}
	err = db.DBconn.Unmarshal(pkDataArray[0], &privateKey)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Private key unmarshal error")
	}
	// before attempting to generate a kubeconfig,
	// check if certificate has been issued and copy it from etcd to mongodb
	if cluster.Specification.Certificate == "" {
		log.Info("Certificate not yet in MongoDB, checking etcd.", log.Fields{"logical cloud": logicalCloud, "cluster ref": clusterReference})

		// access etcd
		clusterName := strings.Join([]string{cluster.Specification.ClusterProvider, "+", cluster.Specification.ClusterName}, "")

		// get the app context handle for the status of this cluster (which should contain the certificate inside, if already issued)
		statusHandle, err := ac.GetClusterStatusHandle(ctx, "logical-cloud", clusterName)

		if err != nil {
			return "", pkgerrors.Wrap(err, "The cluster doesn't contain status, please check if all services are up and running")
		}
		statusRaw, err := ac.GetValue(ctx, statusHandle)
		if err != nil {
			return "", pkgerrors.Wrap(err, "An error occurred while reading the cluster status")
		}

		var rbstatus rb.ResourceBundleStateStatus
		err = json.Unmarshal([]byte(statusRaw.(string)), &rbstatus)
		if err != nil {
			return "", pkgerrors.Wrap(err, "An error occurred while parsing the cluster status")
		}

		if len(rbstatus.CsrStatuses) == 0 {
			return "", pkgerrors.New("A status for the CSR hasn't been returned yet")
		}

		// validate that we indeed obtained a certificate before persisting it in the database:
		approved := false
		for _, c := range rbstatus.CsrStatuses[0].Status.Conditions {
			if c.Type == "Denied" {
				return "", pkgerrors.New("Certificate was denied!")
			}
			if c.Type == "Failed" {
				return "", pkgerrors.New("Certificate issue failed")
			}
			if c.Type == "Approved" {
				approved = true
			}
		}
		if !approved {
			return "", pkgerrors.New("The CSR hasn't been approved yet or the certificate hasn't been issued yet")
		}

		//just double-check certificate field contents aren't empty:
		cert := rbstatus.CsrStatuses[0].Status.Certificate
		if len(cert) > 0 {
			cluster.Specification.Certificate = base64.StdEncoding.EncodeToString([]byte(cert))
		} else {
			return "", pkgerrors.New("Certificate issued was invalid")
		}

		_, err = v.UpdateCluster(ctx, project, logicalCloud, clusterReference, cluster)
		if err != nil {
			return "", pkgerrors.Wrap(err, "An error occurred while storing the certificate")
		}
	} else {
		// certificate is already in MongoDB so just hand it over to create the API response
		log.Info("Certificate already in MongoDB, pass it to API", log.Fields{})
	}

	// sanity check for cluster-issued certificate
	if cluster.Specification.Certificate == "" {
		return "", pkgerrors.New("Failed creating kubeconfig due to unexpected empty certificate")
	}

	// get kubeconfig from L0 cloudconfig respective to the cluster referenced by this logical cloud
	cconfig, err := ccc.GetCloudConfig(ctx, cluster.Specification.ClusterProvider, cluster.Specification.ClusterName, "0", "")
	if err != nil {
		return "", pkgerrors.New("Failed fetching kubeconfig from rsync's CloudConfig")
	}
	adminConfig, err := base64.StdEncoding.DecodeString(cconfig.Config)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed decoding CloudConfig's kubeconfig from base64")
	}

	// unmarshall CloudConfig's kubeconfig into struct
	adminKubeConfig := common.KubeConfig{}
	err = yaml.Unmarshal(adminConfig, &adminKubeConfig)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed parsing CloudConfig's kubeconfig yaml")
	}

	// all data needed for final kubeconfig:
	signedCert := cluster.Specification.Certificate
	clusterCert := adminKubeConfig.Clusters[0].ClusterDef.CertificateAuthorityData
	clusterAddr := adminKubeConfig.Clusters[0].ClusterDef.Server
	namespace := lc.Specification.NameSpace
	userName := lc.Specification.User.UserName
	contextName := userName + "@" + clusterReference

	kubeconfig := common.KubeConfig{
		ApiVersion: "v1",
		Kind:       "Config",
		Clusters: []common.KubeCluster{
			common.KubeCluster{
				ClusterName: clusterReference,
				ClusterDef: common.KubeClusterDef{
					CertificateAuthorityData: clusterCert,
					Server:                   clusterAddr,
				},
			},
		},
		Contexts: []common.KubeContext{
			common.KubeContext{
				ContextName: contextName,
				ContextDef: common.KubeContextDef{
					Cluster:   clusterReference,
					Namespace: namespace,
					User:      userName,
				},
			},
		},
		CurrentContext: contextName,
		Preferences:    map[string]string{},
		Users: []common.KubeUser{
			common.KubeUser{
				UserName: userName,
				UserDef: common.KubeUserDef{
					ClientCertificateData: signedCert,
					ClientKeyData:         privateKey.KeyValue,
				},
			},
		},
	}

	yaml, err := yaml.Marshal(&kubeconfig)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed marshalling user kubeconfig into yaml")
	}

	// now that we have the L1 kubeconfig for this L1 logical cloud,
	// let's give it to rsync so it can get stored in the right place
	_, err = ccc.CreateCloudConfig(
		ctx,
		cluster.Specification.ClusterProvider,
		cluster.Specification.ClusterName,
		lc.Specification.Level,
		lc.Specification.NameSpace,
		base64.StdEncoding.EncodeToString(yaml))

	if err != nil {
		if err.Error() != "CloudConfig already exists" {
			return "", pkgerrors.Wrap(err, "Failed creating a new kubeconfig in rsync's CloudConfig")
		}
	}

	return string(yaml), nil
}
