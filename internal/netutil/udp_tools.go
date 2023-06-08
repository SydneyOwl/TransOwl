package netutil

import (
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/sydneyowl/TransOwl/internal/cfg"
	"github.com/sydneyowl/TransOwl/internal/terminal"
	"github.com/sydneyowl/TransOwl/internal/terminal/related_resp"

	"github.com/gookit/slog"
)

var (
	instance      map[string]*UDPModule
	instanceMutex sync.Mutex

	status      = cfg.STATUS_OK
	statusMutex sync.RWMutex
)

type UDPModule struct {
	targetInterface NetInterface
	// Only one goroutine is allowed to use the io
	sendingMutex   *sync.Mutex
	listeningMutex *sync.Mutex
	handlers       []ActionHandler
}

type UDPModuleWithContext struct {
	targetIP *net.UDPAddr
	conn     *net.UDPConn
}

func init() {
	instance = map[string]*UDPModule{}
}

func SetTerminalState(statusCode int) {
	statusMutex.Lock()
	status = statusCode
	statusMutex.Unlock()
}

func SetTerminalStateOnSuccess(statusCode int) {
	statusMutex.Lock()
	if statusCode == cfg.STATUS_OK {
		status = statusCode
	}
	statusMutex.Unlock()
}

func GetTerminalState() int {
	var stat int
	statusMutex.RLock()
	stat = status
	statusMutex.RUnlock()
	return stat
}

func (udpModule *UDPModule) registerHandlers(actionHandler ...ActionHandler) {
	udpModule.handlers = append(udpModule.handlers, actionHandler...)
}

// Singleton: for concurrent
func NewUDPModule(targetInterface NetInterface) *UDPModule {
	name := targetInterface.RawInterface.Name
	if instance[name] == nil {
		instanceMutex.Lock()
		defer instanceMutex.Unlock()
		if instance[name] == nil {
			instance[name] = &UDPModule{targetInterface: targetInterface, sendingMutex: &sync.Mutex{}, listeningMutex: &sync.Mutex{}}
		}
	}
	return instance[name]
}

// This is not concurrent-safe.
func NewUDPModuleWithContext(currInterface NetInterface, targetIP net.IP) (*UDPModuleWithContext, error) {
	localAddr := &net.UDPAddr{
		IP:   currInterface.CurrentIP,
		Port: cfg.UDP_PORT_INWARD,
	}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, err
	}
	targetAddr := &net.UDPAddr{
		IP:   targetIP,
		Port: cfg.UDP_PORT_INWARD,
	}
	return &UDPModuleWithContext{conn: conn, targetIP: targetAddr}, nil
}

// This is a blocked function!
func (udpModule *UDPModuleWithContext) ReadFromUDPAndParseWithTimeout(timeout time.Duration) (interface{}, uint, error) {
	buf := make([]byte, 4096)
	if timeout.Seconds() > 0 {
		err := udpModule.conn.SetDeadline(time.Now().Add(timeout))
		if err != nil {
			return nil, 0, err
		}
	}
	udpBytes, _, err := udpModule.conn.ReadFromUDP(buf)
	if err != nil {
		return nil, 0, err
	}
	slog.Debug(string(buf))
	parse, u, err := checkResponseAndParse(buf[:udpBytes])
	if err != nil {
		return nil, 0, err
	}
	return parse, u, nil
}
func (udpModule *UDPModuleWithContext) SendToUDP(msg string) error {
	_, err := udpModule.conn.WriteToUDP([]byte(msg), udpModule.targetIP)
	return err
}
func (udpModule *UDPModuleWithContext) ShutConn() {
	_ = udpModule.conn.Close()
}
func (udpModule *UDPModule) StartUDPListeningWithDefaultHandlers(terminalCurr terminal.Terminal, timeout time.Duration, msgChan chan interface{}) {
	udpModule.registerHandlers(udpModule.ReplyDiscoverDevicesHandler, udpModule.ReplySearchDevicesHandler)
	udpModule.StartUDPListeningWithHandlers(terminalCurr, timeout, msgChan)
}
func (udpModule *UDPModule) StartUDPListeningWithExtraHandlers(terminalCurr terminal.Terminal, timeout time.Duration, msgChan chan interface{}, actionHandler ...ActionHandler) {
	udpModule.registerHandlers(udpModule.ReplyDiscoverDevicesHandler, udpModule.ReplySearchDevicesHandler)
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
	udpModule.listeningMutex.Lock()
	defer udpModule.listeningMutex.Unlock()
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
		//slog.Trace("Start reading from udp...")
		udpBytes, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			informChan <- err
			return
		}
		parseResult, reqCode, err := checkResponseAndParse(buf[:udpBytes])
		if err != nil {
			slog.Debugf("Cannot parse : %v", err)
			continue
		}
		wg := sync.WaitGroup{}
		for _, v := range udpModule.handlers {
			wg.Add(1)
			// We don't need t wait 2ms now
			go v(reqCode, parseResult, terminalCurr, informChan, &wg)
		}
		// Wait until all go routine in the same interface are done
		wg.Wait()
		slog.Trace("All go routines exited.")
	}
}

