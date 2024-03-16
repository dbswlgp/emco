// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package anthos

import (
	"context"
	"strings"
	"time"

	"encoding/binary"
	"encoding/json"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
)

type Resource struct {
	ApiVersion    string    `json:"apiVersion"`
	Kind          string    `json:"kind"`
	MetaData      MetaDatas `json:"metadata"`
	Specification Specs     `json:"spec,omitempty"`
}

type MetaDatas struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

type Specs struct {
	Git AnthosGit `json:"git,omitempty"` // for Anthos RepoSync
}

type AnthosGit struct {
	Repo        string `json:"repo"`
	Revision    string `json:"revision"`
	Branch      string `json:"branch"`
	Dir         string `json:"dir"`
	Auth        string `json:"auth"`
	NoSSLVerify bool   `json:"noSSLVerify"`
}

func (p *AnthosProvider) ApplyConfig(ctx context.Context, config interface{}) error {

	// Prepare git changes
	var gp interface{}

	acUtils, err := utils.NewAppContextReference(ctx, p.gitProvider.Cid)
	if err != nil {
		return nil
	}
	appns, applevel := acUtils.GetNamespace(ctx)
	_, name, _, lcns, lclevel, err := acUtils.GetLogicalCloudInfo(ctx)
	log.Debug("Composite App's namespace:", log.Fields{"appns": appns})
	log.Debug("Composite App's level:", log.Fields{"applevel": applevel})
	log.Debug("Logical Cloud's namespace:", log.Fields{"namespace": lcns})
	log.Debug("Logical Cloud's level:", log.Fields{"lclevel": lclevel})
	log.Debug("Git Provider's level:", log.Fields{"p.gitProvider.Level": p.gitProvider.Level})

	if err != nil {
		return err
	}
	if lclevel == "1" {
		// For L1 LC: Create RepoSync for namespace: Google Anthos GitOps backend supports Privileged Logical Clouds using RepoSync.
		// The Google Anthos RepoSync allows resources to be synchronized over GitOps into a GKE namespace.
		// RepoSync automatically creates a Kubernetes Service Account void of any access privileges, including to the defined namespace.
		// Later in the Apply operation, DCM will convert User Permissions into role bindings, and RepoSync's Kubernetes
		// Service Account will obtain the desired level of access to achieve full Privileged Logical Cloud support.

		rsName := strings.Join([]string{"repo-sync-", name}, "")
		rs := Resource{
			ApiVersion: "configsync.gke.io/v1beta1",
			Kind:       "RepoSync",
			MetaData: MetaDatas{
				Name:      rsName,
				Namespace: lcns,
			},
			Specification: Specs{
				Git: AnthosGit{
					// TODO query CloudConfig for Repo URL
					Repo:        p.gitProvider.Url,
					Revision:    "HEAD",
					Branch:      p.gitProvider.Branch,
					Dir:         strings.Join([]string{"/namespaces/", lcns, "/", p.gitProvider.Cluster}, ""),
					Auth:        "none",
					NoSSLVerify: false,
				},
			},
		}
		rsData, err := json.Marshal(&rs)
		if err != nil {
			log.Error("ApplyConfig:: Marshal error for RepoSync", log.Fields{"err": err, "rs": rs})
			return err
		}
		rsFilename := strings.Join([]string{p.GetPath("context"), rsName, "+RepoSync", ".json"}, "")
		gp, err = p.gitProvider.Apply(rsFilename, gp, rsData)
	}

	// Add deployed flag to commit
	path := p.GetDeployedPath("context") + "/deployed"

	t := time.Now()
	timeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(timeBytes, uint32(t.Unix()))
	gp, err = p.gitProvider.Apply(path, gp, timeBytes)

	// Commit
	err = p.gitProvider.Commit(ctx, gp)
	if err != nil {
		log.Error("ApplyConfig:: Commit files err", log.Fields{"err": err, "gp": gp})
	}
	return err
}

func (p *AnthosProvider) DeleteConfig(ctx context.Context, config interface{}) error {

	// Prepare git changes
	var gp interface{}

	acUtils, err := utils.NewAppContextReference(ctx, p.gitProvider.Cid)
	if err != nil {
		return nil
	}
	_, name, _, _, lclevel, err := acUtils.GetLogicalCloudInfo(ctx)

	if err != nil {
		return err
	}
	if lclevel == "1" {
		rsName := strings.Join([]string{"repo-sync-", name}, "")
		rsFilename := strings.Join([]string{p.GetPath("context"), rsName, "+RepoSync", ".json"}, "")

		gp, err := p.gitProvider.Delete(rsFilename, gp, nil)

		err = p.gitProvider.Commit(ctx, gp)
		if err != nil {
			log.Error("DeleteConfig:: Commit files err", log.Fields{"err": err, "gp": gp})
		}
		return err
	}
	return nil
}

/*
	Function to get path of "deployed" tracking file path
	params : string
	return : string
*/
func (p *AnthosProvider) GetDeployedPath(t string) string {
	if p.gitProvider.Level == "0" {
		return "clusters/" + p.gitProvider.Cluster + "/" + t + "/" + p.gitProvider.Cid
	} else {
		// separating by namespaces instead of logical clouds as it achieved the same goal and reduces code complexity
		return "namespaces/" + p.gitProvider.Namespace + "/" + p.gitProvider.Cluster + "/" + t + "/" + p.gitProvider.Cid
	}
}

/*
	Function to get path of files stored in git
	params : string
	return : string
*/
func (p *AnthosProvider) GetPath(t string) string {
	if p.gitProvider.Level == "0" {
		return "clusters/" + p.gitProvider.Cluster + "/" + t + "/" + p.gitProvider.Cid + "/app/" + p.gitProvider.App + "/"
	} else {
		// separating by namespaces instead of logical clouds as it achieved the same goal and reduces code complexity
		return "namespaces/" + p.gitProvider.Namespace + "/" + p.gitProvider.Cluster + "/" + t + "/" + p.gitProvider.Cid + "/app/" + p.gitProvider.App + "/"
	}
}
