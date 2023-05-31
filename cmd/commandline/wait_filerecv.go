package commandline

import (
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"github.com/sydneyowl/TransOwl/internal/cfg"
	"github.com/sydneyowl/TransOwl/internal/netutil"
	"github.com/sydneyowl/TransOwl/internal/terminal"
	"github.com/sydneyowl/TransOwl/internal/terminal/related_resp"
	"time"
)

var savePath string

var waitRecvCmd = &cobra.Command{
	Use:   "waitrecv",
	Short: "Wait for receiving file",
	Long:  `Listening for ASK_FOR_AVAILABLE_DEVICES`,
	Run: func(cmd *cobra.Command, args []string) {
		if savePath == "" {
			slog.Errorf("Specify filepath you want to save file at using --savepath")
			return
		}
		//var targetInterface netutil.NetInterface
		//var interfaceMutex sync.Mutex
		recvFileTransfer := make(chan related_resp.FileTransfer)
		for _, v := range processedInterfaces {
			go func(ver netutil.NetInterface) {
				terminal := netutil.GenerateCurrTerminal(terminal.User{
					IP:       ver.CurrentIP.String(),
					UserName: userName,
				})
				slog.Infof("Interface %s start listening...", ver.RawInterface.Name)
				msgChan := make(chan interface{}, cfg.CACHED_UDP_READ_CHANNEL_MAX_BUFFER)
				udp := netutil.NewUDPModule(ver)
				go udp.StartUDPListeningWithExtraHandlers(terminal, -time.Second, msgChan, udp.ReplyReadyToReceiveFileHandler)
				for {
					ans := <-msgChan
					switch res := ans.(type) {
					case related_resp.FileTransfer:
						slog.Infof("Ready to receive file: %s, size: %s, from: %s", res.File.FileName, res.File.GetHumanReadableFileSize(), res.User.UserName)
						recvFileTransfer <- res
					case error:
						slog.Errorf("Error occurred: %v", ans)
						return
					default:
						slog.Debugf("Unhandled kind: %v", res)
					}
				}
			}(v)
		}
	l:
		for {
			select {
			case <-recvFileTransfer:
				break l
			case <-time.After(cfg.MAX_DEVICE_DISCOVER_TIMEOUT * time.Second):
				slog.Warn("Timeout")
				return
			}
		}
		//
		//TODO: ADD RECV
		slog.Debug("RECV")
	},
}

func init() {
	waitRecvCmd.Flags().StringVar(&savePath, "savepath", "", "file will be saved at path you specified.")
	_ = waitRecvCmd.MarkFlagRequired("savepath")
	// Not available right now!
	BaseCmd.AddCommand(waitRecvCmd)
}
