// Package related_resp refines response related to device discovery.
package related_resp

import (
	"TransOwl/internal/cfg"
	"TransOwl/internal/terminal"
	"encoding/json"
	"github.com/gookit/slog"
)

type FixedHeader struct {
	Type uint   `json:"type"`
	Flag string `json:"flag"`
	Time int64  `json:"time"`
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

func ParseResponseToTerminal(buf []byte) (*terminal.Terminal, uint, error) {
	slog.Debugf("Received: %s", string(buf))
	ter := DeviceDiscovery{}
	err := json.Unmarshal(buf, &ter)
	if err != nil {
		slog.Warnf("Failed to parsE: %v", err)
		return nil, 0, err
	}
	if ter.Flag == cfg.TRANSOWL_FLAG {
		return &ter.Terminal, ter.Type, nil
	}
	slog.Debugf("Not a valid transowl packet")
	return nil, 0, ERR_TRANSOWL_FLAG_WRONG
}
