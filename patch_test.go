package httptap

import (
"testing"

"github.com/evanphx/json-patch"
)

func TestPatches(t *testing.T) {
	cases := []struct {
		name  string
		in    []byte
		patch []byte
		want  []byte
	}{
		{
			name: "add",
			in:   []byte(`{"root":{}}`),
			patch: []byte(`
			[
				{
					"op":"add",
					"path":"/root/key1",
					"value":"value1"
				}
			]`),
			want: []byte(`{"root":{"key1":"value1"}}`),
		},
	}
	for _, cc := range cases {
		t.Run(cc.name, func(tt *testing.T) {
			patch, err := jsonpatch.DecodePatch(cc.patch)
			if err != nil {
				t.Fatalf("error: %s", err)
			}
			res, err := patch.Apply(cc.in)
			if err != nil {
				t.Fatalf("error: %s", err)
			}
			t.Log(string(res))
		})
	}
}
