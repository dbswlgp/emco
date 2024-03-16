// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2022 Intel Corporation

package module

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"

	"context"

	pkgerrors "github.com/pkg/errors"
)

// LogicalCloudManager is an interface that exposes the connection
// functionality
type LogicalCloudManager interface {
	Create(ctx context.Context, project string, c common.LogicalCloud) (common.LogicalCloud, error)
	Get(ctx context.Context, project, name string) (common.LogicalCloud, error)
	GetAll(ctx context.Context, project string) ([]common.LogicalCloud, error)
	GetState(ctx context.Context, p string, lc string) (state.StateInfo, error)
	Delete(ctx context.Context, project, name string) error
	GenericStatus(ctx context.Context, project, name, qStatusInstance, qType, qOutput string, fClusters, fResources []string) (status.StatusResult, error)
	StatusClusters(ctx context.Context, p, lc, qStatusInstance string) (status.LogicalCloudClustersStatus, error)
	StatusResources(ctx context.Context, p, lc, qStatusInstance, qType string, fClusters []string) (status.LogicalCloudResourcesStatus, error)
	Status(ctx context.Context, p, lc, qStatusInstance, qType, qOutput string, fClusters, fResources []string) (status.LogicalCloudStatus, error)
	UpdateLogicalCloud(ctx context.Context, project, name string, c common.LogicalCloud) (common.LogicalCloud, error)
	UpdateInstantiation(ctx context.Context, project, name string, c common.LogicalCloud) (common.LogicalCloud, error)
}

// LogicalCloudClient implements the LogicalCloudManager
// It will also be used to maintain some localized state
type LogicalCloudClient struct {
	storeName string
	tagMeta   string
	tagState  string
}

// LogicalCloudClient returns an instance of the LogicalCloudClient
// which implements the LogicalCloudManager
func NewLogicalCloudClient() *LogicalCloudClient {
	return &LogicalCloudClient{
		storeName: "resources",
		tagMeta:   "data",
		tagState:  "stateInfo",
	}
}

// Create entry for the logical cloud resource in the database
func (v *LogicalCloudClient) Create(ctx context.Context, project string, c common.LogicalCloud) (common.LogicalCloud, error) {

	//Construct key consisting of name
	key := common.LogicalCloudKey{
		Project:          project,
		LogicalCloudName: c.MetaData.Name,
	}

	//Check if this Logical Cloud already exists
	_, err := v.Get(ctx, project, c.MetaData.Name)
	if err == nil {
		return common.LogicalCloud{}, pkgerrors.New("Logical Cloud already exists")
	}

	// if Logical Cloud Level is not specified, it defaults to 1:
	if c.Specification.Level == "" {
		c.Specification.Level = "1"
	}

	err = db.DBconn.Insert(ctx, v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return common.LogicalCloud{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	// Generate and store Logical Cloud user private key (Level 1 only)
	if c.Specification.Level == "1" {
		pk, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return common.LogicalCloud{}, pkgerrors.New("Failure generating Logical Cloud user key")
		}

		pkData := base64.StdEncoding.EncodeToString(pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(pk),
			},
		))

		privKey := common.PrivateKey{
			KeyValue: string(pkData),
		}

		err = db.DBconn.Insert(ctx, v.storeName, key, nil, "privatekey", privKey)
		if err != nil {
			return common.LogicalCloud{}, pkgerrors.New("Failure storing Logical Cloud user key")
		}
	}

	// Add the state info record (initial state)
	s := state.StateInfo{}
	a := state.ActionEntry{
		State:     state.StateEnum.Created,
		ContextId: "",
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)

	err = db.DBconn.Insert(ctx, v.storeName, key, nil, v.tagState, s)
	if err != nil {
		return common.LogicalCloud{}, pkgerrors.Wrap(err, "Error updating the state info of the LogicalCloud: "+c.MetaData.Name)
	}

	return c, nil
}

