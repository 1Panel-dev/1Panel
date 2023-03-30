package service

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/utils/firewall"
	fireClient "github.com/1Panel-dev/1Panel/backend/utils/firewall/client"
	"github.com/jinzhu/copier"
)

type FirewallService struct{}

type IFirewallService interface {
	LoadBaseInfo() (dto.FirewallBaseInfo, error)
	SearchWithPage(search dto.RuleSearch) (int64, interface{}, error)
	OperateFirewall(operation string) error
	OperatePortRule(req dto.PortRuleOperate, reload bool) error
	OperateAddressRule(req dto.AddrRuleOperate, reload bool) error
	UpdatePortRule(req dto.PortRuleUpdate) error
	UpdateAddrRule(req dto.AddrRuleUpdate) error
	BacthOperateRule(req dto.BatchRuleOperate) error
}

func NewIFirewallService() IFirewallService {
	return &FirewallService{}
}

func (u *FirewallService) LoadBaseInfo() (dto.FirewallBaseInfo, error) {
	var baseInfo dto.FirewallBaseInfo
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return baseInfo, err
	}
	baseInfo.Name = client.Name()
	baseInfo.Status, err = client.Status()
	if err != nil {
		return baseInfo, err
	}
	if baseInfo.Status == "not running" {
		baseInfo.Version = "-"
		return baseInfo, err
	}
	baseInfo.Version, err = client.Version()
	if err != nil {
		return baseInfo, err
	}
	return baseInfo, nil
}

func (u *FirewallService) SearchWithPage(req dto.RuleSearch) (int64, interface{}, error) {
	var (
		datas     []fireClient.FireInfo
		backDatas []fireClient.FireInfo
	)
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return 0, nil, err
	}
	if req.Type == "port" {
		ports, err := client.ListPort()
		if err != nil {
			return 0, nil, err
		}
		if len(req.Info) != 0 {
			for _, port := range ports {
				if strings.Contains(port.Port, req.Info) {
					datas = append(datas, port)
				}
			}
		} else {
			datas = ports
		}
	} else {
		addrs, err := client.ListAddress()
		if err != nil {
			return 0, nil, err
		}
		if len(req.Info) != 0 {
			for _, addr := range addrs {
				if strings.Contains(addr.Address, req.Info) {
					datas = append(datas, addr)
				}
			}
		} else {
			datas = addrs
		}
	}
	total, start, end := len(datas), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		backDatas = make([]fireClient.FireInfo, 0)
	} else {
		if end >= total {
			end = total
		}
		backDatas = datas[start:end]
	}

	return int64(total), backDatas, nil
}

func (u *FirewallService) OperateFirewall(operation string) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	switch operation {
	case "start":
		return client.Start()
	case "stop":
		return client.Stop()
	case "reload":
		return client.Reload()
	}
	return fmt.Errorf("not support such operation: %s", operation)
}

func (u *FirewallService) OperatePortRule(req dto.PortRuleOperate, reload bool) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	if client.Name() == "ufw" {
		req.Port = strings.ReplaceAll(req.Port, "-", ":")
		if req.Operation == "remove" && req.Protocol == "tcp/udp" {
			req.Protocol = ""
			return u.operatePort(client, req)
		}
	}

	if req.Protocol == "tcp/udp" {
		req.Protocol = "tcp"
		if err := u.operatePort(client, req); err != nil {
			return err
		}
		req.Protocol = "udp"
	}
	if err := u.operatePort(client, req); err != nil {
		return err
	}
	if reload {
		return client.Reload()
	}
	return nil
}

func (u *FirewallService) OperateAddressRule(req dto.AddrRuleOperate, reload bool) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}

	var fireInfo fireClient.FireInfo
	if err := copier.Copy(&fireInfo, &req); err != nil {
		return err
	}

	addressList := strings.Split(req.Address, ",")
	for _, addr := range addressList {
		if len(addr) == 0 {
			continue
		}
		fireInfo.Address = addr
		if err := client.RichRules(fireInfo, req.Operation); err != nil {
			return err
		}
	}
	if reload {
		return client.Reload()
	}
	return nil
}

func (u *FirewallService) UpdatePortRule(req dto.PortRuleUpdate) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	if err := u.OperatePortRule(req.OldRule, false); err != nil {
		return err
	}
	if err := u.OperatePortRule(req.NewRule, false); err != nil {
		return err
	}
	return client.Reload()
}

func (u *FirewallService) UpdateAddrRule(req dto.AddrRuleUpdate) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	if err := u.OperateAddressRule(req.OldRule, false); err != nil {
		return err
	}
	if err := u.OperateAddressRule(req.NewRule, false); err != nil {
		return err
	}
	return client.Reload()
}

func (u *FirewallService) BacthOperateRule(req dto.BatchRuleOperate) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	if req.Type == "port" {
		for _, rule := range req.Rules {
			if err := u.OperatePortRule(rule, false); err != nil {
				return err
			}
		}
		return client.Reload()
	}
	for _, rule := range req.Rules {
		itemRule := dto.AddrRuleOperate{Operation: rule.Operation, Address: rule.Address, Strategy: rule.Strategy}
		if err := u.OperateAddressRule(itemRule, false); err != nil {
			return err
		}
	}
	return client.Reload()
}

func (u *FirewallService) operatePort(client firewall.FirewallClient, req dto.PortRuleOperate) error {
	var fireInfo fireClient.FireInfo
	if err := copier.Copy(&fireInfo, &req); err != nil {
		return err
	}

	if client.Name() == "ufw" {
		if len(fireInfo.Address) != 0 && fireInfo.Address != "Anywhere" {
			return client.RichRules(fireInfo, req.Operation)
		}
		return client.Port(fireInfo, req.Operation)
	}

	if len(fireInfo.Address) != 0 || fireInfo.Strategy == "drop" {
		return client.RichRules(fireInfo, req.Operation)
	}
	return client.Port(fireInfo, req.Operation)
}
