package hostnetwork

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type PersistentChange struct {
	AddAddresses    []string
	RemoveAddresses []string
	AddRoutes       []Route
	RemoveRoutes    []Route
}

type persistentFile struct {
	path   string
	before []byte
	after  []byte
	mode   os.FileMode
}

type PersistentNetworkReader struct {
	netplanLoaded  bool
	netplanFiles   []persistentFile
	netplanErr     error
	connections    map[string]string
	ifupdownLoaded bool
	ifupdownFiles  []persistentFile
	ifupdownErr    error
}

type PersistentNetwork struct {
	reader             *PersistentNetworkReader
	Backend            string
	Device             string
	Addresses          []string
	Routes             []Route
	files              []persistentFile
	connection         string
	properties         []string
	previousProperties []string
}

func PreparePersistentNetwork(device string, change PersistentChange) (*PersistentNetwork, error) {
	reader := &PersistentNetworkReader{}
	return reader.Prepare(device, change)
}

func (reader *PersistentNetworkReader) Prepare(device string, change PersistentChange) (result *PersistentNetwork, resultErr error) {
	defer func() {
		if resultErr != nil {
			result = nil
		}
	}()
	defer func() {
		if resultErr != nil || result == nil {
			return
		}
		for _, address := range change.RemoveAddresses {
			count := 0
			for _, existing := range result.Addresses {
				if existing == address {
					count++
				}
			}
			if count != 1 {
				resultErr = errors.New("persistent address source is missing or ambiguous")
				return
			}
		}
		for _, route := range change.RemoveRoutes {
			count := 0
			for _, existing := range result.Routes {
				if PersistentRouteEqual(existing, route) {
					count++
				}
			}
			if count != 1 {
				resultErr = errors.New("persistent route source is missing or ambiguous")
				return
			}
		}
		for _, address := range change.AddAddresses {
			if slices.Contains(result.Addresses, address) && !slices.Contains(change.RemoveAddresses, address) {
				resultErr = errors.New("address already exists in persistent configuration")
				return
			}
		}
		for _, route := range change.AddRoutes {
			if slices.ContainsFunc(result.Routes, func(r Route) bool {
				return PersistentRouteEqual(r, route) && !slices.ContainsFunc(change.RemoveRoutes, func(old Route) bool { return PersistentRouteEqual(r, old) })
			}) {
				resultErr = errors.New("route already exists in persistent configuration")
				return
			}
		}
	}()
	if IsReadOnlyVirtualInterface(device) {
		return nil, buserr.New("ErrNetworkVirtualBridgeReadOnly")
	}
	iface, err := net.InterfaceByName(device)
	if err != nil {
		return nil, err
	}
	if iface.Flags&net.FlagLoopback != 0 {
		return nil, buserr.New("ErrNetworkLoopbackReadOnly")
	}
	p := &PersistentNetwork{Device: device, reader: reader}
	found, err := p.netplan(change)
	if err != nil {
		return nil, err
	}
	if found {
		return p, nil
	}
	if reader.connections == nil {
		reader.connections = make(map[string]string)
		output, err := cmd.NewCommandMgr(cmd.WithTimeout(15*time.Second)).RunWithStdout("nmcli", "--terse", "--mode", "multiline", "--escape", "no", "--fields", "GENERAL.DEVICE,GENERAL.CON-UUID", "device", "show")
		if err == nil {
			var name string
			for _, line := range strings.Split(output, "\n") {
				key, value, _ := strings.Cut(line, ":")
				value = strings.TrimSpace(value)
				switch key {
				case "GENERAL.DEVICE":
					name = value
				case "GENERAL.CON-UUID":
					if name != "" && value != "" && value != "--" {
						reader.connections[name] = value
					}
				}
			}
		}
	}
	if uuid := reader.connections[device]; uuid != "" {
		p.Backend = "NetworkManager"
		p.connection = uuid
		return p, p.networkManager(change)
	}
	data, e := os.ReadFile(filepath.Join("/run/systemd/netif/links", strconv.Itoa(iface.Index)))
	if e == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if file, ok := strings.CutPrefix(line, "NETWORK_FILE="); ok && file != "" {
				p.Backend = "networkd"
				return p, p.networkd(strings.Trim(file, "\""), change)
			}
		}
	}
	found, err = p.ifupdown(change)
	if err != nil {
		return nil, err
	}
	if found {
		return p, nil
	}
	return nil, buserr.WithName("ErrNetworkPersistentSourceMissing", device)
}

