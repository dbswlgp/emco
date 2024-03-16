// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"fmt"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	"context"
	pkgerrors "github.com/pkg/errors"
	migrationModel "gitlab.com/project-emco/core/emco-base/src/migration/pkg/model"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

func (c *MigrationClient) AddIntent(ctx context.Context, a migrationModel.DeploymentMigrationIntent, p string, ca string, v string, di string, exists bool) (migrationModel.DeploymentMigrationIntent, error) {
	//Check for the intent already exists here.
	res, dependentErrStaus, err := c.GetIntent(ctx, a.MetaData.Name, p, ca, v, di)
	if err != nil && dependentErrStaus == true {
		log.Error("AddIntent ... Intent dependency check failed", log.Fields{"migrationIntent": a.MetaData.Name, "err": err, "res-received": res})
		return migrationModel.DeploymentMigrationIntent{}, err
	} else if (err == nil) && (!exists) {
		log.Error("AddIntent ... Intent already exists", log.Fields{"migrationIntent": a.MetaData.Name, "err": err, "res-received": res})
		return migrationModel.DeploymentMigrationIntent{}, pkgerrors.New("Intent already exists")
	}

	dbKey := MigrationIntentKey{
		IntentName:            a.MetaData.Name,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	log.Info("AddIntent ... Creating DB entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "migrationIntent": a.MetaData.Name})
	err = db.DBconn.Insert(ctx, c.db.StoreName, dbKey, nil, c.db.TagMetaData, a)
	if err != nil {
		log.Error("AddIntent ... DB Error .. Creating DB entry error", log.Fields{"StoreName": c.db.StoreName, "akey": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "migrationIntent": a.MetaData.Name})
		return migrationModel.DeploymentMigrationIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}
	return a, nil
}



func (c *MigrationClient) UpdateIntent(ctx context.Context, migrationIntentName string, p string, ca string, v string, di string, fieldupdate string) (/*migrationModel.DeploymentMigrationIntent,*/ error) {

	dbKey := MigrationIntentKey{
		IntentName:            migrationIntentName,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	log.Info("AddIntent ... Creating DB entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "migrationIntent": migrationIntentName})
	//err = db.DBconn.Insert(ctx, c.db.StoreName, dbKey, nil, c.db.TagMetaData, a)

	err := db.DBconn.UpdateIntentField(ctx, c.db.StoreName, dbKey, "data.metadata.description", fieldupdate)

	if err != nil {
		log.Error("UpdateIntent ... DB Error .. Updating DB entry error", log.Fields{"StoreName": c.db.StoreName, "akey": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "migrationIntent": migrationIntentName})
		return pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return nil
}


func (c *MigrationClient) GetIntent(ctx context.Context, i string, p string, ca string, v string, di string) (migrationModel.DeploymentMigrationIntent, bool, error) {

	dbKey := MigrationIntentKey{
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetIntent ... DB Error .. Get Intent error", log.Fields{"migrationIntent": i, "err": err})
		return migrationModel.DeploymentMigrationIntent{}, false, err
	}

	if len(result) == 0 {
		return migrationModel.DeploymentMigrationIntent{}, false, pkgerrors.New("Intent not found")
	}

	if result != nil {
		a := migrationModel.DeploymentMigrationIntent{}
		err = db.DBconn.Unmarshal(result[0], &a)
		if err != nil {
			log.Error("GetIntent ... Unmarshalling  Intent error", log.Fields{"migrationIntent": i})
			return migrationModel.DeploymentMigrationIntent{}, false, err
		}
		return a, false, nil
	}
	return migrationModel.DeploymentMigrationIntent{}, false, pkgerrors.New("Unknown Error")
}

func (c MigrationClient) GetAllIntents(ctx context.Context, p string, ca string, v string, di string) ([]migrationModel.DeploymentMigrationIntent, error) {

	dbKey := MigrationIntentKey{
		IntentName:            "",
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetAllIntents ... DB Error .. Get MigrationIntents db error", log.Fields{"StoreName": c.db.StoreName, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "len_result": len(result), "err": err})
		return []migrationModel.DeploymentMigrationIntent{}, err
	}
	log.Info("GetAllIntents ... db result", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di})

	var listOfIntents []migrationModel.DeploymentMigrationIntent
	for i := range result {
		a := migrationModel.DeploymentMigrationIntent{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllIntents ... Unmarshalling MigrationIntents error", log.Fields{"deploymentgroup": di})
				return []migrationModel.DeploymentMigrationIntent{}, err
			}
			listOfIntents = append(listOfIntents, a)
		}
	}
	return listOfIntents, nil
}

func (c MigrationClient) GetAllIntentsByDIG(ctx context.Context, p string, di string) ([]migrationModel.DeploymentMigrationIntent, error) {

	dbKey := MigrationIntentKey{
		IntentName:            "",
		Project:               p,
		CompositeApp:          "",
		Version:               "",
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetAllIntents ... DB Error .. Get MigrationIntents db error", log.Fields{"StoreName": c.db.StoreName, "project": p, "deploymentIntentGroup": di, "len_result": len(result), "err": err})
		return []migrationModel.DeploymentMigrationIntent{}, err
	}
	log.Info("GetAllIntents ... db result", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "deploymentIntentGroup": di })

	var listOfIntents []migrationModel.DeploymentMigrationIntent
	for i := range result {
		a := migrationModel.DeploymentMigrationIntent{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllIntents ... Unmarshalling MigrationIntents error", log.Fields{"deploymentgroup": di})
				return []migrationModel.DeploymentMigrationIntent{}, err
			}
			listOfIntents = append(listOfIntents, a)
		}
	}
	return listOfIntents, nil
}

func (c MigrationClient) GetAllIntentsByApp(ctx context.Context, app string, p string, ca string, v string, di string) ([]migrationModel.DeploymentMigrationIntent, error) {

	dbKey := MigrationIntentKey{
		IntentName:            "",
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)

	if err != nil {
		log.Error("GetAllIntentsByApp .. DB Error", log.Fields{"StoreName": c.db.StoreName, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "len_result": len(result), "err": err})
		return []migrationModel.DeploymentMigrationIntent{}, err
	}
	log.Info("GetAllIntentsByApp ... db result",
		log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "app": app})

	var listOfIntents []migrationModel.DeploymentMigrationIntent
	for i := range result {
		a := migrationModel.DeploymentMigrationIntent{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllIntentsByApp ... Unmarshalling MigrationIntents error", log.Fields{"deploymentgroup": di})
				return []migrationModel.DeploymentMigrationIntent{}, err
			}
			//if a.Spec.AppName == app {
			listOfIntents = append(listOfIntents, a)
			//}

		}
	}

	fmt.Println("\nmigration/pkg/module/hpa-intent-api.go:156")
	fmt.Println("listOfIntents: ", listOfIntents,"\n")

	return listOfIntents, nil
}

