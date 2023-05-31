package test

import (
	"TransOwl/internal/cfg"
	"TransOwl/internal/netutil"
	"TransOwl/pkg/logger"
	"sync"
	"testing"
)

func TestUDPPackSend(t *testing.T) {
	logger.InitTraceLevelLogs()
	availInterfaces, _ := netutil.GetAvailableNetInterfaces()
	wifi, _ := netutil.GetNetInterfacesByName("Wi-Fi", availInterfaces)
	err := netutil.NewUDPModule(*wifi).SendDiscoverDevicesPacket(cfg.NETTYPE_WHOLENET)
	if err != nil {
		t.Fatal(err)
	}
}
func TestConcurrentPackSend(t *testing.T) {
	logger.InitTraceLevelLogs()
	wg := sync.WaitGroup{}
	availInterfaces, _ := netutil.GetAvailableNetInterfaces()
	wifi, _ := netutil.GetNetInterfacesByName("Wi-Fi", availInterfaces)
	udp := netutil.NewUDPModule(*wifi)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			//udp := netutil.NewUDPModule(*wifi, usr)
			err := udp.SendDiscoverDevicesPacket(cfg.NETTYPE_WHOLENET)
			if err != nil {
				t.Error(err)
				return
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
