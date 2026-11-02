package protocol

// VM helper protocol definition shared by the core caller and helper binary.

const (
	Version      = "1.2.0"
	API          = "1.2.0"
	MinVersion   = "1.2.0"
	DefaultURI   = "qemu:///system"
	HelperBinary = "1panel-kvm"
)

const (
	ActionVersion            = "version"
	ActionCheck              = "check"
	ActionHealthCheck        = "health-check"
	ActionList               = "list"
	ActionGet                = "get"
	ActionCreate             = "create"
	ActionUpdate             = "update"
	ActionResizeDisk         = "resizeDisk"
	ActionRename             = "rename"
	ActionDelete             = "delete"
	ActionStart              = "start"
	ActionShutdown           = "shutdown"
	ActionReboot             = "reboot"
	ActionDestroy            = "destroy"
	ActionSuspend            = "suspend"
	ActionResume             = "resume"
	ActionConsole            = "console"
	ActionStats              = "stats"
	ActionListSnapshots      = "listSnapshots"
	ActionCreateSnapshot     = "createSnapshot"
	ActionRecoverSnapshot    = "recoverSnapshot"
	ActionDeleteSnapshot     = "deleteSnapshot"
	ActionCreateTemplate     = "createTemplate"
	ActionCreateFromTemplate = "createFromTemplate"
	ActionResources          = "resources"
	ActionCreateResource     = "createResource"
	ActionUpdateResource     = "updateResource"
	ActionDeleteResource     = "deleteResource"
	ActionScanOrphans        = "scanOrphans"
	ActionCleanOrphans       = "cleanOrphans"
)

type Request struct {
	Version string `json:"version"`
	Action  string `json:"action"`
	URI     string `json:"uri,omitempty"`
	Payload any    `json:"payload,omitempty"`
}

type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Detail != "" {
		return e.Message + ": " + e.Detail
	}
	return e.Message
}

type VersionInfo struct {
	Version string `json:"version"`
	API     string `json:"api"`
}

type VMConfig struct {
	Name             string   `json:"name"`
	CPU              int      `json:"cpu"`
	MemoryMiB        int64    `json:"memoryMiB"`
	DiskSizeGiB      int64    `json:"diskSizeGiB"`
	DiskBus          string   `json:"diskBus,omitempty"`
	DiskPath         string   `json:"diskPath,omitempty"`
	StoragePool      string   `json:"storagePool,omitempty"`
	ISOPaths         []string `json:"isoPaths"`
	ImagePath        string   `json:"imagePath,omitempty"`
	Network          string   `json:"network"`
	BootOrder        string   `json:"bootOrder,omitempty"`
	StartAfterCreate bool     `json:"startAfterCreate"`
	AutoStart        bool     `json:"autoStart"`
}

type VMInfoReq struct {
	Name       string `json:"name"`
	CheckDisks bool   `json:"checkDisks,omitempty"`
}

type VMName struct {
	Name string `json:"name"`
}

type VMDiskResize struct {
	Name    string `json:"name"`
	Target  string `json:"target"`
	SizeGiB int64  `json:"sizeGiB"`
}

type VMRename struct {
	Name    string `json:"name"`
	NewName string `json:"newName"`
}

type VMDelete struct {
	Name           string   `json:"name"`
	Force          bool     `json:"force"`
	DeleteDisk     bool     `json:"deleteDisk"`
	ProtectedPaths []string `json:"protectedPaths,omitempty"`
}

type VMSnapshotReq struct {
	Name         string `json:"name"`
	SnapshotName string `json:"snapshotName,omitempty"`
	Description  string `json:"description,omitempty"`
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

type VMTemplateCreate struct {
	Name        string `json:"name"`
	SourceName  string `json:"sourceName"`
	DiskPath    string `json:"diskPath,omitempty"`
	StoragePool string `json:"storagePool,omitempty"`
}

type VMOrphanScan struct {
	ProtectedPaths []string `json:"protectedPaths,omitempty"`
}

type VMOrphanItem struct {
	Type       string `json:"type"`
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	ModifiedAt int64  `json:"modifiedAt"`
}

type VMOrphanClean struct {
	Items          []VMOrphanItem `json:"items"`
	ProtectedPaths []string       `json:"protectedPaths,omitempty"`
}

type VMCheckItem struct {
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	Message  string `json:"message"`
	IsEnable bool   `json:"isEnable"`
	IsExist  bool   `json:"isExist"`
	IsActive bool   `json:"isActive"`
	Version  string `json:"version"`
}

const (
	HealthCheckLevelPass = "PASS"
	HealthCheckLevelInfo = "INFO"
	HealthCheckLevelWarn = "WARN"
	HealthCheckLevelFail = "FAIL"
)

type VMHealthCheckRequest struct {
	Name string `json:"name"`
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

type VMResourceSearch struct {
	Type string `json:"type"`
}

type VMResourceInfo struct {
	Type            string `json:"type"`
	Name            string `json:"name"`
	SubType         string `json:"subType"`
	BridgeName      string `json:"bridgeName,omitempty"`
	SourceInterface string `json:"sourceInterface,omitempty"`
	CIDR            string `json:"cidr,omitempty"`
	Gateway         string `json:"gateway,omitempty"`
	DHCPStart       string `json:"dhcpStart,omitempty"`
	DHCPEnd         string `json:"dhcpEnd,omitempty"`
	Path            string `json:"path,omitempty"`
	Status          string `json:"status"`
	AutoStart       bool   `json:"autoStart"`
	Capacity        uint64 `json:"capacity,omitempty"`
	Available       uint64 `json:"available,omitempty"`
}

type VMResourceCreate struct {
	Type            string `json:"type"`
	Name            string `json:"name"`
	SubType         string `json:"subType"`
	BridgeName      string `json:"bridgeName,omitempty"`
	SourceInterface string `json:"sourceInterface,omitempty"`
	CIDR            string `json:"cidr,omitempty"`
	Gateway         string `json:"gateway,omitempty"`
	DHCPStart       string `json:"dhcpStart,omitempty"`
	DHCPEnd         string `json:"dhcpEnd,omitempty"`
	Path            string `json:"path,omitempty"`
	AutoStart       bool   `json:"autoStart"`
}

type VMResourceUpdate struct {
	Type      string  `json:"type"`
	Name      string  `json:"name"`
	AutoStart *bool   `json:"autoStart,omitempty"`
	Status    *string `json:"status,omitempty"`
}

type VMResourceDelete struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type VMInfo struct {
	Name        string        `json:"name"`
	UUID        string        `json:"uuid"`
	CPU         int           `json:"cpu"`
	Memory      int64         `json:"memory"`
	DiskSize    int64         `json:"diskSize"`
	DiskBus     string        `json:"diskBus"`
	DiskPath    string        `json:"diskPath"`
	StoragePool string        `json:"storagePool"`
	ISOPaths    []string      `json:"isoPaths"`
	ImagePath   string        `json:"imagePath"`
	Network     string        `json:"network"`
	BootOrder   string        `json:"bootOrder"`
	AutoStart   bool          `json:"autoStart"`
	Status      string        `json:"status"`
	Disks       []VMDisk      `json:"disks"`
	Interfaces  []VMInterface `json:"interfaces"`
	Graphics    *VMGraphics   `json:"graphics"`
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

func OK(data any) Response {
	return Response{Success: true, Data: data}
}

func Fail(code string, message string, detail string) Response {
	return Response{Success: false, Error: &Error{Code: code, Message: message, Detail: detail}}
}
