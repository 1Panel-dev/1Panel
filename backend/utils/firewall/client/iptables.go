package client

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

var NatListRegex = regexp.MustCompile(`^(\d+)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?) .+?:(\d{1,5}(?::\d+)?).+?[ :](.+-.+|(?:.+:)?\d{1,5}(?:-\d{1,5})?)$`)

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

func (iptables *Iptables) Check() error {
	stdout, err := cmd.Exec("cat /proc/sys/net/ipv4/ip_forward")
	if err != nil {
		return err
	}
	if stdout == "0" {
		return fmt.Errorf("disable")
	}

	return nil
}

func (iptables *Iptables) NatList() ([]IptablesNatInfo, error) {
	stdout, err := cmd.Execf("%s iptables -t nat -nL PREROUTING --line", iptables.CmdStr)
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
	rule := fmt.Sprintf("%s iptables -t nat -A PREROUTING -p %s --dport %s -j REDIRECT --to-port %s", iptables.CmdStr, protocol, src, destPort)
	if destIp != "" && destIp != "127.0.0.1" && destIp != "localhost" {
		rule = fmt.Sprintf("%s iptables -t nat -A PREROUTING -p %s --dport %s -j DNAT --to-destination %s:%s", iptables.CmdStr, protocol, src, destIp, destPort)
	}
	stdout, err := cmd.Exec(rule)
	if err != nil {
		return err
	}
	if stdout != "" {
		return errors.New(stdout)
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
	stdout, err := cmd.Execf("%s iptables -t nat -D PREROUTING %s", iptables.CmdStr, num)
	if err != nil {
		return err
	}
	if stdout != "" {
		return fmt.Errorf(stdout)
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
	stdout, err := cmd.Execf("%s iptables -t nat -F && %s iptables -t nat -X", iptables.CmdStr, iptables.CmdStr)
	if err != nil {
		return err
	}
	if stdout != "" {
		return fmt.Errorf(stdout)
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
