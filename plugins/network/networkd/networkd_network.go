// SPDX-License-Identifier: Apache-2.0

package networkd

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"

	"github.com/pm-web/pkg/configfile"
	"github.com/pm-web/pkg/web"
)

type MatchSection struct {
	Name string `json:"Name"`
}

type NetworkSection struct {
	DHCP                string   `json:"DHCP"`
	Address             string   `json:"Address"`
	Gateway             string   `json:"Gateway"`
	DNS                 []string `json:"DNS"`
	Domains             []string `json:"Domains"`
	NTP                 []string `json:"NTP"`
	IPv6AcceptRA        string   `json:"IPv6AcceptRA"`
	LinkLocalAddressing string   `json:"LinkLocalAddressing"`
	MulticastDNS        string   `json:"MulticastDNS"`
}
type AddressSection struct {
	Address string `json:"Address"`
	Peer    string `json:"Peer"`
	Label   string `json:"Label"`
	Scope   string `json:"Scope"`
}
type RouteSection struct {
	Gateway         string `json:"Gateway"`
	GatewayOnlink   string `json:"GatewayOnlink"`
	Destination     string `json:"Destination"`
	Source          string `json:"Source"`
	PreferredSource string `json:"PreferredSource"`
	Table           string `json:"Table"`
	Scope           string `json:"Scope"`
}

type DHCPv4Section struct {
	ClientIdentifier      string `json:"ClientIdentifier"`
	VendorClassIdentifier string `json:"VendorClassIdentifier"`
	RequestOptions        string `json:"RequestOptions"`
	SendOption            string `json:"SendOption"`
	UseDNS                string `json:"UseDNS"`
	UseNTP                string `json:"UseNTP"`
	UseHostname           string `json:"UseHostname"`
	UseDomains            string `json:"UseDomains"`
	UseRoutes             string `json:"UseRoutes"`
	UseMTU                string `json:"UseMTU"`
	UseGateway            string `json:"UseGateway"`
	UseTimezone           string `json:"UUseTimezone"`
}

type Network struct {
	Link            string           `json:"Link"`
	MatchSection    MatchSection     `json:"MatchSection"`
	NetworkSection  NetworkSection   `json:"NetworkSection"`
	DHCPv4Section   DHCPv4Section    `json:"DHCPv4Section"`
	AddressSections []AddressSection `json:"AddressSections"`
	RouteSections   []RouteSection   `json:"RouteSections"`
}

type LinkState struct {
	AddressState     string   `json:"AddressState"`
	AlternativeNames []string `json:"AlternativeNames"`
	CarrierState     string   `json:"CarrierState"`
	Driver           string   `json:"Driver"`
	IPv4AddressState string   `json:"IPv4AddressState"`
	IPv6AddressState string   `json:"IPv6AddressState"`
	Index            int      `json:"Index"`
	LinkFile         string   `json:"LinkFile"`
	Model            string   `json:"Model"`
	Name             string   `json:"Name"`
	NetworkFile      string   `json:"NetworkFile"`
	OnlineState      string   `json:"OnlineState"`
	OperationalState string   `json:"OperationalState"`
	Path             string   `json:"Path"`
	SetupState       string   `json:"SetupState"`
	Type             string   `json:"Type"`
	Vendor           string   `json:"Vendor"`
}

func decodeJSONRequest(r *http.Request) (*Network, error) {
	n := Network{}
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		return &n, err
	}

	return &n, nil
}

func AcquireNetworkLinkProperty(ctx context.Context, w http.ResponseWriter) error {
	c, err := NewSDConnection()
	if err != nil {
		log.Errorf("Failed to establish connection to the system bus: %s", err)
		return err
	}
	defer c.Close()

	links, err := c.DBusNetworkLinkProperty(ctx)
	if err != nil {
		return err
	}

	return web.JSONResponse(links, w)
}

func (n *Network) buildNetworkSection(m *configfile.Meta) {
	if n.NetworkSection.DHCP != "" {
		m.SetKeySectionString("Network", "DHCP", n.NetworkSection.DHCP)
	}
	if n.NetworkSection.Address != "" {
		m.SetKeySectionString("Network", "Address", n.NetworkSection.Address)
	}
	if n.NetworkSection.Gateway != "" {
		m.SetKeySectionString("Network", "Gateway", n.NetworkSection.Gateway)
	}
	if n.NetworkSection.IPv6AcceptRA != "" {
		m.SetKeySectionString("Network", "IPv6AcceptRA", n.NetworkSection.IPv6AcceptRA)
	}
	if n.NetworkSection.LinkLocalAddressing != "" {
		m.SetKeySectionString("Network", "LinkLocalAddressing", n.NetworkSection.LinkLocalAddressing)
	}
	if n.NetworkSection.MulticastDNS != "" {
		m.SetKeySectionString("Network", "MulticastDNS", n.NetworkSection.MulticastDNS)
	}
	if len(n.NetworkSection.Domains) > 0 {
		m.SetKeySectionString("Network", "Domains", strings.Join(n.NetworkSection.Domains, " "))
	}
	if len(n.NetworkSection.DNS) > 0 {
		m.SetKeySectionString("Network", "DNS", strings.Join(n.NetworkSection.DNS, " "))
	}
	if len(n.NetworkSection.NTP) > 0 {
		m.SetKeySectionString("Network", "NTP", strings.Join(n.NetworkSection.NTP, " "))
	}
}

