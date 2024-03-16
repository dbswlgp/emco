package emcogit2go

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	git "github.com/libgit2/git2go/v33"
	v1alpha1 "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	gitUtils "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/utils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Git2go struct {
	Url        string
	Branch     string
	UserName   string
	RepoName   string
	GitToken   string
	FolderName string
}

func NewGit2Go(url, branch, user, repo, token string) (*Git2go, error) {

	folderName := "/tmp/" + user + "-" + repo
	g := Git2go{
		Url:        url,
		Branch:     branch,
		UserName:   user,
		RepoName:   repo,
		GitToken:   token,
		FolderName: folderName,
	}

	//check if git repo exists, if not clone
	check, err := gitUtils.Exists(folderName)
	if err != nil {
		return nil, err
	}
	//Clone the git repo if not already cloned
	if !check {
		if err := os.Mkdir(folderName, os.ModePerm); err != nil {
			log.Error("Error in creating the dir", log.Fields{"Error": err})
			return nil, err
		}
		// // clone the repo
		_, err = git.Clone(url, folderName, &git.CloneOptions{CheckoutBranch: branch, CheckoutOptions: git.CheckoutOptions{Strategy: git.CheckoutSafe}})
		if err != nil {
			log.Error("Error cloning the repo", log.Fields{"Error": err})
			return nil, err
		}

	}
	return &g, nil
}

func getCredCallBack(userName, token string) func(url string, username string, allowedTypes git.CredType) (*git.Credential, error) {
	return func(url string, username string, allowedTypes git.CredType) (*git.Credential, error) {
		username = userName
		password := token
		cred, err := git.NewCredUserpassPlaintext(username, password)
		return cred, err
	}
}

// CommitFile contains high-level information about a file  ed to a commit.
type CommitFile struct {

	// true if file is to be added and false for delete
	Add bool `json:"add"`
	// Path is path where this file is located.
	Path *string `json:"path"`

	FileName *string `json:"filename"`

	// Content is the content of the file.
	Content *string `json:"content,omitempty"`
}

//function to delete a file
func deleteFile(filenName string) error {
	// Removing file from the directory
	// Using Remove() function
	err := os.Remove(filenName)
	if err != nil {
		log.Error("Error in Deleting file from the tmp folder", log.Fields{"err": err})
		return err
	}

	return nil
}

//function to create a new file
func createFile(fileName string, content string) error {
	if err := os.MkdirAll(filepath.Dir(fileName), 0770); err != nil {
		return err
	}

	f, err := os.Create(fileName)

	if err != nil {
		log.Error("Error in Creating file in the tmp folder", log.Fields{"err": err})
		return err
	}

	defer f.Close()

	_, err2 := f.WriteString(content)

	if err2 != nil {
		log.Error("Error in writing file from the tmp folder", log.Fields{"err": err2})
		return err2
	}
	return nil
}

var mutex = sync.Mutex{}

