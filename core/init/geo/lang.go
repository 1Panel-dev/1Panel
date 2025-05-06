package geo

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/cmd"
	fileUtils "github.com/1Panel-dev/1Panel/core/utils/files"
)

func Init() {
	go initLang()
}

func initLang() {
	geoPath := path.Join(global.CONF.Base.InstallDir, "1panel/geo/GeoIP.mmdb")
	isLangExist := fileUtils.Stat("/usr/local/bin/lang/zh.sh")
	isGeoExist := fileUtils.Stat(geoPath)
	if isLangExist && isGeoExist {
		return
	}
	upgradePath := path.Join(global.CONF.Base.InstallDir, "1panel/tmp/upgrade")
	tmpPath, err := loadRestorePath(upgradePath)
	upgradeDir := path.Join(upgradePath, tmpPath, "downloads")
	if err != nil || len(tmpPath) == 0 || !fileUtils.Stat(upgradeDir) {
		if !isLangExist {
			downloadLangFromRemote()
		}
		if !isGeoExist {
			downloadGeoFromRemote(geoPath)
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
	if tmpPath == "no such file" || !fileUtils.Stat(tmpPath) {
		if !isLangExist {
			downloadLangFromRemote()
		}
		if !isGeoExist {
			downloadGeoFromRemote(geoPath)
		}
		return
	}
	if !isLangExist {
		if !fileUtils.Stat(path.Join(tmpPath, "lang")) {
			downloadLangFromRemote()
			return
		}
		std, err := cmd.RunDefaultWithStdoutBashCf("cp -r %s %s", path.Join(tmpPath, "lang"), "/usr/local/bin/")
		if err != nil {
			global.LOG.Errorf("load lang from package failed, std: %s, err: %v", std, err)
			return
		}
		global.LOG.Info("init lang successful")
	}
	if !isGeoExist {
		if !fileUtils.Stat(path.Join(tmpPath, "GeoIP.mmdb")) {
			downloadGeoFromRemote(geoPath)
			return
		}
		std, err := cmd.RunDefaultWithStdoutBashCf("mkdir %s && cp %s %s/", path.Dir(geoPath), path.Join(tmpPath, "GeoIP.mmdb"), path.Dir(geoPath))
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

func downloadLangFromRemote() {
	path := fmt.Sprintf("%s/language/lang.tar.gz", global.CONF.RemoteURL.RepoUrl)
	if err := fileUtils.DownloadFile(path, "/usr/local/bin/lang.tar.gz"); err != nil {
		global.LOG.Errorf("download lang.tar.gz failed, err: %v", err)
		return
	}
	if !fileUtils.Stat("/usr/local/bin/lang.tar.gz") {
		global.LOG.Error("download lang.tar.gz failed, no such file")
		return
	}
	std, err := cmd.RunDefaultWithStdoutBashCf("tar zxvfC %s %s", "/usr/local/bin/lang.tar.gz", "/usr/local/bin/")
	if err != nil {
		fmt.Printf("decompress lang.tar.gz failed, std: %s, err: %v", std, err)
		return
	}
	_ = os.Remove("/usr/local/bin/lang.tar.gz")
	global.LOG.Info("download lang successful")
}
func downloadGeoFromRemote(targetPath string) {
	_ = os.MkdirAll(path.Dir(targetPath), os.ModePerm)
	pathItem := fmt.Sprintf("%s/geo/GeoIP.mmdb", global.CONF.RemoteURL.RepoUrl)
	if err := fileUtils.DownloadFile(pathItem, targetPath); err != nil {
		global.LOG.Errorf("download geo ip failed, err: %v", err)
		return
	}
	global.LOG.Info("download geo ip successful")
}
