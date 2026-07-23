package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	fireClient "github.com/1Panel-dev/1Panel/agent/utils/firewall/client"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client/iptables"
	"github.com/jinzhu/copier"
)

type FirewallService struct{}

type IFirewallService interface {
	LoadBaseInfo(tab string) (dto.FirewallBaseInfo, error)
	SearchWithPage(search dto.RuleSearch) (int64, interface{}, error)
	OperateFirewall(req dto.FirewallOperation) error
	OperatePortRule(req dto.PortRuleOperate, reload bool) error
	OperateAddressRule(req dto.AddrRuleOperate, reload bool) error
	UpdatePortRule(req dto.PortRuleUpdate) error
	UpdateAddrRule(req dto.AddrRuleUpdate) error
	UpdateDescription(req dto.UpdateFirewallDescription) error
	BatchOperateRule(req dto.BatchRuleOperate) error
}

func NewIFirewallService() IFirewallService {
	return &FirewallService{}
}

func (u *FirewallService) LoadBaseInfo(tab string) (dto.FirewallBaseInfo, error) {
	var baseInfo dto.FirewallBaseInfo
	baseInfo.Version = "-"
	baseInfo.Name = "-"
	client, err := firewall.NewFirewallClient()
	if err != nil {
		global.LOG.Errorf("load firewall failed, err: %v", err)
		baseInfo.IsExist = false
		return baseInfo, nil
	}
	baseInfo.IsExist = true
	baseInfo.Name = client.Name()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		baseInfo.PingStatus = firewall.LoadPingStatus()
		baseInfo.Version, _ = client.Version()
	}()
	go func() {
		defer wg.Done()
		baseInfo.IsActive, _ = client.Status()
		baseInfo.IsInit, baseInfo.IsBind = iptables.LoadInitStatus(baseInfo.Name, tab)
	}()
	wg.Wait()
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

	var rules []fireClient.FireInfo
	switch req.Type {
	case "port":
		rules, err = client.ListPort()
	case "address":
		rules, err = client.ListAddress()
	}
	if err != nil {
		return 0, nil, err
	}

	if len(req.Info) != 0 {
		for _, addr := range rules {
			if strings.Contains(addr.Address, req.Info) ||
				strings.Contains(addr.Port, req.Info) ||
				strings.Contains(addr.TargetPort, req.Info) ||
				strings.Contains(addr.TargetIP, req.Info) {
				datas = append(datas, addr)
			}
		}
	} else {
		datas = rules
	}
	if req.Type == "port" {
		apps := u.loadPortByApp()
		for i := 0; i < len(datas); i++ {
			datas[i].UsedStatus = checkPortUsed(datas[i].Port, datas[i].Protocol, apps)
		}
	}

	var datasFilterStrategy []fireClient.FireInfo
	if len(req.Strategy) != 0 {
		for _, data := range datas {
			if req.Strategy == data.Strategy {
				datasFilterStrategy = append(datasFilterStrategy, data)
			}
		}
	} else {
		datasFilterStrategy = datas
	}

	total, start, end := len(datasFilterStrategy), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		backDatas = make([]fireClient.FireInfo, 0)
	} else {
		if end >= total {
			end = total
		}
		backDatas = datasFilterStrategy[start:end]
	}

	datasFromDB, _ := hostRepo.ListFirewallRecord()
	for i := 0; i < len(backDatas); i++ {
		for _, des := range datasFromDB {
			if req.Type != des.Type {
				continue
			}
			if backDatas[i].Port == des.DstPort &&
				req.Type == "port" &&
				backDatas[i].Protocol == des.Protocol &&
				backDatas[i].Strategy == des.Strategy &&
				backDatas[i].Address == des.SrcIP {
				backDatas[i].ID = des.ID
				backDatas[i].Description = des.Description
				break
			}
			if req.Type == "address" && backDatas[i].Strategy == des.Strategy && backDatas[i].Address == des.SrcIP {
				backDatas[i].ID = des.ID
				backDatas[i].Description = des.Description
				break
			}
		}
	}

	if req.Type == "port" || req.Type == "address" {
		go u.cleanUnUsedData(client)
	}

	return int64(total), backDatas, nil
}

