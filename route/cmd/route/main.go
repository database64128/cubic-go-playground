package main

import (
	"flag"
	"log/slog"

	"github.com/database64128/cubic-go-playground/logging/tslog"
)

var (
	dumpAll bool
	logCfg  tslog.Config
)

func init() {
	flag.BoolVar(&dumpAll, "dumpAll", false, "dump all routes and interfaces, not just the default routes and active interfaces")
	flag.BoolVar(&logCfg.NoColor, "logNoColor", false, "disable colors in log output")
	flag.BoolVar(&logCfg.NoTime, "logNoTime", false, "disable timestamps in log output")
	flag.TextVar(&logCfg.Format, "logFormat", tslog.FormatDefault, "log `format`, one of: default, slog-text, slog-json, tint-text")
	flag.TextVar(&logCfg.Level, "logLevel", slog.LevelInfo, "log `level`, one of: DEBUG, INFO, WARN, ERROR")
}
