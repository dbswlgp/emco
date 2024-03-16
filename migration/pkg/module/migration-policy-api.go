// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
//	"fmt"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	"context"
	pkgerrors "github.com/pkg/errors"
	migrationModel "gitlab.com/project-emco/core/emco-base/src/migration/pkg/model"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

/*
AddIntent adds a given intent to the deployment-intent-group and stores in the db.
Other input parameters for it - projectName, compositeAppName, version, DeploymentIntentgroupName
*/
func (c *MigrationClient) AddPolicy(ctx context.Context, a migrationModel.MigrationPolicy, p string, exists bool) (migrationModel.MigrationPolicy, error) {
	//Check for the intent already exists here.
	res, dependentErrStaus, err := c.GetPolicy(ctx, a.MetaData.Name, p)
	if err != nil && dependentErrStaus == true {
		log.Error("AddPolicy ... Policy dependency check failed", log.Fields{"migrationPolicy": a.MetaData.Name, "err": err, "res-received": res})
		return migrationModel.MigrationPolicy{}, err
	} else if (err == nil) && (!exists) {
		log.Error("AddPolicy ... Policy already exists", log.Fields{"migrationPolicy": a.MetaData.Name, "err": err, "res-received": res})
		return migrationModel.MigrationPolicy{}, pkgerrors.New("Policy already exists")
	}

	dbKey := MigrationPolicyKey{
		PolicyName:            a.MetaData.Name,
		Project:               p,
	}

	log.Info("AddPolicy ... Creating DB entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "migrationPolicy": a.MetaData.Name})
	err = db.DBconn.Insert(ctx, c.db.StoreName, dbKey, nil, c.db.TagMetaData, a)
	if err != nil {
		log.Error("AddPolicy ... DB Error .. Creating DB entry error", log.Fields{"StoreName": c.db.StoreName, "akey": dbKey, "project": p, "migrationPolicy": a.MetaData.Name})
		return migrationModel.MigrationPolicy{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}
	return a, nil
}


/*
UpdateIntent adds a given intent to the deployment-intent-group and stores in the db.
Other input parameters for it - projectName, compositeAppName, version, DeploymentIntentgroupName
*/
/*
func (c *MigrationClient) UpdatePolicy(ctx context.Context, migrationPolicyName string, p string, ca string, v string, di string, fieldupdate string) (migrationModel.DeploymentMigrationPolicy, error) {

	dbKey := MigrationPolicyKey{
		IntentName:            migrationPolicyName,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	log.Info("AddIntent ... Creating DB entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "migrationPolicy": migrationPolicyName})
	//err = db.DBconn.Insert(ctx, c.db.StoreName, dbKey, nil, c.db.TagMetaData, a)

	err := db.DBconn.UpdateIntentField(ctx, c.db.StoreName, dbKey, "data.metadata.description", fieldupdate)

	if err != nil {
		log.Error("UpdateIntent ... DB Error .. Updating DB entry error", log.Fields{"StoreName": c.db.StoreName, "akey": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "migrationPolicy": migrationPolicyName})
		return pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return nil
}
*/

/*
GetIntent takes in an IntentName, ProjectName, CompositeAppName, Version and DeploymentIntentGroup.
It returns the Intent.
*/
func (c *MigrationClient) GetPolicy(ctx context.Context, pn string, p string) (migrationModel.MigrationPolicy, bool, error) {

	dbKey := MigrationPolicyKey{
		PolicyName:            pn,
		Project:               p,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetPolicy ... DB Error .. Get Policy error", log.Fields{"migrationPolicy": pn, "err": err})
		return migrationModel.MigrationPolicy{}, false, err
	}

	if len(result) == 0 {
		return migrationModel.MigrationPolicy{}, false, pkgerrors.New("Policy not found")
	}

	if result != nil {
		a := migrationModel.MigrationPolicy{}
		err = db.DBconn.Unmarshal(result[0], &a)
		if err != nil {
			log.Error("GetPolicy ... Unmarshalling  Policy error", log.Fields{"migrationPolicy": pn})
			return migrationModel.MigrationPolicy{}, false, err
		}
		return a, false, nil
	}
	return migrationModel.MigrationPolicy{}, false, pkgerrors.New("Unknown Error")
}

/*
GetAllIntents takes in projectName, CompositeAppName, CompositeAppVersion,
DeploymentIntentName . It returns ListOfIntents.
*/
func (c MigrationClient) GetAllPolicies(ctx context.Context, p string) ([]migrationModel.MigrationPolicy, error) {

	dbKey := MigrationPolicyKey{
		PolicyName:            "",
		Project:               p,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetAllPolicies ... DB Error .. Get MigrationPolicies db error", log.Fields{"StoreName": c.db.StoreName, "project": p, "len_result": len(result), "err": err})
		return []migrationModel.MigrationPolicy{}, err
	}
	log.Info("GetAllPolicies ... db result", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p})

	var listOfPolicies []migrationModel.MigrationPolicy
	for i := range result {
		a := migrationModel.MigrationPolicy{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllPolicies ... Unmarshalling MigrationPolicies error", log.Fields{"project": p})
				return []migrationModel.MigrationPolicy{}, err
			}
			listOfPolicies = append(listOfPolicies, a)
		}
	}
	return listOfPolicies, nil
}

