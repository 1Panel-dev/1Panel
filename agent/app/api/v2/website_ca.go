package v2

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/gin-gonic/gin"
)

// @Tags Website CA
// @Summary Page website ca
// @Accept json
// @Param request body request.WebsiteCASearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ca/search [post]
func (b *BaseApi) PageWebsiteCA(c *gin.Context) {
	var req request.WebsiteCASearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, cas, err := websiteCAService.Page(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: cas,
	})
}

// @Tags Website CA
// @Summary Create website ca
// @Accept json
// @Param request body request.WebsiteCACreate true "request"
// @Success 200 {object} request.WebsiteCACreate
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ca [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建网站 ca [name]","formatEN":"Create website ca [name]"}
func (b *BaseApi) CreateWebsiteCA(c *gin.Context) {
	var req request.WebsiteCACreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteCAService.Create(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website CA
// @Summary Get website ca
// @Accept json
// @Param id path int true "id"
// @Success 200 {object} response.WebsiteCADTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ca/{id} [get]
func (b *BaseApi) GetWebsiteCA(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		return
	}
	res, err := websiteCAService.GetCA(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website CA
// @Summary Delete website ca
// @Accept json
// @Param request body request.WebsiteCommonReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ca/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_cas","output_column":"name","output_value":"name"}],"formatZH":"删除网站 ca [name]","formatEN":"Delete website ca [name]"}
func (b *BaseApi) DeleteWebsiteCA(c *gin.Context) {
	var req request.WebsiteCommonReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteCAService.Delete(req.ID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Website CA
// @Summary Obtain SSL
// @Accept json
// @Param request body request.WebsiteCAObtain true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ca/obtain [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_cas","output_column":"name","output_value":"name"}],"formatZH":"自签 SSL 证书 [name]","formatEN":"Obtain SSL [name]"}
func (b *BaseApi) ObtainWebsiteCA(c *gin.Context) {
	var req request.WebsiteCAObtain
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := websiteCAService.ObtainSSL(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Website CA
// @Summary Obtain SSL
// @Accept json
// @Param request body request.WebsiteCAObtain true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ca/renew [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_cas","output_column":"name","output_value":"name"}],"formatZH":"自签 SSL 证书 [name]","formatEN":"Obtain SSL [name]"}
func (b *BaseApi) RenewWebsiteCA(c *gin.Context) {
	var req request.WebsiteCARenew
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if _, err := websiteCAService.ObtainSSL(request.WebsiteCAObtain{
		SSLID: req.SSLID,
		Renew: true,
		Unit:  "year",
		Time:  1,
	}); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Website CA
// @Summary Download CA file
// @Accept json
// @Param request body request.WebsiteResourceReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router  /websites/ca/download [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_cas","output_column":"name","output_value":"name"}],"formatZH":"下载 CA 证书文件 [name]","formatEN":"download ca file [name]"}
func (b *BaseApi) DownloadCAFile(c *gin.Context) {
	var req request.WebsiteResourceReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	file, err := websiteCAService.DownloadFile(req.ID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
	c.Header("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(info.Name()))
	http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), file)
}

// @Tags Website CA
// @Summary Upload website CA archive
// @Accept multipart/form-data
// @Param file formData file true "ZIP containing ca.crt and ca.key"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ca/upload [post]
func (b *BaseApi) UploadWebsiteCA(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 11<<20)
	if err := c.Request.ParseMultipartForm(11 << 20); err != nil {
		helper.InternalServer(c, buserr.New("ErrCAUpload"))
		return
	}
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	file, err := c.FormFile("file")
	if err != nil || file.Size > 10<<20 || !strings.HasSuffix(strings.ToLower(file.Filename), ".zip") {
		helper.InternalServer(c, buserr.New("ErrCAUpload"))
		return
	}
	stream, err := file.Open()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	defer stream.Close()
	data, err := io.ReadAll(io.LimitReader(stream, (10<<20)+1))
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	if err = websiteCAService.Upload(file.Filename[:len(file.Filename)-len(".zip")], data); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Website CA
// @Summary Import website CA from PEM or server files
// @Accept json
// @Param request body request.WebsiteCAImport true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/ca/import [post]
func (b *BaseApi) ImportWebsiteCA(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 3<<20)
	var req request.WebsiteCAImport
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteCAService.Import(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
