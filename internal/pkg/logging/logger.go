package logging

import (
	"fmt"
	"os"
	"path"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/utils"
)

var FileRotateLogs = new(fileRotateLogs)

type fileRotateLogs struct{}

// GetWriteSyncer Get zapcore.WriteSyncer | 获取 zapcore.WriteSyncer
func (r *fileRotateLogs) GetWriteSyncer(level string) (zapcore.WriteSyncer, error) {
	fileWriter, err := rotatelogs.New(
		path.Join("logs", "%Y-%m-%d", level+".log"),
		rotatelogs.WithClock(rotatelogs.Local),
		rotatelogs.WithMaxAge(time.Duration(30)*24*time.Hour), // Log retention time | 日志留存时间
		rotatelogs.WithRotationTime(time.Hour*24),
	)
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
}

type _zap struct{}

// GetEncoder Get zapcore.Encoder | 获取 zapcore.Encoder
func (z *_zap) GetEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(z.GetEncoderConfig())
}

// GetEncoderConfig Get zapcore.EncoderConfig | 获取zapcore.EncoderConfig
func (z *_zap) GetEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     z.CustomTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// GetEncoderCore Get Encoder's zapcore.Core | 获取Encoder的 zapcore.Core
func (z *_zap) GetEncoderCore(l zapcore.Level, level zap.LevelEnablerFunc) zapcore.Core {
	writer, err := FileRotateLogs.GetWriteSyncer(l.String()) // Use file-rotatelogs for log splitting | 使用file-rotatelogs进行日志分割
	if err != nil {
		fmt.Printf("Get Write Syncer Failed err:%v", err.Error())
		return nil
	}

	return zapcore.NewCore(z.GetEncoder(), writer, level)
}

// CustomTimeEncoder Custom log output time format | 自定义日志输出时间格式
func (z *_zap) CustomTimeEncoder(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
	encoder.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// GetZapCores Get []zapcore.Core based on configuration file Level | 根据配置文件的Level获取 []zapcore.Core
func (z *_zap) GetZapCores() []zapcore.Core {
	cores := make([]zapcore.Core, 0, 7)
	if configs.Debug {
		for level := zapcore.DebugLevel; level <= zapcore.FatalLevel; level++ {
			cores = append(cores, z.GetEncoderCore(level, z.GetLevelPriority(level)))
		}
	} else {
		for level := zapcore.InfoLevel; level <= zapcore.FatalLevel; level++ {
			cores = append(cores, z.GetEncoderCore(level, z.GetLevelPriority(level)))
		}
	}
	return cores
}

// GetLevelPriority Get zap.LevelEnablerFunc based on zapcore.Level | 根据 zapcore.Level 获取 zap.LevelEnablerFunc
func (z *_zap) GetLevelPriority(level zapcore.Level) zap.LevelEnablerFunc {
	switch level {
	case zapcore.DebugLevel:
		return func(level zapcore.Level) bool { // Debug level | 调试级别
			return level == zap.DebugLevel
		}
	case zapcore.InfoLevel:
		return func(level zapcore.Level) bool { // Log level | 日志级别
			return level == zap.InfoLevel
		}
	case zapcore.WarnLevel:
		return func(level zapcore.Level) bool { // Warning level | 警告级别
			return level == zap.WarnLevel
		}
	case zapcore.ErrorLevel:
		return func(level zapcore.Level) bool { // Error level | 错误级别
			return level == zap.ErrorLevel
		}
	case zapcore.DPanicLevel:
		return func(level zapcore.Level) bool { // dpanic level | dpanic级别
			return level == zap.DPanicLevel
		}
	case zapcore.PanicLevel:
		return func(level zapcore.Level) bool { // panic level | panic级别
			return level == zap.PanicLevel
		}
	case zapcore.FatalLevel:
		return func(level zapcore.Level) bool { // Fatal level | 终止级别
			return level == zap.FatalLevel
		}
	default:
		return func(level zapcore.Level) bool { // Debug level | 调试级别
			return level == zap.DebugLevel
		}
	}
}

// Zap Initialize logger | 初始化日志
func Zap() (logger *zap.Logger) {
	if err := utils.CreatNestedFolder("logs"); err != nil { // Check if Director folder exists | 判断是否有Director文件夹
		fmt.Printf("create %v directory err\n", "logs")
	}
	var z = new(_zap)
	cores := z.GetZapCores()
	logger = zap.New(zapcore.NewTee(cores...))
	logger = logger.WithOptions(zap.AddCaller())
	return logger
}