func (n *Network) buildDHCPv4Section(m *configfile.Meta) {
	if n.DHCPv4Section.ClientIdentifier != "" {
		m.SetKeySectionString("DHCPv4", "ClientIdentifier", n.DHCPv4Section.ClientIdentifier)
	}
	if n.DHCPv4Section.VendorClassIdentifier != "" {
		m.SetKeySectionString("DHCPv4", "VendorClassIdentifier", n.DHCPv4Section.VendorClassIdentifier)
	}
	if n.DHCPv4Section.RequestOptions != "" {
		m.SetKeySectionString("DHCPv4", "RequestOptions", n.DHCPv4Section.RequestOptions)
	}
	if n.DHCPv4Section.SendOption != "" {
		m.SetKeySectionString("DHCPv4", "SendOption", n.DHCPv4Section.SendOption)
	}
	if n.DHCPv4Section.UseDNS != "" {
		m.SetKeySectionString("DHCPv4", "UseDNS", n.DHCPv4Section.UseDNS)
	}
	if n.DHCPv4Section.UseDomains != "" {
		m.SetKeySectionString("DHCPv4", "UseDomains", n.DHCPv4Section.UseDomains)
	}
	if n.DHCPv4Section.UseNTP != "" {
		m.SetKeySectionString("DHCPv4", "UseNTP", n.DHCPv4Section.UseNTP)
	}
	if n.DHCPv4Section.UseMTU != "" {
		m.SetKeySectionString("DHCPv4", "UseMTU", n.DHCPv4Section.UseMTU)
	}

	if n.DHCPv4Section.UseGateway != "" {
		m.SetKeySectionString("DHCPv4", "UseGateway", n.DHCPv4Section.UseGateway)
	}
	if n.DHCPv4Section.UseTimezone != "" {
		m.SetKeySectionString("DHCPv4", "UseTimezone", n.DHCPv4Section.UseTimezone)
	}
}

func (n *Network) buildAddressSection(m *configfile.Meta) {
	for _, a := range n.AddressSections {
		if a.Address != "" {
			m.SetKeySectionString("Address", "Address", a.Address)
		}
		if a.Peer != "" {
			m.SetKeySectionString("Address", "Peer", a.Peer)
		}
		if a.Label != "" {
			m.SetKeySectionString("Address", "Label", a.Label)
		}
		if a.Scope != "" {
			m.SetKeySectionString("Address", "Scope", a.Scope)
		}
	}
}

func (n *Network) buildRouteSection(m *configfile.Meta) {
	for _, rt := range n.RouteSections {
		if rt.Gateway != "" {
			m.SetKeySectionString("Route", "Gateway", rt.Gateway)
		}
		if rt.GatewayOnlink != "" {
			m.SetKeySectionString("Route", "GatewayOnlink", rt.GatewayOnlink)
		}
		if rt.Destination != "" {
			m.SetKeySectionString("Route", "Destination", rt.Destination)
		}
		if rt.Source != "" {
			m.SetKeySectionString("Route", "Source", rt.Source)
		}
		if rt.PreferredSource != "" {
			m.SetKeySectionString("Route", "PreferredSource", rt.PreferredSource)
		}
		if rt.Table != "" {
			m.SetKeySectionString("Route", "Table", rt.Table)
		}
		if rt.Scope != "" {
			m.SetKeySectionString("Route", "Scope", rt.Scope)
		}
	}
}

func (n *Network) ConfigureNetwork(ctx context.Context, w http.ResponseWriter) error {
	link, err := netlink.LinkByName(n.Link)
	if err != nil {
		return err
	}

	network, err := CreateOrParseNetworkFile(link)
	if err != nil {
		return err
	}

	m, err := configfile.Load(network)
	if err != nil {
		return err
	}

	n.buildNetworkSection(m)
	n.buildDHCPv4Section(m)
	n.buildAddressSection(m)
	n.buildRouteSection(m)

	if err := m.Save(); err != nil {
		return err
	}

	c, err := NewSDConnection()
	if err != nil {
		log.Errorf("Failed to establish connection to the system bus: %s", err)
		return err
	}
	defer c.Close()

	if err := c.DBusNetworkReload(ctx); err != nil {
		return err
	}

	return web.JSONResponse("configured", w)
}
