package buserr

import (
	"bytes"
	"fmt"
	"sort"
)

type MultiErr map[string]error

func (e MultiErr) Error() string {
	var keys []string
	for key := range e {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	buffer := bytes.NewBufferString("")
	for _, key := range keys {
		buffer.WriteString(fmt.Sprintf("[%s] %s\n", key, e[key]))
	}
	return buffer.String()
}
