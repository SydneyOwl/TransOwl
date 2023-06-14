package commandline

import (
	"fmt"
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"github.com/sydneyowl/TransOwl/internal/cfg"
	"github.com/sydneyowl/TransOwl/internal/netutil"
	"github.com/sydneyowl/TransOwl/internal/terminal"
	"github.com/sydneyowl/TransOwl/internal/terminal/related_resp"
	"net"
	"time"
)

var scanDevicesCmd = &cobra.Command{
	Use:     "scandevices",
	Short:   `Print all devices available in current net.`,
	Long:    `Only devices responding TransOwl UDP packet are accepted.`,
	Example: `./TransOwl scandevices`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, v := range processedInterfaces {
			endChan := make(chan interface{}, cfg.CACHED_UDP_READ_CHANNEL_MAX_BUFFER)
			slog.Infof("Scanning on " + v.RawInterface.Name)
			// be careful when owl is beautiful!
			tTerminal := netutil.GenerateCurrTerminal(terminal.User{
				IP:       v.CurrentIP.String(),
				UserName: userName,
			})
			go ScanDevice(tTerminal, v, scanDeeper, endChan)
		l:
			for {
				ans := <-endChan
				switch result := ans.(type) {
				case related_resp.DeviceDiscovery:
					fmt.Printf("Device found: User: %s, IP: %s, OS: %s, Arch: %s\n", result.User.UserName, result.User.IP, result.Device.OS, result.Device.Arch)
				case net.Error:
					if result.Timeout() {
						slog.Trace("Scan timedout")
						break l
					}
					slog.Errorf("Err occurred: %v", result)
					return
				case error:
					slog.Errorf("Scan suspended due to %v", result)
					return
				default:
					slog.Tracef("Not handled: %v", result)
				}
			}
		}
	},
}

func ScanDevice(t terminal.Terminal, v netutil.NetInterface, scanDeeper bool, endChan chan interface{}) {
	var deepbit uint = cfg.NETTYPE_SAMESEGMENT
	if scanDeeper {
		deepbit = cfg.NETTYPE_WHOLENET
	}
	slog.Debugf("Sending ASK_FOR_DEVICE req to %s...", v.RawInterface.Name)
	udpModule := netutil.NewUDPModule(v)
	err := udpModule.SendDiscoverDevicesPacket(deepbit)
	if err != nil {
		slog.Fatalf("failed to send udp packet: %v", err)
		return
	}
	handler := udpModule.PrintDeviceAckedHandler
	udpModule.StartUDPListeningWithExtraHandlers(t, cfg.MAX_DEVICE_DISCOVER_TIMEOUT*time.Second, endChan, handler)
}

func init() {
	BaseCmd.AddCommand(scanDevicesCmd)
}