func (u *FirewallService) OperateFirewall(req dto.FirewallOperation) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	fail2BanState := newFirewallFail2BanState()
	if client.Name() == "firewalld" && req.Operation == "stop" {
		if err := fail2BanState.rememberBeforeFirewallStop(); err != nil {
			return err
		}
	}
	needRestartDocker := false
	switch req.Operation {
	case "start":
		if err := client.Start(); err != nil {
			return err
		}
		if err := u.addPortsBeforeStart(client); err != nil {
			_ = client.Stop()
			return err
		}
		needRestartDocker = true
	case "stop":
		if err := client.Stop(); err != nil {
			return err
		}
		needRestartDocker = true
	case "restart":
		if err := client.Restart(); err != nil {
			return err
		}
		if err := u.addPortsBeforeStart(client); err != nil {
			return err
		}
		needRestartDocker = true
	case "disableBanPing":
		if err := firewall.UpdatePingStatus("0"); err != nil {
			return err
		}
		_ = settingRepo.Update("BanPing", constant.StatusDisable)
		return nil
	case "enableBanPing":
		if err := firewall.UpdatePingStatus("1"); err != nil {
			return err
		}
		_ = settingRepo.Update("BanPing", constant.StatusEnable)
		return nil
	default:
		return fmt.Errorf("not supported operation: %s", req.Operation)
	}
	if needRestartDocker && req.WithDockerRestart {
		if err := controller.HandleRestart("docker"); err != nil {
			return fmt.Errorf("failed to restart Docker: %v", err)
		}
	}
	if client.Name() == "firewalld" && req.Operation == "start" {
		if err := fail2BanState.restoreAfterFirewallStart(); err != nil {
			return err
		}
	}
	return nil
}

func (u *FirewallService) OperatePortRule(req dto.PortRuleOperate, reload bool) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	return u.operatePortRuleWithClient(client, req, reload)
}

