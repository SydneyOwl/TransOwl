package netutil

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sydneyowl/TransOwl/internal/cfg"
	"github.com/sydneyowl/TransOwl/internal/terminal"
	"github.com/sydneyowl/TransOwl/internal/terminal/related_resp"

	"github.com/gookit/slog"
)

// Handlers should not return any value.
type ActionHandler func(respType uint, target interface{}, currentTerminal terminal.Terminal, informChan chan<- interface{}, wg *sync.WaitGroup)

// if we received device discovery request we call this handler.
func (udpModule *UDPModule) ReplyDiscoverDevicesHandler(bit uint, target interface{}, currentTerminal terminal.Terminal, _ chan<- interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	if bit != cfg.ASK_FOR_AVAILABLE_DEVICES {
		return
	}
	targetTerminal := target.(related_resp.DeviceDiscovery)
	// sleep for a while so server can switch to listen mode
	time.Sleep(time.Millisecond * 2)
	slog.Trace("Replying ACK_ONLINE...")
	if err := udpModule.sendUDPPacket(net.ParseIP(targetTerminal.User.IP), GenerateReplyDeviceQueryJSON(currentTerminal)); err != nil {
		slog.Debugf("Failed to reply ASK_DEVICE request:%v", err)
	}
}
func (udpModule *UDPModule) ReplyReceivedSearchDevicesAckHandler(bit uint, target interface{}, _ terminal.Terminal, informChan chan<- interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	if bit != cfg.ACK_I_AM_THE_DEVICE {
		return
	}
	targetTerminal := target.(related_resp.DeviceDiscovery)
	targetTerminal.FoundAt = udpModule.targetInterface.RawInterface.Name
	informChan <- targetTerminal.Terminal
}

// if we received device search request we call this handler.
func (udpModule *UDPModule) ReplySearchDevicesHandler(bit uint, target interface{}, currentTerminal terminal.Terminal, informChan chan<- interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	if bit != cfg.SEARCH_FOR_DEVICE {
		return
	}
	targetTerminal := target.(related_resp.DeviceDiscovery)
	// sleep for a while so server can switch to listen mode
	time.Sleep(time.Millisecond * 2)
	if targetTerminal.User.UserName != currentTerminal.User.UserName {
		slog.Debugf("Not me!")
		return
	}
	slog.Trace("Replying ACK_I_AM_THE_DEVICE...")
	if err := udpModule.sendUDPPacket(net.ParseIP(targetTerminal.User.IP), GenerateIAmTheDeviceQueryJSON(currentTerminal)); err != nil {
		slog.Debugf("Failed to reply ACK_I_AM_THE_DEVICE request:%v", err)
		return
	}
	//targetTerminal.FoundAt = udpModule.targetInterface.RawInterface.Name
	//informChan <- targetTerminal.Terminal
}
func (udpModule *UDPModule) PrintDeviceAckedHandler(bit uint, target interface{}, _ terminal.Terminal, _ chan<- interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	targetTerminal := target.(related_resp.DeviceDiscovery)
	if bit != cfg.ACK_BEING_DISCOVERED {
		return
	}
	fmt.Printf("Device found: User: %s, IP: %s, OS: %s, Arch: %s\n", targetTerminal.User.UserName, targetTerminal.User.IP, targetTerminal.Device.OS, targetTerminal.Device.Arch)
}
func (udpModule *UDPModule) GatherDeviceAckedHandler(bit uint, target interface{}, _ terminal.Terminal, msgChan chan<- interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	if bit != cfg.ACK_BEING_DISCOVERED {
		return
	}
	targetTerminal := target.(related_resp.DeviceDiscovery)
	msgChan <- targetTerminal.Terminal
}

// if we received send_file request and if we are free we'll call this
func (udpModule *UDPModule) ReplyReadyToReceiveFileHandler(bit uint, target interface{}, currentTerminal terminal.Terminal, msgChan chan<- interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	if bit != cfg.READY_TO_SEND_FILE {
		return
	}
	targetFileReq := target.(related_resp.FileTransfer)
	sendable := cfg.STATUS_OK == GetTerminalState()
	if !sendable {
		slog.Debugf("Refused request from %s since we are busy now", targetFileReq.User.UserName)
		SetTerminalState(cfg.STATUS_OK)
		return
	}
	SetTerminalState(cfg.STATUS_RECV_FILE)
	time.Sleep(2 * time.Millisecond)
	if err := udpModule.sendUDPPacket(net.ParseIP(targetFileReq.User.IP), GenerateReadyToReceiveFileJSON(currentTerminal, sendable)); err != nil {
		slog.Warnf("Failed to reply ASK_DEVICE request:%v", err)
		SetTerminalState(cfg.STATUS_OK)
		return
	}
	slog.Trace("file receive!")
	msgChan <- targetFileReq
}
func (udpModule *UDPModule) InformTCPHandler(bit uint, target interface{}, currentTerminal terminal.Terminal, msgChan chan<- interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	if bit != cfg.READY_TO_RECV_FILE {
		return
	}
	sendable := cfg.STATUS_OK == GetTerminalState()
	targetFileReq := target.(related_resp.FileTransfer)
	if !sendable {
		slog.Debugf("Busy now and do not accept other requests!")
		msgChan <- struct{}{}
		slog.Debugf("Refused request from %s since we are busy now", targetFileReq.User.UserName)
		return
	}
	slog.Trace("receive side acked ready_to_send_file")
	SetTerminalState(cfg.STATUS_SEND_FILE)
	msgChan <- targetFileReq
}
