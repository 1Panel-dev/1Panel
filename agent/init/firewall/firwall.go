package firewall

import (
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client/iptables"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

func Init() {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return
	}
	re.RegisterRegex(iptables.Chian1PanelBasicPortPattern)
	re.RegisterRegex(iptables.Chain1PanelBasicAddressPattern)
	clientName := client.Name()
	if clientName == "ufw" || clientName == "iptables" {
		_ = iptables.LoadRulesFromFile(iptables.FilterTab, iptables.Chain1PanelForward, iptables.ForwardFileName)
		_ = iptables.LoadRulesFromFile(iptables.NatTab, iptables.Chain1PanelPreRouting, iptables.ForwardFileName1)
		_ = iptables.LoadRulesFromFile(iptables.NatTab, iptables.Chain1PanelPostRouting, iptables.ForwardFileName2)
	}
	if clientName == "iptables" {
		_ = iptables.LoadRulesFromFile(iptables.FilterTab, iptables.Chain1PanelBasic, iptables.BasicFileName)
		_ = iptables.LoadRulesFromFile(iptables.FilterTab, iptables.Chain1PanelInput, iptables.InputFileName)
		_ = iptables.LoadRulesFromFile(iptables.FilterTab, iptables.Chain1PanelOutput, iptables.OutputFileName)
	}
}
