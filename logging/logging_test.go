package logging

import (
	"log/slog"
	"net/netip"
	"os"
	"testing"

	"github.com/database64128/cubic-go-playground/logging/tslog"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	ip       = netip.AddrFrom16([16]byte{0x20, 0x01, 0x0d, 0xb8, 0xfa, 0xd6, 0x05, 0x72, 0xac, 0xbe, 0x71, 0x43, 0x14, 0xe5, 0x7a, 0x6e})
	addrPort = netip.AddrPortFrom(ip, 1234)
	prefix   = netip.PrefixFrom(ip, 64)
)

func openDevNull(b *testing.B) *os.File {
	b.Helper()
	f, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		b.Skipf("Failed to open /dev/null: %v", err)
	}
	b.Cleanup(func() {
		_ = f.Close()
	})
	return f
}

func BenchmarkZapConsole(b *testing.B) {
	f := openDevNull(b)

	for _, c := range []struct {
		name      string
		noColor   bool
		noTime    bool
		addCaller bool
	}{
		{"Color", false, false, false},
		{"NoColor", true, false, false},
		{"NoTime", false, true, false},
		{"AddCaller", false, false, true},
	} {
		b.Run(c.name, func(b *testing.B) {
			logger := NewProductionConsoleZapLogger(f, zap.InfoLevel, c.noColor, c.noTime, c.addCaller)
			b.Cleanup(func() {
				_ = logger.Sync()
			})

			benchmarkZapLogger(b, logger)
		})
	}
}

func benchmarkZapLogger(b *testing.B, logger *zap.Logger) {
	b.Run("Info", func(b *testing.B) {
		for b.Loop() {
			logger.Info("Hello, world!")
		}
	})

	b.Run("Info/FieldsString", func(b *testing.B) {
		for b.Loop() {
			logger.Info("Hello, world!",
				zap.String("ip", ip.String()),
				zap.String("addrPort", addrPort.String()),
				zap.String("prefix", prefix.String()),
			)
		}
	})

	b.Run("Info/FieldsStringer", func(b *testing.B) {
		for b.Loop() {
			logger.Info("Hello, world!",
				zap.Stringer("ip", ip),
				zap.Stringer("addrPort", addrPort),
				zap.Stringer("prefix", prefix),
			)
		}
	})

	b.Run("Info/FieldsStringerp", func(b *testing.B) {
		for b.Loop() {
			logger.Info("Hello, world!",
				zap.Stringer("ip", &ip),
				zap.Stringer("addrPort", &addrPort),
				zap.Stringer("prefix", &prefix),
			)
		}
	})

	b.Run("Debug", func(b *testing.B) {
		for b.Loop() {
			logger.Debug("Hello, world!")
		}
	})

	b.Run("Debug/FieldsString", func(b *testing.B) {
		for b.Loop() {
			logger.Debug("Hello, world!",
				zap.String("ip", ip.String()),
				zap.String("addrPort", addrPort.String()),
				zap.String("prefix", prefix.String()),
			)
		}
	})

	b.Run("Debug/FieldsStringer", func(b *testing.B) {
		for b.Loop() {
			logger.Debug("Hello, world!",
				zap.Stringer("ip", ip),
				zap.Stringer("addrPort", addrPort),
				zap.Stringer("prefix", prefix),
			)
		}
	})

	b.Run("Debug/FieldsStringerp", func(b *testing.B) {
		for b.Loop() {
			logger.Debug("Hello, world!",
				zap.Stringer("ip", &ip),
				zap.Stringer("addrPort", &addrPort),
				zap.Stringer("prefix", &prefix),
			)
		}
	})

	for _, lvl := range []zapcore.Level{zap.InfoLevel, zap.DebugLevel} {
		b.Run(lvl.String(), func(b *testing.B) {
			b.Run("CheckNoFields", func(b *testing.B) {
				for b.Loop() {
					if ce := logger.Check(lvl, "Hello, world!"); ce != nil {
						ce.Write()
					}
				}
			})

			b.Run("CheckFieldsString", func(b *testing.B) {
				for b.Loop() {
					if ce := logger.Check(lvl, "Hello, world!"); ce != nil {
						ce.Write(
							zap.String("ip", ip.String()),
							zap.String("addrPort", addrPort.String()),
							zap.String("prefix", prefix.String()),
						)
					}
				}
			})

			b.Run("CheckFieldsStringer", func(b *testing.B) {
				for b.Loop() {
					if ce := logger.Check(lvl, "Hello, world!"); ce != nil {
						ce.Write(
							zap.Stringer("ip", ip),
							zap.Stringer("addrPort", addrPort),
							zap.Stringer("prefix", prefix),
						)
					}
				}
			})

			b.Run("CheckFieldsStringerp", func(b *testing.B) {
				for b.Loop() {
					if ce := logger.Check(lvl, "Hello, world!"); ce != nil {
						ce.Write(
							zap.Stringer("ip", &ip),
							zap.Stringer("addrPort", &addrPort),
							zap.Stringer("prefix", &prefix),
						)
					}
				}
			})
		})
	}
}

