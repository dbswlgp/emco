// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package emcogithub

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	gogithub "github.com/google/go-github/v41/github"
	"gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const (
	githubDomain = "github.com"
	maxrand      = 0x7fffffffffffffff
)

type GithubClient struct {
	gitProviderClient gitprovider.Client
	gogithubClient    *gogithub.Client
}

type Github struct {
	Url      string
	Branch   string
	UserName string
	RepoName string
	GitToken string
	Cluster  string
	Client   interface{}
}

func NewGithub(cluster, url, branch, user, repo, token string) (*Github, error) {

	g := Github{
		Url:      url,
		Branch:   branch,
		UserName: user,
		RepoName: repo,
		GitToken: token,
		Cluster:  cluster,
	}
	client, err := CreateClient(user, token)
	if err != nil {
		log.Error("Error getting git client", log.Fields{"err": err})
		return nil, err
	}
	g.Client = client

	return &g, nil
}

/*
	Function to create githubClient
	params : github token
	return : github client, error
*/
func CreateClient(userName, githubToken string) (GithubClient, error) {

	var client GithubClient
	var err error

	client.gitProviderClient, err = github.NewClient(gitprovider.WithOAuth2Token(githubToken), gitprovider.WithDestructiveAPICalls(true))
	if err != nil {
		return GithubClient{}, err
	}

	tp := gogithub.BasicAuthTransport{
		Username: userName,
		Password: githubToken,
	}
	client.gogithubClient = gogithub.NewClient(tp.Client())

	return client, nil

}

/*
	Helper function to convert interface to GithubClient
	params: files interface{}
	return: GithubClient
*/
func convertToClient(c interface{}) GithubClient {
	return c.(GithubClient)
}

/*
	Function to create a new Repo in github
	params : context, github client, Repository Name, User Name, description
	return : nil/error
*/
func CreateRepo(ctx context.Context, c interface{}, repoName string, userName string, desc string) error {

	// obtain client
	client := convertToClient(c)

	// create repo reference
	userRepoRef := getRepoRef(userName, repoName)

	// Create repoinfo reference
	userRepoInfo := gitprovider.RepositoryInfo{
		Description: &desc,
		Visibility:  gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPublic),
	}

	// Create the repository
	_, err := client.gitProviderClient.UserRepositories().Create(ctx, userRepoRef, userRepoInfo, &gitprovider.RepositoryCreateOptions{
		AutoInit:        gitprovider.BoolVar(true),
		LicenseTemplate: gitprovider.LicenseTemplateVar(gitprovider.LicenseTemplateApache2),
	})

	if err != nil {
		return err
	}
	log.Info("Repo Created", log.Fields{})

	return nil
}

/*
	Function to commit multiple files to the github repo
	params : context, github client, User Name, Repo Name, BranchName, Commit Message, files ([]gitprovider.CommitFile)
	return : nil/error
*/
func (p *Github) CommitFiles(app, commitMessage string, files interface{}) error {

	// obtain client
	client := convertToClient(p.Client)

	n := 0
	var ctx context.Context
	ctx = context.Background()
	for {
		// obtain the sha key for main
		//obtain sha
		latestSHA, err := GetLatestCommitSHA(ctx, client, p.UserName, p.RepoName, p.Branch, "")
		if err != nil {
			return err
		}
		//create a new branch from main
		ra := rand.New(rand.NewSource(time.Now().UnixNano()))
		rn := ra.Int63n(maxrand)
		id := fmt.Sprintf("%v", rn)

		mergeBranch := app + "-" + id
		err = CreateBranch(ctx, client, latestSHA, p.UserName, p.RepoName, mergeBranch)

		// defer deletion of the created branch
		defer DeleteBranch(ctx, client, p.UserName, p.RepoName, mergeBranch)

		if err != nil {
			return err
		}

		// commit the files to this new branch
		// create repo reference
		userRepoRef := getRepoRef(p.UserName, p.RepoName)

		userRepo, err := client.gitProviderClient.UserRepositories().Get(ctx, userRepoRef)
		if err != nil {
			return err
		}

		//Commit file to this repo
		resp, err := userRepo.Commits().Create(ctx, mergeBranch, commitMessage, convertToCommitFile(files))
		if err != nil {
			log.Error("Error in commiting the files", log.Fields{"err": err, "mergeBranch": mergeBranch, "commitMessage": commitMessage, "files": files})
			return err
		}
		log.Debug("CommitResponse for userRepo:", log.Fields{"resp": resp})

		// merge the branch to the main
		err = mergeBranchToMain(ctx, client, p.UserName, p.RepoName, p.Branch, mergeBranch)

		if err != nil {
			// check error for merge conflict "409 Merge conflict"
			if strings.Contains(err.Error(), "409 Merge conflict") && n < 3 {
				// Merge conflict flag
				n++
				log.Error("Merge Conflict, trying again!", log.Fields{"err": err})
				continue
			} else {
				return err
			}
		}
		return nil
	}
	return nil
}

