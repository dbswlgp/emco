// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (

	"fmt"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	"context"
	pkgerrors "github.com/pkg/errors"
	hpaModel "gitlab.com/project-emco/core/emco-base/src/hpa-plc/pkg/model"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

/*
AddIntent adds a given intent to the deployment-intent-group and stores in the db.
Other input parameters for it - projectName, compositeAppName, version, DeploymentIntentgroupName
*/
func (c *HpaPlacementClient) AddIntent(ctx context.Context, a hpaModel.DeploymentHpaIntent, p string, ca string, v string, di string, exists bool) (hpaModel.DeploymentHpaIntent, error) {
	//Check for the intent already exists here.
	res, dependentErrStaus, err := c.GetIntent(ctx, a.MetaData.Name, p, ca, v, di)
	if err != nil && dependentErrStaus == true {
		log.Error("AddIntent ... Intent dependency check failed", log.Fields{"hpaIntent": a.MetaData.Name, "err": err, "res-received": res})
		return hpaModel.DeploymentHpaIntent{}, err
	} else if (err == nil) && (!exists) {
		log.Error("AddIntent ... Intent already exists", log.Fields{"hpaIntent": a.MetaData.Name, "err": err, "res-received": res})
		return hpaModel.DeploymentHpaIntent{}, pkgerrors.New("Intent already exists")
	}

	dbKey := HpaIntentKey{
		IntentName:            a.MetaData.Name,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	log.Info("AddIntent ... Creating DB entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "hpaIntent": a.MetaData.Name})
	err = db.DBconn.Insert(ctx, c.db.StoreName, dbKey, nil, c.db.TagMetaData, a)
	if err != nil {
		log.Error("AddIntent ... DB Error .. Creating DB entry error", log.Fields{"StoreName": c.db.StoreName, "akey": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "hpaIntent": a.MetaData.Name})
		return hpaModel.DeploymentHpaIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}
	return a, nil
}

/*
GetIntent takes in an IntentName, ProjectName, CompositeAppName, Version and DeploymentIntentGroup.
It returns the Intent.
*/
func (c *HpaPlacementClient) GetIntent(ctx context.Context, i string, p string, ca string, v string, di string) (hpaModel.DeploymentHpaIntent, bool, error) {

	dbKey := HpaIntentKey{
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetIntent ... DB Error .. Get Intent error", log.Fields{"hpaIntent": i, "err": err})
		return hpaModel.DeploymentHpaIntent{}, false, err
	}

	if len(result) == 0 {
		return hpaModel.DeploymentHpaIntent{}, false, pkgerrors.New("Intent not found")
	}

	if result != nil {
		a := hpaModel.DeploymentHpaIntent{}
		err = db.DBconn.Unmarshal(result[0], &a)
		if err != nil {
			log.Error("GetIntent ... Unmarshalling  Intent error", log.Fields{"hpaIntent": i})
			return hpaModel.DeploymentHpaIntent{}, false, err
		}
		return a, false, nil
	}
	return hpaModel.DeploymentHpaIntent{}, false, pkgerrors.New("Unknown Error")
}

/*
GetAllIntents takes in projectName, CompositeAppName, CompositeAppVersion,
DeploymentIntentName . It returns ListOfIntents.
*/
func (c HpaPlacementClient) GetAllIntents(ctx context.Context, p string, ca string, v string, di string) ([]hpaModel.DeploymentHpaIntent, error) {

	dbKey := HpaIntentKey{
		IntentName:            "",
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetAllIntents ... DB Error .. Get HpaIntents db error", log.Fields{"StoreName": c.db.StoreName, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "len_result": len(result), "err": err})
		return []hpaModel.DeploymentHpaIntent{}, err
	}
	log.Info("GetAllIntents ... db result", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di})

	var listOfIntents []hpaModel.DeploymentHpaIntent
	for i := range result {
		a := hpaModel.DeploymentHpaIntent{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllIntents ... Unmarshalling HpaIntents error", log.Fields{"deploymentgroup": di})
				return []hpaModel.DeploymentHpaIntent{}, err
			}
			listOfIntents = append(listOfIntents, a)
		}
	}
	return listOfIntents, nil
}

/*
GetAllIntentsByApp takes in appName, projectName, CompositeAppName, CompositeAppVersion,
DeploymentIntentName . It returns ListOfIntents.
*/
func (c HpaPlacementClient) GetAllIntentsByApp(ctx context.Context, app string, p string, ca string, v string, di string) ([]hpaModel.DeploymentHpaIntent, error) {

	dbKey := HpaIntentKey{
		IntentName:            "",
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)

	if err != nil {
		log.Error("GetAllIntentsByApp .. DB Error", log.Fields{"StoreName": c.db.StoreName, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "len_result": len(result), "err": err})
		return []hpaModel.DeploymentHpaIntent{}, err
	}
	log.Info("GetAllIntentsByApp ... db result",
		log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "app": app})

	var listOfIntents []hpaModel.DeploymentHpaIntent
	for i := range result {
		a := hpaModel.DeploymentHpaIntent{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllIntentsByApp ... Unmarshalling HpaIntents error", log.Fields{"deploymentgroup": di})
				return []hpaModel.DeploymentHpaIntent{}, err
			}
			if a.Spec.AppName == app {
				listOfIntents = append(listOfIntents, a)
			}
		}
	}

	fmt.Println("\nhpa-plc/pkg/module/hpa-intent-api.go:160")
	fmt.Println("listOfIntents: ", listOfIntents,"\n")

	return listOfIntents, nil
}

/*
GetIntentByName takes in IntentName, projectName, CompositeAppName, CompositeAppVersion
and deploymentIntentGroupName returns the list of intents under the IntentName.
*/
func (c HpaPlacementClient) GetIntentByName(ctx context.Context, i string, p string, ca string, v string, di string) (hpaModel.DeploymentHpaIntent, error) {

	dbKey := HpaIntentKey{
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetIntentByName ... DB Error .. Get HpaIntent error", log.Fields{"hpaIntent": i})
		return hpaModel.DeploymentHpaIntent{}, err
	}

	if len(result) == 0 {
		return hpaModel.DeploymentHpaIntent{}, pkgerrors.New("Intent not found")
	}

	var a hpaModel.DeploymentHpaIntent
	err = db.DBconn.Unmarshal(result[0], &a)
	if err != nil {
		log.Error("GetIntentByName ...  Unmarshalling HpaIntent error", log.Fields{"hpaIntent": i})
		return hpaModel.DeploymentHpaIntent{}, err
	}
	return a, nil
}

// DeleteIntent deletes a given intent tied to project, composite app and deployment intent group
func (c HpaPlacementClient) DeleteIntent(ctx context.Context, i string, p string, ca string, v string, di string) error {
	dbKey := HpaIntentKey{
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	//Check for the Intent already exists
	_, _, err := c.GetIntent(ctx, i, p, ca, v, di)
	if err != nil {
		log.Error("DeleteIntent ... Intent does not exist", log.Fields{"hpaIntent": i, "err": err})
		return err
	}

	log.Info("DeleteIntent ... Delete Hpa Intent entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "hpaIntent": i})
	err = db.DBconn.Remove(ctx, c.db.StoreName, dbKey)
	if err != nil {
		log.Error("DeleteIntent ... DB Error .. Delete Hpa Intent entry error", log.Fields{"err": err, "StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "hpaIntent": i})
		return pkgerrors.Wrapf(err, "DB Error .. Delete Hpa Intent[%s] DB Error", i)
	}
	return nil
}
