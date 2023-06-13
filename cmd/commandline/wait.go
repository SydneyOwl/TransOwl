package commandline

import (
	"bufio"
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"github.com/sydneyowl/TransOwl/internal/cfg"
	"github.com/sydneyowl/TransOwl/internal/netutil"
	"github.com/sydneyowl/TransOwl/internal/terminal"
	"github.com/sydneyowl/TransOwl/internal/terminal/related_resp"
	"github.com/sydneyowl/TransOwl/pkg/util/terminalutil"
	"net"
	"os"
	"path/filepath"
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
						res.FoundAt = ver.RawInterface.Name
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
		var targetIntName related_resp.FileTransfer
	l:
		for {
			select {
			case targetName := <-recvFileTransfer:
				targetIntName = targetName
				break l
			case <-time.After(cfg.MAX_STANDBY_RECEIVE_TIMEOUT * time.Second):
				slog.Warn("No request received within 300s.")
				return
			}
		}
		fileName := targetIntName.File.FileName
		fileSize := targetIntName.File.FileSize
		uu := targetIntName.TransPswd
		targetInterface, _ := netutil.GetNetInterfacesByName(targetIntName.FoundAt, processedInterfaces)
		tcp, err := netutil.NewReceiverTCPModule(*targetInterface)
		defer tcp.ShutListener()
		if err != nil {
			slog.Errorf("Cannot create receiver side tcp conn: %v", err)
			return
		}
		var conn *net.Conn
		timeoutChan := make(chan struct{})
		go func() {
			for {
				conn, err = tcp.BlockTillAcceptWithTimeout(cfg.MAX_TCP_READ_TIMEOUT * time.Second)
				//defer conn.Close()
				if err != nil {
					slog.Errorf("Cannot establish connection: %v. Retrying...", err)
					continue
				}
				_ = (*conn).SetDeadline(time.Now().Add(cfg.MAX_STANDBY_RECEIVE_TIMEOUT * time.Second))
				buf := make([]byte, 2048)
				n, err := (*conn).Read(buf)
				if err != nil {
					slog.Errorf("Failed to read from sender: %v", err)
					continue
				}
				bt, err := netutil.DecodeCustomBinaryO(buf[:n])
				if err != nil {
					slog.Errorf("Failed to verify: %v", err)
					return
				}
				if string(bt) != uu {
					slog.Errorf("Failed to verify sender!")
					_, _ = (*conn).Write([]byte("ERR"))
					return
				}
				break
			}
			timeoutChan <- struct{}{}
		}()
	m:
		for {
			select {
			case <-timeoutChan:
				break m
			case <-time.After(cfg.MAX_TCP_DIAL_TIMEOUT * time.Second):
				slog.Error("Sender side time out!")
				return
			}
		}
		_, err = (*conn).Write([]byte(cfg.ACK_PACKET_RECV))
		if err != nil {
			slog.Errorf("Failed to tell sender: %v", err)
			return
		}
		slog.Info("Verified sender")
		f, err := os.Create(filepath.Join(savePath, fileName))
		if err != nil {
			slog.Errorf("Failed to create file: %#v", err)
			return
		}
		reader := bufio.NewReaderSize(*conn, 409600)
		binChn := make(chan []byte, 20)
		supChn := make(chan struct{})
		go func() {
			for {
				res, ok := <-binChn
				if !ok {
					supChn <- struct{}{}
					return
				}
				_, err = f.Write(res)
				if err != nil {
					slog.Errorf("Cannot write file: %v", err)
					return
				}
			}
		}()
		bar := terminalutil.GenerateBarConfig(int64(fileSize), "Receiving File")
		for {
			_ = (*conn).SetReadDeadline(time.Now().Add(cfg.MAX_TCP_READ_TIMEOUT * time.Second))
			binary, err := netutil.DecodeCustomBinary(reader)
			if err != nil {
				_ = bar.Exit()
				slog.Errorf("Cannot decode:%v", err)
				_, _ = (*conn).Write([]byte(cfg.ACK_ERROR))
				return
			}
			if string(binary) == cfg.ACK_SEND_DONE {
				slog.Infof("Done!")
				close(binChn)
				<-supChn
				_ = f.Close()
				_, _ = (*conn).Write([]byte(cfg.ACK_RECV_DONE))
				return
			}
			_, _ = (*conn).Write([]byte(cfg.ACK_PACKET_RECV))
			binChn <- binary
			_, _ = bar.Write(binary)
		}
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
