package service

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	fireClient "github.com/1Panel-dev/1Panel/agent/utils/firewall/client"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client/iptables"
	"github.com/jinzhu/copier"
)

const confPath = "/etc/sysctl.conf"
const panelSysctlPath = "/etc/sysctl.d/98-onepanel.conf"

type FirewallService struct{}

type IFirewallService interface {
	LoadBaseInfo(tab string) (dto.FirewallBaseInfo, error)
	SearchWithPage(search dto.RuleSearch) (int64, interface{}, error)
	OperateFirewall(req dto.FirewallOperation) error
	OperatePortRule(req dto.PortRuleOperate, reload bool) error
	OperateForwardRule(req dto.ForwardRuleOperate) error
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
		baseInfo.PingStatus = u.pingStatus()
		baseInfo.Version, _ = client.Version()
	}()
	go func() {
		defer wg.Done()
		baseInfo.IsActive, _ = client.Status()
		baseInfo.IsInit, baseInfo.IsBind = loadInitStatus(baseInfo.Name, tab)
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
	case "forward":
		rules, err = client.ListForward()
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

	var datasFilterStatus []fireClient.FireInfo
	if len(req.Status) != 0 {
		for _, data := range datas {
			if req.Status == "free" && len(data.UsedStatus) == 0 {
				datasFilterStatus = append(datasFilterStatus, data)
			}
			if req.Status == "used" && len(data.UsedStatus) != 0 {
				datasFilterStatus = append(datasFilterStatus, data)
			}
		}
	} else {
		datasFilterStatus = datas
	}

	var datasFilterStrategy []fireClient.FireInfo
	if len(req.Strategy) != 0 {
		for _, data := range datasFilterStatus {
			if req.Strategy == data.Strategy {
				datasFilterStrategy = append(datasFilterStrategy, data)
			}
		}
	} else {
		datasFilterStrategy = datasFilterStatus
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

	go u.cleanUnUsedData(client)

	return int64(total), backDatas, nil
}

func (u *FirewallService) OperateFirewall(req dto.FirewallOperation) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
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
		needRestartDocker = true
	case "disablePing":
		return u.updatePingStatus("0")
	case "enablePing":
		return u.updatePingStatus("1")
	default:
		return fmt.Errorf("not supported operation: %s", req.Operation)
	}
	if needRestartDocker && req.WithDockerRestart {
		if err := controller.HandleRestart("docker"); err != nil {
			return fmt.Errorf("failed to restart Docker: %v", err)
		}
	}
	return nil
}

func (u *FirewallService) OperatePortRule(req dto.PortRuleOperate, reload bool) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	if len(req.Chain) == 0 && client.Name() == "iptables" {
		req.Chain = iptables.Chain1PanelBasic
	}
	protos := strings.Split(req.Protocol, "/")
	itemAddress := strings.Split(strings.TrimSuffix(req.Address, ","), ",")

	if client.Name() == "ufw" {
		if strings.Contains(req.Port, ",") || strings.Contains(req.Port, "-") {
			for _, proto := range protos {
				for _, addr := range itemAddress {
					if len(addr) == 0 {
						addr = "Anywhere"
					}
					req.Address = addr
					req.Port = strings.ReplaceAll(req.Port, "-", ":")
					req.Protocol = proto
					if err := u.operatePort(client, req); err != nil {
						return err
					}
					req.Port = strings.ReplaceAll(req.Port, ":", "-")
					if err := u.addPortRecord(req); err != nil {
						return err
					}
				}
			}
			return nil
		}
		for _, addr := range itemAddress {
			if len(addr) == 0 {
				addr = "Anywhere"
			}
			if req.Protocol == "tcp/udp" {
				req.Protocol = ""
			}
			req.Address = addr
			if err := u.operatePort(client, req); err != nil {
				return err
			}
			if len(req.Protocol) == 0 {
				req.Protocol = "tcp/udp"
			}
			if err := u.addPortRecord(req); err != nil {
				return err
			}
		}
		return nil
	}

	itemPorts := req.Port
	for _, proto := range protos {
		if strings.Contains(req.Port, "-") {
			for _, addr := range itemAddress {
				req.Protocol = proto
				req.Address = addr
				if err := u.operatePort(client, req); err != nil {
					return err
				}
				if err := u.addPortRecord(req); err != nil {
					return err
				}
			}
		} else {
			ports := strings.Split(itemPorts, ",")
			for _, port := range ports {
				if len(port) == 0 {
					continue
				}
				for _, addr := range itemAddress {
					req.Address = addr
					req.Port = port
					req.Protocol = proto
					if err := u.operatePort(client, req); err != nil {
						return err
					}
					if err := u.addPortRecord(req); err != nil {
						return err
					}
				}
			}
		}
	}

	if reload {
		return client.Reload()
	}
	return nil
}

