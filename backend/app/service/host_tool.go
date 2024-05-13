package service

import (
	"bytes"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/ini_conf"
	"github.com/1Panel-dev/1Panel/backend/utils/systemctl"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
)

type HostToolService struct{}

type IHostToolService interface {
	GetToolStatus(req request.HostToolReq) (*response.HostToolRes, error)
	CreateToolConfig(req request.HostToolCreate) error
	OperateTool(req request.HostToolReq) error
	OperateToolConfig(req request.HostToolConfig) (*response.HostToolConfig, error)
	GetToolLog(req request.HostToolLogReq) (string, error)
	OperateSupervisorProcess(req request.SupervisorProcessConfig) error
	GetSupervisorProcessConfig() ([]response.SupervisorProcessConfig, error)
	OperateSupervisorProcessFile(req request.SupervisorProcessFileReq) (string, error)
}

func NewIHostToolService() IHostToolService {
	return &HostToolService{}
}

func (h *HostToolService) GetToolStatus(req request.HostToolReq) (*response.HostToolRes, error) {
	res := &response.HostToolRes{}
	res.Type = req.Type
	switch req.Type {
	case constant.Supervisord:
		supervisorConfig := &response.Supervisor{}
		if !cmd.Which(constant.Supervisord) {
			supervisorConfig.IsExist = false
			res.Config = supervisorConfig
			return res, nil
		}
		supervisorConfig.IsExist = true
		serviceExist, _ := systemctl.IsExist(constant.Supervisord)
		if !serviceExist {
			serviceExist, _ = systemctl.IsExist(constant.Supervisor)
			if !serviceExist {
				supervisorConfig.IsExist = false
				res.Config = supervisorConfig
				return res, nil
			} else {
				supervisorConfig.ServiceName = constant.Supervisor
			}
		} else {
			supervisorConfig.ServiceName = constant.Supervisord
		}

		serviceNameSet, _ := settingRepo.Get(settingRepo.WithByKey(constant.SupervisorServiceName))
		if serviceNameSet.ID != 0 || serviceNameSet.Value != "" {
			supervisorConfig.ServiceName = serviceNameSet.Value
		}

		versionRes, _ := cmd.Exec("supervisord -v")
		supervisorConfig.Version = strings.TrimSuffix(versionRes, "\n")
		_, ctlRrr := exec.LookPath("supervisorctl")
		supervisorConfig.CtlExist = ctlRrr == nil

		active, _ := systemctl.IsActive(supervisorConfig.ServiceName)
		if active {
			supervisorConfig.Status = "running"
		} else {
			supervisorConfig.Status = "stopped"
		}

		pathSet, _ := settingRepo.Get(settingRepo.WithByKey(constant.SupervisorConfigPath))
		if pathSet.ID != 0 || pathSet.Value != "" {
			supervisorConfig.ConfigPath = pathSet.Value
			res.Config = supervisorConfig
			return res, nil
		} else {
			supervisorConfig.Init = true
		}

		servicePath := "/usr/lib/systemd/system/supervisor.service"
		fileOp := files.NewFileOp()
		if !fileOp.Stat(servicePath) {
			servicePath = "/usr/lib/systemd/system/supervisord.service"
		}
		if fileOp.Stat(servicePath) {
			startCmd, _ := ini_conf.GetIniValue(servicePath, "Service", "ExecStart")
			if startCmd != "" {
				args := strings.Fields(startCmd)
				cIndex := -1
				for i, arg := range args {
					if arg == "-c" {
						cIndex = i
						break
					}
				}
				if cIndex != -1 && cIndex+1 < len(args) {
					supervisorConfig.ConfigPath = args[cIndex+1]
				}
			}
		}
		if supervisorConfig.ConfigPath == "" {
			configPath := "/etc/supervisord.conf"
			if !fileOp.Stat(configPath) {
				configPath = "/etc/supervisor/supervisord.conf"
				if fileOp.Stat(configPath) {
					supervisorConfig.ConfigPath = configPath
				}
			}
		}

		res.Config = supervisorConfig
	}
	return res, nil
}

