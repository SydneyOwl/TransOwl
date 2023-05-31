package netutil

import (
	"TransOwl/internal/cfg"
	"TransOwl/internal/terminal"
	"TransOwl/internal/terminal/related_resp"
	"encoding/json"
	"github.com/gookit/slog"
	"os"
	"runtime"
	"time"
)

func GenerateCurrTerminal(user terminal.User) terminal.Terminal {
	return terminal.Terminal{
		Device: terminal.Device{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
		User: user,
	}
}
func GenerateQueryDeviceRequestJSON(netInterface NetInterface) string {
	request := related_resp.DeviceDiscovery{
		FixedHeader: related_resp.FixedHeader{
			Type: cfg.ASK_FOR_AVAILABLE_DEVICES,
			Flag: cfg.TRANSOWL_FLAG,
		},
		Terminal: terminal.Terminal{
			User: terminal.User{
				IP: netInterface.CurrentIP.String(),
			},
		},
	}
	data, err := json.Marshal(request)
	if err != nil {
		// This is not likely to happen
		slog.Warnf("Cannot generate req json:%v", err)
	}
	return string(data)
}
func GenerateReplyDeviceQueryJSON(tarTerminal terminal.Terminal) string {
	response := related_resp.DeviceDiscovery{
		FixedHeader: related_resp.FixedHeader{
			Type: cfg.ACK_BEING_DISCOVERED,
			Flag: cfg.TRANSOWL_FLAG,
			Time: time.Now().Unix(),
		},
		Terminal: tarTerminal,
	}
	data, err := json.Marshal(response)
	if err != nil {
		// This is not likely to happen
		slog.Warnf("Cannot generate req json:%v", err)
	}
	return string(data)
}
func GenerateReadyToSendFileJSON(tarTerminal terminal.Terminal, file os.FileInfo) string {
	request := related_resp.FileTransfer{
		FixedHeader: related_resp.FixedHeader{
			Type: cfg.READY_TO_SEND_FILE,
			Flag: cfg.TRANSOWL_FLAG,
			Time: time.Now().Unix(),
		},
		Terminal: tarTerminal,
		File: terminal.File{
			FileName: file.Name(),
			FileSize: uint64(file.Size()),
		},
	}
	data, err := json.Marshal(request)
	if err != nil {
		// This is not likely to happen
		slog.Warnf("Cannot generate req json:%v", err)
	}
	return string(data)
}