func (u *FirewallService) OperateForwardRule(req dto.ForwardRuleOperate) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}

	rules, _ := client.ListForward()
	i := 0
	for _, rule := range rules {
		shouldKeep := true
		for i := range req.Rules {
			reqRule := &req.Rules[i]
			if reqRule.TargetIP == "" {
				reqRule.TargetIP = "127.0.0.1"
			}

			if reqRule.Operation == "remove" {
				for _, proto := range strings.Split(reqRule.Protocol, "/") {
					if reqRule.Port == rule.Port &&
						reqRule.TargetPort == rule.TargetPort &&
						reqRule.TargetIP == rule.TargetIP &&
						proto == rule.Protocol &&
						reqRule.Interface == rule.Interface {
						shouldKeep = false
						break
					}
				}
			}
		}
		if shouldKeep {
			rules[i] = rule
			i++
		}
	}
	rules = rules[:i]

	for _, rule := range rules {
		for _, reqRule := range req.Rules {
			if reqRule.Operation == "remove" {
				continue
			}

			for _, proto := range strings.Split(reqRule.Protocol, "/") {
				if reqRule.Port == rule.Port &&
					reqRule.TargetPort == rule.TargetPort &&
					reqRule.TargetIP == rule.TargetIP &&
					proto == rule.Protocol &&
					reqRule.Interface == rule.Interface {
					return buserr.New("ErrRecordExist")
				}
			}
		}
	}

	sort.SliceStable(req.Rules, func(i, j int) bool {
		if req.Rules[i].Operation == "remove" && req.Rules[j].Operation != "remove" {
			return true
		}
		if req.Rules[i].Operation != "remove" && req.Rules[j].Operation == "remove" {
			return false
		}
		n1, _ := strconv.Atoi(req.Rules[i].Num)
		n2, _ := strconv.Atoi(req.Rules[j].Num)
		return n1 > n2
	})

	for _, r := range req.Rules {
		for _, p := range strings.Split(r.Protocol, "/") {
			if r.TargetIP == "" {
				r.TargetIP = "127.0.0.1"
			}
			if err = client.PortForward(fireClient.Forward{
				Num:        r.Num,
				Protocol:   p,
				Port:       r.Port,
				TargetIP:   r.TargetIP,
				TargetPort: r.TargetPort,
				Interface:  r.Interface,
			}, r.Operation); err != nil {
				if req.ForceDelete {
					global.LOG.Error(err)
					continue
				}
				return err
			}
		}
	}
	return nil
}

