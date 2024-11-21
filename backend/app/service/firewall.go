package service

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/firewall"
	fireClient "github.com/1Panel-dev/1Panel/backend/utils/firewall/client"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

const confPath = "/etc/sysctl.conf"

type FirewallService struct{}

type IFirewallService interface {
	LoadBaseInfo() (dto.FirewallBaseInfo, error)
	SearchWithPage(search dto.RuleSearch) (int64, interface{}, error)
	OperateFirewall(operation string) error
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

func (u *FirewallService) LoadBaseInfo() (dto.FirewallBaseInfo, error) {
	var baseInfo dto.FirewallBaseInfo
	baseInfo.Status = "not running"
	baseInfo.Version = "-"
	baseInfo.Name = "-"
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return baseInfo, err
	}
	baseInfo.Name = client.Name()

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		baseInfo.PingStatus = u.pingStatus()
	}()
	go func() {
		defer wg.Done()
		baseInfo.Status, _ = client.Status()
	}()
	go func() {
		defer wg.Done()
		baseInfo.Version, _ = client.Version()
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
			if backDatas[i].Port == des.Port &&
				req.Type == "port" &&
				backDatas[i].Protocol == des.Protocol &&
				backDatas[i].Strategy == des.Strategy &&
				backDatas[i].Address == des.Address {
				backDatas[i].Description = des.Description
				break
			}
			if req.Type == "address" && backDatas[i].Strategy == des.Strategy && backDatas[i].Address == des.Address {
				backDatas[i].Description = des.Description
				break
			}
		}
	}

	go u.cleanUnUsedData(client)

	return int64(total), backDatas, nil
}

func (u *FirewallService) OperateFirewall(operation string) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	needRestartDocker := false
	switch operation {
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
		return fmt.Errorf("not supported operation: %s", operation)
	}
	if needRestartDocker {
		if err := restartDocker(); err != nil {
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
						proto == rule.Protocol {
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
					proto == rule.Protocol {
					return constant.ErrRecordExist
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
			}, r.Operation); err != nil {
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
		if err := u.addAddressRecord(req); err != nil {
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
	var firewall model.Firewall
	if err := copier.Copy(&firewall, &req); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
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
	apps, err := appInstallRepo.ListBy()
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
			if records[i].Port == item.Port && records[i].Protocol == item.Protocol && records[i].Strategy == item.Strategy && records[i].Address == item.Address {
				records = append(records[:i], records[i+1:]...)
			}
		}
	}

	for _, record := range records {
		_ = hostRepo.DeleteFirewallRecordByID(record.ID)
	}
}
func (u *FirewallService) pingStatus() string {
	if _, err := os.Stat("/etc/sysctl.conf"); err != nil {
		return constant.StatusNone
	}
	sudo := cmd.SudoHandleCmd()
	command := fmt.Sprintf("%s cat /etc/sysctl.conf | grep net/ipv4/icmp_echo_ignore_all= ", sudo)
	stdout, _ := cmd.Exec(command)
	if stdout == "net/ipv4/icmp_echo_ignore_all=1\n" {
		return constant.StatusEnable
	}
	return constant.StatusDisable
}

func (u *FirewallService) updatePingStatus(enable string) error {
	lineBytes, err := os.ReadFile(confPath)
	if err != nil {
		return err
	}
	files := strings.Split(string(lineBytes), "\n")
	var newFiles []string
	hasLine := false
	for _, line := range files {
		if strings.Contains(line, "net/ipv4/icmp_echo_ignore_all") || strings.HasPrefix(line, "net/ipv4/icmp_echo_ignore_all") {
			newFiles = append(newFiles, "net/ipv4/icmp_echo_ignore_all="+enable)
			hasLine = true
		} else {
			newFiles = append(newFiles, line)
		}
	}
	if !hasLine {
		newFiles = append(newFiles, "net/ipv4/icmp_echo_ignore_all="+enable)
	}
	file, err := os.OpenFile(confPath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(strings.Join(newFiles, "\n"))
	if err != nil {
		return err
	}

	sudo := cmd.SudoHandleCmd()
	command := fmt.Sprintf("%s sysctl -p", sudo)
	stdout, err := cmd.Exec(command)
	if err != nil {
		return fmt.Errorf("update ping status failed, err: %v", stdout)
	}

	return nil
}

func (u *FirewallService) addPortsBeforeStart(client firewall.FirewallClient) error {
	serverPort, err := settingRepo.Get(settingRepo.WithByKey("ServerPort"))
	if err != nil {
		return err
	}
	if err := client.Port(fireClient.FireInfo{Port: serverPort.Value, Protocol: "tcp", Strategy: "accept"}, "add"); err != nil {
		return err
	}
	if err := client.Port(fireClient.FireInfo{Port: "22", Protocol: "tcp", Strategy: "accept"}, "add"); err != nil {
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
		return hostRepo.DeleteFirewallRecord("port", req.Port, req.Protocol, req.Address, req.Strategy)
	}

	if err := hostRepo.SaveFirewallRecord(&model.Firewall{
		Type:        "port",
		Port:        req.Port,
		Protocol:    req.Protocol,
		Address:     req.Address,
		Strategy:    req.Strategy,
		Description: req.Description,
	}); err != nil {
		return fmt.Errorf("add record %s/%s failed (strategy: %s, address: %s), err: %v", req.Port, req.Protocol, req.Strategy, req.Address, err)
	}

	return nil
}

func (u *FirewallService) addAddressRecord(req dto.AddrRuleOperate) error {
	if req.Operation == "remove" {
		return hostRepo.DeleteFirewallRecord("address", "", "", req.Address, req.Strategy)
	}
	if err := hostRepo.SaveFirewallRecord(&model.Firewall{
		Type:        "address",
		Address:     req.Address,
		Strategy:    req.Strategy,
		Description: req.Description,
	}); err != nil {
		return fmt.Errorf("add record failed (strategy: %s, address: %s), err: %v", req.Strategy, req.Address, err)
	}
	return nil
}

func checkPortUsed(ports, proto string, apps []portOfApp) string {
	var portList []int
	if strings.Contains(ports, "-") || strings.Contains(ports, ",") {
		if strings.Contains(ports, "-") {
			port1, err := strconv.Atoi(strings.Split(ports, "-")[0])
			if err != nil {
				global.LOG.Errorf(" convert string %s to int failed, err: %v", strings.Split(ports, "-")[0], err)
				return ""
			}
			port2, err := strconv.Atoi(strings.Split(ports, "-")[1])
			if err != nil {
				global.LOG.Errorf(" convert string %s to int failed, err: %v", strings.Split(ports, "-")[1], err)
				return ""
			}
			for i := port1; i <= port2; i++ {
				portList = append(portList, i)
			}
		} else {
			portLists := strings.Split(ports, ",")
			for _, item := range portLists {
				portItem, _ := strconv.Atoi(item)
				portList = append(portList, portItem)
			}
		}

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
