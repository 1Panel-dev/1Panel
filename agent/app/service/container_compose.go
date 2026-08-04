package service

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"gopkg.in/yaml.v3"
)

const composeProjectLabel = "com.docker.compose.project"
const composeConfigLabel = "com.docker.compose.project.config_files"
const composeWorkdirLabel = "com.docker.compose.project.working_dir"
const composeCreatedBy = "createdBy"

func (u *ContainerService) PageCompose(req dto.SearchWithPage) (int64, interface{}, error) {
	var (
		records   []dto.ComposeInfo
		BackDatas []dto.ComposeInfo
	)
	client, err := docker.NewDockerClient()
	if err != nil {
		return 0, nil, err
	}
	defer client.Close()

	options := container.ListOptions{All: true}
	options.Filters = filters.NewArgs()
	options.Filters.Add("label", composeProjectLabel)

	list, err := client.ContainerList(context.Background(), options)
	if err != nil {
		return 0, nil, err
	}

	composeRecords, _ := composeRepo.ListRecord()
	pinnedByName := make(map[string]bool, len(composeRecords))
	composeCreatedByLocal := make([]model.Compose, 0, len(composeRecords))
	for _, record := range composeRecords {
		pinnedByName[record.Name] = record.IsPinned
		if len(record.Path) != 0 {
			composeCreatedByLocal = append(composeCreatedByLocal, record)
		}
	}
	composeLocalMap := make(map[string]dto.ComposeInfo)
	for _, localItem := range composeCreatedByLocal {
		composeItemLocal := dto.ComposeInfo{
			ContainerCount: 0,
			CreatedAt:      localItem.CreatedAt.Format(constant.DateTimeLayout),
			ConfigFile:     localItem.Path,
			Workdir:        strings.TrimSuffix(localItem.Path, "/docker-compose.yml"),
		}
		composeItemLocal.CreatedBy = "1Panel"
		composeItemLocal.Path = localItem.Path
		composeLocalMap[localItem.Name] = composeItemLocal
	}

	composeMap := make(map[string]dto.ComposeInfo)
	for _, container := range list {
		if name, ok := container.Labels[composeProjectLabel]; ok {
			containerItem := dto.ComposeContainer{
				ContainerID: container.ID,
				Name:        container.Names[0][1:],
				State:       container.State,
				CreateTime:  time.Unix(container.Created, 0).Format(constant.DateTimeLayout),
				Ports:       transPortToStr(container.Ports),
			}
			if compose, has := composeMap[name]; has {
				compose.ContainerCount++
				if strings.ToLower(containerItem.State) == "running" {
					compose.RunningCount++
				}
				compose.Containers = append(compose.Containers, containerItem)
				composeMap[name] = compose
			} else {
				config := container.Labels[composeConfigLabel]
				workdir := container.Labels[composeWorkdirLabel]
				composeItem := dto.ComposeInfo{
					ContainerCount: 1,
					CreatedAt:      time.Unix(container.Created, 0).Format(constant.DateTimeLayout),
					ConfigFile:     config,
					Workdir:        workdir,
					Containers:     []dto.ComposeContainer{containerItem},
				}
				if strings.ToLower(containerItem.State) == "running" {
					composeItem.RunningCount = 1
				}
				createdBy, ok := container.Labels[composeCreatedBy]
				if ok {
					composeItem.CreatedBy = createdBy
				}
				if len(config) != 0 && len(workdir) != 0 && strings.Contains(config, workdir) {
					composeItem.Path = config
				} else {
					composeItem.Path = workdir
				}
				for i := 0; i < len(composeCreatedByLocal); i++ {
					if composeCreatedByLocal[i].Name == name {
						composeItem.CreatedBy = "1Panel"
						composeCreatedByLocal = append(composeCreatedByLocal[:i], composeCreatedByLocal[i+1:]...)
						break
					}
				}
				composeMap[name] = composeItem
			}
		}
	}

	mergedMap := make(map[string]dto.ComposeInfo)
	for key, localItem := range composeLocalMap {
		mergedMap[key] = localItem
	}
	for key, item := range composeMap {
		if existingItem, exists := mergedMap[key]; exists {
			if item.ContainerCount > 0 {
				if existingItem.ContainerCount <= 0 {
					mergedMap[key] = item
				}
			}
		} else {
			mergedMap[key] = item
		}
	}

	for key, value := range mergedMap {
		value.Name = key
		value.ComposeFileExists = composeFileExists(value.Workdir, value.ConfigFile)
		value.IsPinned = pinnedByName[key]
		records = append(records, value)
	}
	if len(req.Info) != 0 {
		length, count := len(records), 0
		for count < length {
			if !strings.Contains(records[count].Name, req.Info) {
				records = append(records[:count], records[(count+1):]...)
				length--
			} else {
				count++
			}
		}
	}
	if req.ExcludeAppStore {
		length, count := len(records), 0
		for count < length {
			if records[count].CreatedBy == "Apps" {
				records = append(records[:count], records[(count+1):]...)
				length--
			} else {
				count++
			}
		}
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].IsPinned != records[j].IsPinned {
			return records[i].IsPinned
		}
		return records[i].CreatedAt > records[j].CreatedAt
	})
	total, start, end := len(records), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		BackDatas = make([]dto.ComposeInfo, 0)
	} else {
		if end >= total {
			end = total
		}
		BackDatas = records[start:end]
	}
	listItem := loadEnv(BackDatas)
	return int64(total), listItem, nil
}

