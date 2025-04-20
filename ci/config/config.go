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

func (c Conf) ToConfString() (ConfString, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return ConfString(b), nil
}