// function to commit files to a branch
func (p *Git2go) CommitFiles(app, message string, files interface{}) error {

	mutex.Lock()
	defer mutex.Unlock()

	var repo *git.Repository

	branchName := p.Branch
	userName := p.UserName

	repo, err := git.OpenRepository(p.FolderName)
	if err != nil {
		log.Error("Error in Opening the git repository", log.Fields{"err": err, "branchName": branchName})
		return err
	}

	signature := &git.Signature{
		Name:  userName,
		Email: userName + "@gmail.com",
		When:  time.Now(),
	}

	// set head to point to the branch
	err = repo.SetHead("refs/heads/" + branchName)
	if err != nil {
		log.Error("Error in settting the head", log.Fields{"err": err, "branchName": branchName})
		return err
	}

	branch, err := repo.References.Lookup("refs/heads/" + branchName)
	if err != nil {
		log.Info("Error in looking up ref", log.Fields{"err": err})
		return err
	}

	//Update the index with files and obtain the latest index
	//loop through all files and update the index
	idx, err := repo.Index()
	if err != nil {
		log.Error("Error in obtaining the repo index", log.Fields{"err": err, "idx": idx})
		return err
	}
	f := convertToCommitFile(files)
	for _, file := range f {
		if file.Add {
			idx, err = addToCommit(idx, *file.Path, *file.FileName, *file.Content)
		} else {
			idx, err = deleteFromCommit(idx, *file.Path, *file.FileName)
		}

		if err != nil {
			log.Error("Error in adding or deleting file to commit", log.Fields{"err": err, "idx": idx})
			return err
		}
	}
	//commit the files to the branch
	treeId, err := idx.WriteTree()
	if err != nil {
		log.Error("Error from idx.WriteTree()", log.Fields{"err": err})
		return err
	}

	err = idx.Write()
	if err != nil {
		log.Error("Error in Deleting file from idx.Write()", log.Fields{"err": err})
		return err
	}

	tree, err := repo.LookupTree(treeId)
	if err != nil {
		log.Error("Error in looking up tree", log.Fields{"err": err, "treeId": treeId})
		return err
	}

	commitTarget, err := repo.LookupCommit(branch.Target())
	if err != nil {
		log.Error("Error in Looking up Commit for commit", log.Fields{"err": err})
		return err
	}

	_, err = repo.CreateCommit("refs/heads/"+branchName, signature, signature, message, tree, commitTarget)
	if err != nil {
		log.Error("Error in creating a commit", log.Fields{"err": err, "branchName": branchName})
		return err
	}

	//push branch to origin remote
	err = p.PushBranch(repo, branchName)
	if err != nil {
		return err
	}

	return nil
}

// function to push branch to remote origin
func (p *Git2go) PushBranch(repo *git.Repository, branchName string) error {
	// push the branch to origin

	userName := p.UserName
	token := p.GitToken

	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		log.Error("Error in obtaining remote", log.Fields{"err": err, "branchName": branchName})
		return err
	}
	cbs := &git.RemoteCallbacks{
		CredentialsCallback: getCredCallBack(userName, token),
	}

	err = remote.Push([]string{"+refs/heads/" + branchName}, &git.PushOptions{RemoteCallbacks: *cbs})
	if err != nil {
		log.Error("Error in Pushing the branch", log.Fields{"err": err, "branchName": branchName})
		return err
	}

	return nil
}

// function to add file for commit
func addToCommit(idx *git.Index, path, fileName, contents string) (*git.Index, error) {

	err := createFile(path, contents)
	if err != nil {
		return nil, err
	}
	// add file to staging area
	err = idx.AddByPath(fileName)
	if err != nil {
		return nil, err
	}
	return idx, nil

}

// function to delete file for commit
func deleteFromCommit(idx *git.Index, path, fileName string) (*git.Index, error) {

	// check if the file Exists
	check, err := gitUtils.Exists(path)
	if err != nil {
		return nil, err
	}
	if check {
		err := idx.RemoveByPath(fileName)
		if err != nil {
			return nil, err
		}
		err = deleteFile(path)
		if err != nil {
			return nil, err
		}
	}
	return idx, nil
}

// function to add file to commit files array
func (p *Git2go) AddToCommit(fileName, content string, ref interface{}) interface{} {
	path := p.FolderName + "/" + fileName
	//check if file exists
	files := append(convertToCommitFile(ref), CommitFile{
		Add:      true,
		Path:     &path,
		FileName: &fileName,
		Content:  &content,
	})

	return files

}

// function to delete file from commit files array
func (p *Git2go) DeleteToCommit(fileName string, ref interface{}) interface{} {
	path := p.FolderName + "/" + fileName
	files := append(convertToCommitFile(ref), CommitFile{
		Add:      false,
		Path:     &path,
		FileName: &fileName,
	})

	return files
}

