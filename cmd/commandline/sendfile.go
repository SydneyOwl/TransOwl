package commandline

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"github.com/sydneyowl/TransOwl/internal/cfg"
	"github.com/sydneyowl/TransOwl/internal/netutil"
	"github.com/sydneyowl/TransOwl/internal/terminal"
)

var sendUser string
var filePath string
var sendFile = &cobra.Command{
	Use:   "sendfile",
	Short: "send file to someone",
	Long:  `example: ./TransOwl sendfile --filepath owl.doc --sendto TransOwlUser-d83hf`,
	Run: func(cmd *cobra.Command, args []string) {
		if sendUser == "" || filePath == "" {
			slog.Error("specify receiver via --sendto and file ready to be sent via --filepath")
			return
		}
		file, err := os.Stat(filePath)
		if err != nil {
			slog.Errorf("Cannot read file: %s", err)
			return
		}
		fileName := file.Name()
		fileSize := file.Size()
		// Deny file > 100MB
		// Since we read directly from os.
		if fileSize > 104857600 {
			slog.Error("We currently don't support send file at that big!")
			return
		}
		slog.Infof("file to be sent: %s, %s", fileName, humanize.Bytes(uint64(fileSize)))
		// get available clients
		availableReceiver := terminal.Terminal{}
		slog.Info("Searching...")
		resRefer := make(chan interface{}, cfg.CACHED_UDP_READ_CHANNEL_MAX_BUFFER)
		waitFinddev := sync.WaitGroup{}
		for _, v := range processedInterfaces {
			waitFinddev.Add(1)
			slog.Trace("Scanning on " + v.RawInterface.Name)
			// be careful when owl is beautiful!
			go func(v netutil.NetInterface) {
				defer waitFinddev.Done()
				FindDevice(sendUser, netutil.GenerateCurrTerminal(terminal.User{
					IP:       v.CurrentIP.String(),
					UserName: userName,
				}), v, scanDeeper, resRefer)
			}(v)
		}
		waitFinddev.Wait()
	l:
		for {
			ans := <-resRefer
			switch result := ans.(type) {
			case terminal.Terminal:
				availableReceiver = result
				break l
			case net.Error:
				if result.Timeout() {
					slog.Warnf("Scan timedout: No user named %s found. Try using --deepscan to scan deeper.", sendUser)
					return
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
		slog.Debugf("users: %s", availableReceiver)
		slog.Noticef("User found at interface %s", availableReceiver.FoundAt)
		targetInterface, _ := netutil.GetNetInterfacesByName(availableReceiver.FoundAt, processedInterfaces)
		tml := netutil.GenerateCurrTerminal(terminal.User{
			IP:       targetInterface.CurrentIP.String(),
			UserName: userName,
		})
		//Now establish connection(p2p) with target.
		udp, err := netutil.NewUDPModuleWithContext(*targetInterface, net.ParseIP(availableReceiver.User.IP))
		err = udp.SendToUDP(netutil.GenerateReadyToSendFileJSON(tml, file))
		if err != nil {
			slog.Errorf("Cannot ask client: %v", err)
			return
		}
		_, tp, err := udp.ReadFromUDPAndParseWithTimeout(time.Second * cfg.MAX_STANDBY_ACK_TIMEOUT)
		if err != nil {
			slog.Errorf("Err occurred: %v", err)
			return
		}
		if tp != cfg.READY_TO_RECV_FILE {
			slog.Errorf("Client sent a message but cant be understood")
			return
		}
		// tell target that we are going send files!
		// we send file now!
		// TODO: ADD SEND
		slog.Info("Ready to send file!")
	},
}

func init() {
	sendFile.Flags().StringVar(&sendUser, "sendto", "", "Send to user")
	sendFile.Flags().StringVar(&filePath, "filepath", "", "file to be sent")
	_ = sendFile.MarkFlagRequired("filepath")
	_ = sendFile.MarkFlagRequired("sendto")
	// Not available right now
	BaseCmd.AddCommand(sendFile)
}
func FindDevice(target string, t terminal.Terminal, v netutil.NetInterface, scanDeeper bool, endChan chan interface{}) {
	var deepbit uint = cfg.NETTYPE_SAMESEGMENT
	if scanDeeper {
		deepbit = cfg.NETTYPE_WHOLENET
	}
	slog.Debugf("Sending ASK_FOR_DEVICE req to %s...", v.RawInterface.Name)
	udpModule := netutil.NewUDPModule(v)
	err := udpModule.SendSearchDevicesPacket(deepbit, target)
	if err != nil {
		slog.Fatalf("failed to send udp packet: %v", err)
		return
	}
	udpModule.StartUDPListeningWithExtraHandlers(t, cfg.MAX_DEVICE_DISCOVER_TIMEOUT*time.Second, endChan, udpModule.ReplyReceivedSearchDevicesAckHandler)
}
