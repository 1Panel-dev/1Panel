package hostnetwork

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/netip"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

func Routes(ctx context.Context, protocol string) ([]Route, error) {
	result := make([]Route, 0)
	for _, family := range []string{"IPv4", "IPv6"} {
		flag := "-4"
		if family == "IPv6" {
			flag = "-6"
		}
		var routes []struct {
			Destination     string          `json:"dst"`
			PreferredSource string          `json:"prefsrc"`
			Gateway         string          `json:"gateway"`
			Interface       string          `json:"dev"`
			Metric          uint32          `json:"metric"`
			Protocol        string          `json:"protocol"`
			Scope           string          `json:"scope"`
			NextHopID       uint32          `json:"nhid"`
			Source          string          `json:"from"`
			TOS             string          `json:"tos"`
			Type            string          `json:"type"`
			Metrics         json.RawMessage `json:"metrics"`
			Flags           []string        `json:"flags"`
			Encap           json.RawMessage `json:"encap"`
			Via             json.RawMessage `json:"via"`
			NewDestination  json.RawMessage `json:"newdst"`
			Realm           json.RawMessage `json:"realm"`
			InputInterface  string          `json:"iif"`
			NextHops        []struct {
				Gateway   string `json:"gateway"`
				Interface string `json:"dev"`
				Weight    uint32 `json:"weight"`
			} `json:"nexthops"`
		}
		args := []string{flag, "route", "show", "table", "main"}
		if protocol != "" {
			args = append(args, "proto", protocol)
		}
		if err := readIPJSON(ctx, 10*time.Second, args, &routes); err != nil {
			if canReadLegacy(ctx, err) {
				return readLegacyRoutes(protocol)
			}
			if strings.Contains(err.Error(), "FIB table does not exist") {
				continue
			}
			return nil, err
		}
		for _, route := range routes {
			if route.Type != "" && route.Type != "unicast" {
				continue
			}
			item := Route{PreferredSource: route.PreferredSource, Destination: route.Destination, Gateway: route.Gateway, Interface: route.Interface, Metric: route.Metric, Table: "main", Protocol: route.Protocol, Scope: route.Scope, Family: family, ReadOnly: len(route.NextHops) > 0 || route.Interface == "" || route.NextHopID != 0 || route.Source != "" || route.TOS != "" || len(route.Metrics) > 0 || len(route.Flags) > 0 || len(route.Encap) > 0 || len(route.Via) > 0 || len(route.NewDestination) > 0 || len(route.Realm) > 0 || route.InputInterface != "", NextHops: make([]NextHop, 0, len(route.NextHops))}
			for _, hop := range route.NextHops {
				item.NextHops = append(item.NextHops, NextHop{Gateway: hop.Gateway, Interface: hop.Interface, Weight: hop.Weight})
			}
			result = append(result, item)
		}
	}
	return result, nil
}

func routeArgs(req Route) ([]string, error) {
	if _, err := net.InterfaceByName(req.Interface); err != nil {
		return nil, err
	}
	args := []string{req.Destination}
	if req.Gateway != "" {
		args = append(args, "via", req.Gateway)
	}
	args = append(args, "dev", req.Interface, "table", "main", "metric", strconv.FormatUint(uint64(req.Metric), 10))
	if req.Protocol != "" {
		if !re.GetRegex(re.HostNetworkRouteTokenPattern).MatchString(req.Protocol) {
			return nil, errors.New("invalid route protocol")
		}
		args = append(args, "proto", req.Protocol)
	}
	if req.Scope != "" {
		if !re.GetRegex(re.HostNetworkRouteTokenPattern).MatchString(req.Scope) {
			return nil, errors.New("invalid route scope")
		}
		args = append(args, "scope", req.Scope)
	}
	return args, nil
}

func ChangeRoute(route Route, operation string) error {
	args, err := routeArgs(route)
	if err != nil {
		return err
	}
	family := "-4"
	if route.Family == "IPv6" {
		family = "-6"
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(10*time.Second)).Run("ip", append([]string{family, "route", operation}, args...)...)
}

func CheckRoute(ctx context.Context, req Route) (*RouteCheck, error) {
	if _, err := routeArgs(req); err != nil {
		return nil, err
	}
	var err error
	target := req.Gateway
	checkHost := false
	if req.Destination != "default" {
		prefix, parseErr := netip.ParsePrefix(req.Destination)
		if parseErr != nil {
			address, _ := netip.ParseAddr(req.Destination)
			prefix = netip.PrefixFrom(address, address.BitLen())
		}
		target = prefix.Masked().Addr().String()
		checkHost = prefix.Bits() == prefix.Addr().BitLen()
		if !checkHost {
			target = prefix.Masked().Addr().Next().String()
		}
	}
	if req.Gateway == "" && !checkHost {
		iface, err := net.InterfaceByName(req.Interface)
		if err != nil {
			return nil, err
		}
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagRunning == 0 {
			return &RouteCheck{Message: "Interface " + req.Interface + " is down or has no running link"}, nil
		}
		return &RouteCheck{Reachable: true, Message: "Direct route exists in the main routing table and interface " + req.Interface + " is running; no host ICMP probe was performed"}, nil
	}
	family := "-4"
	if req.Family == "IPv6" {
		family = "-6"
	}
	var selected []struct {
		Interface string `json:"dev"`
		Gateway   string `json:"gateway"`
	}
	var selectedInterface, selectedGateway string
	if err = readIPJSON(ctx, 5*time.Second, []string{family, "route", "get", target}, &selected); err != nil {
		if !canReadLegacy(ctx, err) {
			return nil, err
		}
		output, err := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(5*time.Second)).RunWithStdout("ip", family, "-o", "route", "get", target)
		if err != nil {
			return nil, err
		}
		line, _, _ := strings.Cut(output, "\n")
		fields := strings.Fields(line)
		for i := 1; i+1 < len(fields); i++ {
			switch fields[i] {
			case "dev":
				selectedInterface = fields[i+1]
				i++
			case "via":
				selectedGateway = fields[i+1]
				i++
			}
		}
	} else if len(selected) != 0 {
		selectedInterface, selectedGateway = selected[0].Interface, selected[0].Gateway
	}
	if selectedInterface != req.Interface || (req.Destination != "default" && selectedGateway != req.Gateway) {
		return &RouteCheck{Message: "Kernel lookup selected a different route for " + target}, nil
	}
	if req.Gateway != "" {
		target = req.Gateway
	}
	if _, err = exec.LookPath("ping"); err != nil {
		return nil, err
	}
	output, err := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(5*time.Second)).RunWithStdout("ping", family, "-n", "-I", req.Interface, "-c", "1", "-W", "2", target)
	result := &RouteCheck{Reachable: err == nil, Message: strings.TrimSpace(output)}
	if err != nil {
		result.Message = err.Error()
	}
	return result, nil
}
