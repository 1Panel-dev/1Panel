package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/vm/types"
	"gorm.io/gorm"
)

func (v *VMService) ListNetworks() ([]dto.VMNetworkItem, error) {
	records, err := vmRepo.ListNetworks()
	if err != nil {
		return nil, err
	}
	items := v.networkRecordsToItems(records)
	return items, nil
}

func (v *VMService) ListNetworkBridges() ([]dto.VMNetworkBridgeItem, error) {
	return listNetworkBridges("/sys/class/net")
}

func (v *VMService) ListNetworkInterfaces() ([]dto.VMNetworkInterfaceItem, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	return listMacvtapInterfaces("/sys/class/net", interfaces)
}

func (v *VMService) CreateNetwork(req dto.VMNetworkCreate) error {
	if req.Type == "macvtap" {
		interfaces, err := v.ListNetworkInterfaces()
		if err != nil {
			return err
		}
		found := false
		for _, item := range interfaces {
			if item.Name == req.SourceInterface {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unavailable macvtap source interface: %s", req.SourceInterface)
		}
	} else {
		req.SourceInterface = ""
	}
	if req.Type != "nat" {
		req.CIDR, req.Gateway, req.DHCPStart, req.DHCPEnd = "", "", "", ""
	}
	if req.Type == "macvtap" {
		req.BridgeName = ""
	}
	ctx, cancel := vmResourceOperationContext(false)
	defer cancel()
	unlock, err := vmResourceLocks.lock(ctx, "network", req.Name)
	if err != nil {
		return err
	}
	defer unlock()
	itemNetwork, err := vmRepo.GetNetwork(repo.WithByName(req.Name))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if itemNetwork.ID != 0 {
		return buserr.New("ErrRecordExist")
	}
	if req.Type == "bridge" && isManagedNetworkBridge(req.BridgeName) {
		return buserr.New("ErrVMBridgeManaged")
	}
	if !v.helperExists() {
		return buserr.New("ErrVMHelperNotFound")
	}
	if _, exists, err := v.findVMResourceWithContext(ctx, "network", req.Name); err != nil {
		return err
	} else if exists {
		return buserr.New("ErrRecordExist")
	}
	record := model.VMNetwork{
		Name: req.Name, Type: req.Type, BridgeName: req.BridgeName, SourceInterface: req.SourceInterface,
		CIDR: req.CIDR, Gateway: req.Gateway, DHCPStart: req.DHCPStart, DHCPEnd: req.DHCPEnd,
		AutoStart: req.AutoStart, Status: vmResourceCreating, Description: req.Description,
	}
	if err := vmRepo.CreateNetwork(&record); err != nil {
		return err
	}
	if err := v.provider.CreateResource(ctx, types.VMResourceCreate{
		Type: "network", Name: req.Name, SubType: req.Type, BridgeName: req.BridgeName, SourceInterface: req.SourceInterface,
		CIDR: req.CIDR, Gateway: req.Gateway, DHCPStart: req.DHCPStart, DHCPEnd: req.DHCPEnd,
		AutoStart: req.AutoStart,
	}); err != nil {
		if updateErr := vmRepo.UpdateNetwork(record.ID, map[string]interface{}{
			"status": constant.StatusFailed, "error": err.Error(),
		}); updateErr != nil {
			return errors.Join(err, fmt.Errorf("save network creation failure: %w", updateErr))
		}
		return err
	}
	return vmRepo.UpdateNetwork(record.ID, map[string]interface{}{
		"status": constant.StatusRunning, "error": "",
	})
}

func (v *VMService) SyncNetworks() error {
	revision, idle := vmResourceLocks.snapshot("network")
	if !idle {
		return nil
	}
	if !v.helperExists() {
		return buserr.New("ErrVMHelperNotFound")
	}
	liveItems, err := v.provider.ListResources(context.Background(), "network")
	if err != nil {
		return err
	}
	return vmResourceLocks.applySnapshot("network", revision, func() error {
		records, err := vmRepo.ListNetworks()
		if err != nil {
			return err
		}
		liveByName := map[string]dto.VMResourceInfo{}
		for _, live := range liveItems {
			if live.Name == "" {
				continue
			}
			liveByName[live.Name] = live
		}
		recordsByName := map[string]model.VMNetwork{}
		for _, record := range records {
			if vmResourceStatusEquals(record.Status, vmResourceCreating) {
				record.Status = constant.StatusFailed
				record.Error = "Network creation was interrupted before its result was saved"
				if err := vmRepo.UpdateNetwork(record.ID, map[string]interface{}{
					"status": record.Status, "error": record.Error,
				}); err != nil {
					return err
				}
			}
			recordsByName[record.Name] = record
		}
		for name, live := range liveByName {
			vars := map[string]interface{}{
				"type":             live.SubType,
				"bridge_name":      live.BridgeName,
				"source_interface": live.SourceInterface,
				"cidr":             live.CIDR,
				"gateway":          live.Gateway,
				"dhcp_start":       live.DHCPStart,
				"dhcp_end":         live.DHCPEnd,
				"auto_start":       live.AutoStart,
				"status":           normalizeVMResourceStatus(live.Status),
			}
			if record, ok := recordsByName[name]; ok {
				if vmResourceStatusEquals(record.Status, constant.StatusFailed) {
					continue
				}
				if err := vmRepo.UpdateNetwork(record.ID, vars); err != nil {
					return err
				}
				continue
			}
			if err := vmRepo.CreateNetwork(&model.VMNetwork{
				Name: live.Name, Type: live.SubType, BridgeName: live.BridgeName, SourceInterface: live.SourceInterface,
				CIDR: live.CIDR, Gateway: live.Gateway, DHCPStart: live.DHCPStart, DHCPEnd: live.DHCPEnd,
				AutoStart: live.AutoStart, Status: normalizeVMResourceStatus(live.Status),
			}); err != nil {
				return err
			}
		}
		for _, record := range recordsByName {
			if vmResourceStatusEquals(record.Status, constant.StatusFailed) {
				continue
			}
			if _, ok := liveByName[record.Name]; !ok {
				if err := vmRepo.UpdateNetwork(record.ID, map[string]interface{}{
					"status": constant.StatusDeleted,
				}); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (v *VMService) UpdateNetwork(req dto.VMNetworkUpdate) error {
	ctx, cancel := vmResourceOperationContext(false)
	defer cancel()
	itemNetwork, err := vmRepo.GetNetwork(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	unlock, err := vmResourceLocks.lock(ctx, "network", itemNetwork.Name)
	if err != nil {
		return err
	}
	defer unlock()
	itemNetwork, err = vmRepo.GetNetwork(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if itemNetwork.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	vars := map[string]interface{}{}
	if req.Description != nil {
		vars["description"] = *req.Description
	}
	if req.AutoStart != nil || req.Status != nil {
		if vmResourceCreationIncomplete(itemNetwork.Status) {
			return buserr.WithDetail("ErrInvalidParams", "network creation has not succeeded", nil)
		}
		if !v.helperExists() {
			return buserr.New("ErrVMHelperNotFound")
		}
		live, ok, err := v.findVMResourceWithContext(ctx, "network", itemNetwork.Name)
		if err != nil {
			return err
		}
		if !ok {
			_ = vmRepo.UpdateNetwork(req.ID, map[string]interface{}{"status": constant.StatusDeleted})
			return buserr.WithDetail("ErrResourceNotExists", "Network does not exist", nil)
		}
		if live.Status != "" {
			vars["status"] = normalizeVMResourceStatus(live.Status)
		}
	}
	if req.AutoStart != nil {
		if err := v.provider.UpdateResource(ctx, types.VMResourceUpdate{
			Type: "network", Name: itemNetwork.Name, AutoStart: req.AutoStart,
		}); err != nil {
			return err
		}
		vars["auto_start"] = *req.AutoStart
	}
	if req.Status != nil {
		status := normalizeVMResourceStatus(*req.Status)
		if !validVMResourceOperateStatus(status) {
			return buserr.WithDetail("ErrInvalidParams", status, nil)
		}
		if err := v.provider.UpdateResource(ctx, types.VMResourceUpdate{
			Type: "network", Name: itemNetwork.Name, Status: &status,
		}); err != nil {
			return err
		}
		vars["status"] = status
	}
	if len(vars) == 0 {
		return nil
	}
	return vmRepo.UpdateNetwork(req.ID, vars)
}

func (v *VMService) DeleteNetwork(req dto.VMResourceDelete) error {
	ctx, cancel := vmResourceOperationContext(req.Force)
	defer cancel()
	itemNetwork, err := vmRepo.GetNetwork(repo.WithByID(req.ID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return buserr.New("ErrRecordNotFound")
		}
		return err
	}
	unlock, cleanup, err := vmResourceLocks.lockForDelete(ctx, "network", itemNetwork.Name, req.Force)
	if err != nil {
		return err
	}
	defer unlock()
	itemNetwork, err = vmRepo.GetNetwork(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if len(loadVMResourceRefs("network")[itemNetwork.Name]) > 0 {
		return buserr.New("ErrVMNetworkInUse")
	}
	if cleanup {
		if err := v.deleteVMResource(ctx, "network", itemNetwork.Name, req.Force); err != nil {
			return err
		}
	}
	return vmRepo.DeleteNetwork(repo.WithByID(req.ID))
}

func (v *VMService) networkRecordsToItems(records []model.VMNetwork) []dto.VMNetworkItem {
	items := make([]dto.VMNetworkItem, 0, len(records))
	vmRefs := loadVMResourceRefs("network")
	usedIPs := v.loadVMNetworkIPUsages(records)
	for _, record := range records {
		status := normalizeVMResourceStatus(record.Status)
		items = append(items, dto.VMNetworkItem{
			ID: record.ID, Name: record.Name, Type: record.Type,
			BridgeName: record.BridgeName, SourceInterface: record.SourceInterface, CIDR: record.CIDR, Gateway: record.Gateway,
			DHCPStart: record.DHCPStart, DHCPEnd: record.DHCPEnd, AutoStart: record.AutoStart,
			Description: record.Description, CreatedAt: record.CreatedAt.Format("2006-01-02 15:04:05"),
			Status: status, Error: record.Error, VMCount: len(vmRefs[record.Name]), VMs: vmRefs[record.Name],
			Registered: true, UsedIPs: usedIPs[record.Name],
		})
	}
	return items
}

func (v *VMService) loadVMNetworkIPUsages(records []model.VMNetwork) map[string][]dto.VMNetworkIPUsage {
	usages := map[string][]dto.VMNetworkIPUsage{}
	if !v.helperExists() {
		return usages
	}
	domains, err := v.provider.List(context.Background())
	if err != nil {
		return usages
	}
	vmRecords, err := vmRepo.List()
	if err != nil {
		return usages
	}
	vmRecordsByName := map[string]model.VirtualMachine{}
	for _, record := range vmRecords {
		vmRecordsByName[record.Name] = record
	}
	networkNames := map[string]struct{}{}
	bridgeToNetwork := map[string]string{}
	for _, record := range records {
		networkNames[record.Name] = struct{}{}
		if record.BridgeName != "" {
			bridgeToNetwork[record.BridgeName] = record.Name
		}
	}
	for _, domain := range domains {
		if record, ok := vmRecordsByName[domain.Name]; ok {
			mergeVMMetadata(&domain, record)
		}
		for _, item := range domain.Interfaces {
			ip := strings.TrimSpace(item.IP)
			if ip == "" {
				continue
			}
			networkName := resolveVMInterfaceNetwork(item.Source, domain.Network, networkNames, bridgeToNetwork)
			if networkName == "" {
				continue
			}
			usages[networkName] = append(usages[networkName], dto.VMNetworkIPUsage{
				IP: ip, VMName: domain.Name, MAC: item.MAC, Source: item.Source,
			})
		}
	}
	for name := range usages {
		sort.Slice(usages[name], func(i, j int) bool {
			if usages[name][i].IP != usages[name][j].IP {
				return usages[name][i].IP < usages[name][j].IP
			}
			return strings.ToLower(usages[name][i].VMName) < strings.ToLower(usages[name][j].VMName)
		})
	}
	return usages
}

func resolveVMInterfaceNetwork(source, vmNetwork string, networkNames map[string]struct{}, bridgeToNetwork map[string]string) string {
	if _, ok := networkNames[source]; ok {
		return source
	}
	if _, ok := networkNames[vmNetwork]; ok {
		return vmNetwork
	}
	return bridgeToNetwork[source]
}

func listNetworkBridges(netClassPath string) ([]dto.VMNetworkBridgeItem, error) {
	entries, err := os.ReadDir(netClassPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []dto.VMNetworkBridgeItem{}, nil
		}
		return nil, err
	}

	items := make([]dto.VMNetworkBridgeItem, 0)
	for _, entry := range entries {
		name := entry.Name()
		netPath := netClassPath + "/" + name
		info, err := os.Stat(netPath)
		if err != nil || !info.IsDir() {
			continue
		}
		if _, err := os.Stat(netPath + "/bridge"); err != nil {
			continue
		}
		item := dto.VMNetworkBridgeItem{
			Name:        name,
			Recommended: true,
		}
		switch {
		case name == "docker0" || strings.HasPrefix(name, "br-"):
			item.Recommended = false
			item.Description = "docker"
		case name == "virbr0" || strings.HasPrefix(name, "virbr"):
			item.Recommended = false
			item.Description = "libvirt"
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Recommended != items[j].Recommended {
			return items[i].Recommended
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	return items, nil
}

func isManagedNetworkBridge(name string) bool {
	return name == "docker0" || strings.HasPrefix(name, "br-") || name == "virbr0" || strings.HasPrefix(name, "virbr")
}

func listMacvtapInterfaces(netClassPath string, interfaces []net.Interface) ([]dto.VMNetworkInterfaceItem, error) {
	items := make([]dto.VMNetworkInterfaceItem, 0)
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || len(iface.HardwareAddr) != 6 {
			continue
		}
		netPath := filepath.Join(netClassPath, iface.Name)
		if _, err := os.Stat(filepath.Join(netPath, "device")); err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			continue
		}
		excluded := false
		for _, name := range []string{"master", "wireless", "bridge"} {
			if _, err := os.Stat(filepath.Join(netPath, name)); err == nil {
				excluded = true
				break
			} else if !os.IsNotExist(err) {
				return nil, err
			}
		}
		if excluded {
			continue
		}
		addresses, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		item := dto.VMNetworkInterfaceItem{Name: iface.Name, Addresses: []string{}}
		for _, address := range addresses {
			item.Addresses = append(item.Addresses, address.String())
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items, nil
}
