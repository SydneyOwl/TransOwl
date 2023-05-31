package netutil

import (
	"TransOwl/internal/cfg"
	"TransOwl/internal/terminal"
	"fmt"
	"github.com/gookit/slog"
	"net"
	"sync"
	"time"
)

// Handlers should not return any value.
type ActionHandler func(respType uint, targetTerminal terminal.Terminal, currentTerminal terminal.Terminal, informChan chan interface{}, wg *sync.WaitGroup)

// if we received device discovery request we call this handler.
func (udpModule *UDPModule) ReplyDiscoverDevicesHandler(bit uint, targetTerminal terminal.Terminal, currentTerminal terminal.Terminal, _ chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	if bit != cfg.ASK_FOR_AVAILABLE_DEVICES {
		return
	}
	// sleep for a while so server can switch to listen mode
	time.Sleep(time.Millisecond * 2)
	slog.Trace("Replying ACK_ONLINE...")
	if err := udpModule.sendUDPPacket(net.ParseIP(targetTerminal.User.IP), GenerateReplyDeviceQueryJSON(currentTerminal)); err != nil {
		slog.Debugf("Failed to reply ASK_DEVICE request:%v", err)
	}
}
func (udpModule *UDPModule) PrintDeviceAckedHandler(bit uint, targetTerminal terminal.Terminal, _ terminal.Terminal, _ chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	if bit != cfg.ACK_BEING_DISCOVERED {
		return
	}
	fmt.Printf("Device found: User: %s, IP: %s, OS: %s, Arch: %s\n", targetTerminal.User.UserName, targetTerminal.User.IP, targetTerminal.Device.OS, targetTerminal.Device.Arch)
}
func (udpModule *UDPModule) GatherDeviceAckedHandler(bit uint, targetTerminal terminal.Terminal, _ terminal.Terminal, msgChan chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	if bit != cfg.ACK_BEING_DISCOVERED {
		return
	}
	msgChan <- targetTerminal
}

// if we received send_file request and if we are free we'll call this
func (udpModule *UDPModule) ReplyReadyToReceiveFile(bit uint, targetTerminal terminal.Terminal, currentTerminal terminal.Terminal, msgChan chan interface{}, wg *sync.WaitGroup) {
	if bit != cfg.READY_TO_SEND_FILE {
		return
	}
	//if err := udpModule.sendUDPPacket(net.ParseIP(targetTerminal.User.IP), GenerateReadyToSendFileJSON(currentTerminal)); err != nil {
	//	slog.Debugf("Failed to reply ASK_DEVICE request:%v", err)
	//}
}
