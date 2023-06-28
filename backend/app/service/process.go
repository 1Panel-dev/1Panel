package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/shirou/gopsutil/v3/process"
)

type ProcessService struct{}

type IProcessService interface {
	StopProcess(req request.ProcessReq) error
}

func NewIProcessService() IProcessService {
	return &ProcessService{}
}

func (p *ProcessService) StopProcess(req request.ProcessReq) error {
	proc, err := process.NewProcess(req.PID)
	if err != nil {
		return err
	}
	if err := proc.Kill(); err != nil {
		return err
	}
	return nil
}
