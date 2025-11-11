package iptables

import (
	"fmt"
	"strings"
)

func AddForward(protocol, srcPort, dest, destPort, iface string, save bool) error {
	if dest != "" && dest != "127.0.0.1" && dest != "localhost" {
		iptablesArg := fmt.Sprintf("-A %s", Chain1PanelPreRouting)
		if iface != "" {
			iptablesArg += fmt.Sprintf(" -i %s", iface)
		}
		iptablesArg += fmt.Sprintf(" -p %s --dport %s -j DNAT --to-destination %s:%s", protocol, srcPort, dest, destPort)
		if err := Run(NatTab, iptablesArg); err != nil {
			return err
		}

		if err := Run(NatTab, fmt.Sprintf("-A %s -d %s -p %s --dport %s -j MASQUERADE", Chain1PanelPostRouting, dest, protocol, destPort)); err != nil {
			return err
		}

		if err := Run(FilterTab, fmt.Sprintf("-A %s -d %s -p %s --dport %s -j ACCEPT", Chain1PanelForward, dest, protocol, destPort)); err != nil {
			return err
		}

		if err := Run(FilterTab, fmt.Sprintf("-A %s -s %s -p %s --sport %s -j ACCEPT", Chain1PanelForward, dest, protocol, destPort)); err != nil {
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
	if err := Run(NatTab, fmt.Sprintf("-D %s %s", Chain1PanelPreRouting, num)); err != nil {
		return err
	}

	if dest != "" && dest != "127.0.0.1" && dest != "localhost" {
		if err := Run(NatTab, fmt.Sprintf("-D %s -d %s -p %s --dport %s -j MASQUERADE", Chain1PanelPostRouting, dest, protocol, destPort)); err != nil {
			return err
		}

		if err := Run(FilterTab, fmt.Sprintf("-D %s -d %s -p %s --dport %s -j ACCEPT", Chain1PanelForward, dest, protocol, destPort)); err != nil {
			return err
		}

		if err := Run(FilterTab, fmt.Sprintf("-D %s -s %s -p %s --sport %s -j ACCEPT", Chain1PanelForward, dest, protocol, destPort)); err != nil {
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
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimFunc(line, func(r rune) bool {
			return r <= 32
		})
		if natListRegex.MatchString(line) {
			match := natListRegex.FindStringSubmatch(line)
			if !strings.Contains(match[13], ":") {
				match[13] = fmt.Sprintf(":%s", match[13])
			}
			forwardList = append(forwardList, IptablesNatInfo{
				Num:         match[1],
				Target:      match[4],
				Protocol:    match[11],
				InIface:     match[7],
				OutIface:    match[8],
				Opt:         match[6],
				Source:      match[9],
				Destination: match[10],
				SrcPort:     match[12],
				DestPort:    match[13],
			})
		}
	}

	return forwardList, nil
}

type IptablesNatInfo struct {
	Num         string `json:"num"`
	Target      string `json:"target"`
	Protocol    string `json:"protocol"`
	InIface     string `json:"inIface"`
	OutIface    string `json:"outIface"`
	Opt         string `json:"opt"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	SrcPort     string `json:"srcPort"`
	DestPort    string `json:"destPort"`
}
