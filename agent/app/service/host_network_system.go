package service

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"slices"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/hostnetwork"
)

func systemNetworkRoute(route dto.HostNetworkRoute) hostnetwork.Route {
	return hostnetwork.Route{Destination: route.Destination, Gateway: route.Gateway, Interface: route.Interface, Metric: route.Metric, Table: route.Table, Protocol: route.Protocol, Scope: route.Scope, Family: route.Family}
}

func applySystemNetworkChange(plan *hostnetwork.PersistentNetwork, change hostnetwork.PersistentChange) (resultErr error) {
	restore, err := hostnetwork.PreserveSecondaryAddresses(plan.Device, change)
	if err != nil {
		return err
	}
	if restore != nil {
		defer func() { resultErr = errors.Join(resultErr, restore()) }()
	}
	var previousRoutes []hostnetwork.Route
	if len(change.RemoveAddresses) > 0 {
		routes, err := hostnetwork.Routes(context.Background(), "")
		if err != nil {
			return err
		}
		for _, route := range routes {
			if route.Interface != plan.Device || route.Protocol != "static" && route.Protocol != "boot" && route.Protocol != "" {
				continue
			}
			for _, address := range change.RemoveAddresses {
				prefix, _ := netip.ParsePrefix(address)
				if route.PreferredSource == prefix.Addr().String() {
					return errors.New("a static route explicitly uses this source address; update that route first")
				}
			}
			if !route.ReadOnly && route.PreferredSource == "" {
				previousRoutes = append(previousRoutes, route)
			}
		}
	}
	if err := plan.Apply(); err != nil {
		return err
	}
	var rollbacks []func() error
	undo := func() error {
		failures := []error{plan.Rollback()}
		for i := len(rollbacks) - 1; i >= 0; i-- {
			failures = append(failures, rollbacks[i]())
		}
		if len(previousRoutes) > 0 {
			routes, err := hostnetwork.Routes(context.Background(), "")
			if err != nil {
				failures = append(failures, err)
			} else {
				for _, route := range previousRoutes {
					if !slices.ContainsFunc(routes, func(r hostnetwork.Route) bool {
						return r.Interface == plan.Device && hostnetwork.PersistentRouteEqual(route, r)
					}) {
						failures = append(failures, hostnetwork.ChangeRoute(route, "add"))
					}
				}
			}
		}
		return errors.Join(failures...)
	}
	failed := func(err error) error { return errors.Join(err, undo()) }
	for _, address := range change.AddAddresses {
		req := hostnetwork.AddressRequest{Interface: plan.Device, Address: address}
		if err := hostnetwork.ChangeSystemAddress(req, "add"); err != nil {
			return failed(err)
		}
		rollbacks = append(rollbacks, func() error { return hostnetwork.ChangeSystemAddress(req, "del") })
		if err := hostnetwork.WaitAddress(req); err != nil {
			return failed(err)
		}
	}
	for _, address := range change.RemoveAddresses {
		req := hostnetwork.AddressRequest{Interface: plan.Device, Address: address}
		if err := hostnetwork.ChangeSystemAddress(req, "del"); err != nil {
			return failed(err)
		}
		rollbacks = append(rollbacks, func() error { return hostnetwork.ChangeSystemAddress(req, "add") })
	}
	for _, route := range change.RemoveRoutes {
		if err := hostnetwork.ChangeRoute(route, "del"); err != nil {
			return failed(err)
		}
		rollbacks = append(rollbacks, func() error { return hostnetwork.ChangeRoute(route, "add") })
	}
	for _, route := range change.AddRoutes {
		if err := hostnetwork.ChangeRoute(route, "add"); err != nil {
			return failed(err)
		}
		rollbacks = append(rollbacks, func() error { return hostnetwork.ChangeRoute(route, "del") })
	}

	links, err := hostnetwork.Links(plan.Device)
	if err != nil {
		return failed(err)
	}
	for _, address := range change.AddAddresses {
		found := slices.ContainsFunc(links, func(link hostnetwork.Link) bool {
			return link.Name == plan.Device && slices.ContainsFunc(link.Addresses, func(a hostnetwork.Address) bool { return a.CIDR() == address && !a.DadFailed && !a.Tentative })
		})
		if !found {
			return failed(fmt.Errorf("address %s did not become active", address))
		}
	}
	for _, address := range change.RemoveAddresses {
		found := slices.ContainsFunc(links, func(link hostnetwork.Link) bool {
			return link.Name == plan.Device && slices.ContainsFunc(link.Addresses, func(a hostnetwork.Address) bool { return a.CIDR() == address })
		})
		if found {
			return failed(fmt.Errorf("address %s was not removed", address))
		}
	}
	routes, err := hostnetwork.Routes(context.Background(), "")
	if err != nil {
		return failed(err)
	}
	restored := false
	for _, route := range previousRoutes {
		if slices.ContainsFunc(change.RemoveRoutes, func(r hostnetwork.Route) bool { return hostnetwork.PersistentRouteEqual(route, r) }) {
			continue
		}
		if !slices.ContainsFunc(routes, func(r hostnetwork.Route) bool {
			return r.Interface == plan.Device && hostnetwork.PersistentRouteEqual(route, r)
		}) {
			if err := hostnetwork.ChangeRoute(route, "add"); err != nil {
				return failed(err)
			}
			restored = true
		}
	}
	if restored {
		routes, err = hostnetwork.Routes(context.Background(), "")
		if err != nil {
			return failed(err)
		}
	}
	for _, route := range change.AddRoutes {
		if !slices.ContainsFunc(routes, func(r hostnetwork.Route) bool {
			return r.Interface == plan.Device && hostnetwork.PersistentRouteEqual(route, r)
		}) {
			return failed(errors.New("route did not become active"))
		}
	}
	for _, route := range change.RemoveRoutes {
		if slices.ContainsFunc(change.AddRoutes, func(r hostnetwork.Route) bool { return hostnetwork.PersistentRouteEqual(route, r) }) {
			continue
		}
		if slices.ContainsFunc(routes, func(r hostnetwork.Route) bool {
			return r.Interface == plan.Device && hostnetwork.PersistentRouteEqual(route, r)
		}) {
			return failed(errors.New("route was not removed"))
		}
	}
	return nil
}
