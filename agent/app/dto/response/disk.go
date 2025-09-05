package response

type DiskInfo struct {
	DiskBasicInfo
	Partitions []DiskBasicInfo `json:"partitions"`
}

type DiskBasicInfo struct {
	Device      string `json:"device"`
	Size        string `json:"size"`
	Model       string `json:"model"`
	DiskType    string `json:"diskType"`
	IsRemovable bool   `json:"isRemovable"`
	IsSystem    bool   `json:"isSystem"`
	Filesystem  string `json:"filesystem"`
	Used        string `json:"used"`
	Avail       string `json:"avail"`
	UsePercent  int    `json:"usePercent"`
	MountPoint  string `json:"mountPoint"`
	IsMounted   bool   `json:"isMounted"`
	Serial      string `json:"serial"`
}

type CompleteDiskInfo struct {
	Disks              []DiskInfo      `json:"disks"`
	UnpartitionedDisks []DiskBasicInfo `json:"unpartitionedDisks"`
	SystemDisk         *DiskInfo       `json:"systemDisk"`
	TotalDisks         int             `json:"totalDisks"`
	TotalCapacity      int64           `json:"totalCapacity"`
}

type MountInfo struct {
	Device     string `json:"device"`
	MountPoint string `json:"mountPoint"`
	Filesystem string `json:"filesystem"`
	Options    string `json:"options"`
}
