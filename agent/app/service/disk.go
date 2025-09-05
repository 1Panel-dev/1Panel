package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	disksWithPartitions, unpartitionedDisks, err := getAllDisksInfo()
	if err != nil {
		return nil, err
	}
	var systemDisk *response.DiskInfo
	var totalCapacity int64
	var filteredDisksWithPartitions []response.DiskInfo

	for i, disk := range disksWithPartitions {
		if disk.IsSystem {
			if systemDisk == nil {
				systemDisk = &disksWithPartitions[i]
			}
		} else {
			filteredDisksWithPartitions = append(filteredDisksWithPartitions, disk)
		}
	}

	completeDiskInfo := &response.CompleteDiskInfo{
		Disks:              filteredDisksWithPartitions,
		UnpartitionedDisks: unpartitionedDisks,
		SystemDisk:         systemDisk,
		TotalDisks:         len(filteredDisksWithPartitions) + len(unpartitionedDisks),
		TotalCapacity:      totalCapacity,
	}
	return completeDiskInfo, nil
}

func getDiskType(rota bool, tran string) string {
	if tran == "" {
		return ""
	}
	if !rota {
		return "SSD"
	}
	return "HDD"
}

func getAllDisksInfo() ([]response.DiskInfo, []response.DiskBasicInfo, error) {
	var disksWithPartitions []response.DiskInfo
	var unpartitionedDisks []response.DiskBasicInfo

	output, err := cmd.RunDefaultWithStdoutBashC("lsblk -J -o NAME,SIZE,TYPE,MOUNTPOINT,FSTYPE,MODEL,SERIAL,TRAN,ROTA")
	if err != nil {
		return nil, nil, err
	}

	var lsblkOutput dto.LsblkOutput
	if err = json.Unmarshal([]byte(output), &lsblkOutput); err != nil {
		return nil, nil, err
	}

	for _, device := range lsblkOutput.BlockDevices {
		if device.Type != "disk" {
			continue
		}

		devicePath := "/dev/" + device.Name
		model := ""
		isPhysical := false

		if device.Tran != "" {
			isPhysical = true
			if device.Model != nil {
				model = *device.Model
			}
		}

		hasPartitions := len(device.Children) > 0

		if hasPartitions {
			disk := response.DiskInfo{
				DiskBasicInfo: response.DiskBasicInfo{
					Device:      devicePath,
					Size:        device.Size,
					DiskType:    getDiskType(device.Rota, device.Tran),
					IsRemovable: isRemovableDisk(devicePath),
					IsSystem:    false,
				},
				Partitions: []response.DiskBasicInfo{},
			}

			if isPhysical {
				disk.Serial = device.Serial
				disk.Model = model
			}

			for _, partition := range device.Children {
				partitionInfo := processPartition(partition)
				if partitionInfo != nil {
					disk.Partitions = append(disk.Partitions, *partitionInfo)
					if partitionInfo.IsSystem {
						disk.IsSystem = true
					}
				}
			}

			if len(disk.Partitions) > 0 {
				disksWithPartitions = append(disksWithPartitions, disk)
			}
		} else {
			if isSystemDiskByDevice(device) {
				continue
			}

			unpartitionedDisk := response.DiskBasicInfo{
				Device:      devicePath,
				Size:        device.Size,
				Model:       model,
				Serial:      device.Serial,
				DiskType:    getDiskType(device.Rota, device.Tran),
				IsRemovable: isRemovableDisk(devicePath),
			}
			unpartitionedDisks = append(unpartitionedDisks, unpartitionedDisk)
		}
	}

	return disksWithPartitions, unpartitionedDisks, nil
}

func (s *DiskService) PartitionDisk(req request.DiskPartitionRequest) (string, error) {
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

func formatDisk(req dto.DiskFormatRequest) error {
	var mkfsCmd *exec.Cmd

	switch req.Filesystem {
	case "ext4":
		mkfsCmd = exec.Command("mkfs.ext4", "-F", req.Device)
	case "xfs":
		if !cmd.Which("mkfs.xfs") {
			return buserr.New("XfsNotFound")
		}
		mkfsCmd = exec.Command("mkfs.xfs", "-f", req.Device)
	default:
		return fmt.Errorf("unsupport type: %s", req.Filesystem)
	}
	if err := mkfsCmd.Run(); err != nil {
		return err
	}
	return nil
}

func parseDiskInfoLinux(fields []string) (response.DiskInfo, error) {
	device := fields[0]
	filesystem := fields[1]
	sizeStr := fields[2]
	usedStr := fields[3]
	availStr := fields[4]
	usePercentStr := fields[5]
	mountPoint := fields[6]

	usePercent := 0
	if strings.HasSuffix(usePercentStr, "%") {
		if percent, err := strconv.Atoi(strings.TrimSuffix(usePercentStr, "%")); err == nil {
			usePercent = percent
		}
	}

	return response.DiskInfo{
		DiskBasicInfo: response.DiskBasicInfo{
			Device:     device,
			Filesystem: filesystem,
			Size:       sizeStr,
			Used:       usedStr,
			Avail:      availStr,
			UsePercent: usePercent,
			MountPoint: mountPoint,
		},
	}, nil
}

func deviceExists(device string) bool {
	_, err := os.Stat(device)
	return err == nil
}

func isDeviceMounted(device string) bool {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[0] == device {
			return true
		}
	}
	return false
}

