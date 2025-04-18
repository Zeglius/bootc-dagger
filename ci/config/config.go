package config

import (
	"encoding/json"
)

// ConfString holds a json of [Conf].
type ConfString = string

func ReadConfString(s ConfString) (Conf, error) {
	var c Conf
	if err := json.Unmarshal([]byte(s), &c); err != nil {
		return Conf{}, err
	}
	return c, nil
}

type Conf struct {
	Jobs []Job `json:"jobs"` // Wow
}

// // Returns the json schema that the config file follows.
// func (m *Ci) ConfigJsonSchema() string {
// 	r := &jsonschema.Reflector{
// 		KeyNamer:       strcase.KebabCase,
// 		ExpandedStruct: true,
// 	}

// 	json, _ := r.Reflect(m.Conf).MarshalJSON()
// 	return string(json)
// }
