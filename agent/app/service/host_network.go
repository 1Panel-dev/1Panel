package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"slices"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/hostnetwork"
	"golang.org/x/sync/errgroup"
)

type HostNetworkService struct{}

var hostNetworkMu sync.Mutex

func (s *HostNetworkService) UpdateAddress(change dto.HostNetworkAddressChange, remove bool) error {
	hostNetworkMu.Lock()
	defer hostNetworkMu.Unlock()
	if _, err := hostNetworkPersistenceSupport(); err != nil {
		return err
	}
	req, prefix, err := normalizeHostNetworkAddress(change.HostNetworkAddressRequest)
	if err != nil {
		return err
	}
	if change.Confirmation != req.Interface {
		return errors.New("network change confirmation does not match the interface")
	}
	links, err := hostnetwork.Links("")
	if err != nil {
		return hostNetworkSystemError(err)
	}
	delta := hostnetwork.PersistentChange{}
	previous := req
	if change.PreviousAddress != "" {
		if remove {
			return errors.New("previous address is not valid for deletion")
		}
		previous, _, err = normalizeHostNetworkAddress(dto.HostNetworkAddressRequest{Interface: req.Interface, Address: change.PreviousAddress})
		if err != nil {
			return err
		}
		if previous.Address == req.Address {
			return nil
		}
	}
	if remove || change.PreviousAddress != "" {
		address := findHostNetworkAddress(links, previous)
		if address == nil || address.Dynamic || address.Scope != "global" {
			return errors.New("only existing static global addresses can be changed")
		}
		_, oldPrefix, _ := normalizeHostNetworkAddress(previous)
		if err := guardHostNetworkPrimaryAddress(links, previous, oldPrefix, address); err != nil {
			return err
		}
		delta.RemoveAddresses = append(delta.RemoveAddresses, previous.Address)
	}
	if !remove {
		if findHostNetworkAddress(links, req) != nil {
			return buserr.New("ErrNetworkPersistenceConflict")
		}
		if !prefix.Addr().IsGlobalUnicast() {
			return errors.New("only global unicast interface addresses are supported")
		}
		delta.AddAddresses = append(delta.AddAddresses, req.Address)
	}
	plan, err := hostnetwork.PreparePersistentNetwork(req.Interface, delta)
	if err != nil {
		return hostNetworkSystemError(err)
	}
	if len(delta.RemoveAddresses) > 0 {
		var before, after []netip.Prefix
		for _, link := range links {
			if link.Name != req.Interface {
				continue
			}
			for _, address := range link.Addresses {
				prefix, err := netip.ParsePrefix(address.CIDR())
				if err != nil {
					continue
				}
				before = append(before, prefix)
				if !slices.Contains(delta.RemoveAddresses, address.CIDR()) {
					after = append(after, prefix)
				}
			}
		}
		for _, address := range delta.AddAddresses {
			prefix, _ := netip.ParsePrefix(address)
			after = append(after, prefix)
		}
		for _, route := range plan.Routes {
			gateway, err := netip.ParseAddr(route.Gateway)
			if err != nil {
				continue
			}
			reachableBefore := slices.ContainsFunc(before, func(prefix netip.Prefix) bool { return prefix.Contains(gateway) })
			reachableAfter := slices.ContainsFunc(after, func(prefix netip.Prefix) bool { return prefix.Contains(gateway) })
			if reachableBefore && !reachableAfter {
				return fmt.Errorf("address change would make gateway %s unreachable; update the dependent route first", route.Gateway)
			}
		}
	}
	err = applySystemNetworkChange(plan, delta)
	return hostNetworkSystemError(err)
}

