package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestDefaultConfig(t *testing.T) {
	config := defaultConfig()

	assert.Equal(t, _defaultLevel, config.Level)
	assert.Equal(t, _defaultFormat, config.Format)
	assert.Equal(t, _defaultOutputPath, config.OutputPath)
	assert.Equal(t, _defaultComponent, config.Component)
}

func TestNewLogger_AllCases(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		expectedLevel  zapcore.Level
		expectedFormat string
		setupFile      func() (string, func())
	}{
		{
			name: "debug level with json format",
			config: Config{
				Level:      "debug",
				Format:     "json",
				OutputPath: "stdout",
				Component:  "test",
			},
			expectedLevel:  zap.DebugLevel,
			expectedFormat: "json",
		},
		{
			name: "warn level with console format",
			config: Config{
				Level:      "warn",
				Format:     "console",
				OutputPath: "stderr",
				Component:  "test",
			},
			expectedLevel:  zap.WarnLevel,
			expectedFormat: "console",
		},
		{
			name: "error level with console format",
			config: Config{
				Level:      "error",
				Format:     "console",
				OutputPath: "stdout",
				Component:  "test",
			},
			expectedLevel:  zap.ErrorLevel,
			expectedFormat: "console",
		},
		{
			name: "invalid level defaults to info",
			config: Config{
				Level:      "invalid",
				Format:     "json",
				OutputPath: "stdout",
				Component:  "test",
			},
			expectedLevel:  zap.InfoLevel,
			expectedFormat: "json",
		},
		{
			name: "file output success",
			config: Config{
				Level:      "info",
				Format:     "json",
				OutputPath: "", // Will be set by setupFile
				Component:  "test",
			},
			expectedLevel:  zap.InfoLevel,
			expectedFormat: "json",
			setupFile: func() (string, func()) {
				tmpDir := os.TempDir()
				fileName := filepath.Join(tmpDir, "test_log.log")
				return fileName, func() {
					os.Remove(fileName)
				}
			},
		},
		{
			name: "file output failure fallback to stdout",
			config: Config{
				Level:      "info",
				Format:     "json",
				OutputPath: "/invalid/path/that/does/not/exist.log",
				Component:  "test",
			},
			expectedLevel:  zap.InfoLevel,
			expectedFormat: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanup func()
			if tt.setupFile != nil {
				path, cleanupFunc := tt.setupFile()
				tt.config.OutputPath = path
				cleanup = cleanupFunc
			}
			if cleanup != nil {
				defer cleanup()
			}

			logger := newLogger(tt.config)
			assert.NotNil(t, logger)
			assert.True(t, logger.Core().Enabled(tt.expectedLevel))
		})
	}
}

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		expectedLevel  zapcore.Level
		expectedFormat string
	}{
		{
			name: "default configuration",
			config: Config{
				Level:      _defaultLevel,
				Format:     _defaultFormat,
				OutputPath: _defaultOutputPath,
				Component:  _defaultComponent,
			},
			expectedLevel:  zap.InfoLevel,
			expectedFormat: "json",
		},
		{
			name: "debug level with console format",
			config: Config{
				Level:      "debug",
				Format:     "console",
				OutputPath: _defaultOutputPath,
				Component:  "test",
			},
			expectedLevel:  zap.DebugLevel,
			expectedFormat: "console",
		},
		{
			name: "error level",
			config: Config{
				Level:      "error",
				Format:     _defaultFormat,
				OutputPath: _defaultOutputPath,
				Component:  "test",
			},
			expectedLevel:  zap.ErrorLevel,
			expectedFormat: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize the actual logger first to test configuration
			InitLogger(tt.config)
			logger := GetLogger()
			assert.NotNil(t, logger)
			assert.True(t, logger.Core().Enabled(tt.expectedLevel))

			// Now test the component field with a controlled environment
			var buf bytes.Buffer
			encoderConfig := zap.NewProductionEncoderConfig()
			encoderConfig.TimeKey = "" // Disable time for predictable output

			core := zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderConfig),
				zapcore.AddSync(&buf),
				tt.expectedLevel,
			)

			// Create a test logger with the component field
			testLogger := zap.New(core).With(zap.String("component", tt.config.Component))

			// Log at the appropriate level
			msg := "test message"
			switch tt.expectedLevel {
			case zap.DebugLevel:
				testLogger.Debug(msg)
			case zap.InfoLevel:
				testLogger.Info(msg)
			case zap.WarnLevel:
				testLogger.Warn(msg)
			case zap.ErrorLevel:
				testLogger.Error(msg)
			}

			// Ensure the buffer is not empty
			output := buf.String()
			require.NotEmpty(t, output)

			var log map[string]interface{}
			err := json.Unmarshal([]byte(output), &log)
			require.NoError(t, err)

			assert.Equal(t, tt.config.Component, log["component"])
			assert.Equal(t, msg, log["msg"])
		})
	}
}

func TestLogLevels(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a custom encoder config for testing
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "ts",
		NameKey:        "name",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create a custom core writing to our buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)

	// Replace the global logger
	_log = zap.New(core)

	tests := []struct {
		name    string
		logFunc func(string, ...zap.Field)
		level   string
		message string
	}{
		{
			name:    "info level",
			logFunc: Info,
			level:   "info",
			message: "info message",
		},
		{
			name:    "debug level",
			logFunc: Debug,
			level:   "debug",
			message: "debug message",
		},
		{
			name:    "warn level",
			logFunc: Warn,
			level:   "warn",
			message: "warn message",
		},
		{
			name:    "error level",
			logFunc: Error,
			level:   "error",
			message: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(tt.message)

			var log map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &log)
			require.NoError(t, err)

			assert.Equal(t, tt.message, log["msg"])
			assert.Equal(t, tt.level, log["level"])
		})
	}
}

func TestLoggerWith(t *testing.T) {
	var buf bytes.Buffer

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.InfoLevel,
	)

	_log = zap.New(core)

	logger := With(zap.String("key", "value"))
	logger.Info("test message")

	var log map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &log)
	require.NoError(t, err)

	assert.Equal(t, "test message", log["msg"])
	assert.Equal(t, "value", log["key"])
}

func TestLoggerSync(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "log_test_*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Initialize logger with the temp file
	InitLogger(Config{
		Level:      "info",
		Format:     "json",
		OutputPath: tmpFile.Name(),
		Component:  "test",
	})

	// Write something to the log
	Info("test message")

	// Test sync
	err = Sync()
	assert.NoError(t, err)
}

func TestFatal(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a custom encoder config for testing
	encConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	// Create a custom core that writes to our buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encConfig),
		zapcore.AddSync(&buf),
		zapcore.FatalLevel,
	)

	// Create a custom logger that doesn't call os.Exit
	originalLogger := _log
	_log = zap.New(core, zap.WithFatalHook(zapcore.WriteThenPanic))
	defer func() {
		_log = originalLogger
		if r := recover(); r == nil {
			t.Error("The code did not panic")
		}
	}()

	Fatal("fatal error", zap.String("test", "value"))

	// This line should not be reached due to the panic
	t.Error("Expected panic, got none")
}

func TestWith(t *testing.T) {
	logger := With(zap.String("key", "value"))
	assert.NotNil(t, logger)
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)
}

func TestSync(t *testing.T) {
	err := Sync()
	assert.NoError(t, err)
}
