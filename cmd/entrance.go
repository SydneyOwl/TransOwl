package cmd

import (
	"github.com/sydneyowl/TransOwl/cmd/commandline"
	"os"
)

func Execute() {
	if err := commandline.BaseCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
