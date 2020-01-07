package mgmirr

import (
	"encoding/json"
	"fmt"
	"gopkg.in/src-d/go-git.v4/config"
	"io/ioutil"
)

type Config struct {
	Origin  config.RemoteConfig
	Remotes []config.RemoteConfig
}

func LoadConfig(config_file string) (Config, error) {

	var err error
	var cfg Config

	file, err := ioutil.ReadFile(config_file)
	if err != nil {
		return cfg, fmt.Errorf("unable to read '%s': %v", config_file, err)
	}

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("Unmarshal of '%s' failed: %v", config_file, err)
	}

	return cfg, nil
}
