package service

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
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
	baseInfo.PingStatus = u.pingStatus()
	baseInfo.Status = "not running"
	baseInfo.Version = "-"
	baseInfo.Name = "-"
	client, err := firewall.NewFirewallClient()
	if err != nil {
		if err.Error() == "no such type" {
			return baseInfo, nil
		}
		return baseInfo, err
	}
	baseInfo.Name = client.Name()
	baseInfo.Status, err = client.Status()
	if err != nil {
		return baseInfo, err
	}
	if baseInfo.Status == "not running" {
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

	datasFromDB, _ := hostRepo.ListFirewallRecord()
	for i := 0; i < len(backDatas); i++ {
		for _, des := range datasFromDB {
			if backDatas[i].Port == des.Port && backDatas[i].Protocol == des.Protocol && backDatas[i].Strategy == des.Strategy && backDatas[i].Address == des.Address {
				backDatas[i].Description = des.Description
				break
			}
		}
	}

	if req.Type == "port" {
		apps := u.loadPortByApp()
		for i := 0; i < len(backDatas); i++ {
			port, _ := strconv.Atoi(backDatas[i].Port)
			backDatas[i].IsUsed = common.ScanPort(port)
			if backDatas[i].Protocol == "udp" {
				backDatas[i].IsUsed = common.ScanUDPPort(port)
				continue
			}
			for _, app := range apps {
				if app.HttpPort == backDatas[i].Port || app.HttpsPort == backDatas[i].Port {
					backDatas[i].APPName = app.AppName
					break
				}
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
	switch operation {
	case "start":
		if err := client.Start(); err != nil {
			return err
		}
		if err := u.addPortsBeforeStart(client); err != nil {
			_ = client.Stop()
			return err
		}
		_, _ = cmd.Exec("systemctl restart docker")
		return nil
	case "stop":
		if err := client.Stop(); err != nil {
			return err
		}
		_, _ = cmd.Exec("systemctl restart docker")
		return nil
	case "disablePing":
		return u.updatePingStatus("0")
	case "enablePing":
		return u.updatePingStatus("1")
	}
	return fmt.Errorf("not support such operation: %s", operation)
}

func (u *FirewallService) OperatePortRule(req dto.PortRuleOperate, reload bool) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	protos := strings.Split(req.Protocol, "/")
	if client.Name() == "ufw" {
		if len(req.Address) == 0 {
			req.Address = "Anywhere"
		}
		if strings.Contains(req.Port, ",") || strings.Contains(req.Port, "-") {
			for _, proto := range protos {
				req.Port = strings.ReplaceAll(req.Port, "-", ":")
				req.Protocol = proto
				if err := u.operatePort(client, req); err != nil {
					return err
				}
				_ = u.addPortRecord(req)
			}
			return nil
		}
		if req.Protocol == "tcp/udp" {
			req.Protocol = ""
		}
		if err := u.operatePort(client, req); err != nil {
			return err
		}
		_ = u.addPortRecord(req)
		return nil
	}

	for _, proto := range protos {
		if strings.Contains(req.Port, "-") {
			req.Protocol = proto
			if err := u.operatePort(client, req); err != nil {
				return err
			}
			_ = u.addPortRecord(req)
		} else {
			ports := strings.Split(req.Port, ",")
			for _, port := range ports {
				if len(port) == 0 {
					continue
				}
				req.Port = port
				req.Protocol = proto
				if err := u.operatePort(client, req); err != nil {
					return err
				}
				_ = u.addPortRecord(req)
			}
		}
	}
	return client.Reload()
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
		_ = u.addAddressRecord(req)
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
	apps := u.loadPortByApp()
	for _, app := range apps {
		if err := client.Port(fireClient.FireInfo{Port: app.HttpPort, Protocol: "tcp", Strategy: "accept"}, "add"); err != nil {
			return err
		}
	}

	return client.Reload()
}

func (u *FirewallService) addPortRecord(req dto.PortRuleOperate) error {
	if req.Operation == "remove" {
		return hostRepo.DeleteFirewallRecord(req.Port, req.Protocol, req.Address, req.Strategy)
	}

	return hostRepo.SaveFirewallRecord(&model.Firewall{
		Port:        req.Port,
		Protocol:    req.Protocol,
		Address:     req.Address,
		Strategy:    req.Strategy,
		Description: req.Description,
	})
}

func (u *FirewallService) addAddressRecord(req dto.AddrRuleOperate) error {
	if req.Operation == "remove" {
		return hostRepo.DeleteFirewallRecord("", "", req.Address, req.Strategy)
	}
	return hostRepo.SaveFirewallRecord(&model.Firewall{
		Address:     req.Address,
		Strategy:    req.Strategy,
		Description: req.Description,
	})
}
