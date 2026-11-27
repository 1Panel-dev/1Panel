package hostnetwork

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func readLegacyLinks(name string) ([]Link, error) {
	data, err := syscall.NetlinkRIB(unix.RTM_GETLINK, unix.AF_UNSPEC)
	if err != nil {
		return nil, err
	}
	messages, err := syscall.ParseNetlinkMessage(data)
	if err != nil {
		return nil, err
	}
	links := make([]Link, 0)
	indices := make(map[int]int)
	for _, msg := range messages {
		if msg.Header.Type != unix.RTM_NEWLINK || len(msg.Data) < unix.SizeofIfInfomsg {
			continue
		}
		attrs, err := syscall.ParseNetlinkRouteAttr(&msg)
		if err != nil {
			return nil, err
		}
		link := Link{
			Index:     int(int32(binary.NativeEndian.Uint32(msg.Data[4:8]))),
			Flags:     legacyLinkFlags(binary.NativeEndian.Uint32(msg.Data[8:12])),
			OperState: "UNKNOWN",
			Addresses: []Address{},
		}
		for _, attr := range attrs {
			switch attr.Attr.Type {
			case unix.IFLA_IFNAME:
				link.Name = strings.TrimRight(string(attr.Value), "\x00")
			case unix.IFLA_MTU:
				if len(attr.Value) >= 4 {
					link.MTU = int(binary.NativeEndian.Uint32(attr.Value[:4]))
				}
			case unix.IFLA_ADDRESS:
				link.MAC = net.HardwareAddr(attr.Value).String()
			case unix.IFLA_OPERSTATE:
				states := [...]string{"UNKNOWN", "NOTPRESENT", "DOWN", "LOWERLAYERDOWN", "TESTING", "DORMANT", "UP"}
				if len(attr.Value) > 0 && int(attr.Value[0]) < len(states) {
					link.OperState = states[attr.Value[0]]
				}
			}
		}
		if link.Index <= 0 || link.Name == "" || (name != "" && name != link.Name) {
			continue
		}
		indices[link.Index] = len(links)
		links = append(links, link)
	}
	if name != "" && len(links) == 0 {
		return nil, fmt.Errorf("network interface %q not found", name)
	}
	data, err = syscall.NetlinkRIB(unix.RTM_GETADDR, unix.AF_UNSPEC)
	if err != nil {
		return nil, err
	}
	messages, err = syscall.ParseNetlinkMessage(data)
	if err != nil {
		return nil, err
	}
	for _, msg := range messages {
		if msg.Header.Type != unix.RTM_NEWADDR || len(msg.Data) < unix.SizeofIfAddrmsg {
			continue
		}
		index, ok := indices[int(binary.NativeEndian.Uint32(msg.Data[4:8]))]
		if !ok {
			continue
		}
		address, err := legacyAddress(&msg)
		if err != nil {
			return nil, err
		}
		if address.Local != "" {
			links[index].Addresses = append(links[index].Addresses, address)
		}
	}
	return links, nil
}

func legacyAddress(msg *syscall.NetlinkMessage) (Address, error) {
	family, bits := "inet", 32
	switch msg.Data[0] {
	case unix.AF_INET:
	case unix.AF_INET6:
		family, bits = "inet6", 128
	default:
		return Address{}, nil
	}
	attrs, err := syscall.ParseNetlinkRouteAttr(msg)
	if err != nil {
		return Address{}, err
	}
	address := Address{Family: family, PrefixLen: int(msg.Data[1]), Scope: legacyScope(msg.Data[3])}
	flags := uint32(msg.Data[2])
	var local, peer netip.Addr
	for _, attr := range attrs {
		switch attr.Attr.Type {
		case unix.IFA_LOCAL:
			local, _ = netip.AddrFromSlice(attr.Value)
		case unix.IFA_ADDRESS:
			peer, _ = netip.AddrFromSlice(attr.Value)
		case unix.IFA_LABEL:
			address.Label = strings.TrimRight(string(attr.Value), "\x00")
		case unix.IFA_FLAGS:
			if len(attr.Value) >= 4 {
				flags = binary.NativeEndian.Uint32(attr.Value[:4])
			}
		case unix.IFA_CACHEINFO:
			if len(attr.Value) >= 8 {
				validTime := binary.NativeEndian.Uint32(attr.Value[4:8])
				address.ValidTime = &validTime
			}
		}
	}
	if !local.IsValid() {
		local = peer
	}
	if local.BitLen() != bits || address.PrefixLen > bits {
		return Address{}, nil
	}
	address.Local = local.String()
	address.Secondary = flags&unix.IFA_F_SECONDARY != 0
	address.Dynamic = flags&unix.IFA_F_PERMANENT == 0
	address.Tentative = flags&unix.IFA_F_TENTATIVE != 0
	address.DadFailed = flags&unix.IFA_F_DADFAILED != 0
	return address, nil
}

