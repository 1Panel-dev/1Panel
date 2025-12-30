package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type LsblkRaw struct {
	Blockdevices []LsblkDevice `json:"blockdevices"`
}

type LsblkDevice struct {
	Name       string        `json:"name"`
	Size       string        `json:"size"`
	Type       string        `json:"type"`
	Mountpoint string        `json:"mountpoint"`
	Fstype     string        `json:"fstype"`
	Model      string        `json:"model"`
	Serial     string        `json:"serial"`
	Tran       string        `json:"tran"`
	Rota       bool          `json:"rota"`
	Children   []LsblkDevice `json:"children"`
}

func parseDevice(dev LsblkDevice) []response.DiskBasicInfo {
	var list []response.DiskBasicInfo

	if strings.HasPrefix(dev.Name, "loop") || strings.HasPrefix(dev.Name, "dm-") || dev.Type == "rom" {
		return list
	}

	if dev.Type == "lvm" {
		return list
	}

	diskType := "Unknown"
	if dev.Type == "disk" || dev.Type == "part" {
		if dev.Rota {
			diskType = "HDD"
		} else {
			diskType = "SSD"
		}
	}

	mountPoint := dev.Mountpoint
	filesystem := dev.Fstype
	size := dev.Size

	var used, avail, totalSize string
	var usePercent int
	isMounted := mountPoint != ""
	isSystem := false

	if dev.Fstype == "LVM2_member" && len(dev.Children) > 0 {
		for _, child := range dev.Children {
			if child.Type == "lvm" && child.Mountpoint != "" {
				devicePath := "/dev/mapper/" + child.Name
				totalSize, used, avail, usePercent, _ := getDiskUsageInfo(devicePath)

				childInfo := response.DiskBasicInfo{
					Device:      dev.Name,
					Size:        totalSize,
					Model:       dev.Model,
					DiskType:    diskType,
					Filesystem:  child.Fstype,
					MountPoint:  child.Mountpoint,
					IsMounted:   true,
					UsePercent:  usePercent,
					Used:        used,
					Avail:       avail,
					Serial:      dev.Serial,
					IsRemovable: dev.Tran == "usb",
					IsSystem:    isSystemDisk(child.Mountpoint),
				}
				list = append(list, childInfo)
			}
		}
		return list
	} else if isMounted {
		isSystem = isSystemDisk(mountPoint)
		devicePath := "/dev/" + dev.Name
		totalSize, used, avail, usePercent, _ = getDiskUsageInfo(devicePath)
		if totalSize != "" {
			size = totalSize
		}
	}

	info := response.DiskBasicInfo{
		Device:      dev.Name,
		Size:        size,
		Model:       dev.Model,
		DiskType:    diskType,
		Filesystem:  filesystem,
		MountPoint:  mountPoint,
		IsMounted:   isMounted,
		UsePercent:  usePercent,
		Used:        used,
		Avail:       avail,
		Serial:      dev.Serial,
		IsRemovable: dev.Tran == "usb",
		IsSystem:    isSystem,
	}

	list = append(list, info)

	for _, child := range dev.Children {
		childList := parseDevice(child)
		list = append(list, childList...)
	}

	return list
}

func parseLsblkJsonOutput(output string) ([]response.DiskBasicInfo, error) {
	raw := &LsblkRaw{}
	if err := json.Unmarshal([]byte(output), raw); err != nil {
		return nil, fmt.Errorf("failed to parse lsblk json output: %v", err)
	}
	var disks []response.DiskBasicInfo

	for _, dev := range raw.Blockdevices {
		if strings.HasPrefix(dev.Name, "loop") ||
			strings.HasPrefix(dev.Name, "dm-") {
			continue
		}
		devList := parseDevice(dev)
		disks = append(disks, devList...)
	}
	return disks, nil
}

