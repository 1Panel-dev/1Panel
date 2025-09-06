package dto

type LsblkDevice struct {
	Name       string        `json:"name"`
	Size       string        `json:"size"`
	Type       string        `json:"type"`
	MountPoint *string       `json:"mountpoint"`
	FsType     *string       `json:"fstype"`
	Model      *string       `json:"model"`
	Serial     string        `json:"serial"`
	Tran       string        `json:"tran"`
	Rota       bool          `json:"rota"`
	Children   []LsblkDevice `json:"children,omitempty"`
}

type LsblkOutput struct {
	BlockDevices []LsblkDevice `json:"blockdevices"`
}

type DiskFormatRequest struct {
	Device      string `json:"device" `
	Filesystem  string `json:"filesystem" `
	Label       string `json:"label,omitempty" `
	QuickFormat bool   `json:"quickFormat,omitempty"`
}
