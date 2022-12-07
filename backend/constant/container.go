package constant

const (
	ContainerOpStart   = "start"
	ContainerOpStop    = "stop"
	ContainerOpRestart = "restart"
	ContainerOpKill    = "kill"
	ContainerOpPause   = "pause"
	ContainerOpUnpause = "unpause"
	ContainerOpRename  = "rename"
	ContainerOpRemove  = "remove"

	ComposeOpStop    = "stop"
	ComposeOpRestart = "restart"
	ComposeOpRemove  = "remove"

	TmpDockerBuildDir  = "/opt/1Panel/data/docker/build"
	TmpComposeBuildDir = "/opt/1Panel/data/docker/compose"
)
