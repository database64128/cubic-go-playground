package logging

import (
	"io"
	"log/slog"
	"time"

	"github.com/database64128/tint"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewProductionConsoleZapLogger creates a new [*zap.Logger] with reasonable defaults for production console environments.
//
// See [NewProductionConsoleEncoderConfig] for information on the default encoder configuration.
func NewProductionConsoleZapLogger(ws zapcore.WriteSyncer, level zapcore.Level, noColor, noTime, addCaller bool) *zap.Logger {
	cfg := NewProductionConsoleEncoderConfig(noColor, noTime)
	enc := zapcore.NewConsoleEncoder(cfg)
	core := zapcore.NewCore(enc, zapcore.Lock(ws), level)
	var opts []zap.Option
	if noTime {
		opts = append(opts, zap.WithClock(fakeClock{})) // Note that the sampler requires a real clock.
	}
	if addCaller {
		opts = append(opts, zap.AddCaller())
	}
	return zap.New(core, opts...)
}

// NewProductionConsoleEncoderConfig returns an opinionated [zapcore.EncoderConfig] for production console environments.
func NewProductionConsoleEncoderConfig(noColor, noTime bool) zapcore.EncoderConfig {
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

	return ec
}

// fakeClock is a fake clock that always returns the zero-value time.
//
// fakeClock implements [zapcore.Clock].
type fakeClock struct{}

// Now implements [zapcore.Clock.Now].
func (fakeClock) Now() time.Time {
	return time.Time{}
}

// NewTicker implements [zapcore.Clock.NewTicker].
func (fakeClock) NewTicker(d time.Duration) *time.Ticker {
	return time.NewTicker(d)
}

// NewTintSlogger creates a new [*slog.Logger] with a tint handler.
func NewTintSlogger(w io.Writer, level slog.Level, noColor, noTime bool) *slog.Logger {
	var replaceAttr func(groups []string, attr slog.Attr) slog.Attr
	if noTime {
		replaceAttr = func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 && attr.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return attr
		}
	}

	return slog.New(tint.NewHandler(w, &tint.Options{
		Level:       level,
		ReplaceAttr: replaceAttr,
		NoColor:     noColor,
	}))
}

// NewZerologLogger creates a new [zerolog.Logger].
func NewZerologLogger(w io.Writer, level zerolog.Level, noTime bool) zerolog.Logger {
	logger := zerolog.New(w).Level(level)
	if noTime {
		return logger
	}
	return logger.With().Timestamp().Logger()
}

// NewZerologPrettyLogger creates a new [zerolog.Logger] with a pretty console writer.
func NewZerologPrettyLogger(w io.Writer, level zerolog.Level, noColor, noTime bool) zerolog.Logger {
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:     w,
		NoColor: noColor,
	}).Level(level)
	if noTime {
		return logger
	}
	return logger.With().Timestamp().Logger()
}
