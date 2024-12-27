package client

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

const Chain = "1PANEL"

const (
	FilterTab = "filter"
	NatTab    = "nat"
)

var (
	natListRegex = regexp.MustCompile(`^(\d+)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)(?:\s+(.+?) .+?:(\d{1,5}(?::\d+)?).+?[ :](.+-.+|(?:.+:)?\d{1,5}(?:-\d{1,5})?))?$`)
)

type Iptables struct {
	CmdStr string
}

func NewIptables() (*Iptables, error) {
	iptables := new(Iptables)
	if cmd.HasNoPasswordSudo() {
		iptables.CmdStr = "sudo"
	}

	return iptables, nil
}

func (iptables *Iptables) runf(tab, rule string, a ...any) error {
	stdout, err := cmd.Execf("%s iptables -t %s %s", iptables.CmdStr, tab, fmt.Sprintf(rule, a...))
	if err != nil {
		return fmt.Errorf("%s, %s", err, stdout)
	}
	if stdout != "" {
		return fmt.Errorf("iptables error: %s", stdout)
	}

	return nil
}

func (iptables *Iptables) Check() error {
	stdout, err := cmd.Exec("cat /proc/sys/net/ipv4/ip_forward")
	if err != nil {
		return fmt.Errorf("%s, %s", err, stdout)
	}
	if stdout == "0" {
		return fmt.Errorf("disable")
	}

	return nil
}

func (iptables *Iptables) NewChain(tab string) error {
	return iptables.runf(tab, "-N %s", Chain)
}

func (iptables *Iptables) AppendChain(tab string, chain string) error {
	return iptables.runf(tab, "-A %s -j %s", chain, Chain)
}

func (iptables *Iptables) NatList(chain ...string) ([]IptablesNatInfo, error) {
	rule := fmt.Sprintf("%s iptables -t %s -nL %s --line", iptables.CmdStr, Chain)
	if len(chain) == 1 {
		rule = fmt.Sprintf("%s iptables -t %s -nL %s --line", iptables.CmdStr, chain[0])
	}
	stdout, err := cmd.Exec(rule)
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
			if !strings.Contains(match[9], ":") {
				match[9] = fmt.Sprintf(":%s", match[9])
			}
			forwardList = append(forwardList, IptablesNatInfo{
				Num:         match[1],
				Target:      match[2],
				Protocol:    match[7],
				Opt:         match[4],
				Source:      match[5],
				Destination: match[6],
				SrcPort:     match[8],
				DestPort:    match[9],
			})
		}
	}

	return forwardList, nil
}

func (iptables *Iptables) NatAdd(protocol, srcPort, dest, destPort string, save bool) error {
	if dest != "" && dest != "127.0.0.1" && dest != "localhost" {
		if err := iptables.runf(NatTab, fmt.Sprintf(
			"-A %s -p %s --dport %s -j DNAT --to-destination %s:%s",
			Chain,
			protocol,
			srcPort,
			dest,
			destPort,
		)); err != nil {
			return err
		}

		// 非本机转发, 按公网流程走
		if err := iptables.runf(NatTab, fmt.Sprintf(
			"-A %s -p %s -d %s --dport %s -j MASQUERADE",
			Chain,
			protocol,
			dest,
			destPort,
		)); err != nil {
			return err
		}

		if err := iptables.runf(FilterTab, fmt.Sprintf(
			"-A %s -d %s -p %s --dport %s -j ACCEPT",
			Chain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}

		if err := iptables.runf(FilterTab, fmt.Sprintf(
			"-A %s -s %s -p %s --sport %s -j ACCEPT",
			Chain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}
	} else {
		if err := iptables.runf(NatTab, fmt.Sprintf(
			"-A %s -p %s --dport %s -j REDIRECT --to-port %s",
			Chain,
			protocol,
			srcPort,
			destPort,
		)); err != nil {
			return err
		}
	}

	if save {
		return global.DB.Save(&model.Forward{
			Protocol:   protocol,
			Port:       srcPort,
			TargetIP:   dest,
			TargetPort: destPort,
		}).Error
	}
	return nil
}

func (iptables *Iptables) NatRemove(num string, protocol, srcPort, dest, destPort string) error {
	if err := iptables.runf(NatTab, "-D %s %s", Chain, num); err != nil {
		return err
	}

	// 删除公网转发规则
	if dest != "" && dest != "127.0.0.1" && dest != "localhost" {
		if err := iptables.runf(NatTab, fmt.Sprintf(
			"-D %s -p %s --dport %s -j DNAT MASQUERADE",
			Chain,
			protocol,
			destPort,
		)); err != nil {
			return err
		}

		if err := iptables.runf(FilterTab, fmt.Sprintf(
			"-D %s -d %s -p %s --dport %s -j ACCEPT",
			Chain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}

		if err := iptables.runf(FilterTab, fmt.Sprintf(
			"-D %s -s %s -p %s --sport %s -j ACCEPT",
			Chain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}
	}

	global.DB.Where(
		"protocol = ? AND port = ? AND target_ip = ? AND target_port = ?",
		protocol,
		srcPort,
		dest,
		destPort,
	).Delete(&model.Forward{})
	return nil
}

func (iptables *Iptables) Reload() error {
	if err := iptables.runf(NatTab, "-F %s", Chain); err != nil {
		return err
	}
	if err := iptables.runf(FilterTab, "-F %s", Chain); err != nil {
		return err
	}

	var rules []model.Forward
	global.DB.Find(&rules)
	for _, forward := range rules {
		if err := iptables.NatAdd(forward.Protocol, forward.Port, forward.TargetIP, forward.TargetPort, false); err != nil {
			return err
		}
	}
	return nil
}