func (u *FirewallService) OperateAddressRule(req dto.AddrRuleOperate, reload bool) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	chain := ""
	if client.Name() == "iptables" {
		chain = iptables.Chain1PanelBasic
	}
	var fireInfo fireClient.FireInfo
	if err := copier.Copy(&fireInfo, &req); err != nil {
		return err
	}

	addressList := strings.Split(req.Address, ",")
	for i := 0; i < len(addressList); i++ {
		if len(addressList[i]) == 0 {
			continue
		}
		fireInfo.Address = addressList[i]
		if err := client.RichRules(fireInfo, req.Operation); err != nil {
			return err
		}
		req.Address = addressList[i]
		if err := u.addAddressRecord(chain, req); err != nil {
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

func (u *FirewallService) operatePort(client firewall.FirewallClient, req dto.PortRuleOperate) error {
	var fireInfo fireClient.FireInfo
	if err := copier.Copy(&fireInfo, &req); err != nil {
		return err
	}

	if client.Name() == "ufw" {
		if len(fireInfo.Address) != 0 && !strings.EqualFold(fireInfo.Address, "Anywhere") {
			return client.RichRules(fireInfo, req.Operation)
		}
		return client.Port(fireInfo, req.Operation)
	}

	if len(fireInfo.Address) != 0 || fireInfo.Strategy == "drop" {
		return client.RichRules(fireInfo, req.Operation)
	}
	return client.Port(fireInfo, req.Operation)
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

func (u *FirewallService) cleanUnUsedData(client firewall.FirewallClient) {
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

func (u *FirewallService) pingStatus() string {
	data, err := os.ReadFile("/proc/sys/net/ipv4/icmp_echo_ignore_all")
	if err != nil {
		return constant.StatusNone
	}
	v6Data, v6err := os.ReadFile("/proc/sys/net/ipv6/icmp/echo_ignore_all")
	if v6err != nil {
		if strings.TrimSpace(string(data)) == "1" {
			return constant.StatusEnable
		}
		return constant.StatusDisable
	} else {
		if strings.TrimSpace(string(data)) == "1" && strings.TrimSpace(string(v6Data)) == "1" {
			return constant.StatusEnable
		}
		return constant.StatusDisable
	}

}

func (u *FirewallService) updatePingStatus(enable string) error {
	var targetPath string
	var applyCmd string

	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		// Debian 13
		targetPath = panelSysctlPath
		applyCmd = fmt.Sprintf("%s sysctl --system", cmd.SudoHandleCmd())
		if err := cmd.RunDefaultBashCf("%s mkdir -p /etc/sysctl.d", cmd.SudoHandleCmd()); err != nil {
			return fmt.Errorf("failed to create directory /etc/sysctl.d: %v", err)
		}
	} else {
		targetPath = confPath
		applyCmd = fmt.Sprintf("%s sysctl -p", cmd.SudoHandleCmd())
	}

	lineBytes, err := os.ReadFile(targetPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %v", targetPath, err)
	}

	if err := cmd.RunDefaultBashCf("echo %s | %s tee /proc/sys/net/ipv4/icmp_echo_ignore_all > /dev/null", enable, cmd.SudoHandleCmd()); err != nil {
		return fmt.Errorf("failed to apply ipv4 ping status temporarily: %v", err)
	}

	var hasIpv6 bool
	if _, err := os.Stat("/proc/sys/net/ipv6/icmp/echo_ignore_all"); err == nil {
		hasIpv6 = true
		if err := cmd.RunDefaultBashCf("echo %s | %s tee /proc/sys/net/ipv6/icmp/echo_ignore_all > /dev/null", enable, cmd.SudoHandleCmd()); err != nil {
			global.LOG.Warnf("failed to apply ipv6 ping status temporarily: %v", err)
		}
	}

	var files []string
	if err == nil {
		files = strings.Split(string(lineBytes), "\n")
	}

	var newFiles []string
	hasIPv4Line, hasIPv6Line := false, false

	for _, line := range files {
		if strings.HasPrefix(strings.TrimSpace(line), "net.ipv4.icmp_echo_ignore_all") {
			newFiles = append(newFiles, "net.ipv4.icmp_echo_ignore_all="+enable)
			hasIPv4Line = true
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "net.ipv6.icmp.echo_ignore_all") {
			newFiles = append(newFiles, "net.ipv6.icmp.echo_ignore_all="+enable)
			hasIPv6Line = true
			continue
		}
		newFiles = append(newFiles, line)
	}

	if !hasIPv4Line {
		newFiles = append(newFiles, "net.ipv4.icmp_echo_ignore_all="+enable)
	}
	if hasIpv6 && !hasIPv6Line {
		newFiles = append(newFiles, "net.ipv6.icmp.echo_ignore_all="+enable)
	}

	file, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, constant.FilePerm)
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", targetPath, err)
	}
	defer file.Close()

	if _, err = file.WriteString(strings.Join(newFiles, "\n")); err != nil {
		return fmt.Errorf("failed to write to %s: %v", targetPath, err)
	}

	if err := cmd.RunDefaultBashC(applyCmd); err != nil {
		global.LOG.Warnf("failed to apply persistent config with '%s': %v", applyCmd, err)
	}

	return nil
}

func (u *FirewallService) addPortsBeforeStart(client firewall.FirewallClient) error {
	if !global.IsMaster {
		if err := client.Port(fireClient.FireInfo{Port: global.CONF.Base.Port, Protocol: "tcp", Strategy: "accept"}, "add"); err != nil {
			return err
		}
	} else {
		var portSetting model.Setting
		_ = global.CoreDB.Where("key = ?", "ServerPort").First(&portSetting).Error
		if len(portSetting.Value) != 0 {
			if err := client.Port(fireClient.FireInfo{Port: portSetting.Value, Protocol: "tcp", Strategy: "accept"}, "add"); err != nil {
				return err
			}
		}
	}
	if err := client.Port(fireClient.FireInfo{Port: loadSSHPort(), Protocol: "tcp", Strategy: "accept"}, "add"); err != nil {
		return err
	}
	if err := client.Port(fireClient.FireInfo{Port: "80", Protocol: "tcp", Strategy: "accept"}, "add"); err != nil {
		return err
	}
	if err := client.Port(fireClient.FireInfo{Port: "443", Protocol: "tcp", Strategy: "accept"}, "add"); err != nil {
		return err
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
			return fmt.Sprintf("(%s)", app.AppName)
		}
	}
	port, err := strconv.Atoi(ports)
	if err != nil {
		global.LOG.Errorf(" convert string %v to int failed, err: %v", port, err)
		return ""
	}
	if common.ScanPortWithProto(port, proto) {
		return "inUsed"
	}
	return ""
}

