package mgmirr_test

import (
	"github.com/jmahler/mgmirr"
	"testing"
)

func TestConfig(t *testing.T) {
	_, err := mgmirr.LoadConfig("testdata/config.json")
	if err != nil {
		t.Errorf("Failed to load config.json: %v", err)
	}
}
