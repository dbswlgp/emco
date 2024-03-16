// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package action

import (
	"time"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdate"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	orchMod "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	con "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/utils"
	"gitlab.com/project-emco/core/emco-base/src/tac/pkg/model"
	"gitlab.com/project-emco/core/emco-base/src/tac/pkg/module"
	eta "gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/emcotemporalapi"

	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

var moduleClient *orchMod.Client

// UpdateAppContext applies the supplied intent against the given AppContext ID

func UpdateAppContext(ctx context.Context, intentName, appContextId, updateFromAppContext string) error {
	// Discern whether or not this is a pre-install loop or pre-update loop
	var hookString string
	if _, err := strconv.Atoi(updateFromAppContext); err != nil {
		hookString = "pre-install"
	} else {
		hookString = "pre-update"
	}

	// load orch client for deploy workers
	moduleClient = orchMod.NewClient()

	// load the app context
	var ac appcontext.AppContext

	_, err := ac.LoadAppContext(ctx, appContextId)
	if err != nil {
		logutils.Error("Failed to get the appContext.",
			logutils.Fields{
				"ID": appContextId})
		return errors.Wrapf(err, "Failed to get the appContext with ID: %s.", appContextId)
	}

	_, err = ac.GetCompositeAppHandle(ctx)
	if err != nil {
		return err
	}

	appContext, err := con.ReadAppContext(ctx, appContextId)
	if err != nil {
		logutils.Error("Failed to get the compositeApp for the appContext.",
			logutils.Fields{
				"ID": appContextId})
		return errors.Wrapf(err, "Failed to get the compositeApp for the appContext with ID: %s", appContextId)
	}

	// find the project, app, version, and group.
	project := appContext.CompMetadata.Project
	app := appContext.CompMetadata.CompositeApp
	version := appContext.CompMetadata.Version
	group := appContext.CompMetadata.DeploymentIntentGroup

	// look up workflow pre install hooks
	pre, err := module.NewClient().WorkflowIntentClient.GetSpecificHooks(ctx, project, app, version, group, hookString)
	if err != nil {
		logutils.Error("Failed to get the intents for the deploymentIntentGroup.",
			logutils.Fields{
				"DeploymentIntentGroup": group})
		return errors.Wrapf(err, "Failed to get the intents for the deploymentIntentGroup: %s", group)
	}

	// get all workers associated with all tac-intents
	toDeploy, err := getWorkers(pre, project, app, version, group)
	if err != nil {
		logutils.Error("Failed to get workers for tac-intents", logutils.Fields{"error": err})
		return errors.Wrapf(err, "Failed to get the workers for the tac intents: %s", group)
	}

	// verify that all workers are either running or they are approved to be run. If not then error out.
	toDeploy, err = verifyWorkers(toDeploy, project)
	if err != nil {
		logutils.Error("Failed to verify all workers were approved to be started or started.", logutils.Fields{"error": err})
		return errors.Wrapf(err, "Failed to verify all workers were approved or already started: %s", group)
	}
	logutils.Info("Obtained all workers, and verified they are all approved or already started.", logutils.Fields{"numer of workers to start": len(toDeploy)})

	// start all of the workers associated with the pre hooks
	err = deployWorkers(toDeploy, project)
	if err != nil {
		logutils.Error("Issue getting all of workers to start on time", logutils.Fields{"error": err})
		return errors.Wrapf(err, "Failed to start all workers before timeout: %s", group)
	}

	// iterate through all pre-install hooks
	logutils.Info("UpdateAppContext TAC ... Iterating through all workflow preinstall intents for this deployment intent group.",
		logutils.Fields{})

	for _, p := range pre {

		// log which pre-install hook we are working on
		logutils.Info(hookString+" hooks",
			logutils.Fields{
				"Temporal hook": p.Metadata.Name})

		// execute the workflow
		err = RunWorkflow(p.Spec.WfClientSpec.WfClientEndpointName,
			strconv.Itoa(p.Spec.WfClientSpec.WfClientEndpointPort), p.Spec.WfTemporalSpec.WfClientName, p.Spec.WfTemporalSpec)
		if err != nil {
			logutils.Error("Failed to run workflow in pre-install hooks.",
				logutils.Fields{
					"Hook name": p.Metadata.Name})
			return errors.Wrapf(err, "Failed to run workflow in pre-install hook: %s", p.Metadata.Name)
		}

	}

	// log that pre-install is finished
	logutils.Info("UpdateAppContext TAC ... Pre-install hooks finished.",
		logutils.Fields{
			"Project": project,
			"App":     app,
			"Version": version,
			"DIG":     group,
		})

	return nil
}

// TerminateAppContext is the to run the pre/post terminate workflows for this deployment intent group
func TerminateAppContext(ctx context.Context, appContextId string) error {
	var ac appcontext.AppContext

	logutils.Info("Terminate TAC ... pre-terminate hooks starting.", logutils.Fields{})

	// load orch client for deploy workers
	moduleClient = orchMod.NewClient()

	_, err := ac.LoadAppContext(ctx, appContextId)
	if err != nil {
		logutils.Error("Failed to get the appContext.",
			logutils.Fields{
				"ID": appContextId})
		return errors.Wrapf(err, "Failed to get the appContext with ID: %s.", appContextId)
	}

	_, err = ac.GetCompositeAppHandle(ctx)
	if err != nil {
		return err
	}

	appContext, err := con.ReadAppContext(ctx, appContextId)
	if err != nil {
		logutils.Error("Failed to get the compositeApp for the appContext.",
			logutils.Fields{
				"ID": appContextId})
		return errors.Wrapf(err, "Failed to get the compositeApp for the appContext with ID: %s", appContextId)
	}

	project := appContext.CompMetadata.Project
	app := appContext.CompMetadata.CompositeApp
	version := appContext.CompMetadata.Version
	group := appContext.CompMetadata.DeploymentIntentGroup

	// look up workflow pre terminate hooks
	pre, err := module.NewClient().WorkflowIntentClient.GetSpecificHooks(ctx, project, app, version, group, "pre-termination")
	if err != nil {
		logutils.Error("Failed to get the intents for the deploymentIntentGroup.",
			logutils.Fields{
				"DeploymentIntentGroup": group})
		return errors.Wrapf(err, "Failed to get the intents for the deploymentIntentGroup: %s", group)
	}

	// get all workers associated with all tac-intents
	toDeploy, err := getWorkers(pre, project, app, version, group)
	if err != nil {
		logutils.Error("Failed to get workers for tac-intents", logutils.Fields{"error": err})
		return errors.Wrapf(err, "Failed to get the workers for the tac intents: %s", group)
	}

	// verify that all workers are either running or they are approved to be run. If not then error out.
	toDeploy, err = verifyWorkers(toDeploy, project)
	if err != nil {
		logutils.Error("Failed to verify all workers were approved to be started or started.", logutils.Fields{"error": err})
		return errors.Wrapf(err, "Failed to verify all workers were approved or already started: %s", group)
	}
	logutils.Info("Obtained all workers, and verified they are all approved or already started.", logutils.Fields{"numer of workers to start": len(toDeploy)})

	// start all of the workers associated with the pre hooks
	err = deployWorkers(toDeploy, project)
	if err != nil {
		logutils.Error("Issue getting all of workers to start on time", logutils.Fields{"error": err})
		return errors.Wrapf(err, "Failed to start all workers before timeout: %s", group)
	}

	// iterate through all pre-terminate hooks
	logutils.Info("Terminate TAC ... Iterating through all workflow pre-terminate intents for this deployment intent group.",
		logutils.Fields{})

	for _, p := range pre {

		// log which pre-install hook we are working on
		logutils.Info("Terminate TAC ... Starting pre-terminate hooks.",
			logutils.Fields{
				"Temporal hook": p.Metadata.Name,
				"Project":       project,
				"App":           app,
				"Version":       version,
				"DIG":           group,
			})

		// execute the workflow
		err = RunWorkflow(p.Spec.WfClientSpec.WfClientEndpointName,
			strconv.Itoa(p.Spec.WfClientSpec.WfClientEndpointPort), p.Spec.WfTemporalSpec.WfClientName, p.Spec.WfTemporalSpec)
		if err != nil {
			logutils.Error("Failed to run workflow in pre-install hooks.",
				logutils.Fields{
					"Hook name": p.Metadata.Name})
			return errors.Wrapf(err, "Failed to run workflow in pre-install hook: %s", p.Metadata.Name)
		}

	}

	// log that pre-terminate is finished
	logutils.Info("Terminate TAC ... Pre-terminate hooks finished.",
		logutils.Fields{
			"Project": project,
			"App":     app,
			"Version": version,
			"DIG":     group,
		})

	return nil
}

func PostEvent(ctx context.Context, appContextId string, et contextupdate.EventType) error {
	eventString := []string{"post-install", "post-termination", "post-update"}
	workerIntent := []string{"pre-install", "pre-termination", "pre-update"}
	var ac appcontext.AppContext

	// load orch client for deploy workers
	moduleClient = orchMod.NewClient()

	logutils.Info("PostEvent TAC ... "+eventString[et]+" hooks starting.", logutils.Fields{})

	_, err := ac.LoadAppContext(ctx, appContextId)
	if err != nil {
		logutils.Error("PostEvent Failed to get the appContext.",
			logutils.Fields{
				"ID": appContextId})
		return errors.Wrapf(err, "PostEvent Failed to get the appContext with ID: %s.", appContextId)
	}

	_, err = ac.GetCompositeAppHandle(ctx)
	if err != nil {
		return err
	}

	appContext, err := con.ReadAppContext(ctx, appContextId)
	if err != nil {
		logutils.Error("PostEvent Failed to get the compositeApp for the appContext.",
			logutils.Fields{
				"ID": appContextId})
		return errors.Wrapf(err, "PostEvent Failed to get the compositeApp for the appContext with ID: %s", appContextId)
	}

	project := appContext.CompMetadata.Project
	app := appContext.CompMetadata.CompositeApp
	version := appContext.CompMetadata.Version
	group := appContext.CompMetadata.DeploymentIntentGroup

	// look up workflow pre install hooks
	post, err := module.NewClient().WorkflowIntentClient.GetSpecificHooks(ctx, project, app, version, group, eventString[et])
	if err != nil {
		logutils.Error("PostEvent Failed to get the intents for the deploymentIntentGroup.",
			logutils.Fields{
				"DeploymentIntentGroup": group})
		return errors.Wrapf(err, "PostEvent Failed to get the intents for the deploymentIntentGroup: %s", group)
	}

	// look up workflow pre install hooks
	pre, err := module.NewClient().WorkflowIntentClient.GetSpecificHooks(ctx, project, app, version, group, workerIntent[et])
	if err != nil {
		logutils.Error("PostEvent Failed to get the intents for the deploymentIntentGroup.",
			logutils.Fields{
				"DeploymentIntentGroup": group})
		return errors.Wrapf(err, "PostEvent Failed to get the intents for the deploymentIntentGroup: %s", group)
	}

	// get all workers associated with all tac-intents
	toTerminate, err := getWorkers(pre, project, app, version, group)
	if err != nil {
		logutils.Error("Failed to get workers for tac-intents", logutils.Fields{"error": err})
		return errors.Wrapf(err, "Failed to get the workers for the tac intents: %s", group)
	}

	// clean up the workers
	err = cleanUpWorkers(toTerminate, project)
	if err != nil {
		logutils.Error("Failed to terminate the workers", logutils.Fields{"error": err})
		return errors.Wrapf(err, "Failed to delete the workers after use.")
	}

	// iterate through all pre-install hooks
	logutils.Info("PostEvent TAC ... Iterating through all workflow "+eventString[et]+" intents for this deployment intent group.",
		logutils.Fields{})

	for _, p := range post {

		// log which pre-install hook we are working on
		logutils.Info("PostEvent TAC ... Starting "+eventString[et]+" hooks.",
			logutils.Fields{
				"Temporal hook": p.Metadata.Name,
				"Project":       project,
				"App":           app,
				"Version":       version,
				"DIG":           group,
			})

		// execute the workflow
		err = RunWorkflow(p.Spec.WfClientSpec.WfClientEndpointName,
			strconv.Itoa(p.Spec.WfClientSpec.WfClientEndpointPort), p.Spec.WfTemporalSpec.WfClientName, p.Spec.WfTemporalSpec)
		if err != nil {
			logutils.Error("Failed to run workflow in "+eventString[et]+" hooks.",
				logutils.Fields{
					"Hook name": p.Metadata.Name})
			return errors.Wrapf(err, "Failed to run workflow in PostEvent hook: %s", p.Metadata.Name)
		}

	}

	// log that PostEvent is finished
	logutils.Info("PostEvent TAC ... "+eventString[et]+" hooks finished.",
		logutils.Fields{
			"Project": project,
			"App":     app,
			"Version": version,
			"DIG":     group,
		})

	return nil
}

func RunWorkflow(clientEndpoint, clientPort, clientName string, spec eta.WfTemporalSpec) error {
	url := "http://" + clientEndpoint + ":" +
		clientPort + "/invoke/" +
		clientName

	jsonBytes, err := json.Marshal(spec)
	if err != nil {
		logutils.Error("Temporal Action Controller.... Error marshalling the workflow temporal spec into post body",
			logutils.Fields{"error": err.Error()})
		return err
	}
	logutils.Info("Temporal Action Controller.... Run workflow",
		logutils.Fields{"url": url, "workflow temporal spec": string(jsonBytes)})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		logutils.Error("Temporal Action Controller.... Run workflow could not post",
			logutils.Fields{"error": err.Error()})
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		logutils.Error("Temporal Action Controller.... Run workflow POST returned error",
			logutils.Fields{"status code": resp.Status, "urL": url})
		return errors.New("Error starting workflow")
	}

	return nil
}

