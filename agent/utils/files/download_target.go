package files

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/1Panel-dev/1Panel/agent/buserr"
)

func validDownloadFilename(name string) bool {
	return strings.TrimSpace(name) != "" && name != "." && name != ".." && len(name) <= 255 && utf8.ValidString(name) &&
		!strings.ContainsAny(name, "/\\:") && strings.IndexFunc(name, unicode.IsControl) < 0
}

func ResolveDownloadDestination(directory, name string) (string, error) {
	if strings.TrimSpace(directory) == "" || !validDownloadFilename(name) {
		return "", buserr.New("ErrInvalidParams")
	}
	parent, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return "", err
	}
	parent, err = filepath.Abs(parent)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(parent)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", buserr.New("ErrInvalidParams")
	}
	destination := filepath.Join(parent, name)
	if filepath.Dir(destination) != parent {
		return "", buserr.New("ErrInvalidParams")
	}
	return destination, nil
}
