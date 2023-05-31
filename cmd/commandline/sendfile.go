package commandline

import (
	"github.com/sydneyowl/TransOwl/internal/cfg"
	"github.com/sydneyowl/TransOwl/internal/netutil"
	"github.com/sydneyowl/TransOwl/internal/terminal"
	"github.com/sydneyowl/TransOwl/internal/terminal/related_resp"
	"github.com/sydneyowl/TransOwl/pkg/util/terminalutil"
	"net"
	"os"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
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
		slog.Debugf("file to be sent: %s, %s", fileName, humanize.Bytes(uint64(fileSize)))
		// get available clients
		availableReceiver := make([]terminal.Terminal, 5)
		slog.Info("Searching...")
		for _, v := range processedInterfaces {
			resRefer := make(chan interface{}, cfg.CACHED_UDP_READ_CHANNEL_MAX_BUFFER)
			slog.Trace("Scanning on " + v.RawInterface.Name)
			// be careful when owl is beautiful!
			go ScanDevice(netutil.GenerateCurrTerminal(terminal.User{
				IP:       v.CurrentIP.String(),
				UserName: userName,
			}), v, scanDeeper, false, resRefer)
		l:
			for {
				ans := <-resRefer
				switch result := ans.(type) {
				case terminal.Terminal:
					result.FoundAt = v.RawInterface.Name
					availableReceiver = append(availableReceiver, result)
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
		// removed duplicated receiver via username
		availableReceiver = terminalutil.RemoveStringDuplicateUseMap(availableReceiver)
		slog.Debugf("users: %s", availableReceiver)
		match := false
		var targetTerminal terminal.Terminal
		for _, v := range availableReceiver {
			if v.User.UserName == sendUser {
				targetTerminal = v
				match = true
				break
			}
		}
		if !match {
			slog.Errorf("User %s not found or did not response.", sendUser)
			return
		}
		// tell target that we are going send files!
		slog.Debugf("User found at interface %s", targetTerminal.FoundAt)
		// Start global listener
		replyChan := make(chan related_resp.FileTransfer)
		// Start listening for other requests, at all interfaces.
		for _, v := range processedInterfaces {
			go func(ver netutil.NetInterface) {
				thisTerminal := netutil.GenerateCurrTerminal(terminal.User{
					IP:       ver.CurrentIP.String(),
					UserName: userName,
				})
				msgChan := make(chan interface{}, cfg.CACHED_UDP_READ_CHANNEL_MAX_BUFFER)
				udp := netutil.NewUDPModule(ver)
				go udp.StartUDPListeningWithExtraHandlers(thisTerminal, -time.Second, msgChan, udp.InformTCPHandler)
				for {
					res := <-msgChan
					switch ps := res.(type) {
					case struct{}:
						continue
					case related_resp.FileTransfer:
						replyChan <- ps
						slog.Trace("Allowdd to save file!")
					case error:
						slog.Errorf("Goroutine met an error: %v", res)
						return
					default:
						slog.Trace("GoRoutine Recv: %v", res)
					}
				}
			}(v)
		}
		face, err := netutil.GetNetInterfacesByName(targetTerminal.FoundAt, processedInterfaces)
		if err != nil {
			slog.Error("No interface found!")
			return
		}
		thisTerminal := netutil.GenerateCurrTerminal(terminal.User{
			IP:       face.CurrentIP.String(),
			UserName: userName,
		})
		// Now we send READY_TO_SEND_FILE to client
		udp := netutil.NewUDPModule(*face)
		err = udp.SendP2PUDPPacket(targetTerminal.User.IP, netutil.GenerateReadyToSendFileJSON(thisTerminal, file))
		if err != nil {
			slog.Errorf("Failed to send with READY_TO_SEND_FILE: %v", err)
			return
		}
	a:
		for {
			select {
			case <-replyChan:
				break a
			case <-time.After(cfg.MAX_DEVICE_DISCOVER_TIMEOUT * time.Second):
				slog.Warn("Timeout")
				return
			}
		}
		// we send file now!
		// TODO: ADD SEND
		slog.Debug("SENN")
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
