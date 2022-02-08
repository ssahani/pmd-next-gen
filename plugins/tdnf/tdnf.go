// SPDX-License-Identifier: Apache-2.0
// Copyright 2022 VMware, Inc.

package tdnf

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/pmd-nextgen/pkg/jobs"
	"github.com/pmd-nextgen/pkg/system"
	"github.com/pmd-nextgen/pkg/validator"
	"github.com/pmd-nextgen/pkg/web"
)

type ListItem struct {
	Name string `json:"Name"`
	Arch string `json:"Arch"`
	Evr  string `json:"Evr"`
	Repo string `json:"Repo"`
}

type Repo struct {
	Repo     string `json:"Repo"`
	RepoName string `json:"RepoName"`
	Enabled  bool   `json:"Enabled"`
}

type Info struct {
	Name        string `json:"Name"`
	Arch        string `json:"Arch"`
	Evr         string `json:"Evr"`
	InstallSize int    `json:"InstallSize"`
	Repo        string `json:"Repo"`
	Summary     string `json:"Summary"`
	Url         string `json:"Url"`
	License     string `json:"License"`
	Description string `json:"Description"`
}

type Options struct {
	AllowErasing    bool     `tdnf:"--allowerasing"`
	Best            bool     `tdnf:"--best"`
	CacheOnly       bool     `tdnf:"--cacheonly"`
	Config          string   `tdnf:"--config"`
	DisableRepo     []string `tdnf:"--disablerepo"`
	DisableExcludes bool     `tdnf:"--disableexcludes"`
	DownloadDir     string   `tdnf:"--downloaddir"`
	DownloadOnly    bool     `tdnf:"--downloadonly"`
	EnableRepo      []string `tdnf:"--enablerepo"`
	Exclude         string   `tdnf:"--exclude"`
	InstallRoot     string   `tdnf:"--installroot"`
	NoAutoRemove    bool     `tdnf:"--noautoremove"`
	NoGPGCheck      bool     `tdnf:"--nogpgcheck"`
	NoPlugins       bool     `tdnf:"--noplugins"`
	RebootRequired  bool     `tdnf:"--rebootrequired"`
	Refresh         bool     `tdnf:"--refresh"`
	ReleaseVer      string   `tdnf:"--releasever"`
	RepoId          string   `tdnf:"--repoid"`
	RepoFromPath    string   `tdnf:"--repofrompath"`
	Security        bool     `tdnf:"--security"`
	SecSeverity     string   `tdnf:"--sec-severity"`
	SetOpt          []string `tdnf:"--setopt"`
	SkipConflicts   bool     `tdnf:"--skipconflicts"`
	SkipDigest      bool     `tdnf:"--skipdigest"`
	SkipObsoletes   bool     `tdnf:"--skipobsoletes"`
	SkipSignature   bool     `tdnf:"--skipsignature"`
}

func TdnfOptions(options *Options) []string {
	var strOptions []string

	v := reflect.ValueOf(options).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if opt := field.Tag.Get("tdnf"); opt != "" {
			value := v.Field(i).Interface()
			switch value.(type) {
			case bool:
				if value.(bool) {
					strOptions = append(strOptions, opt)
				}
			case string:
				if strVal := value.(string); strVal != "" {
					strOptions = append(strOptions, opt+"="+strVal)
				}
			case []string:
				for _, s := range value.([]string) {
					strOptions = append(strOptions, opt+"="+s)
				}
			}
		}
	}
	return strOptions
}

func TdnfExec(options *Options, args ...string) (string, error) {
	args = append([]string{"-j"}, args...)

	if options != nil {
		args = append(TdnfOptions(options), args...)
	}
	fmt.Printf("calling tdnf %v\n", args)
	s, err := system.ExecAndCapture("tdnf", args...)
	if err != nil {
		return "", err
	}
	return s, nil
}

func AcquireList(w http.ResponseWriter, pkg string, options Options) error {
	job := jobs.CreateJob(func() (interface{}, error) {
		var s string
		var err error
		if !validator.IsEmpty(pkg) {
			s, err = TdnfExec(&options, "list", pkg)
		} else {
			s, err = TdnfExec(&options, "list")
		}
		var list interface{}
		if err := json.Unmarshal([]byte(s), &list); err != nil {
			return nil, err
		}

		return list, err
	})
	return jobs.AcceptedResponse(w, job)
}

func AcquireRepoList(w http.ResponseWriter, options Options) error {
	s, err := TdnfExec(&options, "repolist")
	if err != nil {
		log.Errorf("Failed to execute tdnf repolist: %v", err)
		return err
	}

	var repoList interface{}
	if err := json.Unmarshal([]byte(s), &repoList); err != nil {
		return err
	}

	return web.JSONResponse(repoList, w)
}

func AcquireInfoList(w http.ResponseWriter, pkg string, options Options) error {
	job := jobs.CreateJob(func() (interface{}, error) {
		var s string
		var err error
		if !validator.IsEmpty(pkg) {
			s, err = TdnfExec(&options, "info", pkg)
		} else {
			s, err = TdnfExec(&options, "info")
		}
		if err != nil {
			return nil, err
		}

		var list interface{}
		if err := json.Unmarshal([]byte(s), &list); err != nil {
			return nil, err
		}

		return list, err
	})
	return jobs.AcceptedResponse(w, job)
}

func AcquireMakeCache(w http.ResponseWriter, options Options) error {
	job := jobs.CreateJob(func() (interface{}, error) {
		_, err := TdnfExec(&options, "makecache")
		return nil, err
	})
	return jobs.AcceptedResponse(w, job)
}

func AcquireClean(w http.ResponseWriter, options Options) error {
	_, err := TdnfExec(&options, "clean", "all")
	if err != nil {
		log.Errorf("Failed to execute tdnf clean all': %v", err)
		return err
	}

	return web.JSONResponse("cleaned", w)
}