func (s *HostNetworkService) UpdateRoute(change dto.HostNetworkRouteChange, remove bool) error {
	hostNetworkMu.Lock()
	defer hostNetworkMu.Unlock()
	if _, err := hostNetworkPersistenceSupport(); err != nil {
		return err
	}
	req, err := normalizeStaticHostNetworkRoute(change.HostNetworkRoute)
	if err != nil {
		return err
	}
	if change.Confirmation != req.Interface {
		return errors.New("network change confirmation does not match the interface")
	}
	current, err := loadHostNetworkRoutes(context.Background(), "")
	if err != nil {
		return err
	}
	previous := req
	if change.Previous != nil {
		if remove || change.Previous.Interface != req.Interface {
			return errors.New("route edits must keep the same interface")
		}
		previous, err = normalizeHostNetworkRoute(*change.Previous)
		if err != nil {
			return err
		}
	}
	delta := hostnetwork.PersistentChange{}
	if remove || change.Previous != nil {
		index := slices.IndexFunc(current, func(route dto.HostNetworkRoute) bool { return sameHostNetworkRoute(previous, route) })
		if index < 0 || current[index].ReadOnly || current[index].Source != "" {
			return errors.New("route is missing or read-only")
		}
		previous = current[index]
		if previous.Protocol != "static" && previous.Protocol != "boot" && previous.Protocol != "" {
			return errors.New("only static routes can be changed")
		}
		delta.RemoveRoutes = append(delta.RemoveRoutes, systemNetworkRoute(previous))
	}
	if !remove {
		route := systemNetworkRoute(req)
		route.Protocol = "static"
		delta.AddRoutes = append(delta.AddRoutes, route)
	}
	plan, err := hostnetwork.PreparePersistentNetwork(req.Interface, delta)
	if err != nil {
		return hostNetworkSystemError(err)
	}
	err = applySystemNetworkChange(plan, delta)
	return hostNetworkSystemError(err)
}

func (s *HostNetworkService) CheckRoute(req dto.HostNetworkRoute) (*dto.HostNetworkRouteCheck, error) {
	routes, err := loadHostNetworkRoutes(context.Background(), "")
	if err != nil {
		return nil, err
	}
	return checkHostNetworkRoute(context.Background(), req, routes)
}

func (s *HostNetworkService) CheckRoutes(ctx context.Context, requests []dto.HostNetworkRoute) ([]dto.HostNetworkRouteCheck, error) {
	routes, err := loadHostNetworkRoutes(ctx, "")
	if err != nil {
		return nil, err
	}
	byDestination := make(map[string][]dto.HostNetworkRoute, len(routes))
	for _, route := range routes {
		dest := hostNetworkRouteDestination(route.Destination)
		byDestination[dest] = append(byDestination[dest], route)
	}
	results := make([]dto.HostNetworkRouteCheck, len(requests))
	var probes errgroup.Group
	probes.SetLimit(4)
	for i, req := range requests {
		if ctx.Err() != nil {
			break
		}
		probes.Go(func() error {
			if err := ctx.Err(); err != nil {
				return err
			}
			result, err := checkHostNetworkRoute(ctx, req, byDestination[hostNetworkRouteDestination(req.Destination)])
			if err != nil {
				results[i] = dto.HostNetworkRouteCheck{Failed: true, Message: err.Error()}
			} else {
				results[i] = *result
			}
			return ctx.Err()
		})
	}
	if err := probes.Wait(); err != nil {
		return nil, err
	}
	return results, ctx.Err()
}

func hostNetworkReason(err error) string {
	var business buserr.BusinessError
	if errors.As(err, &business) && (business.Msg == "ErrNetworkLoopbackReadOnly" || business.Msg == "ErrNetworkPersistentSourceMissing" || business.Msg == "ErrNetworkVirtualBridgeReadOnly") {
		return business.Msg
	}
	return err.Error()
}

func normalizeHostNetworkAddress(req dto.HostNetworkAddressRequest) (dto.HostNetworkAddressRequest, netip.Prefix, error) {
	prefix, err := netip.ParsePrefix(req.Address)
	if err != nil || prefix.Addr().IsUnspecified() || prefix.Addr().IsMulticast() || prefix.Addr().Is4In6() {
		return req, netip.Prefix{}, fmt.Errorf("invalid interface address: %s", req.Address)
	}
	req.Address = prefix.String()
	return req, prefix, nil
}

