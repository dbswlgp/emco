// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package state

import (
	"context"
	"encoding/json"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	pkgerrors "github.com/pkg/errors"
)

// GetAppContextFromStateInfo loads the appcontext present in the StateInfo input
func GetAppContextFromId(ctx context.Context, ctxid string) (appcontext.AppContext, error) {
	var cc appcontext.AppContext
	_, err := cc.LoadAppContext(ctx, ctxid)
	if err != nil {
		return appcontext.AppContext{}, err
	}
	return cc, nil
}

// GetCurrentStateFromStatInfo gets the last (current) state from StateInfo
func GetCurrentStateFromStateInfo(s StateInfo) (StateValue, error) {
	alen := len(s.Actions)
	if alen == 0 {
		return StateEnum.Undefined, pkgerrors.Errorf("No state information")
	}
	return s.Actions[alen-1].State, nil
}

// GetStatusContextIdForContextId  gets the status context id associated with the
// input context id.  This will be the context id of the most recent "Instantiated"
// state.  If the provided 'ctxid' is not found, then that is an error
func GetStatusContextIdForContextId(s StateInfo, ctxid string) (string, error) {
	found := false
	var pos int
	for i, entry := range s.Actions {
		if ctxid == entry.ContextId {
			found = true
			pos = i
			break
		}
	}
	if !found {
		return "", pkgerrors.Errorf("No state information for %v", ctxid)
	}

	for i := pos; i >= 0; i-- {
		if s.Actions[i].State == StateEnum.Instantiated ||
			s.Actions[i].State == StateEnum.Applied {
			return s.Actions[i].ContextId, nil
		}
	}
	return "", pkgerrors.Errorf("Status context ID not found for %v", ctxid)
}

// GetContextIdForStatusContextId  given a statusContextId (not checked),
// get the most recent ContextId - i.e. either end of the list or up to "Terminated"
// Assumed that 'ctxid' has already been identified as a status contextId.
func GetContextIdForStatusContextId(s StateInfo, ctxid string) (string, error) {
	found := false
	var pos int
	for i, entry := range s.Actions {
		if ctxid == entry.ContextId {
			found = true
			pos = i
			break
		}
	}
	if !found {
		return "", pkgerrors.Errorf("No state information for %v", ctxid)
	}

	for i := pos + 1; i < len(s.Actions); i++ {
		pos = i
		if s.Actions[i].State == StateEnum.Terminated {
			break
		}
	}
	return s.Actions[pos].ContextId, nil
}

// GetLastContextFromStatInfo gets the last (most recent) context id from StateInfo
func GetLastContextIdFromStateInfo(s StateInfo) string {
	alen := len(s.Actions)
	if alen > 0 {
		return s.Actions[alen-1].ContextId
	} else {
		return ""
	}
}

// GetStatusContextIdFromStateInfo gets status AppContext
func GetStatusContextIdFromStateInfo(s StateInfo) string {
	return s.StatusContextId
}

// GetLatestRevisionFromStateInfo returns the latest revision from StateInfo
func GetLatestRevisionFromStateInfo(s StateInfo) (int64, error) {
	alen := len(s.Actions)
	if alen == 0 {
		return -1, pkgerrors.Errorf("No state information")
	}
	return s.Actions[alen-1].Revision, nil
}

// GetMatchingContextIDforRevision returns the matching contextID for a given revision and stateInfo
func GetMatchingContextIDforRevision(s StateInfo, r int64) (string, error) {
	alen := len(s.Actions)
	if alen == 0 {
		return "", pkgerrors.Errorf("No state information")
	}
	for _, eachActionEntry := range s.Actions {
		if eachActionEntry.Revision == r {
			logutils.Info("Found the matching revisionID", logutils.Fields{"Revision": eachActionEntry.Revision, "ContextID": eachActionEntry.ContextId})
			return eachActionEntry.ContextId, nil
		}

	}
	logutils.Info("No the matching revisionID found", logutils.Fields{"Revision": r})
	return "", pkgerrors.Errorf("No matching ContextId found")
}

// GetContextIdsFromStatInfo return a list of the unique AppContext Ids in the StateInfo
func GetContextIdsFromStateInfo(s StateInfo) []string {
	m := make(map[string]string)

	for _, a := range s.Actions {
		if a.ContextId != "" {
			m[a.ContextId] = ""
		}
	}

	ids := make([]string, len(m))
	i := 0
	for k := range m {
		ids[i] = k
		i++
	}

	return ids
}

func GetAppContextStatus(ctx context.Context, ctxid string) (appcontext.AppContextStatus, error) {
	ac, err := GetAppContextFromId(ctx, ctxid)
	if err != nil {
		return appcontext.AppContextStatus{}, err
	}

	h, err := ac.GetCompositeAppHandle(ctx)
	if err != nil {
		return appcontext.AppContextStatus{}, err
	}
	sh, err := ac.GetLevelHandle(ctx, h, "status")
	if err != nil {
		return appcontext.AppContextStatus{}, err
	}
	s, err := ac.GetValue(ctx, sh)
	if err != nil {
		return appcontext.AppContextStatus{}, err
	}
	acStatus := appcontext.AppContextStatus{}
	js, _ := json.Marshal(s)
	json.Unmarshal(js, &acStatus)

	return acStatus, nil

}

func UpdateAppContextStopFlag(ctx context.Context, ctxid string, sf bool) error {
	ac, err := GetAppContextFromId(ctx, ctxid)
	if err != nil {
		return err
	}
	hc, err := ac.GetCompositeAppHandle(ctx)
	if err != nil {
		return err
	}
	sh, err := ac.GetLevelHandle(ctx, hc, "stopflag")
	if sh == nil {
		_, err = ac.AddLevelValue(ctx, hc, "stopflag", sf)
	} else {
		err = ac.UpdateValue(ctx, sh, sf)
	}
	if err != nil {
		return err
	}
	return nil
}

// UpdateAppContextStatusContextID updates status context id in the AppContext
func UpdateAppContextStatusContextID(ctx context.Context, ctxid string, sctxid string) error {
	ac, err := GetAppContextFromId(ctx, ctxid)
	if err != nil {
		return err
	}
	hc, err := ac.GetCompositeAppHandle(ctx)
	if err != nil {
		return err
	}
	sh, err := ac.GetLevelHandle(ctx, hc, "statusappctxid")
	if sh == nil {
		_, err = ac.AddLevelValue(ctx, hc, "statusappctxid", sctxid)
	} else {
		err = ac.UpdateValue(ctx, sh, sctxid)
	}
	if err != nil {
		return err
	}
	return nil
}
