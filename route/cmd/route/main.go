package main

import (
	"flag"
	"log/slog"
)

var (
	dumpAll    bool
	logNoColor bool
	logNoTime  bool
	logKVPairs bool
	logJSON    bool
	logLevel   slog.Level
)

func init() {
	flag.BoolVar(&dumpAll, "dumpAll", false, "Dump all routes and interfaces, not just the default routes and active interfaces")
	flag.BoolVar(&logNoColor, "logNoColor", false, "Disable colors in log output")
	flag.BoolVar(&logNoTime, "logNoTime", false, "Disable timestamps in log output")
	flag.BoolVar(&logKVPairs, "logKVPairs", false, "Use key=value pairs in log output")
	flag.BoolVar(&logJSON, "logJSON", false, "Use JSON in log output")
	flag.TextVar(&logLevel, "logLevel", slog.LevelInfo, "Log level, one of: DEBUG, INFO, WARN, ERROR")
}
