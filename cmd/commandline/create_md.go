package commandline

import (
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"os"
	"path/filepath"
)

var filepos string
var createMarkdown = &cobra.Command{
	Use:   "genmarkdown",
	Short: "Generate Instruction",
	Long:  `create markdown at location specified`,
	Run: func(cmd *cobra.Command, args []string) {
		pwd, err := os.Getwd()
		if err != nil {
			slog.Panicf("Cannot get pwd: %s", err)
		}
		if filepos == "" {
			slog.Warnf("No pos specified. Use %v as default.", pwd)
			filepos = pwd
		} else {
			_, err := os.Stat(filepos)
			if err != nil {
				slog.Warnf("Cannot access directory %s: %v. Use %s as default.", logToFile, err, pwd)
				filepos = pwd
			}
		}
		err = doc.GenMarkdownTree(BaseCmd, filepos)
		if err != nil {
			slog.Warnf("Cannot create markdown at here: %v", err)
			return
		}
		files, err := filepath.Glob(filepath.Join(filepos, "/TransOwl_completion*.md"))
		if err != nil {
			slog.Warnf("Cannot delete files： %v", err)
			return
		}
		for _, f := range files {
			if err := os.RemoveAll(f); err != nil {
				slog.Warnf("Cannot remove: %v", err)
			}
		}
	},
}

func init() {
	BaseCmd.AddCommand(createMarkdown)
	createMarkdown.Flags().StringVar(&filepos, "mdpath", "", "Create markdown at specified location")
}