func BenchmarkTslog(b *testing.B) {
	f := openDevNull(b)

	for _, c := range []struct {
		name    string
		noColor bool
		noTime  bool
		useText bool
		useJSON bool
	}{
		{"Color", false, false, false, false},
		{"NoColor", true, false, false, false},
		{"NoTime", false, true, false, false},
		{"UseText", false, false, true, false},
		{"UseJSON", false, false, false, true},
	} {
		b.Run(c.name, func(b *testing.B) {
			logCfg := tslog.Config{
				Level:          slog.LevelInfo,
				NoColor:        c.noColor,
				NoTime:         c.noTime,
				UseTextHandler: c.useText,
				UseJSONHandler: c.useJSON,
			}
			logger := logCfg.NewLogger(f)

			benchmarkTslogLogger(b, logger)
		})
	}
}

func benchmarkTslogLogger(b *testing.B, logger *tslog.Logger) {
	b.Run("Info", func(b *testing.B) {
		for b.Loop() {
			logger.Info("Hello, world!")
		}
	})

	b.Run("Info/FieldsAny", func(b *testing.B) {
		for b.Loop() {
			logger.Info("Hello, world!",
				slog.Any("ip", ip),
				slog.Any("addrPort", addrPort),
				slog.Any("prefix", prefix),
			)
		}
	})

	b.Run("Info/FieldsAnyp", func(b *testing.B) {
		for b.Loop() {
			logger.Info("Hello, world!",
				slog.Any("ip", &ip),
				slog.Any("addrPort", &addrPort),
				slog.Any("prefix", &prefix),
			)
		}
	})

	b.Run("Info/FieldsString", func(b *testing.B) {
		for b.Loop() {
			logger.Info("Hello, world!",
				tslog.Addr("ip", ip),
				tslog.AddrPort("addrPort", addrPort),
				tslog.Prefix("prefix", prefix),
			)
		}
	})

	b.Run("Info/FieldsMarshalText", func(b *testing.B) {
		for b.Loop() {
			logger.Info("Hello, world!",
				tslog.AddrMarshalText("ip", ip),
				tslog.AddrPortMarshalText("addrPort", addrPort),
				tslog.PrefixMarshalText("prefix", prefix),
			)
		}
	})

	b.Run("Debug", func(b *testing.B) {
		for b.Loop() {
			logger.Debug("Hello, world!")
		}
	})

	b.Run("Debug/FieldsAny", func(b *testing.B) {
		for b.Loop() {
			logger.Debug("Hello, world!",
				slog.Any("ip", ip),
				slog.Any("addrPort", addrPort),
				slog.Any("prefix", prefix),
			)
		}
	})

	b.Run("Debug/FieldsAnyp", func(b *testing.B) {
		for b.Loop() {
			logger.Debug("Hello, world!",
				slog.Any("ip", &ip),
				slog.Any("addrPort", &addrPort),
				slog.Any("prefix", &prefix),
			)
		}
	})

	b.Run("Debug/FieldsString", func(b *testing.B) {
		for b.Loop() {
			logger.Debug("Hello, world!",
				tslog.Addr("ip", ip),
				tslog.AddrPort("addrPort", addrPort),
				tslog.Prefix("prefix", prefix),
			)
		}
	})

	b.Run("Debug/FieldsMarshalText", func(b *testing.B) {
		for b.Loop() {
			logger.Debug("Hello, world!",
				tslog.AddrMarshalText("ip", ip),
				tslog.AddrPortMarshalText("addrPort", addrPort),
				tslog.PrefixMarshalText("prefix", prefix),
			)
		}
	})

	for _, lvl := range []slog.Level{slog.LevelInfo, slog.LevelDebug} {
		b.Run(lvl.String(), func(b *testing.B) {
			b.Run("EnabledFieldsAny", func(b *testing.B) {
				for b.Loop() {
					if logger.Enabled(lvl) {
						logger.Log(lvl, "Hello, world!",
							slog.Any("ip", ip),
							slog.Any("addrPort", addrPort),
							slog.Any("prefix", prefix),
						)
					}
				}
			})

			b.Run("EnabledFieldsAnyp", func(b *testing.B) {
				for b.Loop() {
					if logger.Enabled(lvl) {
						logger.Log(lvl, "Hello, world!",
							slog.Any("ip", &ip),
							slog.Any("addrPort", &addrPort),
							slog.Any("prefix", &prefix),
						)
					}
				}
			})

			b.Run("EnabledFieldsString", func(b *testing.B) {
				for b.Loop() {
					if logger.Enabled(lvl) {
						logger.Log(lvl, "Hello, world!",
							tslog.Addr("ip", ip),
							tslog.AddrPort("addrPort", addrPort),
							tslog.Prefix("prefix", prefix),
						)
					}
				}
			})

			b.Run("EnabledFieldsMarshalText", func(b *testing.B) {
				for b.Loop() {
					if logger.Enabled(lvl) {
						logger.Log(lvl, "Hello, world!",
							tslog.AddrMarshalText("ip", ip),
							tslog.AddrPortMarshalText("addrPort", addrPort),
							tslog.PrefixMarshalText("prefix", prefix),
						)
					}
				}
			})
		})
	}
}