func organizeDiskInfo(diskInfos []response.DiskBasicInfo) response.CompleteDiskInfo {
	var result response.CompleteDiskInfo
	diskMap := make(map[string]*response.DiskInfo)
	partitions := make(map[string][]response.DiskBasicInfo)

	for _, info := range diskInfos {
		isPartition := isPartitionDevice(info.Device)

		if isPartition {
			parentDevice := getParentDevice(info.Device)
			partitions[parentDevice] = append(partitions[parentDevice], info)
		} else {
			disk := &response.DiskInfo{
				DiskBasicInfo: info,
				Partitions:    []response.DiskBasicInfo{},
			}
			diskMap[info.Device] = disk
		}
	}

	for parentDevice, partList := range partitions {
		if disk, exists := diskMap[parentDevice]; exists {
			for index, part := range partList {
				part.Device = fmt.Sprintf("/dev/%s", part.Device)
				if part.IsSystem {
					disk.IsSystem = true
				}
				partList[index] = part
			}
			disk.Partitions = partList
		}
	}

	var totalCapacity int64
	for _, disk := range diskMap {
		capacity := parseSizeToBytes(disk.Size)
		totalCapacity += capacity

		if disk.IsSystem {
			result.SystemDisks = append(result.SystemDisks, *disk)
		} else if len(disk.Partitions) == 0 {
			result.UnpartitionedDisks = append(result.UnpartitionedDisks, disk.DiskBasicInfo)
		} else {
			result.Disks = append(result.Disks, *disk)
		}
	}

	result.TotalDisks = len(diskMap)
	result.TotalCapacity = totalCapacity

	return result
}

func parseLsblkOutput(output string) ([]response.DiskBasicInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("invalid lsblk output")
	}

	var diskInfos []response.DiskBasicInfo
	lvmMap := make(map[string]response.DiskBasicInfo)
	var pendingDevices []map[string]string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := parseKeyValuePairs(line)

		name, ok := fields["NAME"]
		if !ok {
			continue
		}

		if strings.HasPrefix(name, "loop") ||
			strings.HasPrefix(name, "dm-") {
			continue
		}

		diskType := fields["TYPE"]
		mountPoint := fields["MOUNTPOINT"]
		fsType := fields["FSTYPE"]
		size := fields["SIZE"]

		if diskType == "lvm" {
			total, used, avail, usePercent, _ := getDiskUsageInfo("/dev/mapper/" + name)
			if total != "" && fsType != "" {
				size = total
			}

			lvmInfo := response.DiskBasicInfo{
				Device:      name,
				Size:        size,
				Model:       fields["MODEL"],
				DiskType:    "LVM",
				IsRemovable: false,
				IsSystem:    isSystemDisk(mountPoint),
				Filesystem:  fsType,
				Used:        used,
				Avail:       avail,
				UsePercent:  usePercent,
				MountPoint:  mountPoint,
				IsMounted:   mountPoint != "" && mountPoint != "-",
				Serial:      fields["SERIAL"],
			}
			lvmMap[name] = lvmInfo
		} else if diskType == "disk" || diskType == "part" {
			pendingDevices = append(pendingDevices, fields)
		}
	}

	for _, fields := range pendingDevices {
		name := fields["NAME"]
		size := fields["SIZE"]
		mountPoint := fields["MOUNTPOINT"]
		fsType := fields["FSTYPE"]
		model := fields["MODEL"]
		serial := fields["SERIAL"]
		tran := fields["TRAN"]
		rota := fields["ROTA"]

		var (
			used, avail, totalSize string
			usePercent             int
		)
		if mountPoint != "" {
			totalSize, used, avail, usePercent, _ = getDiskUsageInfo("/dev/" + name)
			if totalSize != "" {
				size = totalSize
			}
		}

		actualMountPoint := mountPoint
		actualFsType := fsType
		actualUsed := used
		actualAvail := avail
		actualUsePercent := usePercent
		isMounted := mountPoint != "" && mountPoint != "-"
		isSystemPartition := isSystemDisk(mountPoint)

		if fsType == "LVM2_member" {
			for _, lvmInfo := range lvmMap {
				if lvmInfo.IsMounted {
					lvmDiskInfo := response.DiskBasicInfo{
						Device:      name,
						Size:        lvmInfo.Size,
						Model:       model,
						DiskType:    "LVM",
						IsRemovable: tran == "usb",
						IsSystem:    lvmInfo.IsSystem,
						Filesystem:  lvmInfo.Filesystem,
						Used:        lvmInfo.Used,
						Avail:       lvmInfo.Avail,
						UsePercent:  lvmInfo.UsePercent,
						MountPoint:  lvmInfo.MountPoint,
						IsMounted:   true,
						Serial:      serial,
					}
					diskInfos = append(diskInfos, lvmDiskInfo)
				}
			}
			continue
		}

		info := response.DiskBasicInfo{
			Device:      name,
			Size:        size,
			Model:       model,
			DiskType:    getDiskType(rota),
			IsRemovable: tran == "usb",
			IsSystem:    isSystemPartition,
			Filesystem:  actualFsType,
			Used:        actualUsed,
			Avail:       actualAvail,
			UsePercent:  actualUsePercent,
			MountPoint:  actualMountPoint,
			IsMounted:   isMounted,
			Serial:      serial,
		}

		diskInfos = append(diskInfos, info)
	}

	return diskInfos, nil
}

