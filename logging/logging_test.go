package logging

import (
	"context"
	"log/slog"
	"net/netip"
	"testing"

	"github.com/database64128/cubic-go-playground/logging/tslog"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	ip       = netip.IPv6LinkLocalAllNodes()
	addrPort = netip.AddrPortFrom(ip, 1234)
)

func BenchmarkZapConsole(b *testing.B) {
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
			logger, close, err := NewProductionConsoleZapLogger(c.noColor, c.noTime, c.addCaller, zap.InfoLevel)
			if err != nil {
				b.Fatalf("Failed to build logger: %v", err)
			}
			b.Cleanup(func() {
				_ = logger.Sync()
				_ = close()
			})

			benchmarkZapLogger(b, logger)
		})
	}
}

func benchmarkZapLogger(b *testing.B, logger *zap.Logger) {
	b.Run("Info", func(b *testing.B) {
		for range b.N {
			logger.Info("Hello, world!")
		}
	})

	b.Run("Info/FieldsString", func(b *testing.B) {
		for range b.N {
			logger.Info("Hello, world!",
				zap.String("ip", ip.String()),
				zap.String("addrPort", addrPort.String()),
			)
		}
	})

	b.Run("Info/FieldsStringer", func(b *testing.B) {
		for range b.N {
			logger.Info("Hello, world!",
				zap.Stringer("ip", ip),
				zap.Stringer("addrPort", addrPort),
			)
		}
	})

	b.Run("Info/FieldsStringerp", func(b *testing.B) {
		for range b.N {
			logger.Info("Hello, world!",
				zap.Stringer("ip", &ip),
				zap.Stringer("addrPort", &addrPort),
			)
		}
	})

	b.Run("Debug", func(b *testing.B) {
		for range b.N {
			logger.Debug("Hello, world!")
		}
	})

	b.Run("Debug/FieldsString", func(b *testing.B) {
		for range b.N {
			logger.Debug("Hello, world!",
				zap.String("ip", ip.String()),
				zap.String("addrPort", addrPort.String()),
			)
		}
	})

	b.Run("Debug/FieldsStringer", func(b *testing.B) {
		for range b.N {
			logger.Debug("Hello, world!",
				zap.Stringer("ip", ip),
				zap.Stringer("addrPort", addrPort),
			)
		}
	})

	b.Run("Debug/FieldsStringerp", func(b *testing.B) {
		for range b.N {
			logger.Debug("Hello, world!",
				zap.Stringer("ip", &ip),
				zap.Stringer("addrPort", &addrPort),
			)
		}
	})

	for _, lvl := range []zapcore.Level{zap.InfoLevel, zap.DebugLevel} {
		b.Run(lvl.String(), func(b *testing.B) {
			b.Run("CheckNoFields", func(b *testing.B) {
				for range b.N {
					if ce := logger.Check(lvl, "Hello, world!"); ce != nil {
						ce.Write()
					}
				}
			})

			b.Run("CheckFieldsString", func(b *testing.B) {
				for range b.N {
					if ce := logger.Check(lvl, "Hello, world!"); ce != nil {
						ce.Write(
							zap.String("ip", ip.String()),
							zap.String("addrPort", addrPort.String()),
						)
					}
				}
			})

			b.Run("CheckFieldsStringer", func(b *testing.B) {
				for range b.N {
					if ce := logger.Check(lvl, "Hello, world!"); ce != nil {
						ce.Write(
							zap.Stringer("ip", ip),
							zap.Stringer("addrPort", addrPort),
						)
					}
				}
			})

			b.Run("CheckFieldsStringerp", func(b *testing.B) {
				for range b.N {
					if ce := logger.Check(lvl, "Hello, world!"); ce != nil {
						ce.Write(
							zap.Stringer("ip", &ip),
							zap.Stringer("addrPort", &addrPort),
						)
					}
				}
			})
		})
	}
}

