package log

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/log"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/configs"
	"github.com/1Panel-dev/1Panel/backend/global"

	"github.com/sirupsen/logrus"
)

func Init() {
	l := logrus.New()
	setOutput(l, global.CONF.LogConfig)
	global.LOG = l
	global.LOG.Info("init success")
}

func setOutput(logger *logrus.Logger, config configs.LogConfig) {

	writer, err := log.NewWriterFromConfig(&log.Config{
		LogPath:       config.Path,
		FileName:      config.LogName,
		TimeTagFormat: "2006-01-02-15-04-05",
		MaxRemain:     config.LogBackup,
	})
	if err != nil {
		panic(err)
	}
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		panic(err)
	}
	logger.SetOutput(writer)
	logger.SetLevel(level)
	logger.SetFormatter(new(MineFormatter))
}

type MineFormatter struct{}

const TimeFormat = "2006-01-02 15:04:05"

func (s *MineFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var cstSh, _ = time.LoadLocation(global.CONF.LogConfig.TimeZone)
	detailInfo := ""
	if entry.Caller != nil {
		funcion := strings.ReplaceAll(entry.Caller.Function, "github.com/1Panel-dev/1Panel/backend/", "")
		detailInfo = fmt.Sprintf("(%s: %d)", funcion, entry.Caller.Line)
	}
	if len(entry.Data) == 0 {
		msg := fmt.Sprintf("[%s] [%s] %s %s \n", time.Now().In(cstSh).Format(TimeFormat), strings.ToUpper(entry.Level.String()), entry.Message, detailInfo)
		return []byte(msg), nil
	}
	msg := fmt.Sprintf("[%s] [%s] %s %s {%v} \n", time.Now().In(cstSh).Format(TimeFormat), strings.ToUpper(entry.Level.String()), entry.Message, detailInfo, entry.Data)
	return []byte(msg), nil
}
