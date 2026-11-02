package v2

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// @Tags VirtualMachine
// @Summary Check virtual machine environment
// @Accept json
// @Produce json
// @Success 200 {object} dto.VMEnvironment
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/environment [get]
func (b *BaseApi) CheckVMEnvironment(c *gin.Context) {
	data, err := vmService.CheckEnvironment()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Enable virtual machine environment
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/environment/enable [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"启用虚拟机运行环境","formatEN":"enable VM runtime environment"}
func (b *BaseApi) EnableVMEnvironment(c *gin.Context) {
	if err := vmService.EnableEnvironment(); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags VirtualMachine
// @Summary Page virtual machines
// @Description Returns the VM list and the target node quota: 1 for Community, 5 for Professional, and 0 for unlimited Enterprise.
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Produce json
// @Success 200 {object} dto.VMSearchResult{items=[]dto.VMInfo}
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/search [post]
func (b *BaseApi) SearchVM(c *gin.Context) {
	var req dto.SearchWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, list, err := vmService.Search(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.VMSearchResult{
		PageResult: dto.PageResult{Items: list, Total: total},
		Limit:      xpack.VMCountLimit(c),
	})
}

// @Tags VirtualMachine
// @Summary Sync virtual machines
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/sync [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"同步虚拟机","formatEN":"sync virtual machines"}
func (b *BaseApi) SyncVMs(c *gin.Context) {
	if err := vmService.SyncVMs(); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags VirtualMachine
// @Summary Load virtual machine details
// @Accept json
// @Param request body dto.VMInfoReq true "request"
// @Produce json
// @Success 200 {object} dto.VMInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/info [post]
func (b *BaseApi) GetVM(c *gin.Context) {
	var req dto.VMInfoReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.Get(req.Name, req.CheckDisks)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Expand virtual machine disk
// @Description Returns taskID for tracking the asynchronous operation.
// @Accept json
// @Param request body dto.VMDiskResize true "request"
// @Produce json
// @Success 200 {object} dto.VMTaskResp
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/disks/resize [post]
// @x-panel-log {"bodyKeys":["name","target","sizeGiB"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"扩容虚拟机 [name] 磁盘 [target] 至 [sizeGiB] GiB","formatEN":"expand VM [name] disk [target] to [sizeGiB] GiB"}
func (b *BaseApi) ResizeVMDisk(c *gin.Context) {
	var req dto.VMDiskResize
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.ResizeDisk(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Check virtual machine health
// @Accept json
// @Param request body dto.VMReq true "request"
// @Produce json
// @Success 200 {object} dto.VMHealthCheckResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/health/check [post]
func (b *BaseApi) HealthCheckVM(c *gin.Context) {
	var req dto.VMReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.HealthCheck(c.Request.Context(), req.Name)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Load virtual machine statistics
// @Accept json
// @Param request body dto.VMReq true "request"
// @Produce json
// @Success 200 {object} dto.VMStats
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/stats [post]
func (b *BaseApi) GetVMStats(c *gin.Context) {
	var req dto.VMReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.GetStats(req.Name)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Create virtual machine
// @Description Creates a VM from ISO images or a template. The target node allows 1 VM in Community, 5 in Professional, and unlimited VMs in Enterprise. All VM records count toward the quota, regardless of state; creation fails with ErrVMCountLimit when the quota is reached.
// @Description Returns taskID for tracking the asynchronous operation.
// @Accept json
// @Param request body dto.VMCreate true "request"
// @Produce json
// @Success 200 {object} dto.VMTaskResp
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建虚拟机 [name]","formatEN":"create VM [name]"}
func (b *BaseApi) CreateVM(c *gin.Context) {
	var req dto.VMCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.Create(req, xpack.VMCountLimit(c))
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary List virtual machine templates
// @Accept json
// @Produce json
// @Success 200 {array} dto.VMTemplateInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/templates/list [post]
func (b *BaseApi) ListVMTemplates(c *gin.Context) {
	data, err := vmService.ListTemplates()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Create virtual machine template
// @Description Returns taskID for tracking the asynchronous operation.
// @Accept json
// @Param request body dto.VMTemplateCreate true "request"
// @Produce json
// @Success 200 {object} dto.VMTaskResp
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/templates [post]
// @x-panel-log {"bodyKeys":["name","templateName"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建虚拟机 [name] 模板 [templateName]","formatEN":"create VM [name] template [templateName]"}
func (b *BaseApi) CreateVMTemplate(c *gin.Context) {
	var req dto.VMTemplateCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.CreateTemplate(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Delete virtual machine template
// @Description Returns taskID for tracking the asynchronous operation.
// @Accept json
// @Param request body dto.VMTemplateDelete true "request"
// @Produce json
// @Success 200 {object} dto.VMTaskResp
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/templates/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"vm_templates","output_column":"name","output_value":"name"}],"formatZH":"删除虚拟机模板 [name]","formatEN":"delete VM template [name]"}
func (b *BaseApi) DeleteVMTemplate(c *gin.Context) {
	var req dto.VMTemplateDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.DeleteTemplate(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Update virtual machine
// @Description Returns taskID for tracking the asynchronous operation.
// @Accept json
// @Param request body dto.VMUpdate true "request"
// @Produce json
// @Success 200 {object} dto.VMTaskResp
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/update [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新虚拟机 [name]","formatEN":"update VM [name]"}
func (b *BaseApi) UpdateVM(c *gin.Context) {
	var req dto.VMUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.Update(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Rename virtual machine
// @Description Returns taskID for tracking the asynchronous operation.
// @Accept json
// @Param request body dto.VMRename true "request"
// @Produce json
// @Success 200 {object} dto.VMTaskResp
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/rename [post]
// @x-panel-log {"bodyKeys":["name","newName"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"重命名虚拟机 [name] 为 [newName]","formatEN":"rename VM [name] to [newName]"}
func (b *BaseApi) RenameVM(c *gin.Context) {
	var req dto.VMRename
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.Rename(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Delete virtual machine
// @Description Returns taskID for tracking the asynchronous operation.
// @Accept json
// @Param request body dto.VMDelete true "request"
// @Produce json
// @Success 200 {object} dto.VMTaskResp
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"virtual_machines","output_column":"name","output_value":"name"}],"formatZH":"删除虚拟机 [name]","formatEN":"delete VM [name]"}
func (b *BaseApi) DeleteVM(c *gin.Context) {
	var req dto.VMDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.Delete(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary List virtual machine snapshots
// @Accept json
// @Param request body dto.VMSnapshotSearch true "request"
// @Produce json
// @Success 200 {array} dto.VMSnapshotInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/snapshots/list [post]
func (b *BaseApi) ListVMSnapshots(c *gin.Context) {
	var req dto.VMSnapshotSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.ListSnapshots(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Create virtual machine snapshot
// @Description Returns taskID for tracking the asynchronous operation.
// @Accept json
// @Param request body dto.VMSnapshotReq true "request"
// @Produce json
// @Success 200 {object} dto.VMTaskResp
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/snapshots [post]
// @x-panel-log {"bodyKeys":["name","snapshotName"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建虚拟机 [name] 快照 [snapshotName]","formatEN":"create VM [name] snapshot [snapshotName]"}
func (b *BaseApi) CreateVMSnapshot(c *gin.Context) {
	var req dto.VMSnapshotReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.CreateSnapshot(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Recover virtual machine snapshot
// @Description Returns taskID for tracking the asynchronous operation.
// @Accept json
// @Param request body dto.VMSnapshotReq true "request"
// @Produce json
// @Success 200 {object} dto.VMTaskResp
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/snapshots/recover [post]
// @x-panel-log {"bodyKeys":["name","snapshotName"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"恢复虚拟机 [name] 快照 [snapshotName]","formatEN":"recover VM [name] snapshot [snapshotName]"}
func (b *BaseApi) RecoverVMSnapshot(c *gin.Context) {
	var req dto.VMSnapshotReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.RecoverSnapshot(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Delete virtual machine snapshot
// @Description Returns taskID for tracking the asynchronous operation.
// @Accept json
// @Param request body dto.VMSnapshotReq true "request"
// @Produce json
// @Success 200 {object} dto.VMTaskResp
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/snapshots/del [post]
// @x-panel-log {"bodyKeys":["name","snapshotName"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"删除虚拟机 [name] 快照 [snapshotName]","formatEN":"delete VM [name] snapshot [snapshotName]"}
func (b *BaseApi) DeleteVMSnapshot(c *gin.Context) {
	var req dto.VMSnapshotReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.DeleteSnapshot(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Scan orphaned virtual machine data
// @Accept json
// @Produce json
// @Success 200 {array} dto.VMOrphanItem
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/orphans/scan [post]
func (b *BaseApi) ScanVMOrphans(c *gin.Context) {
	data, err := vmService.ScanOrphans()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Clean orphaned virtual machine data
// @Accept json
// @Param request body dto.VMOrphanClean true "request"
// @Produce json
// @Success 200 {array} dto.VMOrphanItem
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/orphans/clean [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"清理虚拟机孤立数据","formatEN":"clean VM orphaned data"}
func (b *BaseApi) CleanVMOrphans(c *gin.Context) {
	var req dto.VMOrphanClean
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.CleanOrphans(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Operate virtual machine
// @Accept json
// @Param request body dto.VMOperate true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/operate [post]
// @x-panel-log {"bodyKeys":["name","operate"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"操作虚拟机 [name] [operate]","formatEN":"operate VM [name] [operate]"}
func (b *BaseApi) OperateVM(c *gin.Context) {
	var req dto.VMOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := vmService.Operate(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags VirtualMachine
// @Summary Load virtual machine console information
// @Accept json
// @Param request body dto.VMReq true "request"
// @Produce json
// @Success 200 {object} dto.VMGraphics
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/console [post]
func (b *BaseApi) GetVMConsole(c *gin.Context) {
	var req dto.VMReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := vmService.Console(req.Name)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Connect virtual machine VNC console over WebSocket
// @Description Upgrades the connection to WebSocket and forwards VNC traffic.
// @Param name query string true "virtual machine name"
// @Success 101
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/console/ws [get]
func (b *BaseApi) ProxyVMConsole(c *gin.Context) {
	if !websocket.IsWebSocketUpgrade(c.Request) {
		helper.BadRequest(c, errors.New("websocket upgrade required"))
		return
	}
	itemName := c.Query("name")
	if !validateVMName(c, itemName) {
		return
	}
	graphics, err := vmService.Console(itemName)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	addr, err := buildVNCAddress(graphics)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	wsConn, err := vmConsoleUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.LOG.Errorf("upgrade vm console websocket failed, err: %v", err)
		return
	}
	defer wsConn.Close()

	vncConn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		global.LOG.Errorf("connect vm vnc failed, vm=%s, addr=%s, err=%v", itemName, addr, err)
		writeVMConsoleClose(wsConn, fmt.Sprintf("connect vnc failed: %v", err))
		return
	}
	defer vncConn.Close()

	errCh := make(chan error, 2)
	go forwardVNCToWebSocket(wsConn, vncConn, errCh)
	go forwardWebSocketToVNC(wsConn, vncConn, errCh)
	if err := <-errCh; err != nil && !errors.Is(err, io.EOF) {
		global.LOG.Debugf("vm console disconnected, vm=%s, err=%v", itemName, err)
	}
}

// @Tags VirtualMachine
// @Summary Page virtual machine ISO images
// @Accept json
// @Param request body dto.SearchPageWithType true "request"
// @Produce json
// @Success 200 {object} dto.PageResult{items=[]dto.VMIsoInfo}
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/iso/search [post]
func (b *BaseApi) SearchVMIso(c *gin.Context) {
	var req dto.SearchPageWithType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, list, err := vmService.SearchIso(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{Items: list, Total: total})
}

// @Tags VirtualMachine
// @Summary List virtual machine ISO images
// @Accept json
// @Produce json
// @Success 200 {array} dto.VMIsoInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/iso/list [post]
func (b *BaseApi) ListVMIso(c *gin.Context) {
	data, err := vmService.ListIso()
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Create virtual machine ISO image record
// @Accept json
// @Param request body dto.VMIsoCreate true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/iso [post]
// @x-panel-log {"bodyKeys":["name","type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建虚拟机镜像 [name] [type]","formatEN":"create VM ISO [name] [type]"}
func (b *BaseApi) CreateVMIso(c *gin.Context) {
	var req dto.VMIsoCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := vmService.CreateIso(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags VirtualMachine
// @Summary Start virtual machine ISO image import
// @Description Creates an import record and returns its id.
// @Accept json
// @Param request body dto.VMIsoImportStart true "request"
// @Produce json
// @Success 200 {object} map[string]uint
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/iso/import/start [post]
func (b *BaseApi) StartVMIsoImport(c *gin.Context) {
	var req dto.VMIsoImportStart
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	id, err := vmService.StartIsoImport(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, map[string]uint{"id": id})
}

// @Tags VirtualMachine
// @Summary Update virtual machine ISO image import status
// @Accept json
// @Param request body dto.VMIsoImportUpdate true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/iso/import/update [post]
func (b *BaseApi) UpdateVMIsoImport(c *gin.Context) {
	var req dto.VMIsoImportUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := vmService.UpdateIsoImport(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags VirtualMachine
// @Summary Update virtual machine ISO image record
// @Accept json
// @Param request body dto.VMIsoUpdate true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/iso/update [post]
// @x-panel-log {"bodyKeys":["name","type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新虚拟机镜像 [name] [type]","formatEN":"update VM ISO [name] [type]"}
func (b *BaseApi) UpdateVMIso(c *gin.Context) {
	var req dto.VMIsoUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := vmService.UpdateIso(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags VirtualMachine
// @Summary Delete virtual machine ISO image
// @Accept json
// @Param request body dto.VMIsoDelete true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/iso/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"vm_isos","output_column":"name","output_value":"name"}],"formatZH":"删除虚拟机镜像 [name]","formatEN":"delete VM ISO [name]"}
func (b *BaseApi) DeleteVMIso(c *gin.Context) {
	var req dto.VMIsoDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := vmService.DeleteIso(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags VirtualMachine
// @Summary List virtual machine networks
// @Accept json
// @Produce json
// @Success 200 {array} dto.VMNetworkItem
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/networks/list [post]
func (b *BaseApi) ListVMNetworks(c *gin.Context) {
	data, err := vmService.ListNetworks()
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary List host network interfaces for virtual machines
// @Accept json
// @Produce json
// @Success 200 {array} dto.VMNetworkInterfaceItem
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/networks/interfaces [get]
func (b *BaseApi) ListVMNetworkInterfaces(c *gin.Context) {
	data, err := vmService.ListNetworkInterfaces()
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary List host network bridges for virtual machines
// @Accept json
// @Produce json
// @Success 200 {array} dto.VMNetworkBridgeItem
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/networks/bridges [get]
func (b *BaseApi) ListVMNetworkBridges(c *gin.Context) {
	data, err := vmService.ListNetworkBridges()
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Create virtual machine network
// @Accept json
// @Param request body dto.VMNetworkCreate true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/networks [post]
// @x-panel-log {"bodyKeys":["name","type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建虚拟机网络 [name] [type]","formatEN":"create VM network [name] [type]"}
func (b *BaseApi) CreateVMNetwork(c *gin.Context) {
	var req dto.VMNetworkCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := vmService.CreateNetwork(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags VirtualMachine
// @Summary Update virtual machine network
// @Accept json
// @Param request body dto.VMNetworkUpdate true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/networks/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"vm_networks","output_column":"name","output_value":"name"}],"formatZH":"更新虚拟机网络 [name]","formatEN":"update VM network [name]"}
func (b *BaseApi) UpdateVMNetwork(c *gin.Context) {
	var req dto.VMNetworkUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := vmService.UpdateNetwork(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags VirtualMachine
// @Summary Delete virtual machine network
// @Accept json
// @Param request body dto.VMResourceDelete true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/networks/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"vm_networks","output_column":"name","output_value":"name"}],"formatZH":"删除虚拟机网络 [name]","formatEN":"delete VM network [name]"}
func (b *BaseApi) DeleteVMNetwork(c *gin.Context) {
	var req dto.VMResourceDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := vmService.DeleteNetwork(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags VirtualMachine
// @Summary List virtual machine storage pools
// @Accept json
// @Produce json
// @Success 200 {array} dto.VMStorageItem
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/storages/list [post]
func (b *BaseApi) ListVMStorages(c *gin.Context) {
	data, err := vmService.ListStorages()
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags VirtualMachine
// @Summary Create virtual machine storage pool
// @Accept json
// @Param request body dto.VMStorageCreate true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/storages [post]
// @x-panel-log {"bodyKeys":["name","type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建虚拟机存储 [name] [type]","formatEN":"create VM storage [name] [type]"}
func (b *BaseApi) CreateVMStorage(c *gin.Context) {
	var req dto.VMStorageCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := vmService.CreateStorage(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags VirtualMachine
// @Summary Update virtual machine storage pool
// @Accept json
// @Param request body dto.VMStorageUpdate true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/storages/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"vm_storages","output_column":"name","output_value":"name"}],"formatZH":"更新虚拟机存储 [name]","formatEN":"update VM storage [name]"}
func (b *BaseApi) UpdateVMStorage(c *gin.Context) {
	var req dto.VMStorageUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := vmService.UpdateStorage(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags VirtualMachine
// @Summary Delete virtual machine storage pool
// @Accept json
// @Param request body dto.VMResourceDelete true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /vms/storages/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"vm_storages","output_column":"name","output_value":"name"}],"formatZH":"删除虚拟机存储 [name]","formatEN":"delete VM storage [name]"}
func (b *BaseApi) DeleteVMStorage(c *gin.Context) {
	var req dto.VMResourceDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := vmService.DeleteStorage(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func validateVMName(c *gin.Context, name string) bool {
	if re.GetRegex(re.VMNameValidationPattern).MatchString(name) {
		return true
	}
	helper.BadRequest(c, buserr.New("ErrVMInvalidInput"))
	return false
}

var vmConsoleUpgrader = websocket.Upgrader{
	ReadBufferSize:  32 * 1024,
	WriteBufferSize: 32 * 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func buildVNCAddress(graphics *dto.VMGraphics) (string, error) {
	if graphics == nil || strings.ToLower(graphics.Type) != "vnc" {
		return "", errors.New("vnc console is unavailable")
	}
	port, err := strconv.Atoi(graphics.Port)
	if err != nil || port <= 0 {
		return "", fmt.Errorf("invalid vnc port: %s", graphics.Port)
	}
	host := strings.TrimSpace(graphics.Listen)
	if host == "" || host == "0.0.0.0" || host == "::" || host == "::0" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, strconv.Itoa(port)), nil
}

func forwardVNCToWebSocket(wsConn *websocket.Conn, vncConn net.Conn, errCh chan<- error) {
	buf := make([]byte, 32*1024)
	for {
		n, err := vncConn.Read(buf)
		if n > 0 {
			if writeErr := wsConn.WriteMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
				errCh <- writeErr
				return
			}
		}
		if err != nil {
			errCh <- err
			return
		}
	}
}

func forwardWebSocketToVNC(wsConn *websocket.Conn, vncConn net.Conn, errCh chan<- error) {
	for {
		messageType, data, err := wsConn.ReadMessage()
		if err != nil {
			errCh <- err
			return
		}
		if messageType != websocket.BinaryMessage && messageType != websocket.TextMessage {
			continue
		}
		if _, err := vncConn.Write(data); err != nil {
			errCh <- err
			return
		}
	}
}

func writeVMConsoleClose(wsConn *websocket.Conn, message string) {
	_ = wsConn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseTryAgainLater, message),
		time.Now().Add(time.Second),
	)
}
