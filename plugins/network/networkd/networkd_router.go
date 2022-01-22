// SPDX-License-Identifier: Apache-2.0
// Copyright 2022 VMware, Inc.

package networkd

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/pmd-nextgen/pkg/web"
)

func routerConfigureNetwork(w http.ResponseWriter, r *http.Request) {
	n, err := decodeNetworkJSONRequest(r)
	if err != nil {
		http.Error(w, "Error decoding request", http.StatusBadRequest)
		return
	}

	if err := n.ConfigureNetwork(r.Context(), w); err != nil {
		web.JSONResponseError(err, w)
	}
}

func routerRemoveNetwork(w http.ResponseWriter, r *http.Request) {
	n, err := decodeNetworkJSONRequest(r)
	if err != nil {
		http.Error(w, "Error decoding request", http.StatusBadRequest)
		return
	}

	if err := n.RemoveNetwork(r.Context(), w); err != nil {
		web.JSONResponseError(err, w)
	}
}

func routerAcquireLinks(w http.ResponseWriter, r *http.Request) {
	l, err := AcquireLinks(r.Context())
	if  err != nil {
		web.JSONResponseError(err, w)
	}

	web.JSONResponse(l, w)
}

func routerAcquireNetworkState(w http.ResponseWriter, r *http.Request) {
	n, err := AcquireNetworkState(r.Context())
	if err != nil {
		web.JSONResponseError(err, w)
	}

	web.JSONResponse(n, w)
}

func routerConfigureNetDev(w http.ResponseWriter, r *http.Request) {
	n, err := decodeNetDevJSONRequest(r)
	if err != nil {
		http.Error(w, "Error decoding request", http.StatusBadRequest)
		return
	}

	if err := n.ConfigureNetDev(r.Context(), w); err != nil {
		web.JSONResponseError(err, w)
	}
}

func RegisterRouterNetworkd(router *mux.Router) {
	n := router.PathPrefix("/networkd").Subrouter().StrictSlash(false)

	n.HandleFunc("/network/describenetwork", routerAcquireNetworkState).Methods("GET")
	n.HandleFunc("/network/describelinks", routerAcquireLinks).Methods("GET")
	n.HandleFunc("/network/configure", routerConfigureNetwork).Methods("POST")
	n.HandleFunc("/network/remove", routerRemoveNetwork).Methods("DELETE")

	n.HandleFunc("/netdev/configure", routerConfigureNetDev).Methods("POST")
}
