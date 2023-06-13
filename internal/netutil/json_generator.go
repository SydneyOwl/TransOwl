package netutil

import (
	"encoding/json"
	"github.com/gookit/slog"
	"github.com/sydneyowl/TransOwl/internal/cfg"
	"github.com/sydneyowl/TransOwl/internal/terminal"
	"github.com/sydneyowl/TransOwl/internal/terminal/related_resp"
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
func GenerateAskForTargetDeviceQueryJSON(netInterface NetInterface, name string) string {
	response := related_resp.DeviceDiscovery{
		FixedHeader: related_resp.FixedHeader{
			Type: cfg.SEARCH_FOR_DEVICE,
			Flag: cfg.TRANSOWL_FLAG,
			Time: time.Now().Unix(),
		},
		Terminal: terminal.Terminal{User: terminal.User{
			// target username
			UserName: name,
			// This device's ip
			IP: netInterface.CurrentIP.String(),
		},
		},
	}
	data, err := json.Marshal(response)
	if err != nil {
		// This is not likely to happen
		slog.Warnf("Cannot generate req json:%v", err)
	}
	return string(data)
}
func GenerateIAmTheDeviceQueryJSON(tarTerminal terminal.Terminal) string {
	response := related_resp.DeviceDiscovery{
		FixedHeader: related_resp.FixedHeader{
			Type: cfg.ACK_I_AM_THE_DEVICE,
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
func GenerateReadyToSendFileJSON(myTerminal terminal.Terminal, pswd string, file os.FileInfo) string {
	request := related_resp.FileTransfer{
		FixedHeader: related_resp.FixedHeader{
			Type:      cfg.READY_TO_SEND_FILE,
			Flag:      cfg.TRANSOWL_FLAG,
			Time:      time.Now().Unix(),
			TransPswd: pswd,
		},
		Terminal: myTerminal,
		File: terminal.File{
			FileName: file.Name(),
			FileSize: uint64(file.Size()),
		},
	}
	data, _ := json.Marshal(request)
	return string(data)
}
func GenerateReadyToReceiveFileJSON(myTerminal terminal.Terminal, accept bool) string {
	var typ uint = cfg.READY_TO_RECV_FILE
	if !accept {
		typ = cfg.REFUSED_TO_RECV_FILE
	}
	response := related_resp.FileTransfer{
		FixedHeader: related_resp.FixedHeader{
			Type: typ,
			Flag: cfg.TRANSOWL_FLAG,
			Time: time.Now().Unix(),
		},
		Terminal: myTerminal,
	}
	data, _ := json.Marshal(response)
	return string(data)
}