func (u *FirewallService) operatePortRuleWithClient(client firewall.FilterClient, req dto.PortRuleOperate, reload bool) error {
	var rule fireClient.FireInfo
	if err := copier.Copy(&rule, &req); err != nil {
		return err
	}
	for _, unit := range client.ExpandPortRule(rule) {
		if err := client.ApplyPortUnit(unit, req.Operation); err != nil {
			return err
		}
		record := req
		record.Chain = unit.Chain
		record.Address = unit.Record.Address
		record.Port = unit.Record.Port
		record.Protocol = unit.Record.Protocol
		record.Strategy = unit.Record.Strategy
		if err := u.addPortRecord(record); err != nil {
			return err
		}
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
	return u.operateAddressRuleWithClient(client, req, reload)
}

func (u *FirewallService) operateAddressRuleWithClient(client firewall.FilterClient, req dto.AddrRuleOperate, reload bool) error {
	var rule fireClient.FireInfo
	if err := copier.Copy(&rule, &req); err != nil {
		return err
	}
	for _, unit := range client.ExpandAddressRule(rule) {
		if err := client.ApplyAddressUnit(unit, req.Operation); err != nil {
			return err
		}
		record := req
		record.Address = unit.Apply.Address
		if err := u.addAddressRecord(unit.Chain, record); err != nil {
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

func (u *FirewallService) UpdateDescription(req dto.UpdateFirewallDescription) error {
	firewall := model.Firewall{
		Type:        req.Type,
		Chain:       req.Chain,
		SrcIP:       req.SrcIP,
		DstIP:       req.DstIP,
		SrcPort:     req.SrcPort,
		DstPort:     req.DstPort,
		Protocol:    req.Protocol,
		Strategy:    req.Strategy,
		Description: req.Description,
	}

	return hostRepo.SaveFirewallRecord(&firewall)
}

func (u *FirewallService) BatchOperateRule(req dto.BatchRuleOperate) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	if req.Type == "port" {
		for _, rule := range req.Rules {
			_ = u.OperatePortRule(rule, false)
		}
		return client.Reload()
	}
	for _, rule := range req.Rules {
		itemRule := dto.AddrRuleOperate{Operation: rule.Operation, Address: rule.Address, Strategy: rule.Strategy}
		_ = u.OperateAddressRule(itemRule, false)
	}
	return client.Reload()
}

func OperateFirewallPort(oldPorts, newPorts []int) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	return operateFirewallPorts(client, oldPorts, newPorts)
}

func operateFirewallPorts(client firewall.FilterClient, oldPorts, newPorts []int) error {
	for _, port := range newPorts {
		if err := client.Port(fireClient.FireInfo{Port: strconv.Itoa(port), Protocol: "tcp", Strategy: "accept"}, "add"); err != nil {
			return err
		}
	}
	for _, port := range oldPorts {
		if err := client.Port(fireClient.FireInfo{Port: strconv.Itoa(port), Protocol: "tcp", Strategy: "accept"}, "remove"); err != nil {
			return err
		}
	}
	return client.Reload()
}

type portOfApp struct {
	AppName   string
	HttpPort  string
	HttpsPort string
}

func (u *FirewallService) loadPortByApp() []portOfApp {
	var datas []portOfApp
	apps, err := appInstallRepo.ListBy(context.Background())
	if err != nil {
		return datas
	}
	for i := 0; i < len(apps); i++ {
		datas = append(datas, portOfApp{
			AppName:   apps[i].App.Key,
			HttpPort:  strconv.Itoa(apps[i].HttpPort),
			HttpsPort: strconv.Itoa(apps[i].HttpsPort),
		})
	}
	systemPort, err := settingRepo.Get(settingRepo.WithByKey("ServerPort"))
	if err != nil {
		return datas
	}
	datas = append(datas, portOfApp{AppName: "1panel", HttpPort: systemPort.Value})

	return datas
}

func (u *FirewallService) cleanUnUsedData(client firewall.FilterClient) {
	list, _ := client.ListPort()
	addressList, _ := client.ListAddress()
	list = append(list, addressList...)
	if len(list) == 0 {
		return
	}
	records, _ := hostRepo.ListFirewallRecord()
	if len(records) == 0 {
		return
	}
	for _, item := range list {
		for i := 0; i < len(records); i++ {
			if records[i].DstPort == item.Port && records[i].Protocol == item.Protocol && records[i].Strategy == item.Strategy && records[i].SrcIP == item.Address {
				records = append(records[:i], records[i+1:]...)
			}
		}
	}

	for _, record := range records {
		_ = hostRepo.DeleteFirewallRecordByID(record.ID)
	}
}

func (u *FirewallService) addPortsBeforeStart(client firewall.FilterClient) error {
	if client.Name() == "iptables" {
		isInit, _ := iptables.LoadInitStatus("iptables", "base")
		if !isInit {
			return nil
		}
		return syncIptablesFirewallPortWhiteList(true)
	}
	portWhiteList, err := loadFirewallPortWhiteList()
	if err != nil {
		return err
	}
	for _, item := range portWhiteList {
		if err := client.Port(fireClient.FireInfo{Port: item.Port, Protocol: item.Protocol, Strategy: "accept"}, "add"); err != nil {
			return err
		}
	}

	return client.Reload()
}

func (u *FirewallService) addPortRecord(req dto.PortRuleOperate) error {
	if req.Operation == "remove" {
		if req.ID != 0 {
			return hostRepo.DeleteFirewallRecordByID(req.ID)
		}
		return nil
	}

	if len(req.Description) == 0 {
		return nil
	}
	if err := hostRepo.SaveFirewallRecord(&model.Firewall{
		Type:        "port",
		Chain:       req.Chain,
		DstPort:     req.Port,
		Protocol:    req.Protocol,
		SrcIP:       req.Address,
		Strategy:    req.Strategy,
		Description: req.Description,
	}); err != nil {
		return fmt.Errorf("add record %s/%s failed (strategy: %s, address: %s), err: %v", req.Port, req.Protocol, req.Strategy, req.Address, err)
	}

	return nil
}

func (u *FirewallService) addAddressRecord(chain string, req dto.AddrRuleOperate) error {
	if req.Operation == "remove" {
		if req.ID != 0 {
			return hostRepo.DeleteFirewallRecordByID(req.ID)
		}
		return nil
	}

	if err := hostRepo.SaveFirewallRecord(&model.Firewall{
		Type:        "address",
		Chain:       chain,
		SrcIP:       req.Address,
		Strategy:    req.Strategy,
		Description: req.Description,
	}); err != nil {
		return fmt.Errorf("add record failed (strategy: %s, address: %s), err: %v", req.Strategy, req.Address, err)
	}
	return nil
}

func checkPortUsed(ports, proto string, apps []portOfApp) string {
	var portList []int
	rangeSplit := ""
	if strings.Contains(ports, "-") {
		rangeSplit = "-"
	}
	if strings.Contains(ports, ":") {
		rangeSplit = ":"
	}
	if len(rangeSplit) != 0 {
		port1, err := strconv.Atoi(strings.Split(ports, rangeSplit)[0])
		if err != nil {
			global.LOG.Errorf(" convert string %s to int failed, err: %v", strings.Split(ports, rangeSplit)[0], err)
			return ""
		}
		port2, err := strconv.Atoi(strings.Split(ports, rangeSplit)[1])
		if err != nil {
			global.LOG.Errorf(" convert string %s to int failed, err: %v", strings.Split(ports, rangeSplit)[1], err)
			return ""
		}
		for i := port1; i <= port2; i++ {
			portList = append(portList, i)
		}
	}
	if strings.Contains(ports, ",") {
		portLists := strings.Split(ports, ",")
		for _, item := range portLists {
			portItem, _ := strconv.Atoi(item)
			portList = append(portList, portItem)
		}
	}
	if len(portList) != 0 {
		var usedPorts []string
		for _, port := range portList {
			portItem := fmt.Sprintf("%v", port)
			isUsedByApp := false
			for _, app := range apps {
				if app.HttpPort == portItem || app.HttpsPort == portItem {
					isUsedByApp = true
					usedPorts = append(usedPorts, fmt.Sprintf("%s (%s)", portItem, app.AppName))
					break
				}
			}
			if !isUsedByApp && common.ScanPortWithProto(port, proto) {
				usedPorts = append(usedPorts, fmt.Sprintf("%v", port))
			}
		}
		return strings.Join(usedPorts, ",")
	}

	for _, app := range apps {
		if app.HttpPort == ports || app.HttpsPort == ports {
			return app.AppName
		}
	}

	return ""
}
