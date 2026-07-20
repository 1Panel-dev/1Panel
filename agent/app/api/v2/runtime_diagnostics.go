package v2

import (
	"os"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

// @Tags RuntimeDiagnostics
// @Summary Load runtime diagnostics summary
// @Success 200 {object} dto.RuntimeDiagnosticsSummary
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/diagnostics/summary [get]
func (b *BaseApi) LoadRuntimeDiagnosticsSummary(c *gin.Context) {
	data, err := runtimeDiagnosticsService.Summary()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags RuntimeDiagnostics
// @Summary Load grouped goroutine snapshot
// @Success 200 {object} dto.RuntimeGoroutineSnapshot
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/diagnostics/goroutines [get]
func (b *BaseApi) LoadRuntimeGoroutines(c *gin.Context) {
	data, err := runtimeDiagnosticsService.Goroutines()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithDataGzipped(c, data)
}

// @Tags RuntimeDiagnostics
// @Summary Capture runtime profile
// @Param request body dto.RuntimeProfileCreate true "request"
// @Success 200 {file} file
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/diagnostics/profiles [post]
func (b *BaseApi) CreateRuntimeProfile(c *gin.Context) {
	var req dto.RuntimeProfileCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	profile, err := runtimeDiagnosticsService.CreateProfile(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	defer os.Remove(profile.Path)
	c.Header("Content-Disposition", `attachment; filename="`+profile.Name+`"`)
	c.Header("Content-Type", "application/octet-stream")
	c.File(profile.Path)
	c.Abort()
}
