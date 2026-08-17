package firewall

type BaseOperation string

const (
	BaseOperationInit            BaseOperation = "init-base"
	BaseOperationBind            BaseOperation = "bind-base"
	BaseOperationBindWithoutInit BaseOperation = "bind-base-without-init"
	BaseOperationUnbind          BaseOperation = "unbind-base"
)
