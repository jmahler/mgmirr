package main_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	out_bytes, err := exec.Command("rgm", "-h").Output()
	if err != nil {
		t.Fatalf("unable to get help usage: %v", err)
	}
	out := string(out_bytes)

	if !strings.Contains(out, "help") {
		t.Errorf("unexpected help (-h) output")
	}
}
