// SPDX-License-Identifier: Apache-2.0

package systemd

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/pm-web/pkg/web"
)

func routerAcquireSystemdManagerProperty(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)

	if err := ManagerAcquireSystemProperty(r.Context(), w, v["property"]); err != nil {
		web.JSONResponseError(err, w)
	}
}

func routerConfigureSystemdConf(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if err := GetSystemConf(w); err != nil {
			web.JSONResponseError(err, w)
			return
		}
	case "POST":
		if err := UpdateSystemConf(w, r); err != nil {
			web.JSONResponseError(err, w)
		}
	}
}

func routerConfigureUnit(w http.ResponseWriter, r *http.Request) {
	u := new(UnitAction)

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Error decoding request", http.StatusBadRequest)
		return
	}

	if err := u.UnitCommands(r.Context()); err != nil {
		web.JSONResponseError(err, w)
		return
	}

	web.JSONResponse("", w)
}

func routerAcquireAllSystemdUnits(w http.ResponseWriter, r *http.Request) {
	if err := ListUnits(r.Context(), w); err != nil {
		web.JSONResponseError(err, w)
	}
}

func routerAcquireUnitStatus(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	u := UnitAction{
		Unit: v["unit"],
	}

	if err := u.AcquireUnitStatus(r.Context(), w); err != nil {
		web.JSONResponseError(err, w)
	}
}

func routerAcquireUnitProperty(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	u := UnitAction{
		Unit:     v["unit"],
		Property: v["property"],
	}

	if err := u.AcquireUnitProperty(r.Context(), w); err != nil {
		web.JSONResponseError(err, w)
	}
}

func routerAcquireUnitPropertyAll(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	u := UnitAction{
		Unit: v["unit"],
	}

	if err := u.AcquireAllUnitProperty(r.Context(), w); err != nil {
		web.JSONResponseError(err, w)
	}
}

func routerAcquireUnitTypeProperty(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	u := UnitAction{
		Unit:     v["unit"],
		UnitType: v["unittype"],
		Property: v["property"],
	}

	u.AcquireUnitTypeProperty(r.Context(), w)
}

func RegisterRouterSystemd(router *mux.Router) {
	n := router.PathPrefix("/service").Subrouter()

	// systemd unit commands
	n.HandleFunc("/systemd", routerConfigureUnit).Methods("POST")

	// systemd unit status and property
	n.HandleFunc("/systemd/manager/property/{property}", routerAcquireSystemdManagerProperty).Methods("GET")
	n.HandleFunc("/systemd/units", routerAcquireAllSystemdUnits).Methods("GET")
	n.HandleFunc("/systemd/{unit}/status", routerAcquireUnitStatus).Methods("GET")
	n.HandleFunc("/systemd/{unit}/property", routerAcquireUnitProperty).Methods("GET")
	n.HandleFunc("/systemd/{unit}/propertyall", routerAcquireUnitPropertyAll).Methods("GET")
	n.HandleFunc("/systemd/{unit}/property/{unittype}", routerAcquireUnitTypeProperty).Methods("GET")

	// systemd configuration
	n.HandleFunc("/systemd/conf", routerConfigureSystemdConf)
	n.HandleFunc("/systemd/conf/update", routerConfigureSystemdConf)
}
