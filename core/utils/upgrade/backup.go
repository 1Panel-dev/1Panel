package upgrade

import (
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/core/utils/common"
)

func DropBackupCopies(installDir, backupCopies string) error {
	copies, _ := strconv.Atoi(backupCopies)
	if copies == 0 {
		return nil
	}
	backupDir := path.Join(installDir, "1panel/tmp/upgrade")
	upgradeDir, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}
	var versions []string
	for _, item := range upgradeDir {
		if item.IsDir() && strings.HasPrefix(item.Name(), "v") {
			versions = append(versions, item.Name())
		}
	}
	if len(versions) <= copies {
		return nil
	}
	sort.Slice(versions, func(i, j int) bool {
		return common.ComparePanelVersion(versions[i], versions[j])
	})
	for i := copies; i < len(versions); i++ {
		_ = os.RemoveAll(path.Join(backupDir, versions[i]))
	}
	return nil
}