func (p *PersistentNetwork) Apply() error {
	if p.Backend == "NetworkManager" {
		if len(p.properties) == 0 {
			return nil
		}
		for i := 0; i+1 < len(p.previousProperties); i += 2 {
			value, err := cmd.NewCommandMgr(cmd.WithTimeout(15*time.Second)).RunWithStdout("nmcli", "--escape", "no", "-g", p.previousProperties[i], "connection", "show", "uuid", p.connection)
			value = strings.TrimSpace(value)
			if value == "--" {
				value = ""
			}
			if err != nil || value != p.previousProperties[i+1] {
				return errors.New("NetworkManager connection changed externally")
			}
		}
		_, err := cmd.NewCommandMgr(cmd.WithTimeout(15*time.Second)).RunWithStdout("nmcli", append([]string{"connection", "modify", "uuid", p.connection}, p.properties...)...)
		if err != nil {
			return errors.Join(err, p.Rollback())
		}
		return nil
	}
	for _, file := range p.files {
		current, err := os.ReadFile(file.path)
		if err != nil || !bytes.Equal(current, file.before) {
			return fmt.Errorf("network configuration changed externally: %s", file.path)
		}
	}
	for _, file := range p.files {
		if bytes.Equal(file.before, file.after) {
			continue
		}
		if err := writeNetworkFile(file.path, file.after, file.mode); err != nil {
			return errors.Join(err, p.Rollback())
		}
	}
	if p.Backend == "Netplan" {
		if _, err := cmd.NewCommandMgr(cmd.WithTimeout(15*time.Second)).RunWithStdout("netplan", "generate"); err != nil {
			return errors.Join(err, p.Rollback())
		}
	}
	return nil
}

func (p *PersistentNetwork) Rollback() error {
	if p.Backend == "NetworkManager" {
		if len(p.previousProperties) == 0 {
			return nil
		}
		_, err := cmd.NewCommandMgr(cmd.WithTimeout(15*time.Second)).RunWithStdout("nmcli", append([]string{"connection", "modify", "uuid", p.connection}, p.previousProperties...)...)
		return err
	}
	var failures []error
	for _, file := range p.files {
		if bytes.Equal(file.before, file.after) {
			continue
		}
		current, err := os.ReadFile(file.path)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		if bytes.Equal(current, file.before) {
			continue
		}
		if !bytes.Equal(current, file.after) {
			failures = append(failures, fmt.Errorf("refusing to overwrite externally changed file %s", file.path))
			continue
		}
		if err := writeNetworkFile(file.path, file.before, file.mode); err != nil {
			failures = append(failures, err)
		}
	}
	if p.Backend == "Netplan" {
		_, err := cmd.NewCommandMgr(cmd.WithTimeout(15*time.Second)).RunWithStdout("netplan", "generate")
		if err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func readNetworkFile(path string) (persistentFile, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return persistentFile{}, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0022 != 0 {
		return persistentFile{}, fmt.Errorf("unsafe network configuration file: %s", path)
	}
	data, err := os.ReadFile(path)
	return persistentFile{path: path, before: data, after: data, mode: info.Mode().Perm()}, err
}

func writeNetworkFile(path string, data []byte, mode os.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".1panel-network-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if err = file.Chmod(mode); err == nil {
		_, err = file.Write(data)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = os.Rename(file.Name(), path); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func PersistentRouteEqual(a, b Route) bool {
	return routeDestination(a.Destination, a.Family) == routeDestination(b.Destination, b.Family) && a.Family == b.Family && a.Gateway == b.Gateway && (a.MetricAuto || a.Metric == b.Metric)
}

func routeDestination(value, family string) string {
	if value == "default" || value == "0.0.0.0/0" || value == "::/0" {
		if family == "IPv6" {
			return "::/0"
		}
		return "0.0.0.0/0"
	}
	if p, err := netip.ParsePrefix(value); err == nil {
		return p.Masked().String()
	}
	if a, err := netip.ParseAddr(value); err == nil {
		return netip.PrefixFrom(a, a.BitLen()).String()
	}
	return value
}

func persistentRoute(destination, gateway, metric, device string) (Route, error) {
	family := "IPv4"
	if strings.Contains(destination+gateway, ":") {
		family = "IPv6"
	}
	value, err := strconv.ParseUint(metric, 10, 32)
	if metric == "" {
		err = nil
	}
	if err != nil {
		return Route{}, err
	}
	route := Route{Destination: routeDestination(destination, family), Gateway: gateway, Metric: uint32(value), Interface: device, Family: family, Table: "main", Protocol: "static"}
	if _, err := netip.ParsePrefix(route.Destination); err != nil {
		return Route{}, err
	}
	if gateway != "" {
		a, err := netip.ParseAddr(gateway)
		if err != nil || a.Is6() != (family == "IPv6") {
			return Route{}, errors.New("invalid persistent route gateway")
		}
	}
	return route, nil
}
