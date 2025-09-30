package client

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

const (
	PreRoutingChain  = "1PANEL_PREROUTING"
	PostRoutingChain = "1PANEL_POSTROUTING"
	ForwardChain     = "1PANEL_FORWARD"
)

const (
	FilterTab = "filter"
	NatTab    = "nat"
)

const NatChain = "1PANEL"

var (
	natListRegex = regexp.MustCompile(`^(\d+)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)(?:\s+(.+?) .+?:(\d{1,5}(?::\d+)?).+?[ :](.+-.+|(?:.+:)?\d{1,5}(?:-\d{1,5})?))?$`)
)

type Iptables struct {
	CmdStr string
}

func NewIptables() (*Iptables, error) {
	iptables := new(Iptables)
	iptables.CmdStr = cmd.SudoHandleCmd()

	return iptables, nil
}

func (iptables *Iptables) outf(tab, rule string, a ...any) (stdout string, err error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithIgnoreExist1(), cmd.WithTimeout(20*time.Second))
	stdout, err = cmdMgr.RunWithStdoutBashCf("%s iptables -t %s %s", iptables.CmdStr, tab, fmt.Sprintf(rule, a...))
	if err != nil && stdout != "" {
		global.LOG.Errorf("iptables failed, err: %s", stdout)
	}
	return
}

func (iptables *Iptables) runf(tab, rule string, a ...any) error {
	stdout, err := iptables.outf(tab, rule, a...)
	if err != nil {
		return fmt.Errorf("%s, %s", err, stdout)
	}
	if stdout != "" {
		return fmt.Errorf("iptables error: %s", stdout)
	}

	return nil
}

func (iptables *Iptables) Check() error {
	stdout, err := cmd.RunDefaultWithStdoutBashC("cat /proc/sys/net/ipv4/ip_forward")
	if err != nil {
		return fmt.Errorf("check ip_forward error: %w, output: %s", err, stdout)
	}
	if strings.TrimSpace(stdout) == "0" {
		return fmt.Errorf("ipv4 forward disabled")
	}

	chain, err := iptables.outf(NatTab, "-L -n | grep 'Chain %s'", PreRoutingChain)
	if err != nil {
		return fmt.Errorf("failed to check chain: %w", err)
	}
	if strings.TrimSpace(chain) != "" {
		return fmt.Errorf("chain %s already exists", PreRoutingChain)
	}

	return nil
}

func (iptables *Iptables) NewChain(tab, chain string) error {
	return iptables.runf(tab, "-N %s", chain)
}

func (iptables *Iptables) AppendChain(tab string, chain, chain1 string) error {
	return iptables.runf(tab, "-A %s -j %s", chain, chain1)
}

func (iptables *Iptables) NatList(chain ...string) ([]IptablesNatInfo, error) {
	if len(chain) == 0 {
		chain = append(chain, PreRoutingChain)
	}
	stdout, err := iptables.outf(NatTab, "-nvL %s --line-numbers", chain[0])
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

func (iptables *Iptables) NatAdd(protocol, srcPort, dest, destPort, iface string, save bool) error {
	if dest != "" && dest != "127.0.0.1" && dest != "localhost" {
		iptablesArg := fmt.Sprintf("-A %s", PreRoutingChain)
		if iface != "" {
			iptablesArg += fmt.Sprintf(" -i %s", iface)
		}
		iptablesArg += fmt.Sprintf(" -p %s --dport %s -j DNAT --to-destination %s:%s", protocol, srcPort, dest, destPort)
		if err := iptables.runf(NatTab, iptablesArg); err != nil {
			return err
		}

		if err := iptables.runf(NatTab, fmt.Sprintf(
			"-A %s -d %s -p %s --dport %s -j MASQUERADE",
			PostRoutingChain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}

		if err := iptables.runf(FilterTab, fmt.Sprintf(
			"-A %s -d %s -p %s --dport %s -j ACCEPT",
			ForwardChain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}

		if err := iptables.runf(FilterTab, fmt.Sprintf(
			"-A %s -s %s -p %s --sport %s -j ACCEPT",
			ForwardChain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}
	} else {
		iptablesArg := fmt.Sprintf("-A %s", PreRoutingChain)
		if iface != "" {
			iptablesArg += fmt.Sprintf(" -i %s", iface)
		}
		iptablesArg += fmt.Sprintf(" -p %s --dport %s -j REDIRECT --to-port %s", protocol, srcPort, destPort)
		if err := iptables.runf(NatTab, iptablesArg); err != nil {
			return err
		}
	}

	if save {
		return global.DB.Save(&model.Forward{
			Protocol:   protocol,
			Port:       srcPort,
			TargetIP:   dest,
			TargetPort: destPort,
			Interface:  iface,
		}).Error
	}
	return nil
}

func (iptables *Iptables) NatRemove(num string, protocol, srcPort, dest, destPort, iface string) error {
	if err := iptables.runf(NatTab, "-D %s %s", PreRoutingChain, num); err != nil {
		return err
	}

	if dest != "" && dest != "127.0.0.1" && dest != "localhost" {
		if err := iptables.runf(NatTab, fmt.Sprintf(
			"-D %s -d %s -p %s --dport %s -j MASQUERADE",
			PostRoutingChain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}

		if err := iptables.runf(FilterTab, fmt.Sprintf(
			"-D %s -d %s -p %s --dport %s -j ACCEPT",
			ForwardChain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}

		if err := iptables.runf(FilterTab, fmt.Sprintf(
			"-D %s -s %s -p %s --sport %s -j ACCEPT",
			ForwardChain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}
	}

	global.DB.Where(
		"protocol = ? AND port = ? AND target_ip = ? AND target_port = ? AND (interface = ? OR (interface IS NULL AND ? = ''))",
		protocol,
		srcPort,
		dest,
		destPort,
		iface,
		iface,
	).Delete(&model.Forward{})
	return nil
}

func (iptables *Iptables) Reload() error {
	if err := iptables.runf(NatTab, "-F %s", PreRoutingChain); err != nil {
		return err
	}
	if err := iptables.runf(NatTab, "-F %s", PostRoutingChain); err != nil {
		return err
	}
	if err := iptables.runf(FilterTab, "-F %s", ForwardChain); err != nil {
		return err
	}

	var rules []model.Forward
	global.DB.Find(&rules)
	for _, forward := range rules {
		if err := iptables.NatAdd(forward.Protocol, forward.Port, forward.TargetIP, forward.TargetPort, forward.Interface, false); err != nil {
			return err
		}
	}
	return nil
}
