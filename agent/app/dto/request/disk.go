package request

type DiskPartitionRequest struct {
	Device     string `json:"device" validate:"required"`
	Filesystem string `json:"filesystem" validate:"required,oneof=ext4 xfs"`
	Label      string `json:"label"`
	AutoMount  bool   `json:"autoMount"`
	MountPoint string `json:"mountPoint" validate:"required"`
}

type DiskMountRequest struct {
	Device     string `json:"device" validate:"required"`
	MountPoint string `json:"mountPoint" validate:"required"`
	Filesystem string `json:"filesystem" validate:"required,oneof=ext4 xfs"`
	AutoMount  bool   `json:"autoMount"`
}

type DiskUnmountRequest struct {
	MountPoint string `json:"mountPoint" validate:"required"`
}
