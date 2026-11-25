package hostnetwork

import (
	"errors"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

func interfaceFiles(path string, seen map[string]bool) ([]persistentFile, error) {
	path = filepath.Clean(path)
	if seen[path] {
		return nil, nil
	}
	seen[path] = true
	file, err := readNetworkFile(path)
	if err != nil {
		return nil, err
	}
	result := []persistentFile{file}
	for _, line := range strings.Split(string(file.before), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "source" && fields[0] != "source-directory" {
			continue
		}
		for _, pattern := range fields[1:] {
			if !filepath.IsAbs(pattern) {
				pattern = filepath.Join(filepath.Dir(path), pattern)
			}
			if fields[0] == "source-directory" {
				pattern = filepath.Join(pattern, "*")
			}
			paths, err := filepath.Glob(pattern)
			if err != nil {
				return nil, err
			}
			for _, child := range paths {
				if fields[0] == "source-directory" && strings.Contains(filepath.Base(child), ".") {
					continue
				}
				files, err := interfaceFiles(child, seen)
				if err != nil {
					return nil, err
				}
				result = append(result, files...)
			}
		}
	}
	return result, nil
}

func (p *PersistentNetwork) ifupdown(change PersistentChange) (bool, error) {
	reader := p.reader
	if !reader.ifupdownLoaded {
		reader.ifupdownLoaded = true
		if _, err := os.Stat("/etc/network/interfaces"); !os.IsNotExist(err) {
			if _, err := exec.LookPath("ifquery"); err == nil {
				reader.ifupdownFiles, reader.ifupdownErr = interfaceFiles("/etc/network/interfaces", map[string]bool{})
			}
		}
	}
	if reader.ifupdownErr != nil {
		return false, reader.ifupdownErr
	}
	found, err := p.ifupdownFiles(reader.ifupdownFiles, change)
	if !found || err != nil {
		return found, err
	}
	if _, err := cmd.NewCommandMgr(cmd.WithTimeout(15*time.Second)).RunWithStdout("ifquery", p.Device); err != nil {
		return false, nil
	}
	return true, nil
}

func (p *PersistentNetwork) ifupdownFiles(files []persistentFile, change PersistentChange) (bool, error) {
	var err error
	found := false
	families := map[string]int{}
	removedAddresses := map[string]int{}
	removedRoutes := make([]int, len(change.RemoveRoutes))
	for _, file := range files {
		lines := strings.Split(string(file.before), "\n")
		var output []string
		for start := 0; start < len(lines); {
			fields := strings.Fields(lines[start])
			if len(fields) < 4 || fields[0] != "iface" || fields[1] != p.Device {
				output = append(output, lines[start])
				start++
				continue
			}
			found = true
			p.Backend = "ifupdown"
			family := "IPv4"
			if fields[2] == "inet6" {
				family = "IPv6"
			} else if fields[2] != "inet" {
				return true, errors.New("unsupported ifupdown address family")
			}
			families[family]++
			if families[family] > 1 {
				return true, errors.New("multiple ifupdown stanzas for the same address family are read-only")
			}
			if !slices.Contains([]string{"static", "manual", "dhcp", "auto"}, fields[3]) {
				return true, errors.New("unsupported ifupdown method")
			}
			end := start + 1
			for end < len(lines) {
				next := strings.Fields(lines[end])
				if len(next) > 0 && slices.Contains([]string{"iface", "auto", "allow-auto", "allow-hotplug", "source", "source-directory", "mapping"}, next[0]) {
					break
				}
				end++
			}
			stanza := lines[start:end]
			netmask, metric := "", ""
			for _, line := range stanza {
				f := strings.Fields(line)
				if fields[3] == "static" && len(f) >= 2 && f[0] == "broadcast" && f[1] != "+" && f[1] != "-" {
					return true, errors.New("ifupdown configurations with explicit broadcast addresses are read-only")
				}
				if len(f) == 2 {
					if f[0] == "netmask" {
						netmask = f[1]
					}
					if f[0] == "metric" {
						metric = f[1]
					}
				}
			}
			kept := []string{stanza[0]}
			addressCount := 0
			deletedAddress := false
			for _, line := range stanza[1:] {
				f := strings.Fields(line)
				if len(f) < 2 {
					kept = append(kept, line)
					continue
				}
				if f[0] == "address" && fields[3] == "static" {
					address := f[1]
					if !strings.Contains(address, "/") {
						bits, err := strconv.Atoi(netmask)
						if err != nil {
							mask := net.ParseIP(netmask).To4()
							if mask == nil {
								return true, errors.New("cannot resolve ifupdown address prefix")
							}
							var size int
							bits, size = net.IPMask(mask).Size()
							if size != 32 {
								return true, errors.New("invalid ifupdown netmask")
							}
						}
						address += "/" + strconv.Itoa(bits)
					}
					prefix, err := netip.ParsePrefix(address)
					if err != nil {
						return true, err
					}
					address = prefix.String()
					p.Addresses = append(p.Addresses, address)
					if slices.Contains(change.RemoveAddresses, address) {
						removedAddresses[address]++
						deletedAddress = true
						continue
					}
					if len(change.AddAddresses)+len(change.RemoveAddresses) > 0 {
						_, comment, hasComment := strings.Cut(line, "#")
						line = "    address " + address
						if hasComment {
							line += " #" + comment
						}
					}
					addressCount++
				}
				if slices.Contains([]string{"up", "post-up", "down", "pre-down"}, f[0]) {
					address, ok := parseInterfaceAddress(f[1:], p.Device, f[0] == "down" || f[0] == "pre-down")
					if ok {
						isUp := f[0] == "up" || f[0] == "post-up"
						if isUp {
							p.Addresses = append(p.Addresses, address)
						}
						if slices.Contains(change.RemoveAddresses, address) {
							if isUp {
								removedAddresses[address]++
							}
							continue
						}
					}
				}
				var route Route
				isRoute := false
				isDown := false
				if f[0] == "gateway" && fields[3] == "static" {
					destination := "0.0.0.0/0"
					if family == "IPv6" {
						destination = "::/0"
					}
					route, err = persistentRoute(destination, f[1], metric, p.Device)
					if err != nil {
						return true, err
					}
					route.MetricAuto = metric == ""
					isRoute = true
				} else if slices.Contains([]string{"up", "post-up", "down", "pre-down"}, f[0]) {
					isDown = f[0] == "down" || f[0] == "pre-down"
					route, isRoute = parseInterfaceRoute(f[1:], family, p.Device, isDown)
				}
				if isRoute {
					if !isDown {
						p.Routes = append(p.Routes, route)
					}
					remove := false
					for i, r := range change.RemoveRoutes {
						if PersistentRouteEqual(route, r) {
							remove = true
							if !isDown {
								removedRoutes[i]++
							}
						}
					}
					if remove {
						continue
					}
				}
				if f[0] == "netmask" && len(change.AddAddresses)+len(change.RemoveAddresses) > 0 {
					continue
				}
				kept = append(kept, line)
			}
			for _, address := range change.AddAddresses {
				if strings.Contains(address, ":") == (family == "IPv6") {
					if fields[3] == "static" {
						kept = append(kept, "    address "+address)
					} else {
						flag := "-4"
						if family == "IPv6" {
							flag = "-6"
						}
						kept = append(kept, "    up ip "+flag+" address add "+address+" dev "+p.Device, "    down ip "+flag+" address del "+address+" dev "+p.Device)
					}
					addressCount++
				}
			}
			if deletedAddress && addressCount == 0 && fields[3] == "static" {
				return true, errors.New("cannot remove the last address from a static ifupdown stanza")
			}
			for _, route := range change.AddRoutes {
				if route.Family != family {
					continue
				}
				flag := "-4"
				if family == "IPv6" {
					flag = "-6"
				}
				args := routeDestination(route.Destination, route.Family)
				if route.Gateway != "" {
					args += " via " + route.Gateway
				}
				args += " proto static dev " + p.Device + " metric " + strconv.FormatUint(uint64(route.Metric), 10)
				kept = append(kept, "    up ip "+flag+" route add "+args, "    down ip "+flag+" route del "+args)
			}
			output = append(output, kept...)
			start = end
		}
		file.after = []byte(strings.Join(output, "\n"))
		p.files = append(p.files, file)
	}
	if !found {
		return false, nil
	}
	for _, address := range change.RemoveAddresses {
		if removedAddresses[address] != 1 {
			return true, errors.New("persistent address source is missing or ambiguous")
		}
	}
	for _, count := range removedRoutes {
		if count != 1 {
			return true, errors.New("persistent route source is missing or ambiguous")
		}
	}
	for _, address := range change.AddAddresses {
		family := "IPv4"
		if strings.Contains(address, ":") {
			family = "IPv6"
		}
		if families[family] == 0 {
			return true, errors.New("no ifupdown stanza for this address family")
		}
	}
	for _, route := range change.AddRoutes {
		if families[route.Family] == 0 {
			return true, errors.New("no ifupdown stanza for this address family")
		}
	}
	return true, nil
}

func parseInterfaceRoute(fields []string, family, device string, remove bool) (Route, bool) {
	if len(fields) >= 2 && fields[len(fields)-2] == "||" && fields[len(fields)-1] == "true" {
		fields = fields[:len(fields)-2]
	}
	if len(fields) < 4 || fields[0] != "ip" {
		return Route{}, false
	}
	fields = fields[1:]
	if fields[0] == "-4" || fields[0] == "-6" {
		fields = fields[1:]
	}
	if len(fields) < 3 || fields[0] != "route" || fields[1] != "add" && fields[1] != "del" || (fields[1] == "del") != remove {
		return Route{}, false
	}
	destination := fields[2]
	gateway, metric := "", ""
	fields = fields[3:]
	for len(fields) >= 2 {
		switch fields[0] {
		case "via":
			gateway = fields[1]
		case "metric":
			metric = fields[1]
		case "proto":
			if fields[1] != "static" && fields[1] != "boot" {
				return Route{}, false
			}
		case "dev":
			if fields[1] != device {
				return Route{}, false
			}
		default:
			return Route{}, false
		}
		fields = fields[2:]
	}
	if len(fields) != 0 {
		return Route{}, false
	}
	if destination == "default" && family == "IPv6" {
		destination = "::/0"
	}
	route, err := persistentRoute(destination, gateway, metric, device)
	route.MetricAuto = metric == ""
	return route, err == nil
}

func parseInterfaceAddress(fields []string, device string, remove bool) (string, bool) {
	if len(fields) >= 2 && fields[len(fields)-2] == "||" && fields[len(fields)-1] == "true" {
		fields = fields[:len(fields)-2]
	}
	if len(fields) < 6 || fields[0] != "ip" {
		return "", false
	}
	fields = fields[1:]
	if fields[0] == "-4" || fields[0] == "-6" {
		fields = fields[1:]
	}
	if len(fields) != 5 || fields[0] != "address" && fields[0] != "addr" || fields[1] != "add" && fields[1] != "del" || (fields[1] == "del") != remove || fields[3] != "dev" || fields[4] != device {
		return "", false
	}
	prefix, err := netip.ParsePrefix(fields[2])
	if err != nil {
		return "", false
	}
	return prefix.String(), true
}
