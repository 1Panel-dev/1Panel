package lang

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

func Init() {
	go initLang()
}

func initLang() {
	fileOp := files.NewFileOp()
	geoPath := path.Join(global.CONF.System.BaseDir, "1panel/geo/GeoIP.mmdb")
	isLangExist := fileOp.Stat("/usr/local/bin/lang/zh.sh")
	isGeoExist := fileOp.Stat(geoPath)
	if isLangExist && isGeoExist {
		return
	}
	upgradePath := path.Join(global.CONF.System.BaseDir, "1panel/tmp/upgrade")
	tmpPath, err := loadRestorePath(upgradePath)
	upgradeDir := path.Join(upgradePath, tmpPath, "downloads")
	if err != nil || len(tmpPath) == 0 || !fileOp.Stat(upgradeDir) {
		if !isLangExist {
			downloadLangFromRemote(fileOp)
		}
		if !isGeoExist {
			downloadGeoFromRemote(fileOp, geoPath)
		}
		return
	}

	files, _ := os.ReadDir(upgradeDir)
	if len(files) == 0 {
		tmpPath = "no such file"
	} else {
		for _, item := range files {
			if item.IsDir() && strings.HasPrefix(item.Name(), "1panel-") {
				tmpPath = path.Join(upgradePath, tmpPath, "downloads", item.Name())
				break
			}
		}
	}
	if tmpPath == "no such file" || !fileOp.Stat(tmpPath) {
		if !isLangExist {
			downloadLangFromRemote(fileOp)
		}
		if !isGeoExist {
			downloadGeoFromRemote(fileOp, geoPath)
		}
		return
	}
	if !isLangExist {
		if !fileOp.Stat(path.Join(tmpPath, "lang")) {
			downloadLangFromRemote(fileOp)
			return
		}
		std, err := cmd.Execf("cp -r %s %s", path.Join(tmpPath, "lang"), "/usr/local/bin/")
		if err != nil {
			global.LOG.Errorf("load lang from package failed, std: %s, err: %v", std, err)
			return
		}
		global.LOG.Info("init lang successful")
	}
	if !isGeoExist {
		if !fileOp.Stat(path.Join(tmpPath, "GeoIP.mmdb")) {
			downloadGeoFromRemote(fileOp, geoPath)
			return
		}
		std, err := cmd.Execf("mkdir %s && cp %s %s/", path.Dir(geoPath), path.Join(tmpPath, "GeoIP.mmdb"), path.Dir(geoPath))
		if err != nil {
			global.LOG.Errorf("load geo ip from package failed, std: %s, err: %v", std, err)
			return
		}
		global.LOG.Info("init geo ip successful")
	}
}

func loadRestorePath(upgradeDir string) (string, error) {
	if _, err := os.Stat(upgradeDir); err != nil && os.IsNotExist(err) {
		return "no such file", nil
	}
	files, err := os.ReadDir(upgradeDir)
	if err != nil {
		return "", err
	}
	var folders []string
	for _, file := range files {
		if file.IsDir() {
			folders = append(folders, file.Name())
		}
	}
	if len(folders) == 0 {
		return "no such file", nil
	}
	sort.Slice(folders, func(i, j int) bool {
		return folders[i] > folders[j]
	})
	return folders[0], nil
}

func downloadLangFromRemote(fileOp files.FileOp) {
	path := fmt.Sprintf("%s/language/lang.tar.gz", global.CONF.System.RepoUrl)
	if err := fileOp.DownloadFile(path, "/usr/local/bin/lang.tar.gz"); err != nil {
		global.LOG.Errorf("download lang.tar.gz failed, err: %v", err)
		return
	}
	if !fileOp.Stat("/usr/local/bin/lang.tar.gz") {
		global.LOG.Error("download lang.tar.gz failed, no such file")
		return
	}
	std, err := cmd.Execf("tar zxvfC %s %s", "/usr/local/bin/lang.tar.gz", "/usr/local/bin/")
	if err != nil {
		fmt.Printf("decompress lang.tar.gz failed, std: %s, err: %v", std, err)
		return
	}
	_ = os.Remove("/usr/local/bin/lang.tar.gz")
	global.LOG.Info("download lang successful")
}
func downloadGeoFromRemote(fileOp files.FileOp, targetPath string) {
	_ = os.MkdirAll(path.Dir(targetPath), os.ModePerm)
	pathItem := fmt.Sprintf("%s/geo/GeoIP.mmdb", global.CONF.System.RepoUrl)
	if err := fileOp.DownloadFile(pathItem, targetPath); err != nil {
		global.LOG.Errorf("download geo ip failed, err: %v", err)
		return
	}
	global.LOG.Info("download geo ip successful")
}