func (h *HostToolService) CreateToolConfig(req request.HostToolCreate) error {
	switch req.Type {
	case constant.Supervisord:
		fileOp := files.NewFileOp()
		if !fileOp.Stat(req.ConfigPath) {
			return buserr.New("ErrConfigNotFound")
		}
		cfg, err := ini.Load(req.ConfigPath)
		if err != nil {
			return err
		}
		service, err := cfg.GetSection("include")
		if err != nil {
			return err
		}
		targetKey, err := service.GetKey("files")
		if err != nil {
			return err
		}
		if targetKey != nil {
			_, err = service.NewKey(";files", targetKey.Value())
			if err != nil {
				return err
			}
		}
		supervisorDir := path.Join(global.CONF.System.BaseDir, "1panel", "tools", "supervisord")
		includeDir := path.Join(supervisorDir, "supervisor.d")
		if !fileOp.Stat(includeDir) {
			if err = fileOp.CreateDir(includeDir, 0755); err != nil {
				return err
			}
		}
		logDir := path.Join(supervisorDir, "log")
		if !fileOp.Stat(logDir) {
			if err = fileOp.CreateDir(logDir, 0755); err != nil {
				return err
			}
		}
		includePath := path.Join(includeDir, "*.ini")
		targetKey.SetValue(includePath)
		if err = cfg.SaveTo(req.ConfigPath); err != nil {
			return err
		}

		serviceNameSet, _ := settingRepo.Get(settingRepo.WithByKey(constant.SupervisorServiceName))
		if serviceNameSet.ID != 0 {
			if err = settingRepo.Update(constant.SupervisorServiceName, req.ServiceName); err != nil {
				return err
			}
		} else {
			if err = settingRepo.Create(constant.SupervisorServiceName, req.ServiceName); err != nil {
				return err
			}
		}

		configPathSet, _ := settingRepo.Get(settingRepo.WithByKey(constant.SupervisorConfigPath))
		if configPathSet.ID != 0 {
			if err = settingRepo.Update(constant.SupervisorConfigPath, req.ConfigPath); err != nil {
				return err
			}
		} else {
			if err = settingRepo.Create(constant.SupervisorConfigPath, req.ConfigPath); err != nil {
				return err
			}
		}
		if err = systemctl.Restart(req.ServiceName); err != nil {
			global.LOG.Errorf("[init] restart %s failed err %s", req.ServiceName, err.Error())
			return err
		}
	}
	return nil
}

func (h *HostToolService) OperateTool(req request.HostToolReq) error {
	serviceName := req.Type
	if req.Type == constant.Supervisord {
		serviceNameSet, _ := settingRepo.Get(settingRepo.WithByKey(constant.SupervisorServiceName))
		if serviceNameSet.ID != 0 || serviceNameSet.Value != "" {
			serviceName = serviceNameSet.Value
		}
	}
	return systemctl.Operate(req.Operate, serviceName)
}

func (h *HostToolService) OperateToolConfig(req request.HostToolConfig) (*response.HostToolConfig, error) {
	fileOp := files.NewFileOp()
	res := &response.HostToolConfig{}
	configPath := ""
	serviceName := "supervisord"
	switch req.Type {
	case constant.Supervisord:
		pathSet, _ := settingRepo.Get(settingRepo.WithByKey(constant.SupervisorConfigPath))
		if pathSet.ID != 0 || pathSet.Value != "" {
			configPath = pathSet.Value
		}
		serviceNameSet, _ := settingRepo.Get(settingRepo.WithByKey(constant.SupervisorServiceName))
		if serviceNameSet.ID != 0 || serviceNameSet.Value != "" {
			serviceName = serviceNameSet.Value
		}
	}
	switch req.Operate {
	case "get":
		content, err := fileOp.GetContent(configPath)
		if err != nil {
			return nil, err
		}
		res.Content = string(content)
	case "set":
		file, err := fileOp.OpenFile(configPath)
		if err != nil {
			return nil, err
		}
		oldContent, err := fileOp.GetContent(configPath)
		if err != nil {
			return nil, err
		}
		fileInfo, err := file.Stat()
		if err != nil {
			return nil, err
		}
		if err = fileOp.WriteFile(configPath, strings.NewReader(req.Content), fileInfo.Mode()); err != nil {
			return nil, err
		}
		if err = systemctl.Restart(serviceName); err != nil {
			_ = fileOp.WriteFile(configPath, bytes.NewReader(oldContent), fileInfo.Mode())
			return nil, err
		}
	}

	return res, nil
}

func (h *HostToolService) GetToolLog(req request.HostToolLogReq) (string, error) {
	fileOp := files.NewFileOp()
	logfilePath := ""
	switch req.Type {
	case constant.Supervisord:
		configPath := "/etc/supervisord.conf"
		pathSet, _ := settingRepo.Get(settingRepo.WithByKey(constant.SupervisorConfigPath))
		if pathSet.ID != 0 || pathSet.Value != "" {
			configPath = pathSet.Value
		}
		logfilePath, _ = ini_conf.GetIniValue(configPath, "supervisord", "logfile")
	}
	oldContent, err := fileOp.GetContent(logfilePath)
	if err != nil {
		return "", err
	}
	return string(oldContent), nil
}