/*
	Function to delete repo
	params : context, github client , user name, repo name
	return : nil/error
*/
func DeleteRepo(ctx context.Context, c interface{}, userName string, repoName string) error {

	// obtain client
	client := convertToClient(c)

	// create repo reference
	userRepoRef := getRepoRef(userName, repoName)
	// get the reference of the repo to be deleted
	userRepo, err := client.gitProviderClient.UserRepositories().Get(ctx, userRepoRef)

	if err != nil {
		return err
	}
	//delete repo
	err = userRepo.Delete(ctx)

	if err != nil {
		return err
	}

	return nil
}

/*
	Internal function to create a repo refercnce
	params : user name, repo name
	return : repo reference
*/
func getRepoRef(userName string, repoName string) gitprovider.UserRepositoryRef {
	// Create the user reference
	userRef := gitprovider.UserRef{
		Domain:    githubDomain,
		UserLogin: userName,
	}

	// Create the repo reference
	userRepoRef := gitprovider.UserRepositoryRef{
		UserRef:        userRef,
		RepositoryName: repoName,
	}

	return userRepoRef
}

func convertToCommitFile(ref interface{}) []gitprovider.CommitFile {
	var exists bool
	switch ref.(type) {
	case []gitprovider.CommitFile:
		exists = true
	default:
		exists = false
	}
	var rf []gitprovider.CommitFile
	// Create rf is doesn't exist
	if !exists {
		rf = []gitprovider.CommitFile{}
	} else {
		rf = ref.([]gitprovider.CommitFile)
	}
	return rf
}

/*
	Function to Add file to the commit
	params : path , content, files (gitprovider commitfile array)
	return : files (gitprovider commitfile array)
*/
func (p *Github) AddToCommit(path, content string, ref interface{}) interface{} {
	files := append(convertToCommitFile(ref), gitprovider.CommitFile{
		Path:    &path,
		Content: &content,
	})

	return files
}

/*
	Function to Delete file from the commit
	params : path, files (gitprovider commitfile array)
	return : files (gitprovider commitfile array)
*/
func (p *Github) DeleteToCommit(path string, ref interface{}) interface{} {
	files := append(convertToCommitFile(ref), gitprovider.CommitFile{
		Path:    &path,
		Content: nil,
	})

	return files
}

/*
	Function to get files to the github repo
	params : context, github client, User Name, Repo Name, Branch Name, path)
	return : []*gitprovider.CommitFile, nil/error
*/
func GetFiles(ctx context.Context, c interface{}, userName string, repoName string, branch string, path string) ([]*gitprovider.CommitFile, error) {

	// obtain client
	client := convertToClient(c)

	// create repo reference
	userRepoRef := getRepoRef(userName, repoName)
	userRepo, err := client.gitProviderClient.UserRepositories().Get(ctx, userRepoRef)
	if err != nil {
		return nil, err
	}
	// Read the files
	cf, err := userRepo.Files().Get(ctx, path, branch)
	if err != nil {
		return nil, err
	}
	return cf, nil
}

/*
	Function to obtaion the SHA of latest commit
	params : context, github client, User Name, Repo Name, Branch, Path
	return : LatestCommit string, error
*/
func GetLatestCommitSHA(ctx context.Context, c interface{}, userName, repoName, branch, path string) (string, error) {

	// obtain client
	client := convertToClient(c)

	perPage := 1
	page := 1

	lcOpts := &gogithub.CommitsListOptions{
		ListOptions: gogithub.ListOptions{
			PerPage: perPage,
			Page:    page,
		},
		SHA:  branch,
		Path: path,
	}
	//Get the latest SHA
	resp, _, err := client.gogithubClient.Repositories.ListCommits(ctx, userName, repoName, lcOpts)
	if err != nil {
		log.Error("Error in obtaining the list of commits", log.Fields{"err": err})
		return "", err
	}
	if len(resp) == 0 {
		log.Info("File not created yet.", log.Fields{"Latest Commit Array": resp})
		return "", nil
	}
	latestCommitSHA := *resp[0].SHA

	return latestCommitSHA, nil
}

