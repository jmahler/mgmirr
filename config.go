package rpmmirr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"text/template"
)

type Config struct {
	Origin  RemoteConfig
	Remotes []RemoteConfig
}

// Given a config object (template), fill out the variables.
//   "URL": ["https://src.fedoraproject.org/rpms/{{.RPM}}.git"]
func ExecConfigTemplate(cfg Config, rpm string) (Config, error) {

	// This copies the given object to a new object while
	// also filling out the template variables.

	var new_cfg Config

	new_cfg.Origin = RemoteConfig{
		Name: cfg.Origin.Name,
		URL:  cfg.Origin.URL,
	}

	tmpl, err := template.New("URL").Parse(cfg.Origin.URL)

	vars := struct{ RPM string }{rpm}
	out := new(bytes.Buffer)
	err = tmpl.Execute(out, vars)
	if err != nil {
		return new_cfg, fmt.Errorf("unable to exec template '%s' for '%s': %v", cfg.Origin.URL, rpm, err)
	}

	new_cfg.Origin.URL = string(out.String())

	new_cfg.Remotes = make([]RemoteConfig, len(cfg.Remotes))
	for i, remote := range cfg.Remotes {
		tmpl, err := template.New("URL").Parse(remote.URL)

		vars := struct{ RPM string }{rpm}
		out := new(bytes.Buffer)
		err = tmpl.Execute(out, vars)
		if err != nil {
			return new_cfg, fmt.Errorf("unable to exec template '%s' for '%s': %v", remote.URL, rpm, err)
		}
		new_url := string(out.String())

		new_cfg.Remotes[i].Name = remote.Name
		new_cfg.Remotes[i].URL = new_url
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
