package mgmirr

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"log"
	"strings"
)

// For an existing Git repo and an RPM (e.g. cowsay) Setup the remotes.
//
// This is a best effort procedure.  Not all remotes will be available
// (fedora might not have package x).  As long as at least one remote
// works it is a success.
func SetupRpmRemotes(repo *git.Repository, rcs []config.RemoteConfig) error {

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

func setupRpmRemote(repo *git.Repository, cfg *config.RemoteConfig) error {
	_, err := repo.CreateRemote(cfg)
	if err != nil {
		if err == git.ErrRemoteExists {
			// OK
		} else {
			return fmt.Errorf("git add remote for '%v' failed: %v", cfg.Name, err)
		}
	}

	return nil
}

func FetchAll(repo *git.Repository, rcs []config.RemoteConfig) error {
	var one_worked bool = false

	for _, c := range rcs {
		err := repo.Fetch(&git.FetchOptions{
			RemoteName: c.Name,
		})
		if err != nil {
			if err == git.NoErrAlreadyUpToDate {
				// OK
			} else {
				log.Printf("git fetch remote '%v' failed: %v", c.Name, err)
			}
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

// Does the reference name represent a branch?
func isBranch(branch string) bool {
	if strings.Contains(branch, "HEAD") {
		return false
	} else if strings.HasPrefix(branch, "refs/remotes") {
		return true
	} else if strings.HasPrefix(branch, "refs/heads") {
		return true
	}

	return false
}

// For a repo with remote branches the expected local branch name
// is the same but with "remotes/" removed.
//  remotes/fedora/f31 -> fedora/31
// This gets the set of local branches (e.g. fedora/f31) that "should"
// exist based on the remotes that were found.
func getExpectedLocalBranches(repo *git.Repository) ([]string, error) {

	var ref_branches []string
	refs, err := repo.Storer.IterReferences()
	if err != nil {
		return nil, err
	}
	_ = refs.ForEach(func(c *plumbing.Reference) error {
		ref_branch := c.Strings()[0]
		if isBranch(ref_branch) {
			ref_branches = append(ref_branches, ref_branch)
		}
		return nil
	})

	// refs/heads/fedora/f31 -> refs/remotes/fedora/f31
	var branches []string
	for _, ref_branch := range ref_branches {
		prefix := "refs/remotes/"
		if strings.HasPrefix(ref_branch, prefix) {
			branch := strings.TrimPrefix(ref_branch, prefix)
			branches = append(branches, branch)
		}
		// else ignore local (refs/heads) branches,
		//      the're accounted for by the remotes.
	}

	return branches, nil
}

func setupRpmBranch(repo *git.Repository, branch string) error {

	var err error

	// fedora/f31 -> fedora, f31
	branch_parts := strings.Split(branch, "/")
	if len(branch_parts) > 2 {
		return fmt.Errorf("branch '%s' with > 2 parts isn't supported", branch)
	}
	remote := branch_parts[0]       // fedora
	short_branch := branch_parts[1] // f31

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("unable to get worktree: %v", err)
	}

	// First, make sure the branch to create exists (e.g. fedora/f31)
	// If it doesn't, create it with the proper remote.
	// This will update .git/config but git branch won't show it yet.
	//
	// [branch "fedora/f31"]
	// 		remote = fedora
	//		merge = refs/heads/f31
	//
	_, err = repo.Branch(branch) // branch exists?
	if err != nil {
		err := repo.CreateBranch(&config.Branch{
			Name:   branch,
			Remote: remote,
			Merge:  plumbing.NewBranchReferenceName(short_branch),
		})
		if err != nil {
			return fmt.Errorf("create branch '%s' failed: %v", branch, err)
		}
	}

	// Checkout the remote for our desired branch.
	// It will be in the "detached head" state.
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewRemoteReferenceName(remote, short_branch),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout remote branch '%s': %v", branch, err)
	}

	// Checkout the remote in to our new branch and create if necessary.
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: true,
	})
	if err != nil {
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
			Force:  true,
		})
		if err != nil {
			return fmt.Errorf("failed to checkout existing branch '%s': %v", branch, err)
		}
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
			return fmt.Errorf("unable to setup branch '%s': %v", branch, err)
		}
	}

	return nil
}

// Walk all the local branches and perform a git pull.
func PullAll(repo *git.Repository) error {
	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %v", err)
	}

	branches, err := getExpectedLocalBranches(repo)
	if err != nil {
		return err
	}

	for _, branch := range branches {

		// fedora/f31 -> [fedora, f31]
		branch_parts := strings.Split(branch, "/")

		remote := branch_parts[0]
		short_branch := strings.Join(branch_parts[1:], "/")

		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(fmt.Sprintf("%s/%s", remote, short_branch)),
		})
		if err != nil {
			return err
		}

		err = wt.Pull(&git.PullOptions{
			RemoteName:    remote,
			ReferenceName: plumbing.NewBranchReferenceName(short_branch),
			// These correspond to the branch, remote and merge in .git/config
		})
		if err != nil {
			if err == git.NoErrAlreadyUpToDate {
				// OK
			} else {
				return fmt.Errorf("pull failed: %v", err)
			}
		}
	}

	return nil
}
