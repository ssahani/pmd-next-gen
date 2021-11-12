// SPDX-License-Identifier: Apache-2.0

package link

import (
	"net"
	"net/http"

	"github.com/vishvananda/netlink"

	"github.com/pm-web/pkg/web"
)

type Link struct {
	Action  string   `json:"Action"`
	Name    string   `json:"Name"`
	MTU     string   `json:"MTU"`
	Kind    string   `json:"Kind"`
	Mode    string   `json:"Mode"`
	Enslave []string `json:"Enslave"`
}

type LinkInfo struct {
	Index        int                     `json:"Index"`
	Mtu          int                     `json:"MTU"`
	TxQLen       int                     `json:"TxQLen"`
	Name         string                  `json:"Name"`
	HardwareAddr string                  `json:"HardwareAddr"`
	Flags        string                  `json:"Flags"`
	RawFlags     uint32                  `json:"RawFlags"`
	ParentIndex  int                     `json:"ParentIndex"`
	MasterIndex  int                     `json:"MasterIndex"`
	Namespace    string                  `json:"Namespace"`
	Alias        string                  `json:"Alias"`
	Statistics   *netlink.LinkStatistics `json:"Statistics"`

	Promisc int `json:"Promisc"`
	Xdp     struct {
		Fd       int  `json:"Fd"`
		Attached bool `json:"Attached"`
		Flags    int  `json:"Flags"`
		ProgID   int  `json:"ProgId"`
	} `json:"Xdp"`
	EncapType   string `json:"EncapType"`
	Protinfo    string `json:"Protinfo"`
	OperState   string `json:"OperState"`
	NetNsID     int    `json:"NetNsID"`
	NumTxQueues int    `json:"NumTxQueues"`
	NumRxQueues int    `json:"NumRxQueues"`
	GSOMaxSize  uint32 `json:"GSOMaxSize"`
	GSOMaxSegs  uint32 `json:"GSOMaxSegs"`
	Group       uint32 `json:"Group"`
	Slave       string `json:"Slave"`
}

func isUp(v net.Flags) bool {
	return v&net.FlagUp == net.FlagUp
}

func isBroadcastCast(v net.Flags) bool {
	return v&net.FlagBroadcast == net.FlagBroadcast
}

func isMulticastCast(v net.Flags) bool {
	return v&net.FlagMulticast == net.FlagMulticast
}

func fillOneLink(link netlink.Link) LinkInfo {
	l := LinkInfo{
		Index:        link.Attrs().Index,
		Mtu:          link.Attrs().MTU,
		TxQLen:       link.Attrs().TxQLen,
		Name:         link.Attrs().Name,
		HardwareAddr: link.Attrs().HardwareAddr.String(),
		RawFlags:     link.Attrs().RawFlags,
		ParentIndex:  link.Attrs().ParentIndex,
		MasterIndex:  link.Attrs().MasterIndex,
		Alias:        link.Attrs().Alias,
		EncapType:    link.Attrs().EncapType,
		OperState:    link.Attrs().OperState.String(),
		NetNsID:      link.Attrs().NetNsID,
		NumTxQueues:  link.Attrs().NumTxQueues,
		NumRxQueues:  link.Attrs().NumRxQueues,
		GSOMaxSize:   link.Attrs().GSOMaxSize,
		GSOMaxSegs:   link.Attrs().GSOMaxSegs,
		Group:        link.Attrs().Group,
		Statistics:   link.Attrs().Statistics,
		Promisc:      link.Attrs().Promisc,
		Flags:        link.Attrs().Flags.String(),
	}

	if link.Attrs().Protinfo != nil {
		l.Protinfo = link.Attrs().Protinfo.String()
	}

	return l
}

func (link *Link) AcquireLink(w http.ResponseWriter) error {
	if link.Name != "" {
		l, err := netlink.LinkByName(link.Name)
		if err != nil {
			return err
		}

		return web.JSONResponse(fillOneLink(l), w)
	}

	links, err := netlink.LinkList()
	if err != nil {
		return err
	}

	j := []LinkInfo{}
	for _, l := range links {
		j = append(j, fillOneLink(l))
	}

	return web.JSONResponse(j, w)
}
