package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

// @Tags Disk Management
// @Summary Get complete disk information
// @Description Get information about all disks including partitioned and unpartitioned disks
// @Produce json
// @Success 200 {object} response.CompleteDiskInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /disks [get]
func (b *BaseApi) GetCompleteDiskInfo(c *gin.Context) {
	diskInfo, err := diskService.GetCompleteDiskInfo()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, diskInfo)
}

// @Tags Disk Management
// @Summary Partition disk
// @Description Create partition and format disk with specified filesystem
// @Accept json
// @Param request body request.DiskPartitionRequest true "partition request"
// @Success 200 {string} string "Partition created successfully"
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /disks/partition [post]
// @x-panel-log {"bodyKeys":["device", "filesystem", "mountPoint"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"对磁盘 [device] 进行分区，文件系统 [filesystem]，挂载点 [mountPoint]","formatEN":"Partition disk [device] with filesystem [filesystem], mount point [mountPoint]"}
func (b *BaseApi) PartitionDisk(c *gin.Context) {
	var req request.DiskPartitionRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := diskService.PartitionDisk(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Disk Management
// @Summary Mount disk
// @Description Mount partition to specified mount point
// @Accept json
// @Param request body request.DiskMountRequest true "mount request"
// @Success 200 {string} string "Disk mounted successfully"
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /disks/mount [post]
// @x-panel-log {"bodyKeys":["device", "mountPoint"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"挂载磁盘 [device] 到 [mountPoint]","formatEN":"Mount disk [device] to [mountPoint]"}
func (b *BaseApi) MountDisk(c *gin.Context) {
	var req request.DiskMountRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err := diskService.MountDisk(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.Success(c)
}

// @Tags Disk Management
// @Summary Unmount disk
// @Description Unmount partition from mount point
// @Accept json
// @Param request body request.DiskUnmountRequest true "unmount request"
// @Success 200 {string} string "Disk unmounted successfully"
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /disks/unmount [post]
// @x-panel-log {"bodyKeys":["device", "mountPoint"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"卸载磁盘 [device] 从 [mountPoint]","formatEN":"Unmount disk [device] from [mountPoint]"}
func (b *BaseApi) UnmountDisk(c *gin.Context) {
	var req request.DiskUnmountRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err := diskService.UnmountDisk(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.Success(c)
}