/*
	Helper function to convert interface to []git2go.CommitFile
	params: files interface{}
	return: []git2go.CommitFile
*/
func convertToCommitFile(ref interface{}) []CommitFile {
	var exists bool
	switch ref.(type) {
	case []CommitFile:
		exists = true
	default:
		exists = false
	}
	var rf []CommitFile
	// Create rf is doesn't exist
	if !exists {
		rf = []CommitFile{}
	} else {
		rf = ref.([]CommitFile)
	}
	return rf
}

//function to create branch
func CreateBranch(folderName, branchName string) (*git.Branch, error) {
	// // // open a repo
	repo, err := git.OpenRepository(folderName)
	if err != nil {
		log.Error("Error in Opening the git repository", log.Fields{"err": err})
		return nil, err
	}

	// create the new branch
	//checkout the new branch
	err = repo.SetHead("refs/heads/" + "main")
	if err != nil {
		log.Error("Error in settting the head", log.Fields{"err": err, "branchName": branchName})
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		log.Error("Error in obtaining the head of the repo", log.Fields{"err": err})
		return nil, err
	}

	headCommit, err := repo.LookupCommit(head.Target())
	if err != nil {
		log.Error("Error in obtainging the head commit", log.Fields{"err": err, "headCommit": headCommit})
		return nil, err
	}
	branch, err := repo.CreateBranch(branchName, headCommit, false)
	if err != nil {
		log.Error("Error in Creating branch", log.Fields{"err": err, "branchName": branchName, "headCommit": headCommit, "branch": branch})
		return nil, err
	}

	return branch, nil
}

