package mgmirr

import (
	"fmt"
	"github.com/libgit2/git2go"
	"log"
	"strings"
)

type RemoteConfig struct {
	Name string
	URL  string
}

// For an existing Git repo and an RPM (e.g. cowsay) Setup the remotes.
//
// This is a best effort procedure.  Not all remotes will be available
// (fedora might not have package x).  As long as at least one remote
// works it is a success.
func SetupRpmRemotes(repo *git.Repository, rcs []RemoteConfig) error {

	var one_worked bool = false

	for _, rc := range rcs {

		// try to set up the remote, continue if it doesn't work
		err := setupRpmRemote(repo, &rc)
		if err != nil {
			log.Println(err)
		} else {
			one_worked = true
		}
	}

	if one_worked {
		return nil
	} else {
		return fmt.Errorf("unable to setup any remotes")
	}
}

func setupRpmRemote(repo *git.Repository, cfg *RemoteConfig) error {
	_, err := repo.Remotes.Create(cfg.Name, cfg.URL)
	if err != nil {
		return fmt.Errorf("git add remote for '%v' failed: %v", cfg.Name, err)
	}

	return nil
}

func FetchAll(repo *git.Repository) error {
	var one_worked bool = false

	remotes, err := repo.Remotes.List()
	if err != nil {
		return fmt.Errorf("unable to list remotes: %v", err)
	}

	for _, remote := range remotes {
		r, err := repo.Remotes.Lookup(remote) // get Remote obj
		if err != nil {
			log.Printf("unable to find remote '%v': %v", remote, err)
			continue
		}

		err = r.Fetch(nil, nil, "")
		if err != nil {
			log.Printf("git fetch remote '%v' failed: %v", remote, err)
		} else {
			one_worked = true
		}
	}

	if one_worked {
		return nil
	} else {
		return fmt.Errorf("unable to fetch any remotes")
	}
}

// For a repo with remote branches the expected local branch name
// is the same but with "remotes/" removed.
//  remotes/fedora/f31 -> fedora/31
// This gets the set of local branches (e.g. fedora/f31) that "should"
// exist based on the remotes that were found.
func getExpectedLocalBranches(repo *git.Repository) ([]string, error) {

	var branches []string
	iter, err := repo.NewBranchIterator(git.BranchRemote)
	if err != nil {
		return nil, err
	}
	defer iter.Free()
	for {
		ref, branch_type, err := iter.Next()
		if err != nil {
			break
		}
		if branch_type != git.BranchRemote {
			continue
		}
		branch, _ := ref.Branch().Name() // fedora/f31
		branches = append(branches, branch)
	}

	return branches, nil
}

func setupRpmBranch(repo *git.Repository, branch string) error {

	var err error

	// first, create a new local branch from the remote

	remote_branch, err := repo.LookupBranch(branch, git.BranchRemote)
	if err != nil {
		return fmt.Errorf("unable to find remote '%s': %v", branch, err)
	}
	defer remote_branch.Free()

	commit, err := repo.LookupCommit(remote_branch.Target())
	if err != nil {
		return fmt.Errorf("lookup commit failed: %v", err)
	}
	defer commit.Free()

	local_branch, err := repo.LookupBranch(branch, git.BranchLocal)
	if local_branch == nil || err != nil {
		local_branch, err = repo.CreateBranch(branch, commit, false)
		if err != nil {
			return fmt.Errorf("create branch '%s' failed: %v", branch, err)
		}
	}
	if local_branch == nil {
		return fmt.Errorf("Failed to create local branch '%v'.", branch)
	}
	defer local_branch.Free()

	// second, --set-upstream tracking branch

	// checkout -b would do what we want, but git2go doesn't
	// have an exact equivalent.
	//
	//   git checkout -b fedora/f31 fedora/f31
	//
	// It is easiest to just set the .git/config values.
	//
	// git config "branch.fedora/f31.remote" fedora
	// git config "branch.fedora/f31.merge" "refs/heads/f31"

	branch_parts := strings.Split(branch, "/")
	remote := branch_parts[0]                           // fedora
	short_branch := strings.Join(branch_parts[1:], "/") // f31

	cfg, err := repo.Config()
	if err != nil {
		return fmt.Errorf("Failed to get Config: %v", err)
	}
	err = cfg.SetString(fmt.Sprintf("branch.%s.remote", branch), remote)
	if err != nil {
		return fmt.Errorf("Failed to set config remote: %v", err)
	}
	err = cfg.SetString(fmt.Sprintf("branch.%s.merge", branch), fmt.Sprintf("refs/heads/%s", short_branch))
	if err != nil {
		return fmt.Errorf("Failed to set config merge: %v", err)
	}

	return nil
}

