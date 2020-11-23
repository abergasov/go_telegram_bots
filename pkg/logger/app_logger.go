package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var aLogger *zap.Logger

func NewLogger() {
	var eI error
	aLogger, eI = zap.NewDevelopment()
	if eI != nil {
		log.Panicf("error init logger, %s", eI.Error())
	}
}

func Info(message string, args ...zapcore.Field) {
	aLogger.Info(message, args...)
}

func Warning(message string, args ...zapcore.Field) {
	aLogger.Warn(message, args...)
}

func Error(message string, err error) {
	aLogger.Error(message, zap.Error(err))
}

func Fatal(message string, err error, args ...zapcore.Field) {
	if len(args) == 0 {
		aLogger.Fatal(message, zap.Error(err))
		return
	}
	params := make([]zapcore.Field, len(args)+1)
	params[0] = zap.Error(err)
	params = append(params, args...)

	aLogger.Fatal(message, params...)
}