func BenchmarkSlogTint(b *testing.B) {
	f := openDevNull(b)
	logger := NewTintSlogger(f, slog.LevelInfo, false, false)

	benchmarkSlogLogger(b, logger)
}

func benchmarkSlogLogger(b *testing.B, logger *slog.Logger) {
	ctx := b.Context()

	b.Run("Info", func(b *testing.B) {
		for b.Loop() {
			logger.Info("Hello, world!")
		}
	})

	b.Run("Debug", func(b *testing.B) {
		for b.Loop() {
			logger.Debug("Hello, world!")
		}
	})

	for _, lvl := range []slog.Level{slog.LevelInfo, slog.LevelDebug} {
		b.Run(lvl.String(), func(b *testing.B) {
			b.Run("Attr", func(b *testing.B) {
				for b.Loop() {
					logger.LogAttrs(ctx, lvl, "Hello, world!")
				}
			})

			b.Run("AttrFieldsAny", func(b *testing.B) {
				for b.Loop() {
					logger.LogAttrs(ctx, lvl, "Hello, world!",
						slog.Any("ip", ip),
						slog.Any("addrPort", addrPort),
						slog.Any("prefix", prefix),
					)
				}
			})

			b.Run("AttrFieldsAnyp", func(b *testing.B) {
				for b.Loop() {
					logger.LogAttrs(ctx, lvl, "Hello, world!",
						slog.Any("ip", &ip),
						slog.Any("addrPort", &addrPort),
						slog.Any("prefix", &prefix),
					)
				}
			})

			b.Run("AttrFieldsString", func(b *testing.B) {
				for b.Loop() {
					logger.LogAttrs(ctx, lvl, "Hello, world!",
						slog.String("ip", ip.String()),
						slog.String("addrPort", addrPort.String()),
						slog.String("prefix", prefix.String()),
					)
				}
			})

			b.Run("EnabledAttrFieldsAny", func(b *testing.B) {
				for b.Loop() {
					if logger.Enabled(ctx, lvl) {
						logger.LogAttrs(ctx, lvl, "Hello, world!",
							slog.Any("ip", ip),
							slog.Any("addrPort", addrPort),
							slog.Any("prefix", prefix),
						)
					}
				}
			})

			b.Run("EnabledAttrFieldsAnyp", func(b *testing.B) {
				for b.Loop() {
					if logger.Enabled(ctx, lvl) {
						logger.LogAttrs(ctx, lvl, "Hello, world!",
							slog.Any("ip", &ip),
							slog.Any("addrPort", &addrPort),
							slog.Any("prefix", &prefix),
						)
					}
				}
			})

			b.Run("EnabledAttrFieldsString", func(b *testing.B) {
				for b.Loop() {
					if logger.Enabled(ctx, lvl) {
						logger.LogAttrs(ctx, lvl, "Hello, world!",
							slog.String("ip", ip.String()),
							slog.String("addrPort", addrPort.String()),
							slog.String("prefix", prefix.String()),
						)
					}
				}
			})
		})
	}
}

