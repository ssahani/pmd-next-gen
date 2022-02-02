// SPDX-License-Identifier: Apache-2.0
// Copyright 2022 VMware, Inc.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pmd-nextgen/pkg/share"
	"github.com/pmd-nextgen/pkg/web"
	"github.com/pmd-nextgen/plugins/management"
	"github.com/pmd-nextgen/plugins/management/hostname"
	"github.com/pmd-nextgen/plugins/management/timedate"
	"github.com/pmd-nextgen/plugins/network/netlink/address"
	"github.com/pmd-nextgen/plugins/network/netlink/route"
	"github.com/pmd-nextgen/plugins/network/networkd"
	"github.com/pmd-nextgen/plugins/systemd"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemDescribe struct {
	Success bool                `json:"success"`
	Message management.Describe `json:"message"`
	Errors  string              `json:"errors"`
}

func acquireSystemDescribe(host string, token map[string]string) (*management.Describe, error) {
	resp, err := web.DispatchSocket(http.MethodGet, host, "/api/v1/system/describe", token, nil)
	if err != nil {
		fmt.Printf("Failed to acquire system info: %v\n", err)
		return nil, err
	}

	h := SystemDescribe{}
	if err := json.Unmarshal(resp, &h); err != nil {
		fmt.Printf("Failed to decode json message: %v\n", err)
		return nil, err
	}

	if !h.Success {
		fmt.Printf("%v\n", h.Errors)
		return nil, errors.New(h.Errors)
	}

	return &h.Message, nil
}

func displayHostname(h *hostname.Describe) {
	fmt.Printf("              %v %v\n", color.HiBlueString("System Name:"), h.StaticHostname)
	fmt.Printf("                   %v %v (%v) %v\n", color.HiBlueString("Kernel:"), h.KernelName, h.KernelRelease, h.KernelVersion)
	fmt.Printf("                  %v %v\n", color.HiBlueString("Chassis:"), h.Chassis)
	if h.HardwareModel != "" {
		fmt.Printf("           %v %v\n", color.HiBlueString("Hardware Model:"), h.HardwareModel)
	}
	if h.HardwareVendor != "" {
		fmt.Printf("          %v %v\n", color.HiBlueString("Hardware Vendor:"), h.HardwareVendor)
	}
	if h.ProductUUID != "" {
		fmt.Printf("             %v %v\n", color.HiBlueString("Product UUID:"), h.ProductUUID)
	}
	fmt.Printf("         %v %v\n", color.HiBlueString("Operating System:"), h.OperatingSystemPrettyName)
	if h.OperatingSystemHomeURL != "" {
		fmt.Printf("%v %v\n", color.HiBlueString("Operating System Home URL:"), h.OperatingSystemHomeURL)
	}
}

func displayTimeDate(t *timedate.Describe) {
	tm := time.UnixMicro(int64(t.TimeUSec))
	location, _ := time.LoadLocation(t.Timezone)

	fmt.Printf("                %v %v (%v)\n", color.HiBlueString("Time zone:"), t.Timezone, tm.In(location))
	fmt.Printf("         %v %v\n", color.HiBlueString("NTP synchronized:"), t.NTPSynchronized)

	fmt.Printf("       %v %v\n", color.HiBlueString("              Time:"), tm.Format(time.UnixDate))
	tm = time.UnixMicro(int64(t.TimeUSec))
	fmt.Printf("       %v %v\n", color.HiBlueString("          RTC Time:"), tm.UTC())
}

func displaySystemd(sd *systemd.Describe) {
	fmt.Printf("          %v %v\n", color.HiBlueString("Systemd Version:"), sd.Version)
	fmt.Printf("             %v %v\n", color.HiBlueString("Architecture:"), sd.Architecture)
	fmt.Printf("           %v %v\n", color.HiBlueString("Virtualization:"), sd.Virtualization)
}

