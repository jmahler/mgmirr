package mgmirr_test

import (
	"fmt"
	"github.com/jmahler/mgmirr"
	"gopkg.in/libgit2/git2go.v27"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

	repo, err := git.InitRepository(dir, false)
	if err != nil {
		t.Fatalf("git init of empty dir '%s' failed: %v", dir, err)
	}

	abs_testdata, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatalf("source directory template path setup failed: %v", err)
	}

	rpm := "patch"
	cfg := []mgmirr.RemoteConfig{
		mgmirr.RemoteConfig{
			Name: "fedora",
			URL:  filepath.Join(abs_testdata, fmt.Sprintf("%s.fedora", rpm)),
		},
		mgmirr.RemoteConfig{
			Name: "centos",
			URL:  filepath.Join(abs_testdata, fmt.Sprintf("%s.centos", rpm)),
		},
	}

	t.Run("SetupRpmRemotes", func(t *testing.T) {
		err = mgmirr.SetupRpmRemotes(repo, cfg)
		if err != nil {
			t.Fatalf("setup remotes failed: %v", err)
		}

		out_bytes, err := exec.Command("git", "-C", dir, "remote").Output()
		if err != nil {
			t.Fatalf("unable to get remote: %v", err)
		}
		out := string(out_bytes)

		for _, c := range cfg {
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

func testTrackingBranch(t *testing.T, dir string, branch string, tracking_branch string) {
	_, err := exec.Command("git", "-C", dir, "checkout", branch).Output()
	if err != nil {
		t.Fatalf("unable to git -C '%s' checkout '%s': %v", dir, branch, err)
	}

	out_byte, err := exec.Command("git", "-C", dir, "checkout").Output()
	if err != nil {
		t.Fatalf("unable to run git -C '%s' checkout: %v", dir, err)
	}
	out := string(out_byte)

	if !strings.Contains(out, tracking_branch) {
		t.Errorf("incorrect tracking branch, expected '%s' in '%s'", tracking_branch, out)
	}
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
