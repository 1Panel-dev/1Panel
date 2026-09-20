package v2

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/auth"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/gin-gonic/gin"
)

func bindAPIKeyRequest(c *gin.Context, value interface{}) bool {
	if _, err := auth.RequireAPIKeySession(c); err != nil {
		helper.BadAuth(c, "ErrNotLogin", nil)
		return false
	}
	decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 32<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		helper.BadRequest(c, err)
		return false
	}
	if err := decoder.Decode(new(interface{})); err != io.EOF {
		helper.BadRequest(c, errors.New("unexpected JSON data"))
		return false
	}
	return true
}

func apiKeyError(c *gin.Context, err error) {
	var business buserr.BusinessError
	if errors.As(err, &business) {
		helper.ErrorWithDetail(c, http.StatusBadRequest, business.Msg, business.Err)
		return
	}
	helper.InternalServer(c, err)
}

func (b *BaseApi) SearchAPIKeys(c *gin.Context) {
	var req dto.APIKeySearch
	if !bindAPIKeyRequest(c, &req) {
		return
	}
	result, err := service.NewAPIKeyService().Search(c, req)
	if err != nil {
		apiKeyError(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	helper.SuccessWithData(c, result)
}

func (b *BaseApi) CreateAPIKey(c *gin.Context) {
	req := dto.APIKeyCreate{APIKeyFields: dto.APIKeyFields{APIKeyValidityTime: 120}}
	if !bindAPIKeyRequest(c, &req) {
		return
	}
	result, err := service.NewAPIKeyService().Create(c, req)
	if err != nil {
		apiKeyError(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	helper.SuccessWithData(c, result)
}

func (b *BaseApi) UpdateAPIKey(c *gin.Context) {
	var req dto.APIKeyUpdate
	if !bindAPIKeyRequest(c, &req) {
		return
	}
	if err := service.NewAPIKeyService().Update(c, req); err != nil {
		apiKeyError(c, err)
		return
	}
	apiKeyMutationSuccess(c)
}

func (b *BaseApi) SetAPIKeyStatus(c *gin.Context) {
	var req dto.APIKeyStatus
	if !bindAPIKeyRequest(c, &req) {
		return
	}
	if err := service.NewAPIKeyService().Status(c, req); err != nil {
		apiKeyError(c, err)
		return
	}
	apiKeyMutationSuccess(c)
}

func (b *BaseApi) RevokeAPIKey(c *gin.Context) {
	var req dto.APIKeyMutation
	if !bindAPIKeyRequest(c, &req) {
		return
	}
	if err := service.NewAPIKeyService().Revoke(c, req); err != nil {
		apiKeyError(c, err)
		return
	}
	apiKeyMutationSuccess(c)
}

func apiKeyMutationSuccess(c *gin.Context) {
	helper.SuccessWithData(c, gin.H{"terminalClosePending": c.GetBool("API_KEY_TERMINAL_CLOSE_PENDING")})
}
