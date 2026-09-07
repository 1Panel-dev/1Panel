package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/auth"
	baseDto "github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/1Panel-dev/1Panel/core/utils/mfa"
	"github.com/1Panel-dev/1Panel/core/utils/req_helper/proxy_local"
	terminalsession "github.com/1Panel-dev/1Panel/core/utils/terminal_session"
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

func (a *authHelper) PrepareLogout(_ *gin.Context) (*baseDto.LogoutResult, error) {
	return &baseDto.LogoutResult{}, nil
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

func (a *authHelper) CoreAPIAuthMiddleware() gin.HandlerFunc {
	return auth.APIAuthMiddleware(auth.LoadAPIAuthConfig, func(c *gin.Context, _ auth.APIAuthConfig) {
		name, _ := repo.NewISettingRepo().GetValueByKey("UserName")
		c.Set("API_AUTH_USERNAME", name)
		c.Set(psession.GinContextSessionUserKey, psession.SessionUser{
			ID: psession.SuperAdminSessionUserID, Name: name, Role: "ADMIN",
		})
	})
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
	apiKey, err := auth.GenerateApiKey()
	if err != nil {
		return "", err
	}
	userID := psession.SuperAdminSessionUserID
	if err := a.RevokeTerminalSessions("auth_session", userID, terminalsession.APIAuthSessionID(userID)); err != nil {
		global.LOG.Warnf("revoke API terminal sessions after API key generation failed, err: %v", err)
	}
	return apiKey, nil
}
func (a *authHelper) UpdateApiConfig(c *gin.Context, req baseDto.ApiInterfaceConfig) error {
	if err := auth.UpdateApiConfig(req); err != nil {
		return err
	}
	userID := psession.SuperAdminSessionUserID
	if err := a.RevokeTerminalSessions("auth_session", userID, terminalsession.APIAuthSessionID(userID)); err != nil {
		global.LOG.Warnf("revoke API terminal sessions after API config update failed, err: %v", err)
	}
	return nil
}

func (a *authHelper) GetCurrentUserInfo(_ *gin.Context) (*baseDto.CurrentUserInfo, error) {
	return auth.GetCurrentUserInfo()
}
func (a *authHelper) ShouldCheckPasswordExpiration(_ *gin.Context) (bool, error) {
	return true, nil
}
func (a *authHelper) LoadPasswordExpirationTime(c *gin.Context) (string, error) {
	return auth.LoadPasswordExpirationTime(c)
}
func (a *authHelper) SyncPasswordExpirationTime(expirationDays string) error {
	return auth.SyncPasswordExpirationTime(expirationDays)
}
func (a *authHelper) UpdateCurrentUserInfo(c *gin.Context, req baseDto.CurrentUserUpdate) error {
	identity, _ := terminalsession.FromContext(c)
	if err := auth.UpdateCurrentUserInfo(c, req); err != nil {
		return err
	}
	if identity.UserID != "" {
		if err := a.RevokeTerminalSessions("user", identity.UserID, ""); err != nil {
			global.LOG.Warnf("revoke terminal sessions after user update failed, err: %v", err)
		}
	}
	return nil
}
func (a *authHelper) HandlePasswordExpired(c *gin.Context, old, new string) error {
	identity, _ := terminalsession.FromContext(c)
	if err := auth.HandlePasswordExpired(c, old, new); err != nil {
		return err
	}
	if identity.UserID != "" {
		if err := a.RevokeTerminalSessions("user", identity.UserID, ""); err != nil {
			global.LOG.Warnf("revoke terminal sessions after password change failed, err: %v", err)
		}
	}
	return nil
}

func (a *authHelper) RevokeTerminalSessions(scope, userID, authSessionID string) error {
	body, err := json.Marshal(map[string]string{
		"scope": scope, "userId": userID, "authSessionId": authSessionID,
	})
	if err != nil {
		return err
	}
	_, err = proxy_local.NewLocalClientWithContext(
		context.Background(),
		"/api/v2/internal/terminal/sessions/revoke",
		http.MethodPost,
		bytes.NewReader(body),
		nil,
		5*time.Second,
	)
	return err
}
