package compose

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
)

func checkCmd() error {
	if global.CONF.DockerConfig.Command == "" {
		dockerComposeCmd := common.GetDockerComposeCommand()
		if dockerComposeCmd == "" {
			return buserr.New("ErrDockerComposeCmdNotFound")
		}
		global.CONF.DockerConfig.Command = dockerComposeCmd
	}
	return nil
}

func getComposeBaseCmd() (string, []string) {
	cmdStr := strings.TrimSpace(global.CONF.DockerConfig.Command)
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

func Up(filePath string, projectName ...string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	base, extra := getComposeBaseCmd()
	args := appendProjectName(append([]string(nil), extra...), projectName)
	args = append(args, upArgs(filePath, false)...)
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdout(base, args...)
}

func UpWithoutBuild(filePath string, projectName ...string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	base, extra := getComposeBaseCmd()
	args := appendProjectName(append([]string(nil), extra...), projectName)
	args = append(args, upArgs(filePath, true)...)
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdout(base, args...)
}

func upArgs(filePath string, withoutBuild bool) []string {
	args := loadFiles(filePath)
	args = append(args, "up", "-d")
	if withoutBuild {
		args = append(args, "--no-build")
	}
	return args
}

func UpWithTask(filePath string, task *task.Task, forcePull bool, projectName ...string) error {
	if err := PullComposeImages(filePath, forcePull, task, projectName...); err != nil {
		return err
	}
	base, extra := getComposeBaseCmd()
	args := appendProjectName(append([]string(nil), extra...), projectName)
	args = append(args, upArgs(filePath, false)...)
	return cmd.NewCommandMgr(cmd.WithTask(*task), cmd.WithTimeout(20*time.Minute)).Run(base, args...)
}

func BuildWithTask(filePath, projectName string, task *task.Task) error {
	if err := checkCmd(); err != nil {
		return err
	}
	base, extra := getComposeBaseCmd()
	args := append([]string(nil), extra...)
	if projectName != "" {
		args = append(args, "--project-name", projectName)
	}
	args = append(args, loadFiles(filePath)...)
	args = append(args, "build")
	return cmd.NewCommandMgr(cmd.WithTask(*task), cmd.WithTimeout(120*time.Minute)).Run(base, args...)
}

func PullComposeImages(filePath string, forcePull bool, task *task.Task, projectName ...string) error {
	return pullComposeImages(filePath, forcePull, task, projectName...)
}

func pullComposeImages(filePath string, forcePull bool, task *task.Task, projectName ...string) error {
	images, err := GetComposeImages(filePath, projectName...)
	if err != nil {
		return err
	}
	dockerCLi, err := docker.NewClient()
	if err != nil {
		return err
	}
	for _, image := range images {
		if !forcePull {
			if exist, _ := dockerCLi.ImageExists(image); exist {
				if task != nil {
					task.Log(i18n.GetMsgByKey("UseExistImage"))
				}
				continue
			}
		}

		if task != nil {
			task.Log(i18n.GetWithName("PullImageStart", image))
		}
		pullErr := error(nil)
		if task != nil {
			pullErr = dockerCLi.PullImageWithProcess(task, image)
		} else {
			pullErr = docker.PullImage(image)
		}
		if pullErr != nil {
			errMsg := ""
			errOur := pullErr.Error()
			if errOur != "" {
				if strings.Contains(errOur, "no such host") {
					errMsg = i18n.GetMsgByKey("ErrNoSuchHost") + ":"
				}
				if strings.Contains(errOur, "Error response from daemon") {
					errMsg = i18n.GetMsgByKey("PullImageTimeout") + ":"
				}
			}
			message := errMsg + errOur
			installErr := errors.New(message)
			if task != nil {
				task.LogFailedWithErr(i18n.GetMsgByKey("PullImage"), installErr)
			}
			if exist, _ := dockerCLi.ImageExists(image); !exist {
				return installErr
			}
			if task != nil {
				task.Log(i18n.GetMsgByKey("UseExistImage"))
			}
		} else if task != nil {
			task.Log(i18n.GetMsgByKey("PullImageSuccess"))
		}
	}

	return nil
}

func GetComposeImages(filePath string, projectName ...string) ([]string, error) {
	images, err := getComposeImagesByCommand(filePath, projectName...)
	if err == nil {
		return images, nil
	}

	content, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return nil, readErr
	}
	env, _ := os.ReadFile(path.Join(path.Dir(filePath), ".env"))
	images, parseErr := docker.GetImagesFromDockerCompose(env, content)
	if parseErr != nil {
		return nil, fmt.Errorf("get compose images failed, cmd err: %v, parse err: %v", err, parseErr)
	}
	return images, nil
}

