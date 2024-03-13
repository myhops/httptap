package tap

import "gopkg.in/yaml.v3"

type TapCfg struct {
	Name string    `yaml:"name"`
	Kind string    `yaml:"kind"`
	Spec yaml.Node `yaml:"spec"`
}

type TemplateTapConfigDefinition struct {
	Tap  TapCfg         `yaml:"metadata"`
	Spec TemplateObject `yaml:"spec"`
}

