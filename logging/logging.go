package logging

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewProductionConsoleConfig builds a reasonable production logging configuration.
// Logging is enabled at InfoLevel and above, and uses a console encoder.
// Logs are written to standard error.
// Stacktraces are included on logs of ErrorLevel and above.
// DPanicLevel logs will not panic, but will write a stacktrace.
//
// Sampling is enabled at 100:100 by default, meaning that after the first 100 log entries
// with the same level and message in the same second, it will log every 100th entry
// with the same level and message in the same second.
//
// See [NewProductionConsoleEncoderConfig] for information on the default encoder configuration.
func NewProductionConsoleConfig(noColor, noTime, noCaller bool) zap.Config {
	return zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "console",
		EncoderConfig:    NewProductionConsoleEncoderConfig(noColor, noTime, noCaller),
		OutputPaths:      []string{os.DevNull},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// NewProductionConsoleEncoderConfig returns an opinionated EncoderConfig for
// production console environments.
func NewProductionConsoleEncoderConfig(noColor, noTime, noCaller bool) zapcore.EncoderConfig {
	ec := zapcore.EncoderConfig{
		TimeKey:          "T",
		LevelKey:         "L",
		NameKey:          "N",
		CallerKey:        "C",
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       "M",
		StacktraceKey:    "S",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalColorLevelEncoder,
		EncodeTime:       zapcore.ISO8601TimeEncoder,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " ",
	}

	if noColor {
		ec.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	if noTime {
		ec.TimeKey = zapcore.OmitKey
		ec.EncodeTime = nil
	}

	if noCaller {
		ec.CallerKey = zapcore.OmitKey
		ec.EncodeCaller = nil
	}

	return ec
}

// NewTintSlogger creates a new [*slog.Logger] with a tint handler.
func NewTintSlogger(level slog.Level, noColor, noTime bool) (*slog.Logger, func() error, error) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		return nil, nil, err
	}

	var replaceAttr func(groups []string, attr slog.Attr) slog.Attr
	if noTime {
		replaceAttr = func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 && attr.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return attr
		}
	}

	return slog.New(tint.NewHandler(f, &tint.Options{
		Level:       level,
		ReplaceAttr: replaceAttr,
		NoColor:     noColor,
	})), f.Close, nil
}