func legacyScope(scope byte) string {
	switch scope {
	case unix.RT_SCOPE_UNIVERSE:
		return "global"
	case unix.RT_SCOPE_SITE:
		return "site"
	case unix.RT_SCOPE_LINK:
		return "link"
	case unix.RT_SCOPE_HOST:
		return "host"
	case unix.RT_SCOPE_NOWHERE:
		return "nowhere"
	default:
		return strconv.Itoa(int(scope))
	}
}

func legacyLinkFlags(flags uint32) []string {
	result := make([]string, 0)
	if flags&unix.IFF_UP != 0 && flags&unix.IFF_RUNNING == 0 {
		result = append(result, "NO-CARRIER")
	}
	for _, flag := range [...]struct {
		value uint32
		name  string
	}{
		{unix.IFF_LOOPBACK, "LOOPBACK"}, {unix.IFF_BROADCAST, "BROADCAST"}, {unix.IFF_POINTOPOINT, "POINTOPOINT"},
		{unix.IFF_MULTICAST, "MULTICAST"}, {unix.IFF_NOARP, "NOARP"}, {unix.IFF_ALLMULTI, "ALLMULTI"},
		{unix.IFF_PROMISC, "PROMISC"}, {unix.IFF_MASTER, "MASTER"}, {unix.IFF_SLAVE, "SLAVE"},
		{unix.IFF_DEBUG, "DEBUG"}, {unix.IFF_DYNAMIC, "DYNAMIC"}, {unix.IFF_AUTOMEDIA, "AUTOMEDIA"},
		{unix.IFF_PORTSEL, "PORTSEL"}, {unix.IFF_NOTRAILERS, "NOTRAILERS"}, {unix.IFF_UP, "UP"},
		{unix.IFF_LOWER_UP, "LOWER_UP"}, {unix.IFF_DORMANT, "DORMANT"}, {unix.IFF_ECHO, "ECHO"},
	} {
		if flags&flag.value != 0 {
			result = append(result, flag.name)
		}
	}
	return result
}

const legacyRouteNextHopID = 30

func readLegacyRoutes(protocol string) ([]Route, error) {
	var protocolNumber uint64
	if protocol != "" {
		var err error
		protocolNumber, err = strconv.ParseUint(protocol, 10, 8)
		if err != nil {
			return nil, err
		}
	}
	links, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	names := make(map[uint32]string, len(links))
	for _, link := range links {
		names[uint32(link.Index)] = link.Name
	}
	result := make([]Route, 0)
	for _, family := range []int{unix.AF_INET, unix.AF_INET6} {
		dump, err := syscall.NetlinkRIB(unix.RTM_GETROUTE, family)
		if err != nil {
			return nil, err
		}
		messages, err := syscall.ParseNetlinkMessage(dump)
		if err != nil {
			return nil, err
		}
		for _, message := range messages {
			if message.Header.Type != unix.RTM_NEWROUTE || len(message.Data) < unix.SizeofRtMsg || int(message.Data[0]) != family || message.Data[7] != unix.RTN_UNICAST || binary.NativeEndian.Uint32(message.Data[8:12])&unix.RTM_F_CLONED != 0 {
				continue
			}
			if protocol != "" && message.Data[5] != byte(protocolNumber) {
				continue
			}
			route, mainTable, err := parseLegacyRoute(&message, names)
			if err != nil {
				return nil, err
			}
			if mainTable {
				result = append(result, route)
			}
		}
	}
	return result, nil
}

