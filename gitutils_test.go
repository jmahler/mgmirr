package mgmirr_test

import (
	"fmt"
	"github.com/jmahler/mgmirr"
	"gopkg.in/src-d/go-git.v4"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestRpmMirror(t *testing.T) {

	dir, err := ioutil.TempDir("", "mgmirr")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)
	//fmt.Println(dir)
	// Need to debug tests?  Comment out Remove and Print the Git repo dir.

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("git init of empty dir '%s' failed: %v", dir, err)
	}

	cfg_tmpl, err := mgmirr.LoadConfig("testdata/config.json")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	rpm := "patch"
	cfg, err := mgmirr.ExecConfigTemplate(cfg_tmpl, rpm)
	if err != nil {
		t.Fatal(err)
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
		err = mgmirr.FetchAll(repo, cfg.Remotes)
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
			{"remotes/other/my/branch/with/lots/of/parts", true},
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
			{"other/my/branch/with/lots/of/parts", true},
		}
		testBranches(t, dir, cases)
	})

	t.Run("PullAll", func(t *testing.T) {

		// before, all up to date
		cases := []BranchStatusCase{
			{"fedora/f29", true},
			{"fedora/f30", true},
			{"fedora/f31", true},
			{"centos/c6", true},
			{"centos/c7", true},
			{"other/my/branch/with/lots/of/parts", true},
		}
		testBranchStatus(t, dir, cases)

		branches := []string{
			"fedora/f29",
			"fedora/f31",
			"centos/c7",
			"other/my/branch/with/lots/of/parts",
		}
		resetBranches(t, dir, branches)

		// now some are out of date
		cases = []BranchStatusCase{
			{"fedora/f29", false},
			{"fedora/f30", true},
			{"fedora/f31", false},
			{"centos/c6", true},
			{"centos/c7", false},
			{"other/my/branch/with/lots/of/parts", false},
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
			{"other/my/branch/with/lots/of/parts", true},
		}
		testBranchStatus(t, dir, cases)
	})
}

// Split branch output in to lines and trim whitespace.
//
//    git branch -a
//      origin/f29
//    * origin/f30
//      remotes/fedora/f29
//
//    ["origin/f29", "origin/f30", "remotes/fedora/f29"]
//
func splitBranchOutput(in string) []string {
	lines := strings.Split(in, "\n")
	branches := make([]string, len(lines))
	for i, line := range lines {
		branches[i] = strings.TrimSpace(strings.TrimPrefix(line, "*"))
	}

	return branches
}

type BranchCase struct {
	Branch string
	Exists bool
}

func testBranches(t *testing.T, dir string, cases []BranchCase) {
	t.Helper()

	out_byte, err := exec.Command("git", "-C", dir, "branch", "-a").Output()
	if err != nil {
		t.Fatalf("unable to run git branch -a on '%s': %v", dir, err)
	}
	branches := splitBranchOutput(string(out_byte))

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s", tc.Branch), func(t *testing.T) {
			found := false
			for _, branch := range branches {
				if tc.Branch == branch {
					found = true
					break
				}
			}
			if !found && tc.Exists {
				t.Errorf("didn't find branch '%s'", tc.Branch)
			} else if found && !tc.Exists {
				t.Errorf("found unexpected branch '%s'", tc.Branch)
			}
		})
	}
}

// Reset branches back to older versions.
func resetBranches(t *testing.T, dir string, branches []string) {
	t.Helper()

	for _, branch := range branches {
		_, err := exec.Command("git", "-C", dir, "checkout", branch).Output()
		if err != nil {
			t.Fatalf("unable to checkout branch '%s': %v", branch, err)
		}

		_, err = exec.Command("git", "-C", dir, "reset", "--hard", "HEAD~1").Output()
		if err != nil {
			t.Fatalf("unable to reset branch '%s': %v", branch, err)
		}
	}
}

type BranchStatusCase struct {
	Branch   string
	UpToDate bool
}

func testBranchStatus(t *testing.T, dir string, cases []BranchStatusCase) {
	t.Helper()

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s", c.Branch), func(t *testing.T) {
			_, err := exec.Command("git", "-C", dir, "checkout", c.Branch).Output()
			if err != nil {
				t.Fatalf("unable to checkout branch '%s': %v", c.Branch, err)
			}

			out_bytes, err := exec.Command("git", "-C", dir, "status", c.Branch).Output()
			if err != nil {
				t.Fatalf("unable to get status of branch '%s': %v", c.Branch, err)
			}
			out := string(out_bytes)

			if strings.Contains(out, "Your branch is up to date with") != c.UpToDate {
				t.Errorf("branch '%s' has the wrong status", c.Branch)
			}
		})
	}
}
