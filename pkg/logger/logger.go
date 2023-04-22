package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

const (
	Console = "console"
	File    = "file"
)

var (
	// Logger 性能更好但是对使用者不方便，每次需要使用 zap.xxx 传入类型
	Logger *zap.Logger
	// Sugar 性能稍差但是可以不用指定传入类型
	Sugar *zap.SugaredLogger
)

// 编码器（如何写入日志）
func logEncoder() zapcore.Encoder {
	timeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

// 指定将日志写到哪里去
func logWriterSyncer() zapcore.WriteSyncer {
	// 切割归档日志文件
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./log/im.log",
		MaxSize:    1024, // 日志文件的最大大小（MB）
		MaxAge:     7,    // 保留旧文件的最大天数
		MaxBackups: 10,   // 保留旧文档的最大个数
		LocalTime:  false,
		Compress:   false, // 是否压缩旧文件
	})
}

func InitLogger(target string, level zapcore.Level) {
	w := logWriterSyncer()
	var writeSyncer zapcore.WriteSyncer
	// 打印在控制台
	if target == Console {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else if target == File {
		writeSyncer = zapcore.NewMultiWriteSyncer(w)
	}

	core := zapcore.NewCore(
		logEncoder(), // 怎么写
		writeSyncer,  // 写到哪
		level,        // 日志级别
	)

	Logger = zap.New(core, zap.AddCaller()) // 打印调用方信息
	Sugar = Logger.Sugar()
}
