package csvexport

import (
	"encoding/csv"
	"os"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/i18n"
)

func ExportSSHLogs(filename string, logs []dto.SSHHistory) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{
		i18n.GetMsgByKey("ExportIP"),
		i18n.GetMsgByKey("ExportArea"),
		i18n.GetMsgByKey("ExportPort"),
		i18n.GetMsgByKey("ExportAuthMode"),
		i18n.GetMsgByKey("ExportUser"),
		i18n.GetMsgByKey("ExportStatus"),
		i18n.GetMsgByKey("ExportDate"),
	}); err != nil {
		return err
	}

	for _, log := range logs {
		record := []string{
			log.Address,
			log.Area,
			log.Port,
			log.AuthMode,
			log.User,
			log.Status,
			log.Date.Format(constant.DateTimeLayout),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
