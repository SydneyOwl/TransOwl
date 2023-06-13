// Package related_resp refines response related to device discovery.
package related_resp

import (
	"github.com/sydneyowl/TransOwl/internal/terminal"
)

type FixedHeader struct {
	Type      uint   `json:"type"`
	Flag      string `json:"flag"`
	Time      int64  `json:"time"`
	TransPswd string `json:"transPswd"`
}
type DeviceDiscovery struct {
	FixedHeader
	terminal.Terminal
}

type FileTransfer struct {
	FixedHeader
	terminal.Terminal
	File terminal.File `json:"file"`
}
