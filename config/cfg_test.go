package config

import (
	"testing"
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed testcase1.yaml
var testcase1 []byte

func TestUnmarshalSpec(t *testing.T) {
	{
		var obj TapHandler
		if err := yaml.Unmarshal(testcase1, &obj); err != nil {
			t.Errorf("error: %s", err)
		}
		t.Log("done")
	}
}