func composeFileExists(workdir, configFile string) bool {
	workdir = strings.TrimSpace(workdir)
	configFile = strings.TrimSpace(configFile)
	if configFile == "" {
		return false
	}
	for _, item := range strings.Split(configFile, ",") {
		file := strings.TrimSpace(item)
		if file == "" {
			continue
		}
		if !filepath.IsAbs(file) && workdir != "" {
			file = filepath.Join(workdir, file)
		}
		file = filepath.Clean(file)
		info, err := os.Stat(file)
		if err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}

func (u *ContainerService) TestCompose(req dto.ComposeCreate) (bool, error) {
	if err := validateComposeCreateName(req); err != nil {
		return false, err
	}
	if cmd.CheckIllegal(req.Name, req.DirName, req.Path) {
		return false, buserr.New("ErrCmdIllegal")
	}
	if req.From != "path" {
		if err := checkComposeRecordName(composeCreateDirName(req)); err != nil {
			return false, err
		}
	}
	if err := u.loadPath(&req); err != nil {
		return false, err
	}
	if err := newComposeEnv(req.Path, req.Env); err != nil {
		return false, err
	}
	projectName, err := resolveComposeProjectName(req.Path, req.Name)
	if err != nil {
		return false, err
	}
	if err := checkComposeRecordName(projectName); err != nil {
		return false, err
	}
	return true, nil
}

func (u *ContainerService) CreateCompose(req dto.ComposeCreate) error {
	if err := validateComposeCreateName(req); err != nil {
		return err
	}
	if cmd.CheckIllegal(req.Name, req.DirName, req.Path) {
		return buserr.New("ErrCmdIllegal")
	}
	if req.From != "path" {
		if err := checkComposeRecordName(composeCreateDirName(req)); err != nil {
			return err
		}
	}
	if err := u.loadPath(&req); err != nil {
		return err
	}
	if err := newComposeEnv(req.Path, req.Env); err != nil {
		return err
	}
	projectName, err := resolveComposeProjectName(req.Path, req.Name)
	if err != nil {
		return err
	}
	req.Name = projectName
	if err := checkComposeRecordName(req.Name); err != nil {
		return err
	}
	taskItem, err := task.NewTaskWithOps(req.Name, task.TaskCreate, task.TaskScopeCompose, req.TaskID, 1)
	if err != nil {
		return fmt.Errorf("new task for image build failed, err: %v", err)
	}
	go func() {
		taskItem.AddSubTask(i18n.GetMsgByKey("ComposeCreate"), func(t *task.Task) error {
			err := compose.UpWithTask(req.Path, t, req.ForcePull, req.Name)
			t.LogWithStatus(i18n.GetMsgByKey("ComposeCreate"), err)
			if err != nil {
				_, _ = compose.Down(req.Path, req.Name)
				return err
			}
			recordName := strings.ToLower(req.Name)
			record, _ := composeRepo.GetRecord(repo.WithByName(recordName))
			if record.ID == 0 {
				_ = composeRepo.CreateRecord(&model.Compose{Name: recordName, Path: req.Path})
			} else {
				_ = composeRepo.UpdateRecord(recordName, map[string]interface{}{"path": req.Path})
			}
			return nil
		}, nil)
		_ = taskItem.Execute()
	}()

	return nil
}

func checkComposeRecordName(name string) error {
	composeItem, _ := composeRepo.GetRecord(repo.WithByName(name))
	if composeItem.ID != 0 && len(composeItem.Path) != 0 {
		return buserr.New("ErrRecordExist")
	}
	return nil
}

func validateComposeCreateName(req dto.ComposeCreate) error {
	name := strings.TrimSpace(req.Name)
	if name != "" && !re.GetRegex(re.ComposeNamePattern).MatchString(name) {
		return buserr.New("ErrComposeNameInvalid")
	}
	if req.From != "path" && !re.GetRegex(re.ComposeNamePattern).MatchString(composeCreateDirName(req)) {
		return buserr.New("ErrComposeNameInvalid")
	}
	return nil
}

func composeCreateDirName(req dto.ComposeCreate) string {
	dirName := strings.TrimSpace(req.DirName)
	if dirName == "" {
		// Keep compatibility with callers that used name as both the directory and
		// Compose project name before dirName was introduced.
		return strings.TrimSpace(req.Name)
	}
	return dirName
}

func resolveComposeProjectName(composePath, fallbackName string) (string, error) {
	// Preserve the name resolved by Compose (including a top-level name) so the
	// container label and the local record always use the same project identity.
	parentName := normalizeComposeProjectName(path.Base(path.Dir(primaryComposePath(composePath))))
	fallbackName = strings.TrimSpace(fallbackName)
	stdout, err := runComposeConfig(composePath, "")
	if err == nil {
		projectName, parseErr := loadComposeProjectName(stdout)
		if parseErr != nil {
			return "", parseErr
		}
		if projectName != "" {
			if !re.GetRegex(re.ComposeNamePattern).MatchString(projectName) {
				return "", buserr.New("ErrComposeNameInvalid")
			}
			return projectName, nil
		}
		if parentName != "" {
			return parentName, nil
		}
		if fallbackName != "" {
			if _, fallbackErr := runComposeConfig(composePath, fallbackName); fallbackErr != nil {
				return "", fallbackErr
			}
			return fallbackName, nil
		}
		return "", buserr.New("ErrComposeProjectNameEmpty")
	}
	if !isComposeProjectNameEmptyError(err) {
		return "", err
	}

	resolveErr := err
	if parentName != "" {
		if _, parentErr := runComposeConfig(composePath, parentName); parentErr == nil {
			return parentName, nil
		} else {
			resolveErr = parentErr
		}
	}

	if fallbackName != "" && fallbackName != parentName {
		if _, fallbackErr := runComposeConfig(composePath, fallbackName); fallbackErr == nil {
			return fallbackName, nil
		} else {
			return "", fallbackErr
		}
	}
	if parentName == "" && fallbackName == "" {
		return "", buserr.New("ErrComposeProjectNameEmpty")
	}
	return "", resolveErr
}

func runComposeConfig(composePath, projectName string) ([]byte, error) {
	configCmd := getComposeCmd(composePath, "config", projectName)
	stdout, err := configCmd.Output()
	if err != nil {
		var stderr []byte
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = exitErr.Stderr
		}
		return nil, fmt.Errorf("docker-compose config failed, std: %s, err: %v", mergeComposeOutput(stdout, stderr), err)
	}
	return stdout, nil
}

