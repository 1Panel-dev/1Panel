package hostnetwork

import (
	"errors"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (p *PersistentNetwork) networkManager(change PersistentChange) error {
	keys := []string{"addresses", "routes", "gateway", "route-metric", "route-table", "never-default", "method"}
	var fields []string
	for _, family := range []string{"ipv4", "ipv6"} {
		for _, key := range keys {
			fields = append(fields, family+"."+key)
		}
	}
	output, err := cmd.NewCommandMgr(cmd.WithTimeout(15*time.Second)).RunWithStdout("nmcli", "--terse", "--mode", "multiline", "--escape", "no", "--fields", strings.Join(fields, ","), "connection", "show", "uuid", p.connection)
	if err != nil {
		return err
	}
	properties := make(map[string]string, len(fields))
	for _, line := range strings.Split(output, "\n") {
		if key, value, ok := strings.Cut(line, ":"); ok {
			properties[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	for _, family := range []string{"ipv4", "ipv6"} {
		is6 := family == "ipv6"
		values := make(map[string]string, len(keys))
		for _, key := range keys {
			value, ok := properties[family+"."+key]
			if !ok {
				return fmt.Errorf("missing NetworkManager property %s.%s", family, key)
			}
			if value == "--" {
				value = ""
			}
			values[key] = value
		}
		table := strings.Fields(values["route-table"])
		if len(table) > 0 && table[0] != "0" && table[0] != "254" {
			return errors.New("NetworkManager policy routing profiles are read-only")
		}
		addresses := []string{}
		for _, raw := range strings.Split(values["addresses"], ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			address, err := netip.ParsePrefix(raw)
			if err != nil {
				return err
			}
			addresses = append(addresses, address.String())
		}
		p.Addresses = append(p.Addresses, addresses...)
		changed := false
		for _, remove := range change.RemoveAddresses {
			if strings.Contains(remove, ":") != is6 {
				continue
			}
			index := slices.Index(addresses, remove)
			if index < 0 {
				return errors.New("address is not present in the NetworkManager connection")
			}
			addresses = slices.Delete(addresses, index, index+1)
			changed = true
		}
		for _, add := range change.AddAddresses {
			if strings.Contains(add, ":") == is6 {
				addresses = append(addresses, add)
				changed = true
			}
		}
		if changed {
			if values["method"] == "manual" && len(addresses) == 0 {
				return errors.New("cannot remove the last address from a manual NetworkManager connection")
			}
			if values["method"] != "manual" && values["method"] != "auto" {
				return fmt.Errorf("NetworkManager method %s is read-only", values["method"])
			}
			p.properties = append(p.properties, family+".addresses", strings.Join(addresses, ","))
			p.previousProperties = append(p.previousProperties, family+".addresses", values["addresses"])
		}
		rawRoutes := []string{}
		for _, raw := range strings.Split(values["routes"], ",") {
			raw = strings.TrimSpace(raw)
			if raw != "" {
				rawRoutes = append(rawRoutes, raw)
			}
		}
		var parsed []Route
		for _, raw := range rawRoutes {
			fields := strings.Fields(raw)
			if len(fields) == 0 || len(fields) > 3 {
				return errors.New("NetworkManager routes with additional attributes are read-only")
			}
			gateway, metric := "", ""
			remaining := fields[1:]
			if len(remaining) > 0 {
				if address, err := netip.ParseAddr(remaining[0]); err == nil {
					gateway = address.String()
					remaining = remaining[1:]
				}
			}
			if len(remaining) > 1 {
				return errors.New("NetworkManager routes with additional attributes are read-only")
			}
			if len(remaining) == 1 {
				metric = remaining[0]
			}
			if gateway == "0.0.0.0" || gateway == "::" {
				gateway = ""
			}
			if metric == "-1" {
				metric = ""
			}
			route, err := persistentRoute(fields[0], gateway, metric, p.Device)
			if err != nil {
				return err
			}
			route.MetricAuto = metric == ""
			parsed = append(parsed, route)
		}
		p.Routes = append(p.Routes, parsed...)
		gateway := values["gateway"]
		var defaultRoute Route
		if gateway != "" {
			destination := "0.0.0.0/0"
			if is6 {
				destination = "::/0"
			}
			metric := values["route-metric"]
			if metric == "-1" {
				metric = ""
			}
			var err error
			defaultRoute, err = persistentRoute(destination, gateway, metric, p.Device)
			if err != nil {
				return err
			}
			defaultRoute.MetricAuto = metric == ""
			p.Routes = append(p.Routes, defaultRoute)
		}
		changed = false
		for _, remove := range change.RemoveRoutes {
			if (remove.Family == "IPv6") != is6 {
				continue
			}
			if gateway != "" && PersistentRouteEqual(defaultRoute, remove) {
				p.properties = append(p.properties, family+".gateway", "")
				p.previousProperties = append(p.previousProperties, family+".gateway", gateway)
				continue
			}
			index := slices.IndexFunc(parsed, func(r Route) bool { return PersistentRouteEqual(r, remove) })
			if index < 0 {
				return errors.New("route is not present in the NetworkManager connection")
			}
			rawRoutes = slices.Delete(rawRoutes, index, index+1)
			parsed = slices.Delete(parsed, index, index+1)
			changed = true
		}
		for _, route := range change.AddRoutes {
			if (route.Family == "IPv6") != is6 {
				continue
			}
			destination := routeDestination(route.Destination, route.Family)
			if (destination == "0.0.0.0/0" || destination == "::/0") && values["never-default"] == "yes" {
				return errors.New("this NetworkManager connection forbids default routes")
			}
			gateway := route.Gateway
			if gateway == "" {
				gateway = "0.0.0.0"
				if is6 {
					gateway = "::"
				}
			}
			rawRoutes = append(rawRoutes, routeDestination(route.Destination, route.Family)+" "+gateway+" "+strconv.FormatUint(uint64(route.Metric), 10))
			changed = true
		}
		if changed {
			p.properties = append(p.properties, family+".routes", strings.Join(rawRoutes, ","))
			p.previousProperties = append(p.previousProperties, family+".routes", values["routes"])
		}
	}
	return nil
}