/*
	function to create new branch from main
	params : context, go git client, latestCommitSHA, User Name, Repo Name, branch
	return : error
*/
func CreateBranch(ctx context.Context, c interface{}, latestCommitSHA, userName, repoName, branch string) error {

	// obtain client
	client := convertToClient(c)

	// create a new branch
	ref, _, err := client.gogithubClient.Git.CreateRef(ctx, userName, repoName, &gogithub.Reference{
		Ref: gogithub.String("refs/heads/" + branch),
		Object: &gogithub.GitObject{
			SHA: gogithub.String(latestCommitSHA),
		},
	})
	if err != nil {
		log.Error("Git.CreateRef returned error:", log.Fields{"err": err})
		return err

	}
	log.Info("Branch Created: ", log.Fields{"ref": ref})
	return nil
}

/*
	function to merge the branch to main
	params : context, go git client, User Name, Repo Name,branch, mergeBranch
	return : LatestCommit string, error
*/
func mergeBranchToMain(ctx context.Context, c interface{}, userName, repoName, branch, mergeBranch string) error {
	// obtain client
	client := convertToClient(c)

	// merge the branch
	input := &gogithub.RepositoryMergeRequest{
		Base:          gogithub.String(branch),
		Head:          gogithub.String(mergeBranch),
		CommitMessage: gogithub.String("merging " + mergeBranch + " to " + branch),
	}

	commit, _, err := client.gogithubClient.Repositories.Merge(ctx, userName, repoName, input)
	if err != nil {
		log.Error("Error occured while Merging", log.Fields{"err": err})
		return err
	}

	log.Info("Branch Merged, Merge response:", log.Fields{"commit": commit})

	return nil

}

/*
	Function to delete the branch
	params : context, go git client, User Name, Repo Name, mergeBranch
	return : LatestCommit string, error
*/
func DeleteBranch(ctx context.Context, c interface{}, userName, repoName, mergeBranch string) error {

	// obtain client
	client := convertToClient(c)

	// Delete the Git branch
	_, err := client.gogithubClient.Git.DeleteRef(ctx, userName, repoName, "refs/heads/"+mergeBranch)
	if err != nil {
		log.Error("Git.DeleteRef returned error: ", log.Fields{"err": err})
		return err
	}
	log.Info("Branch Deleted", log.Fields{"mergeBranch": mergeBranch})
	return nil
}

/*
	Function to check if file exists
	params : context, go git client, User Name, Repo Name, Branch, Path
	return : LatestCommit string, error
*/
func CheckIfFileExists(ctx context.Context, c interface{}, userName, repoName, branch, path string) (bool, error) {

	latestSHA, err := GetLatestCommitSHA(ctx, c, userName, repoName, branch, path)
	if err != nil {
		return false, err
	}

	if latestSHA == "" {
		return false, nil
	}

	return true, nil

}

