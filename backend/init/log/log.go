package log

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/configs"
	"github.com/1Panel-dev/1Panel/global"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

func Init() {
	l := logrus.New()
	setOutput(l, global.CONF.LogConfig)
	global.LOG = l
}

func setOutput(log *logrus.Logger, config configs.LogConfig) {
	filePath := path.Join(config.Path, config.LogName+config.LogSuffix)
	logPrint := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    config.LogSize,
		MaxBackups: config.LogBackup,
		MaxAge:     config.LogData,
		Compress:   true,
	}
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logPrint)
	log.SetLevel(level)
	log.SetFormatter(new(MineFormatter))
}

type MineFormatter struct{}

const TimeFormat = "2006-01-02 15:04:05"

func (s *MineFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var cstSh, _ = time.LoadLocation(global.CONF.LogConfig.TimeZone)
	detailInfo := ""
	if entry.Caller != nil {
		funcion := strings.ReplaceAll(entry.Caller.Function, "github.com/1Panel-dev/1Panel/", "")
		detailInfo = fmt.Sprintf("(%s: %d)", funcion, entry.Caller.Line)
	}
	if len(entry.Data) == 0 {
		msg := fmt.Sprintf("[%s] [%s] %s %s \n", time.Now().In(cstSh).Format(TimeFormat), strings.ToUpper(entry.Level.String()), entry.Message, detailInfo)
		return []byte(msg), nil
	}
	msg := fmt.Sprintf("[%s] [%s] %s %s {%v} \n", time.Now().In(cstSh).Format(TimeFormat), strings.ToUpper(entry.Level.String()), entry.Message, detailInfo, entry.Data)
	return []byte(msg), nil
}
