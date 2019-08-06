package logger

import (
	"fmt"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/pkg/errors"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"io"
	"newgateway/config"
	"os"
	"runtime"
	"strings"
	"time"
)


var logger = logrus.New()

// 封装logrus.Fields
type Fields logrus.Fields

func GetLogger() *logrus.Logger {
	return logger
}

func init() {
	initLogger()
}
func initLogger() {
	switch level := strings.ToLower(config.GetConfig().Log.Level); level {
	case "debug":
		SetLogLevel(logrus.DebugLevel)
	case "info":
		SetLogLevel(logrus.InfoLevel)
	case "warn":
		SetLogLevel(logrus.WarnLevel)
	case "error":
		SetLogLevel(logrus.ErrorLevel)
	default:
		SetLogLevel(logrus.InfoLevel)
	}

	logPath := config.GetConfig().Log.File.Path
	var errorWriter, infoWriter io.Writer
	if logPath != "" {
		infoLog := logPath + "/info.log"
		errorLog := logPath + "/error.log"
		// 日志分割
		var err error

		logFile := config.GetConfig().Log.File
		errorWriter, err = rotatelogs.New(
			errorLog+".%Y%m%d%H%M",
			rotatelogs.WithLinkName(errorLog),                                        // 生成软链，指向最新日志文件
			rotatelogs.WithMaxAge(time.Duration(logFile.MaxHour)*time.Hour),          // 文件最大保存时间
			rotatelogs.WithRotationTime(time.Duration(logFile.RotateHour)*time.Hour), // 日志切割时间间隔
		)
		if err != nil {
			logger.Errorf("config local file system logger error. %v", errors.WithStack(err))
		}
		infoWriter, err = rotatelogs.New(
			infoLog+".%Y%m%d%H%M",
			rotatelogs.WithLinkName(infoLog),                                         // 生成软链，指向最新日志文件
			rotatelogs.WithMaxAge(time.Duration(logFile.MaxHour)*time.Hour),          // 文件最大保存时间
			rotatelogs.WithRotationTime(time.Duration(logFile.RotateHour)*time.Hour), // 日志切割时间间隔
		)
		if err != nil {
			logger.Errorf("config local file system logger error. %v", errors.WithStack(err))
		}
		formatter := &logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "@timestamp",
				logrus.FieldKeyLevel: "@level",
				logrus.FieldKeyMsg:   "@message",
				logrus.FieldKeyFunc:  "@caller",
			},
		}
		lfHook := lfshook.NewHook(lfshook.WriterMap{
			logrus.DebugLevel: infoWriter, // 为不同级别设置不同的输出目的
			logrus.InfoLevel:  infoWriter,
			logrus.WarnLevel:  infoWriter,
			logrus.ErrorLevel: errorWriter,
			logrus.FatalLevel: errorWriter,
			logrus.PanicLevel: errorWriter,
		}, formatter)
		logger.AddHook(lfHook)
	} else {
		infoWriter = os.Stdout
		errorWriter = os.Stderr
	}
}

func SetLogLevel(level logrus.Level) {
	logger.Level = level
}

func SetLogFormatter(formatter logrus.Formatter) {
	logger.Formatter = formatter
}

// Debug
func Debug(args ...interface{}) {
	if logger.Level >= logrus.DebugLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Debug(args...)
	}
}

// 带有field的Debug
func DebugWithFields(l interface{}, f Fields) {
	if logger.Level >= logrus.DebugLevel {
		entry := logger.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Debug(l)
	}
}

// Info
func Info(args ...interface{}) {
	if logger.Level >= logrus.InfoLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Info(args...)
	}
}

// Info
func Infof(format string, args ...interface{}) {
	arg := fmt.Sprintf(format, args...)
	if logger.Level >= logrus.InfoLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Info(arg)
	}
}

// 带有field的Info
func InfoWithFields(l interface{}, f Fields) {
	if logger.Level >= logrus.InfoLevel {
		entry := logger.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Info(l)
	}
}

// Warn
func Warn(args ...interface{}) {
	if logger.Level >= logrus.WarnLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Warn(args...)
	}
}
func Warnf(format string, args ...interface{}) {
	arg := fmt.Sprintf(format, args...)
	if logger.Level >= logrus.WarnLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Warn(arg)
	}
}

// 带有Field的Warn
func WarnWithFields(l interface{}, f Fields) {
	if logger.Level >= logrus.WarnLevel {
		entry := logger.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Warn(l)
	}
}

// Error
func Error(args ...interface{}) {
	if logger.Level >= logrus.ErrorLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Error(args...)
	}
}

// Error
func Errorf(format string, args ...interface{}) {
	arg := fmt.Sprintf(format, args...)
	if logger.Level >= logrus.ErrorLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Error(arg)
	}
}

// 带有Fields的Error
func ErrorWithFields(l interface{}, f Fields) {
	if logger.Level >= logrus.ErrorLevel {
		entry := logger.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Error(l)
	}
}

// Fatal
func Fatal(args ...interface{}) {
	if logger.Level >= logrus.FatalLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Fatal(args...)
	}
}

// 带有Field的Fatal
func FatalWithFields(l interface{}, f Fields) {
	if logger.Level >= logrus.FatalLevel {
		entry := logger.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Fatal(l)
	}
}

// Panic
func Panic(args ...interface{}) {
	if logger.Level >= logrus.PanicLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Panic(args...)
	}
}

// 带有Field的Panic
func PanicWithFields(l interface{}, f Fields) {
	if logger.Level >= logrus.PanicLevel {
		entry := logger.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Panic(l)
	}
}

func fileInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}
