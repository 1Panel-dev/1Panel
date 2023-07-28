package service

import (
	"bytes"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/ini_conf"
	"github.com/1Panel-dev/1Panel/backend/utils/systemctl"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
	"os/exec"
	"path"
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
}

func NewIHostToolService() IHostToolService {
	return &HostToolService{}
}

func (h *HostToolService) GetToolStatus(req request.HostToolReq) (*response.HostToolRes, error) {
	res := &response.HostToolRes{}
	res.Type = req.Type
	switch req.Type {
	case constant.Supervisord:
		exist, err := systemctl.IsExist(constant.Supervisord)
		if err != nil {
			return nil, err
		}
		supervisorConfig := &response.Supervisor{}
		if !exist {
			supervisorConfig.IsExist = false
			return res, nil
		}
		supervisorConfig.IsExist = true

		versionRes, _ := cmd.Exec("supervisord -v")
		supervisorConfig.Version = strings.TrimSuffix(versionRes, "\n")
		_, ctlRrr := exec.LookPath("supervisorctl")
		supervisorConfig.CtlExist = ctlRrr == nil

		active, err := systemctl.IsActive(constant.Supervisord)
		if err != nil {
			supervisorConfig.Status = "unhealthy"
			supervisorConfig.Msg = err.Error()
			res.Config = supervisorConfig
			return res, nil
		}
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
		}
		supervisorConfig.Init = true

		servicePath := "/usr/lib/systemd/system/supervisord.service"
		fileOp := files.NewFileOp()
		if !fileOp.Stat(servicePath) {
			servicePath = "/lib/systemd/system/supervisord.service"
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
		} else {
			configPath := "/etc/supervisord.conf"
			if !fileOp.Stat(configPath) {
				configPath = "/etc/supervisor/supervisord.conf"
				if !fileOp.Stat("configPath") {
					return nil, errors.New("ErrConfigNotFound")
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
			return errors.New("ErrConfigNotFound")
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
		if err = settingRepo.Create(constant.SupervisorConfigPath, req.ConfigPath); err != nil {
			return err
		}
		go func() {
			if err = systemctl.Restart(constant.Supervisord); err != nil {
				global.LOG.Errorf("[init] restart supervisord failed err %s", err.Error())
			}
		}()
	}
	return nil
}

func (h *HostToolService) OperateTool(req request.HostToolReq) error {
	return systemctl.Operate(req.Operate, req.Type)
}

func (h *HostToolService) OperateToolConfig(req request.HostToolConfig) (*response.HostToolConfig, error) {
	fileOp := files.NewFileOp()
	res := &response.HostToolConfig{}
	configPath := ""
	switch req.Type {
	case constant.Supervisord:
		pathSet, _ := settingRepo.Get(settingRepo.WithByKey(constant.SupervisorConfigPath))
		if pathSet.ID != 0 || pathSet.Value != "" {
			configPath = pathSet.Value
		}
	}
	configPath = "/etc/supervisord.conf"
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
		if err = systemctl.Restart(req.Type); err != nil {
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
	configFile := ini.Empty()
	supervisordDir := path.Join(global.CONF.System.BaseDir, "1panel", "tools", "supervisord")
	logDir := path.Join(supervisordDir, "log")
	includeDir := path.Join(supervisordDir, "supervisor.d")

	section, err := configFile.NewSection(fmt.Sprintf("program:%s", req.Name))
	if err != nil {
		return err
	}
	_, _ = section.NewKey("command", req.Command)
	_, _ = section.NewKey("directory", req.Dir)
	_, _ = section.NewKey("autorestart", "true")
	_, _ = section.NewKey("startsecs", "3")
	_, _ = section.NewKey("stdout_logfile", path.Join(logDir, fmt.Sprintf("%s.out.log", req.Name)))
	_, _ = section.NewKey("stderr_logfile", path.Join(logDir, fmt.Sprintf("%s.err.log", req.Name)))
	_, _ = section.NewKey("stdout_logfile_maxbytes", "2MB")
	_, _ = section.NewKey("stderr_logfile_maxbytes", "2MB")
	_, _ = section.NewKey("user", req.User)
	_, _ = section.NewKey("priority", "999")
	_, _ = section.NewKey("numprocs", req.Numprocs)
	_, _ = section.NewKey("process_name", "%(program_name)s_%(process_num)02d")

	return configFile.SaveTo(path.Join(includeDir, fmt.Sprintf("%s.ini", req.Name)))
}
