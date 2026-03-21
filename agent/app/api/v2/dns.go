package v2

import (
	"strconv"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

// @Tags DNS
// @Summary Page DNS zones
// @Accept json
// @Param request body request.DnsZoneSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dns/zones/search [post]
func (b *BaseApi) PageDnsZone(c *gin.Context) {
	var req request.DnsZoneSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, zones, err := dnsService.PageZone(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: zones,
	})
}

// @Tags DNS
// @Summary Create DNS zone
// @Accept json
// @Param request body request.DnsZoneCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dns/zones [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建 DNS 域 [name]","formatEN":"Create DNS zone [name]"}
func (b *BaseApi) CreateDnsZone(c *gin.Context) {
	var req request.DnsZoneCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := dnsService.CreateZone(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags DNS
// @Summary Update DNS zone
// @Accept json
// @Param request body request.DnsZoneUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dns/zones/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"dns_zones","output_column":"name","output_value":"name"}],"formatZH":"更新 DNS 域 [name]","formatEN":"Update DNS zone [name]"}
func (b *BaseApi) UpdateDnsZone(c *gin.Context) {
	var req request.DnsZoneUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := dnsService.UpdateZone(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags DNS
// @Summary Get DNS zone
// @Accept json
// @Param id path integer true "zone id"
// @Success 200 {object} response.DnsZoneRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dns/zones/:id [get]
func (b *BaseApi) GetDnsZone(c *gin.Context) {
	idStr, ok := c.Params.Get("id")
	if !ok {
		helper.BadRequest(c, nil)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	zone, err := dnsService.GetZone(uint(id))
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, zone)
}

// @Tags DNS
// @Summary Delete DNS zone
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dns/zones/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"dns_zones","output_column":"name","output_value":"name"}],"formatZH":"删除 DNS 域 [name]","formatEN":"Delete DNS zone [name]"}
func (b *BaseApi) DeleteDnsZone(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := dnsService.DeleteZone(req.ID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags DNS
// @Summary Sync DNS zone from PowerDNS
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dns/zones/sync [post]
func (b *BaseApi) SyncDnsZone(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := dnsService.SyncZone(req.ID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags DNS
// @Summary Page DNS records
// @Accept json
// @Param request body request.DnsRecordSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dns/records/search [post]
func (b *BaseApi) PageDnsRecord(c *gin.Context) {
	var req request.DnsRecordSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, records, err := dnsService.PageRecord(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: records,
	})
}

// @Tags DNS
// @Summary Create DNS record
// @Accept json
// @Param request body request.DnsRecordCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dns/records [post]
// @x-panel-log {"bodyKeys":["name","type"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建 DNS 记录 [name][type]","formatEN":"Create DNS record [name][type]"}
func (b *BaseApi) CreateDnsRecord(c *gin.Context) {
	var req request.DnsRecordCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := dnsService.CreateRecord(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags DNS
// @Summary Update DNS record
// @Accept json
// @Param request body request.DnsRecordUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dns/records/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"dns_records","output_column":"name","output_value":"name"}],"formatZH":"更新 DNS 记录 [name]","formatEN":"Update DNS record [name]"}
func (b *BaseApi) UpdateDnsRecord(c *gin.Context) {
	var req request.DnsRecordUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := dnsService.UpdateRecord(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags DNS
// @Summary Delete DNS record
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dns/records/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"dns_records","output_column":"name","output_value":"name"}],"formatZH":"删除 DNS 记录 [name]","formatEN":"Delete DNS record [name]"}
func (b *BaseApi) DeleteDnsRecord(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := dnsService.DeleteRecord(req.ID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags DNS
// @Summary Setup DNS (enable/disable PowerDNS)
// @Accept json
// @Param request body request.DnsSetup true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dns/setup [post]
// @x-panel-log {"bodyKeys":["enable"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"DNS 设置 [enable]","formatEN":"DNS setup [enable]"}
func (b *BaseApi) SetupDns(c *gin.Context) {
	var req request.DnsSetup
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := dnsService.Setup(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags DNS
// @Summary Get DNS status
// @Success 200 {object} response.DnsStatus
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dns/status [get]
func (b *BaseApi) GetDnsStatus(c *gin.Context) {
	status, err := dnsService.GetStatus()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, status)
}
