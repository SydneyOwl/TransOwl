package commandline

import (
	"TransOwl/internal/cfg"
	"TransOwl/internal/netutil"
	"TransOwl/internal/terminal"
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"sync"
	"time"
)

var waitRecvCmd = &cobra.Command{
	Use:   "waitrecv",
	Short: "Wait for receiving file",
	Long:  `Listening for ASK_FOR_AVAILABLE_DEVICES`,
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		//var targetInterface netutil.NetInterface
		//var interfaceMutex sync.Mutex
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
				udp.StartUDPListeningWithExtraHandlers(terminal, -time.Second, msgChan)
				// Received request for receiving files!
				wg.Done()
				//slog.Warnf("GoRoutine on Interface %s failed due to %v\n", ver.RawInterface.Name, err)
			}(v)
		}
		wg.Wait()
	},
}

func init() {
	BaseCmd.AddCommand(waitRecvCmd)
}
