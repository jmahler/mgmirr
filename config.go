package mgmirr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	Origin  RemoteConfig
	Remotes []RemoteConfig
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
