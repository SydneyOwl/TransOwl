package netutil

import (
	"TransOwl/internal/cfg"
	"TransOwl/internal/terminal"
	"TransOwl/internal/terminal/related_resp"
	"github.com/gookit/slog"
	"net"
	"sync"
	"time"
)

var (
	instance map[string]*UDPModule
	mu       sync.Mutex
)

type UDPModule struct {
	targetInterface NetInterface
	// Only one goroutine is allowed to use the io
	sendingMutex *sync.Mutex
	handlers     []ActionHandler
}

func init() {
	instance = map[string]*UDPModule{}
}
func (udpModule *UDPModule) registerHandlers(actionHandler ...ActionHandler) {
	udpModule.handlers = append(udpModule.handlers, actionHandler...)
}

// Singleton: for concurrent
func NewUDPModule(targetInterface NetInterface) *UDPModule {
	name := targetInterface.RawInterface.Name
	if instance[name] == nil {
		mu.Lock()
		defer mu.Unlock()
		if instance[name] == nil {
			instance[name] = &UDPModule{targetInterface: targetInterface, sendingMutex: &sync.Mutex{}}
		}
	}
	return instance[name]
}

func (udpModule *UDPModule) StartUDPListeningWithDefaultHandlers(terminalCurr terminal.Terminal, timeout time.Duration, msgChan chan interface{}) {
	udpModule.registerHandlers(udpModule.ReplyDiscoverDevicesHandler)
	udpModule.StartUDPListeningWithHandlers(terminalCurr, timeout, msgChan)
}
func (udpModule *UDPModule) StartUDPListeningWithExtraHandlers(terminalCurr terminal.Terminal, timeout time.Duration, msgChan chan interface{}, actionHandler ...ActionHandler) {
	udpModule.registerHandlers(udpModule.ReplyDiscoverDevicesHandler)
	udpModule.registerHandlers(actionHandler...)
	udpModule.StartUDPListeningWithHandlers(terminalCurr, timeout, msgChan)
}
func (udpModule *UDPModule) StartUDPListeningWithOutDefaultHandlers(terminalCurr terminal.Terminal, timeout time.Duration, msgChan chan interface{}, actionHandler ...ActionHandler) {
	udpModule.registerHandlers(actionHandler...)
	udpModule.StartUDPListeningWithHandlers(terminalCurr, timeout, msgChan)
}

// Use two ports to avoid concurrent problems
func (udpModule *UDPModule) StartUDPListeningWithHandlers(terminalCurr terminal.Terminal, timeout time.Duration, informChan chan interface{}) {
	localAddr := net.UDPAddr{
		IP:   udpModule.targetInterface.CurrentIP,
		Port: cfg.UDP_PORT_INWARD,
	}
	var conn *net.UDPConn
	var err error
	conn, err = net.ListenUDP("udp", &localAddr)
	defer conn.Close()
	if timeout.Seconds() > 0 {
		err = conn.SetDeadline(time.Now().Add(timeout))
		if err != nil {
			informChan <- err
			return
		}
	}
	if err != nil {
		// Tell main thread StartUDPListeningWithHandlers cannot listen
		informChan <- err
		return
	}
	for {
		buf := make([]byte, 4096)
		for {
			//slog.Trace("Start reading from udp...")
			udpBytes, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				_ = conn.Close()
				informChan <- err
				return
			}
			parseResult, reqCode, err := related_resp.ParseResponseToTerminal(buf[:udpBytes])
			if err != nil {
				slog.Debugf("Cannot parse response to terminal:%v", err)
				continue
			}
			wg := sync.WaitGroup{}
			for _, v := range udpModule.handlers {
				wg.Add(1)
				// We don't need t wait 2ms now
				go v(reqCode, *parseResult, terminalCurr, informChan, &wg)
			}
			// Wait until all go routine in the same interface are done
			wg.Wait()
		}
	}
}

func (udpModule *UDPModule) sendUDPPacket(targetIp net.IP, msg string) error {
	udpModule.sendingMutex.Lock()
	defer udpModule.sendingMutex.Unlock()
	localAddr := net.UDPAddr{
		IP:   udpModule.targetInterface.CurrentIP,
		Port: cfg.UDP_PORT_OUTWARD,
	}
	broadcastAddr := net.UDPAddr{
		IP:   targetIp,
		Port: cfg.UDP_PORT_INWARD,
	}
	conn, err := net.DialUDP("udp", &localAddr, &broadcastAddr)
	if err != nil {
		slog.Warnf("Cannot send udp packet: %v", err)
		return ERR_FAILED_TO_SEND_UDP_PACKET
	}
	_, err = conn.Write([]byte(msg))
	if err != nil {
		slog.Warnf("Cannot send udp packet: %v", err)
		return ERR_FAILED_TO_SEND_UDP_PACKET
	}
	_ = conn.Close()
	slog.Trace("UDP packet sent!")
	return nil
}

// e.g. 192.168.1.x
func (udpModule *UDPModule) SendUDPBroadcastWithinSameSegment(msg string) error {
	return udpModule.sendUDPPacket(udpModule.targetInterface.MaxIP, msg)
}

// 255.255.255.255
func (udpModule *UDPModule) SendUDPBroadcastToWholeNet(msg string) error {
	return udpModule.sendUDPPacket(net.ParseIP("255.255.255.255"), msg)
}
func (udpModule *UDPModule) SendDiscoverDevicesPacket(netType uint) error {
	switch netType {
	case cfg.NETTYPE_SAMESEGMENT:
		return udpModule.SendUDPBroadcastWithinSameSegment(GenerateQueryDeviceRequestJSON(udpModule.targetInterface))
	case cfg.NETTYPE_WHOLENET:
		return udpModule.SendUDPBroadcastToWholeNet(GenerateQueryDeviceRequestJSON(udpModule.targetInterface))
	default:
		return udpModule.SendUDPBroadcastToWholeNet(GenerateQueryDeviceRequestJSON(udpModule.targetInterface))
	}
}