func getComposeImagesByCommand(filePath string, projectName ...string) ([]string, error) {
	if err := checkCmd(); err != nil {
		return nil, err
	}
	base, extra := getComposeBaseCmd()
	args := appendProjectName(append([]string(nil), extra...), projectName)
	args = append(args, loadFiles(filePath)...)
	args = append(args, "config", "--format", "json", "--no-normalize")
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(5*time.Minute)).
		RunWithStdout(base, args...)
	if err != nil {
		return nil, fmt.Errorf("run compose config --format json --no-normalize failed, std: %s, err: %v", stdout, err)
	}

	var composeConfig struct {
		Services map[string]struct {
			Image string `json:"image"`
		} `json:"services"`
	}
	if err = json.Unmarshal([]byte(stdout), &composeConfig); err != nil {
		return nil, fmt.Errorf("parse compose config json failed, std: %s, err: %v", stdout, err)
	}

	var images []string
	seen := make(map[string]struct{})
	for _, service := range composeConfig.Services {
		image := strings.TrimSpace(service.Image)
		if image == "" {
			continue
		}
		if _, ok := seen[image]; ok {
			continue
		}
		seen[image] = struct{}{}
		images = append(images, image)
	}
	if len(images) == 0 {
		return nil, errors.New("no images found from compose config json")
	}
	return images, nil
}

func Down(filePath string, projectName ...string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	base, extra := getComposeBaseCmd()
	args := appendProjectName(append([]string(nil), extra...), projectName)
	args = append(args, loadFiles(filePath)...)
	args = append(args, "down", "--remove-orphans")
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdout(base, args...)
}

func Stop(filePath string, projectName ...string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	base, extra := getComposeBaseCmd()
	args := appendProjectName(append([]string(nil), extra...), projectName)
	args = append(args, loadFiles(filePath)...)
	args = append(args, "stop")
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdout(base, args...)
}

func Restart(filePath string, projectName ...string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	base, extra := getComposeBaseCmd()
	args := appendProjectName(append([]string(nil), extra...), projectName)
	args = append(args, loadFiles(filePath)...)
	args = append(args, "restart")
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdout(base, args...)
}

func Operate(filePath, operation string, projectName ...string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	base, extra := getComposeBaseCmd()
	args := appendProjectName(append([]string(nil), extra...), projectName)
	args = append(args, loadFiles(filePath)...)
	args = append(args, operation)
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Minute)).RunWithStdout(base, args...)
}

func DownAndUp(filePath string, projectName ...string) (string, error) {
	if err := checkCmd(); err != nil {
		return "", err
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(20 * time.Minute))
	base, extra := getComposeBaseCmd()
	argsDown := appendProjectName(append([]string(nil), extra...), projectName)
	argsDown = append(argsDown, loadFiles(filePath)...)
	argsDown = append(argsDown, "down")
	stdout, err := cmdMgr.RunWithStdout(base, argsDown...)
	if err != nil {
		return stdout, err
	}
	argsUp := appendProjectName(append([]string(nil), extra...), projectName)
	argsUp = append(argsUp, loadFiles(filePath)...)
	argsUp = append(argsUp, "up", "-d")
	stdout, err = cmdMgr.RunWithStdout(base, argsUp...)
	return stdout, err
}

func appendProjectName(args []string, projectName []string) []string {
	if len(projectName) > 0 && strings.TrimSpace(projectName[0]) != "" {
		args = append(args, "--project-name", projectName[0])
	}
	return args
}

func loadFiles(filePath string) []string {
	var fileItem []string
	for _, item := range strings.Split(filePath, ",") {
		if len(item) != 0 {
			fileItem = append(fileItem, "-f", item)
		}
	}
	return fileItem
}
