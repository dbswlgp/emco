// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package gitsupport

import (
	"context"
	"fmt"
	"strings"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	pkgerrors "github.com/pkg/errors"

	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"
	emcogit2go "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogit2go"
	emcogithub "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogithub"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
)

type GitProvider struct {
	Cid       string
	Cluster   string
	App       string
	Namespace string
	Level     string
	GitType   string
	GitToken  string
	UserName  string
	Branch    string
	RepoName  string
	Url       string

	gitInterface GitInterfaceProvider
}

type GitInterfaceProvider interface {
	AddToCommit(fileName, content string, ref interface{}) interface{}
	DeleteToCommit(fileName string, ref interface{}) interface{}
	CommitFiles(app, message string, files interface{}) error
	ClusterWatcher(ctx context.Context, cid, app, cluster string, waitTime int) error
}

/*
	Function to create a New gitProvider
	params : cid, app, cluster, level, namespace string
	return : GitProvider, error
*/
func NewGitProvider(ctx context.Context, cid, app, cluster, level, namespace string) (*GitProvider, error) {

	result := strings.SplitN(cluster, "+", 2)

	c, err := utils.GetGitOpsConfig(ctx, cluster, "0", "default")
	if err != nil {
		return nil, err
	}
	// Read from database
	ccc := db.NewCloudConfigClient()
	refObject, err := ccc.GetClusterSyncObjects(ctx, result[0], c.Props.GitOpsReferenceObject)

	if err != nil {
		log.Error("Invalid refObject :", log.Fields{"refObj": c.Props.GitOpsReferenceObject, "error": err})
		return nil, err
	}

	kvRef := refObject.Spec.Kv

	var gitType, gitToken, branch, userName, repoName, url string

	for _, kvpair := range kvRef {
		log.Info("kvpair", log.Fields{"kvpair": kvpair})
		v, ok := kvpair["gitType"]
		if ok {
			gitType = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["gitToken"]
		if ok {
			gitToken = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["repoName"]
		if ok {
			repoName = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["userName"]
		if ok {
			userName = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["branch"]
		if ok {
			branch = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["url"]
		if ok {
			url = fmt.Sprintf("%v", v)
			continue
		}
	}
	if len(gitToken) <= 0 || len(branch) <= 0 || len(userName) <= 0 || len(repoName) <= 0 || len(url) <= 0 {
		log.Error("Missing information for Git", log.Fields{"token": gitToken, "branch": branch, "userName": userName, "repoName": repoName, "url": url})
		return nil, pkgerrors.Errorf("Missing Information for Git")
	}

	p := GitProvider{
		Cid:       cid,
		App:       app,
		Cluster:   cluster,
		Level:     level,
		Namespace: namespace,
		GitType:   gitType,
		GitToken:  gitToken,
		Branch:    branch,
		UserName:  userName,
		RepoName:  repoName,
		Url:       url,
	}

	if strings.EqualFold(gitType, "github") {
		p.gitInterface, err = emcogithub.NewGithub(p.Cluster, p.Url, p.Branch, p.UserName, p.RepoName, p.GitToken)
	} else {
		p.gitInterface, err = emcogit2go.NewGit2Go(p.Url, p.Branch, p.UserName, p.RepoName, p.GitToken)
	}

	return &p, nil
}

/*
	Function to get path of files stored in git
	params : string
	return : string
*/
func (p *GitProvider) GetPath(t string) string {
	return "clusters/" + p.Cluster + "/" + t + "/" + p.Cid + "/app/" + p.App + "/"
}

/*
	Function to create a new resource if the not already existing
	params : name string, ref interface{}, content []byte
	return : interface{}, error
*/
func (p *GitProvider) Create(name string, ref interface{}, content []byte) (interface{}, error) {

	path := p.GetPath("context") + name + ".yaml"
	files := p.gitInterface.AddToCommit(path, string(content), ref)
	return files, nil
}

/*
	Function to apply resource to the cluster
	params : name string, ref interface{}, content []byte
	return : interface{}, error
*/
func (p *GitProvider) Apply(path string, ref interface{}, content []byte) (interface{}, error) {

	files := p.gitInterface.AddToCommit(path, string(content), ref)
	return files, nil

}

/*
	Function to delete resource from the cluster
	params : name string, ref interface{}, content []byte
	return : interface{}, error
*/
func (p *GitProvider) Delete(path string, ref interface{}, content []byte) (interface{}, error) {

	files := p.gitInterface.DeleteToCommit(path, ref)
	return files, nil

}

/*
	Function to get resource from the cluster
	params : name string, gvkRes []byte
	return : []byte, error
*/
func (p *GitProvider) Get(name string, gvkRes []byte) ([]byte, error) {

	return []byte{}, nil
}

/*
	Function to commit resources to the cluster
	params : ctx context.Context, ref interface{}
	return : error
*/
func (p *GitProvider) Commit(ctx context.Context, ref interface{}) error {

	err := p.gitInterface.CommitFiles(p.App, "Commit for "+p.GetPath("context"), ref)
	return err
}

/*
	Function for cluster reachablity test
	params : null
	return : error
*/
func (p *GitProvider) IsReachable() error {
	return nil
}

// Wait time between reading git status (seconds)
var waitTime int = 60

// StartClusterWatcher watches for CR changes in git location
// go routine starts and reads after waitTime
// Thread exists when the AppContext is deleted
func (p *GitProvider) StartClusterWatcher(ctx context.Context) error {
	return p.gitInterface.ClusterWatcher(ctx, p.Cid, p.App, p.Cluster, waitTime)
}
