package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// log defaults to a Nop logger so packages can use the global safely without
// requiring callers (and tests) to call Init first. Init replaces it.
var log = zap.NewNop().Sugar()

func Init(service string, isDev bool) {
	var config zap.Config

	if isDev {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	logger, _ := config.Build(zap.WithCaller(true), zap.AddStacktrace(zapcore.FatalLevel))
	log = logger.Sugar().With("service", service)
}

func Info(msg string, keysAndValues ...interface{}) {
	log.Infow(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...interface{}) {
	log.Errorw(msg, keysAndValues...)
}

func Debug(msg string, keysAndValues ...interface{}) {
	log.Debugw(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...interface{}) {
	log.Warnw(msg, keysAndValues...)
}

func Fatal(msg string, keysAndValues ...interface{}) {
	log.Fatalw(msg, keysAndValues...)
}

func WithFields(keysAndValues ...interface{}) *zap.SugaredLogger {
	return log.With(keysAndValues...)
}
