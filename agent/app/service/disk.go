package service

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
)

type DiskService struct{}

type IDiskService interface {
	GetCompleteDiskInfo() (*response.CompleteDiskInfo, error)
	PartitionDisk(req request.DiskPartitionRequest) (string, error)
	MountDisk(req request.DiskMountRequest) error
	UnmountDisk(req request.DiskUnmountRequest) error
}

func NewIDiskService() IDiskService {
	return &DiskService{}
}

func (s *DiskService) GetCompleteDiskInfo() (*response.CompleteDiskInfo, error) {
	output, err := cmd.RunDefaultWithStdoutBashC("lsblk -P -o NAME,SIZE,TYPE,MOUNTPOINT,FSTYPE,MODEL,SERIAL,TRAN,ROTA")
	if err != nil {
		return nil, fmt.Errorf("failed to execute lsblk command: %v", err)
	}

	diskInfos, err := parseLsblkOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %v", err)
	}

	result := organizeDiskInfo(diskInfos)
	return &result, nil
}

func (s *DiskService) PartitionDisk(req request.DiskPartitionRequest) (string, error) {
	if !strings.HasPrefix("/dev", req.Device) {
		req.Device = "/dev/" + req.Device
	}
	if !deviceExists(req.Device) {
		return "", buserr.WithName("DeviceNotFound", req.Device)
	}

	if isDeviceMounted(req.Device) {
		return "", buserr.WithName("DeviceIsMounted", req.Device)
	}

	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(10 * time.Second))
	if err := cmdMgr.RunBashC(fmt.Sprintf("partprobe %s", req.Device)); err != nil {
		return "", buserr.WithErr("PartitionDiskErr", err)
	}

	if err := cmdMgr.RunBashC(fmt.Sprintf("parted -s %s mklabel gpt", req.Device)); err != nil {
		return "", buserr.WithErr("PartitionDiskErr", err)
	}

	if err := cmdMgr.RunBashC(fmt.Sprintf("parted -s %s mkpart primary 1MiB 100%%", req.Device)); err != nil {
		return "", buserr.WithErr("PartitionDiskErr", err)
	}

	if err := cmdMgr.RunBashC(fmt.Sprintf("partprobe %s", req.Device)); err != nil {
		return "", buserr.WithErr("PartitionDiskErr", err)
	}
	partition := req.Device + "1"

	formatReq := dto.DiskFormatRequest{
		Device:     partition,
		Filesystem: req.Filesystem,
		Label:      req.Label,
	}
	if err := formatDisk(formatReq); err != nil {
		return "", buserr.WithErr("FormatDiskErr", err)
	}

	if req.AutoMount && req.MountPoint != "" {
		mountReq := request.DiskMountRequest{
			Device:     partition,
			MountPoint: req.MountPoint,
			Filesystem: req.Filesystem,
		}
		if err := s.MountDisk(mountReq); err != nil {
			return "", buserr.WithErr("MountDiskErr", err)
		}
	}

	return partition, nil
}

func (s *DiskService) MountDisk(req request.DiskMountRequest) error {
	if !deviceExists(req.Device) {
		return buserr.WithName("DeviceNotFound", req.Device)
	}

	if err := os.MkdirAll(req.MountPoint, 0755); err != nil {
		return err
	}

	fileSystem, err := getFilesystemType(req.Device)
	if err != nil {
		return err
	}
	if fileSystem == "" {
		formatReq := dto.DiskFormatRequest{
			Device:     req.Device,
			Filesystem: req.Filesystem,
		}
		if err := formatDisk(formatReq); err != nil {
			return buserr.WithErr("FormatDiskErr", err)
		}
	}

	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(1 * time.Minute))
	if err := cmdMgr.RunBashC(fmt.Sprintf("mount  -t %s %s %s", req.Filesystem, req.Device, req.MountPoint)); err != nil {
		return buserr.WithErr("MountDiskErr", err)
	}

	if err := addToFstabWithOptions(req.Device, req.MountPoint, req.Filesystem, ""); err != nil {
		return buserr.WithErr("MountDiskErr", err)
	}

	return nil
}

func (s *DiskService) UnmountDisk(req request.DiskUnmountRequest) error {
	if !isPointMounted(req.MountPoint) {
		return buserr.New("MountDiskErr")
	}
	if err := cmd.RunDefaultBashC(fmt.Sprintf("umount -f  %s", req.MountPoint)); err != nil {
		return buserr.WithErr("MountDiskErr", err)
	}
	if err := removeFromFstab(req.MountPoint); err != nil {
		global.LOG.Errorf("remove %s mountPoint err: %v", req.MountPoint, err)
	}
	return nil
}
