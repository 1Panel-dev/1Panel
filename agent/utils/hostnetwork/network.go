package hostnetwork

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type Link struct {
	Index     int       `json:"ifindex"`
	Name      string    `json:"ifname"`
	MTU       int       `json:"mtu"`
	MAC       string    `json:"address"`
	Flags     []string  `json:"flags"`
	OperState string    `json:"operstate"`
	Addresses []Address `json:"addr_info"`
}

type Address struct {
	Family    string  `json:"family"`
	Local     string  `json:"local"`
	PrefixLen int     `json:"prefixlen"`
	Scope     string  `json:"scope"`
	Label     string  `json:"label"`
	Secondary bool    `json:"secondary"`
	Dynamic   bool    `json:"dynamic"`
	Tentative bool    `json:"tentative"`
	DadFailed bool    `json:"dadfailed"`
	ValidTime *uint32 `json:"valid_life_time"`
}

func (a Address) CIDR() string {
	return fmt.Sprintf("%s/%d", a.Local, a.PrefixLen)
}

type AddressRequest struct {
	Interface string
	Address   string
}

type Route struct {
	MetricAuto      bool
	PreferredSource string
	Destination     string
	Gateway         string
	Interface       string
	Metric          uint32
	Table           string
	Protocol        string
	Scope           string
	Family          string
	ReadOnly        bool
	NextHops        []NextHop
}

type NextHop struct {
	Gateway   string
	Interface string
	Weight    uint32
}

type RouteCheck struct {
	Reachable bool
	Message   string
}

type IPRoute2Error struct{ Err error }

func (e *IPRoute2Error) Error() string { return e.Err.Error() }
func (e *IPRoute2Error) Unwrap() error { return e.Err }

func readIPJSON(ctx context.Context, timeout time.Duration, args []string, value interface{}) error {
	command, err := exec.LookPath("ip")
	if err != nil {
		return &IPRoute2Error{Err: err}
	}
	output, err := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(timeout)).RunWithStdout(command, append([]string{"-j"}, args...)...)
	if err != nil {
		message := strings.ToLower(err.Error())
		jsonOption := strings.Contains(message, "-j") || strings.Contains(message, "'j'") || strings.Contains(message, `"j"`)
		busyBoxUsage := strings.Contains(message, "busybox") && strings.Contains(message, "usage: ip ")
		if busyBoxUsage || jsonOption && (strings.Contains(message, "unknown") || strings.Contains(message, "invalid option") || strings.Contains(message, "unrecognized")) {
			return &IPRoute2Error{Err: err}
		}
		return err
	}
	if err := json.Unmarshal([]byte(output), value); err != nil {
		return &IPRoute2Error{Err: err}
	}
	return nil
}

func canReadLegacy(ctx context.Context, err error) bool {
	var unavailable *IPRoute2Error
	if !errors.As(err, &unavailable) {
		return false
	}
	version, err := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(3*time.Second)).RunWithStdout("ip", "-Version")
	return err == nil && strings.HasPrefix(strings.TrimSpace(version), "ip utility, iproute2-")
}

func Links(name string) ([]Link, error) {
	args := []string{"address", "show"}
	if name != "" {
		args = append(args, "dev", name)
	}
	var links []Link
	if err := readIPJSON(context.Background(), 10*time.Second, args, &links); err != nil {
		if canReadLegacy(context.Background(), err) {
			return readLegacyLinks(name)
		}
		return nil, err
	}
	if slices.ContainsFunc(links, func(link Link) bool { return link.Name == "" || link.Index == 0 }) {
		return readLegacyLinks(name)
	}
	return links, nil
}

func WaitAddress(req AddressRequest) error {
	deadline := time.Now().Add(5 * time.Second)
	for {
		links, err := Links(req.Interface)
		if err != nil {
			return err
		}
		var address *Address
		for _, link := range links {
			if link.Name != req.Interface {
				continue
			}
			for _, item := range link.Addresses {
				if item.CIDR() == req.Address {
					address = &item
					break
				}
			}
			if address != nil {
				break
			}
		}
		switch {
		case address == nil:
			return errors.New("interface address is missing")
		case address.DadFailed:
			return errors.New("IPv6 duplicate address detection failed")
		case !address.Tentative:
			return nil
		case time.Now().After(deadline):
			return errors.New("IPv6 duplicate address detection did not finish")
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func ChangeSystemAddress(req AddressRequest, operation string) error {
	if _, err := net.InterfaceByName(req.Interface); err != nil {
		return err
	}
	family := "-4"
	if strings.Contains(req.Address, ":") {
		family = "-6"
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(10*time.Second)).Run("ip", family, "address", operation, req.Address, "dev", req.Interface)
}

func PreserveSecondaryAddresses(device string, change PersistentChange) (func() error, error) {
	for _, removed := range change.RemoveAddresses {
		oldPrefix, err := netip.ParsePrefix(removed)
		if err != nil || !oldPrefix.Addr().Is4() {
			continue
		}
		for _, added := range change.AddAddresses {
			newPrefix, err := netip.ParsePrefix(added)
			if err != nil || oldPrefix.Masked() != newPrefix.Masked() {
				continue
			}
			if _, err := net.InterfaceByName(device); err != nil {
				return nil, err
			}
			path := filepath.Join("/proc/sys/net/ipv4/conf", device, "promote_secondaries")
			previous, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(string(previous)) == "1" {
				return nil, nil
			}
			if err := os.WriteFile(path, []byte("1"), 0644); err != nil {
				return nil, err
			}
			return func() error { return os.WriteFile(path, previous, 0644) }, nil
		}
	}
	return nil, nil
}

const networkAdminCap = 1 << 12

func HasAdmin() bool {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if value, ok := strings.CutPrefix(line, "CapEff:"); ok {
			caps, err := strconv.ParseUint(strings.TrimSpace(value), 16, 64)
			return err == nil && caps&networkAdminCap != 0
		}
	}
	return false
}

func IsReadOnlyVirtualInterface(device string) bool {
	for _, prefix := range []string{"docker", "br-", "veth", "tun", "tap", "virbr", "vnet", "vmnet", "zt"} {
		if strings.HasPrefix(device, prefix) {
			return true
		}
	}
	return false
}
