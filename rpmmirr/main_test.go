package main_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestMain(t *testing.T) {
	out_bytes, err := exec.Command("rpmmirr", "-h").Output()
	if err != nil {
		t.Fatalf("unable to get help usage: %v", err)
	}
	out := string(out_bytes)

	if !strings.Contains(out, "help") {
		t.Errorf("unexpected help (-h) output")
	}
}
