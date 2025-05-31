package logger

import (
	"fmt"

	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	_logger, err := zap.NewProduction(
		zap.AddStacktrace(zap.PanicLevel),
		zap.AddCallerSkip(1),
	)
	if err != nil {
		panic(err)
	}
	logger = _logger
}

func Fatal(msg string) {
	logger.Fatal(msg)
}

func Fatalf(msg string, args ...any) {
	logger.Fatal(fmt.Sprintf(msg, args...))
}
func FatalKV(msg string, kv ...any) {
	logger.Fatal(msg, parseKV(kv...)...)
}

func Panic(msg string) {
	logger.Panic(msg)
}

func Panicf(msg string, args ...any) {
	logger.Panic(fmt.Sprintf(msg, args...))
}

func PanicKV(msg string, kv ...any) {
	logger.Panic(msg, parseKV(kv...)...)
}

func Error(msg string) {
	logger.Error(msg)
}

func Errorf(msg string, args ...any) {
	logger.Error(fmt.Sprintf(msg, args...))
}

func ErrorKV(msg string, kv ...any) {
	logger.Error(msg, parseKV(kv...)...)
}

func Warn(msg string) {
	logger.Warn(msg)
}

func Warnf(msg string, args ...any) {
	logger.Warn(fmt.Sprintf(msg, args...))
}

func WarnKV(msg string, kv ...any) {
	logger.Warn(msg, parseKV(kv...)...)
}

func Info(msg string) {
	logger.Info(msg)
}

func Infof(msg string, args ...any) {
	logger.Info(fmt.Sprintf(msg, args...))
}

func InfoKV(msg string, kv ...any) {
	logger.Info(msg, parseKV(kv...)...)
}

func Debug(msg string) {
	logger.Debug(msg)
}

func Debugf(msg string, args ...any) {
	logger.Debug(fmt.Sprintf(msg, args...))
}

func DebufKV(msg string, kv ...any) {
	logger.Debug(msg, parseKV(kv...)...)
}

func parseKV(kv ...any) []zap.Field {
	if len(kv)%2 != 0 {
		Panic("kv must be pairs")
	}
	kvs := len(kv) / 2
	fields := make([]zap.Field, 0, kvs)
	for i := 0; i < kvs; i += 2 {
		k, ok := kv[i].(string)
		if !ok {
			Panic("kv key must be string")
		}
		fields = append(fields, zap.Any(k, kv[i+1]))
	}
	return fields
}
