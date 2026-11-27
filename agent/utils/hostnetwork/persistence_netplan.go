package hostnetwork

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func yamlField(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func yamlValue(node *yaml.Node, key string) string {
	value := yamlField(node, key)
	if value == nil {
		return ""
	}
	return value.Value
}

func setYAMLField(node *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			if value == nil {
				node.Content = append(node.Content[:i], node.Content[i+2:]...)
			} else {
				node.Content[i+1] = value
			}
			return
		}
	}
	if value != nil {
		node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
	}
}

func (p *PersistentNetwork) netplan(change PersistentChange) (bool, error) {
	iface, err := net.InterfaceByName(p.Device)
	if err != nil {
		return false, err
	}
	reader := p.reader
	if !reader.netplanLoaded {
		reader.netplanLoaded = true
		paths, err := filepath.Glob("/etc/netplan/*.yaml")
		reader.netplanErr = err
		if err == nil {
			for _, path := range paths {
				file, err := readNetworkFile(path)
				if err != nil {
					reader.netplanErr = err
					break
				}
				reader.netplanFiles = append(reader.netplanFiles, file)
			}
		}
	}
	if reader.netplanErr != nil {
		return false, reader.netplanErr
	}
	return p.netplanFiles(reader.netplanFiles, iface.HardwareAddr.String(), change)
}

