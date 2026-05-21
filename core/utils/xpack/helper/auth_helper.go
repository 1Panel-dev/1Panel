package helper

import (
	"time"

	"github.com/1Panel-dev/1Panel/core/app/auth"
	baseDto "github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/1Panel-dev/1Panel/core/utils/mfa"
	"github.com/1Panel-dev/1Panel/core/utils/xpack/providers"
	"github.com/gin-gonic/gin"
)

type authHelper struct{}

func NewIAuthProvider() providers.AuthProvider {
	return &authHelper{}
}

func (a *authHelper) Login(c *gin.Context, info baseDto.Login, entrance string) (*baseDto.UserLoginInfo, string, error) {
	return auth.Login(c, info, entrance)
}

func (a *authHelper) MFALogin(c *gin.Context, info baseDto.MFALogin, entrance string) (*baseDto.UserLoginInfo, string, error) {
	return auth.MFALogin(c, info, entrance)
}

func (a *authHelper) PasskeyBeginLogin(c *gin.Context, entrance string) (*baseDto.PasskeyBeginResponse, string, error) {
	return auth.PasskeyBeginLogin(c, entrance)
}
func (a *authHelper) PasskeyFinishLogin(c *gin.Context, sessionID, entrance string) (*baseDto.UserLoginInfo, string, error) {
	return auth.PasskeyFinishLogin(c, sessionID, entrance)
}
func (a *authHelper) PasskeyBeginRegister(c *gin.Context, name string) (*baseDto.PasskeyBeginResponse, string, error) {
	return auth.PasskeyBeginRegister(c, name)
}
func (a *authHelper) PasskeyFinishRegister(c *gin.Context, sessionID string) (string, error) {
	return auth.PasskeyFinishRegister(c, sessionID)
}
func (a *authHelper) PasskeyList(c *gin.Context) ([]baseDto.PasskeyInfo, error) {
	return auth.PasskeyList()
}
func (a *authHelper) PasskeyDelete(c *gin.Context, id string) error {
	return auth.PasskeyDelete(id)
}
func (a *authHelper) PasskeyStatus(c *gin.Context) bool {
	return auth.PasskeyStatus(c)
}
func (a *authHelper) ClearPasskeys() error {
	return nil
}

func (a *authHelper) ResetSuperAdminUser(name, password string) error {
	return nil
}

func (a *authHelper) LoadSessionTimeout(_ *gin.Context, sessionUser psession.SessionUser) (int, error) {
	return auth.LoadSessionTimeout(sessionUser)
}
func (a *authHelper) LoadExpired(_ *gin.Context, sessionUser psession.SessionUser) (bool, time.Time, error) {
	return auth.LoadExpired(sessionUser)
}

func (a *authHelper) CoreAPIAuthMiddleware() gin.HandlerFunc {
	return auth.APIAuthMiddleware(auth.LoadAPIAuthConfig, nil)
}

func (a *authHelper) CoreRBACMiddlewares() []gin.HandlerFunc { return nil }

func (a *authHelper) LoadMFA(_ *gin.Context, req baseDto.MfaRequest) (mfa.Otp, error) {
	return auth.LoadMFA(req)
}
func (a *authHelper) MFABind(_ *gin.Context, req baseDto.MfaCredential) error {
	return auth.MFABind(req)
}
func (a *authHelper) MFAClose(_ *gin.Context) error {
	return auth.MFAClose()
}
func (a *authHelper) GenerateApiKey(_ *gin.Context) (string, error) {
	return auth.GenerateApiKey()
}
func (a *authHelper) UpdateApiConfig(c *gin.Context, req baseDto.ApiInterfaceConfig) error {
	return auth.UpdateApiConfig(req)
}

func (a *authHelper) GetCurrentUserInfo(_ *gin.Context) (*baseDto.CurrentUserInfo, error) {
	return auth.GetCurrentUserInfo()
}
func (a *authHelper) UpdateCurrentUserInfo(c *gin.Context, req baseDto.CurrentUserUpdate) error {
	return auth.UpdateCurrentUserInfo(c, req)
}
func (a *authHelper) HandlePasswordExpired(c *gin.Context, old, new string) error {
	return auth.HandlePasswordExpired(c, old, new)
}
