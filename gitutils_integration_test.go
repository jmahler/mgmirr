// +build integration

// Perform integration tests by pulling from actual RPM repos.
//
//   go test -tags=integration
//
// **NOTE** These tests are very slow so don't waste your
// time running them until after the local tests have passed.
package mgmirr_test

import (
	"github.com/jmahler/mgmirr"
	"gopkg.in/libgit2/git2go.v27"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestRpmMirrorIntegration(t *testing.T) {
	dir, err := ioutil.TempDir("", "mgmirr-integration")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)
	//fmt.Println(dir)
	// Need to debug tests?  Comment out Remove and Print the Git repo dir.

	cfg_tmpl, err := mgmirr.LoadConfig("testdata/integration_config.json")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	rpm := "patch"
	cfg, err := mgmirr.ExecConfigTemplate(cfg_tmpl, rpm)
	if err != nil {
		t.Fatal(err)
	}

	repo, err := git.Clone(cfg.Origin.URL, dir, &git.CloneOptions{Bare: false})
	if err != nil {
		t.Fatalf("git clone of '%s' to '%s' failed: %v", cfg.Origin.URL, dir, err)
	}

	t.Run("SetupRpmRemotes", func(t *testing.T) {
		err = mgmirr.SetupRpmRemotes(repo, cfg.Remotes)
		if err != nil {
			t.Fatalf("setup remotes failed: %v", err)
		}

		out_bytes, err := exec.Command("git", "-C", dir, "remote").Output()
		if err != nil {
			t.Fatalf("unable to get remote: %v", err)
		}
		out := string(out_bytes)

		for _, c := range cfg.Remotes {
			remote := c.Name
			if !strings.Contains(out, remote) {
				t.Errorf("remote '%s' not found", remote)
			}
		}
	})

	t.Run("FetchAll", func(t *testing.T) {
		err = mgmirr.FetchAll(repo)
		if err != nil {
			t.Fatalf("FetchAll failed: %v", err)
		}

		cases := []BranchCase{
			{"remotes/fedora/f29", true},
			{"remotes/fedora/f31", true},
			{"remotes/fedora/f2", false},
			{"remotes/fedora/f3", false},
			{"remotes/centos/c6", true},
			{"remotes/centos/c7", true},
		}
		testBranches(t, dir, cases)
	})

	t.Run("SetupRpmBranches", func(t *testing.T) {
		err = mgmirr.SetupRpmBranches(repo)
		if err != nil {
			t.Fatalf("SetupRpmBranches failed: %v", err)
		}

		cases := []BranchCase{
			{"fedora/f29", true},
			{"fedora/f31", true},
			{"tes/fedora/f31", false},
			{"fedora/f2", false},
			{"fedora/f3", false},
			{"centos/c6", true},
			{"centos/c7", true},
		}
		testBranches(t, dir, cases)

		testTrackingBranch(t, dir, "fedora/f31", "remotes/fedora/f31")
	})

	t.Run("PullAll", func(t *testing.T) {

		// before, all up to date
		cases := []BranchStatusCase{
			{"fedora/f29", true},
			{"fedora/f30", true},
			{"fedora/f31", true},
			{"centos/c6", true},
			{"centos/c7", true},
		}
		testBranchStatus(t, dir, cases)

		branches := []string{
			"fedora/f29",
			"fedora/f31",
			"centos/c7",
		}
		resetBranches(t, dir, branches)

		// now some are out of date
		cases = []BranchStatusCase{
			{"fedora/f29", false},
			{"fedora/f30", true},
			{"fedora/f31", false},
			{"centos/c6", true},
			{"centos/c7", false},
		}
		testBranchStatus(t, dir, cases)

		err = mgmirr.PullAll(repo)
		if err != nil {
			t.Error(err)
		}

		// now all should be up to date
		cases = []BranchStatusCase{
			{"fedora/f29", true},
			{"fedora/f30", true},
			{"fedora/f31", true},
			{"centos/c6", true},
			{"centos/c7", true},
		}
		testBranchStatus(t, dir, cases)
	})
}
