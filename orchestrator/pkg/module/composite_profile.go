// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"encoding/json"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// CompositeProfile contains the parameters needed for CompositeProfiles
// It implements the interface for managing the CompositeProfiles
type CompositeProfile struct {
	Metadata CompositeProfileMetadata `json:"metadata"`
}

// CompositeProfileMetadata contains the metadata for CompositeProfiles
type CompositeProfileMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// CompositeProfileKey is the key structure that is used in the database
type CompositeProfileKey struct {
	Name         string `json:"compositeProfile"`
	Project      string `json:"project"`
	CompositeApp string `json:"compositeApp"`
	Version      string `json:"compositeAppVersion"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (cpk CompositeProfileKey) String() string {
	out, err := json.Marshal(cpk)
	if err != nil {
		return ""
	}

	return string(out)
}

// CompositeProfileManager exposes the CompositeProfile functionality
type CompositeProfileManager interface {
	CreateCompositeProfile(ctx context.Context, cpf CompositeProfile, p string, ca string,
		v string, exists bool) (CompositeProfile, error)
	GetCompositeProfile(ctx context.Context, compositeProfileName string, projectName string,
		compositeAppName string, version string) (CompositeProfile, error)
	GetCompositeProfiles(ctx context.Context, projectName string, compositeAppName string,
		version string) ([]CompositeProfile, error)
	DeleteCompositeProfile(ctx context.Context, compositeProfileName string, projectName string,
		compositeAppName string, version string) error
}

// CompositeProfileClient implements the Manager
// It will also be used to maintain some localized state
type CompositeProfileClient struct {
	storeName string
	tagMeta   string
}

// NewCompositeProfileClient returns an instance of the CompositeProfileClient
// which implements the Manager
func NewCompositeProfileClient() *CompositeProfileClient {
	return &CompositeProfileClient{
		storeName: "resources",
		tagMeta:   "data",
	}
}

// CreateCompositeProfile creates an entry for CompositeProfile in the database. Other Input parameters for it - projectName, compositeAppName, version
func (c *CompositeProfileClient) CreateCompositeProfile(ctx context.Context, cpf CompositeProfile, p string, ca string,
	v string, exists bool) (CompositeProfile, error) {

	res, err := c.GetCompositeProfile(ctx, cpf.Metadata.Name, p, ca, v)
	if res != (CompositeProfile{}) && !exists {
		return CompositeProfile{}, pkgerrors.New("CompositeProfile already exists")
	}

	cProfkey := CompositeProfileKey{
		Name:         cpf.Metadata.Name,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	err = db.DBconn.Insert(ctx, c.storeName, cProfkey, nil, c.tagMeta, cpf)
	if err != nil {
		return CompositeProfile{}, pkgerrors.Wrap(err, "Create DB entry error")
	}

	return cpf, nil
}

// GetCompositeProfile shall take arguments - name of the composite profile, name of the project, name of the composite app and version of the composite app. It shall return the CompositeProfile if its present.
func (c *CompositeProfileClient) GetCompositeProfile(ctx context.Context, cpf string, p string, ca string, v string) (CompositeProfile, error) {
	key := CompositeProfileKey{
		Name:         cpf,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	result, err := db.DBconn.Find(ctx, c.storeName, key, c.tagMeta)
	if err != nil {
		return CompositeProfile{}, err
	} else if len(result) == 0 {
		return CompositeProfile{}, pkgerrors.New("CompositeProfile not found")
	}

	if result != nil {
		cProf := CompositeProfile{}
		err = db.DBconn.Unmarshal(result[0], &cProf)
		if err != nil {
			return CompositeProfile{}, err
		}
		return cProf, nil
	}

	return CompositeProfile{}, pkgerrors.New("Unknown Error")
}

// GetCompositeProfiles shall take arguments - name of the project, name of the composite profile and version of the composite app. It shall return an array of CompositeProfile.
func (c *CompositeProfileClient) GetCompositeProfiles(ctx context.Context, p string, ca string, v string) ([]CompositeProfile, error) {
	key := CompositeProfileKey{
		Name:         "",
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	values, err := db.DBconn.Find(ctx, c.storeName, key, c.tagMeta)
	if err != nil {
		return []CompositeProfile{}, err
	}

	var resp []CompositeProfile

	for _, value := range values {
		cp := CompositeProfile{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []CompositeProfile{}, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteCompositeProfile deletes the compsiteApp profile from the database
func (c *CompositeProfileClient) DeleteCompositeProfile(ctx context.Context, cpf string, p string, ca string, v string) error {
	key := CompositeProfileKey{
		Name:         cpf,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	err := db.DBconn.Remove(ctx, c.storeName, key)
	return err
}