func BenchmarkZerolog(b *testing.B) {
	f := openDevNull(b)
	logger := NewZerologLogger(f, zerolog.InfoLevel, false)

	benchmarkZerologLogger(b, logger)
}

func BenchmarkZerologPretty(b *testing.B) {
	f := openDevNull(b)
	logger := NewZerologPrettyLogger(f, zerolog.InfoLevel, false, false)

	benchmarkZerologLogger(b, logger)
}

func benchmarkZerologLogger(b *testing.B, logger zerolog.Logger) {
	b.Run("Info", func(b *testing.B) {
		for b.Loop() {
			logger.Info().Msg("Hello, world!")
		}
	})

	b.Run("Info/FieldsString", func(b *testing.B) {
		for b.Loop() {
			logger.Info().
				Str("ip", ip.String()).
				Str("addrPort", addrPort.String()).
				Str("prefix", prefix.String()).
				Msg("Hello, world!")
		}
	})

	b.Run("Info/FieldsStringer", func(b *testing.B) {
		for b.Loop() {
			logger.Info().
				Stringer("ip", ip).
				Stringer("addrPort", addrPort).
				Stringer("prefix", prefix).
				Msg("Hello, world!")
		}
	})

	b.Run("Info/FieldsStringerp", func(b *testing.B) {
		for b.Loop() {
			logger.Info().
				Stringer("ip", &ip).
				Stringer("addrPort", &addrPort).
				Stringer("prefix", &prefix).
				Msg("Hello, world!")
		}
	})

	b.Run("Debug", func(b *testing.B) {
		for b.Loop() {
			logger.Debug().Msg("Hello, world!")
		}
	})

	b.Run("Debug/FieldsString", func(b *testing.B) {
		for b.Loop() {
			logger.Debug().
				Str("ip", ip.String()).
				Str("addrPort", addrPort.String()).
				Str("prefix", prefix.String()).
				Msg("Hello, world!")
		}
	})

	b.Run("Debug/FieldsStringer", func(b *testing.B) {
		for b.Loop() {
			logger.Debug().
				Stringer("ip", ip).
				Stringer("addrPort", addrPort).
				Stringer("prefix", prefix).
				Msg("Hello, world!")
		}
	})

	b.Run("Debug/FieldsStringerp", func(b *testing.B) {
		for b.Loop() {
			logger.Debug().
				Stringer("ip", &ip).
				Stringer("addrPort", &addrPort).
				Stringer("prefix", &prefix).
				Msg("Hello, world!")
		}
	})
}
