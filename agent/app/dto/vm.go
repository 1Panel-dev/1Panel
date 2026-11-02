package dto

import (
	"time"
)

type VMSearchResult struct {
	PageResult
	Limit int `json:"limit"` // Maximum VM count for the target node; 0 means unlimited.
}

type VMCheckItem struct {
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	Message  string `json:"message"`
	IsEnable bool   `json:"isEnable"`
	IsExist  bool   `json:"isExist"`
	IsActive bool   `json:"isActive"`
}

type VMEnvironment struct {
	Arch           string `json:"arch"`
	ArchSupported  bool   `json:"archSupported"`
	Ready          bool   `json:"ready"`
	HelperExist    bool   `json:"helperExist"`
	HelperMessage  string `json:"helperMessage"`
	KVMExist       bool   `json:"kvmExist"`
	KVMMessage     string `json:"kvmMessage"`
	LibvirtExist   bool   `json:"libvirtExist"`
	LibvirtActive  bool   `json:"libvirtActive"`
	LibvirtMessage string `json:"libvirtMessage"`
	StorageExist   bool   `json:"storageExist"`
	StorageMessage string `json:"storageMessage"`
}

type VMHealthCheckItem struct {
	Level      string            `json:"level"`
	Check      string            `json:"check"`
	Reason     string            `json:"reason,omitempty"`
	Params     map[string]string `json:"params,omitempty"`
	Detail     string            `json:"detail,omitempty"`
	FixCommand string            `json:"fixCommand,omitempty"`
}

type VMHealthCheckResult struct {
	Pass  int                 `json:"pass"`
	Warn  int                 `json:"warn"`
	Fail  int                 `json:"fail"`
	Items []VMHealthCheckItem `json:"items"`
}

type VMCreate struct {
	TaskID           string   `json:"taskID"`
	Name             string   `json:"name" validate:"required,vm_name"`
	CPU              int      `json:"cpu" validate:"required,min=1"`
	Memory           int64    `json:"memory" validate:"required,min=1"`
	DiskSize         int64    `json:"diskSize" validate:"required,min=1"`
	OSType           string   `json:"osType" validate:"omitempty,oneof=windows linux other"`
	DiskBus          string   `json:"diskBus" validate:"omitempty,oneof=sata virtio"`
	TemplateID       uint     `json:"templateID"`
	DiskPath         string   `json:"diskPath"`
	StoragePool      string   `json:"storagePool"`
	ISOPaths         []string `json:"isoPaths"`
	ImagePath        string   `json:"imagePath"`
	Network          string   `json:"network" validate:"required,vm_common"`
	BootOrder        string   `json:"bootOrder"`
	StartAfterCreate bool     `json:"startAfterCreate"`
	AutoStart        bool     `json:"autoStart"`
	Description      string   `json:"description"`
}

type VMTemplateCreate struct {
	TaskID       string `json:"taskID"`
	Name         string `json:"name" validate:"required,vm_name"`
	TemplateName string `json:"templateName" validate:"required,vm_name"`
	StoragePool  string `json:"storagePool"`
	Description  string `json:"description"`
}

type VMTemplateDelete struct {
	ID         uint   `json:"id" validate:"required"`
	DeleteDisk bool   `json:"deleteDisk"`
	TaskID     string `json:"taskID"`
}

type VMTemplateInfo struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	SourceVM    string    `json:"sourceVM"`
	CPU         int       `json:"cpu"`
	Memory      int64     `json:"memory"`
	DiskSize    int64     `json:"diskSize"`
	DiskBus     string    `json:"diskBus"`
	DiskPath    string    `json:"diskPath"`
	StoragePool string    `json:"storagePool"`
	Network     string    `json:"network"`
	BootOrder   string    `json:"bootOrder"`
	IsExist     bool      `json:"isExist"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"createdAt"`
	Description string    `json:"description"`
}

