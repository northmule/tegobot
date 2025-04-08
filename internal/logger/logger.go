package logger

import (
	"encoding/json"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger логгер
type Logger struct {
	*zap.SugaredLogger
}

// NewLogger конструктор
func NewLogger(path string, level string) (*Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{path}
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	cfg.Level = lvl
	cfg.Encoding = "console"
	appLogger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	logger := &Logger{appLogger.Sugar()}

	return logger, nil
}

func (l *Logger) WriteToJson(message string, logData interface{}) {
	message = message + ": "
	jsonData, err := json.Marshal(logData)
	if err != nil {
		l.Error(message, err)
		return
	}

	switch l.Level() {
	case zapcore.DebugLevel:
		l.Debug(message, string(jsonData))
	case zapcore.InfoLevel:
		l.Info(message, string(jsonData))
	case zapcore.WarnLevel:
		l.Warn(message, string(jsonData))
	case zapcore.ErrorLevel:
		l.Error(message, string(jsonData))
	case zapcore.PanicLevel:
		l.Panic(message, string(jsonData))
	case zapcore.FatalLevel:
		l.Fatal(message, string(jsonData))
	default:
		l.Debug(message, string(jsonData))
	}
}
