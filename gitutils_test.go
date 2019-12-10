package mgmirr_test

import (
	"fmt"
	"github.com/jmahler/mgmirr"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
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

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("git init of empty dir '%s' failed: %v", dir, err)
	}

	abs_testdata, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatalf("source directory template path setup failed: %v", err)
	}

	rpm := "patch"
	cfg := []config.RemoteConfig{
		config.RemoteConfig{
			Name: "fedora",
			URLs: []string{filepath.Join(abs_testdata, fmt.Sprintf("%s.fedora", rpm))},
		},
		config.RemoteConfig{
			Name: "centos",
			URLs: []string{filepath.Join(abs_testdata, fmt.Sprintf("%s.centos", rpm))},
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
}