func findHostNetworkAddress(links []hostnetwork.Link, req dto.HostNetworkAddressRequest) *hostnetwork.Address {
	for _, link := range links {
		if link.Name != req.Interface {
			continue
		}
		for _, address := range link.Addresses {
			if address.CIDR() == req.Address {
				return &address
			}
		}
	}
	return nil
}

func guardHostNetworkPrimaryAddress(links []hostnetwork.Link, req dto.HostNetworkAddressRequest, prefix netip.Prefix, address *hostnetwork.Address) error {
	if !prefix.Addr().Is4() || address.Secondary {
		return nil
	}
	for _, link := range links {
		if link.Name != req.Interface {
			continue
		}
		for _, other := range link.Addresses {
			if other.Family != "inet" || !other.Secondary || other.PrefixLen != address.PrefixLen || other.Scope != address.Scope {
				continue
			}
			otherIP, err := netip.ParseAddr(other.Local)
			if err == nil && netip.PrefixFrom(otherIP, other.PrefixLen).Masked() == prefix.Masked() {
				return buserr.New("ErrNetworkPrimaryAddress")
			}
		}
	}
	return nil
}

func loadHostNetworkRoutes(ctx context.Context, protocol string) ([]dto.HostNetworkRoute, error) {
	routes, err := hostnetwork.Routes(ctx, protocol)
	if err != nil {
		return nil, hostNetworkSystemError(err)
	}
	result := make([]dto.HostNetworkRoute, 0, len(routes))
	for _, route := range routes {
		item := dto.HostNetworkRoute{Source: route.PreferredSource, Destination: route.Destination, Gateway: route.Gateway, Interface: route.Interface, Metric: route.Metric, Table: route.Table, Protocol: route.Protocol, Scope: route.Scope, Family: route.Family, ReadOnly: route.ReadOnly, Applied: true, NextHops: make([]dto.HostNetworkRouteNextHop, 0, len(route.NextHops))}
		for _, hop := range route.NextHops {
			item.NextHops = append(item.NextHops, dto.HostNetworkRouteNextHop{Gateway: hop.Gateway, Interface: hop.Interface, Weight: hop.Weight})
		}
		result = append(result, item)
	}
	return result, nil
}

func hostNetworkRouteDestination(destination string) string {
	prefix, err := netip.ParsePrefix(destination)
	if err != nil {
		address, err := netip.ParseAddr(destination)
		if err != nil {
			return destination
		}
		prefix = netip.PrefixFrom(address, address.BitLen())
	}
	if prefix.Bits() == 0 {
		return "default"
	}
	return prefix.Masked().String()
}

func currentHostNetworkRoute(req dto.HostNetworkRoute, routes []dto.HostNetworkRoute) (dto.HostNetworkRoute, error) {
	if req.Table != "" && req.Table != "main" {
		return dto.HostNetworkRoute{}, errors.New("only main routing table is supported")
	}
	if req.ReadOnly || len(req.NextHops) > 0 {
		return dto.HostNetworkRoute{}, buserr.New("ErrNetworkRouteReadOnly")
	}
	for _, route := range routes {
		if !route.ReadOnly && route.Family == req.Family && hostNetworkRouteDestination(route.Destination) == hostNetworkRouteDestination(req.Destination) && route.Gateway == req.Gateway && route.Interface == req.Interface && route.Metric == req.Metric && (req.Protocol == "" || route.Protocol == req.Protocol) && (req.Scope == "" || route.Scope == req.Scope) {
			return route, nil
		}
	}
	for _, route := range routes {
		if route.ReadOnly && route.Family == req.Family && hostNetworkRouteDestination(route.Destination) == hostNetworkRouteDestination(req.Destination) && (req.Metric == 0 || route.Metric == req.Metric) {
			return dto.HostNetworkRoute{}, buserr.New("ErrNetworkRouteReadOnly")
		}
	}
	return dto.HostNetworkRoute{}, buserr.New("ErrRecordNotFound")
}

