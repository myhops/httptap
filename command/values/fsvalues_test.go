package values

import (
	"flag"
	"testing"

	"github.com/myhops/httptap/config"
)

const testConfigFile = "/home/peza/DevProjects/httptap/config/testcase2.yaml"

func TestParseTapHandler(t *testing.T) {
	th := &config.TapHandler{}
	fs := NewFlagSet("test", flag.ContinueOnError)

	fs.TapHandlerVar(&th, "tap-config-file", th, "bla")
	err := fs.Parse([]string{"-tap-config-file", testConfigFile})
	if err != nil {
		t.Fatalf("error: %s", err.Error())
	}
}
