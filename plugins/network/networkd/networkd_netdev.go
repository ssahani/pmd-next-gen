// SPDX-License-Identifier: Apache-2.0
// Copyright 2022 VMware, Inc.

package networkd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/pmd-nextgen/pkg/configfile"
	"github.com/pmd-nextgen/pkg/validator"
	"github.com/pmd-nextgen/pkg/web"
	log "github.com/sirupsen/logrus"
)

type VLan struct {
	Id string `json:"Id"`
}

type NetDev struct {
	Link string `json:"Link"`

	MatchSection MatchSection `json:"MatchSection"`

	// [NetDev]
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Kind        string `json:"Kind"`
	MTUBytes    string `json:"MTUBytes"`
	MACAddress  string `json:"MACAddress"`

	// [VLAN]
	VLanSection VLan `json:"VLanSection"`
}

func decodeNetDevJSONRequest(r *http.Request) (*NetDev, error) {
	n := NetDev{}
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		return &n, err
	}

	return &n, nil
}

func (n *NetDev) BuildNetDevSection(m *configfile.Meta) error {
	m.NewSection("NetDev")

	if validator.IsNotEmptyString(n.Description) {
		m.SetKeyToNewSectionString("Description", n.Description)
	}

	if validator.IsEmptyString(n.Name) {
		log.Errorf("Failed to create VLan. Missing NetDev name")
		return errors.New("missing netdev name")

	}
	m.SetKeyToNewSectionString("Name", n.Name)

	if validator.IsEmptyString(n.Kind) {
		log.Errorf("Failed to create VLan. Missing NetDev kind")
		return errors.New("missing netdev kind")
	}
	m.SetKeyToNewSectionString("Kind", n.Kind)

	if validator.IsNotEmptyString(n.MACAddress) {
		if validator.IsNotMAC(n.MACAddress) {
			log.Errorf("Failed to create VLan='%s'. Invalid MACAddress='%s': %v", n.Name, n.MTUBytes)
			return fmt.Errorf("invalid MACAddress='%s'", n.MACAddress)
		}

		m.SetKeyToNewSectionString("MACAddress", n.MACAddress)
	}

	if validator.IsNotEmptyString(n.MTUBytes) {
		if !validator.IsMtu(n.MTUBytes) {
			log.Errorf("Failed to create VLan='%s'. Invalid MTUBytes='%s': %v", n.Name, n.MTUBytes)
			return fmt.Errorf("invalid MTUBytes='%s'", n.MTUBytes)
		}

		m.SetKeyToNewSectionString("MTUBytes", n.MTUBytes)
	}

	return nil
}

func (n *NetDev) BuildKindSection(m *configfile.Meta) error {
	nm, err := CreateOrParseNetworkFile(n.Link)
	if err != nil {
		log.Errorf("Failed to parse network file for link='%s': %v", n.Link, err)
		return err
	}

	switch n.Kind {
	case "vlan":
		m.NewSection("VLAN")

		if validator.IsEmptyString(n.VLanSection.Id) {
			log.Errorf("Failed to create VLan='%s'. Missing Id,", n.Name, err)
			return errors.New("missing vlan id")
		}

		if !validator.IsVLanId(n.VLanSection.Id) {
			log.Errorf("Failed to create VLan='%s'. Invalid Id='%s'", n.Name, n.VLanSection.Id)
			return fmt.Errorf("invalid vlan id='%s'", n.VLanSection.Id)
		}
		m.SetKeyToNewSectionString("Id", n.VLanSection.Id)

		if err := nm.NewKeyToSectionString("Network", "VLAN", n.Name); err != nil {
			log.Errorf("Failed to update .network file of link='%s',", n.Link, err)
			return err
		}
	}

	if err := nm.Save(); err != nil {
		log.Errorf("Failed to update config file='%s': %v", m.Path, err)
		return err
	}

	return nil
}

func (n *NetDev) ConfigureNetDev(ctx context.Context, w http.ResponseWriter) error {
	m, err := CreateNetDevFile(n.Name, n.Kind)
	if err != nil {
		log.Errorf("Failed to parse netdev file for link='%s': %v", n.Name, err)
		return err
	}

	if err = n.BuildNetDevSection(m); err != nil {
		return err
	}
	if err := n.BuildKindSection(m); err != nil {
		return err
	}

	if err := m.Save(); err != nil {
		log.Errorf("Failed to update config file='%s': %v", m.Path, err)
		return err
	}

	// Create .network file for netdev
	if err := CreateNetDevNetworkFile(n.Name, n.Kind); err != nil {
		return err
	}

	c, err := NewSDConnection()
	if err != nil {
		log.Errorf("Failed to establish connection with the system bus: %v", err)
		return err
	}
	defer c.Close()

	if err := c.DBusNetworkReload(ctx); err != nil {
		return err
	}

	return web.JSONResponse("configured", w)
}
