package schema_v12

type smi struct {
	AttachedGpus  string `xml:"attached_gpus"`
	CudaVersion   string `xml:"cuda_version"`
	DriverVersion string `xml:"driver_version"`
	Gpu           []struct {
		ID                       string   `xml:"id,attr"`
		AccountedProcesses       struct{} `xml:"accounted_processes"`
		AccountingMode           string   `xml:"accounting_mode"`
		AccountingModeBufferSize string   `xml:"accounting_mode_buffer_size"`
		AddressingMode           string   `xml:"addressing_mode"`
		ApplicationsClocks       struct {
			GraphicsClock string `xml:"graphics_clock"`
			MemClock      string `xml:"mem_clock"`
		} `xml:"applications_clocks"`
		Bar1MemoryUsage struct {
			Free  string `xml:"free"`
			Total string `xml:"total"`
			Used  string `xml:"used"`
		} `xml:"bar1_memory_usage"`
		BoardID                string `xml:"board_id"`
		BoardPartNumber        string `xml:"board_part_number"`
		CcProtectedMemoryUsage struct {
			Free  string `xml:"free"`
			Total string `xml:"total"`
			Used  string `xml:"used"`
		} `xml:"cc_protected_memory_usage"`
		ClockPolicy struct {
			AutoBoost        string `xml:"auto_boost"`
			AutoBoostDefault string `xml:"auto_boost_default"`
		} `xml:"clock_policy"`
		Clocks struct {
			GraphicsClock string `xml:"graphics_clock"`
			MemClock      string `xml:"mem_clock"`
			SmClock       string `xml:"sm_clock"`
			VideoClock    string `xml:"video_clock"`
		} `xml:"clocks"`
		ClocksEventReasons struct {
			ClocksEventReasonApplicationsClocksSetting string `xml:"clocks_event_reason_applications_clocks_setting"`
			ClocksEventReasonDisplayClocksSetting      string `xml:"clocks_event_reason_display_clocks_setting"`
			ClocksEventReasonGpuIdle                   string `xml:"clocks_event_reason_gpu_idle"`
			ClocksEventReasonHwPowerBrakeSlowdown      string `xml:"clocks_event_reason_hw_power_brake_slowdown"`
			ClocksEventReasonHwSlowdown                string `xml:"clocks_event_reason_hw_slowdown"`
			ClocksEventReasonHwThermalSlowdown         string `xml:"clocks_event_reason_hw_thermal_slowdown"`
			ClocksEventReasonSwPowerCap                string `xml:"clocks_event_reason_sw_power_cap"`
			ClocksEventReasonSwThermalSlowdown         string `xml:"clocks_event_reason_sw_thermal_slowdown"`
			ClocksEventReasonSyncBoost                 string `xml:"clocks_event_reason_sync_boost"`
		} `xml:"clocks_event_reasons"`
		ComputeMode               string `xml:"compute_mode"`
		DefaultApplicationsClocks struct {
			GraphicsClock string `xml:"graphics_clock"`
			MemClock      string `xml:"mem_clock"`
		} `xml:"default_applications_clocks"`
		DeferredClocks struct {
			MemClock string `xml:"mem_clock"`
		} `xml:"deferred_clocks"`
		DisplayActive string `xml:"display_active"`
		DisplayMode   string `xml:"display_mode"`
		DriverModel   struct {
			CurrentDm string `xml:"current_dm"`
			PendingDm string `xml:"pending_dm"`
		} `xml:"driver_model"`
		EccErrors struct {
			Aggregate struct {
				DramCorrectable   string `xml:"dram_correctable"`
				DramUncorrectable string `xml:"dram_uncorrectable"`
				SramCorrectable   string `xml:"sram_correctable"`
				SramUncorrectable string `xml:"sram_uncorrectable"`
			} `xml:"aggregate"`
			Volatile struct {
				DramCorrectable   string `xml:"dram_correctable"`
				DramUncorrectable string `xml:"dram_uncorrectable"`
				SramCorrectable   string `xml:"sram_correctable"`
				SramUncorrectable string `xml:"sram_uncorrectable"`
			} `xml:"volatile"`
		} `xml:"ecc_errors"`
		EccMode struct {
			CurrentEcc string `xml:"current_ecc"`
			PendingEcc string `xml:"pending_ecc"`
		} `xml:"ecc_mode"`
		EncoderStats struct {
			AverageFps     string `xml:"average_fps"`
			AverageLatency string `xml:"average_latency"`
			SessionCount   string `xml:"session_count"`
		} `xml:"encoder_stats"`
		Fabric struct {
			State  string `xml:"state"`
			Status string `xml:"status"`
		} `xml:"fabric"`
		FanSpeed      string `xml:"fan_speed"`
		FbMemoryUsage struct {
			Free     string `xml:"free"`
			Reserved string `xml:"reserved"`
			Total    string `xml:"total"`
			Used     string `xml:"used"`
		} `xml:"fb_memory_usage"`
		FbcStats struct {
			AverageFps     string `xml:"average_fps"`
			AverageLatency string `xml:"average_latency"`
			SessionCount   string `xml:"session_count"`
		} `xml:"fbc_stats"`
		GpuFruPartNumber string `xml:"gpu_fru_part_number"`
		GpuModuleID      string `xml:"gpu_module_id"`
		GpuOperationMode struct {
			CurrentGom string `xml:"current_gom"`
			PendingGom string `xml:"pending_gom"`
		} `xml:"gpu_operation_mode"`
		GpuPartNumber    string `xml:"gpu_part_number"`
		GpuPowerReadings struct {
			CurrentPowerLimit   string `xml:"current_power_limit"`
			DefaultPowerLimit   string `xml:"default_power_limit"`
			MaxPowerLimit       string `xml:"max_power_limit"`
			MinPowerLimit       string `xml:"min_power_limit"`
			PowerDraw           string `xml:"power_draw"`
			PowerState          string `xml:"power_state"`
			RequestedPowerLimit string `xml:"requested_power_limit"`
		} `xml:"gpu_power_readings"`
		GpuResetStatus struct {
			DrainAndResetRecommended string `xml:"drain_and_reset_recommended"`
			ResetRequired            string `xml:"reset_required"`
		} `xml:"gpu_reset_status"`
		GpuVirtualizationMode struct {
			HostVgpuMode       string `xml:"host_vgpu_mode"`
			VirtualizationMode string `xml:"virtualization_mode"`
		} `xml:"gpu_virtualization_mode"`
		GspFirmwareVersion string `xml:"gsp_firmware_version"`
		Ibmnpu             struct {
			RelaxedOrderingMode string `xml:"relaxed_ordering_mode"`
		} `xml:"ibmnpu"`
		InforomVersion struct {
			EccObject  string `xml:"ecc_object"`
			ImgVersion string `xml:"img_version"`
			OemObject  string `xml:"oem_object"`
			PwrObject  string `xml:"pwr_object"`
		} `xml:"inforom_version"`
		MaxClocks struct {
			GraphicsClock string `xml:"graphics_clock"`
			MemClock      string `xml:"mem_clock"`
			SmClock       string `xml:"sm_clock"`
			VideoClock    string `xml:"video_clock"`
		} `xml:"max_clocks"`
		MaxCustomerBoostClocks struct {
			GraphicsClock string `xml:"graphics_clock"`
		} `xml:"max_customer_boost_clocks"`
		MigDevices struct {
			MigDevice []struct {
				Index             string `xml:"index"`
				GpuInstanceID     string `xml:"gpu_instance_id"`
				ComputeInstanceID string `xml:"compute_instance_id"`
				EccErrorCount     struct {
					Text          string `xml:",chardata" json:"text"`
					VolatileCount struct {
						SramUncorrectable string `xml:"sram_uncorrectable"`
					} `xml:"volatile_count" json:"volatile_count"`
				} `xml:"ecc_error_count" json:"ecc_error_count"`
				FbMemoryUsage struct {
					Total    string `xml:"total"`
					Reserved string `xml:"reserved"`
					Used     string `xml:"used"`
					Free     string `xml:"free"`
				} `xml:"fb_memory_usage" json:"fb_memory_usage"`
				Bar1MemoryUsage struct {
					Total string `xml:"total"`
					Used  string `xml:"used"`
					Free  string `xml:"free"`
				} `xml:"bar1_memory_usage" json:"bar1_memory_usage"`
			} `xml:"mig_device" json:"mig_device"`
		} `xml:"mig_devices" json:"mig_devices"`
		MigMode struct {
			CurrentMig string `xml:"current_mig"`
			PendingMig string `xml:"pending_mig"`
		} `xml:"mig_mode"`
		MinorNumber         string `xml:"minor_number"`
		ModulePowerReadings struct {
			CurrentPowerLimit   string `xml:"current_power_limit"`
			DefaultPowerLimit   string `xml:"default_power_limit"`
			MaxPowerLimit       string `xml:"max_power_limit"`
			MinPowerLimit       string `xml:"min_power_limit"`
			PowerDraw           string `xml:"power_draw"`
			PowerState          string `xml:"power_state"`
			RequestedPowerLimit string `xml:"requested_power_limit"`
		} `xml:"module_power_readings"`
		MultigpuBoard string `xml:"multigpu_board"`
		Pci           struct {
			AtomicCapsInbound  string `xml:"atomic_caps_inbound"`
			AtomicCapsOutbound string `xml:"atomic_caps_outbound"`
			PciBridgeChip      struct {
				BridgeChipFw   string `xml:"bridge_chip_fw"`
				BridgeChipType string `xml:"bridge_chip_type"`
			} `xml:"pci_bridge_chip"`
			PciBus         string `xml:"pci_bus"`
			PciBusID       string `xml:"pci_bus_id"`
			PciDevice      string `xml:"pci_device"`
			PciDeviceID    string `xml:"pci_device_id"`
			PciDomain      string `xml:"pci_domain"`
			PciGpuLinkInfo struct {
				LinkWidths struct {
					CurrentLinkWidth string `xml:"current_link_width"`
					MaxLinkWidth     string `xml:"max_link_width"`
				} `xml:"link_widths"`
				PcieGen struct {
					CurrentLinkGen       string `xml:"current_link_gen"`
					DeviceCurrentLinkGen string `xml:"device_current_link_gen"`
					MaxDeviceLinkGen     string `xml:"max_device_link_gen"`
					MaxHostLinkGen       string `xml:"max_host_link_gen"`
					MaxLinkGen           string `xml:"max_link_gen"`
				} `xml:"pcie_gen"`
			} `xml:"pci_gpu_link_info"`
			PciSubSystemID        string `xml:"pci_sub_system_id"`
			ReplayCounter         string `xml:"replay_counter"`
			ReplayRolloverCounter string `xml:"replay_rollover_counter"`
			RxUtil                string `xml:"rx_util"`
			TxUtil                string `xml:"tx_util"`
		} `xml:"pci"`
		PerformanceState string `xml:"performance_state"`
		PersistenceMode  string `xml:"persistence_mode"`
		PowerReadings    struct {
			PowerState         string `xml:"power_state"`
			PowerManagement    string `xml:"power_management"`
			PowerDraw          string `xml:"power_draw"`
			PowerLimit         string `xml:"power_limit"`
			DefaultPowerLimit  string `xml:"default_power_limit"`
			EnforcedPowerLimit string `xml:"enforced_power_limit"`
			MinPowerLimit      string `xml:"min_power_limit"`
			MaxPowerLimit      string `xml:"max_power_limit"`
		} `xml:"power_readings"`
		Processes struct {
			ProcessInfo []struct {
				Pid         string `xml:"pid"`
				Type        string `xml:"type"`
				ProcessName string `xml:"process_name"`
				UsedMemory  string `xml:"used_memory"`
			} `xml:"process_info"`
		} `xml:"processes"`
		ProductArchitecture string `xml:"product_architecture"`
		ProductBrand        string `xml:"product_brand"`
		ProductName         string `xml:"product_name"`
		RemappedRows        struct {
			// Manually added
			Correctable   string `xml:"remapped_row_corr"`
			Uncorrectable string `xml:"remapped_row_unc"`
			Pending       string `xml:"remapped_row_pending"`
			Failure       string `xml:"remapped_row_failure"`
		} `xml:"remapped_rows"`
		RetiredPages struct {
			DoubleBitRetirement struct {
				RetiredCount    string `xml:"retired_count"`
				RetiredPagelist string `xml:"retired_pagelist"`
			} `xml:"double_bit_retirement"`
			MultipleSingleBitRetirement struct {
				RetiredCount    string `xml:"retired_count"`
				RetiredPagelist string `xml:"retired_pagelist"`
			} `xml:"multiple_single_bit_retirement"`
			PendingBlacklist  string `xml:"pending_blacklist"`
			PendingRetirement string `xml:"pending_retirement"`
		} `xml:"retired_pages"`
		Serial          string `xml:"serial"`
		SupportedClocks struct {
			SupportedMemClock []struct {
				SupportedGraphicsClock []string `xml:"supported_graphics_clock"`
				Value                  string   `xml:"value"`
			} `xml:"supported_mem_clock"`
		} `xml:"supported_clocks"`
		SupportedGpuTargetTemp struct {
			GpuTargetTempMax string `xml:"gpu_target_temp_max"`
			GpuTargetTempMin string `xml:"gpu_target_temp_min"`
		} `xml:"supported_gpu_target_temp"`
		Temperature struct {
			GpuTargetTemperature   string `xml:"gpu_target_temperature"`
			GpuTemp                string `xml:"gpu_temp"`
			GpuTempMaxGpuThreshold string `xml:"gpu_temp_max_gpu_threshold"`
			GpuTempMaxMemThreshold string `xml:"gpu_temp_max_mem_threshold"`
			GpuTempMaxThreshold    string `xml:"gpu_temp_max_threshold"`
			GpuTempSlowThreshold   string `xml:"gpu_temp_slow_threshold"`
			GpuTempTlimit          string `xml:"gpu_temp_tlimit"`
			MemoryTemp             string `xml:"memory_temp"`
		} `xml:"temperature"`
		Utilization struct {
			DecoderUtil string `xml:"decoder_util"`
			EncoderUtil string `xml:"encoder_util"`
			GpuUtil     string `xml:"gpu_util"`
			JpegUtil    string `xml:"jpeg_util"`
			MemoryUtil  string `xml:"memory_util"`
			OfaUtil     string `xml:"ofa_util"`
		} `xml:"utilization"`
		UUID         string `xml:"uuid"`
		VbiosVersion string `xml:"vbios_version"`
		Voltage      struct {
			GraphicsVolt string `xml:"graphics_volt"`
		} `xml:"voltage"`
	} `xml:"gpu"`
	Timestamp string `xml:"timestamp"`
}