// helper function to get all workers registered to the list of submitted TAC intents
func getWorkers(tacIntents []model.WorkflowHookIntent, project, app, version, group string) ([]model.WorkerIntent, error) {
	logutils.Info("Getting all workers for relevant tac intents in DIG", logutils.Fields{"DIG": group})
	var toDeploy []model.WorkerIntent

	// loop through all pre hooks, and all the registered workers to a list
	for _, t := range tacIntents {
		// get all workers associated with this tac-intent
		ws, err := module.NewClient().WorkerIntentClient.GetWorkerIntents(project, app, version, group, t.Metadata.Name)
		logutils.Info("Workers for tac intent", logutils.Fields{"tac-intent": t.Metadata.Name, "number of workers": len(ws), "workers": ws})

		// if no error, and this tac-intent has workers then add them to list of workers to deploy
		if err != nil {
			return []model.WorkerIntent{}, err
		} else if ws != nil {
			toDeploy = append(toDeploy, ws...)
		}

	}

	logutils.Info("Obtained all relevant workers for all tac intents", logutils.Fields{"number of workers": len(toDeploy), "workers": toDeploy})
	return toDeploy, nil
}

// helper function to verify that all workers are either already running or approved to be run.
func verifyWorkers(workers []model.WorkerIntent, project string) ([]model.WorkerIntent, error) {
	logutils.Info("Verifying that all workers are either already running or approved to be initialized",
		logutils.Fields{})

	// to deploy is the list of workers that are only approved. We will manage these. We will assume already deployed apps are needed elsewhere
	// instantiationClient is used to get info about DIGs
	var toDeploy []model.WorkerIntent
	instantiationClient := moduleClient.Instantiation

	// loop through all workers and make sure they are valid for run.
	for _, w := range workers {
		res, err := instantiationClient.Status(context.Background(), project, w.Spec.CApp, w.Spec.CAppVersion, w.Spec.DIG, "", "", "", []string{}, []string{}, []string{})
		if err != nil {
			return []model.WorkerIntent{}, err
		}

		logutils.Info("== Status Apps List", logutils.Fields{"response": res, "ready-status": res.State.Actions[len(res.State.Actions)-1].State})

		// approved - we will manage, and is ok
		// instantiated - is already deployed. we will not manage. is ok.
		// any other state is not okay
		if res.State.Actions[len(res.State.Actions)-1].State == "Approved" || res.State.Actions[len(res.State.Actions)-1].State == "Terminated" {
			toDeploy = append(toDeploy, w)
		} else if res.State.Actions[len(res.State.Actions)-1].State == "Instantiated" {
			logutils.Info("App is already instantiated. Will not manage automatically", logutils.Fields{"Worker": w})
		} else {
			return []model.WorkerIntent{}, errors.New("Submitted worker not approved to be instantiated, or already instantiated.")
		}
	}

	return toDeploy, nil
}

