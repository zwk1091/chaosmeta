package loader

import "github.com/traas-stack/chaosmeta/chaosmetad/pkg/injector"

type Args struct {
	SetConfig string `json:"setConfig,omitempty"`
	Load      string `json:"load,omitempty"`
	Info      injector.BaseInfo
	Pid       int    `json:"pid,omitempty"`
	Key       string `json:"key,omitempty"`
}
