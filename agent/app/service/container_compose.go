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
	"gorm.io/gorm"
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
			Workdir:        filepath.Dir(primaryComposePath(localItem.Path)),
		}
		composeItemLocal.CreatedBy = "1Panel"
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
		composeFiles, err := resolveComposeFiles(value.Workdir, value.ConfigFile)
		value.Path = strings.Join(composeFiles, ",")
		value.ComposeFileExists = err == nil && composeFileExists(composeFiles)
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

func (u *ContainerService) TestCompose(req dto.ComposeCreate) (bool, error) {
	if err := validateComposeCreateName(req); err != nil {
		return false, err
	}
	if hasIllegalComposeCreateInput(req) {
		return false, buserr.New("ErrCmdIllegal")
	}
	projectName, err := resolveComposeCreateProjectName(req)
	if err != nil {
		return false, err
	}
	if err := checkComposeCreateDuplicate(req, projectName); err != nil {
		return false, err
	}
	return true, nil
}

func (u *ContainerService) CreateCompose(req dto.ComposeCreate) error {
	if err := validateComposeCreateName(req); err != nil {
		return err
	}
	if hasIllegalComposeCreateInput(req) {
		return buserr.New("ErrCmdIllegal")
	}
	projectName, err := resolveComposeCreateProjectName(req)
	if err != nil {
		return err
	}
	if err := checkComposeCreateDuplicate(req, projectName); err != nil {
		return err
	}
	if err := u.loadPath(&req); err != nil {
		return err
	}
	if err := newComposeEnv(req.Path, req.Env); err != nil {
		return err
	}
	req.Name = projectName
	recordName := strings.ToLower(req.Name)
	if err := saveComposeRecord(recordName, req.Path); err != nil {
		return fmt.Errorf("save compose record failed, err: %v", err)
	}
	taskItem, err := task.NewTaskWithOps(req.Name, task.TaskCreate, task.TaskScopeCompose, req.TaskID, 1)
	if err != nil {
		return fmt.Errorf("new task for image build failed, err: %v", err)
	}
	go func() {
		taskItem.AddSubTask(i18n.GetMsgByKey("ComposeCreate"), func(t *task.Task) error {
			err := compose.UpWithTask(req.Path, t, req.ForcePull, req.Name)
			t.LogWithStatus(i18n.GetMsgByKey("ComposeCreate"), err)
			return err
		}, nil)
		_ = taskItem.Execute()
	}()

	return nil
}

func saveComposeRecord(name, composePath string) error {
	record, err := composeRepo.GetRecord(repo.WithByName(name))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if record.ID == 0 {
		return composeRepo.CreateRecord(&model.Compose{Name: name, Path: composePath})
	}
	return composeRepo.UpdateRecord(name, map[string]interface{}{"path": composePath})
}

func checkComposeRecordName(name string) error {
	composeItem, _ := composeRepo.GetRecord(repo.WithByName(name))
	if composeItem.ID != 0 && len(composeItem.Path) != 0 {
		return buserr.New("ErrRecordExist")
	}
	return nil
}

func checkComposeCreateDuplicate(req dto.ComposeCreate, projectName string) error {
	if err := checkComposeRecordName(projectName); err != nil {
		return err
	}
	if req.From == "path" {
		return nil
	}
	composeItem, _ := composeRepo.GetRecord(repo.WithByPath(composeCreatePath(req)))
	if composeItem.ID != 0 && composeItem.Path != "" {
		return buserr.New("ErrRecordExist")
	}
	return nil
}

func validateComposeCreateName(req dto.ComposeCreate) error {
	if req.From == "path" {
		name := strings.TrimSpace(req.Name)
		if name != "" && !re.GetRegex(re.ComposeNamePattern).MatchString(name) {
			return buserr.New("ErrComposeNameInvalid")
		}
		return nil
	}
	if !re.GetRegex(re.ComposeNamePattern).MatchString(composeCreateDirName(req)) {
		return buserr.New("ErrComposeNameInvalid")
	}
	return nil
}

func hasIllegalComposeCreateInput(req dto.ComposeCreate) bool {
	if req.From == "path" {
		return cmd.CheckIllegal(req.Name, req.Path)
	}
	return cmd.CheckIllegal(composeCreateDirName(req))
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

func composeCreatePath(req dto.ComposeCreate) string {
	return filepath.Join(global.Dir.DataDir, "docker", "compose", composeCreateDirName(req), "docker-compose.yml")
}

func resolveComposeCreateProjectName(req dto.ComposeCreate) (string, error) {
	if req.From == "path" {
		envPath, err := createComposeTempFile(
			filepath.Dir(primaryComposePath(req.Path)),
			".1panel-compose-*.env",
			req.Env,
		)
		if err != nil {
			return "", err
		}
		defer os.Remove(envPath)
		return resolveComposeProjectName(req.Path, req.Name, envPath)
	}

	dir := filepath.Dir(composeCreatePath(req))
	cleanupDir, err := prepareComposeStagingDir(dir)
	if err != nil {
		return "", err
	}
	defer cleanupDir()

	composePath, err := createComposeTempFile(dir, ".1panel-compose-*.yml", req.File)
	if err != nil {
		return "", err
	}
	defer os.Remove(composePath)

	envPath, err := createComposeTempFile(dir, ".1panel-compose-*.env", req.Env)
	if err != nil {
		return "", err
	}
	defer os.Remove(envPath)

	return resolveComposeProjectName(composePath, "", envPath)
}

func createComposeTempFile(dir, pattern, content string) (string, error) {
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	filePath := file.Name()
	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		_ = os.Remove(filePath)
		return "", err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(filePath)
		return "", err
	}
	return filePath, nil
}

