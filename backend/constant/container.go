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

	DaemonJsonDir      = "/System/Volumes/Data/Users/slooop/.docker/daemon.json"
	TmpDockerBuildDir  = "/opt/1Panel/data/docker/build"
	TmpComposeBuildDir = "/opt/1Panel/data/docker/compose"
)
