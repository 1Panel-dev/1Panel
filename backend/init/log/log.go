package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/log"

	"github.com/1Panel-dev/1Panel/backend/configs"
	"github.com/1Panel-dev/1Panel/backend/global"

	"github.com/sirupsen/logrus"
)

const (
	TimeFormat         = "2006-01-02 15:04:05"
	FileTImeFormat     = "2006-01-02"
	RollingTimePattern = "0 0  * * *"
)

func Init() {
	l := logrus.New()
	setOutput(l, global.CONF.LogConfig)
	global.LOG = l
	global.LOG.Info("init logger successfully")
}

func setOutput(logger *logrus.Logger, config configs.LogConfig) {
	writer, err := log.NewWriterFromConfig(&log.Config{
		LogPath:            global.CONF.System.LogPath,
		FileName:           config.LogName,
		TimeTagFormat:      FileTImeFormat,
		MaxRemain:          config.MaxBackup,
		RollingTimePattern: RollingTimePattern,
		LogSuffix:          config.LogSuffix,
	})
	if err != nil {
		panic(err)
	}
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		panic(err)
	}
	fileAndStdoutWriter := io.MultiWriter(writer, os.Stdout)

	logger.SetOutput(fileAndStdoutWriter)
	logger.SetLevel(level)
	logger.SetFormatter(new(MineFormatter))
}

type MineFormatter struct{}

func (s *MineFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	detailInfo := ""
	if entry.Caller != nil {
		function := strings.ReplaceAll(entry.Caller.Function, "github.com/1Panel-dev/1Panel/backend/", "")
		detailInfo = fmt.Sprintf("(%s: %d)", function, entry.Caller.Line)
	}
	if len(entry.Data) == 0 {
		msg := fmt.Sprintf("[%s] [%s] %s %s \n", time.Now().Format(TimeFormat), strings.ToUpper(entry.Level.String()), entry.Message, detailInfo)
		return []byte(msg), nil
	}
	msg := fmt.Sprintf("[%s] [%s] %s %s {%v} \n", time.Now().Format(TimeFormat), strings.ToUpper(entry.Level.String()), entry.Message, detailInfo, entry.Data)
	return []byte(msg), nil
}
