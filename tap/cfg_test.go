package tap

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestUnmarshalSpec(t *testing.T) {
	y := `
name: peter
kind: TemplateTap
spec:
  text: some text
  group: some group
  format: json
`
	{
		var c TapCfg
		err := yaml.Unmarshal([]byte(y), &c)
		if err != nil {
			t.Errorf("error: %s", err.Error())
		}
		if c.Kind != "TemplateTap" {
			t.Errorf("unexpected Kind: %s", c.Kind)
		}
		var tc TemplateTapConfig
		err = c.Spec.Decode(&tc)
		if err != nil {
			t.Errorf("error decoding spec: %s", err.Error())
		}
	}
}