func (h *HostToolService) OperateSupervisorProcess(req request.SupervisorProcessConfig) error {
	var (
		supervisordDir = path.Join(global.CONF.System.BaseDir, "1panel", "tools", "supervisord")
		logDir         = path.Join(supervisordDir, "log")
		includeDir     = path.Join(supervisordDir, "supervisor.d")
		outLog         = path.Join(logDir, fmt.Sprintf("%s.out.log", req.Name))
		errLog         = path.Join(logDir, fmt.Sprintf("%s.err.log", req.Name))
		iniPath        = path.Join(includeDir, fmt.Sprintf("%s.ini", req.Name))
		fileOp         = files.NewFileOp()
	)
	if req.Operate == "update" || req.Operate == "create" {
		if !fileOp.Stat(req.Dir) {
			return buserr.New("ErrConfigDirNotFound")
		}
		_, err := user.Lookup(req.User)
		if err != nil {
			return buserr.WithMap("ErrUserFindErr", map[string]interface{}{"name": req.User, "err": err.Error()}, err)
		}
	}

	switch req.Operate {
	case "create":
		if fileOp.Stat(iniPath) {
			return buserr.New("ErrConfigAlreadyExist")
		}
		configFile := ini.Empty()
		section, err := configFile.NewSection(fmt.Sprintf("program:%s", req.Name))
		if err != nil {
			return err
		}
		_, _ = section.NewKey("command", strings.TrimSpace(req.Command))
		_, _ = section.NewKey("directory", req.Dir)
		_, _ = section.NewKey("autorestart", "true")
		_, _ = section.NewKey("startsecs", "3")
		_, _ = section.NewKey("stdout_logfile", outLog)
		_, _ = section.NewKey("stderr_logfile", errLog)
		_, _ = section.NewKey("stdout_logfile_maxbytes", "2MB")
		_, _ = section.NewKey("stderr_logfile_maxbytes", "2MB")
		_, _ = section.NewKey("user", req.User)
		_, _ = section.NewKey("priority", "999")
		_, _ = section.NewKey("numprocs", req.Numprocs)
		_, _ = section.NewKey("process_name", "%(program_name)s_%(process_num)02d")

		if err = configFile.SaveTo(iniPath); err != nil {
			return err
		}
		if err := operateSupervisorCtl("reread", "", ""); err != nil {
			return err
		}
		return operateSupervisorCtl("update", "", "")
	case "update":
		configFile, err := ini.Load(iniPath)
		if err != nil {
			return err
		}
		section, err := configFile.GetSection(fmt.Sprintf("program:%s", req.Name))
		if err != nil {
			return err
		}

		commandKey := section.Key("command")
		commandKey.SetValue(strings.TrimSpace(req.Command))
		directoryKey := section.Key("directory")
		directoryKey.SetValue(req.Dir)
		userKey := section.Key("user")
		userKey.SetValue(req.User)
		numprocsKey := section.Key("numprocs")
		numprocsKey.SetValue(req.Numprocs)

		if err = configFile.SaveTo(iniPath); err != nil {
			return err
		}
		if err := operateSupervisorCtl("reread", "", ""); err != nil {
			return err
		}
		return operateSupervisorCtl("update", "", "")
	case "restart":
		return operateSupervisorCtl("restart", req.Name, "")
	case "start":
		return operateSupervisorCtl("start", req.Name, "")
	case "stop":
		return operateSupervisorCtl("stop", req.Name, "")
	case "delete":
		_ = operateSupervisorCtl("remove", "", req.Name)
		_ = files.NewFileOp().DeleteFile(iniPath)
		_ = files.NewFileOp().DeleteFile(outLog)
		_ = files.NewFileOp().DeleteFile(errLog)
		if err := operateSupervisorCtl("reread", "", ""); err != nil {
			return err
		}
		return operateSupervisorCtl("update", "", "")
	}

	return nil
}

func (h *HostToolService) GetSupervisorProcessConfig() ([]response.SupervisorProcessConfig, error) {
	var (
		result []response.SupervisorProcessConfig
	)
	configDir := path.Join(global.CONF.System.BaseDir, "1panel", "tools", "supervisord", "supervisor.d")
	fileList, _ := NewIFileService().GetFileList(request.FileOption{FileOption: files.FileOption{Path: configDir, Expand: true, Page: 1, PageSize: 100}})
	if len(fileList.Items) == 0 {
		return result, nil
	}
	for _, configFile := range fileList.Items {
		f, err := ini.Load(configFile.Path)
		if err != nil {
			global.LOG.Errorf("get %s file err %s", configFile.Name, err.Error())
			continue
		}
		if strings.HasSuffix(configFile.Name, ".ini") {
			config := response.SupervisorProcessConfig{}
			name := strings.TrimSuffix(configFile.Name, ".ini")
			config.Name = name
			section, err := f.GetSection(fmt.Sprintf("program:%s", name))
			if err != nil {
				global.LOG.Errorf("get %s file section err %s", configFile.Name, err.Error())
				continue
			}
			if command, _ := section.GetKey("command"); command != nil {
				config.Command = command.Value()
			}
			if directory, _ := section.GetKey("directory"); directory != nil {
				config.Dir = directory.Value()
			}
			if user, _ := section.GetKey("user"); user != nil {
				config.User = user.Value()
			}
			if numprocs, _ := section.GetKey("numprocs"); numprocs != nil {
				config.Numprocs = numprocs.Value()
			}
			_ = getProcessStatus(&config)
			result = append(result, config)
		}
	}
	return result, nil
}