//function to pull branch
func (p *Git2go) GitPull(folderName, branchName string) error {
	url := p.Url
	userName := p.UserName
	check, err := gitUtils.Exists(folderName)

	if !check {
		if err := os.Mkdir(folderName, os.ModePerm); err != nil {
			return err
		}
		// // clone the repo
		_, err := git.Clone(url, folderName, &git.CloneOptions{CheckoutBranch: branchName, CheckoutOptions: git.CheckoutOptions{Strategy: git.CheckoutSafe}})
		if err != nil {
			log.Error("Error in cloning repo Git Pull", log.Fields{"err": err})
			// remove the local folder
			err := os.RemoveAll(folderName)
			return err
		}
	}
	// // // open a repo
	repo, err := git.OpenRepository(folderName)
	if err != nil {
		log.Error("Error in Opening the git repository", log.Fields{"err": err})
		return err
	}

	// Locate remote
	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		log.Error("Error in looking up Origin", log.Fields{"err": err})
		return err
	}

	// Fetch changes from remote
	if err := remote.Fetch([]string{}, nil, ""); err != nil {
		log.Error("Error in Fetching", log.Fields{"err": err})
		return err
	}

	// check if a local branch exist if not do a checkout
	localBranch, err := repo.LookupBranch(branchName, git.BranchLocal)
	// No local branch, lets checkout
	if localBranch == nil || err != nil {
		err = CheckoutBranch(folderName, branchName)
		if err != nil {
			log.Error("Git checkout error", log.Fields{"err": err, "branchName": branchName})
			return err
		}
		return nil
	}
	// set head to point to the created branch
	err = repo.SetHead("refs/heads/" + branchName)
	if err != nil {
		log.Error("Error in settting the head", log.Fields{"err": err, "branchName": branchName})
		return err
	}

	// Get remote master
	remoteBranch, err := repo.References.Lookup("refs/remotes/origin/" + branchName)
	if err != nil {
		return err
	}

	remoteBranchID := remoteBranch.Target()
	// Get annotated commit
	annotatedCommit, err := repo.AnnotatedCommitFromRef(remoteBranch)

	if err != nil {
		return err
	}

	// Do the merge analysis
	mergeHeads := make([]*git.AnnotatedCommit, 1)
	mergeHeads[0] = annotatedCommit
	analysis, _, err := repo.MergeAnalysis(mergeHeads)
	if err != nil {
		return err
	}

	log.Debug("Git Pull Information ", log.Fields{"remoteBranch": remoteBranch, "remoteBranchID": remoteBranchID, "annotatedCommit": annotatedCommit, "analysis": analysis, "mergeHeads": mergeHeads})

	head, err := repo.Head()
	if err != nil {
		return err
	}

	if analysis&git.MergeAnalysisUpToDate != 0 {
		log.Debug("MergeAnalysisUpToDate", log.Fields{"analysis": analysis})
		return nil
	} else if analysis&git.MergeAnalysisNormal != 0 {
		log.Debug("MergeAnalysisNormal", log.Fields{"analysis": analysis})

		// Just merge changes
		if err := repo.Merge([]*git.AnnotatedCommit{annotatedCommit}, nil, nil); err != nil {
			log.Error("Merge Error", log.Fields{"err": err})
			return err
		}
		// Check for conflicts
		index, err := repo.Index()
		if err != nil {
			return err
		}

		if index.HasConflicts() {
			return errors.New("Conflicts encountered. Please resolve them.")
		}

		// Make the merge commit
		sig := &git.Signature{Name: userName, Email: userName + "@cool", When: time.Now()}
		if err != nil {
			return err
		}

		// Get Write Tree
		treeId, err := index.WriteTree()
		if err != nil {
			return err
		}

		// Get repo head

		tree, err := repo.LookupTree(treeId)
		if err != nil {
			return err
		}

		head, err := repo.Head()
		if err != nil {
			return err
		}
		localCommit, err := repo.LookupCommit(head.Target())
		if err != nil {
			return err
		}

		remoteCommit, err := repo.LookupCommit(remoteBranchID)
		if err != nil {
			return err
		}

		_, err = repo.CreateCommit("HEAD", sig, sig, "", tree, localCommit, remoteCommit)
		if err != nil {
			return err
		}
		// Clean up
		repo.StateCleanup()
	} else if analysis&git.MergeAnalysisFastForward != 0 {
		log.Debug("MergeAnalysisFastForward", log.Fields{"analysis": analysis})
		// Fast-forward changes
		// Get remote tree
		remoteTree, err := repo.LookupTree(remoteBranchID)
		if err != nil {
			return err
		}

		// Checkout
		if err := repo.CheckoutTree(remoteTree, nil); err != nil {
			return err
		}

		branchRef, err := repo.References.Lookup("refs/heads/" + branchName)
		if err != nil {
			return err
		}

		// Point branch to the object
		branchRef.SetTarget(remoteBranchID, "")
		if _, err := head.SetTarget(remoteBranchID, ""); err != nil {
			return err
		}

	} else {
		log.Info("Unexpected merge analysis result", log.Fields{"analysis": analysis})
		return fmt.Errorf("Unexpected merge analysis result %d", analysis)
	}
	return nil
}

//function to checkout a branch
func CheckoutBranch(folderName, branchName string) error {
	repo, err := git.OpenRepository(folderName)
	if err != nil {
		return err
	}
	checkoutOpts := &git.CheckoutOpts{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing | git.CheckoutAllowConflicts | git.CheckoutUseTheirs,
	}
	//Getting the reference for the remote branch
	remoteBranch, err := repo.References.Lookup("refs/remotes/origin/" + branchName)
	if err != nil {
		log.Error("Failed to find remote branch: ", log.Fields{"branchName": branchName})
		return err
	}
	defer remoteBranch.Free()

	// Lookup for commit from remote branch
	commit, err := repo.LookupCommit(remoteBranch.Target())
	if err != nil {
		log.Error("Failed to find remote branch commit: ", log.Fields{"branchName": branchName})
		return err
	}
	defer commit.Free()

	localBranch, err := repo.LookupBranch(branchName, git.BranchLocal)
	// No local branch, lets create one
	if localBranch == nil || err != nil {
		// Creating local branch
		localBranch, err = repo.CreateBranch(branchName, commit, false)
		if err != nil {
			log.Error("Failed to create local branch: ", log.Fields{"branchName": branchName})
			return err
		}

		// Setting upstream to origin branch
		err = localBranch.SetUpstream("origin/" + branchName)
		if err != nil {
			log.Error("Failed to create upstream to origin/ ", log.Fields{"branchName": branchName})
			return err
		}
	}
	if localBranch == nil {
		return errors.New("Error while locating/creating local branch")
	}
	defer localBranch.Free()

	// Getting the tree for the branch
	localCommit, err := repo.LookupCommit(localBranch.Target())
	if err != nil {
		log.Error("Failed to lookup for commit in local branch ", log.Fields{"branchName": branchName})
		return err
	}
	defer localCommit.Free()

	tree, err := repo.LookupTree(localCommit.TreeId())
	if err != nil {
		log.Error("Failed to lookup for tree", log.Fields{"branchName": branchName})
		return err
	}
	defer tree.Free()

	// Checkout the tree
	err = repo.CheckoutTree(tree, checkoutOpts)
	if err != nil {
		log.Error("Failed to checkout tree ", log.Fields{"branchName": branchName})
		return err
	}
	return nil
}

