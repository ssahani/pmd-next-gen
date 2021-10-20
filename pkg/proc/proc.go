// SPDX-License-Identifier: Apache-2.0

package proc

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"

	"github.com/pmd/pkg/system"
	"github.com/pmd/pkg/web"
)

const (
	procMiscPath    = "/proc/misc"
	procNetArpPath  = "/proc/net/arp"
	procModulesPath = "/proc/modules"
)

type NetARP struct {
	IPAddress string `json:"IPAddress"`
	HWType    string `json:"HWType"`
	Flags     string `json:"Flags"`
	HWAddress string `json:"HWAddress"`
	Mask      string `json:"Mask"`
	Device    string `json:"Device"`
}

type Modules struct {
	Module     string `json:"Module"`
	MemorySize string `json:"MemorySize"`
	Instances  string `json:"Instances"`
	Dependent  string `json:"Dependent"`
	State      string `json:"State"`
}

func FetchVersion(w http.ResponseWriter) error {
	infoStat, err := host.Info()
	if err != nil {
		return err
	}

	return web.JSONResponse(infoStat, w)
}

func FetchPlatformInformation(w http.ResponseWriter) error {
	platform, family, version, err := host.PlatformInformation()
	if err != nil {
		return err
	}

	p := struct {
		Platform string
		Family   string
		Version  string
	}{
		platform,
		family,
		version,
	}

	return web.JSONResponse(p, w)
}

func FetchVirtualization(w http.ResponseWriter) error {
	system, role, err := host.Virtualization()
	if err != nil {
		return err
	}

	v := struct {
		System string
		Role   string
	}{
		system,
		role,
	}

	return web.JSONResponse(v, w)
}

func FetchUserStat(w http.ResponseWriter) error {
	userStat, err := host.Users()
	if err != nil {
		return err
	}

	return web.JSONResponse(userStat, w)
}

func FetchTemperatureStat(w http.ResponseWriter) error {
	tempStat, err := host.SensorsTemperatures()
	if err != nil {
		return err
	}

	return web.JSONResponse(tempStat, w)
}

// read netstat from proc tcp/udp/sctp
func FetchNetStat(w http.ResponseWriter, protocol string) error {
	conn, err := net.Connections(protocol)
	if err != nil {
		return err
	}

	return web.JSONResponse(conn, w)
}

func FetchNetStatPid(w http.ResponseWriter, protocol string, process string) error {
	pid, err := strconv.ParseInt(process, 10, 32)
	if err != nil || protocol == "" || pid == 0 {
		return errors.New("can't parse request")
	}

	conn, err := net.ConnectionsPid(protocol, int32(pid))
	if err != nil {
		return err
	}

	return web.JSONResponse(conn, w)
}

func FetchProtoCountersStat(w http.ResponseWriter) error {
	protocols := []string{"ip", "icmp", "icmpmsg", "tcp", "udp", "udplite"}

	proto, err := net.ProtoCounters(protocols)
	if err != nil {
		return err
	}

	return web.JSONResponse(proto, w)
}

func FetchNetDev(w http.ResponseWriter) error {
	netDev, err := net.IOCounters(true)
	if err != nil {
		return err
	}

	return web.JSONResponse(netDev, w)
}

func FetchInterfaceStat(w http.ResponseWriter) error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	return web.JSONResponse(interfaces, w)
}

func FetchSwapMemoryStat(w http.ResponseWriter) error {
	swap, err := mem.SwapMemory()
	if err != nil {
		return err
	}

	return web.JSONResponse(swap, w)
}

func FetchVirtualMemoryStat(w http.ResponseWriter) error {
	virt, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	return web.JSONResponse(virt, w)
}

func FetchCPUInfo(w http.ResponseWriter) error {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return err
	}

	return web.JSONResponse(cpuInfo, w)
}

func FetchCPUTimeStat(w http.ResponseWriter) error {
	cpuTime, err := cpu.Times(true)
	if err != nil {
		return err
	}

	return web.JSONResponse(cpuTime, w)
}

func FetchAvgStat(w http.ResponseWriter) error {
	avgStat, err := load.Avg()
	if err != nil {
		return err
	}

	return web.JSONResponse(avgStat, w)
}

func FetchPartitions(w http.ResponseWriter) error {
	part, err := disk.Partitions(true)
	if err != nil {
		return err
	}

	return web.JSONResponse(part, w)
}

