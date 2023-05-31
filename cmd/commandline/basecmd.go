package commandline

import (
	"TransOwl/config"
	"TransOwl/internal/netutil"
	"TransOwl/pkg/logger"
	"github.com/google/uuid"
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
)

var (
	Verbose             = false
	Vverbose            = false
	interfaceSpecified  []string
	userName            string
	processedInterfaces []netutil.NetInterface
	logToFile           = ""
)

var BaseCmd = &cobra.Command{
	Use:     "TransOwl",
	Short:   "TransOwl",
	Version: config.VERSION,
	Long:    `TransOwl - A simple tool for file transition`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.InitLog(Verbose, Vverbose, logToFile)
		if cmd.Name() == "genmarkdown" {
			return
		}
		if userName == "" && cmd.Name() != "netls" {
			slog.Infof("Username not set!")
			rancode, err := uuid.NewUUID()
			if err != nil {
				slog.Panicf("Cannot generate username: %v", err)
			}
			uu := rancode.String()[0:5]
			userName = "TransOwlUser-" + uu
			slog.Noticef("Now we are using `%s` as your name", userName)
		}
		interfaces, err := netutil.GetBroadcastInterfaces()
		if err != nil {
			slog.Panicf("Failed to fetch net interfaces: %v", err)
		}
		if len(interfaceSpecified) > 0 {
			for _, v := range interfaceSpecified {
				res, err := netutil.GetNetInterfacesByName(v, interfaces)
				if err != nil {
					slog.Infof("Cannot find interface %s. Ignored", v)
					continue
				}
				processedInterfaces = append(processedInterfaces, *res)
			}
		} else {
			processedInterfaces = interfaces
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			slog.Infof("No args found. Start with gui by default.")
		}
	},
}

func init() {
	BaseCmd.PersistentFlags().BoolVar(&Verbose, "verbose", false, "Print Debug Level logs")
	BaseCmd.PersistentFlags().BoolVar(&Vverbose, "vverbose", false, "Print Debug/Trace Level logs")
	BaseCmd.PersistentFlags().StringArrayVarP(&interfaceSpecified, "interface", "i", make([]string, 0), "Specify interface you want to search devices in")
	BaseCmd.PersistentFlags().StringVarP(&userName, "user", "u", "", "Specify a username")
	BaseCmd.PersistentFlags().StringVar(&logToFile, "logtofile", "", "Specify a location logs storage in, default is ./TransOwl_*.log")
	BaseCmd.DisableAutoGenTag = true
}