func BenchmarkTslog(b *testing.B) {
	for _, c := range []struct {
		name    string
		noColor bool
		noTime  bool
	}{
		{"Color", false, false},
		{"NoColor", true, false},
		{"NoTime", false, true},
	} {
		b.Run(c.name, func(b *testing.B) {
			logger, close, err := tslog.New(slog.LevelInfo, c.noColor, c.noTime)
			if err != nil {
				b.Fatalf("Failed to create logger: %v", err)
			}
			defer close()

			benchmarkTslogLogger(b, logger)
		})
	}
}

func benchmarkTslogLogger(b *testing.B, logger *tslog.Logger) {
	b.Run("Info", func(b *testing.B) {
		for range b.N {
			logger.Info("Hello, world!")
		}
	})

	b.Run("Info/FieldsAny", func(b *testing.B) {
		for range b.N {
			logger.Info("Hello, world!",
				slog.Any("ip", ip),
				slog.Any("addrPort", addrPort),
			)
		}
	})

	b.Run("Info/FieldsAnyp", func(b *testing.B) {
		for range b.N {
			logger.Info("Hello, world!",
				slog.Any("ip", &ip),
				slog.Any("addrPort", &addrPort),
			)
		}
	})

	b.Run("Info/FieldsString", func(b *testing.B) {
		for range b.N {
			logger.Info("Hello, world!",
				slog.String("ip", ip.String()),
				slog.String("addrPort", addrPort.String()),
			)
		}
	})

	b.Run("Debug", func(b *testing.B) {
		for range b.N {
			logger.Debug("Hello, world!")
		}
	})

	b.Run("Debug/FieldsAny", func(b *testing.B) {
		for range b.N {
			logger.Debug("Hello, world!",
				slog.Any("ip", ip),
				slog.Any("addrPort", addrPort),
			)
		}
	})

	b.Run("Debug/FieldsAnyp", func(b *testing.B) {
		for range b.N {
			logger.Debug("Hello, world!",
				slog.Any("ip", &ip),
				slog.Any("addrPort", &addrPort),
			)
		}
	})

	b.Run("Debug/FieldsString", func(b *testing.B) {
		for range b.N {
			logger.Debug("Hello, world!",
				slog.String("ip", ip.String()),
				slog.String("addrPort", addrPort.String()),
			)
		}
	})

	for _, lvl := range []slog.Level{slog.LevelInfo, slog.LevelDebug} {
		b.Run(lvl.String(), func(b *testing.B) {
			b.Run("EnabledFieldsAny", func(b *testing.B) {
				for range b.N {
					if logger.Enabled(lvl) {
						logger.Log(lvl, "Hello, world!",
							slog.Any("ip", ip),
							slog.Any("addrPort", addrPort),
						)
					}
				}
			})

			b.Run("EnabledFieldsAnyp", func(b *testing.B) {
				for range b.N {
					if logger.Enabled(lvl) {
						logger.Log(lvl, "Hello, world!",
							slog.Any("ip", &ip),
							slog.Any("addrPort", &addrPort),
						)
					}
				}
			})

			b.Run("EnabledFieldsString", func(b *testing.B) {
				for range b.N {
					if logger.Enabled(lvl) {
						logger.Log(lvl, "Hello, world!",
							slog.String("ip", ip.String()),
							slog.String("addrPort", addrPort.String()),
						)
					}
				}
			})
		})
	}
}

func BenchmarkSlogTint(b *testing.B) {
	logger, close, err := NewTintSlogger(slog.LevelInfo, false, false)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer close()

	benchmarkSlogLogger(b, logger)
}