func prepareComposeStagingDir(dir string) (func(), error) {
	if err := os.MkdirAll(filepath.Dir(dir), os.ModePerm); err != nil {
		return nil, err
	}
	created := false
	if err := os.Mkdir(dir, os.ModePerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
	} else {
		created = true
	}
	return func() {
		if created {
			_ = os.Remove(dir)
		}
	}, nil
}

func resolveComposeProjectName(composePath, fallbackName, envFile string) (string, error) {
	// Preserve the name resolved by Compose (including a top-level name) so the
	// container label and the local record always use the same project identity.
	parentName := normalizeComposeProjectName(path.Base(path.Dir(primaryComposePath(composePath))))
	fallbackName = strings.TrimSpace(fallbackName)
	stdout, err := runComposeConfig(composePath, "", envFile)
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
			if _, fallbackErr := runComposeConfig(composePath, fallbackName, envFile); fallbackErr != nil {
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
		if _, parentErr := runComposeConfig(composePath, parentName, envFile); parentErr == nil {
			return parentName, nil
		} else {
			resolveErr = parentErr
		}
	}

	if fallbackName != "" && fallbackName != parentName {
		if _, fallbackErr := runComposeConfig(composePath, fallbackName, envFile); fallbackErr == nil {
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

func runComposeConfig(composePath, projectName, envFile string) ([]byte, error) {
	configCmd := getComposeCmdWithEnv(composePath, "config", envFile, projectName)
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
	if cmd.CheckIllegal(req.Path, req.Operation) {
		return buserr.New("ErrCmdIllegal")
	}
	if req.Operation == "delete" {
		if err := removeContainerForCompose(req.Name, req.Path); err != nil && !req.Force {
			return err
		}
		if req.WithFile {
			for _, item := range strings.Split(req.Path, ",") {
				item = strings.TrimSpace(item)
				if !filepath.IsAbs(item) || filepath.Dir(item) == "/" {
					continue
				}
				info, err := os.Stat(item)
				if err == nil && info.Mode().IsRegular() {
					if err := os.RemoveAll(filepath.Dir(item)); err != nil && !req.Force {
						return err
					}
				}
			}
		}
		return composeRepo.DeleteRecord(repo.WithByName(req.Name))
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
			if err := newComposeEnv(primaryComposePath(req.Path), req.Env); err != nil {
				return err
			}

			if err := compose.UpWithTask(req.Path, t, req.ForcePull, req.Name); err != nil {
				global.LOG.Errorf("update failed when handle compose up, err: %s, now try to recreate the old compose file", err)
				if err := recreateCompose(string(oldFile), req.Path, req.DetailPath, req.Name); err != nil {
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
		composePath := composeCreatePath(*req)
		dir := filepath.Dir(composePath)
		if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
		}

		file, err := os.OpenFile(composePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, constant.FilePerm)
		if err != nil {
			return err
		}
		defer file.Close()
		write := bufio.NewWriter(file)
		_, _ = write.WriteString(string(req.File))
		write.Flush()
		req.Path = composePath
	}
	return nil
}

func removeContainerForCompose(composeName, composePath string) error {
	composeFiles, err := resolveComposeFiles("", composePath)
	filesAvailable := err == nil
	for _, file := range composeFiles {
		info, err := os.Stat(file)
		if err != nil || !info.Mode().IsRegular() {
			filesAvailable = false
			break
		}
	}
	if filesAvailable {
		if stdout, err := compose.Operate(strings.Join(composeFiles, ","), "down", composeName); err != nil {
			return fmt.Errorf("docker-compose down failed, std: %s, err: %w", stdout, err)
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
	var removeErrors []error
	for _, c := range containers {
		if err := client.ContainerRemove(ctx, c.ID, container.RemoveOptions{RemoveVolumes: true, Force: true}); err != nil {
			removeErrors = append(removeErrors, fmt.Errorf("remove container %s failed: %w", c.ID, err))
		}
	}
	return errors.Join(removeErrors...)
}

func recreateCompose(content, composePath, detailPath, projectName string) error {
	file, err := os.OpenFile(detailPath, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(content)
	write.Flush()

	if stdout, err := compose.Up(composePath, projectName); err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

func loadEnv(list []dto.ComposeInfo) []dto.ComposeInfo {
	for i := 0; i < len(list); i++ {
		if list[i].Path == "" {
			continue
		}
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

func resolveComposeFiles(workdir, configFile string) ([]string, error) {
	workdir = strings.TrimSpace(workdir)
	var files []string
	for _, item := range strings.Split(configFile, ",") {
		file := strings.TrimSpace(item)
		if file == "" {
			continue
		}
		if !filepath.IsAbs(file) {
			if !filepath.IsAbs(workdir) {
				return nil, fmt.Errorf("cannot resolve compose file %s without an absolute working directory", file)
			}
			file = filepath.Join(workdir, file)
		}
		files = append(files, filepath.Clean(file))
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("compose config file is not specified")
	}
	return files, nil
}

func composeFileExists(files []string) bool {
	for _, file := range files {
		info, err := os.Stat(file)
		if err == nil && info.Mode().IsRegular() {
			return true
		}
	}
	return false
}