//Function to get files in a path
func GetFilesInPath(path string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		file, err := os.Open(path)
		fileInfo, err := file.Stat()
		if err != nil {
			return err
		}

		//Add only if the path is not a directory
		if !fileInfo.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

//Function to get File Contents
func GetFileContent(filePath string) (string, error) {

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Convert []byte to string and print to screen
	text := string(content)

	return text, nil
}

//function to get the latest commit
func GetLatestCommit(path, branchName string) (*git.Oid, error) {
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Error("Error in Opening repo", log.Fields{"err": err, "path": path})
		return nil, err
	}
	// set head to point to the created branch
	err = repo.SetHead("refs/heads/" + branchName)
	if err != nil {
		log.Error("Error in setting head", log.Fields{"err": err, "branchName": branchName})
		return nil, err
	}
	head, err := repo.Head()
	if err != nil {
		log.Error("Error obtaining head", log.Fields{"err": err, "head": head})
		return nil, err
	}
	headCommit := head.Target()

	return headCommit, nil
}

// function to push branch to remote origin
func PushDeleteBranch(repo *git.Repository, branchName, userName, token string) error {
	// push the branch to origin
	remote, err := repo.Remotes.Lookup("origin")

	cbs := &git.RemoteCallbacks{
		CredentialsCallback: getCredCallBack(userName, token),
	}

	err = remote.Push([]string{":refs/heads/" + branchName}, &git.PushOptions{RemoteCallbacks: *cbs})
	if err != nil {
		log.Error("Error in Pushing the branch", log.Fields{"err": err, "branchName": branchName})
		return err
	}

	return nil
}

//function to merge branch to main
func mergeToMain(repo *git.Repository, branchName string, signature *git.Signature) error {
	// get reference for the target merge branch
	mergeBranch, err := repo.References.Lookup("refs/heads/" + branchName)
	if err != nil {
		log.Error("Error in obtaining the reference for branch to merge to main", log.Fields{"err": err, "branchName": branchName})
		return err
	}

	mergeHeadMergeBranch, err := repo.AnnotatedCommitFromRef(mergeBranch)
	if err != nil {
		log.Error("Error in obtaining the head of the branch to merge", log.Fields{"err": err, "mergeHeadMergeBranch": mergeHeadMergeBranch})
		return err
	}
	mergeHeads := make([]*git.AnnotatedCommit, 1)

	mergeHeads[0] = mergeHeadMergeBranch

	err = repo.Merge(mergeHeads, nil, nil)
	if err != nil {
		log.Error("Error in Merging the branch", log.Fields{"err": err, "mergeHeads": mergeHeads})
		return err
	}

	mergeMessage, err := repo.Message()
	if err != nil {
		return err
	}

	log.Debug("Merge Message", log.Fields{"mergeMessage": mergeMessage})

	err = commitMergeToMaster(repo, signature, "Merge commit to main")
	if err != nil {
		log.Error("Error in commit Merge to main", log.Fields{"err": err})
		return err
	}

	return nil
}

