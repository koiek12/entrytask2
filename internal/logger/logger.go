package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Instance *zap.Logger
var once sync.Once

func Init(path, levelStr string) {
	once.Do(func() {
		var level zapcore.Level
		level.UnmarshalText([]byte(levelStr))
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   path,
			MaxSize:    500,
			MaxBackups: 3,
			MaxAge:     28,
		})
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			w,
			level,
		)
		Instance = zap.New(core, zap.AddStacktrace(zap.ErrorLevel))
	})
}
