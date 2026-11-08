package files

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
)

type fileDeletePolicy struct {
	dataDir, trashRoot string
	fixedRoots         []string
}
type deletePathViews struct {
	lexical, physical string
	info              os.FileInfo
}

func CheckFileDeletePaths(paths ...string) error {
	return (fileDeletePolicy{dataDir: global.Dir.DataDir, trashRoot: global.Dir.RecycleBinDir, fixedRoots: protectedPaths}).check(paths...)
}

func (policy fileDeletePolicy) check(paths ...string) error {
	var fixed []deletePathViews
	for _, root := range policy.fixedRoots {
		view, err := newDeletePathViews(root)
		if err != nil {
			return buserr.New("ErrPathNotDelete")
		}
		fixed = append(fixed, view)
	}
	var data, trash *deletePathViews
	for _, setting := range []struct {
		path        string
		destination **deletePathViews
	}{{policy.dataDir, &data}, {policy.trashRoot, &trash}} {
		if strings.TrimSpace(setting.path) == "" {
			continue
		}
		view, err := newDeletePathViews(setting.path)
		if err != nil {
			return buserr.New("ErrPathNotDelete")
		}
		*setting.destination = &view
	}
	for _, path := range paths {
		target, err := newDeletePathViews(path)
		if err != nil {
			return buserr.New("ErrPathNotDelete")
		}
		for _, root := range fixed {
			if sameDeleteRoot(target, root) {
				return buserr.New("ErrPathNotDelete")
			}
		}
		if data != nil && deleteTargetContainsRoot(target, *data) {
			return buserr.New("ErrPathNotDelete")
		}
		if strings.EqualFold(filepath.Base(target.lexical), ".1panel_clash") || strings.EqualFold(filepath.Base(target.physical), ".1panel_clash") || trash != nil && sameDeleteRoot(target, *trash) {
			return buserr.New("ErrPathNotDelete")
		}
	}
	return nil
}

func newDeletePathViews(value string) (deletePathViews, error) {
	if strings.TrimSpace(value) == "" || strings.IndexByte(value, 0) >= 0 {
		return deletePathViews{}, syscall.EINVAL
	}
	absolute, err := filepath.Abs(value)
	if err != nil {
		return deletePathViews{}, err
	}
	physicalInput := value
	if !filepath.IsAbs(physicalInput) {
		cwd, err := os.Getwd()
		if err != nil {
			return deletePathViews{}, err
		}
		physicalInput = cwd + string(filepath.Separator) + physicalInput
	}
	physical, err := filepath.EvalSymlinks(physicalInput)
	if os.IsNotExist(err) {
		physical, err = resolveMissingDeletePath(value)
	}
	if err != nil {
		return deletePathViews{}, err
	}
	physical, err = filepath.Abs(physical)
	if err != nil {
		return deletePathViews{}, err
	}
	info, err := os.Stat(physical)
	if err != nil && !os.IsNotExist(err) {
		return deletePathViews{}, err
	}
	return deletePathViews{lexical: absolute, physical: physical, info: info}, nil
}

func sameDeleteRoot(target, root deletePathViews) bool {
	for _, candidate := range []string{target.lexical, target.physical} {
		for _, protected := range []string{root.lexical, root.physical} {
			if candidate == protected {
				return true
			}
		}
	}
	return sameDeleteDirectory(target.info, root.info)
}
func sameDeleteDirectory(a, b os.FileInfo) bool {
	return a != nil && b != nil && a.IsDir() && b.IsDir() && os.SameFile(a, b)
}

func deleteTargetContainsRoot(target, root deletePathViews) bool {
	if sameDeleteRoot(target, root) {
		return true
	}
	for _, candidate := range []string{target.lexical, target.physical} {
		for _, protected := range []string{root.lexical, root.physical} {
			relative, err := filepath.Rel(candidate, protected)
			if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				return true
			}
		}
	}
	if target.info != nil && target.info.IsDir() {
		for ancestor := root.physical; ; ancestor = filepath.Dir(ancestor) {
			info, err := os.Stat(ancestor)
			if err == nil && sameDeleteDirectory(target.info, info) {
				return true
			}
			if filepath.Dir(ancestor) == ancestor {
				break
			}
		}
	}
	return false
}

func resolveMissingDeletePath(value string) (string, error) {
	if !filepath.IsAbs(value) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		value = cwd + string(filepath.Separator) + value
	}
	pending := strings.Split(filepath.ToSlash(value), "/")
	current := string(filepath.Separator)
	var currentInfo os.FileInfo
	missing, links := false, 0
	for len(pending) > 0 {
		component := pending[0]
		pending = pending[1:]
		if component == "" {
			continue
		}
		if currentInfo != nil && !currentInfo.IsDir() {
			return "", syscall.ENOTDIR
		}
		if component == "." {
			continue
		}
		if component == ".." {
			if missing {
				return "", os.ErrNotExist
			}
			current = filepath.Dir(current)
			currentInfo = nil
			continue
		}
		next := filepath.Join(current, component)
		if missing {
			current = next
			continue
		}
		info, err := os.Lstat(next)
		if os.IsNotExist(err) {
			missing = true
			current = next
			currentInfo = nil
			continue
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			links++
			if links > 40 {
				return "", syscall.ELOOP
			}
			link, err := os.Readlink(next)
			if err != nil {
				return "", err
			}
			if filepath.IsAbs(link) {
				current = string(filepath.Separator)
			}
			currentInfo = nil
			pending = append(strings.Split(filepath.ToSlash(link), "/"), pending...)
			continue
		}
		current = next
		currentInfo = info
	}
	return current, nil
}
