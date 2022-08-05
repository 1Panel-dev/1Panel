package copier

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// Copy 从一个结构体复制到另一个结构体
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