// helper function to start all workers, and then wait for them to become ready. Will return error if workers can't become ready before timeout.
func deployWorkers(workers []model.WorkerIntent, project string) error {
	logutils.Info("Initializing all relevant workers, and making sure they all enter ready state before timeout.",
		logutils.Fields{})

	instantiationClient := moduleClient.Instantiation

	// start all of the workers
	for _, w := range workers {
		err := instantiationClient.Instantiate(context.Background(), project, w.Spec.CApp, w.Spec.CAppVersion, w.Spec.DIG)
		if err != nil {
			return errors.New("Issue starting one of the workers")
		}
	}

	// Loop to wait for all okay status, or fail if all items are not ready.
	readyStatus := make([]bool, len(workers))
	start := time.Now().UnixMilli()
	for i, w := range workers {
		res, err := instantiationClient.Status(context.Background(), project, w.Spec.CApp, w.Spec.CAppVersion, w.Spec.DIG, "", "ready", "", []string{}, []string{}, []string{})
		if err != nil {
			return err
		}

		if res.ReadyStatus == "Ready" {
			// the application is ready, all good.
			readyStatus[i] = true
		} else if (int(time.Now().UnixMilli() - start)) > w.Spec.StartToCloseTimeout {
			// the application is not ready, and the timeout has expired. Return error
			return errors.New("Timeout reached, and DIG isn't ready.")
		}

		if allReady(readyStatus) {
			// if all workers ready break the loop
			break
		}
	}

	return nil
}

func allReady(status []bool) bool {

	// check to see if all statuses are ready
	for _, s := range status {
		if !s {
			return false
		}
	}

	// if you reach this then all statuses are ready
	return true
}

// Terminate all of the workers after tac has completed
func cleanUpWorkers(workers []model.WorkerIntent, project string) error {
	logutils.Info("Terminating all registered workers now that jobs are done.",
		logutils.Fields{})

	instantiationClient := moduleClient.Instantiation

	// start all of the workers
	for _, w := range workers {
		err := instantiationClient.Terminate(context.Background(), project, w.Spec.CApp, w.Spec.CAppVersion, w.Spec.DIG)
		if err != nil {
			return errors.New("Issue terminating one of the workers")
		}
	}

	return nil
}