type VMUpdate struct {
	TaskID      string   `json:"taskID"`
	Name        string   `json:"name" validate:"required,vm_name"`
	CPU         int      `json:"cpu" validate:"required,min=1"`
	Memory      int64    `json:"memory" validate:"required,min=1"`
	ISOPaths    []string `json:"isoPaths"`
	Network     string   `json:"network" validate:"required,vm_common"`
	BootOrder   string   `json:"bootOrder"`
	AutoStart   bool     `json:"autoStart"`
	Description string   `json:"description"`
}

type VMDiskResize struct {
	TaskID  string `json:"taskID"`
	Name    string `json:"name" validate:"required,vm_name"`
	Target  string `json:"target" validate:"required,vm_common"`
	SizeGiB int64  `json:"sizeGiB" validate:"required,min=1,max=8388607"`
}

type VMRename struct {
	TaskID  string `json:"taskID"`
	Name    string `json:"name" validate:"required,vm_name"`
	NewName string `json:"newName" validate:"required,vm_name"`
}

type VMDelete struct {
	TaskID string `json:"taskID"`
	ID     uint   `json:"id" validate:"required"`
	Force  bool   `json:"force"`
}

type VMOperate struct {
	Name    string `json:"name" validate:"required,vm_name"`
	Operate string `json:"operate" validate:"required,oneof=start shutdown reboot destroy suspend resume"`
}

type VMInfoReq struct {
	Name       string `json:"name" validate:"required,vm_name"`
	CheckDisks bool   `json:"checkDisks"`
}

type VMReq struct {
	Name string `json:"name" validate:"required,vm_name"`
}

type VMTaskResp struct {
	TaskID string `json:"taskID"`
}

type VMOrphanItem struct {
	Type       string `json:"type" validate:"required,oneof=orphan_domain_disk orphan_snapshot partial_file"`
	Path       string `json:"path" validate:"required"`
	Size       int64  `json:"size"`
	ModifiedAt int64  `json:"modifiedAt"`
}

type VMOrphanClean struct {
	Items []VMOrphanItem `json:"items" validate:"required,min=1,dive"`
}

type VMSnapshotReq struct {
	TaskID       string `json:"taskID"`
	Name         string `json:"name" validate:"required,vm_name"`
	SnapshotName string `json:"snapshotName" validate:"required,vm_name"`
	Description  string `json:"description"`
}

type VMSnapshotSearch struct {
	Name string `json:"name" validate:"required,vm_name"`
}

type VMSnapshotInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"createdAt"`
	State       string `json:"state"`
	Parent      string `json:"parent"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	Current     bool   `json:"current"`
	Children    int    `json:"children"`
}

type VMIsoSearch struct {
	PageInfo
	Type string `json:"type" validate:"required"`
	Info string `json:"info"`
}

type VMIsoCreate struct {
	Type        string `json:"type" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Path        string `json:"path" validate:"required"`
	Description string `json:"description"`
}

type VMIsoUpdate struct {
	ID          uint   `json:"id" validate:"required"`
	Type        string `json:"type" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Path        string `json:"path" validate:"required"`
	Description string `json:"description"`
}

type VMIsoImportStart struct {
	Type         string `json:"type" validate:"required"`
	Name         string `json:"name" validate:"required"`
	Path         string `json:"path" validate:"required"`
	Status       string `json:"status" validate:"required,oneof=uploading downloading"`
	TempFileID   string `json:"tempFileID"`
	TempFileType string `json:"tempFileType" validate:"omitempty,oneof=upload download"`
	Description  string `json:"description"`
}

type VMIsoImportUpdate struct {
	ID           uint   `json:"id" validate:"required"`
	Status       string `json:"status" validate:"required,oneof=uploading downloading ready failed"`
	Error        string `json:"error"`
	TempFileID   string `json:"tempFileID"`
	TempFileType string `json:"tempFileType" validate:"omitempty,oneof=upload download"`
}

