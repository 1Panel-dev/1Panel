package types

type VMConfig struct {
	Name             string
	CPU              int
	MemoryMiB        int64
	DiskSizeGiB      int64
	DiskBus          string
	DiskPath         string
	StoragePool      string
	ISOPaths         []string
	ImagePath        string
	Network          string
	BootOrder        string
	StartAfterCreate bool
	AutoStart        bool
}

type VMTemplateCreate struct {
	Name        string
	SourceName  string
	DiskPath    string
	StoragePool string
}

type VMResourceCreate struct {
	Type            string
	Name            string
	SubType         string
	BridgeName      string
	SourceInterface string
	CIDR            string
	Gateway         string
	DHCPStart       string
	DHCPEnd         string
	Path            string
	AutoStart       bool
}

type VMResourceUpdate struct {
	Type      string
	Name      string
	AutoStart *bool
	Status    *string
}
