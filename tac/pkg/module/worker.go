// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"context"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/tac/pkg/model"
)

// WorkflowIntentClient implements the Manager
// It will also be used to maintain some localized state
type WorkerIntentClient struct {
	db db.DbInfo
}

func NewWorkerIntentClient() *WorkerIntentClient {
	return &WorkerIntentClient{
		db: db.DbInfo{
			StoreName: "resources", // should remain the same
			TagMeta:   "data",      // should remain the same
		},
	}
}

type WorkerIntentManager interface {
	// worker backend functions
	CreateOrUpdateWorkerIntent(wi model.WorkerIntent, tac, project, cApp, cAppVer, dig string, exists bool) (model.WorkerIntent, error)
	GetWorkerIntent(workerName, project, cApp, cAppVer, dig, tac string) (model.WorkerIntent, error)
	GetWorkerIntents(project, cApp, cAppVer, dig, tac string) ([]model.WorkerIntent, error)
	DeleteWorkerIntents(project, cApp, cAppVer, dig, tac, workerName string) error
}

func (v *WorkerIntentClient) CreateOrUpdateWorkerIntent(wi model.WorkerIntent, tac, project, cApp, cAppVer, dig string, exists bool) (model.WorkerIntent, error) {
	// print where we are
	log.Info("CreateOrUpdateWorkerIntent", log.Fields{"WorkerIntent": wi, "project": project,
		"cApp": cApp, "tac-intent": tac})

	// create the key for the worker.
	key := model.WorkerKey{
		WorkerName:          wi.Metadata.Name,
		WorkflowHook:        tac,
		DigName:             dig,
		CompositeApp:        cApp,
		Project:             project,
		CompositeAppVersion: cAppVer,
	}

	// check to see if this Worker already exists.
	_, err := v.GetWorkerIntent(wi.Metadata.Name, project, cApp, cAppVer, dig, tac)
	if err == nil && !exists {
		return model.WorkerIntent{}, errors.New("This worker already exists.")
	}

	// if it doesn't exist put it in db
	err = db.DBconn.Insert(context.Background(), v.db.StoreName, key, nil, v.db.TagMeta, wi)
	if err != nil {
		return model.WorkerIntent{}, err
	}

	return wi, nil
}

func (v WorkerIntentClient) GetWorkerIntent(workerName, project, cApp, cAppVer, dig, tac string) (model.WorkerIntent, error) {
	// print where we are
	log.Info("GetWorkerIntent", log.Fields{"WorkerName": workerName, "project": project,
		"cApp": cApp})

	// create the key for the worker.
	key := model.WorkerKey{
		WorkerName:          workerName,
		WorkflowHook:        tac,
		DigName:             dig,
		CompositeApp:        cApp,
		Project:             project,
		CompositeAppVersion: cAppVer,
	}

	// look for the worker in the db
	value, err := db.DBconn.Find(context.Background(), v.db.StoreName, key, v.db.TagMeta)
	if err != nil {
		// if there was an error, return it
		return model.WorkerIntent{}, err
	} else if len(value) == 0 {
		// if it dne then return nil
		return model.WorkerIntent{}, errors.New("Worker Not Found")
	}

	// value needs to be a byte array
	if value == nil {
		return model.WorkerIntent{}, errors.New("Unknown Error")
	}

	wi := model.WorkerIntent{}
	if err = db.DBconn.Unmarshal(value[0], &wi); err != nil {
		return model.WorkerIntent{}, err
	}

	return wi, nil
}

func (v WorkerIntentClient) GetWorkerIntents(project, cApp, cAppVer, dig, tac string) ([]model.WorkerIntent, error) {
	// print where we are
	log.Info("GetWorkerIntents", log.Fields{"tac-intent": tac, "project": project,
		"cApp": cApp})

	// create generic key to receive all workers
	key := model.WorkerKey{
		WorkerName:          "",
		WorkflowHook:        tac,
		DigName:             dig,
		CompositeApp:        cApp,
		Project:             project,
		CompositeAppVersion: cAppVer,
	}

	// query all workers on this tac-intent
	var resp []model.WorkerIntent
	values, err := db.DBconn.Find(context.Background(), v.db.StoreName, key, v.db.TagMeta)
	if err != nil {
		return []model.WorkerIntent{}, nil
	}

	// loop through raw data and put them into json objects
	for _, value := range values {
		// create empty worker intent model, and unmarshall bytes into model
		wi := model.WorkerIntent{}
		err = db.DBconn.Unmarshal(value, &wi)

		if err != nil {
			return []model.WorkerIntent{}, err
		}
		resp = append(resp, wi)
	}

	return resp, nil
}

func (v WorkerIntentClient) DeleteWorkerIntents(project, cApp, cAppVer, dig, tac, workerName string) error {
	// print where we are
	log.Info("DeleteWorkerIntent", log.Fields{"tac-intent": tac, "project": project,
		"cApp": cApp})

	// create the key used to delete the worker intent
	key := model.WorkerKey{
		WorkerName:          workerName,
		WorkflowHook:        tac,
		DigName:             dig,
		CompositeApp:        cApp,
		Project:             project,
		CompositeAppVersion: cAppVer,
	}

	// attempt to delete the entry, and return whatever error is given
	err := db.DBconn.Remove(context.Background(), v.db.StoreName, key)

	return err
}