func displayNetworkState(n *networkd.NetworkDescribe) {
	fmt.Printf("            %-10v %v (%v)\n", color.HiBlueString("Network State:"), n.OperationalState, n.CarrierState)
	if n.OnlineState != "" {
		fmt.Printf("     %-10v %v\n", color.HiBlueString("Network Online State:"), n.OnlineState)
	}
	if len(n.DNS) > 0 {
		fmt.Printf("                      %-10v %v\n", color.HiBlueString("DNS:"), strings.Join(n.DNS, " "))
	}
	if len(n.Domains) > 0 {
		fmt.Printf("                  %-10v %v\n", color.HiBlueString("Domains:"), strings.Join(n.Domains, " "))
	}
	if len(n.NTP) > 0 {
		fmt.Printf("                      %-10v %v\n", color.HiBlueString("NTP:"), strings.Join(n.NTP, " "))
	}
}

func displayNetworkAddresses(addInfo []address.AddressInfo) {
	fmt.Printf("                  %v", color.HiBlueString("Address:"))

	b := true
	for _, addrs := range addInfo {
		if addrs.Name == "lo" {
			continue
		}
		for _, a := range addrs.Addresses {
			if b {
				fmt.Printf(" %v/%v %v %v\n", a.IP, a.Mask, color.HiGreenString("on link"), addrs.Name)
				b = false
			} else {
				fmt.Printf("                           %v/%v %v %v\n", a.IP, a.Mask, color.HiGreenString("on link"), addrs.Name)
			}
		}
	}
}

func displayRoutes(linkRoutes []route.RouteInfo) {
	fmt.Printf("                  %v", color.HiBlueString("Gateway:"))

	b := true
	gws := share.NewSet()
	for _, rt := range linkRoutes {
		if rt.Gw != "" {
			if b {
				fmt.Printf(" %v %v %v\n", rt.Gw, color.HiGreenString("on link"), rt.LinkName)
				gws.Add(rt.LinkName)
				b = false
			} else {
				if !gws.Contains(rt.LinkName) {
					fmt.Printf("                           %v %v %v\n", rt.Gw, color.HiGreenString("on link"), rt.LinkName)
					gws.Add(rt.LinkName)
				}
			}
		}
	}
}

func displayHostInfo(h *host.InfoStat, u []host.UserStat) {
	t := time.Unix(int64(h.BootTime), 0)
	d, _ := share.SecondsToDuration(h.Uptime)
	fmt.Printf("                   %v %v (%v) %v (%v) %v (%v) %v (%v)\n", color.HiBlueString("Uptime:"), color.HiYellowString("Running Since"), d,
		color.HiYellowString("Booted"), t.Format(time.UnixDate), color.HiYellowString("Users"), len(u), color.HiYellowString("Proc"), h.Procs)
}

func displayVMStat(v *mem.VirtualMemoryStat) {
	fmt.Printf("                   %v %v (%v) %v (%v) %v (%v) %v (%v)\n", color.HiBlueString("Memory:"), color.HiYellowString("Total"), v.Total,
		color.HiYellowString("Used"), v.Total, color.HiYellowString("Free"), v.Free, color.HiYellowString("Available"), v.Available)
}

func acquireSystemStatus(host string, token map[string]string) {
	s, err := acquireSystemDescribe(host, token)
	if err != nil {
		return
	}

	displayHostname(s.Hostname)
	displayTimeDate(s.TimeDate)
	displaySystemd(s.Systemd)
	displayNetworkState(s.NetworkDescribe)
	displayNetworkAddresses(s.Addresses)
	displayRoutes(s.Routes)
	displayHostInfo(s.HostInfo, s.UserStat)
	displayVMStat(s.VirtualMemoryStat)
}

func SetHostname(hostName string, host string, token map[string]string) {
	h := hostname.Hostname{
		Method: "SetStaticHostname",
		Value:  hostName,
	}

	resp, err := web.DispatchSocket(http.MethodPost, host, "/api/v1/system/hostname/method", token, h)
	if err != nil {
		fmt.Printf("Failed to set hostname: %v\n", err)
		return
	}

	m := web.JSONResponseMessage{}
	if err := json.Unmarshal(resp, &m); err != nil {
		fmt.Printf("Failed to decode json message: %v\n", err)
		return
	}

	if !m.Success {
		fmt.Printf("Failed to set hostname: %v\n", m.Errors)
	}

	fmt.Println(m.Message)
}
