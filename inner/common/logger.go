package common

import (
	"context"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger структура логгера
type Logger struct {
	*zap.Logger
}

// NewLogger функция-конструктор логгера
func NewLogger(cfg Config) *Logger {
	var zapEncoderCfg = zapcore.EncoderConfig{
		TimeKey:          "timestamp",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.LowercaseLevelEncoder,
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000000"),
		EncodeDuration:   zapcore.MillisDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: "  ",
	}
	var zapCfg = zap.Config{
		Level:       zap.NewAtomicLevelAt(parseLogLevel(cfg.LogLevel)),
		Development: cfg.LogDevelopMode,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		// пишем записи в формате JSON
		Encoding:      "json",
		EncoderConfig: zapEncoderCfg,
		// логируем сообщения и ошибки в консоль
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
	}
	var logger = zap.Must(zapCfg.Build())
	logger.Info("logger construction succeeded")
	var created = &Logger{logger}
	created.setNewFiberZapLogger()
	return created
}

// setNewFiberZapLogger устанавливает логгер для fiber
func (l *Logger) setNewFiberZapLogger() {
	var fiberzapLogger = fiberzap.NewLogger(fiberzap.LoggerConfig{
		SetLogger: l.Logger,
	})
	log.SetLogger(fiberzapLogger)
}

// parseLogLevel парсит уровень логирования из строки в zapcore.Level
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug", "DEBUG":
		return zapcore.DebugLevel
	case "info", "INFO":
		return zapcore.InfoLevel
	case "warn", "WARN":
		return zapcore.WarnLevel
	case "error", "ERROR":
		return zapcore.ErrorLevel
	case "panic", "PANIC":
		return zapcore.PanicLevel
	case "fatal", "FATAL":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// NewTestLogger создает логгер для тестов
func NewTestLogger() *Logger {
	var zapEncoderCfg = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000000"),
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Конфигурация для тестов - минимальное логирование
	var zapCfg = zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.ErrorLevel), // Только ошибки в тестах
		Development:      true,
		Encoding:         "console", // Простой формат для тестов
		EncoderConfig:    zapEncoderCfg,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
	}

	var logger = zap.Must(zapCfg.Build())
	return &Logger{logger}
}

// ключ для получения requestId из контекста
var ridKey = requestid.ConfigDefault.ContextKey.(string)

// функция логирования с добавлением requestId
func (l *Logger) DebugCtx(
	ctx context.Context,
	msg string,
	fields ...zap.Field,
) {
	var rid string
	if v := ctx.Value(ridKey); v != nil {
		rid = v.(string)
	}
	fields = append(fields, zap.String(ridKey, rid))
	l.Debug(msg, fields...)
}

func (l *Logger) InfoCtx(
	ctx context.Context,
	msg string,
	fields ...zap.Field,
) {
	var rid string
	if v := ctx.Value(ridKey); v != nil {
		rid = v.(string)
	}
	fields = append(fields, zap.String(ridKey, rid))
	l.Info(msg, fields...)
}

func (l *Logger) ErrorCtx(
	ctx context.Context,
	msg string,
	fields ...zap.Field,
) {
	var rid string
	if v := ctx.Value(ridKey); v != nil {
		rid = v.(string)
	}
	fields = append(fields, zap.String(ridKey, rid))
	l.Error(msg, fields...)
}
