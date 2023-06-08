package commandline

import (
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"github.com/sydneyowl/TransOwl/internal/cfg"
	"github.com/sydneyowl/TransOwl/internal/netutil"
	"github.com/sydneyowl/TransOwl/internal/terminal"
	"github.com/sydneyowl/TransOwl/internal/terminal/related_resp"
	"sync"
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
			case <-time.After(cfg.MAX_STANDBY_RECEIVE_TIMEOUT * time.Second):
				slog.Warn("No request received within 300s.")
				return
			}
		}
		//
		//TODO: ADD RECV
		slog.Debug("RECV")
	},
}

var waitScanCmd = &cobra.Command{
	Use:   "waitscan",
	Short: "Wait for respond scan req",
	Long:  `Listen for scan-only`,
	Run: func(cmd *cobra.Command, args []string) {
		wg := sync.WaitGroup{}
		for _, v := range processedInterfaces {
			wg.Add(1)
			go func(ver netutil.NetInterface) {
				terminal := netutil.GenerateCurrTerminal(terminal.User{
					IP:       ver.CurrentIP.String(),
					UserName: userName,
				})
				slog.Infof("Interface %s start listening...", ver.RawInterface.Name)
				msgChan := make(chan interface{}, cfg.CACHED_UDP_READ_CHANNEL_MAX_BUFFER)
				udp := netutil.NewUDPModule(ver)
				defer wg.Done()
				go udp.StartUDPListeningWithDefaultHandlers(terminal, -time.Second, msgChan)
				for {
					ans := <-msgChan
					switch res := ans.(type) {
					case error:
						slog.Errorf("Error occurred: %v", ans)
						return
					default:
						slog.Debugf("Unhandled kind: %v", res)
					}
				}
			}(v)
		}
		wg.Wait()
		slog.Errorf("All go routines exited!!")
	},
}

func init() {
	waitRecvCmd.Flags().StringVar(&savePath, "savepath", "", "file will be saved at path you specified.")
	_ = waitRecvCmd.MarkFlagRequired("savepath")
	BaseCmd.AddCommand(waitScanCmd)
	BaseCmd.AddCommand(waitRecvCmd)
}
