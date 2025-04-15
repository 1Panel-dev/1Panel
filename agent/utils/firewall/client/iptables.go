package client

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

const NatChain = "1PANEL"

var NatListRegex = regexp.MustCompile(`^(\d+)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)(?:\s+(.+?) .+?:(\d{1,5}(?::\d+)?).+?[ :](.+-.+|(?:.+:)?\d{1,5}(?:-\d{1,5})?))?$`)

type Iptables struct {
	CmdStr string
}

func NewIptables() (*Iptables, error) {
	iptables := new(Iptables)
	iptables.CmdStr = cmd.SudoHandleCmd()

	return iptables, nil
}

func (iptables *Iptables) run(rule string) error {
	stdout, err := cmd.RunDefaultWithStdoutBashCf("%s iptables -t nat %s", iptables.CmdStr, rule)
	if err != nil {
		return fmt.Errorf("%s, %s", err, stdout)
	}
	if stdout != "" {
		return fmt.Errorf("iptables error: %s", stdout)
	}

	return nil
}

func (iptables *Iptables) runf(rule string, a ...any) error {
	return iptables.run(fmt.Sprintf(rule, a...))
}

func (iptables *Iptables) Check() error {
	stdout, err := cmd.RunDefaultWithStdoutBashC("cat /proc/sys/net/ipv4/ip_forward")
	if err != nil {
		return fmt.Errorf("%s, %s", err, stdout)
	}
	if stdout == "0" {
		return fmt.Errorf("disable")
	}

	return nil
}

func (iptables *Iptables) NatNewChain() error {
	return iptables.runf("-N %s", NatChain)
}

func (iptables *Iptables) NatAppendChain() error {
	return iptables.runf("-A PREROUTING -j %s", NatChain)
}

func (iptables *Iptables) NatList(chain ...string) ([]IptablesNatInfo, error) {
	rule := fmt.Sprintf("%s iptables -t nat -nL %s --line", iptables.CmdStr, NatChain)
	if len(chain) == 1 {
		rule = fmt.Sprintf("%s iptables -t nat -nL %s --line", iptables.CmdStr, chain[0])
	}
	stdout, err := cmd.RunDefaultWithStdoutBashC(rule)
	if err != nil {
		return nil, err
	}

	var forwardList []IptablesNatInfo
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimFunc(line, func(r rune) bool {
			return r <= 32
		})
		if NatListRegex.MatchString(line) {
			match := NatListRegex.FindStringSubmatch(line)
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

func (iptables *Iptables) NatAdd(protocol, src, destIp, destPort string, save bool) error {
	rule := fmt.Sprintf("-A %s -p %s --dport %s -j REDIRECT --to-port %s", NatChain, protocol, src, destPort)
	if destIp != "" && destIp != "127.0.0.1" && destIp != "localhost" {
		rule = fmt.Sprintf("-A %s -p %s --dport %s -j DNAT --to-destination %s:%s", NatChain, protocol, src, destIp, destPort)
	}
	if err := iptables.run(rule); err != nil {
		return err
	}

	if save {
		return global.DB.Save(&model.Forward{
			Protocol:   protocol,
			Port:       src,
			TargetIP:   destIp,
			TargetPort: destPort,
		}).Error
	}
	return nil
}

func (iptables *Iptables) NatRemove(num string, protocol, src, destIp, destPort string) error {
	if err := iptables.runf("-D %s %s", NatChain, num); err != nil {
		return err
	}

	global.DB.Where(
		"protocol = ? AND port = ? AND target_ip = ? AND target_port = ?",
		protocol,
		src,
		destIp,
		destPort,
	).Delete(&model.Forward{})
	return nil
}

func (iptables *Iptables) Reload() error {
	if err := iptables.runf("-F %s", NatChain); err != nil {
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