//function to delete branch
func DeleteBranch(repo *git.Repository, branchName string) error {
	branchA, err := repo.LookupBranch(branchName, git.BranchLocal)
	err = branchA.Delete()
	if err != nil {
		return err
	}
	return nil
}

func commitMergeToMaster(repo *git.Repository, signature *git.Signature, message string) error {
	//commit the merge to main
	branchName := "main"
	idx, err := repo.Index()
	if err != nil {
		log.Error("commitMergeToMaster: Error in obtaining the repo index", log.Fields{"err": err, "idx": idx})
		return err
	}

	branchMain, err := repo.LookupBranch(branchName, git.BranchLocal)
	if err != nil {
		return err
	}

	treeId, err := idx.WriteTree()
	if err != nil {
		log.Error("commitMergeToMaster: Error from idx.WriteTree()", log.Fields{"err": err})
		return err
	}

	err = idx.Write()
	if err != nil {
		log.Error("commitMergeToMaster: Error in Deleting file from idx.Write()", log.Fields{"err": err})
		return err
	}

	tree, err := repo.LookupTree(treeId)
	if err != nil {
		log.Error("commitMergeToMaster: Error in looking up tree", log.Fields{"err": err, "treeId": treeId})
		return err
	}

	commitTarget, err := repo.LookupCommit(branchMain.Target())
	if err != nil {
		log.Error("commitMergeToMaster: Error in Looking up Commit for commit", log.Fields{"err": err})
		return err
	}

	_, err = repo.CreateCommit("refs/heads/"+branchName, signature, signature, message, tree, commitTarget)
	if err != nil {
		log.Error("commitMergeToMaster:Error in creating a commit", log.Fields{"err": err, "branchName": branchName})
		return err
	}

	return nil
}