func mergeComposeOutput(stdout, stderr []byte) string {
	outputs := make([]string, 0, 2)
	if output := strings.TrimSpace(string(stdout)); output != "" {
		outputs = append(outputs, output)
	}
	if output := strings.TrimSpace(string(stderr)); output != "" {
		outputs = append(outputs, output)
	}
	return strings.Join(outputs, "\n")
}

func loadComposeProjectName(config []byte) (string, error) {
	var project struct {
		Name string `yaml:"name"`
	}
	if err := yaml.Unmarshal(config, &project); err != nil {
		return "", buserr.WithDetail("ErrComposeProjectNameParse", err.Error(), err)
	}
	return strings.TrimSpace(project.Name), nil
}

func primaryComposePath(composePath string) string {
	if index := strings.Index(composePath, ","); index >= 0 {
		return composePath[:index]
	}
	return composePath
}

func normalizeComposeProjectName(name string) string {
	name = re.GetRegex(re.ComposeDisallowedCharsPattern).
		ReplaceAllString(strings.ToLower(strings.TrimSpace(name)), "")
	return strings.TrimLeft(name, "_-")
}

func isComposeProjectNameEmptyError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "project name must not be empty") ||
		strings.Contains(message, "project name can't be empty")
}

func (u *ContainerService) ComposeOperation(req dto.ComposeOperation) error {
	if len(req.Path) == 0 && req.Operation == "delete" {
		_ = composeRepo.DeleteRecord(repo.WithByName(req.Name))
		return nil
	}
	if cmd.CheckIllegal(req.Path, req.Operation) {
		return buserr.New("ErrCmdIllegal")
	}
	if req.Operation == "delete" {
		if err := removeContainerForCompose(req.Name, req.Path); err != nil && !req.Force {
			return err
		}
		if req.WithFile {
			for _, item := range strings.Split(req.Path, ",") {
				if len(item) != 0 {
					_ = os.RemoveAll(path.Dir(item))
				}
			}
		}
		_ = composeRepo.DeleteRecord(repo.WithByName(req.Name))
		return nil
	}
	if req.Operation == "up" {
		if stdout, err := compose.Up(req.Path, req.Name); err != nil {
			return fmt.Errorf("docker-compose up failed, std: %s, err: %v", stdout, err)
		}
	} else if req.Operation == "rebuild" {
		if stdout, err := compose.DownAndUp(req.Path, req.Name); err != nil {
			return fmt.Errorf("docker-compose rebuild failed, std: %s, err: %v", stdout, err)
		}
	} else {
		if stdout, err := compose.Operate(req.Path, req.Operation, req.Name); err != nil {
			return fmt.Errorf("docker-compose %s failed, std: %s, err: %v", req.Operation, stdout, err)
		}
	}
	return nil
}

