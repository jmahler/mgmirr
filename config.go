package mgmirr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/src-d/go-git.v4/config"
	"io/ioutil"
	"text/template"
)

type Config struct {
	Origin  config.RemoteConfig
	Remotes []config.RemoteConfig
}

// Given a config object (template), fill out the variables.
//   "URLs": ["https://src.fedoraproject.org/rpms/{{.RPM}}.git"]
func ExecConfigTemplate(cfg Config, rpm string) (Config, error) {

	// This copies the given object to a new object while
	// also filling out the template variables.

	var new_cfg Config

	new_cfg.Origin = cfg.Origin

	new_cfg.Remotes = make([]config.RemoteConfig, len(cfg.Remotes))
	for i, remote := range cfg.Remotes {
		new_urls := make([]string, len(remote.URLs))
		for j, URL := range remote.URLs {
			tmpl, err := template.New("URLs").Parse(URL)

			vars := struct{ RPM string }{rpm}
			out := new(bytes.Buffer)
			err = tmpl.Execute(out, vars)
			if err != nil {
				return new_cfg, fmt.Errorf("unable to exec template '%s' for '%s': %v", URL, rpm, err)
			}

			new_urls[j] = string(out.String())
		}
		new_cfg.Remotes[i].Name = remote.Name
		new_cfg.Remotes[i].URLs = new_urls
	}

	return new_cfg, nil // OK
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
