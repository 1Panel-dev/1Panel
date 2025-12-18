package iptables

import (
	"fmt"
	"strings"
)

func AddForward(protocol, srcPort, dest, destPort, iface string, save bool) error {
	srcPort = strings.ReplaceAll(srcPort, "-", ":")
	itemDstPort := strings.ReplaceAll(destPort, "-", ":")
	if dest != "" && dest != "127.0.0.1" && dest != "localhost" {
		iptablesArg := fmt.Sprintf("-A %s", Chain1PanelPreRouting)
		if iface != "" {
			iptablesArg += fmt.Sprintf(" -i %s", iface)
		}
		iptablesArg += fmt.Sprintf(" -p %s --dport %s -j DNAT --to-destination %s:%s", protocol, srcPort, dest, destPort)
		if err := Run(NatTab, iptablesArg); err != nil {
			return err
		}

		if err := Run(NatTab, fmt.Sprintf("-A %s -d %s -p %s --dport %s -j MASQUERADE", Chain1PanelPostRouting, dest, protocol, itemDstPort)); err != nil {
			return err
		}

		if err := Run(FilterTab, fmt.Sprintf("-A %s -d %s -p %s --dport %s -j ACCEPT", Chain1PanelForward, dest, protocol, itemDstPort)); err != nil {
			return err
		}

		if err := Run(FilterTab, fmt.Sprintf("-A %s -s %s -p %s --sport %s -j ACCEPT", Chain1PanelForward, dest, protocol, itemDstPort)); err != nil {
			return err
		}
	} else {
		iptablesArg := fmt.Sprintf("-A %s", Chain1PanelPreRouting)
		if iface != "" {
			iptablesArg += fmt.Sprintf(" -i %s", iface)
		}
		iptablesArg += fmt.Sprintf(" -p %s --dport %s -j REDIRECT --to-port %s", protocol, srcPort, destPort)
		if err := Run(NatTab, iptablesArg); err != nil {
			return err
		}
	}
	return nil
}

func DeleteForward(num string, protocol, srcPort, dest, destPort, iface string) error {
	itemDstPort := strings.ReplaceAll(destPort, "-", ":")
	if err := Run(NatTab, fmt.Sprintf("-D %s %s", Chain1PanelPreRouting, num)); err != nil {
		return err
	}

	if dest != "" && dest != "127.0.0.1" && dest != "localhost" {
		if err := Run(NatTab, fmt.Sprintf("-D %s -d %s -p %s --dport %s -j MASQUERADE", Chain1PanelPostRouting, dest, protocol, itemDstPort)); err != nil {
			return err
		}

		if err := Run(FilterTab, fmt.Sprintf("-D %s -d %s -p %s --dport %s -j ACCEPT", Chain1PanelForward, dest, protocol, itemDstPort)); err != nil {
			return err
		}

		if err := Run(FilterTab, fmt.Sprintf("-D %s -s %s -p %s --sport %s -j ACCEPT", Chain1PanelForward, dest, protocol, itemDstPort)); err != nil {
			return err
		}
	}
	return nil
}

func ListForward(chain ...string) ([]IptablesNatInfo, error) {
	if len(chain) == 0 {
		chain = append(chain, Chain1PanelPreRouting)
	}
	stdout, err := RunWithStd(NatTab, fmt.Sprintf("-nvL %s --line-numbers", chain[0]))
	if err != nil {
		return nil, err
	}
	var forwardList []IptablesNatInfo
	lines := strings.Split(stdout, "\n")
	for i := 0; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) < 13 {
			continue
		}
		item := IptablesNatInfo{
			Num:      fields[0],
			Protocol: loadProtocol(fields[4]),
			InIface:  fields[6],
			OutIface: fields[7],
			Source:   fields[8],
			SrcPort:  loadNatSrcPort(fields[11]),
		}
		if len(fields) == 15 && fields[13] == "ports" {
			item.DestPort = fields[14]
		}
		if len(fields) == 13 && strings.HasPrefix(fields[12], "to:") {
			parts := strings.Split(fields[12], ":")
			if len(parts) > 2 {
				item.DestPort = parts[2]
				item.Destination = parts[1]
			}
		}
		if len(item.Destination) == 0 {
			item.Destination = "127.0.0.1"
		}
		forwardList = append(forwardList, item)
	}

	return forwardList, nil
}

func loadNatSrcPort(portStr string) string {
	var portItem string
	if strings.Contains(portStr, "dpt:") {
		portItem = strings.ReplaceAll(portStr, "dpt:", "")
	}
	if strings.Contains(portStr, "dpts:") {
		portItem = strings.ReplaceAll(portStr, "dpts:", "")
	}
	portItem = strings.ReplaceAll(portItem, ":", "-")
	return portItem
}

type IptablesNatInfo struct {
	Num         string `json:"num"`
	Protocol    string `json:"protocol"`
	InIface     string `json:"inIface"`
	OutIface    string `json:"outIface"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	SrcPort     string `json:"srcPort"`
	DestPort    string `json:"destPort"`
}