type VMIsoDelete struct {
	ID         uint `json:"id" validate:"required"`
	DeleteFile bool `json:"deleteFile"`
}

type VMIsoInfo struct {
	ID           uint      `json:"id"`
	Type         string    `json:"type"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	IsExist      bool      `json:"isExist"`
	Size         int64     `json:"size"`
	Status       string    `json:"status"`
	Error        string    `json:"error"`
	TempFileID   string    `json:"tempFileID"`
	TempFileType string    `json:"tempFileType"`
	CreatedAt    time.Time `json:"createdAt"`
	Description  string    `json:"description"`
}

type VMResourceInfo struct {
	Type            string `json:"type"`
	Name            string `json:"name"`
	SubType         string `json:"subType"`
	BridgeName      string `json:"bridgeName"`
	SourceInterface string `json:"sourceInterface"`
	CIDR            string `json:"cidr"`
	Gateway         string `json:"gateway"`
	DHCPStart       string `json:"dhcpStart"`
	DHCPEnd         string `json:"dhcpEnd"`
	Path            string `json:"path"`
	Status          string `json:"status"`
	AutoStart       bool   `json:"autoStart"`
	Capacity        uint64 `json:"capacity"`
	Available       uint64 `json:"available"`
}

type VMResourceDelete struct {
	ID    uint `json:"id" validate:"required"`
	Force bool `json:"force"`
}

type VMNetworkCreate struct {
	Name            string `json:"name" validate:"required,vm_common"`
	Type            string `json:"type" validate:"required,oneof=nat bridge macvtap"`
	BridgeName      string `json:"bridgeName"`
	SourceInterface string `json:"sourceInterface" validate:"required_if=Type macvtap"`
	CIDR            string `json:"cidr"`
	Gateway         string `json:"gateway"`
	DHCPStart       string `json:"dhcpStart"`
	DHCPEnd         string `json:"dhcpEnd"`
	AutoStart       bool   `json:"autoStart"`
	Description     string `json:"description"`
}

type VMNetworkUpdate struct {
	ID          uint    `json:"id" validate:"required"`
	AutoStart   *bool   `json:"autoStart"`
	Status      *string `json:"status"`
	Description *string `json:"description"`
}

type VMNetworkItem struct {
	ID              uint               `json:"id"`
	Name            string             `json:"name"`
	Type            string             `json:"type"`
	BridgeName      string             `json:"bridgeName"`
	SourceInterface string             `json:"sourceInterface"`
	CIDR            string             `json:"cidr"`
	Gateway         string             `json:"gateway"`
	DHCPStart       string             `json:"dhcpStart"`
	DHCPEnd         string             `json:"dhcpEnd"`
	AutoStart       bool               `json:"autoStart"`
	Description     string             `json:"description"`
	CreatedAt       string             `json:"createdAt"`
	Status          string             `json:"status"`
	VMCount         int                `json:"vmCount"`
	VMs             []VMResourceRef    `json:"vms"`
	Registered      bool               `json:"registered"`
	UsedIPs         []VMNetworkIPUsage `json:"usedIPs"`
	Error           string             `json:"error"`
}

type VMResourceRef struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type VMNetworkIPUsage struct {
	IP     string `json:"ip"`
	VMName string `json:"vmName"`
	MAC    string `json:"mac"`
	Source string `json:"source"`
}

type VMNetworkInterfaceItem struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
}

type VMNetworkBridgeItem struct {
	Name        string `json:"name"`
	Recommended bool   `json:"recommended"`
	Description string `json:"description"`
}

type VMStorageCreate struct {
	Name        string `json:"name" validate:"required,vm_common"`
	Type        string `json:"type" validate:"required,oneof=dir"`
	Path        string `json:"path" validate:"required"`
	AutoStart   bool   `json:"autoStart"`
	Description string `json:"description"`
}

type VMStorageUpdate struct {
	ID          uint    `json:"id" validate:"required"`
	Name        string  `json:"name" validate:"omitempty,vm_common"`
	Path        string  `json:"path"`
	AutoStart   *bool   `json:"autoStart"`
	Status      *string `json:"status"`
	Description *string `json:"description"`
}

type VMStorageItem struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Path        string          `json:"path"`
	AutoStart   bool            `json:"autoStart"`
	Description string          `json:"description"`
	CreatedAt   string          `json:"createdAt"`
	Status      string          `json:"status"`
	Capacity    uint64          `json:"capacity"`
	Available   uint64          `json:"available"`
	VMCount     int             `json:"vmCount"`
	VMs         []VMResourceRef `json:"vms"`
	Error       string          `json:"error"`
}

type VMInfo struct {
	ID           uint          `json:"id"`
	Name         string        `json:"name"`
	UUID         string        `json:"uuid"`
	CPU          int           `json:"cpu"`
	Memory       int64         `json:"memory"`
	DiskSize     int64         `json:"diskSize"`
	DiskBus      string        `json:"diskBus"`
	DiskPath     string        `json:"diskPath"`
	StoragePool  string        `json:"storagePool"`
	ISOPaths     []string      `json:"isoPaths"`
	ImagePath    string        `json:"imagePath"`
	Network      string        `json:"network"`
	BootOrder    string        `json:"bootOrder"`
	AutoStart    bool          `json:"autoStart"`
	Status       string        `json:"status"`
	Disks        []VMDisk      `json:"disks"`
	Interfaces   []VMInterface `json:"interfaces"`
	Graphics     *VMGraphics   `json:"graphics"`
	TaskID       string        `json:"taskID"`
	CreateTaskID string        `json:"createTaskID"`
	Description  string        `json:"description"`
}

type VMDisk struct {
	Capacity      int64  `json:"capacity"`
	Format        string `json:"format"`
	ReadOnly      bool   `json:"readOnly"`
	Shareable     bool   `json:"shareable"`
	ResizeBlocked string `json:"resizeBlocked"`
	Device        string `json:"device"`
	Type          string `json:"type"`
	Source        string `json:"source"`
	Target        string `json:"target"`
	Bus           string `json:"bus"`
}

type VMInterface struct {
	Type   string `json:"type"`
	Source string `json:"source"`
	Model  string `json:"model"`
	MAC    string `json:"mac"`
	IP     string `json:"ip"`
}

type VMGraphics struct {
	Type   string `json:"type"`
	Port   string `json:"port"`
	Listen string `json:"listen"`
}

type VMStats struct {
	Name      string         `json:"name"`
	Status    string         `json:"status"`
	Timestamp int64          `json:"timestamp"`
	CPUTime   uint64         `json:"cpuTime"`
	VCPU      int            `json:"vcpu"`
	Memory    VMMemoryStats  `json:"memory"`
	Disk      VMDiskStats    `json:"disk"`
	Network   VMNetworkStats `json:"network"`
}

type VMMemoryStats struct {
	AvailableKiB uint64 `json:"availableKiB"`
	UnusedKiB    uint64 `json:"unusedKiB"`
	UsableKiB    uint64 `json:"usableKiB"`
	RSSKiB       uint64 `json:"rssKiB"`
	ActualKiB    uint64 `json:"actualKiB"`
}

type VMDiskStats struct {
	ReadBytes  uint64 `json:"readBytes"`
	WriteBytes uint64 `json:"writeBytes"`
	ReadReqs   uint64 `json:"readReqs"`
	WriteReqs  uint64 `json:"writeReqs"`
}

type VMNetworkStats struct {
	RxBytes   uint64 `json:"rxBytes"`
	TxBytes   uint64 `json:"txBytes"`
	RxPackets uint64 `json:"rxPackets"`
	TxPackets uint64 `json:"txPackets"`
}
