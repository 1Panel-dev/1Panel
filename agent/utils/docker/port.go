package docker

import (
	"fmt"
	"sort"
	"strings"

	containertypes "github.com/docker/docker/api/types/container"
)

// PortRangeItem describes the dimensions shared by Docker port range displays.
// Key contains dimensions that must stay equal (for example address, protocol,
// and an optional protection policy identity).
type PortRangeItem struct {
	Key            string
	PublicPort     uint16
	PrivatePort    uint16
	HasPrivatePort bool
	Position       int
}

type PortRange struct {
	Start PortRangeItem
	End   PortRangeItem
	Items []PortRangeItem
}

func MergePortRanges(items []PortRangeItem) []PortRange {
	sorted := append([]PortRangeItem(nil), items...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Key != sorted[j].Key {
			return sorted[i].Key < sorted[j].Key
		}
		if sorted[i].PublicPort != sorted[j].PublicPort {
			return sorted[i].PublicPort < sorted[j].PublicPort
		}
		return sorted[i].PrivatePort < sorted[j].PrivatePort
	})

	ranges := make([]PortRange, 0, len(sorted))
	for _, item := range sorted {
		if len(ranges) == 0 {
			ranges = append(ranges, PortRange{Start: item, End: item, Items: []PortRangeItem{item}})
			continue
		}
		last := &ranges[len(ranges)-1]
		publicContinuous := (last.End.PublicPort == 0 && item.PublicPort == 0) || item.PublicPort == last.End.PublicPort+1
		privateContinuous := (!last.End.HasPrivatePort && !item.HasPrivatePort) ||
			(last.End.HasPrivatePort && item.HasPrivatePort && item.PrivatePort == last.End.PrivatePort+1)
		if item.Key == last.End.Key && publicContinuous && privateContinuous {
			last.End = item
			last.Items = append(last.Items, item)
			continue
		}
		ranges = append(ranges, PortRange{Start: item, End: item, Items: []PortRangeItem{item}})
	}
	return ranges
}

func SimplifyPorts(ports []containertypes.Port) []string {
	items := make([]PortRangeItem, 0, len(ports))
	for i, port := range ports {
		family := "ipv4"
		if strings.Contains(port.IP, ":") {
			family = "ipv6"
		}
		items = append(items, PortRangeItem{
			Key:            family + "|" + port.IP + "|" + port.Type,
			PublicPort:     port.PublicPort,
			PrivatePort:    port.PrivatePort,
			HasPrivatePort: true,
			Position:       i,
		})
	}

	result := make([]string, 0, len(items))
	for _, portRange := range MergePortRanges(items) {
		start, end := ports[portRange.Start.Position], ports[portRange.End.Position]
		ip := ""
		if start.IP != "" {
			ip = start.IP + ":"
		}
		privatePorts := fmt.Sprintf("%d", start.PrivatePort)
		if start.PrivatePort != end.PrivatePort {
			privatePorts = fmt.Sprintf("%d-%d", start.PrivatePort, end.PrivatePort)
		}
		value := fmt.Sprintf("%s%s/%s", ip, privatePorts, start.Type)
		if start.PublicPort != 0 {
			publicPorts := fmt.Sprintf("%d", start.PublicPort)
			if start.PublicPort != end.PublicPort {
				publicPorts = fmt.Sprintf("%d-%d", start.PublicPort, end.PublicPort)
			}
			value = fmt.Sprintf("%s%s->%s/%s", ip, publicPorts, privatePorts, start.Type)
		}
		result = append(result, value)
	}
	return result
}
