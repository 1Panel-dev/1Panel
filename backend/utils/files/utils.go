package files

import (
	"github.com/gabriel-vasile/mimetype"
	"os"
	"os/user"
	"strconv"
)

func IsSymlink(mode os.FileMode) bool {
	return mode&os.ModeSymlink != 0
}

func GetUsername(uid uint32) string {
	usr, err := user.LookupId(strconv.Itoa(int(uid)))
	if err != nil {
		return ""
	}
	return usr.Username
}

func GetGroup(gid uint32) string {
	usr, err := user.LookupGroupId(strconv.Itoa(int(gid)))
	if err != nil {
		return ""
	}
	return usr.Name
}

func GetMimeType(path string) string {
	mime, err := mimetype.DetectFile(path)
	if err != nil {
		return ""
	}
	return mime.String()
}

func GetSymlink(path string) string {
	linkPath, err := os.Readlink(path)
	if err != nil {
		return ""
	}
	return linkPath
}