func (h *HostToolService) OperateSupervisorProcessFile(req request.SupervisorProcessFileReq) (string, error) {
	var (
		fileOp     = files.NewFileOp()
		group      = fmt.Sprintf("program:%s", req.Name)
		configPath = path.Join(global.CONF.System.BaseDir, "1panel", "tools", "supervisord", "supervisor.d", fmt.Sprintf("%s.ini", req.Name))
	)
	switch req.File {
	case "err.log":
		logPath, err := ini_conf.GetIniValue(configPath, group, "stderr_logfile")
		if err != nil {
			return "", err
		}
		switch req.Operate {
		case "get":
			content, err := fileOp.GetContent(logPath)
			if err != nil {
				return "", err
			}
			return string(content), nil
		case "clear":
			if err = fileOp.WriteFile(logPath, strings.NewReader(""), 0755); err != nil {
				return "", err
			}
		}

	case "out.log":
		logPath, err := ini_conf.GetIniValue(configPath, group, "stdout_logfile")
		if err != nil {
			return "", err
		}
		switch req.Operate {
		case "get":
			content, err := fileOp.GetContent(logPath)
			if err != nil {
				return "", err
			}
			return string(content), nil
		case "clear":
			if err = fileOp.WriteFile(logPath, strings.NewReader(""), 0755); err != nil {
				return "", err
			}
		}

	case "config":
		switch req.Operate {
		case "get":
			content, err := fileOp.GetContent(configPath)
			if err != nil {
				return "", err
			}
			return string(content), nil
		case "update":
			if req.Content == "" {
				return "", buserr.New("ErrConfigIsNull")
			}
			if err := fileOp.WriteFile(configPath, strings.NewReader(req.Content), 0755); err != nil {
				return "", err
			}
			return "", operateSupervisorCtl("update", "", req.Name)
		}

	}
	return "", nil
}

func operateSupervisorCtl(operate, name, group string) error {
	processNames := []string{operate}
	if name != "" {
		includeDir := path.Join(global.CONF.System.BaseDir, "1panel", "tools", "supervisord", "supervisor.d")
		f, err := ini.Load(path.Join(includeDir, fmt.Sprintf("%s.ini", name)))
		if err != nil {
			return err
		}
		section, err := f.GetSection(fmt.Sprintf("program:%s", name))
		if err != nil {
			return err
		}
		numprocsNum := ""
		if numprocs, _ := section.GetKey("numprocs"); numprocs != nil {
			numprocsNum = numprocs.Value()
		}
		if numprocsNum == "" {
			return buserr.New("ErrConfigParse")
		}
		processNames = append(processNames, getProcessName(name, numprocsNum)...)
	}
	if group != "" {
		processNames = append(processNames, group)
	}

	output, err := exec.Command("supervisorctl", processNames...).Output()
	if err != nil {
		if output != nil {
			return errors.New(string(output))
		}
		return err
	}
	return nil
}

func getProcessName(name, numprocs string) []string {
	var (
		processNames []string
	)
	num, err := strconv.Atoi(numprocs)
	if err != nil {
		return processNames
	}
	if num == 1 {
		processNames = append(processNames, fmt.Sprintf("%s:%s_00", name, name))
	} else {
		for i := 0; i < num; i++ {
			processName := fmt.Sprintf("%s:%s_0%s", name, name, strconv.Itoa(i))
			if i >= 10 {
				processName = fmt.Sprintf("%s:%s_%s", name, name, strconv.Itoa(i))
			}
			processNames = append(processNames, processName)
		}
	}
	return processNames
}

func getProcessStatus(config *response.SupervisorProcessConfig) error {
	var (
		processNames = []string{"status"}
	)
	processNames = append(processNames, getProcessName(config.Name, config.Numprocs)...)
	output, _ := exec.Command("supervisorctl", processNames...).Output()
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			status := response.ProcessStatus{
				Name:   fields[0],
				Status: fields[1],
			}
			if fields[1] == "RUNNING" {
				status.PID = strings.TrimSuffix(fields[3], ",")
				status.Uptime = fields[5]
			} else {
				status.Msg = strings.Join(fields[2:], " ")
			}
			config.Status = append(config.Status, status)
		}
	}
	return nil
}
