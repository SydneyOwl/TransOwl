package commandline

import (
	"TransOwl/internal/cfg"
	"TransOwl/internal/netutil"
	"TransOwl/internal/terminal"
	"fmt"
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"net"
)

var (
	scanDeeper = false
)

var scanDevicesCmd = &cobra.Command{
	Use:   "scandevices",
	Short: "Print all devices available in current net.",
	Long:  `Only devices responding TransOwl UDP packet are accepted.`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, v := range processedInterfaces {
			endChan := make(chan interface{}, cfg.CACHED_UDP_READ_CHANNEL_MAX_BUFFER)
			slog.Infof("Scanning on " + v.RawInterface.Name)
			// be careful when owl is beautiful!
			tTerminal := netutil.GenerateCurrTerminal(terminal.User{
				IP:       v.CurrentIP.String(),
				UserName: userName,
			})
			go ScanDevice(tTerminal, v, true, scanDeeper, endChan)
		l:
			for {
				ans := <-endChan
				switch result := ans.(type) {
				case terminal.Terminal:
					fmt.Printf("Device found: User: %s, IP: %s, OS: %s, Arch: %s\n", result.User.UserName, result.User.IP, result.Device.OS, result.Device.Arch)
				case net.Error:
					if result.Timeout() {
						slog.Trace("Scan timedout")
						break l
					}
					slog.Errorf("Err occurred: %v", result)
				case error:
					slog.Errorf("Scan suspended due to %v", result)
				default:
					slog.Tracef("Not handled: %v", result)
				}
			}
		}
	},
}

func ScanDevice(t terminal.Terminal, v netutil.NetInterface, scanDeeper bool, toStd bool, endChan chan interface{}) {
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
	if !toStd {
		handler = udpModule.GatherDeviceAckedHandler
	}
	udpModule.StartUDPListeningWithExtraHandlers(t, cfg.MAX_DEVICE_DISCOVER_TIMEOUT, endChan, handler)
}

func init() {
	scanDevicesCmd.Flags().BoolVarP(&scanDeeper, "deepscan", "d", false, "Scan in 255.255.255.255; If not specified, devices with the same network segment as the NIC are scanned.")
	BaseCmd.AddCommand(scanDevicesCmd)
}