func loadInitStatus(clientName, tab string) (bool, bool) {
	if clientName == "firewalld" {
		return true, true
	}
	if clientName == "ufw" && tab != "forward" {
		return true, true
	}
	switch tab {
	case "base":
		if isExist, _ := iptables.CheckChainExist(iptables.FilterTab, iptables.Chain1PanelBasicBefore); !isExist {
			return false, false
		}
		if exist := iptables.CheckRuleExist(iptables.FilterTab, iptables.Chain1PanelBasicBefore, iptables.IoRuleIn); !exist {
			return false, false
		}
		if exist := iptables.CheckRuleExist(iptables.FilterTab, iptables.Chain1PanelBasicBefore, iptables.EstablishedRule); !exist {
			return false, false
		}
		if exist, _ := iptables.CheckChainExist(iptables.FilterTab, iptables.Chain1PanelBasic); !exist {
			return false, false
		}
		if exist, _ := iptables.CheckChainExist(iptables.FilterTab, iptables.Chain1PanelBasicAfter); !exist {
			return false, false
		}
		if exist := iptables.CheckRuleExist(iptables.FilterTab, iptables.Chain1PanelBasicAfter, iptables.DropAllTcp); !exist {
			return false, false
		}
		if exist := iptables.CheckRuleExist(iptables.FilterTab, iptables.Chain1PanelBasicAfter, iptables.DropAllUdp); !exist {
			return false, false
		}
		if bind, _ := iptables.CheckChainBind(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasicBefore); !bind {
			return true, false
		}
		if bind, _ := iptables.CheckChainBind(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasic); !bind {
			return true, false
		}
		if bind, _ := iptables.CheckChainBind(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasicAfter); !bind {
			return true, false
		}
		return true, true
	case "advance":
		isExist, _ := iptables.CheckChainExist(iptables.FilterTab, iptables.Chain1PanelInput)
		if !isExist {
			return false, false
		}
		isExist, _ = iptables.CheckChainExist(iptables.FilterTab, iptables.Chain1PanelOutput)
		if !isExist {
			return false, false
		}

		isBind, _ := iptables.CheckChainBind(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelInput)
		if !isBind {
			return true, false
		}
		isBind, _ = iptables.CheckChainBind(iptables.FilterTab, iptables.ChainOutput, iptables.Chain1PanelOutput)
		return true, isBind
	case "forward":
		stdout, err := cmd.RunDefaultWithStdoutBashC("cat /proc/sys/net/ipv4/ip_forward")
		if err != nil {
			global.LOG.Errorf("check /proc/sys/net/ipv4/ip_forward failed, err: %v", err)
			return false, false
		}
		if strings.TrimSpace(stdout) == "0" {
			return false, false
		}

		exist, _ := iptables.CheckChainExist(iptables.NatTab, iptables.Chain1PanelPreRouting)
		if !exist {
			return false, false
		}
		exist, _ = iptables.CheckChainExist(iptables.NatTab, iptables.Chain1PanelPostRouting)
		if !exist {
			return false, false
		}
		exist, _ = iptables.CheckChainExist(iptables.FilterTab, iptables.Chain1PanelForward)
		if !exist {
			return false, false
		}
		isBind, _ := iptables.CheckChainBind(iptables.NatTab, "PREROUTING", iptables.Chain1PanelPreRouting)
		if !isBind {
			return false, false
		}
		isBind, _ = iptables.CheckChainBind(iptables.NatTab, "POSTROUTING", iptables.Chain1PanelPostRouting)
		if !isBind {
			return false, false
		}
		isBind, _ = iptables.CheckChainBind(iptables.FilterTab, "FORWARD", iptables.Chain1PanelForward)
		return true, isBind
	default:
		return false, false
	}
}
