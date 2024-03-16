// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	"context"
	pkgerrors "github.com/pkg/errors"
	migrationModel "gitlab.com/project-emco/core/emco-base/src/migration/pkg/model"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

/*
AddAppIntent ... AddAppIntent adds a given consumenr to the hpa-intent-name and stores in the db.
Other input parameters for it - projectName, compositeAppName, version, DeploymentIntentgroupName, intentName
*/
func (c *MigrationClient) AddAppIntent(ctx context.Context, a migrationModel.MigrationAppIntent, p string, ca string, v string, di string, i string, exists bool) (migrationModel.MigrationAppIntent, error) {
	//Check for the Consumer already exists here.
	res, dependentErrStaus, err := c.GetAppIntent(ctx, a.MetaData.Name, p, ca, v, di, i)
	if err != nil && dependentErrStaus == true {
		log.Error("AddAppIntent ... Consumer dependency check failed", log.Fields{"migrationAppIntent": a.MetaData.Name, "err": err, "res-received": res})
		return migrationModel.MigrationAppIntent{}, err
	} else if err == nil && !exists {
		log.Error("AddAppIntent ... Consumer already exists", log.Fields{"migrationAppIntent": a.MetaData.Name, "err": err, "res-received": res})
		return migrationModel.MigrationAppIntent{}, pkgerrors.New("Consumer already exists")
	}
	dbKey := MigrationAppIntentKey{
		AppIntentName:         a.MetaData.Name,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	log.Info("AddAppIntent ... Creating DB entry entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "migrationIntent": i, "migrationAppIntent": a.MetaData.Name})
	err = db.DBconn.Insert(ctx, c.db.StoreName, dbKey, nil, c.db.TagMetaData, a)
	if err != nil {
		log.Error("AddAppIntent ...  DB Error .. Creating DB entry error", log.Fields{"migrationAppIntent": a.MetaData.Name, "err": err})
		return migrationModel.MigrationAppIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}
	return a, nil
}

func (c *MigrationClient) UpdateAppIntent(ctx context.Context, migrationIntentName string, migrationAppIntentName string, p string, ca string, v string, di string, fieldupdate string) (/*migrationModel.DeploymentMigrationIntent,*/ error) {

        dbKey := MigrationAppIntentKey{
		AppIntentName:         migrationAppIntentName,
                IntentName:            migrationIntentName,
                Project:               p,
                CompositeApp:          ca,
                Version:               v,
                DeploymentIntentGroup: di,
        }

        log.Info("AddIntent ... Creating DB entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "migrationIntent": migrationIntentName, "migrationAppIntent": migrationAppIntentName})
        //err = db.DBconn.Insert(ctx, c.db.StoreName, dbKey, nil, c.db.TagMetaData, a)

        err := db.DBconn.UpdateIntentField(ctx, c.db.StoreName, dbKey, "data.metadata.description", fieldupdate)

        if err != nil {
                log.Error("UpdateIntent ... DB Error .. Updating DB entry error", log.Fields{"StoreName": c.db.StoreName, "akey": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "migrationIntent": migrationIntentName})
                return pkgerrors.Wrap(err, "Updating DB Entry")
        }
        return nil
}



/*
GetAppIntent ... takes in an IntentName, ProjectName, CompositeAppName, Version, DeploymentIntentGroup and intentName.
It returns the Consumer.
*/
func (c *MigrationClient) GetAppIntent(ctx context.Context, cn string, p string, ca string, v string, di string, i string) (migrationModel.MigrationAppIntent, bool, error) {

	dbKey := MigrationAppIntentKey{
		AppIntentName:         cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetAppIntent ... DB Error .. Get Consumer error", log.Fields{"migrationAppIntent": cn})
		return migrationModel.MigrationAppIntent{}, false, err
	}

	if len(result) == 0 {
		return migrationModel.MigrationAppIntent{}, false, pkgerrors.New("Consumer not found")
	}

	if result != nil {
		a := migrationModel.MigrationAppIntent{}
		err = db.DBconn.Unmarshal(result[0], &a)
		if err != nil {
			log.Error("GetAppIntent ... Unmarshalling  HpaConsumer error", log.Fields{"migrationAppIntent": cn})
			return migrationModel.MigrationAppIntent{}, false, err
		}
		return a, false, nil
	}
	return migrationModel.MigrationAppIntent{}, false, pkgerrors.New("Unknown Error")
}

/*
GetAllAppIntents ... takes in projectName, CompositeAppName, CompositeAppVersion, DeploymentGroup,
DeploymentIntentName . It returns ListOfConsumers.
*/
func (c MigrationClient) GetAllAppIntents(ctx context.Context, p, ca, v, di, i string) ([]migrationModel.MigrationAppIntent, error) {

	dbKey := MigrationAppIntentKey{
		AppIntentName:         "",
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetAllAppIntents ... DB Error .. Get HpaConsumers db error", log.Fields{"migrationIntent": i})
		return []migrationModel.MigrationAppIntent{}, err
	}
	log.Info("GetAllAppIntents ... db result", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di})

	var listOfMapOfAppIntents []migrationModel.MigrationAppIntent
	for i := range result {
		a := migrationModel.MigrationAppIntent{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllAppIntents ... Unmarshalling Consumer error.", log.Fields{"index": i, "migrationAppIntent": result[i], "err": err})
				return []migrationModel.MigrationAppIntent{}, err
			}
			listOfMapOfAppIntents = append(listOfMapOfAppIntents, a)
		}
	}

	return listOfMapOfAppIntents, nil
}

/*
GetAppIntentByName ... takes in IntentName, projectName, CompositeAppName, CompositeAppVersion,
deploymentIntentGroupName and intentName returns the list of consumers under the IntentName.
*/
func (c MigrationClient) GetAppIntentByName(ctx context.Context, cn, p, ca, v, di, i string) (migrationModel.MigrationAppIntent, error) {

	dbKey := MigrationAppIntentKey{
		AppIntentName:         cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetAppIntentByName ... DB Error .. Get HpaConsumer error", log.Fields{"migrationAppIntent": cn})
		return migrationModel.MigrationAppIntent{}, err
	}

	if len(result) == 0 {
		return migrationModel.MigrationAppIntent{}, pkgerrors.New("Consumer not found")
	}

	var a migrationModel.MigrationAppIntent
	err = db.DBconn.Unmarshal(result[0], &a)
	if err != nil {
		log.Error("GetAppIntentByName ... Unmarshalling Consumer error", log.Fields{"migrationAppIntent": cn})
		return migrationModel.MigrationAppIntent{}, err
	}
	return a, nil
}

// DeleteAppIntent ... deletes a given intent consumer tied to project, composite app and deployment intent group, intent name
func (c MigrationClient) DeleteAppIntent(ctx context.Context, cn, p string, ca string, v string, di string, i string) error {
	dbKey := MigrationAppIntentKey{
		AppIntentName:         cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	//Check for the Consumer already exists
	_, _, err := c.GetAppIntent(ctx, cn, p, ca, v, di, i)
	if err != nil {
		log.Error("DeleteAppIntent ... migrationAppIntent does not exist", log.Fields{"migrationAppIntent": cn, "err": err})
		return err
	}

	log.Info("DeleteAppIntent ... Delete Migration App Intent entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "app-intent-name": cn})
	err = db.DBconn.Remove(ctx, c.db.StoreName, dbKey)
	if err != nil {
		log.Error("DeleteAppIntent ... DB Error .. Delete Migration App Intent entry error", log.Fields{"err": err, "StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "app-intent-name": cn})
		return pkgerrors.Wrap(err, "DB Error .. Delete Migration App Intent entry error")
	}
	return nil
}
