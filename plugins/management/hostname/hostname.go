// SPDX-License-Identifier: Apache-2.0

package hostname

import (
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/pm-web/pkg/web"
)

type Hostname struct {
	Method string `json:"Method"`
	Value  string `json:"Value"`
}

func (h *Hostname) SetHostname() error {
	conn, err := NewSDConnection()
	if err != nil {
		log.Errorf("Failed to establish connection to the system bus: %s", err)
		return err
	}
	defer conn.Close()

	return conn.SetHostName(h.Method, h.Value)
}

func AcquireHostnameProperties(w http.ResponseWriter) error {
	conn, err := NewSDConnection()
	if err != nil {
		log.Errorf("Failed to establish connection to the system bus: %s", err)
		return err
	}
	defer conn.Close()

	hostNameProperties := map[string]string{}

	var wg sync.WaitGroup
	wg.Add(16)

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("Hostname"); err == nil {
			hostNameProperties["Hostname"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("StaticHostname"); err == nil {
			hostNameProperties["StaticHostname"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("PrettyHostname"); err == nil {
			hostNameProperties["PrettyHostname"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("IconName"); err == nil {
			hostNameProperties["IconName"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("Chassis"); err == nil {
			hostNameProperties["Chassis"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("Deployment"); err == nil {
			hostNameProperties["Deployment"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("Location"); err == nil {
			hostNameProperties["Location"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("KernelName"); err == nil {
			hostNameProperties["KernelName"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("KernelRelease"); err == nil {
			hostNameProperties["LKernelRelease"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("KernelVersion"); err == nil {
			hostNameProperties["KernelVersion"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("OperatingSystemPrettyName"); err == nil {
			hostNameProperties["OperatingSystemPrettyName"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("OperatingSystemCPEName"); err == nil {
			hostNameProperties["OperatingSystemCPEName"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("OperatingSystemHomeURL"); err == nil {
			hostNameProperties["OperatingSystemHomeURL"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("HardwareVendor"); err == nil {
			hostNameProperties["HardwareVendor"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("HardwareModel"); err == nil {
			hostNameProperties["HardwareModel"] = p
		}
	}()

	go func() {
		defer wg.Done()

		if p, err := conn.GetHostName("ProductUUID"); err == nil {
			hostNameProperties["ProductUUID"] = p
		}
	}()

	wg.Wait()

	return web.JSONResponse(hostNameProperties, w)
}
