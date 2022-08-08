package copier

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func Copy(to, from interface{}) error {
	b, err := json.Marshal(from)
	if err != nil {
		return errors.Wrap(err, "marshal from data err")
	}

	err = json.Unmarshal(b, to)
	if err != nil {
		return errors.Wrap(err, "unmarshal to data err")
	}

	return nil
}