func isPointMounted(mountPoint string) bool {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[1] == mountPoint {
			return true
		}
	}
	return false
}

func addToFstabWithOptions(device, mountPoint, filesystem, options string) error {
	if filesystem == "" {
		fsType, err := getFilesystemType(device)
		if err != nil {
			filesystem = "auto"
		} else {
			filesystem = fsType
		}
	}

	if options == "" {
		options = "defaults"
	}

	entry := fmt.Sprintf("%s %s %s %s 0 2\n", device, mountPoint, filesystem, options)

	file, err := os.OpenFile("/etc/fstab", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(entry)
	return err
}

func removeFromFstab(mountPoint string) error {
	file, err := os.Open("/etc/fstab")
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == mountPoint {
			continue
		}
		lines = append(lines, line)
	}

	return os.WriteFile("/etc/fstab", []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

func getFilesystemType(device string) (string, error) {
	output, err := cmd.RunDefaultWithStdoutBashC(fmt.Sprintf("blkid -o value -s TYPE %s", device))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func isSystemDisk(mountPoint string) bool {
	systemMountPoints := []string{
		"/",
		"/boot",
		"/boot/efi",
		"/usr",
		"/var",
		"/home",
	}

	for _, sysMount := range systemMountPoints {
		if mountPoint == sysMount {
			return true
		}
	}

	return false
}

func isSystemDiskByDevice(device dto.LsblkDevice) bool {
	for _, child := range device.Children {
		if child.Type == "part" {
			if child.MountPoint != nil && isSystemDisk(*child.MountPoint) {
				return true
			}
			for _, lvmChild := range child.Children {
				if lvmChild.Type == "lvm" && lvmChild.MountPoint != nil {
					if isSystemDisk(*lvmChild.MountPoint) {
						return true
					}
				}
			}
		}
	}
	return false
}

func isRemovableDisk(device string) bool {
	deviceName := strings.TrimPrefix(device, "/dev/")
	for i := len(deviceName) - 1; i >= 0; i-- {
		if deviceName[i] < '0' || deviceName[i] > '9' {
			deviceName = deviceName[:i+1]
			break
		}
	}

	removablePath := filepath.Join("/sys/block", deviceName, "removable")
	if data, err := os.ReadFile(removablePath); err == nil {
		removable := strings.TrimSpace(string(data))
		return removable == "1"
	}

	return false
}

func processPartition(partition dto.LsblkDevice) *response.DiskBasicInfo {
	if partition.Type != "part" {
		return nil
	}

	devicePath := "/dev/" + partition.Name

	var mountPoint, filesystem string
	if partition.MountPoint != nil {
		mountPoint = *partition.MountPoint
	}
	if partition.FsType != nil {
		filesystem = *partition.FsType
	}

	var (
		actualMountPoint, actualFilesystem string
		size, used, avail                  string
		usePercent                         int
		isMounted                          bool
		isSystem                           bool
	)

	if len(partition.Children) > 0 {
		for _, child := range partition.Children {
			if child.Type == "lvm" {
				lvmDevicePath := "/dev/mapper/" + child.Name

				if child.MountPoint != nil {
					actualMountPoint = *child.MountPoint
					isMounted = true
				}
				if child.FsType != nil {
					actualFilesystem = *child.FsType
				}

				if actualMountPoint != "" {
					diskUsage, err := getDiskUsageInfo(lvmDevicePath)
					if err == nil {
						used = diskUsage.Used
						avail = diskUsage.Avail
						size = diskUsage.Size
						usePercent = diskUsage.UsePercent
					}
				}

				isSystem = isSystemDisk(actualMountPoint)
				break
			}
		}
	} else {
		actualMountPoint = mountPoint
		actualFilesystem = filesystem
		isMounted = mountPoint != ""

		if actualMountPoint != "" {
			diskUsage, err := getDiskUsageInfo(devicePath)
			if err == nil {
				used = diskUsage.Used
				avail = diskUsage.Avail
				size = diskUsage.Size
				usePercent = diskUsage.UsePercent
			}
		}

		isSystem = isSystemDisk(actualMountPoint)
	}

	if size == "" {
		size = partition.Size
	}

	return &response.DiskBasicInfo{
		Device:     devicePath,
		Size:       size,
		Used:       used,
		Avail:      avail,
		UsePercent: usePercent,
		MountPoint: actualMountPoint,
		Filesystem: actualFilesystem,
		IsMounted:  isMounted,
		IsSystem:   isSystem,
	}
}

func getDiskUsageInfo(device string) (response.DiskInfo, error) {
	output, err := cmd.RunDefaultWithStdoutBashC(fmt.Sprintf("df -hT -P %s", device))
	if err != nil {
		return response.DiskInfo{}, err
	}

	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return response.DiskInfo{}, err
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 7 {
		return response.DiskInfo{}, err
	}

	return parseDiskInfoLinux(fields)
}
