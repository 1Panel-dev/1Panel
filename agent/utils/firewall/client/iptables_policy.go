package client

import (
	"fmt"
	"strings"
)

// IptablesChain
type IptablesChain struct {
	Name          string
	DefaultPolicy string
	FirstRule     *IptablesPolicyChainItem
	LastRule      *IptablesPolicyChainItem
}

func (c *IptablesChain) ParseLine(line string) error {
	cmd := strings.Split(line, " ")

	if cmd[0] == "-P" {
		c.Name = cmd[1]
		c.DefaultPolicy = cmd[2]
		return nil
	}
	if cmd[0] == "-A" {
		if cmd[1] != c.Name {
			return fmt.Errorf("invalid chain name in rule line: %s", line)
		}
		policy := IptablesPolicy{}
		for i := 2; i < len(cmd); i++ {
			switch cmd[i] {
			case "-p":
				i++
				policy.Protocol = cmd[i]
			case "--dport":
				i++
				// parse port
				var port uint16
				fmt.Sscanf(cmd[i], "%d", &port)
				policy.DstPort = port
			case "--sport":
				i++
				var port uint16
				fmt.Sscanf(cmd[i], "%d", &port)
				policy.SrcPort = port
			case "-s":
				i++
				policy.SourceIP = cmd[i]
			case "-d":
				i++
				policy.DestIP = cmd[i]
			case "-j":
				i++
				policy.Action = cmd[i]
			case "-m":
				// skip
				i++
			case "--comment":
				i++
				policy.Comment = strings.Trim(cmd[i], "\"")
			}
		}
		newItem := &IptablesPolicyChainItem{
			P: policy,
		}
		if c.FirstRule == nil {
			c.FirstRule = newItem
			c.LastRule = newItem
		} else {
			current := c.LastRule
			current.SetNext(newItem)
			c.LastRule = newItem
		}
		return nil
	}
	return fmt.Errorf("invalid iptables rule line: %s", line)
}

type IptablesPolicyChainItem struct {
	next *IptablesPolicyChainItem
	P    IptablesPolicy
}

func (item *IptablesPolicyChainItem) SetNext(next *IptablesPolicyChainItem) {
	item.next = next
}

func (item *IptablesPolicyChainItem) Next() *IptablesPolicyChainItem {
	return item.next
}

type IptablesPolicy struct {
	Protocol string
	SrcPort  uint16
	DstPort  uint16
	SourceIP string
	DestIP   string
	Action   string // ACCEPT, DROP, REJECT
	Comment  string
}
