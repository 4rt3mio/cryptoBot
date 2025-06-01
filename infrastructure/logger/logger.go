package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type ZapLogger struct {
    sugar *zap.SugaredLogger
}

func NewZapLogger() (*ZapLogger, error) {
    cfg := zap.Config{
        Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
        Development: false,
        Encoding:    "json",
        EncoderConfig: zapcore.EncoderConfig{
            TimeKey:        "time",
            LevelKey:       "level",
            NameKey:        "logger",
            CallerKey:      "caller",
            MessageKey:     "msg",
            StacktraceKey:  "stacktrace",
            LineEnding:     zapcore.DefaultLineEnding,
            EncodeLevel:    zapcore.LowercaseLevelEncoder,
            EncodeTime:     zapcore.ISO8601TimeEncoder,
            EncodeCaller:   zapcore.ShortCallerEncoder,
        },
        OutputPaths:      []string{"stdout"},
        ErrorOutputPaths: []string{"stderr"},
    }
    base, err := cfg.Build()
    if err != nil {
        return nil, err
    }
    sugar := base.Sugar()
    return &ZapLogger{sugar: sugar}, nil
}

func (l *ZapLogger) Debug(msg string, kvs ...interface{}) { l.sugar.Debugw(msg, kvs...) }
func (l *ZapLogger) Info(msg string, kvs ...interface{})  { l.sugar.Infow(msg, kvs...) }
func (l *ZapLogger) Warn(msg string, kvs ...interface{})  { l.sugar.Warnw(msg, kvs...) }
func (l *ZapLogger) Error(msg string, kvs ...interface{}) { l.sugar.Errorw(msg, kvs...) }