package pkg

import (
	"github.com/ghodss/yaml"
	neat "github.com/nikhilsbhat/kubectl-neat/cmd"
	log "github.com/sirupsen/logrus"
)

func (drift *Drift) neat(content []byte) (string, error) {
	if drift.New {
		manifest, err := NewHelmTemplate(string(content)).DropStandardLabels(drift.log)
		if err != nil {
			log.Errorf("dropping labels from template errored with '%v'", err)
		}
		content = []byte(manifest)
	}

	inputBytes, err := yaml.YAMLToJSON(content)
	if err != nil {
		drift.log.Errorf("error converting from yaml to json : %v", err)
		return "", err
	}

	neatFile, err := neat.Neat(string(inputBytes))
	if err != nil {
		drift.log.Errorf("errored while running neat on the manifests : %v", err)
		return "", err
	}

	out, err := yaml.JSONToYAML([]byte(neatFile))
	if err != nil {
		drift.log.Errorf("error converting from json to yaml : %v", err)
		return "", err
	}

	return string(out), nil
}
