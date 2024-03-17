package yamljson

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

func Y2J(b []byte) ([]byte, error) {
	var obj any
	if err := yaml.Unmarshal(b, &obj); err != nil {
		return nil, err
	}
	res, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func J2Y(b []byte) ([]byte, error) {
	var obj any
	if err := json.Unmarshal(b, &obj); err != nil {
		return nil, err
	}
	res, err := yaml.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return res, nil
}