func (p *Github) ClusterWatcher(ctx context.Context, cid, app, cluster string, waitTime int) error {

	// obtain the sha key for main
	latestSHA, err := GetLatestCommitSHA(ctx, p.Client, p.UserName, p.RepoName, p.Branch, "")
	if err != nil {
		return err
	}
	// create branch for status
	branch := cluster + "-" + cid + "-" + app
	err = CreateBranch(ctx, p.Client, latestSHA, p.UserName, p.RepoName, branch)
	if err != nil {
		if !strings.Contains(err.Error(), "422 Reference already exists") {
			return err
		}
	}

	// Start thread to sync monitor CR
	go func() error {
		// This function is executed asynchronously, so we must create
		// a new (not derived) context to prevent the context from
		// being cancelled when the caller completes: a cancelled
		// context will cause the below work to exit early.  A link is
		// used so that the traces can be associated.
		tracer := otel.Tracer("rsync")
		ctx, span := tracer.Start(context.Background(), "StartClusterWatcher",
			trace.WithLinks(trace.LinkFromContext(ctx)),
		)
		defer span.End()

		var lastCommitSHA string
		for {
			select {
			case <-time.After(time.Duration(waitTime) * time.Second):
				if ctx.Err() != nil {
					return ctx.Err()
				}
				// Check if AppContext doesn't exist then exit the thread
				if _, err := utils.NewAppContextReference(ctx, cid); err != nil {
					// Delete the Status CR updated by Monitor running on the cluster
					log.Info("Deleting cluster StatusCR", log.Fields{})
					p.DeleteClusterStatusCR(ctx, cid, app, cluster)
					// AppContext deleted - Exit thread
					return nil
				}
				//path := p.GetPath("status")
				path := "clusters/" + cluster + "/" + "status" + "/" + cid + "/app/" + app + "/"
				// branch to track
				branch := p.Cluster + "-" + cid + "-" + app

				latestCommitSHA, err := GetLatestCommitSHA(ctx, p.Client, p.UserName, p.RepoName, branch, path)
				if err != nil {
					log.Error("Error in obtaining latest commit SHA", log.Fields{"err": err})
				}

				if lastCommitSHA != latestCommitSHA {
					// new commit get files
					// Read file
					log.Debug("New Status File, pulling files", log.Fields{"LatestSHA": latestCommitSHA, "LastSHA": lastCommitSHA})
					c, err := GetFiles(ctx, p.Client, p.UserName, p.RepoName, branch, path)
					if err != nil {
						log.Debug("Status file not available", log.Fields{"error": err, "cluster": p.Cluster, "resource": path})
						continue
					}
					//cp := c.([]*gitprovider.CommitFile)
					if len(c) > 0 {
						// Only one file expected in the location
						content := &v1alpha1.ResourceBundleState{}
						_, err := utils.DecodeYAMLData(*c[0].Content, content)
						if err != nil {
							log.Error("", log.Fields{"error": err, "cluster": p.Cluster, "resource": path})
							return err
						}
						status.HandleResourcesStatus(ctx, cid, app, p.Cluster, content)
					}
					lastCommitSHA = latestCommitSHA
				}

			// Check if the context is canceled
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}()
	return nil
}

func (p *Github) DeleteClusterStatusCR(ctx context.Context, cid, app, cluster string) error {

	//Delete the CR from context folder
	var ref interface{}
	path := "clusters/" + cluster + "/context/" + cid + "/app/" + app + "/" + cid + "-" + app + ".yaml"
	files := p.DeleteToCommit(path, ref)
	err := p.CommitFiles(app, "Deleting status CR files "+path, files)
	if err != nil {
		log.Error("Error in commiting files to Delete", log.Fields{"path": path})
	}

	// Delete the status branch
	branch := p.Cluster + "-" + cid + "-" + app

	// ctx := context.Background()

	// Delete the branch
	err = DeleteBranch(ctx, p.Client, p.UserName, p.RepoName, branch)
	if err != nil {
		log.Error("Error in deleting branch", log.Fields{"err": err})
		return err
	}

	return nil

}

/*
	Function to commit multiple files to the github repo
	params : context, Branch Name, Commit Message, appName, files ([]gitprovider.CommitFile)
	return : nil/error
*/
func (p *Github) CommitStatus(commitMessage, branchName, cid, app string, files interface{}) error {

	// obtain client
	client := convertToClient(p.Client)
	mergeBranch := branchName + "-" + cid + "-" + app
	ctx := context.Background()

	// commit the files to this new branch
	// create repo reference
	log.Info("Creating Repo Reference. ", log.Fields{})
	userRepoRef := getRepoRef(p.UserName, p.RepoName)
	log.Info("UserRepoRef:", log.Fields{"UserRepoRef": userRepoRef})

	log.Info("Obtaining user repo. ", log.Fields{})
	userRepo, err := client.gitProviderClient.UserRepositories().Get(ctx, userRepoRef)
	if err != nil {
		log.Error("Error in commiting the files", log.Fields{"err": err, "mergeBranch": mergeBranch, "commitMessage": commitMessage, "files": files})
		return err
	}
	log.Info("UserRepo:", log.Fields{"UserRepo": userRepo})

	log.Info("Commiting Files:", log.Fields{"files": files})
	//Commit file to this repo
	resp, err := userRepo.Commits().Create(ctx, mergeBranch, commitMessage, convertToCommitFile(files))
	if err != nil {
		if !strings.Contains(err.Error(), "404 Not Found") {
			log.Error("Error in commiting the files", log.Fields{"err": err, "mergeBranch": mergeBranch, "commitMessage": commitMessage, "files": files})
		}
		return err

	}
	log.Info("CommitResponse for userRepo:", log.Fields{"resp": resp})
	return nil
}