func (udpModule *UDPModule) sendUDPPacket(targetIp net.IP, msg string) error {
	udpModule.sendingMutex.Lock()
	defer udpModule.sendingMutex.Unlock()
	slog.Tracef("Packet: %s sending to %s", msg, targetIp.String())
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

func (udpModule *UDPModule) SendP2PUDPPacket(targetIp string, msg string) error {
	return udpModule.sendUDPPacket(net.ParseIP(targetIp), msg)
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
func (udpModule *UDPModule) SendSearchDevicesPacket(netType uint, name string) error {
	switch netType {
	case cfg.NETTYPE_SAMESEGMENT:
		return udpModule.SendUDPBroadcastWithinSameSegment(GenerateAskForTargetDeviceQueryJSON(udpModule.targetInterface, name))
	case cfg.NETTYPE_WHOLENET:
		return udpModule.SendUDPBroadcastToWholeNet(GenerateAskForTargetDeviceQueryJSON(udpModule.targetInterface, name))
	default:
		return udpModule.SendUDPBroadcastToWholeNet(GenerateAskForTargetDeviceQueryJSON(udpModule.targetInterface, name))
	}
}
func checkResponseAndParse(buf []byte) (interface{}, uint, error) {
	slog.Debugf("Received: %s", string(buf))
	ter := related_resp.FixedHeader{}
	err := json.Unmarshal(buf, &ter)
	if err != nil {
		slog.Warnf("Failed to parsE: %v", err)
		return nil, 0, err
	}
	if ter.Flag == cfg.TRANSOWL_FLAG {
		switch ter.Type {
		case cfg.ACK_BEING_DISCOVERED, cfg.ASK_FOR_AVAILABLE_DEVICES, cfg.SEARCH_FOR_DEVICE, cfg.ACK_I_AM_THE_DEVICE:
			ter := related_resp.DeviceDiscovery{}
			err = json.Unmarshal(buf, &ter)
			if err != nil {
				slog.Warnf("Failed to parsE: %v", err)
				return nil, ter.Type, err
			}
			return ter, ter.Type, nil
		case cfg.READY_TO_SEND_FILE, cfg.READY_TO_RECV_FILE, cfg.REFUSED_TO_RECV_FILE:
			fileInfo := related_resp.FileTransfer{}
			err = json.Unmarshal(buf, &fileInfo)
			if err != nil {
				slog.Warnf("Failed to parsE: %v", err)
				return nil, ter.Type, err
			}
			return fileInfo, ter.Type, nil
		}
	}
	slog.Debugf("Not a valid transowl type")
	return nil, 0, ERR_TYPE_NOT_DEFINED
}