func benchmarkSlogLogger(b *testing.B, logger *slog.Logger) {
	ctx := context.Background()

	b.Run("Info", func(b *testing.B) {
		for range b.N {
			logger.Info("Hello, world!")
		}
	})

	b.Run("Debug", func(b *testing.B) {
		for range b.N {
			logger.Debug("Hello, world!")
		}
	})

	for _, lvl := range []slog.Level{slog.LevelInfo, slog.LevelDebug} {
		b.Run(lvl.String(), func(b *testing.B) {
			b.Run("Attr", func(b *testing.B) {
				for range b.N {
					logger.LogAttrs(ctx, lvl, "Hello, world!")
				}
			})

			b.Run("AttrFieldsAny", func(b *testing.B) {
				for range b.N {
					logger.LogAttrs(ctx, lvl, "Hello, world!",
						slog.Any("ip", ip),
						slog.Any("addrPort", addrPort),
					)
				}
			})

			b.Run("AttrFieldsAnyp", func(b *testing.B) {
				for range b.N {
					logger.LogAttrs(ctx, lvl, "Hello, world!",
						slog.Any("ip", &ip),
						slog.Any("addrPort", &addrPort),
					)
				}
			})

			b.Run("AttrFieldsString", func(b *testing.B) {
				for range b.N {
					logger.LogAttrs(ctx, lvl, "Hello, world!",
						slog.String("ip", ip.String()),
						slog.String("addrPort", addrPort.String()),
					)
				}
			})

			b.Run("EnabledAttrFieldsAny", func(b *testing.B) {
				for range b.N {
					if logger.Enabled(ctx, lvl) {
						logger.LogAttrs(ctx, lvl, "Hello, world!",
							slog.Any("ip", ip),
							slog.Any("addrPort", addrPort),
						)
					}
				}
			})

			b.Run("EnabledAttrFieldsAnyp", func(b *testing.B) {
				for range b.N {
					if logger.Enabled(ctx, lvl) {
						logger.LogAttrs(ctx, lvl, "Hello, world!",
							slog.Any("ip", &ip),
							slog.Any("addrPort", &addrPort),
						)
					}
				}
			})

			b.Run("EnabledAttrFieldsString", func(b *testing.B) {
				for range b.N {
					if logger.Enabled(ctx, lvl) {
						logger.LogAttrs(ctx, lvl, "Hello, world!",
							slog.String("ip", ip.String()),
							slog.String("addrPort", addrPort.String()),
						)
					}
				}
			})
		})
	}
}

func BenchmarkZerolog(b *testing.B) {
	logger, close, err := NewZerologLogger(zerolog.InfoLevel, false)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer close()

	benchmarkZerologLogger(b, logger)
}

func BenchmarkZerologPretty(b *testing.B) {
	logger, close, err := NewZerologPrettyLogger(zerolog.InfoLevel, false, false)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer close()

	benchmarkZerologLogger(b, logger)
}

func benchmarkZerologLogger(b *testing.B, logger zerolog.Logger) {
	b.Run("Info", func(b *testing.B) {
		for range b.N {
			logger.Info().Msg("Hello, world!")
		}
	})

	b.Run("Info/FieldsString", func(b *testing.B) {
		for range b.N {
			logger.Info().
				Str("ip", ip.String()).
				Str("addrPort", addrPort.String()).
				Msg("Hello, world!")
		}
	})

	b.Run("Info/FieldsStringer", func(b *testing.B) {
		for range b.N {
			logger.Info().
				Stringer("ip", ip).
				Stringer("addrPort", addrPort).
				Msg("Hello, world!")
		}
	})

	b.Run("Info/FieldsStringerp", func(b *testing.B) {
		for range b.N {
			logger.Info().
				Stringer("ip", &ip).
				Stringer("addrPort", &addrPort).
				Msg("Hello, world!")
		}
	})

	b.Run("Debug", func(b *testing.B) {
		for range b.N {
			logger.Debug().Msg("Hello, world!")
		}
	})

	b.Run("Debug/FieldsString", func(b *testing.B) {
		for range b.N {
			logger.Debug().
				Str("ip", ip.String()).
				Str("addrPort", addrPort.String()).
				Msg("Hello, world!")
		}
	})

	b.Run("Debug/FieldsStringer", func(b *testing.B) {
		for range b.N {
			logger.Debug().
				Stringer("ip", ip).
				Stringer("addrPort", addrPort).
				Msg("Hello, world!")
		}
	})

	b.Run("Debug/FieldsStringerp", func(b *testing.B) {
		for range b.N {
			logger.Debug().
				Stringer("ip", &ip).
				Stringer("addrPort", &addrPort).
				Msg("Hello, world!")
		}
	})
}