func parseLegacyRoute(message *syscall.NetlinkMessage, names map[uint32]string) (Route, bool, error) {
	data := message.Data
	attrs, err := syscall.ParseNetlinkRouteAttr(message)
	if err != nil {
		return Route{}, false, err
	}
	table := uint32(data[4])
	for _, attr := range attrs {
		if attr.Attr.Type == unix.RTA_TABLE && len(attr.Value) == 4 {
			table = binary.NativeEndian.Uint32(attr.Value)
		}
	}
	if table != unix.RT_TABLE_MAIN {
		return Route{}, false, nil
	}
	family := "IPv4"
	if data[0] == unix.AF_INET6 {
		family = "IPv6"
	}
	route := Route{
		Destination: "default",
		Table:       "main",
		Family:      family,
		Protocol:    legacyRouteProtocol(data[5]),
		Scope:       legacyScope(data[6]),
		ReadOnly:    data[2] != 0 || data[3] != 0,
		NextHops:    make([]NextHop, 0),
	}
	var destination netip.Addr
	for _, attr := range attrs {
		switch attr.Attr.Type {
		case unix.RTA_DST:
			destination, _ = netip.AddrFromSlice(attr.Value)
		case unix.RTA_PREFSRC:
			route.PreferredSource = legacyRouteIP(attr.Value)
		case unix.RTA_GATEWAY:
			route.Gateway = legacyRouteIP(attr.Value)
			route.ReadOnly = route.ReadOnly || route.Gateway == ""
		case unix.RTA_OIF:
			if len(attr.Value) == 4 {
				route.Interface = names[binary.NativeEndian.Uint32(attr.Value)]
			}
		case unix.RTA_PRIORITY:
			if len(attr.Value) == 4 {
				route.Metric = binary.NativeEndian.Uint32(attr.Value)
			} else {
				route.ReadOnly = true
			}
		case unix.RTA_MULTIPATH:
			route.ReadOnly = true
			route.NextHops, err = parseLegacyNextHops(attr.Value, names)
			if err != nil {
				return Route{}, false, err
			}
		case unix.RTA_VIA:
			if len(attr.Value) >= 2 {
				route.Gateway = legacyRouteIP(attr.Value[2:])
			}
			route.ReadOnly = true
		case unix.RTA_SRC, legacyRouteNextHopID, unix.RTA_NEWDST, unix.RTA_ENCAP_TYPE, unix.RTA_ENCAP, unix.RTA_IIF, unix.RTA_FLOW, unix.RTA_METRICS:
			route.ReadOnly = true
		}
	}
	if bits := int(data[1]); bits != 0 {
		prefix := netip.PrefixFrom(destination, bits)
		if !prefix.IsValid() || destination.Is4() != (family == "IPv4") {
			return Route{}, false, errors.New("invalid route destination in netlink response")
		}
		route.Destination = prefix.Masked().String()
		if bits == destination.BitLen() {
			route.Destination = destination.String()
		}
	}
	route.ReadOnly = route.ReadOnly || route.Interface == ""
	return route, true, nil
}

func parseLegacyNextHops(data []byte, names map[uint32]string) ([]NextHop, error) {
	result := make([]NextHop, 0)
	for len(data) != 0 {
		if len(data) < unix.SizeofRtNexthop {
			return nil, errors.New("invalid route next hop in netlink response")
		}
		length := int(binary.NativeEndian.Uint16(data[:2]))
		aligned := (length + 3) &^ 3
		if length < unix.SizeofRtNexthop || aligned > len(data) {
			return nil, errors.New("invalid route next hop length in netlink response")
		}
		hop := NextHop{Interface: names[binary.NativeEndian.Uint32(data[4:8])], Weight: uint32(data[3]) + 1}
		nested := data[unix.SizeofRtNexthop:length]
		for len(nested) != 0 {
			if len(nested) < unix.SizeofRtAttr {
				return nil, errors.New("invalid route next hop attribute in netlink response")
			}
			attrLength := int(binary.NativeEndian.Uint16(nested[:2]))
			attrAligned := (attrLength + 3) &^ 3
			if attrLength < unix.SizeofRtAttr || attrAligned > len(nested) {
				return nil, errors.New("invalid route next hop attribute length in netlink response")
			}
			value := nested[unix.SizeofRtAttr:attrLength]
			switch binary.NativeEndian.Uint16(nested[2:4]) {
			case unix.RTA_GATEWAY:
				hop.Gateway = legacyRouteIP(value)
			case unix.RTA_VIA:
				if len(value) >= 2 {
					hop.Gateway = legacyRouteIP(value[2:])
				}
			}
			nested = nested[attrAligned:]
		}
		result = append(result, hop)
		data = data[aligned:]
	}
	return result, nil
}

func legacyRouteIP(value []byte) string {
	if address, ok := netip.AddrFromSlice(value); ok {
		return address.String()
	}
	return ""
}

func legacyRouteProtocol(value byte) string {
	switch value {
	case unix.RTPROT_UNSPEC:
		return ""
	case unix.RTPROT_REDIRECT:
		return "redirect"
	case unix.RTPROT_KERNEL:
		return "kernel"
	case unix.RTPROT_BOOT:
		return "boot"
	case unix.RTPROT_STATIC:
		return "static"
	case unix.RTPROT_RA:
		return "ra"
	case unix.RTPROT_DHCP:
		return "dhcp"
	default:
		return strconv.Itoa(int(value))
	}
}