func normalizeHostNetworkRoute(req dto.HostNetworkRoute) (dto.HostNetworkRoute, error) {
	if req.ReadOnly || len(req.NextHops) > 0 {
		return req, buserr.New("ErrNetworkRouteReadOnly")
	}
	if req.Family != "IPv4" && req.Family != "IPv6" {
		return req, errors.New("invalid route address family")
	}
	if req.Table != "" && req.Table != "main" {
		return req, errors.New("only main routing table is supported")
	}
	is4 := req.Family == "IPv4"
	if req.Destination != "default" {
		prefix, err := netip.ParsePrefix(req.Destination)
		if err != nil {
			address, err := netip.ParseAddr(req.Destination)
			if err != nil || address.Zone() != "" {
				return req, errors.New("invalid route destination")
			}
			prefix = netip.PrefixFrom(address, address.BitLen())
		}
		if prefix.Addr().Is4() != is4 || prefix.Addr().Is4In6() || prefix.Addr().IsMulticast() {
			return req, errors.New("route destination address family mismatch")
		}
		req.Destination = "default"
		if prefix.Bits() != 0 {
			req.Destination = prefix.Masked().String()
		}
	}
	if req.Gateway != "" {
		gateway, err := netip.ParseAddr(req.Gateway)
		if err != nil || gateway.Zone() != "" || gateway.Is4() != is4 || gateway.Is4In6() || gateway.IsUnspecified() || gateway.IsMulticast() {
			return req, errors.New("invalid route gateway")
		}
		req.Gateway = gateway.String()
	}
	return req, nil
}

func normalizeStaticHostNetworkRoute(req dto.HostNetworkRoute) (dto.HostNetworkRoute, error) {
	req, err := normalizeHostNetworkRoute(req)
	if err != nil {
		return req, err
	}
	if req.Family == "IPv6" && req.Metric == 0 {
		req.Metric = 1024
	}
	return dto.HostNetworkRoute{
		Destination: req.Destination,
		Gateway:     req.Gateway,
		Interface:   req.Interface,
		Metric:      req.Metric,
		Table:       "main",
		Protocol:    "static",
		Family:      req.Family,
		NextHops:    []dto.HostNetworkRouteNextHop{},
	}, nil
}

func sameHostNetworkRoute(a, b dto.HostNetworkRoute) bool {
	return a.Family == b.Family && hostNetworkRouteDestination(a.Destination) == hostNetworkRouteDestination(b.Destination) && a.Gateway == b.Gateway && a.Interface == b.Interface && a.Metric == b.Metric
}

func checkHostNetworkRoute(ctx context.Context, req dto.HostNetworkRoute, routes []dto.HostNetworkRoute) (*dto.HostNetworkRouteCheck, error) {
	if hostnetwork.IsReadOnlyVirtualInterface(req.Interface) {
		return &dto.HostNetworkRouteCheck{Skipped: true, Message: "ErrNetworkVirtualBridgeReadOnly"}, nil
	}
	if iface, err := net.InterfaceByName(req.Interface); err == nil && iface.Flags&net.FlagLoopback != 0 {
		return &dto.HostNetworkRouteCheck{Skipped: true, Message: "ErrNetworkLoopbackReadOnly"}, nil
	}
	req, err := currentHostNetworkRoute(req, routes)
	if err != nil {
		return nil, err
	}
	if _, err = normalizeHostNetworkRoute(req); err != nil {
		return nil, err
	}
	result, err := hostnetwork.CheckRoute(ctx, hostnetwork.Route{Destination: req.Destination, Gateway: req.Gateway, Interface: req.Interface, Metric: req.Metric, Table: req.Table, Protocol: req.Protocol, Scope: req.Scope, Family: req.Family})
	if err != nil {
		return nil, hostNetworkSystemError(err)
	}
	return &dto.HostNetworkRouteCheck{Reachable: result.Reachable, Message: result.Message}, nil
}