// Get returns Logical Cloud corresponding to logical cloud name
func (v *LogicalCloudClient) Get(ctx context.Context, project, logicalCloudName string) (common.LogicalCloud, error) {

	//Construct the composite key to select the entry
	key := common.LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloudName,
	}
	value, err := db.DBconn.Find(ctx, v.storeName, key, v.tagMeta)
	if err != nil {
		return common.LogicalCloud{}, err
	}

	if len(value) == 0 {
		return common.LogicalCloud{}, pkgerrors.New("Logical Cloud not found")
	}

	//value is a byte array
	if value != nil {
		lc := common.LogicalCloud{}
		err = db.DBconn.Unmarshal(value[0], &lc)
		if err != nil {
			return common.LogicalCloud{}, err
		}
		return lc, nil
	}

	return common.LogicalCloud{}, pkgerrors.New("Unknown Error")
}

// GetAll returns Logical Clouds in the project
func (v *LogicalCloudClient) GetAll(ctx context.Context, project string) ([]common.LogicalCloud, error) {

	//Construct the composite key to select the entry
	key := common.LogicalCloudKey{
		Project:          project,
		LogicalCloudName: "",
	}

	var resp []common.LogicalCloud
	values, err := db.DBconn.Find(ctx, v.storeName, key, v.tagMeta)
	if err != nil {
		return []common.LogicalCloud{}, err
	}

	for _, value := range values {
		lc := common.LogicalCloud{}
		err = db.DBconn.Unmarshal(value, &lc)
		if err != nil {
			return []common.LogicalCloud{}, err
		}
		resp = append(resp, lc)
	}

	return resp, nil
}

// GetState returns the LogicalCloud StateInfo with a given logical cloud name and project
func (v *LogicalCloudClient) GetState(ctx context.Context, p string, lc string) (state.StateInfo, error) {

	key := common.LogicalCloudKey{
		Project:          p,
		LogicalCloudName: lc,
	}

	result, err := db.DBconn.Find(ctx, v.storeName, key, v.tagState)
	if err != nil {
		return state.StateInfo{}, err
	}

	if len(result) == 0 {
		return state.StateInfo{}, pkgerrors.New("LogicalCloud StateInfo not found")
	}

	if result != nil {
		s := state.StateInfo{}
		err = db.DBconn.Unmarshal(result[0], &s)
		if err != nil {
			return state.StateInfo{}, err
		}
		return s, nil
	}

	return state.StateInfo{}, pkgerrors.New("Unknown Error")
}