func (p *Git2go) ClusterWatcher(ctx context.Context, cid, app, cluster string, waitTime int) error {

	// foldername for status
	folderName := "/tmp/" + cluster + "-status"
	// Start thread to sync monitor CR
	var lastCommitSHA *git.Oid
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
					p.DeleteClusterStatusCR(cid, app, cluster)
					// AppContext deleted - Exit thread
					return nil
				}
				path := "clusters/" + cluster + "/" + "status" + "/" + cid + "/app/" + app + "/"
				// branch to track
				branch := cluster

				// // git pull
				mutex.Lock()
				err := p.GitPull(folderName, branch)
				mutex.Unlock()

				if err != nil {
					log.Error("Error in Pulling Branch", log.Fields{"err": err})
					continue
				}

				//obtain the latest commit SHA
				mutex.Lock()
				latestCommitSHA, err := GetLatestCommit(folderName, branch)
				mutex.Unlock()
				if err != nil {
					log.Error("Error in obtaining latest commit SHA", log.Fields{"err": err})
				}

				if lastCommitSHA != latestCommitSHA || lastCommitSHA == nil {
					log.Debug("New Status File, pulling files", log.Fields{"LatestSHA": latestCommitSHA, "LastSHA": lastCommitSHA})
					files, err := GetFilesInPath(folderName + "/" + path)
					if err != nil {
						log.Debug("Status file not available", log.Fields{"error": err, "resource": path})
						continue
					}
					log.Info("Files to track status", log.Fields{"files": files})

					if len(files) > 0 {
						// Only one file expected in the location
						fileContent, err := GetFileContent(files[0])
						if err != nil {
							log.Error("", log.Fields{"error": err, "cluster": cluster, "resource": path})
							return err
						}
						content := &v1alpha1.ResourceBundleState{}
						_, err = utils.DecodeYAMLData(fileContent, content)
						if err != nil {
							log.Error("", log.Fields{"error": err, "cluster": cluster, "resource": path})
							return err
						}
						status.HandleResourcesStatus(ctx, cid, app, cluster, content)
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

// DeleteClusterStatusCR deletes the status CR provided by the monitor on the cluster
func (p *Git2go) DeleteClusterStatusCR(cid, app, cluster string) error {

	// // // // open a repo
	//obtain files to be delete
	path := "clusters/" + cluster + "/context/" + cid + "/app/" + app + "/" + cid + "-" + app + ".yaml"
	check, err := gitUtils.Exists(p.FolderName + "/" + path)
	if err != nil {
		return err
	}
	var ref interface{}
	if check {
		files := p.DeleteToCommit(path, ref)
		err := p.CommitFiles(app, "Deleting status CR files "+path, files)
		if err != nil {
			log.Error("Error in commiting files to Delete", log.Fields{"path": path})
		}
	}
	return nil
}

// function to commit files to a branch
func (p *Git2go) CommitStatus(commitMessage, branchName, cid, app string, files interface{}) error {

	userName := p.UserName
	folderName := p.FolderName
	// // // open a repo
	repo, err := git.OpenRepository(folderName)
	if err != nil {
		log.Error("Error in Opening the git repository", log.Fields{"err": err, "branchName": branchName})
		return err
	}

	signature := &git.Signature{
		Name:  userName,
		Email: userName + "@gmail.com",
		When:  time.Now(),
	}

	var targetID *git.Oid
	//create a new branch
	//check if branch already exists, if yes then skip create branch
	// check if a local branch exists, if not do a checkout
	localBranch, err := repo.LookupBranch(branchName, git.BranchLocal)
	// No local branch, lets create one
	if localBranch == nil || err != nil {
		branchHandle, err := CreateBranch(folderName, branchName)
		if err != nil {
			if !strings.Contains(err.Error(), "a reference with that name already exists") {
				return err
			}
		}
		targetID = branchHandle.Target()
	} else {
		branchRef, err := repo.References.Lookup("refs/heads/" + branchName)
		if err != nil {
			log.Info("Error in looking up ref", log.Fields{"err": err})
			return err
		}

		targetID = branchRef.Target()
	}

	//commit files to the branch
	//push the branch

	// set head to point to the created branch
	err = repo.SetHead("refs/heads/" + branchName)
	if err != nil {
		log.Error("Error in settting the head", log.Fields{"err": err, "branchName": branchName})
		return err
	}

	//Update the index with files and obtain the latest index
	//loop through all files and update the index
	idx, err := repo.Index()
	if err != nil {
		log.Error("Error in obtaining the repo index", log.Fields{"err": err, "idx": idx})
		return err
	}
	f := convertToCommitFile(files)
	for _, file := range f {
		if file.Add {
			idx, err = addToCommit(idx, *file.Path, *file.FileName, *file.Content)
		} else {
			idx, err = deleteFromCommit(idx, *file.Path, *file.FileName)
		}

		if err != nil {
			log.Error("Error in adding or deleting file to commit", log.Fields{"err": err, "idx": idx})
			return err
		}
	}
	//commit the files to the branch
	treeId, err := idx.WriteTree()
	if err != nil {
		log.Error("Error from idx.WriteTree()", log.Fields{"err": err})
		return err
	}

	err = idx.Write()
	if err != nil {
		log.Error("Error in Deleting file from idx.Write()", log.Fields{"err": err})
		return err
	}

	tree, err := repo.LookupTree(treeId)
	if err != nil {
		log.Error("Error in looking up tree", log.Fields{"err": err, "treeId": treeId})
		return err
	}

	commitTarget, err := repo.LookupCommit(targetID)
	if err != nil {
		log.Error("Error in Looking up Commit for commit", log.Fields{"err": err})
		return err
	}

	_, err = repo.CreateCommit("refs/heads/"+branchName, signature, signature, commitMessage, tree, commitTarget)
	if err != nil {
		log.Error("Error in creating a commit", log.Fields{"err": err, "branchName": branchName})
		return err
	}
	err = p.PushBranch(repo, branchName)

	return nil
}
