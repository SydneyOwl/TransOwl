package logger

import (
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"os"
	"path/filepath"
)

var (
	info_log_name  = "TransOwl_info.log"
	error_log_name = "TransOwl_error.log"

	logTemplate         = "[{{datetime}}] [{{level}}] {{message}}\n"
	logCodeLineTemplate = "[{{datetime}}] [{{level}}] {{message}} (from:{{caller}})\n"

	DefaultNormalLevels = slog.Levels{slog.NoticeLevel, slog.InfoLevel}
	VerboseLevels       = slog.Levels{slog.NoticeLevel, slog.InfoLevel, slog.DebugLevel}
	VVboseLevels        = slog.Levels{slog.NoticeLevel, slog.InfoLevel, slog.DebugLevel, slog.TraceLevel}
)

func InitLog(verbose bool, vverbose bool, logToFile string) {
	slog.Configure(func(l *slog.SugaredLogger) {
		f := l.Formatter.(*slog.TextFormatter)
		f.TimeFormat = "2006/01/02 15:04:05"
		if verbose || vverbose {
			f.SetTemplate(logCodeLineTemplate)
		} else {
			f.SetTemplate(logTemplate)
		}
		// slog config
	})
	logLevel := DefaultNormalLevels
	if verbose {
		logLevel = VerboseLevels
	}
	if vverbose {
		logLevel = VVboseLevels
	}
	slog.SetLogLevel(logLevel[len(logLevel)-1])
	if logToFile != "" {
		_, err := os.Stat(logToFile)
		if err != nil {
			pwd, err := os.Getwd()
			slog.Warnf("Cannot access directory %s:%v. Use %s as default.", logToFile, err, pwd)
			if err != nil {
				slog.Panicf("Cannot get pwd: %s", err)
			}
			info_log_name = filepath.Join(pwd, info_log_name)
			error_log_name = filepath.Join(pwd, error_log_name)
		} else {
			info_log_name = filepath.Join(logToFile, info_log_name)
			error_log_name = filepath.Join(logToFile, error_log_name)
		}
		h1 := handler.MustFileHandler(info_log_name, handler.WithLogLevels(logLevel))
		h2 := handler.MustFileHandler(error_log_name, handler.WithLogLevels(slog.DangerLevels))
		slog.PushHandler(h1)
		slog.PushHandler(h2)
	}
}
func InitTraceLevelLogs() {
	slog.Configure(func(l *slog.SugaredLogger) {
		f := l.Formatter.(*slog.TextFormatter)
		f.TimeFormat = "2006/01/02 15:04:05"
		f.SetTemplate(logTemplate)
	})
	slog.SetLogLevel(slog.TraceLevel)
}
