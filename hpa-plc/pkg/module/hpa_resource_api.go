// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
	hpaModel "gitlab.com/project-emco/core/emco-base/src/hpa-plc/pkg/model"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

/*
 AddResource ... AddResource adds a given consumer to the hpa-hpaIntent and stores in the db.
 Other input parameters for it - projectName, compositeAppName, version, DeploymentIntentgroupName, intentName, consumerName
*/
func (c *HpaPlacementClient) AddResource(ctx context.Context, a hpaModel.HpaResourceRequirement, p string, ca string, v string, di string, i string, cn string, exists bool) (hpaModel.HpaResourceRequirement, error) {
	//Check for the Resource already exists here.
	res, _, err := c.GetResource(ctx, a.MetaData.Name, p, ca, v, di, i, cn)
	if err == nil && !exists {
		log.Error("AddResource ... Resource already exists", log.Fields{"hpaResource": a.MetaData.Name, "err": err, "res-received": res})
		return hpaModel.HpaResourceRequirement{}, pkgerrors.New("Resource already exists")
	}

	dbKey := HpaResourceKey{
		ResourceName:          a.MetaData.Name,
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	log.Info("AddResource ... Creating DB entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "dep-group": di, "hpaIntent": i, "hpaConsumer": cn, "hpaResource": a.MetaData.Name})
	err = db.DBconn.Insert(ctx, c.db.StoreName, dbKey, nil, c.db.TagMetaData, a)
	if err != nil {
		log.Error("AddResource ... DB Error .. Creating DB entry error", log.Fields{"hpaResource": a.MetaData.Name})
		return hpaModel.HpaResourceRequirement{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}
	return a, nil
}

/*
 GetResource ... takes in an ConsumerName, IntentName, ProjectName, CompositeAppName, Version, DeploymentIntentGroup and intentName.
 It returns the Resource.
*/
func (c *HpaPlacementClient) GetResource(ctx context.Context, rn string, p string, ca string, v string, di string, i string, cn string) (hpaModel.HpaResourceRequirement, bool, error) {

	dbKey := HpaResourceKey{
		ResourceName:          rn,
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetResource ... DB Error .. Get Resource error", log.Fields{"hpaResource": rn})
		return hpaModel.HpaResourceRequirement{}, false, err
	}

	if len(result) == 0 {
		return hpaModel.HpaResourceRequirement{}, false, pkgerrors.New("Resource not found")
	}

	if result != nil {
		a := hpaModel.HpaResourceRequirement{}
		err = db.DBconn.Unmarshal(result[0], &a)
		if err != nil {
			log.Error("GetResource ... Unmarshalling HpaResource error", log.Fields{"hpaResource": rn})
			return hpaModel.HpaResourceRequirement{}, false, err
		}
		return a, false, nil
	}
	return hpaModel.HpaResourceRequirement{}, false, pkgerrors.New("Unknown Error")
}

/*
 GetAllResources ... takes in projectName, CompositeAppName, CompositeAppVersion, DeploymentGroup,
 DeploymentIntentName, ConsumerName . It returns ListOfResources.
*/
func (c HpaPlacementClient) GetAllResources(ctx context.Context, p, ca, v, di, i, cn string) ([]hpaModel.HpaResourceRequirement, error) {

	dbKey := HpaResourceKey{
		ResourceName:          "",
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetAllResources ... DB Error .. Get HpaResources db error", log.Fields{"hpaConsumer": cn})
		return []hpaModel.HpaResourceRequirement{}, err
	}
	log.Info("GetAllResources ... db result", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "dep-group": di, "hpaConsumer": cn})

	var listOfMapOfResources []hpaModel.HpaResourceRequirement
	for i := range result {
		a := hpaModel.HpaResourceRequirement{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllResources ... Unmarshalling Resources error", log.Fields{"hpaConsumer": cn})
				return []hpaModel.HpaResourceRequirement{}, err
			}
			listOfMapOfResources = append(listOfMapOfResources, a)
		}
	}
	return listOfMapOfResources, nil
}

/*
 GetResourceByName ... takes in IntentName, projectName, CompositeAppName, CompositeAppVersion,
 deploymentIntentGroupName, intentName and consumerName returns the list of resource under the consumerName.
*/
func (c HpaPlacementClient) GetResourceByName(ctx context.Context, rn, p, ca, v, di, i, cn string) (hpaModel.HpaResourceRequirement, error) {

	dbKey := HpaResourceKey{
		ResourceName:          rn,
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetResourceByName ... DB Error .. Get HpaResource error", log.Fields{"hpaResource": rn})
		return hpaModel.HpaResourceRequirement{}, err
	}

	if len(result) == 0 {
		return hpaModel.HpaResourceRequirement{}, pkgerrors.New("Resource not found")
	}

	var a hpaModel.HpaResourceRequirement
	err = db.DBconn.Unmarshal(result[0], &a)
	if err != nil {
		log.Error("GetResourceByName ... Unmarshalling Resource error", log.Fields{"hpaResource": rn})
		return hpaModel.HpaResourceRequirement{}, err
	}
	return a, nil
}

// DeleteResource ... deletes a given resource tied to project, composite app and deployment intent group, intent name, consumer name
func (c HpaPlacementClient) DeleteResource(ctx context.Context, rn string, p string, ca string, v string, di string, i string, cn string) error {
	dbKey := HpaResourceKey{
		ResourceName:          rn,
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	//Check for the Resource already exists
	_, _, err := c.GetResource(ctx, rn, p, ca, v, di, i, cn)
	if err != nil {
		log.Error("DeleteResource ... Resource does not exist", log.Fields{"hpaResource": rn, "err": err})
		return err
	}

	log.Info("DeleteResource ... Delete Hpa Consumer entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "dep-group": di, "hpaIntent": i, "hpaConsumer": cn, "hpaResource": rn})
	err = db.DBconn.Remove(ctx, c.db.StoreName, dbKey)
	if err != nil {
		log.Error("DeleteResource ... DB Error .. Delete Hpa Resource entry error", log.Fields{"err": err, "StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "dep-group": di, "hpaIntent": i, "hpaConsumer": cn})
		return pkgerrors.Wrap(err, "DB Error .. Delete Hpa Resource entry error")
	}
	return nil
}