func (c MigrationClient) GetIntentByName(ctx context.Context, i string, p string, ca string, v string, di string) (migrationModel.DeploymentMigrationIntent, error) {

	dbKey := MigrationIntentKey{
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetIntentByName ... DB Error .. Get MigrationIntent error", log.Fields{"migrationIntent": i})
		return migrationModel.DeploymentMigrationIntent{}, err
	}

	if len(result) == 0 {
		return migrationModel.DeploymentMigrationIntent{}, pkgerrors.New("Intent not found")
	}

	var a migrationModel.DeploymentMigrationIntent
	err = db.DBconn.Unmarshal(result[0], &a)
	if err != nil {
		log.Error("GetIntentByName ...  Unmarshalling MigrationIntent error", log.Fields{"migrationIntent": i})
		return migrationModel.DeploymentMigrationIntent{}, err
	}
	return a, nil
}

func (c MigrationClient) DeleteIntent(ctx context.Context, i string, p string, ca string, v string, di string) error {
	dbKey := MigrationIntentKey{
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	//Check for the Intent already exists
	_, _, err := c.GetIntent(ctx, i, p, ca, v, di)
	if err != nil {
		log.Error("DeleteIntent ... Intent does not exist", log.Fields{"migrationIntent": i, "err": err})
		return err
	}

	log.Info("DeleteIntent ... Delete Migration Intent entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "migrationIntent": i})
	err = db.DBconn.Remove(ctx, c.db.StoreName, dbKey)
	if err != nil {
		log.Error("DeleteIntent ... DB Error .. Delete Migration Intent entry error", log.Fields{"err": err, "StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "migrationIntent": i})
		return pkgerrors.Wrapf(err, "DB Error .. Delete Migration Intent[%s] DB Error", i)
	}
	return nil
}
