package commandline

import (
	"fmt"
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"github.com/sydneyowl/TransOwl/internal/netutil"
)

var lsNetCmd = &cobra.Command{
	Use:   "netls",
	Short: "List net available",
	Long:  `List net which is "UP" and "BROADCAST" but not "LOOPBACK"`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Getting interfaces")
		interfaces, err := netutil.GetBroadcastInterfaces()
		if err != nil {
			slog.Fatalf("Failed to fetch net interfaces: %v", err)
			return
		}
		for i, v := range interfaces {
			fmt.Printf("Interface %d, Name:%s, MAC:%s, ip:%s, MTU:%d\n", i, v.RawInterface.Name, v.RawInterface.HardwareAddr, v.CurrentIP, v.RawInterface.MTU)
		}
		fmt.Println("You may choose one of the interfaces.")
	},
}

func init() {
	BaseCmd.AddCommand(lsNetCmd)
}
