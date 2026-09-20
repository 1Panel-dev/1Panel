package auth

import (
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/gin-gonic/gin"
)

type PanelAPIKeyOwnerProvider struct{}

func (p PanelAPIKeyOwnerProvider) CurrentAPIKeyOwner(c *gin.Context) (APIKeyOwner, error) {
	value, ok := c.Get(psession.GinContextSessionUserKey)
	u, valid := value.(psession.SessionUser)
	if !ok || !valid || u.ID != psession.SuperAdminSessionUserID {
		return APIKeyOwner{}, buserr.New("ErrNotLogin")
	}
	return p.ResolveAPIKeyOwner(c, APIKeyOwnerPanelAdmin, u.ID)
}

func (PanelAPIKeyOwnerProvider) ResolveAPIKeyOwner(_ *gin.Context, ownerType, ownerID string) (APIKeyOwner, error) {
	if ownerType != APIKeyOwnerPanelAdmin || ownerID != psession.SuperAdminSessionUserID {
		return APIKeyOwner{}, buserr.New("ErrApiConfigStatusInvalid")
	}
	name, err := repo.NewISettingRepo().GetValueByKey("UserName")
	if err != nil || name == "" {
		return APIKeyOwner{}, buserr.New("ErrApiConfigStatusInvalid")
	}
	return APIKeyOwner{Type: ownerType, ID: ownerID, Name: name, IsSuperAdmin: true}, nil
}

func (p PanelAPIKeyOwnerProvider) LoadLegacyAPIKey(owner APIKeyOwner) (APIAuthConfig, error) {
	if _, err := p.ResolveAPIKeyOwner(nil, owner.Type, owner.ID); err != nil {
		return APIAuthConfig{}, err
	}
	return LoadAPIAuthConfig(nil)
}

func (p PanelAPIKeyOwnerProvider) SaveLegacyAPIKey(owner APIKeyOwner, config APIAuthConfig) error {
	if _, err := p.ResolveAPIKeyOwner(nil, owner.Type, owner.ID); err != nil {
		return err
	}
	return StoreLegacyAPIConfig(dto.ApiInterfaceConfig{ApiInterfaceStatus: config.ApiInterfaceStatus, ApiKey: config.ApiKey, IpWhiteList: config.IpWhiteList, ApiTrustedProxies: config.ApiTrustedProxies, ApiKeyValidityTime: config.ApiKeyValidityTime})
}
