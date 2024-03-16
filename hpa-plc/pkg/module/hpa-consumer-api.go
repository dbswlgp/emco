// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	"context"
	pkgerrors "github.com/pkg/errors"
	hpaModel "gitlab.com/project-emco/core/emco-base/src/hpa-plc/pkg/model"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

/*
AddConsumer ... AddConsumer adds a given consumenr to the hpa-intent-name and stores in the db.
Other input parameters for it - projectName, compositeAppName, version, DeploymentIntentgroupName, intentName
*/
func (c *HpaPlacementClient) AddConsumer(ctx context.Context, a hpaModel.HpaResourceConsumer, p string, ca string, v string, di string, i string, exists bool) (hpaModel.HpaResourceConsumer, error) {
	//Check for the Consumer already exists here.
	res, dependentErrStaus, err := c.GetConsumer(ctx, a.MetaData.Name, p, ca, v, di, i)
	if err != nil && dependentErrStaus == true {
		log.Error("AddConsumer ... Consumer dependency check failed", log.Fields{"hpaConsumer": a.MetaData.Name, "err": err, "res-received": res})
		return hpaModel.HpaResourceConsumer{}, err
	} else if err == nil && !exists {
		log.Error("AddConsumer ... Consumer already exists", log.Fields{"hpaConsumer": a.MetaData.Name, "err": err, "res-received": res})
		return hpaModel.HpaResourceConsumer{}, pkgerrors.New("Consumer already exists")
	}
	dbKey := HpaConsumerKey{
		ConsumerName:          a.MetaData.Name,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	log.Info("AddConsumer ... Creating DB entry entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "hpaIntent": i, "hpaConsumer": a.MetaData.Name})
	err = db.DBconn.Insert(ctx, c.db.StoreName, dbKey, nil, c.db.TagMetaData, a)
	if err != nil {
		log.Error("AddConsumer ...  DB Error .. Creating DB entry error", log.Fields{"hpaConsumer": a.MetaData.Name, "err": err})
		return hpaModel.HpaResourceConsumer{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}
	return a, nil
}

/*
GetConsumer ... takes in an IntentName, ProjectName, CompositeAppName, Version, DeploymentIntentGroup and intentName.
It returns the Consumer.
*/
func (c *HpaPlacementClient) GetConsumer(ctx context.Context, cn string, p string, ca string, v string, di string, i string) (hpaModel.HpaResourceConsumer, bool, error) {

	dbKey := HpaConsumerKey{
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetConsumer ... DB Error .. Get Consumer error", log.Fields{"hpaConsumer": cn})
		return hpaModel.HpaResourceConsumer{}, false, err
	}

	if len(result) == 0 {
		return hpaModel.HpaResourceConsumer{}, false, pkgerrors.New("Consumer not found")
	}

	if result != nil {
		a := hpaModel.HpaResourceConsumer{}
		err = db.DBconn.Unmarshal(result[0], &a)
		if err != nil {
			log.Error("GetConsumer ... Unmarshalling  HpaConsumer error", log.Fields{"hpaConsumer": cn})
			return hpaModel.HpaResourceConsumer{}, false, err
		}
		return a, false, nil
	}
	return hpaModel.HpaResourceConsumer{}, false, pkgerrors.New("Unknown Error")
}

/*
GetAllConsumers ... takes in projectName, CompositeAppName, CompositeAppVersion, DeploymentGroup,
DeploymentIntentName . It returns ListOfConsumers.
*/
func (c HpaPlacementClient) GetAllConsumers(ctx context.Context, p, ca, v, di, i string) ([]hpaModel.HpaResourceConsumer, error) {

	dbKey := HpaConsumerKey{
		ConsumerName:          "",
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetAllConsumers ... DB Error .. Get HpaConsumers db error", log.Fields{"hpaIntent": i})
		return []hpaModel.HpaResourceConsumer{}, err
	}
	log.Info("GetAllConsumers ... db result", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di})

	var listOfMapOfConsumers []hpaModel.HpaResourceConsumer
	for i := range result {
		a := hpaModel.HpaResourceConsumer{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllConsumers ... Unmarshalling Consumer error.", log.Fields{"index": i, "hpaConsumer": result[i], "err": err})
				return []hpaModel.HpaResourceConsumer{}, err
			}
			listOfMapOfConsumers = append(listOfMapOfConsumers, a)
		}
	}

	return listOfMapOfConsumers, nil
}

/*
GetConsumerByName ... takes in IntentName, projectName, CompositeAppName, CompositeAppVersion,
deploymentIntentGroupName and intentName returns the list of consumers under the IntentName.
*/
func (c HpaPlacementClient) GetConsumerByName(ctx context.Context, cn, p, ca, v, di, i string) (hpaModel.HpaResourceConsumer, error) {

	dbKey := HpaConsumerKey{
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetConsumerByName ... DB Error .. Get HpaConsumer error", log.Fields{"hpaConsumer": cn})
		return hpaModel.HpaResourceConsumer{}, err
	}

	if len(result) == 0 {
		return hpaModel.HpaResourceConsumer{}, pkgerrors.New("Consumer not found")
	}

	var a hpaModel.HpaResourceConsumer
	err = db.DBconn.Unmarshal(result[0], &a)
	if err != nil {
		log.Error("GetConsumerByName ... Unmarshalling Consumer error", log.Fields{"hpaConsumer": cn})
		return hpaModel.HpaResourceConsumer{}, err
	}
	return a, nil
}

// DeleteConsumer ... deletes a given intent consumer tied to project, composite app and deployment intent group, intent name
func (c HpaPlacementClient) DeleteConsumer(ctx context.Context, cn, p string, ca string, v string, di string, i string) error {
	dbKey := HpaConsumerKey{
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	//Check for the Consumer already exists
	_, _, err := c.GetConsumer(ctx, cn, p, ca, v, di, i)
	if err != nil {
		log.Error("DeleteConsumer ... hpaConsumer does not exist", log.Fields{"hpaConsumer": cn, "err": err})
		return err
	}

	log.Info("DeleteConsumer ... Delete Hpa Consumer entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "consumer-name": cn})
	err = db.DBconn.Remove(ctx, c.db.StoreName, dbKey)
	if err != nil {
		log.Error("DeleteConsumer ... DB Error .. Delete Hpa Consumer entry error", log.Fields{"err": err, "StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "consumer-name": cn})
		return pkgerrors.Wrap(err, "DB Error .. Delete Hpa Consumer entry error")
	}
	return nil
}