// Setup a local branch corresponding to each remote branch.
//
//  git branch -a
//  ...
//  fedora/31 -> remotes/fedora/f31
//
// This makes sure all the local branches exist and are up to date.
func SetupRpmBranches(repo *git.Repository) error {

	branches, err := getExpectedLocalBranches(repo)
	if err != nil {
		return fmt.Errorf("unable to get branches: %v", err)
	}

	for _, branch := range branches {
		err = setupRpmBranch(repo, branch)
		if err != nil {
			return err
		}
	}

	return nil
}

// Walk all the local branches and perform a git pull.
func PullAll(repo *git.Repository) error {

	err := FetchAll(repo)
	if err != nil {
		return fmt.Errorf("Unable to fetch for pull: %v", err)
	}

	branches, err := getExpectedLocalBranches(repo)
	if err != nil {
		return err
	}

	for _, branch := range branches {

		err = repo.SetHead("refs/heads/" + branch)
		if err != nil {
			return fmt.Errorf("unable to set head to '%s': %v", branch, err)
		}

		err = repo.CheckoutHead(&git.CheckoutOpts{
			Strategy: git.CheckoutForce,
		})
		if err != nil {
			return fmt.Errorf("unable to checkout head: %v", err)
		}

		local_branch, err := repo.LookupBranch(branch, git.BranchLocal)
		if err != nil {
			return fmt.Errorf("unable to lookup branch '%s': %v", branch, err)
		}
		defer local_branch.Free()

		remote_branch, err := repo.LookupBranch(branch, git.BranchRemote)
		if err != nil {
			return fmt.Errorf("unable to lookup remote branch '%s': %v", branch, err)
		}
		defer remote_branch.Free()

		commit, err := repo.AnnotatedCommitFromRef(remote_branch.Reference)
		if err != nil {
			return fmt.Errorf("unable to lookup commit for branch '%v': %v", branch, err)
		}
		defer commit.Free()

		commits := make([]*git.AnnotatedCommit, 1)
		commits[0] = commit

		analysis, _, err := repo.MergeAnalysis(commits)
		if err != nil {
			return fmt.Errorf("unable to run merge analysis: %v", err)
		}

		if (analysis & git.MergeAnalysisUpToDate) != 0 {
			// OK
		} else if (analysis & git.MergeAnalysisFastForward) != 0 {
			local_branch_ref := local_branch.Reference
			_, err := local_branch_ref.SetTarget(remote_branch.Reference.Target(), "pull: Fast-forward")
			if err != nil {
				return fmt.Errorf("Fast-forward merge of '%s' failed: %v", branch, err)
			}

			err = repo.CheckoutHead(&git.CheckoutOpts{
				Strategy: git.CheckoutForce,
			})
			if err != nil {
				return fmt.Errorf("unable to checkout head: %v", err)
			}
		} else if (analysis & git.MergeAnalysisNormal) != 0 {
			return fmt.Errorf("A merge is required for '%s'", branch)
		} else {
			return fmt.Errorf("Unhandled MergeAnalysis? '%v'", analysis)
		}
	}

	return nil
}
