// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package gitsupport

import (
	"context"
	"os"
	"strings"

	pkgerrors "github.com/pkg/errors"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	//emcogit "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogit"
	emcogit2go "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogit2go"
	emcogithub "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogithub"
)

type GitProvider struct {
	GitUser      string
	GitRepo      string
	GitToken     string
	Cluster      string
	GitType      string
	GitBranch    string
	Url          string
	gitInterface GitInterfaceProvider
}

type GitInterfaceProvider interface {
	AddToCommit(fileName, content string, ref interface{}) interface{}
	DeleteToCommit(fileName string, ref interface{}) interface{}
	CommitFiles(app, message string, files interface{}) error
	ClusterWatcher(ctx context.Context, cid, app, cluster string, waitTime int) error
	CommitStatus(commitMessage, branchName, cid, app string, files interface{}) error
}

/*
	Function to create a New gitProvider
	params : cid, app, cluster, level, namespace string
	return : GitProvider, error
*/
func NewGitProvider() (*GitProvider, error) {

	gitBranch := os.Getenv("GIT_BRANCH")
	gitType := os.Getenv("GIT_TYPE")
	gitUser := os.Getenv("GIT_USERNAME")
	gitToken := os.Getenv("GIT_TOKEN")
	gitRepo := os.Getenv("GIT_REPO")
	clusterName := os.Getenv("GIT_CLUSTERNAME")
	gitUrl := os.Getenv("GIT_URL")

	// If any value is not provided then can't store in Git location
	if len(gitRepo) <= 0 || len(gitToken) <= 0 || len(gitUser) <= 0 || len(clusterName) <= 0 || len(gitUrl) <= 0 {
		log.Info("Git information not found:: Skipping Git storage", log.Fields{})
		return nil, pkgerrors.Errorf("Missing Information for Git")
	}

	log.Info("Git Info found", log.Fields{"gitRepo::": gitRepo, "cluster::": clusterName})

	p := GitProvider{
		GitUser:   gitUser,
		GitRepo:   gitRepo,
		GitToken:  gitToken,
		GitType:   gitType,
		GitBranch: gitBranch,
		Cluster:   clusterName,
		Url:       gitUrl,
	}

	var err error
	if strings.EqualFold(gitType, "github") {
		p.gitInterface, err = emcogithub.NewGithub(p.Cluster, p.Url, p.GitBranch, p.GitUser, p.GitRepo, p.GitToken)
		if err != nil {
			log.Error("Error in creating a github client", log.Fields{"err": err})
			return nil, err
		}
	} else {
		p.gitInterface, err = emcogit2go.NewGit2Go(p.Url, p.GitBranch, p.GitUser, p.GitRepo, p.GitToken)
		if err != nil {
			log.Error("Error in creating a emcogit2go client", log.Fields{"err": err})
			return nil, err
		}
	}

	return &p, nil
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
func (p *GitProvider) CommitStatus(commitMessage, branchName, cid, app string, files interface{}) error {

	err := p.gitInterface.CommitStatus(commitMessage, branchName, cid, app, files)
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
