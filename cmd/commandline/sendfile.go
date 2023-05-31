package commandline

import (
	"TransOwl/internal/cfg"
	"TransOwl/internal/netutil"
	"TransOwl/internal/terminal"
	"TransOwl/pkg/util/terminalutil"
	"github.com/dustin/go-humanize"
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"os"
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
			ScanDevice(netutil.GenerateCurrTerminal(terminal.User{
				IP:       v.CurrentIP.String(),
				UserName: userName,
			}), v, false, scanDeeper, resRefer)
			//We don't need show result we scanned immediately
			//close(resRefer)
			//for ans := range resRefer {
			//	ans = v
			//	availableReceiver = append(availableReceiver, ans)
			//}
		}
		// removed duplicated receiver via username
		availableReceiver = terminalutil.RemoveStringDuplicateUseMap(availableReceiver)
		slog.Debugf("Users:%s", availableReceiver)
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
		}
		// tell target that we are going send files!
		slog.Debugf("User found at interface %s", targetTerminal.User.UserName)
		//udp := netutil.NewUDPModule(targetTerminal.FoundAt)
		//resRefer := make(chan terminal.Terminal, cfg.CACHED_UDP_READ_CHANNEL_MAX_BUFFER)
		//endChan := make(chan struct{})
		//udp.StartUDPListeningWithHandlers()
	},
}

func init() {
	sendFile.Flags().StringVar(&filePath, "sendto", "", "Send to user")
	sendFile.Flags().StringVar(&sendUser, "filepath", "", "file to be sent")
	sendFile.Flags().BoolVarP(&scanDeeper, "deepscan", "d", false, "Scan in 255.255.255.255; If not specified, devices with the same network segment as the NIC are scanned.")
	BaseCmd.AddCommand(sendFile)
}
