package log

import (
	"github.com/1Panel-dev/1Panel/configs"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"path"
)

func Init() {
	l := logrus.New()
	setOutput(l, global.Config.LogConfig)
	global.Logger = l
}

func setOutput(log *logrus.Logger, config configs.LogConfig) {
	filePath := path.Join(config.Path, config.LogName+config.LogSuffix)
	logPrint := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    config.LogSize,   // 日志文件大小，单位是 MB
		MaxBackups: config.LogBackup, // 最大过期日志保留个数
		MaxAge:     config.LogData,   // 保留过期文件最大时间，单位 天
		Compress:   true,             // 是否压缩日志，默认是不压缩。这里设置为true，压缩日志
	}
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logPrint)
	log.SetLevel(level)
}
