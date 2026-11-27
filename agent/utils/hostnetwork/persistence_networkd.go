package hostnetwork

import (
	"errors"
	"net"
	"net/netip"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type networkSection struct {
	name  string
	lines []string
}

func networkSections(content string) []networkSection {
	sections := []networkSection{{}}
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			sections = append(sections, networkSection{name: strings.Trim(trimmed, "[]"), lines: []string{line}})
		} else {
			last := len(sections) - 1
			sections[last].lines = append(sections[last].lines, line)
		}
	}
	return sections
}

func sectionValues(section networkSection) map[string][]string {
	values := make(map[string][]string)
	for _, line := range section.lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if key, value, ok := strings.Cut(line, "="); ok {
			key = strings.TrimSpace(key)
			values[key] = append(values[key], strings.TrimSpace(value))
		}
	}
	return values
}

func (p *PersistentNetwork) networkd(path string, change PersistentChange) error {
	if !strings.HasPrefix(filepath.Clean(path), "/etc/systemd/network/") {
		return errors.New("generated or vendor networkd configuration is read-only")
	}
	paths := []string{path}
	drops, err := filepath.Glob(path + ".d/*.conf")
	if err != nil {
		return err
	}
	paths = append(paths, drops...)
	for _, dir := range []string{"/run/systemd/network", "/usr/lib/systemd/network"} {
		drops, _ := filepath.Glob(filepath.Join(dir, filepath.Base(path)+".d", "*.conf"))
		if len(drops) > 0 {
			return errors.New("networkd drop-ins from multiple configuration directories are read-only")
		}
	}
	return p.networkdFiles(paths, change)
}

func (p *PersistentNetwork) networkdFiles(paths []string, change PersistentChange) error {
	removedAddresses := make(map[string]int)
	removedRoutes := make([]int, len(change.RemoveRoutes))
	for fileIndex, path := range paths {
		file, err := readNetworkFile(path)
		if err != nil {
			return err
		}
		sections := networkSections(string(file.before))
		if fileIndex == 0 {
			matched := false
			for _, section := range sections {
				if section.name != "Match" {
					continue
				}
				values := sectionValues(section)
				if names := values["Name"]; len(names) == 1 && names[0] == p.Device {
					matched = true
				}
				if macs := values["MACAddress"]; len(macs) == 1 {
					if _, err := net.ParseMAC(macs[0]); err == nil {
						matched = true
					}
				}
			}
			if !matched {
				return errors.New("networkd configuration is not restricted to a single interface")
			}
		} else {
			for _, section := range sections {
				if section.name == "Match" {
					return errors.New("networkd match overrides require source reconciliation before editing")
				}
			}
		}
		for index := range sections {
			section := &sections[index]
			values := sectionValues(*section)
			if section.name == "Network" || section.name == "Address" {
				for _, address := range values["Address"] {
					if address == "" {
						return errors.New("networkd address reset directives are read-only")
					}
					if section.name != "Address" || len(values) == 1 {
						prefix, err := netip.ParsePrefix(address)
						if err != nil {
							return err
						}
						p.Addresses = append(p.Addresses, prefix.String())
					}
				}
				var kept []string
				removedAddress := false
				for _, line := range section.lines {
					key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
					key = strings.TrimSpace(key)
					prefix, prefixErr := netip.ParsePrefix(strings.TrimSpace(value))
					if ok && key == "Address" && prefixErr == nil && slices.Contains(change.RemoveAddresses, prefix.String()) {
						if section.name == "Address" && len(values) > 1 {
							return errors.New("networkd addresses with extra attributes are read-only")
						}
						removedAddresses[prefix.String()]++
						removedAddress = true
						continue
					}
					if ok && key == "Gateway" {
						route, err := persistentRoute("default", strings.TrimSpace(value), "", p.Device)
						if err != nil {
							return err
						}
						route.MetricAuto = true
						p.Routes = append(p.Routes, route)
						remove := false
						for i, r := range change.RemoveRoutes {
							if PersistentRouteEqual(route, r) {
								removedRoutes[i]++
								remove = true
							}
						}
						if remove {
							continue
						}
					}
					kept = append(kept, line)
				}
				if section.name == "Address" && removedAddress && len(sectionValues(networkSection{lines: kept})["Address"]) == 0 {
					kept = nil
				}
				section.lines = kept
			}
			if section.name != "Route" {
				continue
			}
			supported := true
			for key, entries := range values {
				if !slices.Contains([]string{"Destination", "Gateway", "Metric", "Table"}, key) || len(entries) != 1 {
					supported = false
				}
			}
			if !supported {
				continue
			}
			get := func(key string) string {
				if len(values[key]) > 0 {
					return values[key][0]
				}
				return ""
			}
			if get("Table") != "" && get("Table") != "254" && get("Table") != "main" {
				continue
			}
			destination := get("Destination")
			if destination == "" {
				destination = "default"
			}
			route, err := persistentRoute(destination, get("Gateway"), get("Metric"), p.Device)
			if err != nil {
				return err
			}
			route.MetricAuto = get("Metric") == ""
			p.Routes = append(p.Routes, route)
			for i, r := range change.RemoveRoutes {
				if PersistentRouteEqual(route, r) {
					removedRoutes[i]++
					section.lines = nil
				}
			}
		}
		if fileIndex == 0 {
			for _, address := range change.AddAddresses {
				sections = append(sections, networkSection{lines: []string{"", "[Address]", "Address=" + address}})
			}
			for _, route := range change.AddRoutes {
				lines := []string{"", "[Route]", "Destination=" + routeDestination(route.Destination, route.Family), "Metric=" + strconv.FormatUint(uint64(route.Metric), 10)}
				if route.Gateway != "" {
					lines = append(lines, "Gateway="+route.Gateway)
				}
				sections = append(sections, networkSection{lines: lines})
			}
		}
		var lines []string
		for _, section := range sections {
			lines = append(lines, section.lines...)
		}
		file.after = []byte(strings.Join(lines, "\n"))
		p.files = append(p.files, file)
	}
	for _, address := range change.RemoveAddresses {
		if removedAddresses[address] != 1 {
			return errors.New("persistent address source is missing or ambiguous")
		}
	}
	for _, count := range removedRoutes {
		if count != 1 {
			return errors.New("persistent route source is missing or ambiguous")
		}
	}
	return nil
}
