package v1

import (
	"strconv"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// @Tags Container
// @Summary Page containers
// @Accept json
// @Param request body dto.PageContainer true "request"
// @Produce json
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/search [post]
func (b *BaseApi) SearchContainer(c *gin.Context) {
	var req dto.PageContainer
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := containerService.Page(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Container
// @Summary List containers
// @Accept json
// @Produce json
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/list [post]
func (b *BaseApi) ListContainer(c *gin.Context) {
	list, err := containerService.List()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags Container Compose
// @Summary Page composes
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/compose/search [post]
func (b *BaseApi) SearchCompose(c *gin.Context) {
	var req dto.SearchWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := containerService.PageCompose(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Container Compose
// @Summary Test compose
// @Accept json
// @Param request body dto.ComposeCreate true "request"
// @Success 200 {boolean} isOK
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/compose/test [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"检测 compose [name] 格式","formatEN":"check compose [name]"}
func (b *BaseApi) TestCompose(c *gin.Context) {
	var req dto.ComposeCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	isOK, err := containerService.TestCompose(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, isOK)
}

// @Tags Container Compose
// @Summary Create compose
// @Accept json
// @Param request body dto.ComposeCreate true "request"
// @Success 200 {string} log
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/compose [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建 compose [name]","formatEN":"create compose [name]"}
func (b *BaseApi) CreateCompose(c *gin.Context) {
	var req dto.ComposeCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	log, err := containerService.CreateCompose(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, log)
}

// @Tags Container Compose
// @Summary Operate compose
// @Accept json
// @Param request body dto.ComposeOperation true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/compose/operate [post]
// @x-panel-log {"bodyKeys":["name","operation"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"compose [operation] [name]","formatEN":"compose [operation] [name]"}
func (b *BaseApi) OperatorCompose(c *gin.Context) {
	var req dto.ComposeOperation
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := containerService.ComposeOperation(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container
// @Summary Update container
// @Accept json
// @Param request body dto.ContainerOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/update [post]
// @x-panel-log {"bodyKeys":["name","image"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新容器 [name][image]","formatEN":"update container [name][image]"}
func (b *BaseApi) ContainerUpdate(c *gin.Context) {
	var req dto.ContainerOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := containerService.ContainerUpdate(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container
// @Summary Load container info
// @Accept json
// @Param request body dto.OperationWithName true "request"
// @Success 200 {object} dto.ContainerOperate
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/info [post]
func (b *BaseApi) ContainerInfo(c *gin.Context) {
	var req dto.OperationWithName
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := containerService.ContainerInfo(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Summary Load container limits
// @Success 200 {object} dto.ResourceLimit
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/limit [get]
func (b *BaseApi) LoadResourceLimit(c *gin.Context) {
	data, err := containerService.LoadResourceLimit()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Summary Load container stats
// @Success 200 {array} dto.ContainerListStats
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/list/stats [get]
func (b *BaseApi) ContainerListStats(c *gin.Context) {
	data, err := containerService.ContainerListStats()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Container
// @Summary Create container
// @Accept json
// @Param request body dto.ContainerOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers [post]
// @x-panel-log {"bodyKeys":["name","image"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建容器 [name][image]","formatEN":"create container [name][image]"}
func (b *BaseApi) ContainerCreate(c *gin.Context) {
	var req dto.ContainerOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := containerService.ContainerCreate(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container
// @Summary Upgrade container
// @Accept json
// @Param request body dto.ContainerUpgrade true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/upgrade [post]
// @x-panel-log {"bodyKeys":["name","image"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新容器镜像 [name][image]","formatEN":"upgrade container image [name][image]"}
func (b *BaseApi) ContainerUpgrade(c *gin.Context) {
	var req dto.ContainerUpgrade
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := containerService.ContainerUpgrade(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container
// @Summary Clean container
// @Accept json
// @Param request body dto.ContainerPrune true "request"
// @Success 200 {object} dto.ContainerPruneReport
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/prune [post]
// @x-panel-log {"bodyKeys":["pruneType"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"清理容器 [pruneType]","formatEN":"clean container [pruneType]"}
func (b *BaseApi) ContainerPrune(c *gin.Context) {
	var req dto.ContainerPrune
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	report, err := containerService.Prune(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, report)
}

// @Tags Container
// @Summary Clean container log
// @Accept json
// @Param request body dto.OperationWithName true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/clean/log [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"清理容器 [name] 日志","formatEN":"clean container [name] logs"}
func (b *BaseApi) CleanContainerLog(c *gin.Context) {
	var req dto.OperationWithName
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := containerService.ContainerLogClean(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container
// @Summary Load container log
// @Accept json
// @Param request body dto.OperationWithNameAndType true "request"
// @Success 200 {string} content
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/load/log [post]
func (b *BaseApi) LoadContainerLog(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	content := containerService.LoadContainerLogs(req)
	helper.SuccessWithData(c, content)
}

// @Tags Container
// @Summary Rename Container
// @Accept json
// @Param request body dto.ContainerRename true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/rename [post]
// @x-panel-log {"bodyKeys":["name","newName"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"容器重命名 [name] => [newName]","formatEN":"rename container [name] => [newName]"}
func (b *BaseApi) ContainerRename(c *gin.Context) {
	var req dto.ContainerRename
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := containerService.ContainerRename(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container
// @Summary Commit Container
// @Accept json
// @Param request body dto.ContainerCommit true "request"
// @Success 200
// @Router /containers/commit [post]
func (b *BaseApi) ContainerCommit(c *gin.Context) {
	var req dto.ContainerCommit
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := containerService.ContainerCommit(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container
// @Summary Operate Container
// @Accept json
// @Param request body dto.ContainerOperation true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/operate [post]
// @x-panel-log {"bodyKeys":["names","operation"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"容器 [names] 执行 [operation]","formatEN":"container [operation] [names]"}
func (b *BaseApi) ContainerOperation(c *gin.Context) {
	var req dto.ContainerOperation
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := containerService.ContainerOperation(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container
// @Summary Container stats
// @Param id path integer true "container id"
// @Success 200 {object} dto.ContainerStats
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/stats/{id} [get]
func (b *BaseApi) ContainerStats(c *gin.Context) {
	containerID, ok := c.Params.Get("id")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error container id in path"))
		return
	}

	result, err := containerService.ContainerStats(containerID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags Container
// @Summary Container inspect
// @Accept json
// @Param request body dto.InspectReq true "request"
// @Success 200 {string} result
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/inspect [post]
func (b *BaseApi) Inspect(c *gin.Context) {
	var req dto.InspectReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := containerService.Inspect(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags Container
// @Summary Container logs
// @Param container query string false "container name"
// @Param since query string false "since"
// @Param follow query string false "follow"
// @Param tail query string false "tail"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/search/log [post]
func (b *BaseApi) ContainerLogs(c *gin.Context) {
	wsConn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.LOG.Errorf("gin context http handler failed, err: %v", err)
		return
	}
	defer wsConn.Close()

	container := c.Query("container")
	since := c.Query("since")
	follow := c.Query("follow") == "true"
	tail := c.Query("tail")

	if err := containerService.ContainerLogs(wsConn, "container", container, since, tail, follow); err != nil {
		_ = wsConn.WriteMessage(1, []byte(err.Error()))
		return
	}
}

// @Tags Container
// @Summary Download Container logs
// @Accept json
// @Param request body dto.ContainerLog true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/download/log [post]
func (b *BaseApi) DownloadContainerLogs(c *gin.Context) {
	var req dto.ContainerLog
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := containerService.DownloadContainerLogs(req.ContainerType, req.Container, req.Since, strconv.Itoa(int(req.Tail)), c)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
	}
}

// @Tags Container Network
// @Summary Page networks
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Produce json
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/network/search [post]
func (b *BaseApi) SearchNetwork(c *gin.Context) {
	var req dto.SearchWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := containerService.PageNetwork(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Container Network
// @Summary List networks
// @Accept json
// @Produce json
// @Success 200 {array} dto.Options
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/network [get]
func (b *BaseApi) ListNetwork(c *gin.Context) {
	list, err := containerService.ListNetwork()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags Container Network
// @Summary Delete network
// @Accept json
// @Param request body dto.BatchDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/network/del [post]
// @x-panel-log {"bodyKeys":["names"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"删除容器网络 [names]","formatEN":"delete container network [names]"}
func (b *BaseApi) DeleteNetwork(c *gin.Context) {
	var req dto.BatchDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := containerService.DeleteNetwork(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container Network
// @Summary Create network
// @Accept json
// @Param request body dto.NetworkCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/network [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建容器网络 name","formatEN":"create container network [name]"}
func (b *BaseApi) CreateNetwork(c *gin.Context) {
	var req dto.NetworkCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := containerService.CreateNetwork(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container Volume
// @Summary Page Container Volumes
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Produce json
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/volume/search [post]
func (b *BaseApi) SearchVolume(c *gin.Context) {
	var req dto.SearchWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := containerService.PageVolume(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Container Volume
// @Summary List Container Volumes
// @Accept json
// @Produce json
// @Success 200 {array} dto.Options
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/volume [get]
func (b *BaseApi) ListVolume(c *gin.Context) {
	list, err := containerService.ListVolume()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags Container Volume
// @Summary Delete Container Volume
// @Accept json
// @Param request body dto.BatchDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/volume/del [post]
// @x-panel-log {"bodyKeys":["names"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"删除容器存储卷 [names]","formatEN":"delete container volume [names]"}
func (b *BaseApi) DeleteVolume(c *gin.Context) {
	var req dto.BatchDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := containerService.DeleteVolume(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container Volume
// @Summary Create Container Volume
// @Accept json
// @Param request body dto.VolumeCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/volume [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建容器存储卷 [name]","formatEN":"create container volume [name]"}
func (b *BaseApi) CreateVolume(c *gin.Context) {
	var req dto.VolumeCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := containerService.CreateVolume(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container Compose
// @Summary Update Container Compose
// @Accept json
// @Param request body dto.ComposeUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/compose/update [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新 compose [name]","formatEN":"update compose information [name]"}
func (b *BaseApi) ComposeUpdate(c *gin.Context) {
	var req dto.ComposeUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := containerService.ComposeUpdate(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container Compose
// @Summary Container Compose logs
// @Param compose query string false "compose file address"
// @Param since query string false "date"
// @Param follow query string false "follow"
// @Param tail query string false "tail"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/compose/search/log [get]
func (b *BaseApi) ComposeLogs(c *gin.Context) {
	wsConn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.LOG.Errorf("gin context http handler failed, err: %v", err)
		return
	}
	defer wsConn.Close()

	compose := c.Query("compose")
	since := c.Query("since")
	follow := c.Query("follow") == "true"
	tail := c.Query("tail")

	if err := containerService.ContainerLogs(wsConn, "compose", compose, since, tail, follow); err != nil {
		_ = wsConn.WriteMessage(1, []byte(err.Error()))
		return
	}
}
