package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	_defaultLevel      = "info"
	_defaultFormat     = "json"
	_defaultOutputPath = "stdout"
	_defaultComponent  = "default"
)

var _log *zap.Logger

type Config struct {
	// Level defines the minimum enabled logging level
	Level string
	// Format specifies the output format (json or console)
	Format string
	// OutputPath specifies where to write the logs
	OutputPath string
	// Component identifies the component in the logs
	Component string
}

func init() {
	initLogger(Config{
		Level:      _defaultLevel,
		Format:     _defaultFormat,
		OutputPath: _defaultOutputPath,
		Component:  _defaultComponent,
	})
}

func InitLogger(cfg Config) {
	initLogger(cfg)
}

func initLogger(cfg Config) {
	level := zap.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}

	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	var output zapcore.WriteSyncer
	switch cfg.OutputPath {
	case "stdout":
		output = zapcore.AddSync(os.Stdout)
	case "stderr":
		output = zapcore.AddSync(os.Stderr)
	default:
		file, _, err := zap.Open(cfg.OutputPath)
		if err != nil {
			// If file opening fails, fallback to stdout
			output = zapcore.AddSync(os.Stdout)
		} else {
			output = zapcore.AddSync(file)
		}
	}

	core := zapcore.NewCore(encoder, output, zap.NewAtomicLevelAt(level))
	_log = zap.New(core).With(zap.String("component", cfg.Component))
}

func Info(msg string, fields ...zap.Field) {
	_log.Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	_log.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	_log.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	_log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	_log.Fatal(msg, fields...)
}

func With(fields ...zap.Field) *zap.Logger {
	return _log.With(fields...)
}

func Sync() error {
	return _log.Sync()
}

func GetLogger() *zap.Logger {
	return _log
}
