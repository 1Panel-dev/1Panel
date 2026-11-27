package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

// @Tags Website DNS
// @Summary List zones of a website dns account
// @Accept json
// @Param request body request.DnsZoneSearch true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/dns/record/zones [post]
func (b *BaseApi) ListWebsiteDnsZones(c *gin.Context) {
	var req request.DnsZoneSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	zones, err := websiteDnsRecordService.ListZones(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, zones)
}

// @Tags Website DNS
// @Summary List dns records of a zone
// @Accept json
// @Param request body request.DnsRecordSearch true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/dns/record/search [post]
func (b *BaseApi) ListWebsiteDnsRecords(c *gin.Context) {
	var req request.DnsRecordSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	records, err := websiteDnsRecordService.ListRecords(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, records)
}

// @Tags Website DNS
// @Summary Create dns record
// @Accept json
// @Param request body request.DnsRecordOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/dns/record [post]
// @x-panel-log {"bodyKeys":["zoneName","name","type"],"paramKeys":[],"BeforeFunctions":[],"formatEN":"Create dns record [name] [type] in [zoneName]"}
func (b *BaseApi) CreateWebsiteDnsRecord(c *gin.Context) {
	var req request.DnsRecordOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteDnsRecordService.CreateRecord(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Website DNS
// @Summary Update dns record
// @Accept json
// @Param request body request.DnsRecordOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/dns/record/update [post]
// @x-panel-log {"bodyKeys":["zoneName","name","type"],"paramKeys":[],"BeforeFunctions":[],"formatEN":"Update dns record [name] [type] in [zoneName]"}
func (b *BaseApi) UpdateWebsiteDnsRecord(c *gin.Context) {
	var req request.DnsRecordOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteDnsRecordService.UpdateRecord(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Website DNS
// @Summary Delete dns record
// @Accept json
// @Param request body request.DnsRecordDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/dns/record/del [post]
// @x-panel-log {"bodyKeys":["zoneName","name","type"],"paramKeys":[],"BeforeFunctions":[],"formatEN":"Delete dns record [name] [type] in [zoneName]"}
func (b *BaseApi) DeleteWebsiteDnsRecord(c *gin.Context) {
	var req request.DnsRecordDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteDnsRecordService.DeleteRecord(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