func FetchIOCounters(w http.ResponseWriter) error {
	ioCounters, err := disk.IOCounters()
	if err != nil {
		return err
	}

	return web.JSONResponse(ioCounters, w)
}

func FetchDiskUsage(w http.ResponseWriter) error {
	u, err := disk.Usage("/")
	if err != nil {
		return err
	}

	return web.JSONResponse(u, w)
}

func FetchMisc(w http.ResponseWriter) error {
	lines, err := system.ReadFullFile(procMiscPath)
	if err != nil {
		log.Fatalf("Failed to read: %s", procMiscPath)
		return err
	}

	miscMap := make(map[int]string)
	for _, line := range lines {
		fields := strings.Fields(line)

		deviceNum, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		miscMap[deviceNum] = fields[1]
	}

	return web.JSONResponse(miscMap, w)
}

func FetchNetArp(w http.ResponseWriter) error {
	lines, err := system.ReadFullFile(procNetArpPath)
	if err != nil {
		log.Errorf("Failed to read '%s': %v", procNetArpPath, err)
		return err
	}

	netARP := make([]NetARP, len(lines)-1)
	for i, line := range lines {
		if i == 0 {
			continue
		}

		fields := strings.Fields(line)

		arp := NetARP{}
		arp.IPAddress = fields[0]
		arp.HWType = fields[1]
		arp.Flags = fields[2]
		arp.HWAddress = fields[3]
		arp.Mask = fields[4]
		arp.Device = fields[5]
		netARP[i-1] = arp
	}

	return web.JSONResponse(netARP, w)
}

func FetchModules(w http.ResponseWriter) error {
	lines, err := system.ReadFullFile(procModulesPath)
	if err != nil {
		log.Fatalf("Failed to read '%s': %v", procModulesPath, err)
		return err
	}

	modules := make([]Modules, len(lines))
	for i, line := range lines {
		fields := strings.Fields(line)

		module := Modules{}

		for j, field := range fields {
			switch j {
			case 0:
				module.Module = field

			case 1:
				module.MemorySize = field

			case 2:
				module.Instances = field

			case 3:
				module.Dependent = field

			case 4:
				module.State = field
			}
		}

		modules[i] = module
	}

	return web.JSONResponse(modules, w)
}

func FetchProcessInfo(w http.ResponseWriter, proc string, property string) error {
	pid, err := strconv.ParseInt(proc, 10, 32)
	if err != nil {
		return err
	}

	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return err
	}

	switch property {
	case "pid-connections":
		conn, err := p.Connections()
		if err != nil {
			return err
		}

		return web.JSONResponse(conn, w)

	case "pid-rlimit":
		rlimit, err := p.Rlimit()
		if err != nil {
			return err
		}

		return web.JSONResponse(rlimit, w)

	case "pid-rlimit-usage":
		rlimit, err := p.RlimitUsage(true)
		if err != nil {
			return err
		}

		return web.JSONResponse(rlimit, w)

	case "pid-status":
		s, err := p.Status()
		if err != nil {
			return err
		}

		return web.JSONResponse(s, w)

	case "pid-username":
		u, err := p.Username()
		if err != nil {
			return err
		}

		return web.JSONResponse(u, w)

	case "pid-open-files":
		f, err := p.OpenFiles()
		if err != nil {
			return err
		}

		return web.JSONResponse(f, w)

	case "pid-fds":
		f, err := p.NumFDs()
		if err != nil {
			return err
		}

		return web.JSONResponse(f, w)

	case "pid-name":
		n, err := p.Name()
		if err != nil {
			return err
		}

		return web.JSONResponse(n, w)

	case "pid-memory-percent":
		m, err := p.MemoryPercent()
		if err != nil {
			return err
		}

		return web.JSONResponse(m, w)

	case "pid-memory-maps":
		m, err := p.MemoryMaps(true)
		if err != nil {
			return err
		}

		return web.JSONResponse(m, w)

	case "pid-memory-info":
		m, err := p.MemoryInfo()
		if err != nil {
			return err
		}

		return web.JSONResponse(m, w)

	case "pid-io-counters":
		m, err := p.IOCounters()
		if err != nil {
			return err
		}

		return web.JSONResponse(m, w)
	}

	return nil
}