func (p *PersistentNetwork) netplanFiles(files []persistentFile, macAddress string, change PersistentChange) (bool, error) {
	var err error
	var document yaml.Node
	var target *yaml.Node
	var selected persistentFile
	for _, file := range files {
		var doc yaml.Node
		if err = yaml.Unmarshal(file.before, &doc); err != nil {
			return false, err
		}
		if len(doc.Content) == 0 {
			continue
		}
		network := yamlField(doc.Content[0], "network")
		ethernets := yamlField(network, "ethernets")
		if ethernets == nil {
			continue
		}
		for i := 0; i+1 < len(ethernets.Content); i += 2 {
			id, node := ethernets.Content[i].Value, ethernets.Content[i+1]
			match := yamlField(node, "match")
			matched := id == p.Device && match == nil || yamlValue(node, "set-name") == p.Device
			if match != nil {
				matched = true
				for j := 0; j+1 < len(match.Content); j += 2 {
					switch match.Content[j].Value {
					case "name":
						ok, _ := filepath.Match(match.Content[j+1].Value, p.Device)
						matched = matched && (ok || yamlValue(node, "set-name") == p.Device)
					case "macaddress":
						matched = matched && strings.EqualFold(match.Content[j+1].Value, macAddress)
					default:
						matched = false
					}
				}
			}
			if !matched {
				continue
			}
			if match != nil && yamlValue(match, "macaddress") == "" {
				name := yamlValue(match, "name")
				if name == "" || strings.ContainsAny(name, "*?[") {
					return true, errors.New("Netplan interface patterns matching multiple devices are read-only")
				}
			}
			if target != nil {
				return true, errors.New("multiple Netplan definitions match this interface; configuration is read-only")
			}
			if node.Kind != yaml.MappingNode || yamlField(node, "<<") != nil {
				return true, errors.New("merged Netplan interface definitions are read-only")
			}
			pending := []*yaml.Node{node}
			for len(pending) > 0 {
				current := pending[len(pending)-1]
				pending = pending[:len(pending)-1]
				if current.Anchor != "" || current.Kind == yaml.AliasNode {
					return true, errors.New("Netplan interface aliases and anchors are read-only")
				}
				pending = append(pending, current.Content...)
			}
			if strings.Contains(strings.ToLower(filepath.Base(file.path)), "cloud-init") {
				return true, errors.New("cloud-init generated network configuration is read-only")
			}
			selected = file
			document = doc
			target = node
		}
	}
	if target == nil {
		return false, nil
	}
	for _, dir := range []string{"/run/netplan", "/lib/netplan"} {
		extra, _ := filepath.Glob(filepath.Join(dir, "*.yaml"))
		if len(extra) > 0 {
			return true, fmt.Errorf("Netplan configuration in %s requires source reconciliation before editing", dir)
		}
	}
	p.Backend = "Netplan"
	addresses := yamlField(target, "addresses")
	if addresses == nil {
		addresses = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	}
	if addresses.Kind != yaml.SequenceNode {
		return true, errors.New("unsupported Netplan addresses")
	}
	for _, address := range addresses.Content {
		if address.Kind != yaml.ScalarNode {
			return true, errors.New("Netplan address mappings are read-only")
		}
		prefix, err := netip.ParsePrefix(address.Value)
		if err != nil {
			return true, err
		}
		p.Addresses = append(p.Addresses, prefix.String())
	}
	for _, address := range change.RemoveAddresses {
		index := slices.IndexFunc(addresses.Content, func(n *yaml.Node) bool {
			prefix, err := netip.ParsePrefix(n.Value)
			return err == nil && prefix.String() == address
		})
		if index < 0 {
			return true, errors.New("address is not present in the persistent Netplan configuration")
		}
		addresses.Content = slices.Delete(addresses.Content, index, index+1)
	}
	for _, address := range change.AddAddresses {
		addresses.Content = append(addresses.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: address})
	}
	if len(change.AddAddresses)+len(change.RemoveAddresses) > 0 {
		setYAMLField(target, "addresses", addresses)
	}
	routes := yamlField(target, "routes")
	if routes == nil {
		routes = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	}
	if routes.Kind != yaml.SequenceNode {
		return true, errors.New("unsupported Netplan routes")
	}
	var editableNodes []*yaml.Node
	for _, node := range routes.Content {
		supported := node.Kind == yaml.MappingNode
		for i := 0; i+1 < len(node.Content); i += 2 {
			if !slices.Contains([]string{"to", "via", "metric", "table"}, node.Content[i].Value) {
				supported = false
			}
		}
		table := yamlValue(node, "table")
		if table != "" && table != "254" {
			supported = false
		}
		if !supported {
			continue
		}
		route, err := persistentRoute(yamlValue(node, "to"), yamlValue(node, "via"), yamlValue(node, "metric"), p.Device)
		if err != nil {
			return true, err
		}
		route.MetricAuto = yamlValue(node, "metric") == ""
		p.Routes = append(p.Routes, route)
		editableNodes = append(editableNodes, node)
	}
	for _, key := range []string{"gateway4", "gateway6"} {
		if value := yamlValue(target, key); value != "" {
			destination := "0.0.0.0/0"
			if key == "gateway6" {
				destination = "::/0"
			}
			route, err := persistentRoute(destination, value, "", p.Device)
			if err != nil {
				return true, err
			}
			route.MetricAuto = true
			p.Routes = append(p.Routes, route)
			editableNodes = append(editableNodes, yamlField(target, key))
		}
	}
	for _, remove := range change.RemoveRoutes {
		index := slices.IndexFunc(p.Routes, func(route Route) bool { return PersistentRouteEqual(route, remove) })
		if index < 0 {
			return true, errors.New("route is not present in the persistent Netplan configuration")
		}
		node := editableNodes[index]
		if node == yamlField(target, "gateway4") {
			setYAMLField(target, "gateway4", nil)
		} else if node == yamlField(target, "gateway6") {
			setYAMLField(target, "gateway6", nil)
		} else {
			routes.Content = slices.DeleteFunc(routes.Content, func(n *yaml.Node) bool { return n == node })
		}
	}
	for _, route := range change.AddRoutes {
		node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		setYAMLField(node, "to", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: routeDestination(route.Destination, route.Family)})
		if route.Gateway != "" {
			setYAMLField(node, "via", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: route.Gateway})
		}
		setYAMLField(node, "metric", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.FormatUint(uint64(route.Metric), 10)})
		routes.Content = append(routes.Content, node)
	}
	if len(change.AddRoutes)+len(change.RemoveRoutes) > 0 {
		setYAMLField(target, "routes", routes)
	}
	if len(change.AddAddresses)+len(change.RemoveAddresses)+len(change.AddRoutes)+len(change.RemoveRoutes) > 0 {
		selected.after, err = yaml.Marshal(&document)
		if err != nil {
			return true, err
		}
	}
	p.files = append(p.files, selected)
	return true, nil
}