func getDiskType(rota string) string {
	if rota == "0" {
		return "SSD"
	} else if rota == "1" {
		return "HDD"
	}
	return "Unknown"
}

func parseKeyValuePairs(line string) map[string]string {
	fields := make(map[string]string)

	matches := re.GetRegex(re.DiskKeyValuePattern).FindAllStringSubmatch(line, -1)
	for _, m := range matches {
		key := m[1]
		raw := m[2]

		val := raw
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			if unq, err := strconv.Unquote(val); err == nil {
				val = unq
			} else {
				val = val[1 : len(val)-1]
			}
		}
		fields[key] = val
	}
	return fields
}

func isPartitionDevice(device string) bool {
	if strings.Contains(device, "nvme") {
		return strings.Contains(device, "p") &&
			strings.ContainsAny(device[strings.LastIndex(device, "p")+1:], "0123456789")
	} else if strings.HasPrefix(device, "sd") || strings.HasPrefix(device, "hd") {
		return len(device) > 3 &&
			strings.ContainsAny(device[len(device)-1:], "0123456789")
	} else if strings.HasPrefix(device, "vd") {
		return len(device) > 3 &&
			strings.ContainsAny(device[len(device)-1:], "0123456789")
	}

	return false
}

func getParentDevice(device string) string {
	if strings.Contains(device, "nvme") {
		if idx := strings.LastIndex(device, "p"); idx != -1 {
			return device[:idx]
		}
	} else {
		return strings.TrimRight(device, "0123456789")
	}

	return device
}

func getDiskUsageInfo(device string) (size, used, avail string, usePercent int, err error) {
	output, err := cmd.RunDefaultWithStdoutBashC(fmt.Sprintf("df -h %s | tail -1", device))
	if err != nil {
		return "", "", "", 0, nil
	}

	fields := strings.Fields(output)
	if len(fields) >= 5 {
		size = fields[1]
		used = fields[2]
		avail = fields[3]
		usePercentStr := strings.TrimSuffix(fields[4], "%")
		usePercent, _ = strconv.Atoi(usePercentStr)
	}

	return size, used, avail, usePercent, nil
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
	uuid, err := getDeviceUUID(device)
	if err != nil {
		return fmt.Errorf("failed to get UUID for device %s: %v", device, err)
	}

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

	entry := fmt.Sprintf("UUID=%s %s %s %s 0 2\n", uuid, mountPoint, filesystem, options)

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
		"/etc",
	}

	for _, sysMount := range systemMountPoints {
		if mountPoint == sysMount {
			return true
		}
	}

	return false
}

func parseSizeToBytes(sizeStr string) int64 {
	if sizeStr == "" || sizeStr == "-" {
		return 0
	}

	sizeStr = strings.TrimSpace(sizeStr)
	if len(sizeStr) < 2 {
		return 0
	}

	unit := strings.ToUpper(sizeStr[len(sizeStr)-1:])
	valueStr := sizeStr[:len(sizeStr)-1]

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0
	}

	switch unit {
	case "K":
		return int64(value * 1024)
	case "M":
		return int64(value * 1024 * 1024)
	case "G":
		return int64(value * 1024 * 1024 * 1024)
	case "T":
		return int64(value * 1024 * 1024 * 1024 * 1024)
	default:
		val, _ := strconv.ParseInt(sizeStr, 10, 64)
		return val
	}
}

func getDeviceUUID(device string) (string, error) {
	output, err := cmd.RunDefaultWithStdoutBashC(fmt.Sprintf("blkid -s UUID -o value %s", device))
	if err != nil {
		return "", err
	}
	uuid := strings.TrimSpace(output)
	if uuid == "" {
		return "", fmt.Errorf("no UUID found for device %s", device)
	}
	return uuid, nil
}
