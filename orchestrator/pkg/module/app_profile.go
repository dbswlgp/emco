// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module

import (
	"context"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// AppProfile contains the parameters needed for AppProfiles
// It implements the interface for managing the AppProfiles
type AppProfile struct {
	Metadata AppProfileMetadata `json:"metadata"`
	Spec     AppProfileSpec     `json:"spec"`
}

type AppProfileContent struct {
	Profile string `json:"profile"`
}

// AppProfileMetadata contains the metadata for AppProfiles
type AppProfileMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// AppProfileSpec contains the Spec for AppProfiles
type AppProfileSpec struct {
	AppName string `json:"app"`
}

// AppProfileKey is the key structure that is used in the database
type AppProfileKey struct {
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	CompositeProfile    string `json:"compositeProfile"`
	Profile             string `json:"appProfile"`
}

type AppProfileQueryKey struct {
	AppName string `json:"app"`
}

type AppProfileFindByAppKey struct {
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	CompositeProfile    string `json:"compositeProfile"`
	AppName             string `json:"app"`
}

// AppProfileManager exposes the AppProfile functionality
type AppProfileManager interface {
	CreateAppProfile(ctx context.Context, provider, compositeApp, compositeAppVersion, compositeProfile string, ap AppProfile, ac AppProfileContent, exists bool) (AppProfile, error)
	GetAppProfile(ctx context.Context, project, compositeApp, compositeAppVersion, compositeProfile, profile string) (AppProfile, error)
	GetAppProfiles(ctx context.Context, project, compositeApp, compositeAppVersion, compositeProfile string) ([]AppProfile, error)
	GetAppProfileByApp(ctx context.Context, project, compositeApp, compositeAppVersion, compositeProfile, appName string) (AppProfile, error)
	GetAppProfileContent(ctx context.Context, project, compositeApp, compositeAppVersion, compositeProfile, profile string) (AppProfileContent, error)
	GetAppProfileContentByApp(ctx context.Context, project, compositeApp, compositeAppVersion, compositeProfile, appName string) (AppProfileContent, error)
	DeleteAppProfile(ctx context.Context, project, compositeApp, compositeAppVersion, compositeProfile, profile string) error
}

// AppProfileClient implements the Manager
// It will also be used to maintain some localized state
type AppProfileClient struct {
	storeName  string
	tagMeta    string
	tagContent string
}

// NewAppProfileClient returns an instance of the AppProfileClient
// which implements the Manager
func NewAppProfileClient() *AppProfileClient {
	return &AppProfileClient{
		storeName:  "resources",
		tagMeta:    "data",
		tagContent: "profilecontent",
	}
}

// CreateAppProfile creates an entry for AppProfile in the database.
func (c *AppProfileClient) CreateAppProfile(ctx context.Context, project, compositeApp, compositeAppVersion, compositeProfile string, ap AppProfile, ac AppProfileContent, exists bool) (AppProfile, error) {
	key := AppProfileKey{
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
		CompositeProfile:    compositeProfile,
		Profile:             ap.Metadata.Name,
	}
	qkey := AppProfileQueryKey{
		AppName: ap.Spec.AppName,
	}

	res, err := c.GetAppProfile(ctx, project, compositeApp, compositeAppVersion, compositeProfile, ap.Metadata.Name)
	if res != (AppProfile{}) && !exists {
		return AppProfile{}, pkgerrors.New("AppProfile already exists")
	}

	res, err = c.GetAppProfileByApp(ctx, project, compositeApp, compositeAppVersion, compositeProfile, ap.Spec.AppName)
	if res != (AppProfile{}) && !exists {
		return AppProfile{}, pkgerrors.New("App already has an AppProfile")
	}

	//Check if composite profile exists (success assumes existance of all higher level 'parent' objects)
	_, err = NewCompositeProfileClient().GetCompositeProfile(ctx, compositeProfile, project, compositeApp, compositeAppVersion)
	if err != nil {
		return AppProfile{}, err
	}

	// TODO: (after app api is ready) check that the app Spec.AppName exists as part of the composite app

	err = db.DBconn.Insert(ctx, c.storeName, key, qkey, c.tagMeta, ap)
	if err != nil {
		return AppProfile{}, pkgerrors.Wrap(err, "Create DB entry error")
	}
	err = db.DBconn.Insert(ctx, c.storeName, key, qkey, c.tagContent, ac)
	if err != nil {
		return AppProfile{}, pkgerrors.Wrap(err, "Create DB entry error")
	}

	return ap, nil
}

// GetAppProfile - return specified App Profile
func (c *AppProfileClient) GetAppProfile(ctx context.Context, project, compositeApp, compositeAppVersion, compositeProfile, profile string) (AppProfile, error) {
	key := AppProfileKey{
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
		CompositeProfile:    compositeProfile,
		Profile:             profile,
	}

	value, err := db.DBconn.Find(ctx, c.storeName, key, c.tagMeta)
	if err != nil {
		return AppProfile{}, err
	} else if len(value) == 0 {
		return AppProfile{}, pkgerrors.New("AppProfile not found")
	}

	if value != nil {
		ap := AppProfile{}
		err = db.DBconn.Unmarshal(value[0], &ap)
		if err != nil {
			return AppProfile{}, err
		}
		return ap, nil
	}

	return AppProfile{}, pkgerrors.New("Unknown Error")

}

// GetAppProfile - return all App Profiles for given composite profile
func (c *AppProfileClient) GetAppProfiles(ctx context.Context, project, compositeApp, compositeAppVersion, compositeProfile string) ([]AppProfile, error) {

	key := AppProfileKey{
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
		CompositeProfile:    compositeProfile,
		Profile:             "",
	}

	var resp []AppProfile
	values, err := db.DBconn.Find(ctx, c.storeName, key, c.tagMeta)
	if err != nil {
		return []AppProfile{}, err
	}

	for _, value := range values {
		ap := AppProfile{}
		err = db.DBconn.Unmarshal(value, &ap)
		if err != nil {
			return []AppProfile{}, err
		}
		resp = append(resp, ap)
	}

	return resp, nil
}

// GetAppProfileByApp - return all App Profiles for given composite profile
func (c *AppProfileClient) GetAppProfileByApp(ctx context.Context, project, compositeApp, compositeAppVersion, compositeProfile, appName string) (AppProfile, error) {

	key := AppProfileFindByAppKey{
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
		CompositeProfile:    compositeProfile,
		AppName:             appName,
	}

	value, err := db.DBconn.Find(ctx, c.storeName, key, c.tagMeta)
	if err != nil {
		return AppProfile{}, err
	} else if len(value) == 0 {
		return AppProfile{}, pkgerrors.New("AppProfile not found")
	}

	if value != nil {
		ap := AppProfile{}
		err = db.DBconn.Unmarshal(value[0], &ap)
		if err != nil {
			return AppProfile{}, err
		}
		return ap, nil
	}

	return AppProfile{}, pkgerrors.New("Unknown Error")
}

func (c *AppProfileClient) GetAppProfileContent(ctx context.Context, project, compositeApp, compositeAppVersion, compositeProfile, profile string) (AppProfileContent, error) {
	key := AppProfileKey{
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
		CompositeProfile:    compositeProfile,
		Profile:             profile,
	}

	value, err := db.DBconn.Find(ctx, c.storeName, key, c.tagContent)
	if err != nil {
		return AppProfileContent{}, err
	} else if len(value) == 0 {
		return AppProfileContent{}, pkgerrors.New("AppProfileContent not found")
	}

	//value is a byte array
	if value != nil {
		ac := AppProfileContent{}
		err = db.DBconn.Unmarshal(value[0], &ac)
		if err != nil {
			return AppProfileContent{}, err
		}
		return ac, nil
	}

	return AppProfileContent{}, pkgerrors.New("Unknown Error")
}

func (c *AppProfileClient) GetAppProfileContentByApp(ctx context.Context, project, compositeApp, compositeAppVersion, compositeProfile, appName string) (AppProfileContent, error) {
	key := AppProfileFindByAppKey{
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
		CompositeProfile:    compositeProfile,
		AppName:             appName,
	}

	value, err := db.DBconn.Find(ctx, c.storeName, key, c.tagContent)
	if err != nil {
		return AppProfileContent{}, err
	} else if len(value) == 0 {
		return AppProfileContent{}, pkgerrors.New("AppProfileContent not found")
	}

	//value is a byte array
	if value != nil {
		ac := AppProfileContent{}
		err = db.DBconn.Unmarshal(value[0], &ac)
		if err != nil {
			return AppProfileContent{}, err
		}
		return ac, nil
	}

	return AppProfileContent{}, pkgerrors.New("Unknown Error")
}

// Delete AppProfile from the database
func (c *AppProfileClient) DeleteAppProfile(ctx context.Context, project, compositeApp, compositeAppVersion, compositeProfile, profile string) error {
	key := AppProfileKey{
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
		CompositeProfile:    compositeProfile,
		Profile:             profile,
	}

	err := db.DBconn.Remove(ctx, c.storeName, key)
	return err
}