// Delete the Logical Cloud entry from database
func (v *LogicalCloudClient) Delete(ctx context.Context, project, logicalCloudName string) error {

	//Construct the composite key to select the entry
	key := common.LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloudName,
	}
	//Check if this Logical Cloud exists
	_, err := v.Get(ctx, project, logicalCloudName)
	if err != nil {
		return err
	}

	// Check if there was a previous appCtx for this logical cloud
	s, err := v.GetState(ctx, project, logicalCloudName)
	if err != nil {
		return err
	}
	cid := state.GetLastContextIdFromStateInfo(s)

	// If there's no appCtx for Logical Cloud, just go ahead and delete it now
	if cid == "" {
		err = db.DBconn.Remove(ctx, v.storeName, key)
		if err != nil {
			return pkgerrors.Wrap(err, "Error when deleting Logical Cloud (scenario with no context)")
		}
		return nil
	}

	ac, err := state.GetAppContextFromId(ctx, cid)
	if err != nil {
		return err
	}

	// Make sure rsync status for this logical cloud is Terminated,
	// otherwise we can't re-instantiate logical cloud yet
	acStatus, err := state.GetAppContextStatus(ctx, cid) // new from state
	if err != nil {
		return err
	}
	switch acStatus.Status {
	case appcontext.AppContextStatusEnum.Terminating:
		log.Error("The Logical Cloud can't be deleted yet, it is being terminated", log.Fields{"logicalcloud": logicalCloudName})
		return pkgerrors.New("The Logical Cloud can't be deleted yet, it is being terminated")
	case appcontext.AppContextStatusEnum.Instantiated:
		log.Error("The Logical Cloud is instantiated, please terminate first", log.Fields{"logicalcloud": logicalCloudName})
		return pkgerrors.New("The Logical Cloud is instantiated, please terminate first")
	case appcontext.AppContextStatusEnum.Instantiating:
		log.Error("The Logical Cloud is instantiating, please wait and then terminate", log.Fields{"logicalcloud": logicalCloudName})
		return pkgerrors.New("The Logical Cloud is instantiating, please wait and then terminate")
	case appcontext.AppContextStatusEnum.InstantiateFailed:
		log.Error("The Logical Cloud has failed instantiating, for safety please terminate and try again", log.Fields{"logicalcloud": logicalCloudName})
		return pkgerrors.New("The Logical Cloud has failed instantiating, for safety please terminate and try again")
	case appcontext.AppContextStatusEnum.TerminateFailed:
		log.Info("The Logical Cloud has failed terminating, proceeding with the delete operation", log.Fields{"logicalcloud": logicalCloudName})
		// try to delete anyway since termination failed
		fallthrough
	case appcontext.AppContextStatusEnum.Terminated:
		// remove the appcontext
		err := ac.DeleteCompositeApp(ctx)
		if err != nil {
			log.Error("Error deleting AppContext CompositeApp Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.Wrap(err, "Error deleting AppContext CompositeApp Logical Cloud")
		}

		stateVal, err := state.GetCurrentStateFromStateInfo(s)
		if err != nil {
			return pkgerrors.Wrap(err, "Error getting current state from Logical Cloud StateInfo")
		}
		if stateVal == state.StateEnum.Terminated || stateVal == state.StateEnum.TerminateStopped {
			// remove all associated app contexts
			for _, id := range state.GetContextIdsFromStateInfo(s) {
				log.Debug("Delete(): attempting to get appcontext", log.Fields{"id": id})
				appCtx, err := state.GetAppContextFromId(ctx, id)
				if err == nil {
					err = appCtx.DeleteCompositeApp(ctx)
					if err != nil {
						return pkgerrors.Wrap(err, "Error deleting appcontext for Logical Cloud")
					}
					log.Debug("Delete(): deleted composite app", log.Fields{"id": id})
				} else {
					log.Debug("Delete(): failed getting appcontext, silently skipping", log.Fields{"id": id})
				}
			}
		}

		err = db.DBconn.Remove(ctx, v.storeName, key)
		if err != nil {
			log.Error("Error when deleting Logical Cloud (scenario with Terminated status)", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.Wrap(err, "Error when deleting Logical Cloud (scenario with Terminated status)")
		}
		log.Info("Deleted Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
		return nil
	default:
		log.Error("The Logical Cloud isn't in an expected status so not taking any action", log.Fields{"logicalcloud": logicalCloudName, "status": acStatus.Status})
		return pkgerrors.New("The Logical Cloud isn't in an expected status so not taking any action")
	}
}

// Update an entry for the Logical Cloud in the database
func (v *LogicalCloudClient) UpdateLogicalCloud(ctx context.Context, project, logicalCloudName string, c common.LogicalCloud) (common.LogicalCloud, error) {

	key := common.LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloudName,
	}
	// Check for mismatch, logicalCloudName and payload logical cloud name
	if c.MetaData.Name != logicalCloudName {
		return common.LogicalCloud{}, pkgerrors.New("Logical Cloud name mismatch")
	}
	//Check if this Logical Cloud exists
	_, err := v.Get(ctx, project, logicalCloudName)
	if err != nil {
		return common.LogicalCloud{}, err
	}
	err = db.DBconn.Insert(ctx, v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return common.LogicalCloud{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}

// Update an entry for the Logical Cloud in the database
func (v *LogicalCloudClient) UpdateInstantiation(ctx context.Context, project, logicalCloudName string, c common.LogicalCloud) (common.LogicalCloud, error) {

	key := common.LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloudName,
	}

	//Check if this Logical Cloud exists
	logicalCloud, err := v.Get(ctx, project, logicalCloudName)
	if err != nil {
		return common.LogicalCloud{}, err
	}

	// If Logical Cloud was already instantiated, then prepare new appcontext to give to rsync
	lcStateInfo, err := v.GetState(ctx, project, logicalCloudName)
	if err != nil {
		return common.LogicalCloud{}, err
	}
	log.Debug("", log.Fields{"lcStateInfo": lcStateInfo})
	oldCID := state.GetLastContextIdFromStateInfo(lcStateInfo)
	if oldCID != "" {
		log.Debug("", log.Fields{"oldCID": oldCID})

		// Since there's a appCtx associated, if the logical cloud isn't fully Terminated then prevent
		// clusters from being added since this is a functional scenario not currently supported
		contextStatus, err := state.GetAppContextStatus(ctx, oldCID) // new from state
		if err != nil {
			return common.LogicalCloud{}, pkgerrors.New("Logical Cloud is not in a state where a cluster can be created")
		}
		// If Logical Cloud is instantiated, it's safe to proceed with an updated-instantiation
		if contextStatus.Status == appcontext.AppContextStatusEnum.Instantiated {
			// We need to know which clusters to instantiate on
			clusterList, err := NewClusterClient().GetAllClusters(ctx, project, logicalCloudName)
			if err != nil {
				return common.LogicalCloud{}, err
			}

			// We need to know the Level as that influences how to build the appcontext
			level := logicalCloud.Specification.Level

			var newCID string
			// Prepare new appcontext to replace previous one
			if level == "1" {
				// For L1, we need to know what quotas and user permissions to use
				quotaList, err := NewQuotaClient().GetAllQuotas(ctx, project, logicalCloudName)
				if err != nil {
					return common.LogicalCloud{}, err
				}
				userPermissionList, err := NewUserPermissionClient().GetAllUserPerms(ctx, project, logicalCloudName)
				if err != nil {
					return common.LogicalCloud{}, err
				}
				_, newCID, err = blindInstantiateL1(ctx, oldCID, project, logicalCloud, v, clusterList, quotaList, userPermissionList)
				log.Debug("", log.Fields{"newCID": newCID})
			} else if level == "0" {
				_, newCID, err = blindInstantiateL0(ctx, project, logicalCloud, v, clusterList)
				log.Debug("", log.Fields{"newCID": newCID})
			}
			if err != nil {
				return common.LogicalCloud{}, err
			}

			// Update DB Status CID
			err = state.UpdateAppContextStatusContextID(ctx, newCID, oldCID)
			if err != nil {
				return common.LogicalCloud{}, err
			}

			// Call rsync to update Logical Cloud in clusters (calculate differences and bring clusters up to DCM state)
			err = callRsyncUpdate(ctx, oldCID, newCID)
			if err != nil {
				log.Error("Failed calling rsync update", log.Fields{"err": err})
				return common.LogicalCloud{}, pkgerrors.Wrap(err, "Failed calling rsync update")
			}

			latestRev, err := state.GetLatestRevisionFromStateInfo(lcStateInfo)
			if err != nil {
				log.Error("Latest revision not found", log.Fields{})
				return common.LogicalCloud{}, err
			}
			// TODO: make atomic
			newRev := latestRev + 1

			a := state.ActionEntry{
				State:     state.StateEnum.Updated,
				ContextId: newCID,
				TimeStamp: time.Now(),
				Revision:  newRev,
			}
			lcStateInfo.Actions = append(lcStateInfo.Actions, a)

			err = db.DBconn.Insert(ctx, v.storeName, key, nil, v.tagState, lcStateInfo)
			if err != nil {
				log.Error("Error updating the state info of the LogicalCloud: ", log.Fields{"logicalCloud": logicalCloud})
				return common.LogicalCloud{}, err
			}

			// TODO: enhancement: also check if any L1 cluster actually got added, if not then no need for ReadyNotify:
			if level == "1" {
				// Call rsync grpc streaming api, which launches a goroutine to wait for the response of
				// every cluster (function should know how many clusters are expected and only finish when
				// all respective certificates have been obtained and all kubeconfigs stored in CloudConfig)
				err = callRsyncReadyNotify(ctx, oldCID)
				if err != nil {
					log.Error("Failed calling rsync ready-notify", log.Fields{"err": err})
					return common.LogicalCloud{}, pkgerrors.Wrap(err, "Failed calling rsync ready-notify")
				}
			}
		}
		return common.LogicalCloud{}, nil
	}

	return c, nil
}

func (v *LogicalCloudClient) GenericStatus(ctx context.Context, p, lc, qStatusInstance, qType, qOutput string, fClusters, fResources []string) (status.StatusResult, error) {
	stateInfo, err := v.GetState(ctx, p, lc)
	if err != nil {
		return status.StatusResult{}, pkgerrors.Wrap(err, "LogicalCloud state not found: "+lc)
	}

	qInstance, err := state.GetContextIdForStatusContextId(stateInfo, qStatusInstance)
	if err != nil {
		return status.StatusResult{}, err
	}

	statusResponse, err := status.GenericPrepareStatusResult(ctx, status.LcStatusQuery, stateInfo, qInstance, qType, qOutput, make([]string, 0), fClusters, fResources)
	if err != nil {
		return status.StatusResult{}, err
	}
	statusResponse.Name = lc

	return statusResponse, nil
}

func (v *LogicalCloudClient) StatusClusters(ctx context.Context, p, lc, qStatusInstance string) (status.LogicalCloudClustersStatus, error) {
	lcState, err := v.GetState(ctx, p, lc)
	if err != nil {
		return status.LogicalCloudClustersStatus{}, pkgerrors.Wrap(err, "Logical Cloud state not found")
	}

	statusResponse, err := status.PrepareClustersByAppStatusResult(ctx, lcState, qStatusInstance, []string{"logical-cloud"})
	if err != nil {
		return status.LogicalCloudClustersStatus{}, err
	}
	statusResponse.Name = lc
	lcStatus := status.LogicalCloudClustersStatus{
		Project:             p,
		ClustersByAppResult: statusResponse,
	}

	return lcStatus, nil
}

func (v *LogicalCloudClient) StatusResources(ctx context.Context, p, lc, qStatusInstance, qType string, fClusters []string) (status.LogicalCloudResourcesStatus, error) {
	lcState, err := v.GetState(ctx, p, lc)
	if err != nil {
		return status.LogicalCloudResourcesStatus{}, pkgerrors.Wrap(err, "Logical Cloud state not found")
	}

	statusResponse, err := status.PrepareResourcesByAppStatusResult(ctx, lcState, qStatusInstance, qType, []string{"logical-cloud"}, fClusters)
	if err != nil {
		return status.LogicalCloudResourcesStatus{}, err
	}
	statusResponse.Name = lc
	lcStatus := status.LogicalCloudResourcesStatus{
		Project:              p,
		ResourcesByAppResult: statusResponse,
	}

	return lcStatus, nil
}

func (v *LogicalCloudClient) Status(ctx context.Context, p, lc, qStatusInstance, qType, qOutput string, fClusters, fResources []string) (status.LogicalCloudStatus, error) {
	lcState, err := v.GetState(ctx, p, lc)
	if err != nil {
		return status.LogicalCloudStatus{}, pkgerrors.Wrap(err, "Logical Cloud state not found")
	}

	statusResponse, err := status.PrepareStatusResult(ctx, lcState, qStatusInstance, qType, qOutput, make([]string, 0), fClusters, fResources)
	if err != nil {
		return status.LogicalCloudStatus{}, err
	}
	statusResponse.Name = lc
	lcStatus := status.LogicalCloudStatus{
		Project:      p,
		StatusResult: statusResponse,
	}
	if len(statusResponse.Apps) < 1 {
		// there will be no composite app for the logical cloud when it's L0 (admin)
		return lcStatus, nil
	}
	lcStatus.Clusters = statusResponse.Apps[0].Clusters
	lcStatus.StatusResult.Apps = nil

	return lcStatus, nil
}
