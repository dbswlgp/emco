// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// Quota contains the parameters needed for a Quota
type Quota struct {
	MetaData QMetaDataList `json:"metadata"`
	// Specification QSpec         `json:"spec"`
	Specification map[string]string `json:"spec"`
}

// MetaData contains the parameters needed for metadata
type QMetaDataList struct {
	QuotaName   string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// TODO: use QSpec fields to validate quota keys
// Spec contains the parameters needed for spec
type QSpec struct {
	LimitsCPU                   string `json:"limits.cpu"`
	LimitsMemory                string `json:"limits.memory"`
	RequestsCPU                 string `json:"requests.cpu"`
	RequestsMemory              string `json:"requests.memory"`
	RequestsStorage             string `json:"requests.storage"`
	LimitsEphemeralStorage      string `json:"limits.ephemeral.storage"`
	PersistentVolumeClaims      string `json:"persistentvolumeclaims"`
	Pods                        string `json:"pods"`
	ConfigMaps                  string `json:"configmaps"`
	ReplicationControllers      string `json:"replicationcontrollers"`
	ResourceQuotas              string `json:"resourcequotas"`
	Services                    string `json:"services"`
	ServicesLoadBalancers       string `json:"services.loadbalancers"`
	ServicesNodePorts           string `json:"services.nodeports"`
	Secrets                     string `json:"secrets"`
	CountReplicationControllers string `json:"count/replicationcontrollers"`
	CountDeploymentsApps        string `json:"count/deployments.apps"`
	CountReplicasetsApps        string `json:"count/replicasets.apps"`
	CountStatefulSets           string `json:"count/statefulsets.apps"`
	CountJobsBatch              string `json:"count/jobs.batch"`
	CountCronJobsBatch          string `json:"count/cronjobs.batch"`
	CountDeploymentsExtensions  string `json:"count/deployments.extensions"`
}

// QuotaKey is the key structure that is used in the database
type QuotaKey struct {
	Project          string `json:"project"`
	LogicalCloudName string `json:"logicalCloud"`
	QuotaName        string `json:"clusterQuota"`
}

// QuotaManager is an interface that exposes the connection
// functionality
type QuotaManager interface {
	CreateQuota(ctx context.Context, project, logicalCloud string, c Quota) (Quota, error)
	GetQuota(ctx context.Context, project, logicalCloud, name string) (Quota, error)
	GetAllQuotas(ctx context.Context, project, logicalCloud string) ([]Quota, error)
	DeleteQuota(ctx context.Context, project, logicalCloud, name string) error
	UpdateQuota(ctx context.Context, project, logicalCloud, name string, c Quota) (Quota, error)
}

// QuotaClient implements the QuotaManager
// It will also be used to maintain some localized state
type QuotaClient struct {
	storeName string
	tagMeta   string
}

// QuotaClient returns an instance of the QuotaClient
// which implements the QuotaManager
func NewQuotaClient() *QuotaClient {
	return &QuotaClient{
		storeName: "resources",
		tagMeta:   "data",
	}
}

// Create entry for the quota resource in the database
func (v *QuotaClient) CreateQuota(ctx context.Context, project, logicalCloud string, c Quota) (Quota, error) {

	//Construct key consisting of name
	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        c.MetaData.QuotaName,
	}
	lcClient := NewLogicalCloudClient()

	//Check if Logical Cloud Level 0 & then avoid creating Quotas
	lc, err := lcClient.Get(ctx, project, logicalCloud)
	if err != nil {
		return Quota{}, err
	}
	if lc.Specification.Level == "0" {
		return Quota{}, pkgerrors.New("Cluster Quotas not allowed for Logical Cloud Level 0")
	}
	//Check if this Quota already exists
	_, err = v.GetQuota(ctx, project, logicalCloud, c.MetaData.QuotaName)
	if err == nil {
		return Quota{}, pkgerrors.New("Quota already exists")
	}

	err = db.DBconn.Insert(ctx, v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return Quota{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return c, nil
}

// Get returns Quota for corresponding quota name
func (v *QuotaClient) GetQuota(ctx context.Context, project, logicalCloud, quotaName string) (Quota, error) {

	//Construct the composite key to select the entry
	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        quotaName,
	}
	value, err := db.DBconn.Find(ctx, v.storeName, key, v.tagMeta)
	if err != nil {
		return Quota{}, err
	}

	if len(value) == 0 {
		return Quota{}, pkgerrors.New("Cluster Quota not found")
	}

	//value is a byte array
	if value != nil {
		q := Quota{}
		err = db.DBconn.Unmarshal(value[0], &q)
		if err != nil {
			return Quota{}, err
		}
		return q, nil
	}

	return Quota{}, pkgerrors.New("Unknown Error")
}

// GetAll returns all cluster quotas in the logical cloud
func (v *QuotaClient) GetAllQuotas(ctx context.Context, project, logicalCloud string) ([]Quota, error) {
	//Construct the composite key to select the entry
	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        "",
	}
	var resp []Quota
	values, err := db.DBconn.Find(ctx, v.storeName, key, v.tagMeta)
	if err != nil {
		return []Quota{}, err
	}

	for _, value := range values {
		q := Quota{}
		err = db.DBconn.Unmarshal(value, &q)
		if err != nil {
			return []Quota{}, err
		}
		resp = append(resp, q)
	}

	return resp, nil
}

// Delete the Quota entry from database
func (v *QuotaClient) DeleteQuota(ctx context.Context, project, logicalCloud, quotaName string) error {
	//Construct the composite key to select the entry
	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        quotaName,
	}
	err := db.DBconn.Remove(ctx, v.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Quota")
	}
	return nil
}

// Update an entry for the Quota in the database
func (v *QuotaClient) UpdateQuota(ctx context.Context, project, logicalCloud, quotaName string, c Quota) (Quota, error) {

	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        quotaName,
	}
	//Check quota URL name against the quota json name
	if c.MetaData.QuotaName != quotaName {
		return Quota{}, pkgerrors.New("Update Error - Quota name mismatch")
	}
	//Check if this Quota exists
	_, err := v.GetQuota(ctx, project, logicalCloud, quotaName)
	if err != nil {
		return Quota{}, err
	}
	err = db.DBconn.Insert(ctx, v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return Quota{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}
