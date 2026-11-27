package service

import (
	"context"
	"os"
	"runtime"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/hostnetwork"
	"github.com/1Panel-dev/1Panel/agent/utils/psutil"
)

var hostNetworkDistro = &psutil.HostInfoState{}

func (s *HostNetworkService) Overview() (dto.HostNetworkOverview, error) {
	hostNetworkMu.Lock()
	defer hostNetworkMu.Unlock()

	result := dto.HostNetworkOverview{}
	if runtime.GOOS == "linux" {
		result.Persistence.OS = hostNetworkDistro.GetDistro()
	}
	reason, supportErr := hostNetworkPersistenceSupport()
	result.Persistence.Reason = reason
	links, err := hostnetwork.Links("")
	if err != nil {
		return result, hostNetworkSystemError(err)
	}
	routes, err := loadHostNetworkRoutes(context.Background(), "")
	if err != nil {
		return result, err
	}
	plans := map[string]*hostnetwork.PersistentNetwork{}
	reasons := map[string]string{}
	reader := &hostnetwork.PersistentNetworkReader{}
	result.Interfaces = hostNetworkInterfaces(links, plans, reasons, reader)
	hostNetworkRoutes(routes, plans, reasons, reader)
	result.Routes = routes
	result.Persistence.Supported = supportErr == nil
	return result, nil
}

func hostNetworkPersistenceSupport() (string, error) {
	if runtime.GOOS != "linux" {
		return "platform", buserr.New("ErrNetworkPersistencePlatform")
	}
	if os.Geteuid() != 0 || !hostnetwork.HasAdmin() {
		return "privilege", buserr.New("ErrNetworkPersistencePrivilege")
	}
	return "", nil
}
