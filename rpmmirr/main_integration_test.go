// +build integration

package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func TestRpmMirror(t *testing.T) {
	path, err := ioutil.TempDir("", "rpmmirr-main_integration_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(path)

	out_bytes, err := exec.Command("rpmmirr", "-C", path, "-c", "testdata/integration_config.json", "-r", "patch").Output()
	if err != nil {
		out := string(out_bytes)
		fmt.Println(out)
		t.Fatalf("unable to mirror RPM: %v", err)
	}
}
