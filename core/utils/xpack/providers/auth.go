package providers

import (
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/1Panel-dev/1Panel/core/utils/mfa"
	"github.com/gin-gonic/gin"
)

type AuthProvider interface {
	Login(c *gin.Context, info dto.Login, entrance string) (*dto.UserLoginInfo, string, error)
	MFALogin(c *gin.Context, info dto.MFALogin, entrance string) (*dto.UserLoginInfo, string, error)

	ResetSuperAdminUser(name, password string) error

	LoadSessionTimeout(c *gin.Context, sessionUser psession.SessionUser) (int, error)
	LoadExpired(c *gin.Context, sessionUser psession.SessionUser) (bool, time.Time, error)

	LoadMFA(c *gin.Context, req dto.MfaRequest) (mfa.Otp, error)
	MFABind(c *gin.Context, req dto.MfaCredential) error
	MFAClose(c *gin.Context) error

	GenerateApiKey(c *gin.Context) (string, error)
	UpdateApiConfig(c *gin.Context, req dto.ApiInterfaceConfig) error

	PasskeyBeginLogin(c *gin.Context, entrance string) (*dto.PasskeyBeginResponse, string, error)
	PasskeyFinishLogin(c *gin.Context, sessionID, entrance string) (*dto.UserLoginInfo, string, error)
	PasskeyBeginRegister(c *gin.Context, name string) (*dto.PasskeyBeginResponse, string, error)
	PasskeyFinishRegister(c *gin.Context, sessionID string) (string, error)
	PasskeyList(c *gin.Context) ([]dto.PasskeyInfo, error)
	PasskeyDelete(c *gin.Context, id string) error
	PasskeyStatus(c *gin.Context) bool
	ClearPasskeys() error

	GetCurrentUserInfo(c *gin.Context) (*dto.CurrentUserInfo, error)
	UpdateCurrentUserInfo(c *gin.Context, req dto.CurrentUserUpdate) error
	HandlePasswordExpired(c *gin.Context, old, new string) error

	CoreAPIAuthMiddleware() gin.HandlerFunc
	CoreRBACMiddlewares() []gin.HandlerFunc
}
