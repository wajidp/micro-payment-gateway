package logger

import (
	"fmt"
	"time"

	"github.com/wajidp/micro-payment-gateway/internal/app/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (

	// log encoders
	Json    = "json"
	Console = "console"
)

var log *zap.Logger

// Log is intended as global logger instance pre-initialized by the
// framework
func Log() *zap.Logger {
	if log == nil {
		SetUp()
	}
	return log
}

//SetUp bla
func SetUp() {
	setLog(zap.DebugLevel, "json")
}

var atom = zap.NewAtomicLevel()

var state = zap.String("state", "bootstrapping")

func syslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("_2 Jan 2006 15:04:05.000"))
}
func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}

func setLog(level zapcore.Level, encoding string) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.StacktraceKey = "stack"

	cfg.EncoderConfig.MessageKey = "msg"
	cfg.EncoderConfig.TimeKey = "time"
	cfg.DisableCaller = true
	cfg.EncoderConfig.EncodeTime = syslogTimeEncoder

	cfg.Encoding = encoding
	cfg.Level = atom

	logger, err := cfg.Build()

	if err != nil {
		if log != nil {
			log.With(zap.Error(err)).Warn("New settings not applied.")
		}
		return
	}
	atom.SetLevel(level)

	if log != nil {
		log.Sync()
	}

	log = logger
}

//Debugf ..
func Debugf(format string, a ...interface{}) {
	Log().Debug(fmt.Sprintf(format, a...))
}

//Infof ..
func Infof(format string, a ...interface{}) {
	Log().Info(fmt.Sprintf(format, a...))
}

//SInfof .. structured logging
func SInfof(msg string, fields ...zap.Field) {
	Log().Info(msg, fields...)
}

//SErrorf .. structured logging
func SErrorf(msg string, fields ...zap.Field) {
	Log().Error(msg, fields...)
}

//SWarnf .. structured logging
func SWarnf(msg string, fields ...zap.Field) {
	Log().Warn(msg, fields...)
}

//SDebugf .. structured logging
func SDebugf(msg string, fields ...zap.Field) {
	Log().Debug(msg, fields...)
}

//Errorf ...
func Errorf(format string, a ...interface{}) {
	Log().Error(fmt.Sprintf(format, a...))
}

//Panicf ...
func Panicf(format string, a ...interface{}) {
	Log().Panic(fmt.Sprintf(format, a...))
}

//Warnf ...
func Warnf(format string, a ...interface{}) {
	Log().Warn(fmt.Sprintf(format, a...))
}

// Info logs a message at the Info level.
func Info(msg string, fields ...zap.Field) {
	Log().Info(msg, fields...)
}
func IsLogEncodingJSON() bool {
	return config.AppConfig.LogEncoding == Json
}