func (u *ContainerService) ComposeUpdate(req dto.ComposeUpdate) error {
	if cmd.CheckIllegal(req.Name, req.Path) {
		return buserr.New("ErrCmdIllegal")
	}
	taskItem, err := task.NewTaskWithOps(req.Name, task.TaskUpdate, task.TaskScopeCompose, req.TaskID, 1)
	if err != nil {
		global.LOG.Errorf("new task for update compose failed, err: %v", err)
		return err
	}
	go func() {
		taskItem.AddSubTask(i18n.GetMsgByKey("TaskUpdate"), func(t *task.Task) error {
			oldFile, err := os.ReadFile(req.DetailPath)
			if err != nil {
				return fmt.Errorf("load file with path %s failed, %v", req.DetailPath, err)
			}
			file, err := os.OpenFile(req.DetailPath, os.O_WRONLY|os.O_TRUNC, 0640)
			if err != nil {
				return err
			}
			defer file.Close()
			write := bufio.NewWriter(file)
			_, _ = write.WriteString(req.Content)
			write.Flush()

			global.LOG.Infof("docker-compose.yml %s has been replaced, now start to docker-compose restart", req.DetailPath)
			if err := newComposeEnv(req.DetailPath, req.Env); err != nil {
				return err
			}

			if err := compose.UpWithTask(req.Path, t, req.ForcePull, req.Name); err != nil {
				global.LOG.Errorf("update failed when handle compose up, err: %s, now try to recreate the old compose file", err)
				if err := recreateCompose(string(oldFile), req.Path, req.Name); err != nil {
					return fmt.Errorf("update failed and recreate old compose file also failed, err: %v", err)
				}
				return fmt.Errorf("update failed when handle compose up, err: %s", err)
			}

			return nil
		}, nil)
		_ = taskItem.Execute()
	}()

	return nil
}

