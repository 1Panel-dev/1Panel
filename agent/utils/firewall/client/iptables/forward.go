package iptables

import (
	"strings"
)

func AddForward(protocol, srcPort, dest, destPort, iface string, save bool) error {
	srcPort = strings.ReplaceAll(srcPort, "-", ":")
	itemDstPort := strings.ReplaceAll(destPort, "-", ":")
	if dest != "" && dest != "127.0.0.1" && dest != "localhost" {
		args := []string{"-A", Chain1PanelPreRouting}
		if iface != "" {
			args = append(args, "-i", iface)
		}
		args = append(args, "-p", protocol, "--dport", srcPort, "-j", "DNAT", "--to-destination", dest+":"+destPort)
		if err := Run(NatTab, args...); err != nil {
			return err
		}

		if err := Run(NatTab, "-A", Chain1PanelPostRouting, "-d", dest, "-p", protocol, "--dport", itemDstPort, "-j", "MASQUERADE"); err != nil {
			return err
		}

		if err := Run(FilterTab, "-A", Chain1PanelForward, "-d", dest, "-p", protocol, "--dport", itemDstPort, "-j", "ACCEPT"); err != nil {
			return err
		}

		if err := Run(FilterTab, "-A", Chain1PanelForward, "-s", dest, "-p", protocol, "--sport", itemDstPort, "-j", "ACCEPT"); err != nil {
			return err
		}
	} else {
		args := []string{"-A", Chain1PanelPreRouting}
		if iface != "" {
			args = append(args, "-i", iface)
		}
		args = append(args, "-p", protocol, "--dport", srcPort, "-j", "REDIRECT", "--to-port", destPort)
		if err := Run(NatTab, args...); err != nil {
			return err
		}
	}
	return nil
}

func DeleteForward(num string, protocol, srcPort, dest, destPort, iface string) error {
	itemDstPort := strings.ReplaceAll(destPort, "-", ":")
	if err := Run(NatTab, "-D", Chain1PanelPreRouting, num); err != nil {
		return err
	}

	if dest != "" && dest != "127.0.0.1" && dest != "localhost" {
		if err := Run(NatTab, "-D", Chain1PanelPostRouting, "-d", dest, "-p", protocol, "--dport", itemDstPort, "-j", "MASQUERADE"); err != nil {
			return err
		}

		if err := Run(FilterTab, "-D", Chain1PanelForward, "-d", dest, "-p", protocol, "--dport", itemDstPort, "-j", "ACCEPT"); err != nil {
			return err
		}

		if err := Run(FilterTab, "-D", Chain1PanelForward, "-s", dest, "-p", protocol, "--sport", itemDstPort, "-j", "ACCEPT"); err != nil {
			return err
		}
	}
	return nil
}

func ListForward(chain ...string) ([]IptablesNatInfo, error) {
	if len(chain) == 0 {
		chain = append(chain, Chain1PanelPreRouting)
	}
	stdout, err := RunWithStd(NatTab, "-nvL", chain[0], "--line-numbers")
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