func hostNetworkSystemError(err error) error {
	var unavailable *hostnetwork.IPRoute2Error
	if errors.As(err, &unavailable) {
		return buserr.WithDetail("ErrNetworkIPRoute2", unavailable.Err.Error(), err)
	}
	return err
}

func hostNetworkInterfaces(links []hostnetwork.Link, plans map[string]*hostnetwork.PersistentNetwork, reasons map[string]string, reader *hostnetwork.PersistentNetworkReader) []dto.HostNetworkInterface {
	result := make([]dto.HostNetworkInterface, 0, len(links))
	for _, link := range links {
		plan, planErr := reader.Prepare(link.Name, hostnetwork.PersistentChange{})
		plans[link.Name] = plan
		if planErr != nil {
			reasons[link.Name] = hostNetworkReason(planErr)
		}
		item := dto.HostNetworkInterface{Index: link.Index, Name: link.Name, MTU: link.MTU, MAC: link.MAC, Flags: link.Flags, State: "DOWN", LinkState: link.OperState, Addresses: make([]dto.HostNetworkAddress, 0, len(link.Addresses))}
		if planErr != nil {
			item.Reason = reasons[link.Name]
		} else {
			item.Backend = plan.Backend
		}
		if slices.Contains(link.Flags, "UP") {
			item.State = "UP"
		}
		if slices.Contains(link.Flags, "LOWER_UP") {
			item.LinkState = "UP"
		}
		for _, address := range link.Addresses {
			family := "IPv4"
			if address.Family == "inet6" {
				family = "IPv6"
			}
			cidr := address.CIDR()
			entry := dto.HostNetworkAddress{Address: cidr, Family: family, Scope: address.Scope, Applied: !address.Tentative && !address.DadFailed, Reason: item.Reason}
			if planErr == nil {
				entry.Editable = !address.Dynamic && address.Scope == "global" && slices.Contains(plan.Addresses, cidr)
				entry.Managed = slices.Contains(plan.Addresses, cidr)
				if !entry.Editable {
					entry.Reason = "NetworkAddressReadOnlyReason"
				}
			}
			item.Addresses = append(item.Addresses, entry)
		}
		result = append(result, item)
	}
	return result
}

func hostNetworkRoutes(routes []dto.HostNetworkRoute, plans map[string]*hostnetwork.PersistentNetwork, reasons map[string]string, reader *hostnetwork.PersistentNetworkReader) {
	for i := range routes {
		route := &routes[i]
		if _, exists := plans[route.Interface]; !exists {
			plan, err := reader.Prepare(route.Interface, hostnetwork.PersistentChange{})
			plans[route.Interface] = plan
			if err != nil {
				reasons[route.Interface] = hostNetworkReason(err)
			}
		}
		route.Reason = reasons[route.Interface]
		if route.Reason == "ErrNetworkVirtualBridgeReadOnly" || route.Reason == "ErrNetworkLoopbackReadOnly" {
			route.ReadOnly = true
		}
		if plan := plans[route.Interface]; plan != nil {
			route.Editable = !route.ReadOnly && route.Source == "" && (route.Protocol == "static" || route.Protocol == "boot" || route.Protocol == "") && slices.ContainsFunc(plan.Routes, func(r hostnetwork.Route) bool { return hostnetwork.PersistentRouteEqual(r, systemNetworkRoute(*route)) })
			route.Managed = route.Editable
			if !route.Editable {
				route.Reason = "NetworkRouteReadOnlyReason"
			}
		}
		if route.Protocol == "kernel" && route.Gateway == "" && !route.ReadOnly {
			route.Reason = "NetworkRouteConnectedReason"
		}
	}
}
