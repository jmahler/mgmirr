package mgmirr_test

import (
	"github.com/jmahler/mgmirr"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	cfg_tmpl, err := mgmirr.LoadConfig("testdata/config.json")
	if err != nil {
		t.Errorf("Failed to load config.json: %v", err)
	}

	// Fill out some data and make sure the template
	// doesn't get corrupted.
	bad_rpm := "badrpmXXX"
	cfg, err := mgmirr.ExecConfigTemplate(cfg_tmpl, bad_rpm)
	if err != nil {
		t.Fatal(err)
	}
	any_url := cfg_tmpl.Remotes[0].URLs[0]
	if strings.Contains(any_url, bad_rpm) {
		t.Fatalf("ExecConfigTemplate corrupted the template: '%s'", any_url)
	}

	rpm := "patch"
	cfg, err = mgmirr.ExecConfigTemplate(cfg_tmpl, rpm)
	if err != nil {
		t.Fatal(err)
	}
	url := cfg.Remotes[0].URLs[0]
	if !strings.Contains(url, rpm) {
		t.Fatalf("Filled out template '%s' missing rpm '%s'", url, rpm)
	}
}