func (u *ContainerService) ComposePin(req dto.ComposePin) error {
	record, _ := composeRepo.GetRecord(repo.WithByName(req.Name))
	if record.ID == 0 {
		if !req.IsPinned {
			return nil
		}
		return composeRepo.CreateRecord(&model.Compose{Name: req.Name, IsPinned: true})
	}
	if !req.IsPinned && len(record.Path) == 0 {
		return composeRepo.DeleteRecord(repo.WithByName(req.Name))
	}
	return composeRepo.UpdateRecord(req.Name, map[string]interface{}{"is_pinned": req.IsPinned})
}

func (u *ContainerService) ComposeLogClean(req dto.ComposeLogClean) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()

	options := container.ListOptions{All: true}
	options.Filters = filters.NewArgs()
	options.Filters.Add("label", composeProjectLabel)

	list, err := client.ContainerList(context.Background(), options)
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, item := range list {
		if name, ok := item.Labels[composeProjectLabel]; ok {
			if name != req.Name {
				continue
			}
			containerItem, err := client.ContainerInspect(ctx, item.ID)
			if err != nil {
				return err
			}
			if err := client.ContainerStop(ctx, containerItem.ID, container.StopOptions{}); err != nil {
				return err
			}
			file, err := os.OpenFile(containerItem.LogPath, os.O_RDWR|os.O_CREATE, constant.FilePerm)
			if err != nil {
				return err
			}
			defer file.Close()
			if err = file.Truncate(0); err != nil {
				return err
			}
			_, _ = file.Seek(0, 0)

			files, _ := filepath.Glob(fmt.Sprintf("%s.*", containerItem.LogPath))
			for _, file := range files {
				_ = os.Remove(file)
			}
		}
	}
	return u.ComposeOperation(dto.ComposeOperation{
		Name:      req.Name,
		Path:      req.Path,
		Operation: "restart",
	})
}

func (u *ContainerService) LoadComposeEnv(name string) (string, error) {
	envFilePath := path.Join(path.Dir(name), ".env")
	file, err := os.ReadFile(envFilePath)
	if err != nil {
		return "", err
	}
	return string(file), nil
}

func (u *ContainerService) loadPath(req *dto.ComposeCreate) error {
	if req.From == "template" || req.From == "edit" {
		dir := fmt.Sprintf("%s/docker/compose/%s", global.Dir.DataDir, composeCreateDirName(*req))
		if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
		}

		path := fmt.Sprintf("%s/docker-compose.yml", dir)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, constant.FilePerm)
		if err != nil {
			return err
		}
		defer file.Close()
		write := bufio.NewWriter(file)
		_, _ = write.WriteString(string(req.File))
		write.Flush()
		req.Path = path
	}
	return nil
}

func removeContainerForCompose(composeName, composePath string) error {
	if _, err := os.Stat(composePath); err == nil {
		if stdout, err := compose.Operate(composePath, "down", composeName); err != nil {
			return errors.New(stdout)
		}
		return nil
	}
	var options container.ListOptions
	options.All = true
	options.Filters = filters.NewArgs()
	options.Filters.Add("label", "com.docker.compose.project="+composeName)
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	ctx := context.Background()
	containers, err := client.ContainerList(ctx, options)
	if err != nil {
		return err
	}
	for _, c := range containers {
		_ = client.ContainerRemove(ctx, c.ID, container.RemoveOptions{RemoveVolumes: true, Force: true})
	}
	return nil
}

func recreateCompose(content, path, projectName string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(content)
	write.Flush()

	if stdout, err := compose.Up(path, projectName); err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

func loadEnv(list []dto.ComposeInfo) []dto.ComposeInfo {
	for i := 0; i < len(list); i++ {
		tmpPath := list[i].Path
		if strings.Contains(list[i].Path, ",") {
			tmpPath = strings.Split(list[i].Path, ",")[0]
		}
		envFilePath := path.Join(path.Dir(tmpPath), ".env")
		file, err := os.ReadFile(envFilePath)
		if err != nil {
			continue
		}
		list[i].Env = string(file)
	}
	return list
}

func newComposeEnv(pathItem string, env string) error {
	envFilePath := path.Join(path.Dir(pathItem), ".env")
	file, err := os.OpenFile(envFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, constant.FilePerm)
	if err != nil {
		global.LOG.Errorf("failed to create env file: %v", err)
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(env); err != nil {
		global.LOG.Errorf("failed to write env to file: %v", err)
		return err
	}
	return nil
}