/*
GetAllExistIntents takes in projectName, CompositeAppName, CompositeAppVersion,
DeploymentIntentName . It returns ListOfIntents.
*/
/*
func (c MigrationClient) GetAllIntentsByDIG(ctx context.Context, p string, di string) ([]migrationModel.DeploymentMigrationPolicy, error) {

	dbKey := MigrationPolicyKey{
		IntentName:            "",
		Project:               p,
		CompositeApp:          "",
		Version:               "",
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetAllIntents ... DB Error .. Get MigrationPolicys db error", log.Fields{"StoreName": c.db.StoreName, "project": p, "deploymentIntentGroup": di, "len_result": len(result), "err": err})
		return []migrationModel.DeploymentMigrationPolicy{}, err
	}
	log.Info("GetAllIntents ... db result", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "deploymentIntentGroup": di })

	var listOfPolicies []migrationModel.DeploymentMigrationPolicy
	for i := range result {
		a := migrationModel.DeploymentMigrationPolicy{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllIntents ... Unmarshalling MigrationPolicys error", log.Fields{"deploymentgroup": di})
				return []migrationModel.DeploymentMigrationPolicy{}, err
			}
			listOfPolicies = append(listOfPolicies, a)
		}
	}
	return listOfPolicies, nil
}
*/

/*
GetAllIntentsByApp takes in appName, projectName, CompositeAppName, CompositeAppVersion,
DeploymentIntentName . It returns ListOfIntents.
*/
/*
func (c MigrationClient) GetAllIntentsByApp(ctx context.Context, app string, p string, ca string, v string, di string) ([]migrationModel.DeploymentMigrationPolicy, error) {

	dbKey := MigrationPolicyKey{
		IntentName:            "",
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)

	if err != nil {
		log.Error("GetAllIntentsByApp .. DB Error", log.Fields{"StoreName": c.db.StoreName, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "len_result": len(result), "err": err})
		return []migrationModel.DeploymentMigrationPolicy{}, err
	}
	log.Info("GetAllIntentsByApp ... db result",
		log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "compositeApp": ca, "compositeAppVersion": v, "deploymentIntentGroup": di, "app": app})

	var listOfPolicies []migrationModel.DeploymentMigrationPolicy
	for i := range result {
		a := migrationModel.DeploymentMigrationPolicy{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllIntentsByApp ... Unmarshalling MigrationPolicys error", log.Fields{"deploymentgroup": di})
				return []migrationModel.DeploymentMigrationPolicy{}, err
			}
			//if a.Spec.AppName == app {
			listOfPolicies = append(listOfPolicies, a)
			//}

		}
	}

	fmt.Println("\nmigration/pkg/module/hpa-intent-api.go:156")
	fmt.Println("listOfPolicies: ", listOfPolicies,"\n")

	return listOfPolicies, nil
}
*/


/*
GetIntentByName takes in IntentName, projectName, CompositeAppName, CompositeAppVersion
and deploymentIntentGroupName returns the list of intents under the IntentName.
*/
func (c MigrationClient) GetPolicyByName(ctx context.Context, pn string, p string) (migrationModel.MigrationPolicy, error) {

	dbKey := MigrationPolicyKey{
		PolicyName:            pn,
		Project:               p,
	}

	result, err := db.DBconn.Find(ctx, c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetPolicyByName ... DB Error .. Get MigrationPolicy error", log.Fields{"migrationPolicy": pn})
		return migrationModel.MigrationPolicy{}, err
	}

	if len(result) == 0 {
		return migrationModel.MigrationPolicy{}, pkgerrors.New("Policy not found")
	}

	var a migrationModel.MigrationPolicy
	err = db.DBconn.Unmarshal(result[0], &a)
	if err != nil {
		log.Error("GetPolicyByName ...  Unmarshalling MigrationPolicy error", log.Fields{"migrationPolicy": pn})
		return migrationModel.MigrationPolicy{}, err
	}
	return a, nil
}



// DeleteIntent deletes a given intent tied to project, composite app and deployment intent group
func (c MigrationClient) DeletePolicy(ctx context.Context, pn string, p string) error {
	dbKey := MigrationPolicyKey{
		PolicyName:            pn,
		Project:               p,
	}

	//Check for the Policy already exists
	_, _, err := c.GetPolicy(ctx, pn, p)
	if err != nil {
		log.Error("DeletePolicy ... Policy does not exist", log.Fields{"migrationPolicy": pn, "err": err})
		return err
	}

	log.Info("DeletePolicy ... Delete Migration Policy entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "migrationPolicy": pn})
	err = db.DBconn.Remove(ctx, c.db.StoreName, dbKey)
	if err != nil {
		log.Error("DeletePolicy ... DB Error .. Delete Migration Policy entry error", log.Fields{"err": err, "StoreName": c.db.StoreName, "key": dbKey, "project": p, "migrationPolicy": pn})
		return pkgerrors.Wrapf(err, "DB Error .. Delete Migration Policy[%s] DB Error", pn)
	}
	return nil
}
