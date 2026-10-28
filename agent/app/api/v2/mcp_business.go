package v2

import (
	"encoding/base64"
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/utils/mcpapi"
	"github.com/gin-gonic/gin"
)

type mcpBusinessAPI struct {
	installApp func(request.AppInstallCreate, bool) (*model.AppInstall, error)
}

func MCPBusinessDispatch() gin.HandlerFunc {
	return (&mcpBusinessAPI{installApp: (&service.AppService{}).InstallForMCP}).dispatch()
}

func (api *mcpBusinessAPI) dispatch() gin.HandlerFunc {
	return func(c *gin.Context) {
		var handler gin.HandlerFunc
		switch c.Request.Method + " " + c.FullPath() {
		case "POST /api/v2/databases":
			handler = api.createMysql
		case "POST /api/v2/databases/status":
			handler = api.mysqlStatus
		case "GET /api/v2/databases/db/list/:type":
			handler = api.databases
		case "POST /api/v2/apps/install":
			handler = api.install
		case "GET /api/v2/apps/installed/info/:appInstallId":
			handler = api.appInfo
		default:
			c.Next()
			return
		}
		present, trusted := mcpapi.Check(c.Request)
		if !present {
			c.Next()
			return
		}
		if !trusted {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		handler(c)
		c.Abort()
	}
}

func (api *mcpBusinessAPI) createMysql(c *gin.Context) {
	var req dto.MysqlDBCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if len(req.Password) != 0 {
		password, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			helper.BadRequest(c, err)
			return
		}
		req.Password = string(password)
	}
	created, err := (&service.MysqlService{}).CreateForMCP(c.Request.Context(), req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	nodeID, _ := mcpapi.NodeID(c.Request)
	helper.SuccessWithData(c, map[string]any{"id": created.ID, "database": created.MysqlName, "name": created.Name, "nodeId": nodeID})
}

func (api *mcpBusinessAPI) mysqlStatus(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := (&service.MysqlService{}).LoadStatusForMCP(c.Request.Context(), req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

func (api *mcpBusinessAPI) databases(c *gin.Context) {
	dbType, err := helper.GetStrParamByKey(c, "type")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	list, err := (&service.DatabaseService{}).ListForMCP(c.Request.Context(), dbType)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, list)
}

func (api *mcpBusinessAPI) install(c *gin.Context) {
	var req request.AppInstallCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if value, exists := req.Params["WEBSITE_DIR"]; exists {
		directory, ok := value.(string)
		if !ok || directory == "" || directory != settingService.GetWebsiteDir() {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
	installed, err := api.installApp(req, true)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, installed)
}

func (api *mcpBusinessAPI) appInfo(c *gin.Context) {
	id, err := helper.GetIntParamByKey(c, "appInstallId")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	info, err := (&service.AppInstallService{}).GetAppInstallInfoForMCP(c.Request.Context(), id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, info)
}
