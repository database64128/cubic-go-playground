package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/netip"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	zapConf  = flag.String("zapConf", "", "Preset name or path to JSON configuration file for building the zap logger.\nAvailable presets: console (default), systemd, production, development")
	logLevel = flag.String("logLevel", "", "Override the logger configuration's log level.\nAvailable levels: debug, info, warn, error, dpanic, panic, fatal")
)

func main() {
	flag.Parse()

	var zc zap.Config

	switch *zapConf {
	case "console", "":
		zc = NewProductionConsoleConfig(false)
	case "systemd":
		zc = NewProductionConsoleConfig(true)
	case "production":
		zc = zap.NewProductionConfig()
	case "development":
		zc = zap.NewDevelopmentConfig()
	default:
		f, err := os.Open(*zapConf)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		d := json.NewDecoder(f)
		d.DisallowUnknownFields()
		err = d.Decode(&zc)
		f.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if *logLevel != "" {
		l, err := zapcore.ParseLevel(*logLevel)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		zc.Level.SetLevel(l)
	}

	logger, err := zc.Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer logger.Sync()

	addr := netip.IPv6Unspecified()

	if ce := logger.Check(zap.DebugLevel, "Did it escape to heap?"); ce != nil {
		ce.Write(zap.Stringer("addr", addr))
	}
}

// NewProductionConsoleConfig is a reasonable production logging configuration.
// Logging is enabled at InfoLevel and above.
//
// It uses a console encoder, writes to standard error, and enables sampling.
// Stacktraces are automatically included on logs of ErrorLevel and above.
func NewProductionConsoleConfig(suppressTimestamps bool) zap.Config {
	return zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "console",
		EncoderConfig:    NewProductionConsoleEncoderConfig(suppressTimestamps),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// NewProductionConsoleEncoderConfig returns an opinionated EncoderConfig for
// production console environments.
func NewProductionConsoleEncoderConfig(suppressTimestamps bool) zapcore.EncoderConfig {
	var (
		timeKey    string
		encodeTime zapcore.TimeEncoder
	)

	if !suppressTimestamps {
		timeKey = "T"
		encodeTime = zapcore.ISO8601TimeEncoder
	}

	return zapcore.EncoderConfig{
		TimeKey:          timeKey,
		LevelKey:         "L",
		NameKey:          "N",
		CallerKey:        "C",
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       "M",
		StacktraceKey:    "S",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       encodeTime,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " ",
	}
}
