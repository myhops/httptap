package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type TapHandler struct {
	ListenAddress string               `yaml:"listenAddress"`
	Logging       *Logger              `yaml:"logging"`
	Upstream      string               `yaml:"upstream"`
	Header        HeaderIncludeExclude `yaml:"header"`

	Taps []*Tap `yaml:"taps"`
}

type Operation struct {
	Op    string `json:"op" yaml:"op"`
	Path  string `json:"path" yaml:"path"`
	Value string `json:"value,omitempty" yaml:"value"`
}

type HeaderIncludeExclude struct {
	Exclude []string `yaml:"exclude"`
	Include []string `yaml:"include"`
}

type Logger struct {
	LogLevel string `yaml:"logLevel"`
	LogFile  string `yaml:"logFile"`
}

type Body struct {
	Body      bool        `yaml:"body"`
	BodyJSON  bool        `yaml:"bodyJSON"`
	BodyPatch []Operation `yaml:"bodyPatch"`
}

type Tap struct {
	Name     string               `yaml:"name"`
	Patterns []string             `yaml:"patterns"`
	Header   HeaderIncludeExclude `yaml:"header"`

	// Add more taps when they become available.
	LogTap      *LogTap      `yaml:"logTap,omitempty"`
	TemplateTap *TemplateTap `yaml:"templateTap,omitempty"`

	RequestIn  *Body `yaml:"requestIn,omitempty"`
	RequestOut *Body `yaml:"requestOut,omitempty"`
	Response   *Body `yaml:"response,omitempty"`
}

type LogTap struct {
	LogFile string `yaml:"logFile"`
}

type TemplateTap struct {
	Template string `yaml:"template"`
	LogFile  string `yaml:"logFile"`
}

func LoadTapHandler(name string) (*TapHandler, error) {
	blob, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("error reading taphandler config file: %w", err)
	}

	var obj = &TapHandler{}
	if err := yaml.Unmarshal(blob, obj); err != nil {
		return nil, fmt.Errorf("error unmarshalling taphandler config file: %w", err)
	}
	return obj, nil
}
