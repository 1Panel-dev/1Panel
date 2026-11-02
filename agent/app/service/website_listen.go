package service

import (
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
)

var websiteListenMutation sync.Mutex

func websiteListenPorts(website model.Website, httpPort, httpsPort int) map[int]struct{} {
	ports := make(map[int]struct{})
	for _, domain := range website.Domains {
		if domain.Port > 0 {
			ports[domain.Port] = struct{}{}
		}
	}
	if website.Protocol == constant.ProtocolHTTPS {
		plan := buildWebsiteTLSPlan(website.Domains, httpPort, httpsPort)
		for _, port := range plan.httpsPorts {
			if port > 0 {
				ports[port] = struct{}{}
			}
		}
	}
	return ports
}

func checkWebsiteListenConflict(website model.Website, httpPort, httpsPort int) error {
	return checkWebsiteListenPorts(website, websiteListenPorts(website, httpPort, httpsPort), httpPort, httpsPort)
}

func checkWebsiteListenPorts(website model.Website, ports map[int]struct{}, httpPort, httpsPort int) error {
	if len(website.Domains) == 0 || len(ports) == 0 {
		return nil
	}
	names := make(map[string]struct{}, len(website.Domains))
	for _, domain := range website.Domains {
		names[strings.ToLower(domain.Domain)] = struct{}{}
	}
	existing, err := websiteRepo.GetBy()
	if err != nil {
		return err
	}
	domains, err := websiteDomainRepo.GetBy()
	if err != nil {
		return err
	}
	byWebsite := make(map[uint][]model.WebsiteDomain)
	for _, domain := range domains {
		byWebsite[domain.WebsiteID] = append(byWebsite[domain.WebsiteID], domain)
	}
	for _, other := range existing {
		if website.ID != 0 && other.ID == website.ID || other.Type == constant.Stream {
			continue
		}
		other.Domains = byWebsite[other.ID]
		for _, domain := range other.Domains {
			if _, match := names[strings.ToLower(domain.Domain)]; !match {
				continue
			}
			for port := range websiteListenPorts(other, httpPort, httpsPort) {
				if _, occupied := ports[port]; occupied {
					return buserr.WithName("ErrDomainIsUsed", other.PrimaryDomain)
				}
			}
			break
		}
	}
	return nil
}
